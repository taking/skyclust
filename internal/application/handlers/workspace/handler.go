package workspace

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"skyclust/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles workspace-related HTTP requests using improved patterns
type Handler struct {
	*handlers.BaseHandler
	workspaceService  domain.WorkspaceService
	userService       domain.UserService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new workspace handler
func NewHandler(workspaceService domain.WorkspaceService, userService domain.UserService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("workspace"),
		workspaceService:  workspaceService,
		userService:       userService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// CreateWorkspace handles workspace creation requests using decorator pattern
func (h *Handler) CreateWorkspace(c *gin.Context) {
	var req domain.CreateWorkspaceRequest

	handler := h.Compose(
		h.createWorkspaceHandler(req),
		h.StandardCRUDDecorators("create_workspace")...,
	)

	handler(c)
}

// createWorkspaceHandler is the core business logic for workspace creation
func (h *Handler) createWorkspaceHandler(req domain.CreateWorkspaceRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		userID := h.extractUserID(c)

		// Extract request from body (JSON binding only, without OwnerID validation)
		var req domain.CreateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeValidationFailed, "Request validation failed", 400), "create_workspace")
			return
		}

		// Set OwnerID from authenticated user (handler responsibility, not client input)
		req.OwnerID = userID.String()

		// Now validate with OwnerID set
		if err := req.Validate(); err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		h.logWorkspaceCreationAttempt(c, userID, req)

		workspace, err := h.workspaceService.CreateWorkspace(ctx, req)
		if err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		h.logWorkspaceCreationSuccess(c, userID, workspace)
		h.setTelemetryAttributes(span, userID, workspace.ID, workspace.Name, "create_workspace")
		h.Created(c, workspace, readability.SuccessMsgUserCreated)
	}
}

// GetWorkspaces handles workspace listing requests using decorator pattern
func (h *Handler) GetWorkspaces(c *gin.Context) {
	handler := h.Compose(
		h.getWorkspacesHandler(),
		h.StandardCRUDDecorators("get_workspaces")...,
	)

	handler(c)
}

// getWorkspacesHandler is the core business logic for getting workspaces
func (h *Handler) getWorkspacesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		userID := h.extractUserID(c)

		h.logWorkspacesRequest(c, userID)

		limit, offset := h.ParsePaginationParams(c)

		workspaces, err := h.workspaceService.GetUserWorkspaces(ctx, userID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspaces")
			return
		}

		h.setTelemetryAttributes(span, userID, "", "", "get_workspaces")
		h.OK(c, gin.H{
			"workspaces": workspaces,
			"pagination": gin.H{
				"total":        len(workspaces),
				"limit":        limit,
				"offset":       offset,
				"current_page": (offset / limit) + 1,
				"total_pages":  (len(workspaces) + limit - 1) / limit,
			},
		}, "Workspaces retrieved successfully")
	}
}

// GetWorkspace handles workspace retrieval requests using decorator pattern
func (h *Handler) GetWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.getWorkspaceHandler(),
		h.StandardCRUDDecorators("get_workspace")...,
	)

	handler(c)
}

// getWorkspaceHandler is the core business logic for getting a single workspace
func (h *Handler) getWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID := h.parseWorkspaceID(c)
		userID := h.extractUserID(c)

		h.logWorkspaceRequest(c, userID, workspaceID)

		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspace")
			return
		}

		h.setTelemetryAttributes(span, userID, workspace.ID, workspace.Name, "get_workspace")
		h.OK(c, workspace, "Workspace retrieved successfully")
	}
}

// UpdateWorkspace handles workspace update requests using decorator pattern
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	var req domain.UpdateWorkspaceRequest

	handler := h.Compose(
		h.updateWorkspaceHandler(req),
		h.StandardCRUDDecorators("update_workspace")...,
	)

	handler(c)
}

// updateWorkspaceHandler is the core business logic for updating workspaces
func (h *Handler) updateWorkspaceHandler(req domain.UpdateWorkspaceRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID := h.parseWorkspaceID(c)
		req = h.extractValidatedUpdateRequest(c)
		userID := h.extractUserID(c)

		// Get workspace to check permissions
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		// Check if user has permission to update (owner or admin)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "update_workspace")
			return
		}

		// Only owner or admin can update
		if userID != ownerID && userRole != domain.AdminRoleType {
			h.Forbidden(c, "Insufficient permissions to update this workspace")
			return
		}

		h.logWorkspaceUpdateAttempt(c, userID, workspaceID)

		// Update workspace (owner_id remains unchanged)
		updatedWorkspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceID.String(), req)
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		h.logWorkspaceUpdateSuccess(c, userID, updatedWorkspace)
		h.setTelemetryAttributes(span, userID, updatedWorkspace.ID, updatedWorkspace.Name, "update_workspace")
		h.OK(c, updatedWorkspace, readability.SuccessMsgUserUpdated)
	}
}

// DeleteWorkspace handles workspace deletion requests using decorator pattern
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.deleteWorkspaceHandler(),
		h.StandardCRUDDecorators("delete_workspace")...,
	)

	handler(c)
}

// deleteWorkspaceHandler is the core business logic for deleting workspaces
func (h *Handler) deleteWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID := h.parseWorkspaceID(c)
		userID := h.extractUserID(c)

		h.logWorkspaceDeletionAttempt(c, userID, workspaceID)

		err := h.workspaceService.DeleteWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "delete_workspace")
			return
		}

		h.logWorkspaceDeletionSuccess(c, userID, workspaceID)
		h.setTelemetryAttributes(span, userID, workspaceID.String(), "", "delete_workspace")
		h.OK(c, gin.H{"message": "Workspace deleted successfully"}, readability.SuccessMsgUserDeleted)
	}
}

// Helper methods for better readability

func (h *Handler) extractValidatedUpdateRequest(c *gin.Context) domain.UpdateWorkspaceRequest {
	var req domain.UpdateWorkspaceRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_workspace")
		return domain.UpdateWorkspaceRequest{}
	}
	return req
}

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	
	// Convert to uuid.UUID (handle both string and uuid.UUID types)
	switch v := userIDValue.(type) {
	case uuid.UUID:
		return v
	case string:
		parsedUserID, err := uuid.Parse(v)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID format", 401), "extract_user_id")
			return uuid.Nil
		}
		return parsedUserID
	default:
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID type", 401), "extract_user_id")
		return uuid.Nil
	}
}

func (h *Handler) parseWorkspaceID(c *gin.Context) uuid.UUID {
	workspaceIDStr := c.Param("id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid workspace ID format", 400), "parse_workspace_id")
		return uuid.Nil
	}
	return workspaceID
}

func (h *Handler) setTelemetryAttributes(span telemetry.Span, userID uuid.UUID, workspaceID, workspaceName, action string) {
	if span != nil {
		telemetry.SetAttributes(span, map[string]interface{}{
			"user_id":        userID.String(),
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
			"action":         action,
		})
	}
}

// Logging helper methods

func (h *Handler) logWorkspaceCreationAttempt(c *gin.Context, userID uuid.UUID, req domain.CreateWorkspaceRequest) {
	h.LogBusinessEvent(c, "workspace_creation_attempted", userID.String(), "", map[string]interface{}{
		"workspace_name": req.Name,
		"description":    req.Description,
	})
}

func (h *Handler) logWorkspaceCreationSuccess(c *gin.Context, userID uuid.UUID, workspace *domain.Workspace) {
	h.LogBusinessEvent(c, "workspace_created", userID.String(), workspace.ID, map[string]interface{}{
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
		"owner_id":       workspace.OwnerID,
	})
}

func (h *Handler) logWorkspacesRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "workspaces_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_workspaces",
	})
}

func (h *Handler) logWorkspaceRequest(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_requested", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

func (h *Handler) logWorkspaceUpdateAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_update_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

func (h *Handler) logWorkspaceUpdateSuccess(c *gin.Context, userID uuid.UUID, workspace *domain.Workspace) {
	h.LogBusinessEvent(c, "workspace_updated", userID.String(), workspace.ID, map[string]interface{}{
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
	})
}

func (h *Handler) logWorkspaceDeletionAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_deletion_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

func (h *Handler) logWorkspaceDeletionSuccess(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_deleted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

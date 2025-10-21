package workspace

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"skyclust/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles workspace-related HTTP requests
type Handler struct {
	*handlers.BaseHandler
	workspaceService domain.WorkspaceService
	userService      domain.UserService
}

// NewHandler creates a new workspace handler
func NewHandler(workspaceService domain.WorkspaceService, userService domain.UserService) *Handler {
	return &Handler{
		BaseHandler:      handlers.NewBaseHandler("workspace"),
		workspaceService: workspaceService,
		userService:      userService,
	}
}

// CreateWorkspace handles workspace creation requests
func (h *Handler) CreateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	var req domain.CreateWorkspaceRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_workspace")
		return
	}

	// Set owner ID from authenticated user
	req.OwnerID = userID.String()

	workspace, err := h.workspaceService.CreateWorkspace(ctx, req)
	if err != nil {
		h.HandleError(c, err, "create_workspace")
		return
	}

	// Log successful creation
	h.LogInfo(c, "Workspace created successfully",
		zap.String("workspace_id", workspace.ID),
		zap.String("workspace_name", workspace.Name),
		zap.String("owner_id", workspace.OwnerID),
	)

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":        userID.String(),
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
		"action":         "create_workspace",
	})

	h.Created(c, workspace, "Workspace created successfully")
}

// GetWorkspaces handles workspace listing requests
func (h *Handler) GetWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_workspaces")
		return
	}

	// Parse pagination parameters
	limit, offset := h.ParsePaginationParams(c)

	// Get workspaces
	workspaces, err := h.workspaceService.GetUserWorkspaces(ctx, userID.String())
	if err != nil {
		h.HandleError(c, err, "get_workspaces")
		return
	}

	// Log successful retrieval
	h.LogInfo(c, "Workspaces retrieved successfully",
		zap.Int("count", len(workspaces)),
	)

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id": userID.String(),
		"count":   len(workspaces),
		"action":  "get_workspaces",
	})

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

// GetWorkspace handles workspace retrieval requests
func (h *Handler) GetWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_workspace")
		return
	}

	// Get workspace
	workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
	if err != nil {
		h.HandleError(c, err, "get_workspace")
		return
	}

	// Log successful retrieval
	h.LogInfo(c, "Workspace retrieved successfully",
		zap.String("workspace_id", workspace.ID),
		zap.String("workspace_name", workspace.Name),
	)

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspace.ID,
		"action":       "get_workspace",
	})

	h.OK(c, workspace, "Workspace retrieved successfully")
}

// UpdateWorkspace handles workspace update requests
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	var req domain.UpdateWorkspaceRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_workspace")
		return
	}

	// Update workspace
	workspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceID.String(), req)
	if err != nil {
		h.HandleError(c, err, "update_workspace")
		return
	}

	// Log successful update
	h.LogInfo(c, "Workspace updated successfully",
		zap.String("workspace_id", workspace.ID),
		zap.String("workspace_name", workspace.Name),
	)

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspace.ID,
		"action":       "update_workspace",
	})

	h.OK(c, workspace, "Workspace updated successfully")
}

// DeleteWorkspace handles workspace deletion requests
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_workspace")
		return
	}

	// Delete workspace
	err = h.workspaceService.DeleteWorkspace(ctx, workspaceID.String())
	if err != nil {
		h.HandleError(c, err, "delete_workspace")
		return
	}

	// Log successful deletion
	h.LogInfo(c, "Workspace deleted successfully",
		zap.String("workspace_id", workspaceID.String()),
	)

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspaceID.String(),
		"action":       "delete_workspace",
	})

	h.OK(c, gin.H{"message": "Workspace deleted successfully"}, "Workspace deleted successfully")
}

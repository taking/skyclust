package http

import (
	"net/http"
	"skyclust/internal/domain"

	"skyclust/pkg/logger"
	"skyclust/pkg/telemetry"

	"github.com/gin-gonic/gin"
)

// WorkspaceHandler handles workspace-related HTTP requests
type WorkspaceHandler struct {
	workspaceService domain.WorkspaceService
	userService      domain.UserService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(workspaceService domain.WorkspaceService, userService domain.UserService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		userService:      userService,
	}
}

// CreateWorkspace handles workspace creation requests
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req domain.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		UnauthorizedResponse(c, "Invalid user ID")
		return
	}

	// Set the owner ID from the authenticated user
	req.OwnerID = userIDStr

	workspace, err := h.workspaceService.CreateWorkspace(ctx, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Errorf("Failed to create workspace: %v", err)
		InternalServerErrorResponse(c, "Failed to create workspace")
		return
	}

	telemetry.AddEvent(span, "workspace.created", map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner_id":     workspace.OwnerID,
	})

	CreatedResponse(c, gin.H{"workspace": workspace}, "Workspace created successfully")
}

// GetWorkspace handles workspace retrieval requests
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(domain.NewDomainError(domain.ErrCodeValidationFailed, "Workspace ID is required", 400))
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to get workspace")
		_ = c.Error(err)
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"workspace_id": workspace.ID,
		"action":       "get_workspace",
	})

	OKResponse(c, gin.H{"workspace": workspace}, "Workspace retrieved successfully")
}

// UpdateWorkspace handles workspace update requests
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(domain.NewDomainError(domain.ErrCodeValidationFailed, "Workspace ID is required", 400))
		return
	}

	var req domain.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(domain.NewDomainError(domain.ErrCodeValidationFailed, "Invalid request body", 400))
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceID, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to update workspace")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "workspace.updated", map[string]interface{}{
		"workspace_id": workspace.ID,
	})

	OKResponse(c, gin.H{"workspace": workspace}, "Workspace retrieved successfully")
}

// DeleteWorkspace handles workspace deletion requests
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(domain.NewDomainError(domain.ErrCodeValidationFailed, "Workspace ID is required", 400))
		return
	}

	err := h.workspaceService.DeleteWorkspace(ctx, workspaceID)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to delete workspace")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "workspace.deleted", map[string]interface{}{
		"workspace_id": workspaceID,
	})

	c.JSON(http.StatusNoContent, nil)
}

// ListWorkspaces handles workspace listing requests
// GetWorkspaces handles workspace list requests
func (h *WorkspaceHandler) GetWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		UnauthorizedResponse(c, "Invalid user ID")
		return
	}

	workspaces, err := h.workspaceService.GetUserWorkspaces(ctx, userIDStr)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Errorf("Failed to list workspaces: %v", err)
		InternalServerErrorResponse(c, "Failed to list workspaces")
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":         userIDStr,
		"workspace_count": len(workspaces),
		"action":          "list_workspaces",
	})

	OKResponse(c, gin.H{"workspaces": workspaces}, "Workspaces retrieved successfully")
}

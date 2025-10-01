package http

import (
	"net/http"

	"cmp/internal/domain"
	"cmp/pkg/shared/errors"
	"cmp/pkg/shared/logger"
	"cmp/pkg/shared/telemetry"

	"github.com/gin-gonic/gin"
)

// CreateWorkspace handles workspace creation requests
func (h *Handler) CreateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req domain.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(errors.NewValidationError("Invalid request body"))
		return
	}

	workspace, err := h.container.WorkspaceService.CreateWorkspace(ctx, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to create workspace")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "workspace.created", map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner_id":     workspace.OwnerID,
	})

	h.CreatedResponse(c, gin.H{"workspace": workspace}, "Workspace created successfully")
}

// GetWorkspace handles workspace retrieval requests
func (h *Handler) GetWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(errors.NewValidationError("Workspace ID is required"))
		return
	}

	workspace, err := h.container.WorkspaceService.GetWorkspace(ctx, workspaceID)
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

	h.OKResponse(c, gin.H{"workspace": workspace}, "Workspace retrieved successfully")
}

// UpdateWorkspace handles workspace update requests
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(errors.NewValidationError("Workspace ID is required"))
		return
	}

	var req domain.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(errors.NewValidationError("Invalid request body"))
		return
	}

	workspace, err := h.container.WorkspaceService.UpdateWorkspace(ctx, workspaceID, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to update workspace")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "workspace.updated", map[string]interface{}{
		"workspace_id": workspace.ID,
	})

	h.OKResponse(c, gin.H{"workspace": workspace}, "Workspace retrieved successfully")
}

// DeleteWorkspace handles workspace deletion requests
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		_ = c.Error(errors.NewValidationError("Workspace ID is required"))
		return
	}

	err := h.container.WorkspaceService.DeleteWorkspace(ctx, workspaceID)
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
func (h *Handler) ListWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	// Get user ID from query parameter or context
	userID := c.Query("user_id")
	if userID == "" {
		_ = c.Error(errors.NewValidationError("User ID is required"))
		return
	}

	workspaces, err := h.container.WorkspaceService.GetUserWorkspaces(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to list workspaces")
		_ = c.Error(err)
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":         userID,
		"workspace_count": len(workspaces),
		"action":          "list_workspaces",
	})

	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
}

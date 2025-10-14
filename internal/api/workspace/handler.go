package workspace

import (
	"net/http"
	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/utils"
	"strconv"

	"skyclust/pkg/logger"
	"skyclust/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles workspace-related HTTP requests
type Handler struct {
	workspaceService   domain.WorkspaceService
	userService        domain.UserService
	tokenExtractor     *utils.TokenExtractor
	performanceTracker *common.PerformanceTracker
	requestLogger      *common.RequestLogger
	validationRules    *common.ValidationRules
}

// NewHandler creates a new workspace handler
func NewHandler(workspaceService domain.WorkspaceService, userService domain.UserService) *Handler {
	return &Handler{
		workspaceService:   workspaceService,
		userService:        userService,
		tokenExtractor:     utils.NewTokenExtractor(),
		performanceTracker: common.NewPerformanceTracker("workspace"),
		requestLogger:      common.NewRequestLogger(nil),
		validationRules:    common.NewValidationRules(),
	}
}

// CreateWorkspace handles workspace creation requests
func (h *Handler) CreateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	var req domain.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(ctx, req)
	if err != nil {
		logger.Errorf("Failed to create workspace: %v", err)
		common.InternalServerError(c, "Failed to create workspace")
		return
	}

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":        userID.String(),
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
		"action":         "create_workspace",
	})

	common.Success(c, http.StatusCreated, workspace, "Workspace created successfully")
}

// GetWorkspaces handles workspace listing requests
func (h *Handler) GetWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		common.InternalServerError(c, "Invalid user ID format")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err2 := strconv.Atoi(offsetStr)
	if err2 != nil || offset < 0 {
		offset = 0
	}

	// Get workspaces for user
	workspaces, err := h.workspaceService.GetUserWorkspaces(c.Request.Context(), userID.String())
	if err != nil {
		common.InternalServerError(c, "Failed to retrieve workspaces")
		return
	}

	// Apply pagination manually
	total := len(workspaces)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	if start < 0 {
		start = 0
	}

	workspaces = workspaces[start:end]

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":         userID.String(),
		"workspace_count": len(workspaces),
		"limit":           limit,
		"offset":          offset,
		"action":          "list_workspaces",
	})

	common.OK(c, gin.H{
		"workspaces": workspaces,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	}, "Workspaces retrieved successfully")
}

// GetWorkspace handles single workspace retrieval requests
func (h *Handler) GetWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceIDStr := c.Param("id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid workspace ID format")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceIDStr)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "Workspace not found")
			return
		}
		logger.Errorf("Failed to get workspace: %v", err)
		common.InternalServerError(c, "Failed to get workspace")
		return
	}

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspaceID.String(),
		"action":       "get_workspace",
	})

	common.OK(c, workspace, "Workspace retrieved successfully")
}

// UpdateWorkspace handles workspace update requests
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceIDStr := c.Param("id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid workspace ID format")
		return
	}

	var req domain.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceIDStr, req)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "Workspace not found")
			return
		}
		logger.Errorf("Failed to update workspace: %v", err)
		common.InternalServerError(c, "Failed to update workspace")
		return
	}

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspaceID.String(),
		"action":       "update_workspace",
	})

	common.OK(c, workspace, "Workspace updated successfully")
}

// DeleteWorkspace handles workspace deletion requests
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)

	workspaceIDStr := c.Param("id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid workspace ID format")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	err = h.workspaceService.DeleteWorkspace(ctx, workspaceIDStr)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "Workspace not found")
			return
		}
		logger.Errorf("Failed to delete workspace: %v", err)
		common.InternalServerError(c, "Failed to delete workspace")
		return
	}

	// Set telemetry attributes
	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id":      userID.String(),
		"workspace_id": workspaceID.String(),
		"action":       "delete_workspace",
	})

	common.OK(c, gin.H{"message": "Workspace deleted successfully"}, "Workspace deleted successfully")
}

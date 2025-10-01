package http

import (
	"net/http"

	"cmp/internal/domain"
	"cmp/pkg/shared/errors"
	"cmp/pkg/shared/logger"
	"cmp/pkg/shared/telemetry"

	"github.com/gin-gonic/gin"
)

// CreateUser handles user creation requests
func (h *Handler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(errors.NewValidationError("Invalid request body"))
		return
	}

	user, err := h.container.UserService.CreateUser(ctx, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to create user")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "user.created", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})

	h.CreatedResponse(c, gin.H{"user": user}, "User created successfully")
}

// GetUser handles user retrieval requests
func (h *Handler) GetUser(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	userID := c.Param("id")
	if userID == "" {
		_ = c.Error(errors.NewValidationError("User ID is required"))
		return
	}

	user, err := h.container.UserService.GetUser(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to get user")
		_ = c.Error(err)
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"user_id": user.ID,
		"action":  "get_user",
	})

	h.OKResponse(c, gin.H{"user": user}, "User retrieved successfully")
}

// UpdateUser handles user update requests
func (h *Handler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	userID := c.Param("id")
	if userID == "" {
		_ = c.Error(errors.NewValidationError("User ID is required"))
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(errors.NewValidationError("Invalid request body"))
		return
	}

	user, err := h.container.UserService.UpdateUser(ctx, userID, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to update user")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "user.updated", map[string]interface{}{
		"user_id": user.ID,
	})

	h.OKResponse(c, gin.H{"user": user}, "User retrieved successfully")
}

// DeleteUser handles user deletion requests
func (h *Handler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	userID := c.Param("id")
	if userID == "" {
		_ = c.Error(errors.NewValidationError("User ID is required"))
		return
	}

	err := h.container.UserService.DeleteUser(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to delete user")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "user.deleted", map[string]interface{}{
		"user_id": userID,
	})

	c.JSON(http.StatusNoContent, nil)
}

// Authenticate handles user authentication requests
func (h *Handler) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		_ = c.Error(errors.NewValidationError("Invalid request body"))
		return
	}

	user, err := h.container.UserService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Authentication failed")
		_ = c.Error(err)
		return
	}

	telemetry.AddEvent(span, "user.authenticated", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})

	h.OKResponse(c, gin.H{"user": user}, "User retrieved successfully")
}

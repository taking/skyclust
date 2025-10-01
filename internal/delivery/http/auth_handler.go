package http

import (
	"cmp/internal/domain"
	"cmp/pkg/shared/logger"
	"cmp/pkg/shared/telemetry"

	"github.com/gin-gonic/gin"
)

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		h.BadRequestResponse(c, err.Error())
		return
	}

	user, err := h.container.UserService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		telemetry.RecordError(span, err)
		h.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Generate JWT token (simplified)
	token := "mock-jwt-token" // TODO: Implement proper JWT generation

	telemetry.SetAttributes(span, map[string]interface{}{
		"user.id": user.ID,
		"login":   true,
	})

	h.OKResponse(c, gin.H{
		"token": token,
		"user":  user,
	}, "Login successful")
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	// TODO: Implement token blacklisting
	telemetry.SetAttributes(span, map[string]interface{}{
		"logout": true,
	})

	h.OKResponse(c, nil, "Logged out successfully")
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		h.BadRequestResponse(c, err.Error())
		return
	}

	user, err := h.container.UserService.CreateUser(ctx, req)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to register user")
		h.InternalServerErrorResponse(c, "Failed to create user")
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"user.id":       user.ID,
		"user.username": user.Username,
		"register":      true,
	})

	h.CreatedResponse(c, gin.H{"user": user}, "User registered successfully")
}

// ChangePassword handles password change requests
func (h *Handler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	span := telemetry.SpanFromContext(ctx)
	defer span.End()

	userID := c.Param("id")
	if userID == "" {
		h.BadRequestResponse(c, "User ID is required")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.RecordError(span, err)
		h.BadRequestResponse(c, err.Error())
		return
	}

	err := h.container.UserService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		telemetry.RecordError(span, err)
		logger.Error("Failed to change password")
		h.InternalServerErrorResponse(c, "Failed to change password")
		return
	}

	telemetry.SetAttributes(span, map[string]interface{}{
		"user.id":         userID,
		"password_change": true,
	})

	h.OKResponse(c, nil, "Password changed successfully")
}

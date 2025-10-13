package http

import (
	"net/http"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService   domain.AuthService
	userService   domain.UserService
	logoutService *usecase.LogoutService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService domain.AuthService, userService domain.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// NewAuthHandlerWithLogout creates a new authentication handler with logout service
func NewAuthHandlerWithLogout(authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		userService:   userService,
		logoutService: logoutService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Create validation error
		validationErr := domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			"Invalid request body",
			http.StatusBadRequest,
		).WithDetails("binding_error", err.Error())

		DomainErrorResponse(c, validationErr)
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		// Use domain error handling
		if domain.IsDomainError(err) {
			DomainErrorResponse(c, domain.GetDomainError(err))
		} else {
			// Convert to domain error
			internalErr := domain.NewDomainError(
				domain.ErrCodeInternalError,
				"Failed to register user",
				http.StatusInternalServerError,
			).WithDetails("original_error", err.Error())

			DomainErrorResponse(c, internalErr)
		}
		return
	}

	CreatedResponse(c, gin.H{
		"user":  user,
		"token": token,
	}, "User registered successfully")
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Create validation error
		validationErr := domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			"Invalid request body",
			http.StatusBadRequest,
		).WithDetails("binding_error", err.Error())

		DomainErrorResponse(c, validationErr)
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// Use domain error handling
		if domain.IsDomainError(err) {
			DomainErrorResponse(c, domain.GetDomainError(err))
		} else {
			// Convert to domain error
			internalErr := domain.NewDomainError(
				domain.ErrCodeInternalError,
				"Failed to login",
				http.StatusInternalServerError,
			).WithDetails("original_error", err.Error())

			DomainErrorResponse(c, internalErr)
		}
		return
	}

	OKResponse(c, gin.H{
		"user":  user,
		"token": token,
	}, "Login successful")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		BadRequestResponse(c, "Authorization header required")
		return
	}

	// Extract token from "Bearer <token>"
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Try to get user ID from context (if authenticated)
	var userUUID uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if userUUID, ok := userID.(uuid.UUID); ok {
			// User is authenticated, use the user ID from context
			_ = userUUID // Use the userUUID variable
		} else {
			InternalServerErrorResponse(c, "Invalid user ID")
			return
		}
	} else {
		// User is not authenticated, try to extract user ID from token
		// This is a fallback for cases where logout is called without authentication middleware
		// Parse JWT token to extract user_id
		userUUID = h.extractUserIDFromToken(token)
	}

	// Use logout service for blacklist functionality
	if h.logoutService != nil {
		req := usecase.LogoutRequest{
			UserID:   userUUID,
			Token:    token,
			AuthType: "jwt",
		}

		resp, err := h.logoutService.Logout(c.Request.Context(), req)
		if err != nil {
			if domain.IsDomainError(err) {
				DomainErrorResponse(c, domain.GetDomainError(err))
			} else {
				InternalServerErrorResponse(c, "Failed to logout")
			}
			return
		}

		OKResponse(c, gin.H{
			"success": resp.Success,
			"message": resp.Message,
		}, "Logout successful")
		return
	}

	// Fallback to basic auth service logout
	if err := h.authService.Logout(userUUID, token); err != nil {
		InternalServerErrorResponse(c, "Failed to logout")
		return
	}

	OKResponse(c, gin.H{
		"message":           "Logout successful",
		"token_invalidated": true,
	}, "Logout successful")
}

// Me returns the current user information
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userUUID.String())
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get user")
		return
	}

	if user == nil {
		NotFoundResponse(c, "User not found")
		return
	}

	OKResponse(c, user, "User information retrieved successfully")
}

// GetUsers returns a list of users with pagination
func (h *AuthHandler) GetUsers(c *gin.Context) {
	// Parse query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	active := c.Query("active")

	// Convert to integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		BadRequestResponse(c, "Invalid page parameter")
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 || limitInt > 100 {
		BadRequestResponse(c, "Invalid limit parameter (max 100)")
		return
	}

	// Calculate offset
	offset := (pageInt - 1) * limitInt

	// Build filters
	filters := make(map[string]interface{})
	if search != "" {
		filters["search"] = search
	}
	if active != "" {
		if active == "true" {
			filters["is_active"] = true
		} else if active == "false" {
			filters["is_active"] = false
		}
	}

	// Get users from service
	users, total, err := h.userService.GetUsers(c.Request.Context(), limitInt, offset, filters)
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get users")
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limitInt) - 1) / int64(limitInt)

	OKResponse(c, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":         pageInt,
			"limit":        limitInt,
			"total":        total,
			"total_pages":  totalPages,
			"has_next":     pageInt < int(totalPages),
			"has_previous": pageInt > 1,
		},
	}, "Users retrieved successfully")
}

// UpdateUser handles user update requests
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "User ID is required")
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErr := domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			"Invalid request body",
			http.StatusBadRequest,
		).WithDetails("binding_error", err.Error())

		DomainErrorResponse(c, validationErr)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		DomainErrorResponse(c, domain.GetDomainError(err))
		return
	}

	// Update user
	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		if domain.IsDomainError(err) {
			DomainErrorResponse(c, domain.GetDomainError(err))
		} else {
			InternalServerErrorResponse(c, "Failed to update user")
		}
		return
	}

	OKResponse(c, gin.H{"user": user}, "User updated successfully")
}

// DeleteUser handles user deletion requests
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "User ID is required")
		return
	}

	// Delete user
	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		if domain.IsDomainError(err) {
			DomainErrorResponse(c, domain.GetDomainError(err))
		} else {
			InternalServerErrorResponse(c, "Failed to delete user")
		}
		return
	}

	OKResponse(c, gin.H{"message": "User deleted successfully"}, "User deleted successfully")
}

// extractUserIDFromToken extracts user ID from JWT token
func (h *AuthHandler) extractUserIDFromToken(token string) uuid.UUID {
	// This is a simplified JWT parsing for logout purposes
	// In production, you might want to use a proper JWT library
	// For now, we'll return Nil and let the logout service handle it
	return uuid.Nil
}

// GetUser returns a specific user by ID
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequestResponse(c, "User ID is required")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			NotFoundResponse(c, "User not found")
			return
		}
		InternalServerErrorResponse(c, "Failed to get user")
		return
	}

	OKResponse(c, user, "User retrieved successfully")
}

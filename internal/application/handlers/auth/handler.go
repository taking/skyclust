package auth

import (
	"net/http"
	"skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// Handler handles authentication-related HTTP requests
type Handler struct {
	*handlers.BaseHandler
	authService   domain.AuthService
	userService   domain.UserService
	logoutService *service.LogoutService
}

// NewHandler creates a new authentication handler
func NewHandler(authService domain.AuthService, userService domain.UserService) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("auth"),
		authService: authService,
		userService: userService,
	}
}

// NewHandlerWithLogout creates a new authentication handler with logout service
func NewHandlerWithLogout(authService domain.AuthService, userService domain.UserService, logoutService *service.LogoutService) *Handler {
	return &Handler{
		BaseHandler:   handlers.NewBaseHandler("auth"),
		authService:   authService,
		userService:   userService,
		logoutService: logoutService,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "register", http.StatusOK)

	var req domain.CreateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "user_registration_attempt", "", "", map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	})

	// Log audit event
	h.LogAuditEvent(c, "user_registration", "user", "", "", map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	})

	// Check if auth service is available
	if h.authService == nil {
		h.HandleError(c, &domain.DomainError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Authentication service is not available",
		}, "register")
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		// Log error
		h.LogError(c, err, "Failed to register user")
		h.HandleError(c, err, "register")
		return
	}

	// Log successful registration
	h.LogBusinessEvent(c, "user_registration_success", user.ID.String(), "", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	// Log successful audit event
	h.LogAuditEvent(c, "user_registration", "user", user.ID.String(), user.ID.String(), map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})

	h.Created(c, gin.H{
		"token": token,
		"user":  user,
	}, "User registered successfully")
}

// Login handles user login requests
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Extract client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check if auth service is available
	if h.authService == nil {
		h.HandleError(c, &domain.DomainError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Authentication service is not available",
		}, "login")
		return
	}

	user, token, err := h.authService.LoginWithContext(req.Email, req.Password, clientIP, userAgent)
	if err != nil {
		h.HandleError(c, err, "login")
		return
	}

	h.OK(c, gin.H{
		"token": token,
		"user":  user,
	}, "Login successful")
}

// Logout handles user logout requests
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "logout")
		return
	}

	// Get Bearer token from Authorization header
	token, err := h.GetBearerTokenFromHeader(c)
	if err != nil {
		h.HandleError(c, err, "logout")
		return
	}

	// Check if auth service is available
	if h.authService == nil {
		h.HandleError(c, &domain.DomainError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Authentication service is not available",
		}, "logout")
		return
	}

	err = h.authService.Logout(userID, token)
	if err != nil {
		h.HandleError(c, err, "logout")
		return
	}

	h.OK(c, gin.H{"message": "Logout successful"}, "Logout successful")
}

// Me returns the current user's information
func (h *Handler) Me(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "me")
		return
	}

	// Get user by ID
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "me")
		return
	}

	h.OK(c, user, "User retrieved successfully")
}

// GetUsers handles getting all users (admin only)
func (h *Handler) GetUsers(c *gin.Context) {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_users")
		return
	}

	// Check if user has permission to list users (admin role only)
	if userRole != domain.AdminRoleType {
		h.Forbidden(c, "Insufficient permissions to list users")
		return
	}

	// Parse pagination parameters
	limit, offset := h.ParsePaginationParams(c)

	// Get users from service
	users, total, err := h.userService.GetUsersWithFilters(domain.UserFilters{
		Limit: limit,
	})
	if err != nil {
		h.HandleError(c, err, "get_users")
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	currentPage := (offset / limit) + 1

	h.OK(c, gin.H{
		"users": users,
		"pagination": gin.H{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"current_page": currentPage,
			"total_pages":  totalPages,
		},
	}, "Users retrieved successfully")
}

// GetUser handles getting a specific user by ID
func (h *Handler) GetUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "get_user")
		return
	}

	h.OK(c, user, "User retrieved successfully")
}

// UpdateUser handles updating a user
func (h *Handler) UpdateUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	currentUserRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	// Check if user can update this user (own profile or admin)
	if currentUserID != userID && currentUserRole != domain.AdminRoleType {
		h.Forbidden(c, "Insufficient permissions to update this user")
		return
	}

	var req domain.UpdateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	// Update user fields
	if req.Username != nil && *req.Username != "" {
		user.Username = *req.Username
	}
	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	h.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser handles deleting a user
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	currentUserRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	// Check if user can delete this user (admin only, cannot delete self)
	if currentUserRole != domain.AdminRoleType {
		h.Forbidden(c, "Only administrators can delete users")
		return
	}

	// Prevent self-deletion
	if currentUserID == userID {
		h.Forbidden(c, "Cannot delete your own account")
		return
	}

	// Check if user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	// Delete user
	err = h.userService.DeleteUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	h.OK(c, gin.H{"message": "User deleted successfully"}, "User deleted successfully")
}

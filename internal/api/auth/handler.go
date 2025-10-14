package auth

import (
	"net/http"
	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"
	"skyclust/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles authentication-related HTTP requests
type Handler struct {
	authService        domain.AuthService
	userService        domain.UserService
	logoutService      *usecase.LogoutService
	tokenExtractor     *utils.TokenExtractor
	performanceTracker *common.PerformanceTracker
	requestLogger      *common.RequestLogger
	auditLogger        *common.AuditLogger
	validationRules    *common.ValidationRules
}

// NewHandler creates a new authentication handler
func NewHandler(authService domain.AuthService, userService domain.UserService) *Handler {
	return &Handler{
		authService:        authService,
		userService:        userService,
		tokenExtractor:     utils.NewTokenExtractor(),
		performanceTracker: common.NewPerformanceTracker("auth"),
		requestLogger:      common.NewRequestLogger(nil),
		auditLogger:        common.NewAuditLogger(nil),
		validationRules:    common.NewValidationRules(),
	}
}

// NewHandlerWithLogout creates a new authentication handler with logout service
func NewHandlerWithLogout(authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) *Handler {
	return &Handler{
		authService:    authService,
		userService:    userService,
		logoutService:  logoutService,
		tokenExtractor: utils.NewTokenExtractor(),
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	// Start performance tracking
	tracker := common.NewPerformanceTracker("register")
	defer tracker.TrackRequest(c, "", http.StatusOK)

	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Enhanced validation with new validation rules
		validationErrors := make(map[string]string)
		validationErrors["binding"] = err.Error()

		// Additional validation using new validation rules
		if !h.validationRules.ValidateEmail(req.Email) {
			validationErrors["email"] = "Invalid email format"
		}
		if !h.validationRules.ValidateUsername(req.Username) {
			validationErrors["username"] = "Invalid username format"
		}
		if !h.validationRules.ValidatePassword(req.Password) {
			validationErrors["password"] = "Password must be at least 8 characters"
		}

		common.ValidationError(c, validationErrors)
		return
	}

	// Log business event
	common.LogBusinessEvent(c, "user_registration_attempt", "", "", map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	})

	// Log audit event
	auditCtx := common.NewAuditContext("", "user_registration", "user", "").
		WithDetails(map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})

	user, token, err := h.authService.Register(req)
	if err != nil {
		// Log error
		common.LogError(c, err, "Failed to register user")

		// Log failed audit event
		auditCtx.LogFailure(c, h.auditLogger, err.Error())

		// Use domain error handling
		if domain.IsDomainError(err) {
			common.DomainError(c, domain.GetDomainError(err))
		} else {
			// Convert to domain error
			internalErr := domain.NewDomainError(
				domain.ErrCodeInternalError,
				"Failed to register user",
				http.StatusInternalServerError,
			).WithDetails("original_error", err.Error())

			common.DomainError(c, internalErr)
		}
		return
	}

	// Log successful registration
	common.LogBusinessEvent(c, "user_registration_success", user.ID.String(), "", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	// Log successful audit event
	auditCtx.UserID = user.ID.String()
	auditCtx.ResourceID = user.ID.String()
	auditCtx.LogSuccess(c, h.auditLogger)

	common.Success(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	}, "User registered successfully")
}

// Login handles user login requests
func (h *Handler) Login(c *gin.Context) {
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

		common.DomainError(c, validationErr)
		return
	}

	// Extract client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	user, token, err := h.authService.LoginWithContext(req.Username, req.Password, clientIP, userAgent)
	if err != nil {
		// Use domain error handling
		if domain.IsDomainError(err) {
			common.DomainError(c, domain.GetDomainError(err))
		} else {
			// Convert to domain error
			internalErr := domain.NewDomainError(
				domain.ErrCodeInternalError,
				"Failed to login",
				http.StatusInternalServerError,
			).WithDetails("original_error", err.Error())

			common.DomainError(c, internalErr)
		}
		return
	}

	common.Success(c, http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	}, "Login successful")
}

// Logout handles user logout requests
func (h *Handler) Logout(c *gin.Context) {
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

	// Get Bearer token from Authorization header
	token, err := h.tokenExtractor.GetBearerTokenFromHeader(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get Bearer token from header")
		}
		return
	}

	err = h.authService.Logout(userID, token)
	if err != nil {
		// Use domain error handling
		if domain.IsDomainError(err) {
			common.DomainError(c, domain.GetDomainError(err))
		} else {
			// Convert to domain error
			internalErr := domain.NewDomainError(
				domain.ErrCodeInternalError,
				"Failed to logout",
				http.StatusInternalServerError,
			).WithDetails("original_error", err.Error())

			common.DomainError(c, internalErr)
		}
		return
	}

	common.Success(c, http.StatusOK, gin.H{"message": "Logout successful"}, "Logout successful")
}

// Me returns the current user's information
func (h *Handler) Me(c *gin.Context) {
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

	// Get user by ID
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "User not found")
			return
		}
		common.InternalServerError(c, "Failed to get user")
		return
	}

	common.OK(c, user, "User retrieved successfully")
}

// GetUsers handles getting all users (admin only)
func (h *Handler) GetUsers(c *gin.Context) {
	// Get current user role from token for authorization
	userRole, err := h.tokenExtractor.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user role from token")
		}
		return
	}

	// Check if user has permission to list users (admin role only)
	if userRole != domain.AdminRoleType {
		common.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Insufficient permissions to list users",
			http.StatusForbidden,
		))
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get users from service
	users, total, err := h.userService.GetUsersWithFilters(domain.UserFilters{
		Limit: limit,
	})
	if err != nil {
		common.InternalServerError(c, "Failed to get users")
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	currentPage := (offset / limit) + 1

	common.OK(c, gin.H{
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
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "User not found")
			return
		}
		common.InternalServerError(c, "Failed to get user")
		return
	}

	common.OK(c, user, "User retrieved successfully")
}

// UpdateUser handles updating a user
func (h *Handler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	currentUserRole, err := h.tokenExtractor.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user role from token")
		}
		return
	}

	// Check if user can update this user (own profile or admin)
	if currentUserID != userID && currentUserRole != domain.AdminRoleType {
		common.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Insufficient permissions to update this user",
			http.StatusForbidden,
		))
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "User not found")
			return
		}
		common.InternalServerError(c, "Failed to get user")
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
		common.InternalServerError(c, "Failed to update user")
		return
	}

	common.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser handles deleting a user
func (h *Handler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	currentUserRole, err := h.tokenExtractor.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user role from token")
		}
		return
	}

	// Check if user can delete this user (admin only, cannot delete self)
	if currentUserRole != domain.AdminRoleType {
		common.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Only administrators can delete users",
			http.StatusForbidden,
		))
		return
	}

	// Prevent self-deletion
	if currentUserID == userID {
		common.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Cannot delete your own account",
			http.StatusForbidden,
		))
		return
	}

	// Check if user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "User not found")
			return
		}
		common.InternalServerError(c, "Failed to get user")
		return
	}

	// Delete user
	err = h.userService.DeleteUserByID(userID)
	if err != nil {
		common.InternalServerError(c, "Failed to delete user")
		return
	}

	common.OK(c, gin.H{"message": "User deleted successfully"}, "User deleted successfully")
}

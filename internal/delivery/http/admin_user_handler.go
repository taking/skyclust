package http

import (
	"strconv"
	"time"

	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminUserHandler handles admin user management operations
type AdminUserHandler struct {
	userService domain.UserService
	rbacService domain.RBACService
	logger      *logger.Logger
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(userService domain.UserService, rbacService domain.RBACService, logger *logger.Logger) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
		rbacService: rbacService,
		logger:      logger,
	}
}

// GetUsers retrieves all users with pagination and filtering
func (h *AdminUserHandler) GetUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get users with filters
	users, total, err := h.userService.GetUsersWithFilters(domain.UserFilters{
		Search: search,
		Role:   role,
		Status: status,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		h.logger.Errorf("Failed to get users: %v", err)
		InternalServerErrorResponse(c, "Failed to retrieve users")
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	// Get user roles for each user
	userList := make([]gin.H, len(users))
	for i, user := range users {
		roles, err := h.rbacService.GetUserRoles(user.ID)
		if err != nil {
			h.logger.Warnf("Failed to get roles for user %s: %v", user.ID, err)
			roles = []domain.Role{}
		}

		userList[i] = gin.H{
			"id":            user.ID,
			"username":      user.Username,
			"email":         user.Email,
			"is_active":     user.IsActive,
			"roles":         roles,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
			"oidc_provider": user.OIDCProvider,
			"oidc_subject":  user.OIDCSubject,
		}
	}

	OKResponse(c, gin.H{
		"users": userList,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}, "Users retrieved successfully")
}

// GetUser retrieves a specific user by ID
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			NotFoundResponse(c, "User not found")
			return
		}
		h.logger.Errorf("Failed to get user %s: %v", userID, err)
		InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	// Get user roles
	roles, err := h.rbacService.GetUserRoles(user.ID)
	if err != nil {
		h.logger.Warnf("Failed to get roles for user %s: %v", user.ID, err)
		roles = []domain.Role{}
	}

	// Get user permissions
	permissions, err := h.rbacService.GetUserEffectivePermissions(user.ID)
	if err != nil {
		h.logger.Warnf("Failed to get permissions for user %s: %v", user.ID, err)
		permissions = []domain.Permission{}
	}

	userResponse := gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"is_active":     user.IsActive,
		"roles":         roles,
		"permissions":   permissions,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"oidc_provider": user.OIDCProvider,
		"oidc_subject":  user.OIDCSubject,
	}

	OKResponse(c, gin.H{"user": userResponse}, "User retrieved successfully")
}

// UpdateUser updates a user's information
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid user ID")
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			NotFoundResponse(c, "User not found")
			return
		}
		h.logger.Errorf("Failed to get user %s: %v", userID, err)
		InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	// Update user fields
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.logger.Errorf("Failed to update user %s: %v", userID, err)
		InternalServerErrorResponse(c, "Failed to update user")
		return
	}

	// Get updated roles
	roles, err := h.rbacService.GetUserRoles(updatedUser.ID)
	if err != nil {
		h.logger.Warnf("Failed to get roles for user %s: %v", updatedUser.ID, err)
		roles = []domain.Role{}
	}

	userResponse := gin.H{
		"id":            updatedUser.ID,
		"username":      updatedUser.Username,
		"email":         updatedUser.Email,
		"is_active":     updatedUser.IsActive,
		"roles":         roles,
		"created_at":    updatedUser.CreatedAt,
		"updated_at":    updatedUser.UpdatedAt,
		"oidc_provider": updatedUser.OIDCProvider,
		"oidc_subject":  updatedUser.OIDCSubject,
	}

	OKResponse(c, gin.H{"user": userResponse}, "User updated successfully")
}

// DeleteUser deletes a user
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid user ID")
		return
	}

	// Check if user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			NotFoundResponse(c, "User not found")
			return
		}
		h.logger.Errorf("Failed to get user %s: %v", userID, err)
		InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	// Delete user
	err = h.userService.DeleteUserByID(userID)
	if err != nil {
		h.logger.Errorf("Failed to delete user %s: %v", userID, err)
		InternalServerErrorResponse(c, "Failed to delete user")
		return
	}

	OKResponse(c, gin.H{}, "User deleted successfully")
}

// AssignRole assigns a role to a user
func (h *AdminUserHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid user ID")
		return
	}

	var req struct {
		Role domain.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Validate role
	validRoles := []domain.Role{domain.AdminRoleType, domain.UserRoleType, domain.ViewerRoleType}
	validRole := false
	for _, role := range validRoles {
		if req.Role == role {
			validRole = true
			break
		}
	}
	if !validRole {
		BadRequestResponse(c, "Invalid role")
		return
	}

	// Assign role
	err = h.rbacService.AssignRole(userID, req.Role)
	if err != nil {
		h.logger.Errorf("Failed to assign role %s to user %s: %v", req.Role, userID, err)
		InternalServerErrorResponse(c, "Failed to assign role")
		return
	}

	OKResponse(c, gin.H{
		"user_id": userID,
		"role":    req.Role,
	}, "Role assigned successfully")
}

// RemoveRole removes a role from a user
func (h *AdminUserHandler) RemoveRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid user ID")
		return
	}

	var req struct {
		Role domain.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Remove role
	err = h.rbacService.RemoveRole(userID, req.Role)
	if err != nil {
		h.logger.Errorf("Failed to remove role %s from user %s: %v", req.Role, userID, err)
		InternalServerErrorResponse(c, "Failed to remove role")
		return
	}

	OKResponse(c, gin.H{
		"user_id": userID,
		"role":    req.Role,
	}, "Role removed successfully")
}

// GetUserStats retrieves user statistics
func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	// Get user statistics
	stats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Errorf("Failed to get user stats: %v", err)
		InternalServerErrorResponse(c, "Failed to retrieve user statistics")
		return
	}

	// Get role distribution
	roleStats, err := h.rbacService.GetRoleDistribution()
	if err != nil {
		h.logger.Warnf("Failed to get role distribution: %v", err)
		roleStats = make(map[domain.Role]int)
	}

	OKResponse(c, gin.H{
		"total_users":       stats.TotalUsers,
		"active_users":      stats.ActiveUsers,
		"inactive_users":    stats.InactiveUsers,
		"new_users_today":   stats.NewUsersToday,
		"role_distribution": roleStats,
		"last_updated":      time.Now().Format(time.RFC3339),
	}, "User statistics retrieved successfully")
}

package admin

import (
	"strconv"
	"time"

	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/utils"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles admin user management operations
type Handler struct {
	userService    domain.UserService
	rbacService    domain.RBACService
	logger         *logger.Logger
	tokenExtractor *utils.TokenExtractor
}

// NewHandler creates a new admin user handler
func NewHandler(userService domain.UserService, rbacService domain.RBACService, logger *logger.Logger) *Handler {
	return &Handler{
		userService:    userService,
		rbacService:    rbacService,
		logger:         logger,
		tokenExtractor: utils.NewTokenExtractor(),
	}
}

// GetUsers retrieves all users with pagination and filtering
func (h *Handler) GetUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build filters
	filters := domain.UserFilters{
		Limit:  limit,
		Search: search,
		Role:   role,
		Status: status,
	}

	// Get users from service
	users, total, err := h.userService.GetUsersWithFilters(filters)
	if err != nil {
		h.logger.Errorf("Failed to get users: %v", err)
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

// GetUser retrieves a specific user by ID
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
		h.logger.Errorf("Failed to get user: %v", err)
		common.InternalServerError(c, "Failed to get user")
		return
	}

	common.OK(c, user, "User retrieved successfully")
}

// UpdateUser updates a specific user
func (h *Handler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
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
		h.logger.Errorf("Failed to get user: %v", err)
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
		h.logger.Errorf("Failed to update user: %v", err)
		common.InternalServerError(c, "Failed to update user")
		return
	}

	common.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser deletes a specific user
func (h *Handler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	// Check if user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "User not found")
			return
		}
		h.logger.Errorf("Failed to get user: %v", err)
		common.InternalServerError(c, "Failed to get user")
		return
	}

	// Delete user
	err = h.userService.DeleteUserByID(userID)
	if err != nil {
		h.logger.Errorf("Failed to delete user: %v", err)
		common.InternalServerError(c, "Failed to delete user")
		return
	}

	common.OK(c, gin.H{"message": "User deleted successfully"}, "User deleted successfully")
}

// AssignRole assigns a role to a user
func (h *Handler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Validate role
	var roleType domain.Role
	switch req.Role {
	case "admin":
		roleType = domain.AdminRoleType
	case "user":
		roleType = domain.UserRoleType
	case "viewer":
		roleType = domain.ViewerRoleType
	default:
		common.BadRequest(c, "Invalid role")
		return
	}

	// Assign role
	err = h.rbacService.AssignRole(userID, roleType)
	if err != nil {
		h.logger.Errorf("Failed to assign role: %v", err)
		common.InternalServerError(c, "Failed to assign role")
		return
	}

	common.OK(c, gin.H{"message": "Role assigned successfully"}, "Role assigned successfully")
}

// RemoveRole removes a role from a user
func (h *Handler) RemoveRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Validate role
	var roleType domain.Role
	switch req.Role {
	case "admin":
		roleType = domain.AdminRoleType
	case "user":
		roleType = domain.UserRoleType
	case "viewer":
		roleType = domain.ViewerRoleType
	default:
		common.BadRequest(c, "Invalid role")
		return
	}

	// Remove role
	err = h.rbacService.RemoveRole(userID, roleType)
	if err != nil {
		h.logger.Errorf("Failed to remove role: %v", err)
		common.InternalServerError(c, "Failed to remove role")
		return
	}

	common.OK(c, gin.H{"message": "Role removed successfully"}, "Role removed successfully")
}

// GetUserStats retrieves user statistics
func (h *Handler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Errorf("Failed to get user stats: %v", err)
		common.InternalServerError(c, "Failed to get user stats")
		return
	}

	// Get role distribution
	roleStats, err := h.rbacService.GetRoleDistribution()
	if err != nil {
		h.logger.Errorf("Failed to get role distribution: %v", err)
		// Continue without role stats
		roleStats = make(map[domain.Role]int)
	}

	common.OK(c, gin.H{
		"total_users":       stats.TotalUsers,
		"active_users":      stats.ActiveUsers,
		"inactive_users":    stats.InactiveUsers,
		"new_users_today":   stats.NewUsersToday,
		"role_distribution": roleStats,
		"last_updated":      time.Now().Format(time.RFC3339),
	}, "User statistics retrieved successfully")
}

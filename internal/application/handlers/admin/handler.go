package admin

import (
	"net/http"
	"strconv"
	"time"

	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles admin user management operations using improved patterns
type Handler struct {
	*handlers.BaseHandler
	userService       domain.UserService
	rbacService       domain.RBACService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new admin user handler
func NewHandler(userService domain.UserService, rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("admin"),
		userService:       userService,
		rbacService:       rbacService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// GetUsers retrieves all users with pagination and filtering using decorator pattern
func (h *Handler) GetUsers(c *gin.Context) {
	handler := h.Compose(
		h.getUsersHandler(),
		h.StandardCRUDDecorators("get_users")...,
	)

	handler(c)
}

// getUsersHandler is the core business logic for getting users
func (h *Handler) getUsersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		if !h.checkAdminPermission(c) {
			return
		}

		h.logAdminUsersRequest(c)

		filters := h.parseUserFilters(c)

		users, total, err := h.userService.GetUsersWithFilters(filters)
		if err != nil {
			h.HandleError(c, err, "get_users")
			return
		}

		// Use standardized pagination metadata
		paginationMeta := h.CalculatePaginationMeta(total, filters.Page, filters.Limit)

		h.OK(c, gin.H{
			"users":      users,
			"pagination": paginationMeta,
		}, "Users retrieved successfully")
	}
}

// GetUser retrieves a specific user by ID
func (h *Handler) GetUser(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "get_user")
		return
	}

	h.OK(c, user, "User retrieved successfully")
}

// UpdateUser updates a specific user
func (h *Handler) UpdateUser(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user")
		return
	}

	var req domain.UpdateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_user")
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
	if req.Password != nil && *req.Password != "" {
		// Hash new password
		hashedPassword, err := h.userService.HashPassword(*req.Password)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500), "update_user")
			return
		}
		user.PasswordHash = hashedPassword
	}
	if req.IsActive != nil {
		user.Active = *req.IsActive
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	h.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser deletes a specific user
func (h *Handler) DeleteUser(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user")
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

// AssignRole assigns a role to a user
func (h *Handler) AssignRole(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "assign_role")
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
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400), "assign_role")
		return
	}

	// Assign role
	err = h.rbacService.AssignRole(userID, roleType)
	if err != nil {
		h.HandleError(c, err, "assign_role")
		return
	}

	h.OK(c, gin.H{"message": "Role assigned successfully"}, "Role assigned successfully")
}

// RemoveRole removes a role from a user
func (h *Handler) RemoveRole(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "assign_role")
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
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400), "assign_role")
		return
	}

	// Remove role
	err = h.rbacService.RemoveRole(userID, roleType)
	if err != nil {
		h.HandleError(c, err, "remove_role")
		return
	}

	h.OK(c, gin.H{"message": "Role removed successfully"}, "Role removed successfully")
}

// GetUserStats retrieves user statistics
func (h *Handler) GetUserStats(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	stats, err := h.userService.GetUserStats()
	if err != nil {
		h.HandleError(c, err, "get_user_stats")
		return
	}

	// Get role distribution
	roleStats, err := h.rbacService.GetRoleDistribution()
	if err != nil {
		h.LogError(c, err, "Failed to get role distribution")
		// Continue without role stats
		roleStats = make(map[domain.Role]int)
	}

	h.OK(c, gin.H{
		"total_users":       stats.TotalUsers,
		"active_users":      stats.ActiveUsers,
		"inactive_users":    stats.InactiveUsers,
		"new_users_today":   stats.NewUsersToday,
		"role_distribution": roleStats,
		"last_updated":      time.Now().Format(time.RFC3339),
	}, "User statistics retrieved successfully")
}

// checkAdminPermission checks if the current user has admin permission
func (h *Handler) checkAdminPermission(c *gin.Context) bool {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			h.HandleError(c, domainErr, "check_admin_permission")
		} else {
			h.HandleError(c, err, "check_admin_permission")
		}
		return false
	}

	// Check if user has admin role
	if userRole != domain.AdminRoleType {
		h.HandleError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Insufficient permissions - admin role required",
			http.StatusForbidden,
		), "check_admin_permission")
		return false
	}

	return true
}

// Helper methods for better readability

func (h *Handler) parseUserFilters(c *gin.Context) domain.UserFilters {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	return domain.UserFilters{
		Limit:  limit,
		Page:   page,
		Search: search,
		Role:   role,
		Status: status,
	}
}


// Logging helper methods

func (h *Handler) logAdminUsersRequest(c *gin.Context) {
	h.LogBusinessEvent(c, "admin_users_requested", "", "", map[string]interface{}{
		"operation": "get_users",
	})
}

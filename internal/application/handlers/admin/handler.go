package admin

import (
	"net/http"
	"time"

	"skyclust/internal/application/handlers/auth"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: 관리자 사용자 관리 작업을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	userService       domain.UserService
	rbacService       domain.RBACService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler: 새로운 관리자 사용자 핸들러를 생성합니다
func NewHandler(userService domain.UserService, rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("admin"),
		userService:       userService,
		rbacService:       rbacService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// GetUsers: 페이지네이션과 필터링을 포함한 모든 사용자를 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetUsers(c *gin.Context) {
	handler := h.Compose(
		h.getUsersHandler(),
		h.StandardCRUDDecorators("get_users")...,
	)

	handler(c)
}

// getUsersHandler: 사용자 조회의 핵심 비즈니스 로직을 처리합니다
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

		// Use standardized paginated response (direct array: data[])
		h.BuildPaginatedResponse(c, users, filters.Page, filters.Limit, total, "Users retrieved successfully")
	}
}

// GetUser: ID로 특정 사용자를 조회합니다
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

// UpdateUser: 특정 사용자를 업데이트합니다
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

	// Handler layer DTO 사용 (auth 패키지의 타입 재사용)
	var handlerReq auth.UpdateUserRequest
	if err := h.ValidateRequest(c, &handlerReq); err != nil {
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
	if handlerReq.Username != nil && *handlerReq.Username != "" {
		user.Username = *handlerReq.Username
	}
	if handlerReq.Email != nil && *handlerReq.Email != "" {
		user.Email = *handlerReq.Email
	}
	if handlerReq.Password != nil && *handlerReq.Password != "" {
		// Hash new password
		hashedPassword, err := h.userService.HashPassword(*handlerReq.Password)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500), "update_user")
			return
		}
		user.PasswordHash = hashedPassword
	}
	if handlerReq.IsActive != nil {
		user.Active = *handlerReq.IsActive
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	h.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser: 특정 사용자를 삭제합니다
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

// GetUserStats: 사용자 통계를 조회합니다
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

// checkAdminPermission: 현재 사용자가 관리자 권한을 가지고 있는지 확인합니다
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

// 헬퍼 메서드들

// parseUserFilters: 쿼리 파라미터로부터 사용자 필터를 파싱합니다
func (h *Handler) parseUserFilters(c *gin.Context) domain.UserFilters {
	// Use standardized query params parsing
	params := h.ParseStandardQueryParams(c)

	// Extract specific filters
	role := c.Query("role")
	status := c.Query("status")

	// Also check filter_ prefix for role and status
	if role == "" {
		if roleFilter, ok := params.Filters["role"].(string); ok {
			role = roleFilter
		}
	}
	if status == "" {
		if statusFilter, ok := params.Filters["status"].(string); ok {
			status = statusFilter
		}
	}

	return domain.UserFilters{
		Limit:  params.Limit,
		Page:   params.Page,
		Search: params.Search,
		Role:   role,
		Status: status,
	}
}

// 로깅 헬퍼 메서드들

// logAdminUsersRequest: 관리자 사용자 조회 요청 로그를 기록합니다
func (h *Handler) logAdminUsersRequest(c *gin.Context) {
	h.LogBusinessEvent(c, "admin_users_requested", "", "", map[string]interface{}{
		"operation": "get_users",
	})
}

package rbac

import (
	"net/http"

	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: RBAC (역할 기반 접근 제어) 작업을 처리하는 핸들러
type Handler struct {
	rbacService domain.RBACService
	*handlers.BaseHandler
}

// NewHandler: 새로운 RBAC 핸들러를 생성합니다
func NewHandler(rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("rbac"),
		rbacService: rbacService,
	}
}

// AssignRole: 사용자에게 역할을 할당합니다
func (h *Handler) AssignRole(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "assign_role")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "assign_role")
		return
	}

	roleType, err := h.parseRole(req.Role)
	if err != nil {
		h.HandleError(c, err, "assign_role")
		return
	}

	err = h.rbacService.AssignRole(userID, roleType)
	if err != nil {
		h.HandleError(c, err, "assign_role")
		return
	}

	h.OK(c, gin.H{"message": "Role assigned successfully"}, "Role assigned successfully")
}

// RemoveRole: 사용자로부터 역할을 제거합니다
func (h *Handler) RemoveRole(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "remove_role")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "remove_role")
		return
	}

	roleType, err := h.parseRole(req.Role)
	if err != nil {
		h.HandleError(c, err, "remove_role")
		return
	}

	err = h.rbacService.RemoveRole(userID, roleType)
	if err != nil {
		h.HandleError(c, err, "remove_role")
		return
	}

	h.OK(c, gin.H{"message": "Role removed successfully"}, "Role removed successfully")
}

// GrantPermission: 역할에 권한을 부여합니다
func (h *Handler) GrantPermission(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	var req struct {
		Role       string `json:"role" binding:"required"`
		Permission string `json:"permission" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "grant_permission")
		return
	}

	roleType, err := h.parseRole(req.Role)
	if err != nil {
		h.HandleError(c, err, "grant_permission")
		return
	}

	permission := domain.Permission(req.Permission)

	err = h.rbacService.GrantPermission(roleType, permission)
	if err != nil {
		h.HandleError(c, err, "grant_permission")
		return
	}

	h.OK(c, gin.H{"message": "Permission granted successfully"}, "Permission granted successfully")
}

// RevokePermission: 역할로부터 권한을 회수합니다
func (h *Handler) RevokePermission(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	var req struct {
		Role       string `json:"role" binding:"required"`
		Permission string `json:"permission" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "revoke_permission")
		return
	}

	roleType, err := h.parseRole(req.Role)
	if err != nil {
		h.HandleError(c, err, "revoke_permission")
		return
	}

	permission := domain.Permission(req.Permission)

	err = h.rbacService.RevokePermission(roleType, permission)
	if err != nil {
		h.HandleError(c, err, "revoke_permission")
		return
	}

	h.OK(c, gin.H{"message": "Permission revoked successfully"}, "Permission revoked successfully")
}

// GetRolePermissions: 역할의 모든 권한을 반환합니다
func (h *Handler) GetRolePermissions(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	roleStr := c.Param("role")
	roleType, err := h.parseRole(roleStr)
	if err != nil {
		h.HandleError(c, err, "get_role_permissions")
		return
	}

	permissions, err := h.rbacService.GetRolePermissions(roleType)
	if err != nil {
		h.HandleError(c, err, "get_role_permissions")
		return
	}

	h.OK(c, gin.H{
		"role":        roleStr,
		"permissions": permissions,
	}, "Role permissions retrieved successfully")
}

// CheckUserPermission: 사용자가 특정 권한을 가지고 있는지 확인합니다
func (h *Handler) CheckUserPermission(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "check_user_permission")
		return
	}

	permissionStr := c.Query("permission")
	if permissionStr == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Permission parameter required", 400), "check_user_permission")
		return
	}

	permission := domain.Permission(permissionStr)

	hasPermission, err := h.rbacService.CheckPermission(userID, permission)
	if err != nil {
		h.HandleError(c, err, "check_user_permission")
		return
	}

	h.OK(c, gin.H{
		"user_id":        userIDStr,
		"permission":     permissionStr,
		"has_permission": hasPermission,
	}, "Permission check completed")
}

// GetUserEffectivePermissions: 사용자의 모든 유효 권한을 반환합니다
func (h *Handler) GetUserEffectivePermissions(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user_effective_permissions")
		return
	}

	permissions, err := h.rbacService.GetUserEffectivePermissions(userID)
	if err != nil {
		h.HandleError(c, err, "get_user_effective_permissions")
		return
	}

	h.OK(c, gin.H{
		"user_id":     userIDStr,
		"permissions": permissions,
	}, "User effective permissions retrieved successfully")
}

// GetUserRoles: 사용자의 모든 역할을 반환합니다
func (h *Handler) GetUserRoles(c *gin.Context) {
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user_roles")
		return
	}

	roles, err := h.rbacService.GetUserRoles(userID)
	if err != nil {
		h.HandleError(c, err, "get_user_roles")
		return
	}

	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = h.roleToString(role)
	}

	h.OK(c, gin.H{
		"user_id": userIDStr,
		"roles":    roleStrings,
	}, "User roles retrieved successfully")
}

// 헬퍼 메서드들

// parseRole: 역할 문자열을 파싱하여 해당하는 domain.Role을 반환합니다
func (h *Handler) parseRole(roleStr string) (domain.Role, error) {
	switch roleStr {
	case "admin":
		return domain.AdminRoleType, nil
	case "user":
		return domain.UserRoleType, nil
	case "viewer":
		return domain.ViewerRoleType, nil
	default:
		return domain.Role(""), domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400)
	}
}

// roleToString: domain.Role을 문자열 표현으로 변환합니다
func (h *Handler) roleToString(role domain.Role) string {
	switch role {
	case domain.AdminRoleType:
		return "admin"
	case domain.UserRoleType:
		return "user"
	case domain.ViewerRoleType:
		return "viewer"
	default:
		return string(role)
	}
}

// checkAdminPermission: 현재 사용자가 관리자 권한을 가지고 있는지 확인합니다
func (h *Handler) checkAdminPermission(c *gin.Context) bool {
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "check_admin_permission")
		return false
	}

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


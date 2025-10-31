package admin

import (
	"net/http"

	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionHandler handles permission management operations
type PermissionHandler struct {
	rbacService domain.RBACService
	*handlers.BaseHandler
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(rbacService domain.RBACService) *PermissionHandler {
	return &PermissionHandler{
		BaseHandler: handlers.NewBaseHandler("admin-permission"),
		rbacService: rbacService,
	}
}

// GrantPermission grants a permission to a role
func (h *PermissionHandler) GrantPermission(c *gin.Context) {
	// Check admin permission
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
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400), "grant_permission")
		return
	}

	// Validate permission
	permission := domain.Permission(req.Permission)

	// Grant permission
	err := h.rbacService.GrantPermission(roleType, permission)
	if err != nil {
		h.HandleError(c, err, "grant_permission")
		return
	}

	h.OK(c, gin.H{"message": "Permission granted successfully"}, "Permission granted successfully")
}

// RevokePermission revokes a permission from a role
func (h *PermissionHandler) RevokePermission(c *gin.Context) {
	// Check admin permission
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
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400), "revoke_permission")
		return
	}

	// Validate permission
	permission := domain.Permission(req.Permission)

	// Revoke permission
	err := h.rbacService.RevokePermission(roleType, permission)
	if err != nil {
		h.HandleError(c, err, "revoke_permission")
		return
	}

	h.OK(c, gin.H{"message": "Permission revoked successfully"}, "Permission revoked successfully")
}

// GetRolePermissions returns all permissions for a role
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}

	roleStr := c.Param("role")
	var roleType domain.Role
	switch roleStr {
	case "admin":
		roleType = domain.AdminRoleType
	case "user":
		roleType = domain.UserRoleType
	case "viewer":
		roleType = domain.ViewerRoleType
	default:
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid role", 400), "get_role_permissions")
		return
	}

	// Get role permissions
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

// CheckUserPermission checks if a user has a specific permission
func (h *PermissionHandler) CheckUserPermission(c *gin.Context) {
	// Check admin permission
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

	// Check permission
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

// GetUserEffectivePermissions returns all effective permissions for a user
func (h *PermissionHandler) GetUserEffectivePermissions(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "get_user_effective_permissions")
		return
	}

	// Get effective permissions
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

// checkAdminPermission checks if the current user has admin permission
func (h *PermissionHandler) checkAdminPermission(c *gin.Context) bool {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "check_admin_permission")
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

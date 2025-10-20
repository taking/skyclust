package admin

import (
	"net/http"

	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionHandler handles permission management operations
type PermissionHandler struct {
	rbacService    domain.RBACService
	tokenExtractor *utils.TokenExtractor
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(rbacService domain.RBACService) *PermissionHandler {
	return &PermissionHandler{
		rbacService:    rbacService,
		tokenExtractor: utils.NewTokenExtractor(),
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

	// Validate permission
	permission := domain.Permission(req.Permission)

	// Grant permission
	err := h.rbacService.GrantPermission(roleType, permission)
	if err != nil {
		common.InternalServerError(c, "Failed to grant permission")
		return
	}

	common.OK(c, gin.H{"message": "Permission granted successfully"}, "Permission granted successfully")
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

	// Validate permission
	permission := domain.Permission(req.Permission)

	// Revoke permission
	err := h.rbacService.RevokePermission(roleType, permission)
	if err != nil {
		common.InternalServerError(c, "Failed to revoke permission")
		return
	}

	common.OK(c, gin.H{"message": "Permission revoked successfully"}, "Permission revoked successfully")
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
		common.BadRequest(c, "Invalid role")
		return
	}

	// Get role permissions
	permissions, err := h.rbacService.GetRolePermissions(roleType)
	if err != nil {
		common.InternalServerError(c, "Failed to get role permissions")
		return
	}

	common.OK(c, gin.H{
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
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	permissionStr := c.Query("permission")
	if permissionStr == "" {
		common.BadRequest(c, "Permission parameter required")
		return
	}

	permission := domain.Permission(permissionStr)

	// Check permission
	hasPermission, err := h.rbacService.CheckPermission(userID, permission)
	if err != nil {
		common.InternalServerError(c, "Failed to check permission")
		return
	}

	common.OK(c, gin.H{
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
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	// Get effective permissions
	permissions, err := h.rbacService.GetUserEffectivePermissions(userID)
	if err != nil {
		common.InternalServerError(c, "Failed to get user effective permissions")
		return
	}

	common.OK(c, gin.H{
		"user_id":     userIDStr,
		"permissions": permissions,
	}, "User effective permissions retrieved successfully")
}

// checkAdminPermission checks if the current user has admin permission
func (h *PermissionHandler) checkAdminPermission(c *gin.Context) bool {
	// Get current user role from token for authorization
	userRole, err := h.tokenExtractor.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user role from token")
		}
		return false
	}

	// Check if user has admin role
	if userRole != domain.AdminRoleType {
		common.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Insufficient permissions - admin role required",
			http.StatusForbidden,
		))
		return false
	}

	return true
}

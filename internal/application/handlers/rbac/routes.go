package rbac

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up RBAC management routes
func SetupRoutes(router *gin.RouterGroup, rbacService domain.RBACService) {
	rbacHandler := NewHandler(rbacService)

	// Role management routes
	router.POST("/users/:user_id/roles", rbacHandler.AssignRole)   // POST /api/v1/admin/rbac/users/:user_id/roles
	router.DELETE("/users/:user_id/roles", rbacHandler.RemoveRole) // DELETE /api/v1/admin/rbac/users/:user_id/roles
	router.GET("/users/:user_id/roles", rbacHandler.GetUserRoles)  // GET /api/v1/admin/rbac/users/:user_id/roles

	// Permission management routes (RESTful)
	// POST /api/v1/admin/rbac/roles/:role/permissions - Grant permission to role
	router.POST("/roles/:role/permissions", rbacHandler.GrantPermission)
	// DELETE /api/v1/admin/rbac/roles/:role/permissions/:permission - Revoke permission from role
	router.DELETE("/roles/:role/permissions/:permission", rbacHandler.RevokePermission)
	// GET /api/v1/admin/rbac/roles/:role/permissions - Get all permissions for a role
	router.GET("/roles/:role/permissions", rbacHandler.GetRolePermissions)
	// GET /api/v1/admin/rbac/users/:user_id/permissions/check - Check if user has specific permission
	router.GET("/users/:user_id/permissions/check", rbacHandler.CheckUserPermission)
	// GET /api/v1/admin/rbac/users/:user_id/permissions/effective - Get all effective permissions for a user
	router.GET("/users/:user_id/permissions/effective", rbacHandler.GetUserEffectivePermissions)
}

package rbac

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up RBAC management routes
func SetupRoutes(router *gin.RouterGroup, rbacService domain.RBACService) {
	rbacHandler := NewHandler(rbacService)

	// Role management routes
	router.POST("/users/:user_id/roles", rbacHandler.AssignRole)       // POST /api/v1/admin/rbac/users/:user_id/roles
	router.DELETE("/users/:user_id/roles", rbacHandler.RemoveRole)   // DELETE /api/v1/admin/rbac/users/:user_id/roles
	router.GET("/users/:user_id/roles", rbacHandler.GetUserRoles)     // GET /api/v1/admin/rbac/users/:user_id/roles

	// Permission management routes
	router.POST("/permissions/grant", rbacHandler.GrantPermission)                               // POST /api/v1/admin/rbac/permissions/grant
	router.POST("/permissions/revoke", rbacHandler.RevokePermission)                             // POST /api/v1/admin/rbac/permissions/revoke
	router.GET("/permissions/roles/:role", rbacHandler.GetRolePermissions)                       // GET /api/v1/admin/rbac/permissions/roles/:role
	router.GET("/permissions/users/:user_id/check", rbacHandler.CheckUserPermission)             // GET /api/v1/admin/rbac/permissions/users/:user_id/check
	router.GET("/permissions/users/:user_id/effective", rbacHandler.GetUserEffectivePermissions) // GET /api/v1/admin/rbac/permissions/users/:user_id/effective
}


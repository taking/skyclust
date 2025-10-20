package admin

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up admin user management routes
func SetupUserRoutes(router *gin.RouterGroup, userService domain.UserService, rbacService domain.RBACService, logger *logger.Logger) {
	adminHandler := NewHandler(userService, rbacService, logger)

	// User management
	router.GET("/", adminHandler.GetUsers)          // GET /api/v1/admin/users
	router.GET("/stats", adminHandler.GetUserStats) // GET /api/v1/admin/users/stats
	router.GET("/:id", adminHandler.GetUser)        // GET /api/v1/admin/users/:id
	router.PUT("/:id", adminHandler.UpdateUser)     // PUT /api/v1/admin/users/:id
	router.DELETE("/:id", adminHandler.DeleteUser)  // DELETE /api/v1/admin/users/:id

	// Role management
	router.POST("/:id/roles", adminHandler.AssignRole)   // POST /api/v1/admin/users/:id/roles
	router.DELETE("/:id/roles", adminHandler.RemoveRole) // DELETE /api/v1/admin/users/:id/roles
}

// SetupPermissionRoutes sets up admin permission management routes
func SetupPermissionRoutes(router *gin.RouterGroup, rbacService domain.RBACService) {
	permissionHandler := NewPermissionHandler(rbacService)

	// Permission management
	router.POST("/grant", permissionHandler.GrantPermission)                               // POST /api/v1/admin/permissions/grant
	router.POST("/revoke", permissionHandler.RevokePermission)                             // POST /api/v1/admin/permissions/revoke
	router.GET("/roles/:role", permissionHandler.GetRolePermissions)                       // GET /api/v1/admin/permissions/roles/:role
	router.GET("/users/:user_id/check", permissionHandler.CheckUserPermission)             // GET /api/v1/admin/permissions/users/:user_id/check
	router.GET("/users/:user_id/effective", permissionHandler.GetUserEffectivePermissions) // GET /api/v1/admin/permissions/users/:user_id/effective
}

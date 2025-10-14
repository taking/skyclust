package routes

import (
	"skyclust/internal/delivery/http"

	"github.com/gin-gonic/gin"
)

// SetupAdminUserRoutes sets up admin user management routes
func SetupAdminUserRoutes(router *gin.RouterGroup, adminUserHandler *http.AdminUserHandler) {
	// User management
	router.GET("/", adminUserHandler.GetUsers)          // GET /api/v1/admin/users
	router.GET("/stats", adminUserHandler.GetUserStats) // GET /api/v1/admin/users/stats
	router.GET("/:id", adminUserHandler.GetUser)        // GET /api/v1/admin/users/:id
	router.PUT("/:id", adminUserHandler.UpdateUser)     // PUT /api/v1/admin/users/:id
	router.DELETE("/:id", adminUserHandler.DeleteUser)  // DELETE /api/v1/admin/users/:id

	// Role management
	router.POST("/:id/roles", adminUserHandler.AssignRole)   // POST /api/v1/admin/users/:id/roles
	router.DELETE("/:id/roles", adminUserHandler.RemoveRole) // DELETE /api/v1/admin/users/:id/roles
}

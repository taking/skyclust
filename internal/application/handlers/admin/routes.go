package admin

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up admin user management routes
func SetupUserRoutes(router *gin.RouterGroup, userService domain.UserService, rbacService domain.RBACService, logger *logger.Logger) {
	adminHandler := NewHandler(userService, rbacService)

	// User management
	router.GET("/", adminHandler.GetUsers)          // GET /api/v1/admin/users
	router.GET("/stats", adminHandler.GetUserStats) // GET /api/v1/admin/users/stats
	router.GET("/:id", adminHandler.GetUser)        // GET /api/v1/admin/users/:id
	router.PUT("/:id", adminHandler.UpdateUser)     // PUT /api/v1/admin/users/:id
	router.DELETE("/:id", adminHandler.DeleteUser)  // DELETE /api/v1/admin/users/:id
}

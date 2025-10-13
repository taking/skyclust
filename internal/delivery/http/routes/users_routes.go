package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupUsersRoutes sets up user management routes (RESTful)
func SetupUsersRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService) {
	authHandler := httpDelivery.NewAuthHandler(authService, userService)

	// User management routes
	router.POST("/users", authHandler.Register)         // Create user
	router.GET("/users", authHandler.GetUsers)          // List users
	router.GET("/users/:id", authHandler.GetUser)       // Get specific user
	router.PUT("/users/:id", authHandler.UpdateUser)    // Update user
	router.DELETE("/users/:id", authHandler.DeleteUser) // Delete user
}

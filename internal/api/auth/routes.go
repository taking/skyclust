package auth

import (
	"skyclust/internal/domain"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up authentication routes
func SetupRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) {
	authHandler := NewHandlerWithLogout(authService, userService, logoutService)

	// Public authentication routes (no authentication required)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)

	// Protected authentication routes (authentication required)
	router.GET("/me", authHandler.Me)
}

// SetupUserRoutes sets up user management routes (RESTful)
func SetupUserRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService) {
	authHandler := NewHandler(authService, userService)

	// User management routes
	router.POST("/users", authHandler.Register)         // Create user
	router.GET("/users", authHandler.GetUsers)          // List users
	router.GET("/users/:id", authHandler.GetUser)       // Get specific user
	router.PUT("/users/:id", authHandler.UpdateUser)    // Update user
	router.DELETE("/users/:id", authHandler.DeleteUser) // Delete user
}

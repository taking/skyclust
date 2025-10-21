package auth

import (
	"skyclust/internal/application/services"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up authentication routes
func SetupRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService, logoutService *service.LogoutService) {
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
	router.POST("", authHandler.Register)         // Create user
	router.GET("", authHandler.GetUsers)          // List users
	router.GET("/:id", authHandler.GetUser)       // Get specific user
	router.PUT("/:id", authHandler.UpdateUser)    // Update user
	router.DELETE("/:id", authHandler.DeleteUser) // Delete user
}

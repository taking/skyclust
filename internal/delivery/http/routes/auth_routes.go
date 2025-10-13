package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupPublicAuthRoutes sets up public authentication routes
func SetupPublicAuthRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) {
	authHandler := httpDelivery.NewAuthHandlerWithLogout(authService, userService, logoutService)

	// Public authentication routes (no authentication required)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)
}

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) {
	authHandler := httpDelivery.NewAuthHandlerWithLogout(authService, userService, logoutService)

	// Protected authentication routes (authentication required)
	router.GET("/me", authHandler.Me)
}

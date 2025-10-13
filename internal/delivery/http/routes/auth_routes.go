package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.RouterGroup, authService domain.AuthService, userService domain.UserService, logoutService *usecase.LogoutService) {
	authHandler := httpDelivery.NewAuthHandlerWithLogout(authService, userService, logoutService)

	// Basic authentication routes
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)
	router.GET("/me", authHandler.Me)
}

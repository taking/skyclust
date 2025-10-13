package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupLogoutRoutes sets up logout routes
func SetupLogoutRoutes(router *gin.RouterGroup, logoutService *usecase.LogoutService) {
	logoutHandler := httpDelivery.NewLogoutHandler(logoutService)

	// Unified logout endpoint
	router.POST("/logout", logoutHandler.UnifiedLogout)

	// Batch logout for multiple devices
	router.POST("/logout/batch", logoutHandler.BatchLogout)

	// Logout statistics
	router.GET("/logout/stats", logoutHandler.GetLogoutStats)

	// Cleanup expired tokens
	router.POST("/logout/cleanup", logoutHandler.CleanupExpiredTokens)
}


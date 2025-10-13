package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupSessionsManagementRoutes sets up session management routes (RESTful)
func SetupSessionsManagementRoutes(router *gin.RouterGroup, logoutService *usecase.LogoutService) {
	logoutHandler := httpDelivery.NewLogoutHandler(logoutService)

	// Session management routes
	router.DELETE("/sessions", logoutHandler.UnifiedLogout)                // Delete session (logout)
	router.DELETE("/sessions/batch", logoutHandler.BatchLogout)            // Delete multiple sessions
	router.GET("/sessions/statistics", logoutHandler.GetLogoutStats)       // Get session statistics
	router.DELETE("/sessions/expired", logoutHandler.CleanupExpiredTokens) // Cleanup expired sessions
}


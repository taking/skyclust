package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupOIDCProvidersRoutes sets up OIDC provider routes (RESTful)
func SetupOIDCProvidersRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := httpDelivery.NewOIDCHandler(oidcService)

	// OIDC provider routes
	router.GET("/oidc/providers", oidcHandler.GetProviders)                       // List OIDC providers
	router.GET("/oidc/providers/:provider/auth-urls", oidcHandler.GetAuthURL)     // Get auth URL
	router.GET("/oidc/providers/:provider/callbacks", oidcHandler.Callback)       // Handle callback
	router.GET("/oidc/providers/:provider/logout-urls", oidcHandler.GetLogoutURL) // Get logout URL
	router.POST("/oidc/sessions", oidcHandler.Login)                              // Create OIDC session
	router.DELETE("/oidc/sessions", oidcHandler.Logout)                           // Delete OIDC session
}


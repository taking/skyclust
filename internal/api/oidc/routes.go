package oidc

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up OIDC authentication routes
func SetupRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := NewHandler(oidcService)

	router.GET("/:provider", oidcHandler.GetAuthURL)
	router.GET("/:provider/callback", oidcHandler.Callback)
	router.POST("/login", oidcHandler.Login)
	router.POST("/logout", oidcHandler.Logout)
	router.GET("/logout/url", oidcHandler.GetLogoutURL)
}

// SetupProviderRoutes sets up OIDC provider routes (RESTful)
func SetupProviderRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := NewHandler(oidcService)

	// OIDC provider routes
	router.GET("/oidc/providers", oidcHandler.GetProviders)                       // List OIDC providers
	router.GET("/oidc/providers/:provider/auth-urls", oidcHandler.GetAuthURL)     // Get auth URL
	router.GET("/oidc/providers/:provider/callbacks", oidcHandler.Callback)       // Handle callback
	router.GET("/oidc/providers/:provider/logout-urls", oidcHandler.GetLogoutURL) // Get logout URL
	router.POST("/oidc/sessions", oidcHandler.Login)                              // Create OIDC session
	router.DELETE("/oidc/sessions", oidcHandler.Logout)                           // Delete OIDC session
}

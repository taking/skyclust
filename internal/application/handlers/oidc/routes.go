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

// SetupProviderRoutes sets up public OIDC provider routes (list available providers)
func SetupProviderRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := NewHandler(oidcService)

	// Public OIDC provider routes (list available providers)
	router.GET("", oidcHandler.GetProviders) // GET /api/v1/oidc-providers
}

// SetupUserProviderRoutes sets up user OIDC provider management routes (protected, needs auth)
// Note: router is already scoped to /api/v1/oidc, so paths are relative to that
func SetupUserProviderRoutes(router *gin.RouterGroup, oidcService domain.OIDCService, providerRepo domain.OIDCProviderRepository) {
	oidcHandler := NewHandler(oidcService)
	oidcHandler.SetOIDCProviderRepository(providerRepo)

	// User OIDC provider management routes
	// Final paths: /api/v1/oidc/providers
	providerRoutes := router.Group("/providers")
	{
		providerRoutes.POST("", oidcHandler.CreateProvider)       // POST /api/v1/oidc/providers
		providerRoutes.GET("", oidcHandler.GetUserProviders)      // GET /api/v1/oidc/providers
		providerRoutes.GET("/:id", oidcHandler.GetProvider)       // GET /api/v1/oidc/providers/:id
		providerRoutes.PUT("/:id", oidcHandler.UpdateProvider)    // PUT /api/v1/oidc/providers/:id
		providerRoutes.DELETE("/:id", oidcHandler.DeleteProvider) // DELETE /api/v1/oidc/providers/:id
	}
}

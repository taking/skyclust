package oidc

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up OIDC authentication routes
func SetupRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := NewHandler(oidcService)

	// Public OIDC routes
	router.GET("/:provider", oidcHandler.GetAuthURL)              // GET /api/v1/auth/oidc/:provider
	router.GET("/:provider/callback", oidcHandler.Callback)       // GET /api/v1/auth/oidc/:provider/callback
	router.GET("/:provider/logout-url", oidcHandler.GetLogoutURL) // GET /api/v1/auth/oidc/:provider/logout-url

	// Session management (RESTful)
	sessionsGroup := router.Group("/sessions")
	{
		sessionsGroup.POST("", oidcHandler.CreateSession)      // POST /api/v1/auth/oidc/sessions (create session)
		sessionsGroup.DELETE("/me", oidcHandler.DeleteSession) // DELETE /api/v1/auth/oidc/sessions/me (delete current session)
	}
}

// SetupProviderRoutes sets up public OIDC provider routes (list available provider types)
// router is scoped to /api/v1/oidc/providers
func SetupProviderRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := NewHandler(oidcService)

	// Public OIDC provider type routes
	// GET /api/v1/oidc/providers/types - Get available provider types
	router.GET("/types", oidcHandler.GetProviders)
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

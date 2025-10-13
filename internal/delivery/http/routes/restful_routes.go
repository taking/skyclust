package routes

import (
	"skyclust/internal/domain"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SetupRESTfulRoutes sets up all RESTful API routes
func SetupRESTfulRoutes(router *gin.Engine,
	authService domain.AuthService,
	userService domain.UserService,
	oidcService domain.OIDCService,
	logoutService *usecase.LogoutService) {

	// API v1 group
	v1 := router.Group("/api/v1")

	// Authentication routes (public)
	authGroup := v1.Group("/auth")
	SetupPublicAuthRoutes(authGroup, authService, userService, logoutService)

	// OIDC routes (public)
	oidcGroup := v1.Group("/auth/oidc")
	SetupOIDCRoutes(oidcGroup, oidcService)

	// Users management (RESTful)
	usersGroup := v1.Group("")
	SetupUsersRoutes(usersGroup, authService, userService)

	// OIDC providers (RESTful)
	oidcProvidersGroup := v1.Group("")
	SetupOIDCProvidersRoutes(oidcProvidersGroup, oidcService)
}

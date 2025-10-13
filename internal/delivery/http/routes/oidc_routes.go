package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupOIDCRoutes sets up OIDC authentication routes
func SetupOIDCRoutes(router *gin.RouterGroup, oidcService domain.OIDCService) {
	oidcHandler := httpDelivery.NewOIDCHandler(oidcService)

	router.GET("/:provider", oidcHandler.GetAuthURL)
	router.GET("/:provider/callback", oidcHandler.Callback)
	router.POST("/login", oidcHandler.Login)
	router.POST("/logout", oidcHandler.Logout)
	router.GET("/logout/url", oidcHandler.GetLogoutURL)
}

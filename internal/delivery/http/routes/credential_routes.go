package routes

import (
	"skyclust/internal/domain"
	httpDelivery "skyclust/internal/delivery/http"

	"github.com/gin-gonic/gin"
)

// SetupCredentialRoutes sets up credential management routes
func SetupCredentialRoutes(router *gin.RouterGroup, credentialService domain.CredentialService) {
	credentialHandler := httpDelivery.NewCredentialHandler(credentialService)

	router.POST("", credentialHandler.CreateCredential)
	router.GET("", credentialHandler.GetCredentials)
	router.GET("/:id", credentialHandler.GetCredential)
	router.PUT("/:id", credentialHandler.UpdateCredential)
	router.DELETE("/:id", credentialHandler.DeleteCredential)
	router.GET("/provider/:provider", credentialHandler.GetCredentialsByProvider)
}

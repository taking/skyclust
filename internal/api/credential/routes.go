package credential

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up credential management routes
func SetupRoutes(router *gin.RouterGroup, credentialService domain.CredentialService) {
	credentialHandler := NewHandler(credentialService)

	router.POST("", credentialHandler.CreateCredential)
	router.GET("", credentialHandler.GetCredentials)
	router.GET("/:id", credentialHandler.GetCredential)
	router.PUT("/:id", credentialHandler.UpdateCredential)
	router.DELETE("/:id", credentialHandler.DeleteCredential)
	router.GET("/provider/:provider", credentialHandler.GetCredentialsByProvider)
}

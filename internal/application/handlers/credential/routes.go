package credential

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up credential management routes (RESTful)
func SetupRoutes(router *gin.RouterGroup, credentialService domain.CredentialService) {
	credentialHandler := NewHandler(credentialService)

	// Standard RESTful routes
	router.POST("", credentialHandler.CreateCredential)       // POST /api/v1/credentials
	router.GET("", credentialHandler.GetCredentials)          // GET /api/v1/credentials (supports ?provider=aws query param)
	router.GET("/:id", credentialHandler.GetCredential)       // GET /api/v1/credentials/:id
	router.PUT("/:id", credentialHandler.UpdateCredential)    // PUT /api/v1/credentials/:id
	router.DELETE("/:id", credentialHandler.DeleteCredential) // DELETE /api/v1/credentials/:id

	// File upload route (special case for multipart/form-data)
	router.POST("/upload", credentialHandler.CreateCredentialFromFile) // POST /api/v1/credentials/upload
}

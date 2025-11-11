package providers

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all provider Kubernetes handlers
// It embeds ProviderBaseHandler to provide standardized provider handler functionality
type BaseHandler struct {
	*handlers.ProviderBaseHandler[*kubernetesservice.Service]
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
	provider string,
	handlerName string,
) *BaseHandler {
	return &BaseHandler{
		ProviderBaseHandler: handlers.NewProviderBaseHandler(
			k8sService,
			credentialService,
			provider,
			handlerName,
		),
	}
}

// Helper methods for parsing request parameters
// These methods delegate to ProviderBaseHandler methods for consistency

func (h *BaseHandler) parseClusterName(c *gin.Context) string {
	return h.ParseClusterName(c)
}

func (h *BaseHandler) parseRegion(c *gin.Context) string {
	return h.ParseRegion(c)
}

// GetK8sService returns the Kubernetes service instance
func (h *BaseHandler) GetK8sService() *kubernetesservice.Service {
	return h.GetService()
}

// GetCredentialService returns the credential service instance
func (h *BaseHandler) GetCredentialService() domain.CredentialService {
	return h.ProviderBaseHandler.GetCredentialService()
}

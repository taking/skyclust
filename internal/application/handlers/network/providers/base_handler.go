package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all provider network handlers
// It embeds ProviderBaseHandler to provide standardized provider handler functionality
type BaseHandler struct {
	*handlers.ProviderBaseHandler[*networkservice.Service]
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
	provider string,
	handlerName string,
) *BaseHandler {
	return &BaseHandler{
		ProviderBaseHandler: handlers.NewProviderBaseHandler(
			networkService,
			credentialService,
			provider,
			handlerName,
		),
	}
}

// Helper methods for parsing request parameters
// These methods delegate to ProviderBaseHandler methods for consistency

func (h *BaseHandler) parseRegion(c *gin.Context) string {
	return h.ParseRegion(c)
}

func (h *BaseHandler) parseVPCID(c *gin.Context) string {
	return h.ParseVPCID(c)
}

func (h *BaseHandler) parseSubnetID(c *gin.Context) string {
	return h.ParseSubnetID(c)
}

// GetNetworkService returns the network service instance
func (h *BaseHandler) GetNetworkService() *networkservice.Service {
	return h.GetService()
}

// GetCredentialService returns the credential service instance
func (h *BaseHandler) GetCredentialService() domain.CredentialService {
	return h.ProviderBaseHandler.GetCredentialService()
}

package providers

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all provider handlers
type BaseHandler struct {
	*handlers.BaseHandler
	k8sService        *kubernetesservice.Service
	credentialService domain.CredentialService
	provider          string
	readabilityHelper *readability.ReadabilityHelper
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
	provider string,
	handlerName string,
) *BaseHandler {
	return &BaseHandler{
		BaseHandler:       handlers.NewBaseHandler(handlerName),
		k8sService:        k8sService,
		credentialService: credentialService,
		provider:          provider,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// Helper methods for parsing request parameters

func (h *BaseHandler) parseClusterName(c *gin.Context) string {
	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "parse_cluster_name")
		return ""
	}
	return clusterName
}

func (h *BaseHandler) parseRegion(c *gin.Context) string {
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "parse_region")
		return ""
	}
	return region
}

// NotImplemented handles unimplemented endpoints
func (h *BaseHandler) NotImplemented(c *gin.Context, operation string) {
	h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotImplemented, operation+" not yet implemented", 501), operation)
}

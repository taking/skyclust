package handlers

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
)

// ProviderBaseHandler provides common functionality for all provider-specific handlers
// T is the service type (e.g., *kubernetesservice.Service, *networkservice.Service)
type ProviderBaseHandler[T any] struct {
	*BaseHandler
	service           T
	credentialService domain.CredentialService
	provider          string
	readabilityHelper *readability.ReadabilityHelper
}

// NewProviderBaseHandler creates a new provider base handler with common dependencies
func NewProviderBaseHandler[T any](
	service T,
	credentialService domain.CredentialService,
	provider string,
	handlerName string,
) *ProviderBaseHandler[T] {
	return &ProviderBaseHandler[T]{
		BaseHandler:       NewBaseHandler(handlerName),
		service:           service,
		credentialService: credentialService,
		provider:          provider,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// GetService returns the service instance
func (h *ProviderBaseHandler[T]) GetService() T {
	return h.service
}

// GetCredentialService returns the credential service
func (h *ProviderBaseHandler[T]) GetCredentialService() domain.CredentialService {
	return h.credentialService
}

// GetProvider returns the provider name
func (h *ProviderBaseHandler[T]) GetProvider() string {
	return h.provider
}

// GetReadabilityHelper returns the readability helper
func (h *ProviderBaseHandler[T]) GetReadabilityHelper() *readability.ReadabilityHelper {
	return h.readabilityHelper
}

// Helper methods for parsing common request parameters

// ParseRegion extracts and validates region from query parameter
func (h *ProviderBaseHandler[T]) ParseRegion(c *gin.Context) string {
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "parse_region")
		return ""
	}
	return region
}

// ParseVPCID extracts and validates VPC ID from query parameter or path parameter
func (h *ProviderBaseHandler[T]) ParseVPCID(c *gin.Context) string {
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		vpcID = c.Param("vpc_id")
	}
	if vpcID == "" {
		vpcID = c.Param("id")
	}
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "parse_vpc_id")
		return ""
	}
	return vpcID
}

// ParseSubnetID extracts and validates Subnet ID from query parameter or path parameter
func (h *ProviderBaseHandler[T]) ParseSubnetID(c *gin.Context) string {
	subnetID := c.Query("subnet_id")
	if subnetID == "" {
		subnetID = c.Param("subnet_id")
	}
	if subnetID == "" {
		subnetID = c.Param("id")
	}
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "subnet_id is required", 400), "parse_subnet_id")
		return ""
	}
	return subnetID
}

// ParseClusterName extracts and validates cluster name from path parameter
func (h *ProviderBaseHandler[T]) ParseClusterName(c *gin.Context) string {
	clusterName := c.Param("name")
	if clusterName == "" {
		clusterName = c.Param("cluster_name")
	}
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "parse_cluster_name")
		return ""
	}
	return clusterName
}

// NotImplemented handles unimplemented endpoints with consistent error response
func (h *ProviderBaseHandler[T]) NotImplemented(c *gin.Context, operation string) {
	h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotImplemented, operation+" not yet implemented", 501), operation)
}

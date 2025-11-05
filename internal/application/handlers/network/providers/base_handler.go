package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all provider network handlers
type BaseHandler struct {
	*handlers.BaseHandler
	networkService    *networkservice.Service
	credentialService domain.CredentialService
	provider          string
	readabilityHelper *readability.ReadabilityHelper
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
	provider string,
	handlerName string,
) *BaseHandler {
	return &BaseHandler{
		BaseHandler:       handlers.NewBaseHandler(handlerName),
		networkService:    networkService,
		credentialService: credentialService,
		provider:          provider,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// Helper methods for parsing request parameters

func (h *BaseHandler) parseRegion(c *gin.Context) string {
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "parse_region")
		return ""
	}
	return region
}

func (h *BaseHandler) parseVPCID(c *gin.Context) string {
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "parse_vpc_id")
		return ""
	}
	return vpcID
}

// NotImplemented handles unimplemented endpoints
func (h *BaseHandler) NotImplemented(c *gin.Context, operation string) {
	h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotImplemented, operation+" not yet implemented", 501), operation)
}


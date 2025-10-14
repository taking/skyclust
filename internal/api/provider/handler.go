package provider

import (
	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/plugin"
	"skyclust/internal/utils"
	plugininterfaces "skyclust/pkg/plugin"

	"github.com/gin-gonic/gin"
)

// Handler handles provider-related HTTP requests
type Handler struct {
	pluginManager  *plugin.Manager
	auditLogRepo   domain.AuditLogRepository
	tokenExtractor *utils.TokenExtractor
}

// NewHandler creates a new provider handler
func NewHandler(pluginManager *plugin.Manager, auditLogRepo domain.AuditLogRepository) *Handler {
	return &Handler{
		pluginManager:  pluginManager,
		auditLogRepo:   auditLogRepo,
		tokenExtractor: utils.NewTokenExtractor(),
	}
}

// GetProviders returns the list of available providers
func (h *Handler) GetProviders(c *gin.Context) {
	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get available providers from plugin manager
	providers := h.pluginManager.ListProviders()

	// TODO: Create audit log
	_ = userID

	common.OK(c, gin.H{
		"providers": providers,
	}, "Providers retrieved successfully")
}

// GetProvider returns information about a specific provider
func (h *Handler) GetProvider(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider information
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, provider, "Provider information retrieved successfully")
}

// GetInstances returns instances for a specific provider
func (h *Handler) GetInstances(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Get instances
	instances, err := provider.ListInstances(c.Request.Context())
	if err != nil {
		common.InternalServerError(c, "Failed to get instances")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, gin.H{
		"instances": instances,
	}, "Instances retrieved successfully")
}

// GetInstance returns information about a specific instance
func (h *Handler) GetInstance(c *gin.Context) {
	providerName := c.Param("name")
	instanceID := c.Param("id")

	if providerName == "" || instanceID == "" {
		common.BadRequest(c, "Provider name and instance ID are required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Get instance status
	status, err := provider.GetInstanceStatus(c.Request.Context(), instanceID)
	if err != nil {
		common.NotFound(c, "Instance not found")
		return
	}

	// Create instance info response
	instance := gin.H{
		"id":       instanceID,
		"provider": providerName,
		"status":   status,
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, instance, "Instance information retrieved successfully")
}

// CreateInstance creates a new instance
func (h *Handler) CreateInstance(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	var req plugininterfaces.CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Create instance
	instance, err := provider.CreateInstance(c.Request.Context(), req)
	if err != nil {
		common.InternalServerError(c, "Failed to create instance")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, instance, "Instance created successfully")
}

// DeleteInstance deletes an instance
func (h *Handler) DeleteInstance(c *gin.Context) {
	providerName := c.Param("name")
	instanceID := c.Param("id")

	if providerName == "" || instanceID == "" {
		common.BadRequest(c, "Provider name and instance ID are required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Delete instance
	err = provider.DeleteInstance(c.Request.Context(), instanceID)
	if err != nil {
		common.InternalServerError(c, "Failed to delete instance")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, gin.H{"message": "Instance deleted successfully"}, "Instance deleted successfully")
}

// GetRegions returns available regions for a provider
func (h *Handler) GetRegions(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Get regions
	regions, err := provider.ListRegions(c.Request.Context())
	if err != nil {
		common.InternalServerError(c, "Failed to get regions")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, gin.H{
		"regions": regions,
	}, "Regions retrieved successfully")
}

// GetCostEstimates returns cost estimates for a provider
func (h *Handler) GetCostEstimates(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get provider
	_, err = h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// TODO: Implement proper cost estimation retrieval
	// For now, return empty estimates as this requires specific parameters
	estimates := []gin.H{}

	// TODO: Create audit log
	_ = userID

	common.OK(c, gin.H{
		"estimates": estimates,
	}, "Cost estimates retrieved successfully")
}

// CreateCostEstimate creates a new cost estimate
func (h *Handler) CreateCostEstimate(c *gin.Context) {
	providerName := c.Param("name")
	if providerName == "" {
		common.BadRequest(c, "Provider name is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	var req plugininterfaces.CostEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		common.NotFound(c, "Provider not found")
		return
	}

	// Create cost estimate
	estimate, err := provider.GetCostEstimate(c.Request.Context(), req)
	if err != nil {
		common.InternalServerError(c, "Failed to create cost estimate")
		return
	}

	// TODO: Create audit log
	_ = userID

	common.OK(c, estimate, "Cost estimate created successfully")
}

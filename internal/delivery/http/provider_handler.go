package http

import (
	"skyclust/internal/domain"
	"skyclust/internal/plugin"
	"skyclust/internal/plugin/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProviderHandler handles provider-related HTTP requests
type ProviderHandler struct {
	pluginManager *plugin.Manager
	auditLogRepo  domain.AuditLogRepository
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(pluginManager *plugin.Manager, auditLogRepo domain.AuditLogRepository) *ProviderHandler {
	return &ProviderHandler{
		pluginManager: pluginManager,
		auditLogRepo:  auditLogRepo,
	}
}

// GetProviders returns the list of available providers
func (h *ProviderHandler) GetProviders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	// Get list of loaded providers
	providers := h.pluginManager.ListProviders()

	// Log provider list request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionProviderList,
		Resource: "GET /api/v1/providers",
		Details: map[string]interface{}{
			"providers": providers,
		},
	})

	OKResponse(c, gin.H{
		"providers": providers,
	}, "Providers retrieved successfully")
}

// GetProvider returns information about a specific provider
func (h *ProviderHandler) GetProvider(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Get provider information
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	providerInfo := gin.H{
		"name":    provider.GetName(),
		"version": provider.GetVersion(),
	}

	// Log provider info request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionProviderList,
		Resource: "GET /api/v1/providers/" + providerName,
		Details: map[string]interface{}{
			"provider": providerName,
		},
	})

	OKResponse(c, providerInfo, "Provider information retrieved successfully")
}

// GetInstances returns instances for a specific provider
func (h *ProviderHandler) GetInstances(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Get region parameter
	region := c.Query("region")
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// List instances
	instances, err := provider.ListInstances(c.Request.Context())
	if err != nil {
		InternalServerErrorResponse(c, "Failed to list instances")
		return
	}

	// Filter instances by region if specified
	if region != "" {
		filteredInstances := make([]interfaces.Instance, 0)
		for _, instance := range instances {
			if instance.Region == region {
				filteredInstances = append(filteredInstances, instance)
			}
		}
		instances = filteredInstances
	}

	// Log instance list request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionInstanceList,
		Resource: "GET /api/v1/providers/" + providerName + "/instances",
		Details: map[string]interface{}{
			"provider": providerName,
			"region":   region,
			"count":    len(instances),
		},
	})

	OKResponse(c, gin.H{
		"instances": instances,
		"provider":  providerName,
		"region":    region,
		"count":     len(instances),
	}, "Instances retrieved successfully")
}

// GetInstance returns information about a specific instance
func (h *ProviderHandler) GetInstance(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	instanceID := c.Param("id")

	if providerName == "" || instanceID == "" {
		BadRequestResponse(c, "Provider name and instance ID are required")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// Get instance details - List instances and find the specific one
	instances, err := provider.ListInstances(c.Request.Context())
	if err != nil {
		InternalServerErrorResponse(c, "Failed to list instances")
		return
	}

	var instance *interfaces.Instance
	for _, inst := range instances {
		if inst.ID == instanceID {
			instance = &inst
			break
		}
	}

	if instance == nil {
		NotFoundResponse(c, "Instance not found")
		return
	}

	// Log instance get request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionInstanceList,
		Resource: "GET /api/v1/providers/" + providerName + "/instances/" + instanceID,
		Details: map[string]interface{}{
			"provider":    providerName,
			"instance_id": instanceID,
		},
	})

	OKResponse(c, instance, "Instance retrieved successfully")
}

// GetRegions returns available regions for a provider
func (h *ProviderHandler) GetRegions(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// Get regions
	regions, err := provider.ListRegions(c.Request.Context())
	if err != nil {
		InternalServerErrorResponse(c, "Failed to list regions")
		return
	}

	// Log regions request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   "list_regions",
		Resource: "GET /api/v1/providers/" + providerName + "/regions",
		Details: map[string]interface{}{
			"provider": providerName,
			"count":    len(regions),
		},
	})

	OKResponse(c, gin.H{
		"regions":  regions,
		"provider": providerName,
		"count":    len(regions),
	}, "Regions retrieved successfully")
}

// CreateInstance creates a new instance
func (h *ProviderHandler) CreateInstance(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Parse request body
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// Create instance
	createReq := interfaces.CreateInstanceRequest{
		Name:     req["name"].(string),
		Type:     req["type"].(string),
		Region:   req["region"].(string),
		ImageID:  req["image_id"].(string),
		Tags:     make(map[string]string),
		UserData: req["user_data"].(string),
	}
	instance, err := provider.CreateInstance(c.Request.Context(), createReq)
	if err != nil {
		InternalServerErrorResponse(c, "Failed to create instance")
		return
	}

	// Log instance creation
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionInstanceCreate,
		Resource: "POST /api/v1/providers/" + providerName + "/instances",
		Details: map[string]interface{}{
			"provider": providerName,
			"request":  req,
		},
	})

	CreatedResponse(c, instance, "Instance created successfully")
}

// DeleteInstance deletes an instance
func (h *ProviderHandler) DeleteInstance(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	instanceID := c.Param("id")

	if providerName == "" || instanceID == "" {
		BadRequestResponse(c, "Provider name and instance ID are required")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// Delete instance
	if err := provider.DeleteInstance(c.Request.Context(), instanceID); err != nil {
		InternalServerErrorResponse(c, "Failed to delete instance")
		return
	}

	// Log instance deletion
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   domain.ActionInstanceDelete,
		Resource: "DELETE /api/v1/providers/" + providerName + "/instances/" + instanceID,
		Details: map[string]interface{}{
			"provider":    providerName,
			"instance_id": instanceID,
		},
	})

	OKResponse(c, gin.H{"message": "Instance deleted successfully"}, "Instance deleted successfully")
}

// GetCostEstimates returns cost estimates for a provider
func (h *ProviderHandler) GetCostEstimates(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Get provider
	_, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// For now, return empty list - in real implementation, you would store and retrieve cost estimates
	estimates := []interface{}{}

	// Log cost estimates request
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   "get_cost_estimates",
		Resource: "GET /api/v1/providers/" + providerName + "/cost-estimates",
		Details: map[string]interface{}{
			"provider": providerName,
		},
	})

	OKResponse(c, gin.H{
		"estimates": estimates,
		"count":     len(estimates),
	}, "Cost estimates retrieved successfully")
}

// CreateCostEstimate creates a new cost estimate for a provider
func (h *ProviderHandler) CreateCostEstimate(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	providerName := c.Param("name")
	if providerName == "" {
		BadRequestResponse(c, "Provider name is required")
		return
	}

	// Parse request body
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Get provider
	provider, err := h.pluginManager.GetProvider(providerName)
	if err != nil {
		NotFoundResponse(c, "Provider not found")
		return
	}

	// Get cost estimate
	costReq := interfaces.CostEstimateRequest{
		InstanceType: req["instance_type"].(string),
		Region:       req["region"].(string),
		Duration:     req["duration"].(string),
	}
	estimate, err := provider.GetCostEstimate(c.Request.Context(), costReq)
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get cost estimate")
		return
	}

	// Log cost estimate creation
	_ = h.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   "create_cost_estimate",
		Resource: "POST /api/v1/providers/" + providerName + "/cost-estimates",
		Details: map[string]interface{}{
			"provider":      providerName,
			"instance_type": req["instance_type"],
			"region":        req["region"],
			"duration":      req["duration"],
		},
	})

	CreatedResponse(c, estimate, "Cost estimate created successfully")
}

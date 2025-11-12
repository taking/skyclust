package resourcegroup

import (
	resourcegroupservice "skyclust/internal/application/services/resourcegroup"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up Azure Resource Group routes
func SetupRoutes(router *gin.RouterGroup, resourceGroupService *resourcegroupservice.Service, credentialService domain.CredentialService) {
	handler := NewHandler(resourceGroupService, credentialService)

	// Resource Group management
	// Path: /api/v1/azure/iam/resource-groups
	router.GET("/resource-groups", handler.ListResourceGroups)
	router.POST("/resource-groups", handler.CreateResourceGroup)
	router.GET("/resource-groups/:name", handler.GetResourceGroup)
	router.PUT("/resource-groups/:name", handler.UpdateResourceGroup)
	router.DELETE("/resource-groups/:name", handler.DeleteResourceGroup)
}


package handlers

import (
	"net/http"

	"cmp/internal/services"

	"github.com/gin-gonic/gin"
)

// ListProviders lists all available cloud providers
func ListProviders(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		providers, err := cloudService.ListProviders()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list providers"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"providers": providers})
	}
}

// GetProvider gets a specific provider
func GetProvider(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("name")

		provider, err := cloudService.GetProvider(providerName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"provider": provider})
	}
}

// InitializeProvider initializes a cloud provider
func InitializeProvider(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("name")

		var req struct {
			Config map[string]interface{} `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := cloudService.InitializeProvider(providerName, req.Config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize provider"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Provider initialized successfully"})
	}
}

// ListVMs lists all VMs in a workspace
func ListVMs(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		provider := c.Query("provider")

		vms, err := cloudService.ListVMs(workspaceID.(string), provider)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list VMs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"vms": vms})
	}
}

// CreateVM creates a new VM
func CreateVM(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		var req struct {
			Provider string `json:"provider" binding:"required"`
			Name     string `json:"name" binding:"required"`
			Type     string `json:"type" binding:"required"`
			Region   string `json:"region" binding:"required"`
			ImageID  string `json:"image_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert to CreateInstanceRequest
		instanceReq := services.CreateInstanceRequest{
			Name:    req.Name,
			Type:    req.Type,
			Region:  req.Region,
			ImageID: req.ImageID,
		}

		vm, err := cloudService.CreateVM(workspaceID.(string), req.Provider, instanceReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create VM"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"vm": vm})
	}
}

// GetVM gets a specific VM
func GetVM(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		vmID := c.Param("vmId")
		provider := c.Query("provider")

		vm, err := cloudService.GetVM(workspaceID.(string), provider, vmID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "VM not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"vm": vm})
	}
}

// DeleteVM deletes a VM
func DeleteVM(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		vmID := c.Param("vmId")
		provider := c.Query("provider")

		if err := cloudService.DeleteVM(workspaceID.(string), provider, vmID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete VM"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "VM deleted successfully"})
	}
}

// StartVM starts a VM
func StartVM(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		vmID := c.Param("vmId")
		provider := c.Query("provider")

		if err := cloudService.StartVM(workspaceID.(string), provider, vmID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start VM"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "VM started successfully"})
	}
}

// StopVM stops a VM
func StopVM(cloudService services.CloudService) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		vmID := c.Param("vmId")
		provider := c.Query("provider")

		if err := cloudService.StopVM(workspaceID.(string), provider, vmID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop VM"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "VM stopped successfully"})
	}
}

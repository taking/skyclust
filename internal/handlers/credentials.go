package handlers

import (
	"net/http"

	"cmp/pkg/credentials"

	"github.com/gin-gonic/gin"
)

// ListCredentials lists all credentials for a workspace
func ListCredentials(credentialsService credentials.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		creds, err := credentialsService.ListCredentials(c.Request.Context(), workspaceID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"credentials": creds})
	}
}

// CreateCredentials creates new credentials
func CreateCredentials(credentialsService credentials.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		var req struct {
			Provider string            `json:"provider" binding:"required"`
			Config   map[string]string `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cred, err := credentialsService.CreateCredentials(c.Request.Context(), workspaceID.(string), req.Provider, req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credentials"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"credentials": cred})
	}
}

// GetCredentials gets specific credentials
func GetCredentials(credentialsService credentials.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		credID := c.Param("credId")

		cred, err := credentialsService.GetCredentials(c.Request.Context(), workspaceID.(string), credID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credentials not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"credentials": cred})
	}
}

// UpdateCredentials updates credentials
func UpdateCredentials(credentialsService credentials.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		credID := c.Param("credId")

		var req struct {
			Config map[string]string `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cred, err := credentialsService.UpdateCredentials(c.Request.Context(), workspaceID.(string), credID, req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"credentials": cred})
	}
}

// DeleteCredentials deletes credentials
func DeleteCredentials(credentialsService credentials.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		credID := c.Param("credId")

		if err := credentialsService.DeleteCredentials(c.Request.Context(), workspaceID.(string), credID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Credentials deleted successfully"})
	}
}

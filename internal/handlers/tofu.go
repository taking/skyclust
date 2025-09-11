package handlers

import (
	"net/http"

	"cmp/pkg/iac"

	"github.com/gin-gonic/gin"
)

// PlanTofu plans OpenTofu execution
func PlanTofu(iacService iac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		var req struct {
			Config string `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		execution, err := iacService.Plan(c.Request.Context(), workspaceID.(string), req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to plan OpenTofu"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"execution": execution})
	}
}

// ApplyTofu applies OpenTofu configuration
func ApplyTofu(iacService iac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		var req struct {
			Config string `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		execution, err := iacService.Apply(c.Request.Context(), workspaceID.(string), req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply OpenTofu"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"execution": execution})
	}
}

// DestroyTofu destroys OpenTofu resources
func DestroyTofu(iacService iac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		var req struct {
			Config string `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		execution, err := iacService.Destroy(c.Request.Context(), workspaceID.(string), req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to destroy OpenTofu"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"execution": execution})
	}
}

// ListExecutions lists all OpenTofu executions
func ListExecutions(iacService iac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")

		executions, err := iacService.ListExecutions(c.Request.Context(), workspaceID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list executions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"executions": executions})
	}
}

// GetExecution gets a specific execution
func GetExecution(iacService iac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID, _ := c.Get("workspace_id")
		executionID := c.Param("execId")

		execution, err := iacService.GetExecution(c.Request.Context(), workspaceID.(string), executionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"execution": execution})
	}
}

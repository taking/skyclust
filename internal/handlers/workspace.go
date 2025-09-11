package handlers

import (
	"net/http"

	"cmp/pkg/workspace"

	"github.com/gin-gonic/gin"
)

// ListWorkspaces lists all workspaces for the current user
func ListWorkspaces(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		workspaces, err := workspaceService.ListWorkspaces(c.Request.Context(), userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workspaces"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
	}
}

// CreateWorkspace creates a new workspace
func CreateWorkspace(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ws, err := workspaceService.CreateWorkspace(c.Request.Context(), req.Name, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workspace"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"workspace": ws})
	}
}

// GetWorkspace gets a specific workspace
func GetWorkspace(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")

		ws, err := workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"workspace": ws})
	}
}

// UpdateWorkspace updates a workspace
func UpdateWorkspace(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")

		var req struct {
			Name     string                 `json:"name"`
			Settings map[string]interface{} `json:"settings"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get existing workspace
		ws, err := workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
			return
		}

		// Update fields
		if req.Name != "" {
			ws.Name = req.Name
		}
		if req.Settings != nil {
			ws.Settings = req.Settings
		}

		if err := workspaceService.UpdateWorkspace(c.Request.Context(), ws); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workspace"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"workspace": ws})
	}
}

// DeleteWorkspace deletes a workspace
func DeleteWorkspace(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")

		if err := workspaceService.DeleteWorkspace(c.Request.Context(), workspaceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workspace"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted successfully"})
	}
}

// WorkspaceMiddleware validates workspace access
func WorkspaceMiddleware(workspaceService workspace.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		userID, _ := c.Get("user_id")

		// Check if user has access to workspace
		hasPermission, err := workspaceService.HasPermission(c.Request.Context(), workspaceID, userID.(string), "read")
		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to workspace"})
			c.Abort()
			return
		}

		c.Set("workspace_id", workspaceID)
		c.Next()
	}
}

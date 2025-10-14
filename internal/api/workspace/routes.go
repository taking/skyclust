package workspace

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up workspace management routes
func SetupRoutes(router *gin.RouterGroup, workspaceService domain.WorkspaceService, userService domain.UserService) {
	workspaceHandler := NewHandler(workspaceService, userService)

	router.POST("", workspaceHandler.CreateWorkspace)
	router.GET("", workspaceHandler.GetWorkspaces)
	router.GET("/:id", workspaceHandler.GetWorkspace)
	router.PUT("/:id", workspaceHandler.UpdateWorkspace)
	router.DELETE("/:id", workspaceHandler.DeleteWorkspace)
}

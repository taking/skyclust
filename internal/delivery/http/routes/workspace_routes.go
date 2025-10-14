package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupWorkspaceRoutes sets up workspace management routes
func SetupWorkspaceRoutes(router *gin.RouterGroup, workspaceService domain.WorkspaceService, userService domain.UserService) {
	workspaceHandler := httpDelivery.NewWorkspaceHandler(workspaceService, userService)

	router.POST("", workspaceHandler.CreateWorkspace)
	router.GET("", workspaceHandler.GetWorkspaces)
	router.GET("/:id", workspaceHandler.GetWorkspace)
	router.PUT("/:id", workspaceHandler.UpdateWorkspace)
	router.DELETE("/:id", workspaceHandler.DeleteWorkspace)
}

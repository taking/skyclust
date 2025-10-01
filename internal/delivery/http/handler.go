package http

import (
	"cmp/internal/container"
	"cmp/pkg/shared/logger"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests
type Handler struct {
	container *container.Container
}

// NewHandler creates a new HTTP handler
func NewHandler(container *container.Container) *Handler {
	return &Handler{
		container: container,
	}
}

// SetupRoutes sets up all HTTP routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("/", h.CreateUser)
			users.GET("/:id", h.GetUser)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
			users.POST("/authenticate", h.Authenticate)
		}

		// Workspace routes
		workspaces := v1.Group("/workspaces")
		{
			workspaces.POST("/", h.CreateWorkspace)
			workspaces.GET("/:id", h.GetWorkspace)
			workspaces.PUT("/:id", h.UpdateWorkspace)
			workspaces.DELETE("/:id", h.DeleteWorkspace)
			workspaces.GET("/", h.ListWorkspaces)
		}

		// VM routes
		vms := v1.Group("/vms")
		{
			vms.POST("/", h.CreateVM)
			vms.GET("/:id", h.GetVM)
			vms.DELETE("/:id", h.DeleteVM)
			vms.GET("/", h.ListVMs)
			vms.POST("/:id/start", h.StartVM)
			vms.POST("/:id/stop", h.StopVM)
		}
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database health
	if err := h.container.Health(ctx); err != nil {
		logger.Error("Health check failed")
		h.InternalServerErrorResponse(c, "Health check failed")
		return
	}

	h.OKResponse(c, gin.H{"status": "healthy"}, "Health check successful")
}

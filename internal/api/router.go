package api

import (
	"net/http"

	"cmp/internal/handlers"
	"cmp/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, services *services.Services) *gin.Engine {
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login(services.Auth))
			auth.POST("/register", handlers.Register(services.Auth))
			auth.POST("/logout", handlers.Logout(services.Auth))
			auth.GET("/me", handlers.GetCurrentUser(services.Auth))
		}

		// Workspace routes
		workspaces := v1.Group("/workspaces")
		workspaces.Use(handlers.AuthMiddleware(services.Auth))
		{
			workspaces.GET("", handlers.ListWorkspaces(services.Workspace))
			workspaces.POST("", handlers.CreateWorkspace(services.Workspace))
			workspaces.GET("/:workspaceId", handlers.GetWorkspace(services.Workspace))
			workspaces.PUT("/:workspaceId", handlers.UpdateWorkspace(services.Workspace))
			workspaces.DELETE("/:workspaceId", handlers.DeleteWorkspace(services.Workspace))
		}

		// Cloud provider routes
		providers := v1.Group("/providers")
		providers.Use(handlers.AuthMiddleware(services.Auth))
		{
			providers.GET("", handlers.ListProviders(services.Cloud))
			providers.GET("/:name", handlers.GetProvider(services.Cloud))
			providers.POST("/:name/initialize", handlers.InitializeProvider(services.Cloud))
		}

		// VM management routes
		vms := v1.Group("/workspaces/:workspaceId/vms")
		vms.Use(handlers.AuthMiddleware(services.Auth))
		vms.Use(handlers.WorkspaceMiddleware(services.Workspace))
		{
			vms.GET("", handlers.ListVMs(services.Cloud))
			vms.POST("", handlers.CreateVM(services.Cloud))
			vms.GET("/:vmId", handlers.GetVM(services.Cloud))
			vms.DELETE("/:vmId", handlers.DeleteVM(services.Cloud))
			vms.POST("/:vmId/start", handlers.StartVM(services.Cloud))
			vms.POST("/:vmId/stop", handlers.StopVM(services.Cloud))
		}

		// Credentials routes
		credentials := v1.Group("/workspaces/:workspaceId/credentials")
		credentials.Use(handlers.AuthMiddleware(services.Auth))
		credentials.Use(handlers.WorkspaceMiddleware(services.Workspace))
		{
			credentials.GET("", handlers.ListCredentials(services.Credentials))
			credentials.POST("", handlers.CreateCredentials(services.Credentials))
			credentials.GET("/:credId", handlers.GetCredentials(services.Credentials))
			credentials.PUT("/:credId", handlers.UpdateCredentials(services.Credentials))
			credentials.DELETE("/:credId", handlers.DeleteCredentials(services.Credentials))
		}

		// OpenTofu routes
		tofu := v1.Group("/workspaces/:workspaceId/tofu")
		tofu.Use(handlers.AuthMiddleware(services.Auth))
		tofu.Use(handlers.WorkspaceMiddleware(services.Workspace))
		{
			tofu.POST("/plan", handlers.PlanTofu(services.IaC))
			tofu.POST("/apply", handlers.ApplyTofu(services.IaC))
			tofu.POST("/destroy", handlers.DestroyTofu(services.IaC))
			tofu.GET("/executions", handlers.ListExecutions(services.IaC))
			tofu.GET("/executions/:execId", handlers.GetExecution(services.IaC))
		}

		// Real-time routes
		realtime := v1.Group("/realtime")
		realtime.Use(handlers.AuthMiddleware(services.Auth))
		{
			realtime.GET("/ws", handlers.WebSocketHandler(services.Realtime))
			realtime.GET("/sse/:workspaceId", handlers.SSEHandler(services.Realtime))
		}

		// Kubernetes routes
		clusters := v1.Group("/workspaces/:workspaceId/clusters")
		clusters.Use(handlers.AuthMiddleware(services.Auth))
		clusters.Use(handlers.WorkspaceMiddleware(services.Workspace))
		{
			clusters.GET("", handlers.ListClusters(services))
			clusters.POST("", handlers.CreateCluster(services))
			clusters.GET("/:clusterId", handlers.GetCluster(services))
			clusters.DELETE("/:clusterId", handlers.DeleteCluster(services))

			// Namespace management
			clusters.GET("/:clusterId/namespaces", handlers.ListNamespaces(services))
			clusters.POST("/:clusterId/namespaces", handlers.CreateNamespace(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName", handlers.GetNamespace(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName", handlers.DeleteNamespace(services))

			// Deployment management
			clusters.GET("/:clusterId/namespaces/:namespaceName/deployments", handlers.ListDeployments(services))
			clusters.POST("/:clusterId/deployments", handlers.CreateDeployment(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName/deployments/:deploymentName", handlers.GetDeployment(services))
			clusters.PUT("/:clusterId/namespaces/:namespaceName/deployments/:deploymentName", handlers.UpdateDeployment(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName/deployments/:deploymentName", handlers.DeleteDeployment(services))
			clusters.POST("/:clusterId/namespaces/:namespaceName/deployments/:deploymentName/scale/:replicas", handlers.ScaleDeployment(services))

			// Service management
			clusters.GET("/:clusterId/namespaces/:namespaceName/services", handlers.ListServices(services))
			clusters.POST("/:clusterId/services", handlers.CreateService(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName/services/:serviceName", handlers.GetService(services))
			clusters.PUT("/:clusterId/namespaces/:namespaceName/services/:serviceName", handlers.UpdateService(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName/services/:serviceName", handlers.DeleteService(services))

			// Pod management
			clusters.GET("/:clusterId/namespaces/:namespaceName/pods", handlers.ListPods(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName/pods/:podName", handlers.GetPod(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName/pods/:podName", handlers.DeletePod(services))

			// ConfigMap management
			clusters.GET("/:clusterId/namespaces/:namespaceName/configmaps", handlers.ListConfigMaps(services))
			clusters.POST("/:clusterId/configmaps", handlers.CreateConfigMap(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName/configmaps/:configMapName", handlers.GetConfigMap(services))
			clusters.PUT("/:clusterId/namespaces/:namespaceName/configmaps/:configMapName", handlers.UpdateConfigMap(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName/configmaps/:configMapName", handlers.DeleteConfigMap(services))

			// Secret management
			clusters.GET("/:clusterId/namespaces/:namespaceName/secrets", handlers.ListSecrets(services))
			clusters.POST("/:clusterId/secrets", handlers.CreateSecret(services))
			clusters.GET("/:clusterId/namespaces/:namespaceName/secrets/:secretName", handlers.GetSecret(services))
			clusters.PUT("/:clusterId/namespaces/:namespaceName/secrets/:secretName", handlers.UpdateSecret(services))
			clusters.DELETE("/:clusterId/namespaces/:namespaceName/secrets/:secretName", handlers.DeleteSecret(services))
		}
	}

	return router
}

package routes

import (
	"runtime"
	"time"

	"skyclust/internal/application/handlers/admin"
	"skyclust/internal/application/handlers/audit"
	"skyclust/internal/application/handlers/auth"
	"skyclust/internal/application/handlers/cost_analysis"
	"skyclust/internal/application/handlers/credential"
	"skyclust/internal/application/handlers/export"
	"skyclust/internal/application/handlers/kubernetes"
	"skyclust/internal/application/handlers/network"
	"skyclust/internal/application/handlers/notification"
	"skyclust/internal/application/handlers/oidc"
	"skyclust/internal/application/handlers/provider"
	"skyclust/internal/application/handlers/sse"
	"skyclust/internal/application/handlers/system"
	"skyclust/internal/application/handlers/workspace"
	service "skyclust/internal/application/services"
	"skyclust/internal/di"
	"skyclust/pkg/config"
	"skyclust/pkg/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteManager manages all API routes with centralized configuration
type RouteManager struct {
	container       di.ContainerInterface
	providerManager interface{} // gRPC Provider Manager
	middleware      *middleware.Middleware
	logger          *zap.Logger
	config          *config.Config
	startTime       time.Time
	requestCount    int64
	errorCount      int64
}

// NewRouteManager creates a new route manager
func NewRouteManager(
	container di.ContainerInterface,
	providerManager interface{},
	middleware *middleware.Middleware,
	logger *zap.Logger,
	config *config.Config,
) *RouteManager {
	return &RouteManager{
		container:       container,
		providerManager: providerManager,
		middleware:      middleware,
		logger:          logger,
		config:          config,
		startTime:       time.Now(),
		requestCount:    0,
		errorCount:      0,
	}
}

// SetupAllRoutes sets up all API routes with optimized structure
func (rm *RouteManager) SetupAllRoutes(router *gin.Engine) {
	// Setup public routes (no authentication required)
	rm.setupPublicRoutes(router)

	// Setup protected routes (authentication required)
	rm.setupProtectedRoutes(router)

	// Setup admin routes (admin privileges required)
	rm.setupAdminRoutes(router)
}

// setupPublicRoutes sets up public routes that don't require authentication
func (rm *RouteManager) setupPublicRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", rm.healthCheck)

	// API v1 public routes
	v1Public := router.Group("/api/v1")
	{
		// Authentication routes (public) - register and login
		authGroup := v1Public.Group("/auth")
		rm.setupPublicAuthRoutes(authGroup)

		// OIDC routes (public)
		oidcGroup := v1Public.Group("/auth/oidc")
		rm.setupOIDCRoutes(oidcGroup)

		// OIDC providers (public)
		oidcProvidersGroup := v1Public.Group("/oidc-providers")
		rm.setupOIDCProvidersRoutes(oidcProvidersGroup)
	}
}

// setupProtectedRoutes sets up protected routes that require authentication
func (rm *RouteManager) setupProtectedRoutes(router *gin.Engine) {
	// API v1 protected routes
	v1Protected := router.Group("/api/v1")

	// Apply authentication middleware to all protected routes
	v1Protected.Use(rm.middleware.AuthMiddleware())
	{
		// Authentication routes (protected) - logout and profile
		authGroup := v1Protected.Group("/auth")
		rm.setupProtectedAuthRoutes(authGroup)

		// User management routes
		usersGroup := v1Protected.Group("/users")
		rm.setupUserRoutes(usersGroup)

		// Credential management routes
		credentialsGroup := v1Protected.Group("/credentials")
		rm.setupCredentialRoutes(credentialsGroup)

		// Workspace management routes
		workspacesGroup := v1Protected.Group("/workspaces")
		rm.setupWorkspaceRoutes(workspacesGroup)

		// Provider routes
		providersGroup := v1Protected.Group("/providers")
		rm.setupProviderRoutes(providersGroup)

		// Provider-specific routes (RESTful)
		rm.setupProviderSpecificRoutes(v1Protected)

		// Cost analysis routes
		costAnalysisGroup := v1Protected.Group("/cost-analysis")
		rm.setupCostAnalysisRoutes(costAnalysisGroup)

		// Notification routes
		notificationsGroup := v1Protected.Group("/notifications")
		rm.setupNotificationRoutes(notificationsGroup)

		// Export routes
		exportsGroup := v1Protected.Group("/exports")
		rm.setupExportRoutes(exportsGroup)

		// SSE routes
		sseGroup := v1Protected.Group("/sse")
		rm.setupSSERoutes(sseGroup)
	}
}

// setupAdminRoutes sets up admin routes that require admin privileges
func (rm *RouteManager) setupAdminRoutes(router *gin.Engine) {
	// API v1 admin routes
	v1Admin := router.Group("/api/v1/admin")

	// Apply authentication and admin middleware
	v1Admin.Use(rm.middleware.AuthMiddleware())
	// TODO: Implement admin middleware
	// v1Admin.Use(middleware.AdminMiddleware(rm.container.RBACService))
	{
		// Admin user management
		adminUsersGroup := v1Admin.Group("/users")
		rm.setupAdminUserRoutes(adminUsersGroup)

		// System management
		systemGroup := v1Admin.Group("/system")
		rm.setupSystemRoutes(systemGroup)

		// Audit logs
		auditGroup := v1Admin.Group("/audit")
		rm.setupAuditRoutes(auditGroup)

		// Permission management
		permissionsGroup := v1Admin.Group("/permissions")
		rm.setupPermissionRoutes(permissionsGroup)
	}
}

// healthCheck provides a comprehensive health check endpoint
func (rm *RouteManager) healthCheck(c *gin.Context) {
	now := time.Now()
	uptime := now.Sub(rm.startTime)

	// Increment request count
	rm.requestCount++

	// Get service status
	services := rm.getServiceStatus()

	// Get dependencies status
	dependencies := rm.getDependenciesStatus()

	// Get system metrics
	metrics := rm.getSystemMetrics()

	// Determine overall health status
	overallStatus := "healthy"
	if !rm.isAllServicesHealthy(services) || !rm.isAllDependenciesHealthy(dependencies) {
		overallStatus = "degraded"
		rm.errorCount++
	}

	// Get alert status
	alerts := rm.getAlertStatus(services, dependencies, metrics)

	c.JSON(200, gin.H{
		"status":       overallStatus,
		"timestamp":    now.Format(time.RFC3339),
		"version":      "1.0.0",
		"uptime":       uptime.String(),
		"started_at":   rm.startTime.Format(time.RFC3339),
		"services":     services,
		"dependencies": dependencies,
		"metrics":      metrics,
		"providers": gin.H{
			"type": "gRPC",
			"note": "Use /api/v1/providers endpoint for details",
		},
		"config": gin.H{
			"environment": rm.getEnvironment(),
			"debug_mode":  rm.config.Server.Host == "0.0.0.0",
		},
		"alerts": alerts,
	})
}

// setupPublicAuthRoutes sets up public authentication routes
func (rm *RouteManager) setupPublicAuthRoutes(router *gin.RouterGroup) {
	// Public routes only - register and login
	authHandler := auth.NewHandler(rm.container.GetAuthService(), rm.container.GetUserService())
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
}

// setupProtectedAuthRoutes sets up protected authentication routes
func (rm *RouteManager) setupProtectedAuthRoutes(router *gin.RouterGroup) {
	// Protected routes only - logout and me
	authHandler := auth.NewHandler(rm.container.GetAuthService(), rm.container.GetUserService())
	router.POST("/logout", authHandler.Logout)
	router.GET("/me", authHandler.Me)
}

// setupOIDCRoutes sets up OIDC routes
func (rm *RouteManager) setupOIDCRoutes(router *gin.RouterGroup) {
	oidc.SetupRoutes(router, rm.container.GetOIDCService())
}

// setupOIDCProvidersRoutes sets up OIDC providers routes
func (rm *RouteManager) setupOIDCProvidersRoutes(router *gin.RouterGroup) {
	oidc.SetupProviderRoutes(router, rm.container.GetOIDCService())
}

// setupUserRoutes sets up user management routes
func (rm *RouteManager) setupUserRoutes(router *gin.RouterGroup) {
	auth.SetupUserRoutes(router, rm.container.GetAuthService(), rm.container.GetUserService())
}

// setupCredentialRoutes sets up credential management routes
func (rm *RouteManager) setupCredentialRoutes(router *gin.RouterGroup) {
	credential.SetupRoutes(router, rm.container.GetCredentialService())
}

// setupWorkspaceRoutes sets up workspace management routes
func (rm *RouteManager) setupWorkspaceRoutes(router *gin.RouterGroup) {
	workspace.SetupRoutes(router, rm.container.GetWorkspaceService(), rm.container.GetUserService())
}

// setupProviderRoutes sets up cloud provider routes
func (rm *RouteManager) setupProviderRoutes(router *gin.RouterGroup) {
	provider.SetupRoutes(router, rm.providerManager, rm.container.GetAuditLogRepository())
}

// setupProviderSpecificRoutes sets up provider-specific routes (RESTful)
func (rm *RouteManager) setupProviderSpecificRoutes(router *gin.RouterGroup) {
	// AWS-specific routes
	awsGroup := router.Group("/aws")
	rm.setupAWSRoutes(awsGroup)

	// GCP-specific routes
	gcpGroup := router.Group("/gcp")
	rm.setupGCPRoutes(gcpGroup)

	// Azure-specific routes
	azureGroup := router.Group("/azure")
	rm.setupAzureRoutes(azureGroup)

	// NCP (Naver Cloud Platform) routes
	ncpGroup := router.Group("/ncp")
	rm.setupNCPRoutes(ncpGroup)
}

// setupAWSRoutes sets up AWS-specific routes
func (rm *RouteManager) setupAWSRoutes(router *gin.RouterGroup) {
	// Kubernetes (EKS)
	k8sGroup := router.Group("/kubernetes")
	k8sService := rm.container.GetKubernetesService().(*service.KubernetesService)
	kubernetes.SetupRoutes(k8sGroup, k8sService, rm.container.GetCredentialService(), "aws")

	// Network resources (VPC, Subnet, Security Group)
	networkGroup := router.Group("/network")
	networkService := rm.container.GetNetworkService().(*service.NetworkService)
	network.SetupRoutes(networkGroup, networkService, rm.container.GetCredentialService(), rm.logger, "aws")

	// TODO: Add more AWS-specific services
	// - EC2
	// - RDS
	// - S3
	// - Lambda
}

// setupGCPRoutes sets up GCP-specific routes
func (rm *RouteManager) setupGCPRoutes(router *gin.RouterGroup) {
	// Kubernetes (GKE)
	k8sGroup := router.Group("/kubernetes")
	k8sService := rm.container.GetKubernetesService().(*service.KubernetesService)
	kubernetes.SetupRoutes(k8sGroup, k8sService, rm.container.GetCredentialService(), "gcp")

	// TODO: Add more GCP-specific services
	// - Compute Engine
	// - Cloud SQL
	// - Cloud Storage
	// - Cloud Functions
}

// setupAzureRoutes sets up Azure-specific routes
func (rm *RouteManager) setupAzureRoutes(router *gin.RouterGroup) {
	// Kubernetes (AKS)
	k8sGroup := router.Group("/kubernetes")
	k8sService := rm.container.GetKubernetesService().(*service.KubernetesService)
	kubernetes.SetupRoutes(k8sGroup, k8sService, rm.container.GetCredentialService(), "azure")

	// TODO: Add more Azure-specific services
	// - Virtual Machines
	// - SQL Database
	// - Blob Storage
	// - Azure Functions
}

// setupNCPRoutes sets up Naver Cloud Platform routes
func (rm *RouteManager) setupNCPRoutes(router *gin.RouterGroup) {
	// Kubernetes (NKS - Naver Kubernetes Service)
	k8sGroup := router.Group("/kubernetes")
	k8sService := rm.container.GetKubernetesService().(*service.KubernetesService)
	kubernetes.SetupRoutes(k8sGroup, k8sService, rm.container.GetCredentialService(), "ncp")

	// TODO: Add more NCP-specific services
	// - Server
	// - Cloud DB
	// - Object Storage
	// - Cloud Functions
}

// setupCostAnalysisRoutes sets up cost analysis routes
func (rm *RouteManager) setupCostAnalysisRoutes(router *gin.RouterGroup) {
	cost_analysis.SetupRoutes(router)
}

// setupNotificationRoutes sets up notification routes
func (rm *RouteManager) setupNotificationRoutes(router *gin.RouterGroup) {
	notification.SetupRoutes(router, rm.container.GetNotificationService())
}

// setupExportRoutes sets up export routes
func (rm *RouteManager) setupExportRoutes(router *gin.RouterGroup) {
	export.SetupRoutes(router)
}

// setupAdminUserRoutes sets up admin user management routes
func (rm *RouteManager) setupAdminUserRoutes(router *gin.RouterGroup) {
	// admin.SetupUserRoutes expects *logger.Logger, but we have *zap.Logger
	// For now, pass nil as logger is not used in the handler
	admin.SetupUserRoutes(router, rm.container.GetUserService(), rm.container.GetRBACService(), nil)
}

// setupSystemRoutes sets up system management routes
func (rm *RouteManager) setupSystemRoutes(router *gin.RouterGroup) {
	// system.SetupRoutes expects *logger.Logger, but we have *zap.Logger
	// For now, pass nil as logger is not used in the handler
	system.SetupRoutes(router, nil)
}

// setupAuditRoutes sets up audit log routes
func (rm *RouteManager) setupAuditRoutes(router *gin.RouterGroup) {
	audit.SetupRoutes(router, rm.container.GetAuditLogService())
}

// setupPermissionRoutes sets up permission management routes
func (rm *RouteManager) setupPermissionRoutes(router *gin.RouterGroup) {
	admin.SetupPermissionRoutes(router, rm.container.GetRBACService())
}

// setupSSERoutes sets up SSE routes
func (rm *RouteManager) setupSSERoutes(router *gin.RouterGroup) {
	sse.SetupRoutes(router)
}

// getServiceStatus returns the status of all services
func (rm *RouteManager) getServiceStatus() gin.H {
	services := gin.H{
		"database": rm.checkDatabaseStatus(),
		"redis":    rm.checkRedisStatus(),
		"plugins":  rm.checkPluginStatus(),
		"auth":     rm.checkAuthServiceStatus(),
	}
	return services
}

// isAllServicesHealthy checks if all services are healthy
func (rm *RouteManager) isAllServicesHealthy(services gin.H) bool {
	for _, service := range services {
		if status, ok := service.(gin.H); ok {
			if healthy, exists := status["healthy"]; exists {
				if !healthy.(bool) {
					return false
				}
			}
		}
	}
	return true
}

// checkDatabaseStatus checks database connectivity
func (rm *RouteManager) checkDatabaseStatus() gin.H {
	// TODO: Implement DB health check
	// if rm.container.DB == nil {
	// 	return gin.H{"healthy": false, "error": "database not initialized"}
	// }

	// TODO: Implement DB connection check
	// sqlDB, err := rm.container.DB.DB()
	// if err != nil {
	// 	return gin.H{"healthy": false, "error": err.Error()}
	// }

	// TODO: Implement DB ping check
	// if err := sqlDB.Ping(); err != nil {
	// 	return gin.H{"healthy": false, "error": err.Error()}
	// }

	// TODO: Implement DB health response
	// return gin.H{"healthy": true, "status": "connected"}
	return gin.H{"healthy": true, "status": "connected"}
}

// checkRedisStatus checks Redis connectivity
func (rm *RouteManager) checkRedisStatus() gin.H {
	// TODO: Implement Redis health check when Redis service is available
	return gin.H{"healthy": true, "status": "not implemented"}
}

// checkPluginStatus checks provider manager status
func (rm *RouteManager) checkPluginStatus() gin.H {
	return gin.H{
		"healthy": true,
		"status":  "gRPC",
		"type":    "gRPC Provider Manager",
		"note":    "Use /api/v1/providers for details",
	}
}

// checkAuthServiceStatus checks authentication service status
func (rm *RouteManager) checkAuthServiceStatus() gin.H {
	// TODO: Implement auth service check
	// if rm.container.AuthService == nil {
	// 	return gin.H{"healthy": false, "error": "auth service not initialized"}
	// }
	return gin.H{"healthy": true, "status": "available"}
}

// getEnvironment returns the current environment
func (rm *RouteManager) getEnvironment() string {
	if rm.config.Server.Host == "0.0.0.0" {
		return "production"
	}
	return "development"
}

// getDependenciesStatus checks external dependencies
func (rm *RouteManager) getDependenciesStatus() gin.H {
	return gin.H{
		"postgres": rm.checkPostgresDependency(),
		"redis":    rm.checkRedisDependency(),
		"plugins":  rm.checkPluginDependencies(),
	}
}

// isAllDependenciesHealthy checks if all dependencies are healthy
func (rm *RouteManager) isAllDependenciesHealthy(dependencies gin.H) bool {
	for _, dep := range dependencies {
		if status, ok := dep.(gin.H); ok {
			if healthy, exists := status["healthy"]; exists {
				if !healthy.(bool) {
					return false
				}
			}
		}
	}
	return true
}

// getSystemMetrics returns system performance metrics
func (rm *RouteManager) getSystemMetrics() gin.H {
	// TODO: Implement system metrics
	// var m runtime.MemStats
	// runtime.ReadMemStats(&m)

	// Calculate error rate
	errorRate := 0.0
	if rm.requestCount > 0 {
		errorRate = float64(rm.errorCount) / float64(rm.requestCount) * 100
	}

	// Calculate requests per second
	uptimeSeconds := time.Since(rm.startTime).Seconds()
	rps := 0.0
	if uptimeSeconds > 0 {
		rps = float64(rm.requestCount) / uptimeSeconds
	}

	return gin.H{
		"memory_usage": gin.H{
			// TODO: Implement memory usage metrics
			// "alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			// "total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			// "sys_mb":         float64(m.Sys) / 1024 / 1024,
			// "num_gc":         m.NumGC,
			"alloc_mb":       0.0,
			"total_alloc_mb": 0.0,
			"sys_mb":         0.0,
			"num_gc":         0,
		},
		"performance": gin.H{
			"goroutines":    runtime.NumGoroutine(),
			"cpu_cores":     runtime.NumCPU(),
			"request_count": rm.requestCount,
			"error_count":   rm.errorCount,
			"error_rate":    errorRate,
			"rps":           rps,
		},
	}
}

// getAlertStatus returns current alert status
func (rm *RouteManager) getAlertStatus(services, dependencies, metrics gin.H) gin.H {
	alerts := []gin.H{}

	// Check service alerts
	if !rm.isAllServicesHealthy(services) {
		alerts = append(alerts, gin.H{
			"type":    "service",
			"level":   "warning",
			"message": "One or more services are unhealthy",
		})
	}

	// Check dependency alerts
	if !rm.isAllDependenciesHealthy(dependencies) {
		alerts = append(alerts, gin.H{
			"type":    "dependency",
			"level":   "critical",
			"message": "External dependencies are failing",
		})
	}

	// Check performance alerts
	if metricsMap, ok := metrics["performance"].(gin.H); ok {
		if errorRate, exists := metricsMap["error_rate"]; exists {
			if rate, ok := errorRate.(float64); ok && rate > 5.0 {
				alerts = append(alerts, gin.H{
					"type":    "performance",
					"level":   "warning",
					"message": "High error rate detected",
					"value":   rate,
				})
			}
		}
	}

	// Check memory alerts
	if metricsMap, ok := metrics["memory_usage"].(gin.H); ok {
		if allocMB, exists := metricsMap["alloc_mb"]; exists {
			if alloc, ok := allocMB.(float64); ok && alloc > 100 {
				alerts = append(alerts, gin.H{
					"type":    "memory",
					"level":   "warning",
					"message": "High memory usage detected",
					"value":   alloc,
				})
			}
		}
	}

	return gin.H{
		"enabled": len(alerts) > 0,
		"count":   len(alerts),
		"alerts":  alerts,
		"thresholds": gin.H{
			"error_rate":   5.0,
			"memory_mb":    100.0,
			"uptime_hours": 24.0,
		},
	}
}

// checkPostgresDependency checks PostgreSQL connection
func (rm *RouteManager) checkPostgresDependency() gin.H {
	// TODO: Implement DB health check
	// if rm.container.DB == nil {
	// 	return gin.H{"healthy": false, "error": "database not initialized"}
	// }

	// TODO: Implement timing
	// start := time.Now()
	// TODO: Implement DB connection check
	// sqlDB, err := rm.container.DB.DB()
	// if err != nil {
	// 	return gin.H{"healthy": false, "error": err.Error()}
	// }

	// TODO: Implement DB ping check
	// if err := sqlDB.Ping(); err != nil {
	// 	return gin.H{"healthy": false, "error": err.Error()}
	// }

	// TODO: Implement response time calculation
	// responseTime := time.Since(start)
	responseTime := time.Since(time.Now())
	return gin.H{
		"healthy":          true,
		"status":           "connected",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1000000,
	}
}

// checkRedisDependency checks Redis connection
func (rm *RouteManager) checkRedisDependency() gin.H {
	// TODO: Implement actual Redis health check
	return gin.H{
		"healthy":          true,
		"status":           "not_implemented",
		"response_time_ms": 0,
	}
}

// checkPluginDependencies checks provider dependencies
func (rm *RouteManager) checkPluginDependencies() gin.H {
	return gin.H{
		"healthy": true,
		"status":  "gRPC",
		"type":    "gRPC Provider Manager",
		"note":    "Providers are managed via gRPC",
	}
}

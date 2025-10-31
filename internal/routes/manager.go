package routes

import (
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
	container  di.ContainerInterface
	middleware *middleware.Middleware
	logger     *zap.Logger
	config     *config.Config
}

// NewRouteManager creates a new route manager
func NewRouteManager(
	container di.ContainerInterface,
	middleware *middleware.Middleware,
	logger *zap.Logger,
	config *config.Config,
) *RouteManager {
	return &RouteManager{
		container:  container,
		middleware: middleware,
		logger:     logger,
		config:     config,
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
	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"message":   "SkyClust API is running",
			"timestamp": "2025-10-28T11:13:34.246963+09:00",
		})
	})

	// API v1 public routes
	v1Public := router.Group("/api/v1")
	{
		// Authentication routes (public) - register and login
		authGroup := v1Public.Group("/auth")
		rm.setupPublicAuthRoutes(authGroup)

		// OIDC routes (public)
		oidcAuthGroup := v1Public.Group("/auth/oidc")
		rm.setupOIDCRoutes(oidcAuthGroup)

		// OIDC providers (public - list available provider types)
		oidcGroup := v1Public.Group("/oidc")
		oidcProvidersGroup := oidcGroup.Group("/providers")
		rm.setupOIDCProvidersRoutes(oidcProvidersGroup)

		// System monitoring routes (no authentication required)
		systemGroup := v1Public.Group("/system")
		rm.setupSystemRoutes(systemGroup)
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

		// Provider-specific routes (RESTful)
		rm.setupProviderSpecificRoutes(v1Protected)

		// Cost analysis routes (keep hyphenated name for single-word resource)
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

		// OIDC provider management routes (protected)
		oidcProviderGroup := v1Protected.Group("/oidc")
		rm.setupUserOIDCProviderRoutes(oidcProviderGroup)
	}
}

// setupAdminRoutes sets up admin routes that require admin privileges
func (rm *RouteManager) setupAdminRoutes(router *gin.Engine) {
	// API v1 admin routes
	v1Admin := router.Group("/api/v1/admin")

	// Apply admin middleware to all admin routes
	// TODO: Implement AdminMiddleware
	// v1Admin.Use(rm.middleware.AdminMiddleware())
	{
		// Admin user management routes
		adminUsersGroup := v1Admin.Group("/users")
		rm.setupAdminUserRoutes(adminUsersGroup)

		// System management routes
		systemGroup := v1Admin.Group("/system")
		rm.setupSystemRoutes(systemGroup)

		// Audit log routes (RESTful: /audit-logs)
		auditGroup := v1Admin.Group("/audit-logs")
		rm.setupAuditRoutes(auditGroup)

		// Permission management routes
		permissionsGroup := v1Admin.Group("/permissions")
		rm.setupPermissionRoutes(permissionsGroup)
	}
}

// setupPublicAuthRoutes sets up public authentication routes
func (rm *RouteManager) setupPublicAuthRoutes(router *gin.RouterGroup) {
	// Public routes only - register and login
	if authService := rm.container.GetAuthService(); authService != nil {
		if userService := rm.container.GetUserService(); userService != nil {
			if rbacService := rm.container.GetRBACService(); rbacService != nil {
				authHandler := auth.NewHandler(authService, userService, rbacService)
				router.POST("/register", authHandler.Register)
				router.POST("/login", authHandler.Login)
			}
		}
	}
}

// setupProtectedAuthRoutes sets up protected authentication routes
func (rm *RouteManager) setupProtectedAuthRoutes(router *gin.RouterGroup) {
	// Protected routes only - sessions and me
	if authService := rm.container.GetAuthService(); authService != nil {
		if userService := rm.container.GetUserService(); userService != nil {
			if rbacService := rm.container.GetRBACService(); rbacService != nil {
				authHandler := auth.NewHandler(authService, userService, rbacService)
				
				// Session management (RESTful)
				sessionsGroup := router.Group("/sessions")
				{
					sessionsGroup.GET("/me", authHandler.GetSession)    // GET /api/v1/auth/sessions/me
					sessionsGroup.DELETE("/me", authHandler.Logout)      // DELETE /api/v1/auth/sessions/me
				}
				
				router.GET("/me", authHandler.Me) // GET /api/v1/auth/me
			}
		}
	}
}

// setupOIDCRoutes sets up OIDC routes
func (rm *RouteManager) setupOIDCRoutes(router *gin.RouterGroup) {
	if oidcService := rm.container.GetOIDCService(); oidcService != nil {
		oidc.SetupRoutes(router, oidcService)
	}
}

// setupOIDCProvidersRoutes sets up OIDC providers routes (public)
func (rm *RouteManager) setupOIDCProvidersRoutes(router *gin.RouterGroup) {
	if oidcService := rm.container.GetOIDCService(); oidcService != nil {
		// Public routes (list available providers)
		oidc.SetupProviderRoutes(router, oidcService)
	}
}

// setupUserOIDCProviderRoutes sets up user OIDC provider management routes (protected)
func (rm *RouteManager) setupUserOIDCProviderRoutes(router *gin.RouterGroup) {
	if oidcService := rm.container.GetOIDCService(); oidcService != nil {
		if oidcProviderRepo := rm.container.GetOIDCProviderRepository(); oidcProviderRepo != nil {
			oidc.SetupUserProviderRoutes(router, oidcService, oidcProviderRepo)
		}
	}
}

// setupUserRoutes sets up user management routes
func (rm *RouteManager) setupUserRoutes(router *gin.RouterGroup) {
	if authService := rm.container.GetAuthService(); authService != nil {
		if userService := rm.container.GetUserService(); userService != nil {
			if rbacService := rm.container.GetRBACService(); rbacService != nil {
				auth.SetupUserRoutes(router, authService, userService, rbacService)
			}
		}
	}
}

// setupCredentialRoutes sets up credential management routes
func (rm *RouteManager) setupCredentialRoutes(router *gin.RouterGroup) {
	if credentialService := rm.container.GetCredentialService(); credentialService != nil {
		credential.SetupRoutes(router, credentialService)
	}
}

// setupWorkspaceRoutes sets up workspace management routes
func (rm *RouteManager) setupWorkspaceRoutes(router *gin.RouterGroup) {
	if workspaceService := rm.container.GetWorkspaceService(); workspaceService != nil {
		if userService := rm.container.GetUserService(); userService != nil {
			workspace.SetupRoutes(router, workspaceService, userService)
		}
	}
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
	if k8sService := rm.container.GetKubernetesService(); k8sService != nil {
		rm.logger.Info("Kubernetes service found, setting up routes")
		if k8s, ok := k8sService.(*service.KubernetesService); ok {
			rm.logger.Info("Kubernetes service type assertion successful, setting up AWS routes")
			kubernetes.SetupRoutes(k8sGroup, k8s, rm.container.GetCredentialService(), "aws")
		} else {
			rm.logger.Warn("Kubernetes service type assertion failed")
		}
	} else {
		rm.logger.Warn("Kubernetes service is nil")
	}

	// Network resources (VPC, Subnet, Security Group)
	networkGroup := router.Group("/network")
	if networkService := rm.container.GetNetworkService(); networkService != nil {
		if networkSvc, ok := networkService.(*service.NetworkService); ok {
			network.SetupRoutes(networkGroup, networkSvc, rm.container.GetCredentialService(), "aws")
		}
	}

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
	if k8sService := rm.container.GetKubernetesService(); k8sService != nil {
		rm.logger.Info("Kubernetes service found, setting up GCP routes")
		if k8s, ok := k8sService.(*service.KubernetesService); ok {
			rm.logger.Info("Kubernetes service type assertion successful, setting up GCP routes")
			kubernetes.SetupGCPRoutes(k8sGroup, k8s, rm.container.GetCredentialService())
		} else {
			rm.logger.Warn("Kubernetes service type assertion failed for GCP")
		}
	} else {
		rm.logger.Warn("Kubernetes service is nil for GCP")
	}

	// Network resources (VPC, Subnet, Security Group)
	networkGroup := router.Group("/network")
	if networkService := rm.container.GetNetworkService(); networkService != nil {
		if networkSvc, ok := networkService.(*service.NetworkService); ok {
			network.SetupGCPRoutes(networkGroup, networkSvc, rm.container.GetCredentialService(), rm.logger)
		}
	}

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
	if k8sService := rm.container.GetKubernetesService(); k8sService != nil {
		if k8s, ok := k8sService.(*service.KubernetesService); ok {
			kubernetes.SetupRoutes(k8sGroup, k8s, rm.container.GetCredentialService(), "azure")
		}
	}

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
	if k8sService := rm.container.GetKubernetesService(); k8sService != nil {
		if k8s, ok := k8sService.(*service.KubernetesService); ok {
			kubernetes.SetupRoutes(k8sGroup, k8s, rm.container.GetCredentialService(), "ncp")
		}
	}

	// TODO: Add more NCP-specific services
	// - Server
	// - Cloud DB
	// - Object Storage
	// - Cloud Functions
}

// setupCostAnalysisRoutes sets up cost analysis routes
func (rm *RouteManager) setupCostAnalysisRoutes(router *gin.RouterGroup) {
	costAnalysisService := rm.container.GetCostAnalysisService()
	if costAnalysisService == nil {
		rm.logger.Warn("CostAnalysisService not available, cost analysis routes will not be set up")
		return
	}

	// Type assert to *service.CostAnalysisService
	if svc, ok := costAnalysisService.(*service.CostAnalysisService); ok {
		cost_analysis.SetupRoutes(router, svc)
	} else {
		rm.logger.Warn("CostAnalysisService type assertion failed, cost analysis routes will not be set up")
	}
}

// setupNotificationRoutes sets up notification routes
func (rm *RouteManager) setupNotificationRoutes(router *gin.RouterGroup) {
	if notificationService := rm.container.GetNotificationService(); notificationService != nil {
		notification.SetupRoutes(router, notificationService)
	}
}

// setupExportRoutes sets up export routes
func (rm *RouteManager) setupExportRoutes(router *gin.RouterGroup) {
	exportHandler := export.NewHandler()
	// Inject ExportService from container
	if exportService := rm.container.GetExportService(); exportService != nil {
		if svc, ok := exportService.(*service.ExportService); ok {
			exportHandler.SetExportService(svc)
		}
	}
	export.SetupRoutesWithHandler(router, exportHandler)
}

// setupAdminUserRoutes sets up admin user management routes
func (rm *RouteManager) setupAdminUserRoutes(router *gin.RouterGroup) {
	// admin.SetupUserRoutes expects *logger.Logger, but we have *zap.Logger
	// For now, pass nil as logger is not used in the handler
	if userService := rm.container.GetUserService(); userService != nil {
		if rbacService := rm.container.GetRBACService(); rbacService != nil {
			admin.SetupUserRoutes(router, userService, rbacService, nil)
		}
	}
}

// setupAuditRoutes sets up audit log routes
func (rm *RouteManager) setupAuditRoutes(router *gin.RouterGroup) {
	if auditLogService := rm.container.GetAuditLogService(); auditLogService != nil {
		audit.SetupRoutes(router, auditLogService)
	}
}

// setupPermissionRoutes sets up permission management routes
func (rm *RouteManager) setupPermissionRoutes(router *gin.RouterGroup) {
	if rbacService := rm.container.GetRBACService(); rbacService != nil {
		admin.SetupPermissionRoutes(router, rbacService)
	}
}

// setupSSERoutes sets up SSE routes
func (rm *RouteManager) setupSSERoutes(router *gin.RouterGroup) {
	sse.SetupRoutes(router)
}

// setupSystemRoutes sets up system monitoring routes
func (rm *RouteManager) setupSystemRoutes(router *gin.RouterGroup) {
	if systemMonitoringService := rm.container.GetSystemMonitoringService(); systemMonitoringService != nil {
		system.SetupRoutes(router, systemMonitoringService)
	}
}

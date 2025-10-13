package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"skyclust/internal/container"
	httpdelivery "skyclust/internal/delivery/http"
	"skyclust/internal/delivery/http/routes"
	"skyclust/internal/plugin"
	"skyclust/pkg/config"
	"skyclust/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile string
	pluginDir  string
	port       string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cmp-server",
		Short: "Cloud Management Portal Server",
		Long:  "A plugin-based cloud management platform",
		Run:   runServer,
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")
	rootCmd.Flags().StringVarP(&pluginDir, "plugins", "p", "plugins", "Plugin directory path")
	rootCmd.Flags().StringVarP(&port, "port", "P", "8081", "Server port")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// getConfigFileByEnvironment returns the appropriate config file based on environment
func getConfigFileByEnvironment() string {
	env := strings.ToLower(os.Getenv("CMP_ENV"))

	switch env {
	case "production", "prod":
		return "configs/config.prod.yaml"
	case "development", "dev", "":
		return "configs/config.dev.yaml"
	default:
		return "configs/config.dev.yaml" // Default to development
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Determine config file based on environment
	if configFile == "config.yaml" {
		configFile = getConfigFileByEnvironment()
	}

	// Load configuration
	var cfg *config.Config
	var err error

	// Try to load from YAML file first
	if configFile != "" && configFile != "config.yaml" {
		cfg, err = config.LoadConfigFromFile(configFile)
		if err != nil {
			log.Printf("Warning: failed to load config from file %s: %v", configFile, err)
			log.Println("Falling back to environment variables")
			cfg = config.LoadConfig()
		} else {
			log.Printf("Loaded configuration from %s", configFile)
		}
	} else {
		// Load from environment variables
		cfg = config.LoadConfig()
	}

	// Override with command line flags
	if pluginDir != "" {
		cfg.Plugins.Directory = pluginDir
	}
	if port != "" {
		// Parse port from string to int
		if portInt, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = portInt
		}
	}

	// Create context
	ctx := context.Background()

	// Initialize dependency injection container
	appContainer, err := container.NewContainer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer appContainer.Close()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Initialize plugin manager
	pluginManager := plugin.NewManager()

	// Load plugins
	if err := pluginManager.LoadPlugins(cfg.Plugins.Directory); err != nil {
		log.Printf("Warning: failed to load plugins: %v", err)
	}

	// TODO: Initialize plugins with configuration
	// This would be done after loading plugins
	// for _, providerName := range pluginManager.ListProviders() {
	//     if err := pluginManager.InitializeProvider(providerName, cfg.Providers[providerName]); err != nil {
	//         log.Printf("Warning: failed to initialize provider %s: %v", providerName, err)
	//     }
	// }

	// Setup router
	router := setupRouter(appContainer, pluginManager, cfg, logger)

	// Start server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Starting CMP server on port %d", cfg.Server.Port)
	log.Printf("Plugin directory: %s", cfg.Plugins.Directory)
	log.Printf("Loaded providers: %v", pluginManager.ListProviders())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(container *container.Container, pluginManager *plugin.Manager, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create error handler
	errorHandler := httpdelivery.NewErrorHandler(logger)

	// Middleware
	router.Use(errorHandler.RequestIDMiddleware())
	router.Use(errorHandler.LoggingMiddleware())
	router.Use(errorHandler.RecoveryMiddleware())
	router.Use(errorHandler.ErrorHandlerMiddleware())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// SSE endpoint (인증 필요)
	router.GET("/api/events", func(c *gin.Context) {
		// SSE 핸들러는 별도 구현 필요
		// TODO: Add authentication middleware
		c.JSON(http.StatusNotImplemented, gin.H{"message": "SSE endpoint will be implemented"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		// RESTful authentication routes
		routes.SetupRESTfulRoutes(router, container.AuthService, container.UserService, container.OIDCService, container.LogoutService)

		// Protected routes (require authentication)
		protected := api.Group("")

		// Create authentication middleware
		authMiddleware := middleware.NewMiddleware(
			logger,
			&middleware.MiddlewareConfig{
				APIVersion:        "1.0",
				DefaultPageSize:   10,
				MaxPageSize:       100,
				LoggingEnabled:    true,
				StructuredLogging: true,
			},
			nil,                   // rateLimiter
			container.AuthService, // authService
			nil,                   // rbacService
			nil,                   // auditLogger
		)

		// Apply authentication middleware
		protected.Use(authMiddleware.AuthMiddleware())
		{
			// Credential management
			credentials := protected.Group("/credentials")
			routes.SetupCredentialRoutes(credentials, container.CredentialService)

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			routes.SetupWorkspaceRoutes(workspaces, container.WorkspaceService, container.UserService)

			// Provider routes
			providers := protected.Group("/providers")
			routes.SetupProviderRoutes(providers, pluginManager, container.AuditLogRepo)

			// Cost analysis routes
			costAnalysis := protected.Group("/cost-analysis")
			costAnalysisHandler := httpdelivery.NewCostAnalysisHandler(logger, container.CostAnalysisService)
			routes.SetupCostAnalysisRoutes(costAnalysis, costAnalysisHandler)

			// Notification routes
			notifications := protected.Group("/notifications")
			notificationHandler := httpdelivery.NewNotificationHandler(container.NotificationService, logger)
			routes.SetupNotificationRoutes(notifications, notificationHandler)

			// Export routes
			exports := protected.Group("/exports")
			exportHandler := httpdelivery.NewExportHandler(logger, container.ExportService)
			routes.SetupExportRoutes(exports, exportHandler)

			// User management routes
			users := protected.Group("/users")
			routes.SetupUsersRoutes(users, container.AuthService, container.UserService)
		}
	}

	return router
}

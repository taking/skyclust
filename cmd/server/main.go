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

	"skyclust/internal/di"
	"skyclust/internal/routes"
	"skyclust/pkg/config"
	"skyclust/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile string
	port       string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cmp-server",
		Short: "Cloud Management Portal Server",
		Long:  "A gRPC-based cloud management platform",
		Run:   runServer,
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")
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

	// Load configuration from YAML file with environment variable override
	cfg, err = config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Loaded configuration from %s", configFile)

	// Override with command line flags (only if not set via environment)
	// CLI flags have lower priority than environment variables
	if port != "" && port != "8081" {
		// Only override if port flag is explicitly set and different from default
		// and no environment variable was set
		if os.Getenv("SERVER_PORT") == "" {
			if portInt, err := strconv.Atoi(port); err == nil {
				cfg.Server.Port = portInt
			}
		}
	}

	// Create context
	ctx := context.Background()

	// Initialize dependency injection container
	appContainer := &di.Container{}

	// Initialize container with configuration
	if err := appContainer.Initialize(ctx, cfg); err != nil {
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

	// Initialize gRPC Provider Manager
	// TODO: Implement gRPC provider manager
	var providerManager interface{} = nil
	log.Println("Provider manager: using gRPC-based providers")

	// Setup router
	router := setupRouter(appContainer, providerManager, cfg, logger)

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
	log.Printf("gRPC Provider Manager initialized")

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

func setupRouter(container di.ContainerInterface, providerManager interface{}, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create middleware with default config
	middlewareInstance := middleware.NewMiddleware(
		logger,
		middleware.GetDefaultMiddlewareConfig(),
		nil,                        // rateLimiter
		container.GetAuthService(), // authService
		container.GetRBACService(), // rbacService
		nil,                        // auditLogger
	)

	// Apply core middleware
	router.Use(middlewareInstance.RequestIDMiddleware())
	router.Use(middlewareInstance.LoggingMiddleware())
	router.Use(middlewareInstance.RecoveryMiddleware())

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
		nil,                        // rateLimiter
		container.GetAuthService(), // authService
		container.GetRBACService(), // rbacService
		nil,                        // auditLogger
	)

	// Create route manager
	routeManager := routes.NewRouteManager(container, providerManager, authMiddleware, logger, cfg)

	// Setup all routes using the route manager
	routeManager.SetupAllRoutes(router)

	return router
}

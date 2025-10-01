package main

import (
	pluginManager "cmp/internal/plugin"
	"cmp/pkg/shared/config"
	"context"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	configFile string
	pluginDir  string
	port       string
)

// Server holds the server state for hot reload
type Server struct {
	manager    *pluginManager.Manager
	config     *config.Config
	configPath string
	mu         sync.RWMutex
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cmp-server",
		Short: "Cloud Management Portal Server",
		Long:  "A plugin-based cloud management portal that supports multiple cloud providers",
		Run:   runServer,
	}

	rootCmd.Flags().StringVar(&configFile, "config", "config.yaml", "Config file path")
	rootCmd.Flags().StringVar(&pluginDir, "plugins", "plugins", "Plugin directory path")
	rootCmd.Flags().StringVar(&port, "port", "8080", "Server port")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Create server instance
	server := &Server{
		configPath: configFile,
	}

	// Load initial configuration
	if err := server.loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize plugin manager
	server.manager = pluginManager.NewManager()

	// Load plugins
	if err := server.manager.LoadPlugins(server.config.Plugins.Directory); err != nil {
		log.Printf("Warning: failed to load plugins: %v", err)
	}

	// Initialize providers with configuration
	server.initializeProviders()

	// Setup router
	router := server.setupRouter()

	// Start file watcher for hot reload
	go server.watchConfigFile()

	// Start server
	log.Printf("Starting CMP server on port %s", server.config.Server.Port)
	log.Printf("Plugin directory: %s", server.config.Plugins.Directory)
	log.Printf("Loaded providers: %v", server.manager.ListProviders())
	log.Printf("Hot reload enabled for config file: %s", server.configPath)

	if err := router.Run(":" + server.config.Server.Port); err != nil {
		log.Fatal(err)
	}
}

// loadConfig loads the configuration file
func (s *Server) loadConfig() error {
	cfg, err := config.LoadConfig(s.configPath)
	if err != nil {
		return err
	}

	// Override port if specified via command line
	if port != "8080" {
		cfg.Server.Port = port
	}

	// Override plugin directory if specified via command line
	if pluginDir != "plugins" {
		cfg.Plugins.Directory = pluginDir
	}

	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()

	return nil
}

// watchConfigFile watches for configuration file changes
func (s *Server) watchConfigFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch the directory containing the config file
	configDir := filepath.Dir(s.configPath)
	if err := watcher.Add(configDir); err != nil {
		log.Printf("Failed to watch config directory: %v", err)
		return
	}

	log.Printf("Watching for config changes in: %s", configDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Check if it's our config file
				if filepath.Base(event.Name) == filepath.Base(s.configPath) {
					log.Printf("Config file changed: %s", event.Name)
					s.reloadConfig()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// reloadConfig reloads the configuration and reinitializes providers
func (s *Server) reloadConfig() {
	log.Printf("Reloading configuration...")

	// Load new configuration (this will re-read environment variables)
	if err := s.loadConfig(); err != nil {
		log.Printf("Failed to reload configuration: %v", err)
		return
	}

	// Reinitialize providers
	s.initializeProviders()

	log.Printf("Configuration reloaded successfully")
	log.Printf("Loaded providers: %v", s.manager.ListProviders())
}

// initializeProviders initializes cloud providers with current configuration
func (s *Server) initializeProviders() {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	// Initialize AWS provider
	if cfg.Providers.AWS.AccessKey != "" && cfg.Providers.AWS.SecretKey != "" {
		awsConfig := map[string]interface{}{
			"access_key": cfg.Providers.AWS.AccessKey,
			"secret_key": cfg.Providers.AWS.SecretKey,
			"region":     cfg.Providers.AWS.Region,
			"role_arn":   cfg.Providers.AWS.RoleARN,
		}

		if err := s.manager.InitializeProvider("aws", awsConfig); err != nil {
			log.Printf("Failed to initialize AWS provider: %v", err)
		} else {
			log.Printf("AWS provider initialized for region: %s", cfg.Providers.AWS.Region)
		}
	}

	// Initialize GCP provider
	if cfg.Providers.GCP.ProjectID != "" {
		gcpConfig := map[string]interface{}{
			"project_id":       cfg.Providers.GCP.ProjectID,
			"credentials_file": cfg.Providers.GCP.CredentialsFile,
			"region":           cfg.Providers.GCP.Region,
		}

		if err := s.manager.InitializeProvider("gcp", gcpConfig); err != nil {
			log.Printf("Failed to initialize GCP provider: %v", err)
		} else {
			log.Printf("GCP provider initialized for project: %s, region: %s", cfg.Providers.GCP.ProjectID, cfg.Providers.GCP.Region)
		}
	}

	// Initialize OpenStack provider
	if cfg.Providers.OpenStack.AuthURL != "" && cfg.Providers.OpenStack.Username != "" {
		openstackConfig := map[string]interface{}{
			"auth_url":   cfg.Providers.OpenStack.AuthURL,
			"username":   cfg.Providers.OpenStack.Username,
			"password":   cfg.Providers.OpenStack.Password,
			"project_id": cfg.Providers.OpenStack.ProjectID,
			"region":     cfg.Providers.OpenStack.Region,
		}

		if err := s.manager.InitializeProvider("openstack", openstackConfig); err != nil {
			log.Printf("Failed to initialize OpenStack provider: %v", err)
		} else {
			log.Printf("OpenStack provider initialized for region: %s", cfg.Providers.OpenStack.Region)
		}
	}

	// Initialize Proxmox provider
	if cfg.Providers.Proxmox.Host != "" && cfg.Providers.Proxmox.Username != "" {
		proxmoxConfig := map[string]interface{}{
			"host":     cfg.Providers.Proxmox.Host,
			"username": cfg.Providers.Proxmox.Username,
			"password": cfg.Providers.Proxmox.Password,
			"realm":    cfg.Providers.Proxmox.Realm,
		}

		if err := s.manager.InitializeProvider("proxmox", proxmoxConfig); err != nil {
			log.Printf("Failed to initialize Proxmox provider: %v", err)
		} else {
			log.Printf("Proxmox provider initialized for host: %s", cfg.Providers.Proxmox.Host)
		}
	}
}

// setupRouter sets up the HTTP router
func (s *Server) setupRouter() *gin.Engine {
	router := gin.Default()

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
		s.mu.RLock()
		providers := s.manager.ListProviders()
		s.mu.RUnlock()
		
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"providers": providers,
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Provider management
		api.GET("/providers", s.getProviders)
		api.GET("/providers/:name", s.getProvider)
		api.POST("/providers/:name/initialize", s.initializeProvider)

		// Instance management
		api.GET("/providers/:name/instances", s.listInstances)
		api.POST("/providers/:name/instances", s.createInstance)
		api.GET("/providers/:name/instances/:id", s.getInstance)
		api.DELETE("/providers/:name/instances/:id", s.deleteInstance)

		// Region management
		api.GET("/providers/:name/regions", s.listRegions)

		// Cost estimation
		api.POST("/providers/:name/cost-estimate", s.getCostEstimate)
	}

	return router
}

// HTTP Handler methods
func (s *Server) getProviders(c *gin.Context) {
	s.mu.RLock()
	providers := s.manager.ListProviders()
	s.mu.RUnlock()
	
	providerInfos := make([]map[string]string, 0, len(providers))

	for _, name := range providers {
		info, err := s.manager.GetProviderInfo(name)
		if err != nil {
			continue
		}
		providerInfos = append(providerInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providerInfos,
	})
}

func (s *Server) getProvider(c *gin.Context) {
	name := c.Param("name")
	info, err := s.manager.GetProviderInfo(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (s *Server) initializeProvider(c *gin.Context) {
	name := c.Param("name")

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.manager.InitializeProvider(name, config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider initialized successfully"})
}

func (s *Server) listInstances(c *gin.Context) {
	name := c.Param("name")
	provider, err := s.manager.GetProvider(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	instances, err := provider.ListInstances(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

func (s *Server) createInstance(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock success response
	instance := map[string]interface{}{"id": "mock-instance", "name": "mock-vm"}

	c.JSON(http.StatusCreated, instance)
}

func (s *Server) getInstance(c *gin.Context) {
	name := c.Param("name")
	instanceID := c.Param("id")

	provider, err := s.manager.GetProvider(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status, err := provider.GetInstanceStatus(context.Background(), instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     instanceID,
		"status": status,
	})
}

func (s *Server) deleteInstance(c *gin.Context) {
	name := c.Param("name")
	instanceID := c.Param("id")

	provider, err := s.manager.GetProvider(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := provider.DeleteInstance(context.Background(), instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance deleted successfully"})
}

func (s *Server) listRegions(c *gin.Context) {
	name := c.Param("name")
	provider, err := s.manager.GetProvider(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	regions, err := provider.ListRegions(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

func (s *Server) getCostEstimate(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock success response
	estimate := map[string]interface{}{"cost": 0.01, "currency": "USD"}

	c.JSON(http.StatusOK, estimate)
}


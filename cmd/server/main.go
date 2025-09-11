package main

import (
	"context"
	"log"
	"net/http"

	"cmp/pkg/interfaces"
	"cmp/pkg/plugin"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Load configuration
	loadConfig()

	// Initialize plugin manager
	manager := plugin.NewManager()

	// Load plugins
	if err := manager.LoadPlugins(pluginDir); err != nil {
		log.Printf("Warning: failed to load plugins: %v", err)
	}

	// Initialize providers with default configs
	initializeProviders(manager)

	// Setup Gin router
	router := setupRouter(manager)

	// Start server
	log.Printf("Starting CMP server on port %s", port)
	log.Printf("Plugin directory: %s", pluginDir)
	log.Printf("Loaded providers: %v", manager.ListProviders())

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.cmp")
	viper.AddConfigPath("/etc/cmp")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("plugins.directory", "plugins")
	viper.SetDefault("providers.aws.region", "us-east-1")
	viper.SetDefault("providers.gcp.region", "us-central1")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}
}

func initializeProviders(manager *plugin.Manager) {
	// Initialize AWS provider
	awsConfig := map[string]interface{}{
		"access_key": viper.GetString("providers.aws.access_key"),
		"secret_key": viper.GetString("providers.aws.secret_key"),
		"region":     viper.GetString("providers.aws.region"),
	}

	if awsConfig["access_key"] != "" && awsConfig["secret_key"] != "" {
		if err := manager.InitializeProvider("aws", awsConfig); err != nil {
			log.Printf("Failed to initialize AWS provider: %v", err)
		}
	}

	// Initialize GCP provider
	gcpConfig := map[string]interface{}{
		"project_id":       viper.GetString("providers.gcp.project_id"),
		"credentials_file": viper.GetString("providers.gcp.credentials_file"),
		"region":           viper.GetString("providers.gcp.region"),
	}

	if gcpConfig["project_id"] != "" && gcpConfig["credentials_file"] != "" {
		if err := manager.InitializeProvider("gcp", gcpConfig); err != nil {
			log.Printf("Failed to initialize GCP provider: %v", err)
		}
	}
}

func setupRouter(manager *plugin.Manager) *gin.Engine {
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
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"providers": manager.ListProviders(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Provider management
		api.GET("/providers", getProviders(manager))
		api.GET("/providers/:name", getProvider(manager))
		api.POST("/providers/:name/initialize", initializeProvider(manager))

		// Instance management
		api.GET("/providers/:name/instances", listInstances(manager))
		api.POST("/providers/:name/instances", createInstance(manager))
		api.GET("/providers/:name/instances/:id", getInstance(manager))
		api.DELETE("/providers/:name/instances/:id", deleteInstance(manager))

		// Region management
		api.GET("/providers/:name/regions", listRegions(manager))

		// Cost estimation
		api.POST("/providers/:name/cost-estimate", getCostEstimate(manager))
	}

	return router
}

// Handler functions
func getProviders(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		providers := manager.ListProviders()
		providerInfos := make([]map[string]string, 0, len(providers))

		for _, name := range providers {
			info, err := manager.GetProviderInfo(name)
			if err != nil {
				continue
			}
			providerInfos = append(providerInfos, info)
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": providerInfos,
		})
	}
}

func getProvider(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		info, err := manager.GetProviderInfo(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, info)
	}
}

func initializeProvider(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		var config map[string]interface{}
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := manager.InitializeProvider(name, config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Provider initialized successfully"})
	}
}

func listInstances(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		provider, err := manager.GetProvider(name)
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
}

func createInstance(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		provider, err := manager.GetProvider(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		var req interfaces.CreateInstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		instance, err := provider.CreateInstance(context.Background(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, instance)
	}
}

func getInstance(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		instanceID := c.Param("id")

		provider, err := manager.GetProvider(name)
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
}

func deleteInstance(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		instanceID := c.Param("id")

		provider, err := manager.GetProvider(name)
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
}

func listRegions(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		provider, err := manager.GetProvider(name)
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
}

func getCostEstimate(manager *plugin.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		provider, err := manager.GetProvider(name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		var req interfaces.CostEstimateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		estimate, err := provider.GetCostEstimate(context.Background(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, estimate)
	}
}

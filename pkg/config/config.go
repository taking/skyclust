package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database"`

	// Security Configuration
	Security SecurityConfig `json:"security"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt"`

	// Encryption Configuration
	Encryption EncryptionConfig `json:"encryption"`

	// Logging Configuration
	Logging LoggingConfig `json:"logging"`

	// Cache Configuration
	Cache CacheConfig `json:"cache"`

	// Monitoring Configuration
	Monitoring MonitoringConfig `json:"monitoring"`

	// NATS Configuration
	NATS NATSConfig `json:"nats"`

	// Redis Configuration
	Redis RedisConfig `json:"redis"`

	// Plugin Configuration
	Plugins PluginsConfig `json:"plugins"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
	MaxConns int    `json:"max_conns"`
	MinConns int    `json:"min_conns"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret     string        `json:"jwt_secret"`
	JWTExpiration time.Duration `json:"jwt_expiration"`
	BCryptCost    int           `json:"bcrypt_cost"`
	EncryptionKey string        `json:"encryption_key"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `json:"secret"`
	Expiration time.Duration `json:"expiration"`
	Issuer     string        `json:"issuer"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Key string `json:"key"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Type     string        `json:"type"`
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Password string        `json:"password"`
	DB       int           `json:"db"`
	TTL      time.Duration `json:"ttl"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled     bool   `json:"enabled"`
	MetricsPort int    `json:"metrics_port"`
	HealthPort  int    `json:"health_port"`
	TraceURL    string `json:"trace_url"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `json:"url"`
	Cluster  string `json:"cluster"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// PluginsConfig holds plugins configuration
type PluginsConfig struct {
	Directory string `json:"directory"`
}

// LoadConfigFromFile loads configuration from YAML file
func LoadConfigFromFile(configFile string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// Override with environment variables if they exist
	overrideWithEnvVars(&config)

	return &config, nil
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "cmp"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 25),
			MinConns: getEnvInt("DB_MIN_CONNS", 5),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			BCryptCost:    getEnvInt("BCRYPT_COST", 12),
			EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-here"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			Issuer:     getEnv("JWT_ISSUER", "skyclust"),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-here"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
		},
		Cache: CacheConfig{
			Type:     getEnv("CACHE_TYPE", "memory"),
			Host:     getEnv("CACHE_HOST", "localhost"),
			Port:     getEnvInt("CACHE_PORT", 6379),
			Password: getEnv("CACHE_PASSWORD", ""),
			DB:       getEnvInt("CACHE_DB", 0),
			TTL:      getEnvDuration("CACHE_TTL", 5*time.Minute),
		},
		Monitoring: MonitoringConfig{
			Enabled:     getEnvBool("MONITORING_ENABLED", true),
			MetricsPort: getEnvInt("METRICS_PORT", 9090),
			HealthPort:  getEnvInt("HEALTH_PORT", 8081),
			TraceURL:    getEnv("TRACE_URL", ""),
		},
		NATS: NATSConfig{
			URL:      getEnv("NATS_URL", "nats://localhost:4222"),
			Cluster:  getEnv("NATS_CLUSTER", "test-cluster"),
			Username: getEnv("NATS_USERNAME", ""),
			Password: getEnv("NATS_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		Plugins: PluginsConfig{
			Directory: getEnv("PLUGIN_DIRECTORY", "./plugins"),
		},
	}
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			Name:     "cmp",
			SSLMode:  "disable",
			MaxConns: 25,
			MinConns: 5,
		},
		Security: SecurityConfig{
			JWTSecret:     "your-secret-key",
			JWTExpiration: 24 * time.Hour,
			BCryptCost:    12,
			EncryptionKey: "your-32-byte-encryption-key-here",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
		Cache: CacheConfig{
			Type:     "memory",
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			TTL:      5 * time.Minute,
		},
		Monitoring: MonitoringConfig{
			Enabled:     true,
			MetricsPort: 9090,
			HealthPort:  8081,
			TraceURL:    "",
		},
	}
}

// overrideWithEnvVars overrides config values with environment variables if they exist
func overrideWithEnvVars(config *Config) {
	// Server config
	if host := getEnv("SERVER_HOST", ""); host != "" {
		config.Server.Host = host
	}
	if port := getEnvInt("SERVER_PORT", 0); port != 0 {
		config.Server.Port = port
	}

	// Database config
	if host := getEnv("DB_HOST", ""); host != "" {
		config.Database.Host = host
	}
	if port := getEnvInt("DB_PORT", 0); port != 0 {
		config.Database.Port = port
	}
	if user := getEnv("DB_USER", ""); user != "" {
		config.Database.User = user
	}
	if password := getEnv("DB_PASSWORD", ""); password != "" {
		config.Database.Password = password
	}
	if name := getEnv("DB_NAME", ""); name != "" {
		config.Database.Name = name
	}
	if sslMode := getEnv("DB_SSL_MODE", ""); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Security config
	if jwtSecret := getEnv("JWT_SECRET", ""); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}
	if encryptionKey := getEnv("ENCRYPTION_KEY", ""); encryptionKey != "" {
		config.Security.EncryptionKey = encryptionKey
	}

	// NATS config
	if url := getEnv("NATS_URL", ""); url != "" {
		config.NATS.URL = url
	}
	if cluster := getEnv("NATS_CLUSTER", ""); cluster != "" {
		config.NATS.Cluster = cluster
	}

	// Redis config
	if host := getEnv("REDIS_HOST", ""); host != "" {
		config.Redis.Host = host
	}
	if port := getEnvInt("REDIS_PORT", 0); port != 0 {
		config.Redis.Port = port
	}
	if password := getEnv("REDIS_PASSWORD", ""); password != "" {
		config.Redis.Password = password
	}

	// Plugins config
	if directory := getEnv("PLUGIN_DIRECTORY", ""); directory != "" {
		config.Plugins.Directory = directory
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	// Validate required fields
	if c.Security.JWTSecret == "your-secret-key" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}

	if c.Security.EncryptionKey == "your-32-byte-encryption-key-here" {
		return fmt.Errorf("ENCRYPTION_KEY must be set to a secure value")
	}

	if len(c.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if len(c.Security.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters long")
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetMetricsAddress returns the metrics server address
func (c *Config) GetMetricsAddress() string {
	return fmt.Sprintf(":%d", c.Monitoring.MetricsPort)
}

// GetHealthAddress returns the health check server address
func (c *Config) GetHealthAddress() string {
	return fmt.Sprintf(":%d", c.Monitoring.HealthPort)
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(getEnv("ENVIRONMENT", "development")) == "production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(getEnv("ENVIRONMENT", "development")) == "development"
}

// IsTesting returns true if running in testing mode
func (c *Config) IsTesting() bool {
	return strings.ToLower(getEnv("ENVIRONMENT", "development")) == "testing"
}

package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Security Configuration
	Security SecurityConfig `json:"security" yaml:"security"`

	// Encryption Configuration
	Encryption EncryptionConfig `json:"encryption" yaml:"encryption"`

	// Logging Configuration
	Logging LoggingConfig `json:"logging" yaml:"logging"`

	// Cache Configuration
	Cache CacheConfig `json:"cache" yaml:"cache"`

	// Monitoring Configuration
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`

	// NATS Configuration
	NATS NATSConfig `json:"nats" yaml:"nats"`

	// Redis Configuration
	Redis RedisConfig `json:"redis" yaml:"redis"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`
	MaxConns int    `json:"max_conns" yaml:"max_conns"`
	MinConns int    `json:"min_conns" yaml:"min_conns"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret          string        `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration      time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`           // Access Token expiration
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry" yaml:"refresh_token_expiry"` // Refresh Token expiration
	JWTIssuer          string        `json:"jwt_issuer" yaml:"jwt_issuer"`
	BCryptCost         int           `json:"bcrypt_cost" yaml:"bcrypt_cost"`
	EncryptionKey      string        `json:"encryption_key" yaml:"encryption_key"`
	EnableTokenRotation bool         `json:"enable_token_rotation" yaml:"enable_token_rotation"` // Enable refresh token rotation
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
	URL                  string `json:"url"`
	Cluster              string `json:"cluster"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	CompressionType      string `json:"compression_type"`      // "none", "gzip", "snappy"
	CompressionThreshold int    `json:"compression_threshold"` // Minimum size in bytes to compress
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// EnvMapping defines environment variable mapping
type EnvMapping struct {
	EnvKey    string
	FieldPath string
	Type      string
	Required  bool
}

// EnvCache provides caching for environment variables
type EnvCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

// ConfigOverrider handles environment variable overrides
type ConfigOverrider struct {
	config      *Config
	envCache    *EnvCache
	envMappings []EnvMapping
}

// Environment variable mappings
var envMappings = []EnvMapping{
	// Server configuration
	{"SERVER_HOST", "Server.Host", "string", false},
	{"SERVER_PORT", "Server.Port", "int", false},
	{"SERVER_READ_TIMEOUT", "Server.ReadTimeout", "duration", false},
	{"SERVER_WRITE_TIMEOUT", "Server.WriteTimeout", "duration", false},
	{"SERVER_IDLE_TIMEOUT", "Server.IdleTimeout", "duration", false},

	// Database configuration
	{"DB_HOST", "Database.Host", "string", false},
	{"DB_PORT", "Database.Port", "int", false},
	{"DB_USER", "Database.User", "string", false},
	{"DB_PASSWORD", "Database.Password", "string", false},
	{"DB_NAME", "Database.Name", "string", false},
	{"DB_SSL_MODE", "Database.SSLMode", "string", false},
	{"DB_MAX_CONNS", "Database.MaxConns", "int", false},
	{"DB_MIN_CONNS", "Database.MinConns", "int", false},

	// Security configuration
	{"JWT_SECRET", "Security.JWTSecret", "string", false},
	{"ENCRYPTION_KEY", "Security.EncryptionKey", "string", false},
	{"JWT_ISSUER", "Security.JWTIssuer", "string", false},
	{"BCRYPT_COST", "Security.BCryptCost", "int", false},

	// Redis configuration
	{"REDIS_HOST", "Redis.Host", "string", false},
	{"REDIS_PORT", "Redis.Port", "int", false},
	{"REDIS_PASSWORD", "Redis.Password", "string", false},
	{"REDIS_DB", "Redis.DB", "int", false},
	{"REDIS_POOL_SIZE", "Redis.PoolSize", "int", false},

	// NATS configuration
	{"NATS_URL", "NATS.URL", "string", false},
	{"NATS_CLUSTER", "NATS.Cluster", "string", false},
	{"NATS_USERNAME", "NATS.Username", "string", false},
	{"NATS_PASSWORD", "NATS.Password", "string", false},

	// Logging configuration
	{"LOG_LEVEL", "Logging.Level", "string", false},
	{"LOG_FORMAT", "Logging.Format", "string", false},
	{"LOG_OUTPUT", "Logging.Output", "string", false},

	// Cache configuration
	{"CACHE_TYPE", "Cache.Type", "string", false},
	{"CACHE_HOST", "Cache.Host", "string", false},
	{"CACHE_PORT", "Cache.Port", "int", false},
	{"CACHE_PASSWORD", "Cache.Password", "string", false},

	// Monitoring configuration
	{"MONITORING_ENABLED", "Monitoring.Enabled", "bool", false},
	{"METRICS_PORT", "Monitoring.MetricsPort", "int", false},
	{"HEALTH_PORT", "Monitoring.HealthPort", "int", false},
	{"TRACE_URL", "Monitoring.TraceURL", "string", false},
}

// NewEnvCache creates a new environment variable cache
func NewEnvCache() *EnvCache {
	return &EnvCache{
		cache: make(map[string]string),
	}
}

// Get retrieves an environment variable with caching
func (e *EnvCache) Get(key string) string {
	e.mu.RLock()
	if val, ok := e.cache[key]; ok {
		e.mu.RUnlock()
		return val
	}
	e.mu.RUnlock()

	val := os.Getenv(key)
	e.mu.Lock()
	e.cache[key] = val
	e.mu.Unlock()
	return val
}

// NewConfigOverrider creates a new config overrider
func NewConfigOverrider(config *Config) *ConfigOverrider {
	return &ConfigOverrider{
		config:      config,
		envCache:    NewEnvCache(),
		envMappings: envMappings,
	}
}

// Override applies environment variable overrides to the config
func (c *ConfigOverrider) Override() error {
	for _, mapping := range c.envMappings {
		if val := c.envCache.Get(mapping.EnvKey); val != "" {
			if err := c.setFieldByPath(mapping.FieldPath, val, mapping.Type); err != nil {
				return fmt.Errorf("failed to set %s: %w", mapping.FieldPath, err)
			}
		}
	}
	return nil
}

// setFieldByPath sets a field value by its path
func (c *ConfigOverrider) setFieldByPath(fieldPath, value, fieldType string) error {
	switch fieldPath {
	// Server configuration
	case "Server.Host":
		c.config.Server.Host = value
	case "Server.Port":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid port value '%s': must be positive", value)
		} else {
			c.config.Server.Port = intVal
		}
	case "Server.ReadTimeout":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid read timeout value '%s': %w", value, err)
		} else {
			c.config.Server.ReadTimeout = duration
		}
	case "Server.WriteTimeout":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid write timeout value '%s': %w", value, err)
		} else {
			c.config.Server.WriteTimeout = duration
		}
	case "Server.IdleTimeout":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid idle timeout value '%s': %w", value, err)
		} else {
			c.config.Server.IdleTimeout = duration
		}

	// Database configuration
	case "Database.Host":
		c.config.Database.Host = value
	case "Database.Port":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid database port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid database port value '%s': must be positive", value)
		} else {
			c.config.Database.Port = intVal
		}
	case "Database.User":
		c.config.Database.User = value
	case "Database.Password":
		c.config.Database.Password = value
	case "Database.Name":
		c.config.Database.Name = value
	case "Database.SSLMode":
		c.config.Database.SSLMode = value
	case "Database.MaxConns":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid max connections value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid max connections value '%s': must be positive", value)
		} else {
			c.config.Database.MaxConns = intVal
		}
	case "Database.MinConns":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid min connections value '%s': %w", value, err)
		} else if intVal < 0 {
			return fmt.Errorf("invalid min connections value '%s': must be non-negative", value)
		} else {
			c.config.Database.MinConns = intVal
		}

	// Security configuration
	case "Security.JWTSecret":
		c.config.Security.JWTSecret = value
	case "Security.EncryptionKey":
		c.config.Security.EncryptionKey = value
	case "Security.JWTIssuer":
		c.config.Security.JWTIssuer = value
	case "Security.BCryptCost":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid bcrypt cost value '%s': %w", value, err)
		} else if intVal < 4 || intVal > 31 {
			return fmt.Errorf("invalid bcrypt cost value '%s': must be between 4 and 31", value)
		} else {
			c.config.Security.BCryptCost = intVal
		}

	// Redis configuration
	case "Redis.Host":
		c.config.Redis.Host = value
	case "Redis.Port":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid redis port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid redis port value '%s': must be positive", value)
		} else {
			c.config.Redis.Port = intVal
		}
	case "Redis.Password":
		c.config.Redis.Password = value
	case "Redis.DB":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid redis db value '%s': %w", value, err)
		} else if intVal < 0 {
			return fmt.Errorf("invalid redis db value '%s': must be non-negative", value)
		} else {
			c.config.Redis.DB = intVal
		}
	case "Redis.PoolSize":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid redis pool size value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid redis pool size value '%s': must be positive", value)
		} else {
			c.config.Redis.PoolSize = intVal
		}

	// NATS configuration
	case "NATS.URL":
		c.config.NATS.URL = value
	case "NATS.Cluster":
		c.config.NATS.Cluster = value
	case "NATS.Username":
		c.config.NATS.Username = value
	case "NATS.Password":
		c.config.NATS.Password = value

	// Logging configuration
	case "Logging.Level":
		c.config.Logging.Level = value
	case "Logging.Format":
		c.config.Logging.Format = value
	case "Logging.Output":
		c.config.Logging.Output = value

	// Cache configuration
	case "Cache.Type":
		c.config.Cache.Type = value
	case "Cache.Host":
		c.config.Cache.Host = value
	case "Cache.Port":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid cache port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid cache port value '%s': must be positive", value)
		} else {
			c.config.Cache.Port = intVal
		}
	case "Cache.Password":
		c.config.Cache.Password = value

	// Monitoring configuration
	case "Monitoring.Enabled":
		c.config.Monitoring.Enabled = parseBoolEnv(value, c.config.Monitoring.Enabled)
	case "Monitoring.MetricsPort":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid metrics port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid metrics port value '%s': must be positive", value)
		} else {
			c.config.Monitoring.MetricsPort = intVal
		}
	case "Monitoring.HealthPort":
		if intVal, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid health port value '%s': %w", value, err)
		} else if intVal <= 0 {
			return fmt.Errorf("invalid health port value '%s': must be positive", value)
		} else {
			c.config.Monitoring.HealthPort = intVal
		}
	case "Monitoring.TraceURL":
		c.config.Monitoring.TraceURL = value

	default:
		return fmt.Errorf("unknown field path: %s", fieldPath)
	}

	return nil
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	// Read YAML file directly first
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML directly
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables using the new ConfigOverrider
	overrider := NewConfigOverrider(&config)
	if err := overrider.Override(); err != nil {
		return nil, fmt.Errorf("failed to override config with environment variables: %w", err)
	}

	// Validate and generate secrets if needed
	if err := validateAndGenerateSecrets(&config); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return &config, nil
}

// parseBoolEnv parses an environment variable as bool
func parseBoolEnv(val string, defaultVal bool) bool {
	switch strings.ToLower(val) {
	case "true", "yes", "1", "on":
		return true
	case "false", "no", "0", "off":
		return false
	default:
		return defaultVal
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
			JWTSecret:          "your-secret-key",
			JWTExpiration:      24 * time.Hour,
			RefreshTokenExpiry:  7 * 24 * time.Hour, // 7 days
			JWTIssuer:          "skyclust",
			BCryptCost:         12,
			EncryptionKey:      "your-32-byte-encryption-key-here",
			EnableTokenRotation: true,
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

// validateAndGenerateSecrets validates configuration and generates secrets if needed
func validateAndGenerateSecrets(config *Config) error {
	// Validate JWT secret
	if config.Security.JWTSecret == "" || config.Security.JWTSecret == "your-super-secret-jwt-key-change-in-production" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		config.Security.JWTSecret = secret
		fmt.Println("WARNING: Generated new JWT secret. Please set JWT_SECRET environment variable for production.")
	}

	// Validate encryption key
	fmt.Printf("[CONFIG] Loaded encryption_key length: %d, first 10 chars: %s...\n",
		len(config.Security.EncryptionKey),
		config.Security.EncryptionKey[:min(10, len(config.Security.EncryptionKey))])

	if config.Security.EncryptionKey == "" || config.Security.EncryptionKey == "your-32-byte-encryption-key-here" {
		key, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		config.Security.EncryptionKey = key
		config.Encryption.Key = key
		fmt.Println("WARNING: Generated new encryption key. Please set ENCRYPTION_KEY environment variable for production.")
	} else {
		fmt.Println("[CONFIG] Using encryption_key from config file")
	}

	// Additional security checks for production
	if isProduction() {
		if config.Database.SSLMode == "disable" {
			return fmt.Errorf("SSL must be enabled in production environment")
		}
		if config.Security.JWTSecret == "your-super-secret-jwt-key-change-in-production" {
			return fmt.Errorf("default JWT secret not allowed in production")
		}
		if config.Security.EncryptionKey == "your-32-byte-encryption-key-here" {
			return fmt.Errorf("default encryption key not allowed in production")
		}
	}

	return nil
}

// generateRandomSecret generates a random secret of specified length
func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// isProduction checks if the application is running in production mode
func isProduction() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "production" || env == "prod"
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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

	if len(c.Security.EncryptionKey) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be at least 32 characters long")
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

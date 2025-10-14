package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
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
	JWTIssuer     string        `json:"jwt_issuer"`
	BCryptCost    int           `json:"bcrypt_cost"`
	EncryptionKey string        `json:"encryption_key"`
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
	// Note: LoadConfigFromFile uses direct YAML parsing, so we don't use viper here

	return &config, nil
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	// Set up viper
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.skyclust")
	viper.AddConfigPath("/etc/skyclust")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	overrideWithEnvVars()

	// Unmarshal into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Handle empty values by falling back to environment variables
	handleEmptyValues(&config)

	// Validate and generate secrets if needed
	if err := validateAndGenerateSecrets(&config); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return &config, nil
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
			JWTIssuer:     "skyclust",
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

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "skyclust")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)

	// Security defaults
	viper.SetDefault("security.jwt_secret", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("security.jwt_expiration", 24)
	viper.SetDefault("security.jwt_issuer", "skyclust")
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("security.encryption_key", "your-32-byte-encryption-key-here")

	// Encryption defaults
	viper.SetDefault("encryption.key", "your-32-byte-encryption-key-here")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)

	// Cache defaults
	viper.SetDefault("cache.type", "memory")
	viper.SetDefault("cache.host", "localhost")
	viper.SetDefault("cache.port", 6379)
	viper.SetDefault("cache.password", "")
	viper.SetDefault("cache.db", 0)
	viper.SetDefault("cache.ttl", 300)

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
	viper.SetDefault("monitoring.health_port", 8081)
	viper.SetDefault("monitoring.trace_url", "")

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.cluster", "skyclust-cluster")
	viper.SetDefault("nats.username", "")
	viper.SetDefault("nats.password", "")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// Plugins defaults
	viper.SetDefault("plugins.directory", "./plugins")
}

// overrideWithEnvVars overrides configuration with environment variables
func overrideWithEnvVars() {
	// Server environment variables
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = viper.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	// Database environment variables
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.name", "DB_NAME")
	_ = viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	_ = viper.BindEnv("database.max_conns", "DB_MAX_CONNS")
	_ = viper.BindEnv("database.min_conns", "DB_MIN_CONNS")

	// Security environment variables
	_ = viper.BindEnv("security.jwt_secret", "JWT_SECRET")
	_ = viper.BindEnv("security.jwt_expiration", "JWT_EXPIRATION")
	_ = viper.BindEnv("security.jwt_issuer", "JWT_ISSUER")
	_ = viper.BindEnv("security.bcrypt_cost", "BCRYPT_COST")
	_ = viper.BindEnv("security.encryption_key", "ENCRYPTION_KEY")

	// Encryption environment variables
	_ = viper.BindEnv("encryption.key", "ENCRYPTION_KEY")

	// Logging environment variables
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")
	_ = viper.BindEnv("logging.output", "LOG_OUTPUT")
	_ = viper.BindEnv("logging.max_size", "LOG_MAX_SIZE")
	_ = viper.BindEnv("logging.max_backups", "LOG_MAX_BACKUPS")
	_ = viper.BindEnv("logging.max_age", "LOG_MAX_AGE")

	// Cache environment variables
	_ = viper.BindEnv("cache.type", "CACHE_TYPE")
	_ = viper.BindEnv("cache.host", "CACHE_HOST")
	_ = viper.BindEnv("cache.port", "CACHE_PORT")
	_ = viper.BindEnv("cache.password", "CACHE_PASSWORD")
	_ = viper.BindEnv("cache.db", "CACHE_DB")
	_ = viper.BindEnv("cache.ttl", "CACHE_TTL")

	// Monitoring environment variables
	_ = viper.BindEnv("monitoring.enabled", "MONITORING_ENABLED")
	_ = viper.BindEnv("monitoring.metrics_port", "METRICS_PORT")
	_ = viper.BindEnv("monitoring.health_port", "HEALTH_PORT")
	_ = viper.BindEnv("monitoring.trace_url", "TRACE_URL")

	// NATS environment variables
	_ = viper.BindEnv("nats.url", "NATS_URL")
	_ = viper.BindEnv("nats.cluster", "NATS_CLUSTER")
	_ = viper.BindEnv("nats.username", "NATS_USERNAME")
	_ = viper.BindEnv("nats.password", "NATS_PASSWORD")

	// Redis environment variables
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")
	_ = viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")

	// Plugins environment variables
	_ = viper.BindEnv("plugins.directory", "PLUGIN_DIRECTORY")
}

// handleEmptyValues handles empty values by falling back to environment variables
func handleEmptyValues(config *Config) {
	// This function can be used to handle any additional fallback logic
	// For now, viper handles most of the environment variable binding
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
	if config.Security.EncryptionKey == "" || config.Security.EncryptionKey == "your-32-byte-encryption-key-here" {
		key, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		config.Security.EncryptionKey = key
		config.Encryption.Key = key
		fmt.Println("WARNING: Generated new encryption key. Please set ENCRYPTION_KEY environment variable for production.")
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

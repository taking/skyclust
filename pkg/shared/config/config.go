package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cmp/pkg/shared/security"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	NATS       NATSConfig       `mapstructure:"nats"`
	Plugins    PluginsConfig    `mapstructure:"plugins"`
	Providers  ProvidersConfig  `mapstructure:"providers"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int    `mapstructure:"max_connections"`
	MinConns int    `mapstructure:"min_connections"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration_hours"`
	Issuer     string `mapstructure:"issuer"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Key string `mapstructure:"key"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL     string `mapstructure:"url"`
	Cluster string `mapstructure:"cluster"`
}

// PluginsConfig holds plugins configuration
type PluginsConfig struct {
	Directory string `mapstructure:"directory"`
}

// ProvidersConfig holds cloud providers configuration
type ProvidersConfig struct {
	AWS       AWSConfig       `mapstructure:"aws"`
	GCP       GCPConfig       `mapstructure:"gcp"`
	OpenStack OpenStackConfig `mapstructure:"openstack"`
	Proxmox   ProxmoxConfig   `mapstructure:"proxmox"`
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	RoleARN   string `mapstructure:"role_arn"`
}

// GCPConfig holds GCP configuration
type GCPConfig struct {
	ProjectID       string `mapstructure:"project_id"`
	CredentialsFile string `mapstructure:"credentials_file"`
	Region          string `mapstructure:"region"`
}

// OpenStackConfig holds OpenStack configuration
type OpenStackConfig struct {
	AuthURL   string `mapstructure:"auth_url"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	ProjectID string `mapstructure:"project_id"`
	Region    string `mapstructure:"region"`
}

// ProxmoxConfig holds Proxmox configuration
type ProxmoxConfig struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Realm    string `mapstructure:"realm"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	// Set up viper
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.cmp")
	viper.AddConfigPath("/etc/cmp")

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

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "cmp_user")
	viper.SetDefault("database.password", "cmp_password")
	viper.SetDefault("database.name", "cmp")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.min_connections", 5)

	// JWT defaults
	viper.SetDefault("jwt.expiration_hours", 24)
	viper.SetDefault("jwt.issuer", "cmp")

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.cluster", "cmp-cluster")

	// Plugins defaults
	viper.SetDefault("plugins.directory", "plugins")

	// Provider defaults
	viper.SetDefault("providers.aws.region", "us-east-1")
	viper.SetDefault("providers.gcp.region", "us-central1")
	viper.SetDefault("providers.openstack.region", "RegionOne")
	viper.SetDefault("providers.proxmox.realm", "pve")
}

// overrideWithEnvVars overrides configuration with environment variables
func overrideWithEnvVars() {
	// Server environment variables
	_ = viper.BindEnv("server.port", "CMP_PORT")
	_ = viper.BindEnv("server.host", "CMP_HOST")
	_ = viper.BindEnv("server.read_timeout", "CMP_READ_TIMEOUT")
	_ = viper.BindEnv("server.write_timeout", "CMP_WRITE_TIMEOUT")
	_ = viper.BindEnv("server.idle_timeout", "CMP_IDLE_TIMEOUT")

	// Database environment variables
	_ = viper.BindEnv("database.host", "CMP_DB_HOST")
	_ = viper.BindEnv("database.port", "CMP_DB_PORT")
	_ = viper.BindEnv("database.user", "CMP_DB_USER")
	_ = viper.BindEnv("database.password", "CMP_DB_PASSWORD")
	_ = viper.BindEnv("database.name", "CMP_DB_NAME")
	_ = viper.BindEnv("database.sslmode", "CMP_DB_SSLMODE")
	_ = viper.BindEnv("database.max_connections", "CMP_DB_MAX_CONNS")
	_ = viper.BindEnv("database.min_connections", "CMP_DB_MIN_CONNS")

	// JWT environment variables
	_ = viper.BindEnv("jwt.secret", "CMP_JWT_SECRET")
	_ = viper.BindEnv("jwt.expiration_hours", "CMP_JWT_EXPIRATION")
	_ = viper.BindEnv("jwt.issuer", "CMP_JWT_ISSUER")

	// Encryption environment variables
	_ = viper.BindEnv("encryption.key", "CMP_ENCRYPTION_KEY")

	// NATS environment variables
	_ = viper.BindEnv("nats.url", "CMP_NATS_URL")
	_ = viper.BindEnv("nats.cluster", "CMP_NATS_CLUSTER")

	// Plugins environment variables
	_ = viper.BindEnv("plugins.directory", "CMP_PLUGINS_DIR")

	// Provider environment variables
	_ = viper.BindEnv("providers.aws.access_key", "CMP_AWS_ACCESS_KEY")
	_ = viper.BindEnv("providers.aws.secret_key", "CMP_AWS_SECRET_KEY")
	_ = viper.BindEnv("providers.aws.region", "CMP_AWS_REGION")
	_ = viper.BindEnv("providers.aws.role_arn", "CMP_AWS_ROLE_ARN")

	_ = viper.BindEnv("providers.gcp.project_id", "CMP_GCP_PROJECT_ID")
	_ = viper.BindEnv("providers.gcp.credentials_file", "CMP_GCP_CREDENTIALS_FILE")
	_ = viper.BindEnv("providers.gcp.region", "CMP_GCP_REGION")

	_ = viper.BindEnv("providers.openstack.auth_url", "CMP_OPENSTACK_AUTH_URL")
	_ = viper.BindEnv("providers.openstack.username", "CMP_OPENSTACK_USERNAME")
	_ = viper.BindEnv("providers.openstack.password", "CMP_OPENSTACK_PASSWORD")
	_ = viper.BindEnv("providers.openstack.project_id", "CMP_OPENSTACK_PROJECT_ID")
	_ = viper.BindEnv("providers.openstack.region", "CMP_OPENSTACK_REGION")

	_ = viper.BindEnv("providers.proxmox.host", "CMP_PROXMOX_HOST")
	_ = viper.BindEnv("providers.proxmox.username", "CMP_PROXMOX_USERNAME")
	_ = viper.BindEnv("providers.proxmox.password", "CMP_PROXMOX_PASSWORD")
	_ = viper.BindEnv("providers.proxmox.realm", "CMP_PROXMOX_REALM")
}

// validateAndGenerateSecrets validates configuration and generates secrets if needed
func validateAndGenerateSecrets(config *Config) error {
	// Validate environment
	if err := security.ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	// Validate database configuration
	if err := security.ValidateDatabaseConfig(
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.SSLMode,
	); err != nil {
		return fmt.Errorf("database configuration validation failed: %w", err)
	}

	// Validate JWT secret
	if config.JWT.Secret == "" || config.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		config.JWT.Secret = secret
		fmt.Println("WARNING: Generated new JWT secret. Please set CMP_JWT_SECRET environment variable for production.")
	} else {
		// Validate existing JWT secret strength
		if err := security.ValidateJWTSecret(config.JWT.Secret); err != nil {
			return fmt.Errorf("JWT secret validation failed: %w", err)
		}
	}

	// Validate encryption key
	if config.Encryption.Key == "" || config.Encryption.Key == "your-32-byte-encryption-key-here" {
		key, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		config.Encryption.Key = key
		fmt.Println("WARNING: Generated new encryption key. Please set CMP_ENCRYPTION_KEY environment variable for production.")
	} else {
		// Validate existing encryption key strength
		if err := security.ValidateEncryptionKey(config.Encryption.Key); err != nil {
			return fmt.Errorf("encryption key validation failed: %w", err)
		}
	}

	// Additional security checks for production
	if security.IsProductionEnvironment() {
		if config.Database.SSLMode == "disable" {
			return fmt.Errorf("SSL must be enabled in production environment")
		}

		if config.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
			return fmt.Errorf("default JWT secret not allowed in production")
		}

		if config.Encryption.Key == "your-32-byte-encryption-key-here" {
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

// GetEnvOrDefault gets environment variable or returns default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvIntOrDefault gets environment variable as int or returns default value
func GetEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// IsProduction checks if the application is running in production mode
func IsProduction() bool {
	env := strings.ToLower(os.Getenv("CMP_ENV"))
	return env == "production" || env == "prod"
}

// IsDevelopment checks if the application is running in development mode
func IsDevelopment() bool {
	env := strings.ToLower(os.Getenv("CMP_ENV"))
	return env == "development" || env == "dev" || env == ""
}

// IsTesting checks if the application is running in testing mode
func IsTesting() bool {
	env := strings.ToLower(os.Getenv("CMP_ENV"))
	return env == "testing" || env == "test"
}

// handleEmptyValues handles empty values by falling back to environment variables
func handleEmptyValues(config *Config) {
	// AWS credentials fallback
	if config.Providers.AWS.AccessKey == "" {
		if envKey := os.Getenv("CMP_AWS_ACCESS_KEY"); envKey != "" {
			config.Providers.AWS.AccessKey = envKey
		}
	}
	if config.Providers.AWS.SecretKey == "" {
		if envSecret := os.Getenv("CMP_AWS_SECRET_KEY"); envSecret != "" {
			config.Providers.AWS.SecretKey = envSecret
		}
	}
	if config.Providers.AWS.Region == "" {
		if envRegion := os.Getenv("CMP_AWS_REGION"); envRegion != "" {
			config.Providers.AWS.Region = envRegion
		}
	}

	// GCP credentials fallback
	if config.Providers.GCP.ProjectID == "" {
		if envProject := os.Getenv("CMP_GCP_PROJECT_ID"); envProject != "" {
			config.Providers.GCP.ProjectID = envProject
		}
	}
	if config.Providers.GCP.CredentialsFile == "" {
		if envCreds := os.Getenv("CMP_GCP_CREDENTIALS_FILE"); envCreds != "" {
			config.Providers.GCP.CredentialsFile = envCreds
		}
	}
	if config.Providers.GCP.Region == "" {
		if envRegion := os.Getenv("CMP_GCP_REGION"); envRegion != "" {
			config.Providers.GCP.Region = envRegion
		}
	}

	// OpenStack credentials fallback
	if config.Providers.OpenStack.AuthURL == "" {
		if envAuthURL := os.Getenv("CMP_OPENSTACK_AUTH_URL"); envAuthURL != "" {
			config.Providers.OpenStack.AuthURL = envAuthURL
		}
	}
	if config.Providers.OpenStack.Username == "" {
		if envUsername := os.Getenv("CMP_OPENSTACK_USERNAME"); envUsername != "" {
			config.Providers.OpenStack.Username = envUsername
		}
	}
	if config.Providers.OpenStack.Password == "" {
		if envPassword := os.Getenv("CMP_OPENSTACK_PASSWORD"); envPassword != "" {
			config.Providers.OpenStack.Password = envPassword
		}
	}
	if config.Providers.OpenStack.ProjectID == "" {
		if envProject := os.Getenv("CMP_OPENSTACK_PROJECT_ID"); envProject != "" {
			config.Providers.OpenStack.ProjectID = envProject
		}
	}
	if config.Providers.OpenStack.Region == "" {
		if envRegion := os.Getenv("CMP_OPENSTACK_REGION"); envRegion != "" {
			config.Providers.OpenStack.Region = envRegion
		}
	}

	// Proxmox credentials fallback
	if config.Providers.Proxmox.Host == "" {
		if envHost := os.Getenv("CMP_PROXMOX_HOST"); envHost != "" {
			config.Providers.Proxmox.Host = envHost
		}
	}
	if config.Providers.Proxmox.Username == "" {
		if envUsername := os.Getenv("CMP_PROXMOX_USERNAME"); envUsername != "" {
			config.Providers.Proxmox.Username = envUsername
		}
	}
	if config.Providers.Proxmox.Password == "" {
		if envPassword := os.Getenv("CMP_PROXMOX_PASSWORD"); envPassword != "" {
			config.Providers.Proxmox.Password = envPassword
		}
	}
	if config.Providers.Proxmox.Realm == "" {
		if envRealm := os.Getenv("CMP_PROXMOX_REALM"); envRealm != "" {
			config.Providers.Proxmox.Realm = envRealm
		}
	}
}

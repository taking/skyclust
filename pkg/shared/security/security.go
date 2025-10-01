package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MinPasswordLength  int
	RequireSpecialChar bool
	RequireNumber      bool
	RequireUppercase   bool
	RequireLowercase   bool
	MaxLoginAttempts   int
	LockoutDuration    int // minutes
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MinPasswordLength:  8,
		RequireSpecialChar: true,
		RequireNumber:      true,
		RequireUppercase:   true,
		RequireLowercase:   true,
		MaxLoginAttempts:   5,
		LockoutDuration:    15,
	}
}

// ValidatePassword validates password strength
func ValidatePassword(password string, config *SecurityConfig) error {
	if len(password) < config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", config.MinPasswordLength)
	}

	if config.RequireUppercase {
		if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	if config.RequireLowercase {
		if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
	}

	if config.RequireNumber {
		if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	if config.RequireSpecialChar {
		if matched, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password); !matched {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword creates a SHA-256 hash of the password
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ValidateEnvironment validates that all required environment variables are set
// This function is now optional and only validates if environment variables are explicitly set
func ValidateEnvironment() error {
	// Only validate if we're in production mode and environment variables are explicitly required
	if IsProductionEnvironment() {
		requiredVars := []string{
			"CMP_DB_HOST",
			"CMP_DB_USER",
			"CMP_DB_PASSWORD",
			"CMP_DB_NAME",
		}

		var missing []string
		for _, varName := range requiredVars {
			if os.Getenv(varName) == "" {
				missing = append(missing, varName)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
		}
	}

	return nil
}

// ValidateSecretStrength validates that secrets meet minimum security requirements
func ValidateSecretStrength(secret, secretName string) error {
	if len(secret) < 32 {
		return fmt.Errorf("%s must be at least 32 characters long", secretName)
	}

	// Check for common weak secrets
	weakSecrets := []string{
		"your-super-secret-jwt-key-change-in-production",
		"your-32-byte-encryption-key-here",
		"password",
		"123456",
		"secret",
		"admin",
	}

	secretLower := strings.ToLower(secret)
	for _, weak := range weakSecrets {
		if strings.Contains(secretLower, weak) {
			return fmt.Errorf("%s contains weak or default value", secretName)
		}
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	re := regexp.MustCompile(`[<>'"&]`)
	return re.ReplaceAllString(input, "")
}

// IsProductionEnvironment checks if the application is running in production
func IsProductionEnvironment() bool {
	env := strings.ToLower(os.Getenv("CMP_ENV"))
	return env == "production" || env == "prod"
}

// GetEnvironmentName returns the current environment name
func GetEnvironmentName() string {
	env := strings.ToLower(os.Getenv("CMP_ENV"))
	if env == "" {
		return "development"
	}
	return env
}

// ValidateDatabaseConfig validates database configuration for security
func ValidateDatabaseConfig(host, user, password, sslMode string) error {
	// Check for localhost in production
	if IsProductionEnvironment() && (host == "localhost" || host == "127.0.0.1") {
		return fmt.Errorf("localhost database connection not allowed in production")
	}

	// Check SSL mode in production
	if IsProductionEnvironment() && sslMode == "disable" {
		return fmt.Errorf("SSL must be enabled in production environment")
	}

	// Check for weak passwords
	if len(password) < 8 {
		return fmt.Errorf("database password must be at least 8 characters long")
	}

	return nil
}

// MaskSensitiveData masks sensitive data in logs
func MaskSensitiveData(data string) string {
	if len(data) <= 8 {
		return "****"
	}
	return data[:4] + "****" + data[len(data)-4:]
}

// ValidateJWTSecret validates JWT secret strength
func ValidateJWTSecret(secret string) error {
	return ValidateSecretStrength(secret, "JWT secret")
}

// ValidateEncryptionKey validates encryption key strength
func ValidateEncryptionKey(key string) error {
	return ValidateSecretStrength(key, "encryption key")
}

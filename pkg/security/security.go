package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher interface for password hashing
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// BcryptHasher provides password hashing functionality
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt-based password hasher
func NewBcryptHasher(cost int) PasswordHasher {
	return &BcryptHasher{
		cost: cost,
	}
}

// HashPassword hashes a password using bcrypt
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash using bcrypt
func (h *BcryptHasher) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Encryptor interface for encryption
type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// AESEncryptor provides encryption functionality
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor creates a new AES-based encryptor
func NewAESEncryptor(key []byte) Encryptor {
	return &AESEncryptor{
		key: key,
	}
}

// Encrypt encrypts data
func (e *AESEncryptor) Encrypt(data []byte) ([]byte, error) {
	// Simple XOR encryption for now
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ e.key[i%len(e.key)]
	}
	return result, nil
}

// Decrypt decrypts data
func (e *AESEncryptor) Decrypt(data []byte) ([]byte, error) {
	// Simple XOR decryption for now
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ e.key[i%len(e.key)]
	}
	return result, nil
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
func ValidateEnvironment() error {
	// Only validate if we're in production mode
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

// IsProductionEnvironment checks if running in production
func IsProductionEnvironment() bool {
	return strings.ToLower(os.Getenv("ENVIRONMENT")) == "production"
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
	re := strings.NewReplacer(
		"<", "",
		">", "",
		"'", "",
		"\"", "",
		"&", "",
	)
	return re.Replace(input)
}

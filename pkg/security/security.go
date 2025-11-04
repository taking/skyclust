package security

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/chacha20poly1305"
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

// ChaChaEncryptor provides encryption functionality using ChaCha20-Poly1305
type ChaChaEncryptor struct {
	aead cipher.AEAD
}

// NewAESEncryptor creates a new ChaCha20-Poly1305 encryptor
// Note: Renamed to maintain backward compatibility, but uses ChaCha20-Poly1305
func NewAESEncryptor(key []byte) Encryptor {
	// Derive a 32-byte key using SHA-256
	hash := sha256.Sum256(key)

	aead, err := chacha20poly1305.NewX(hash[:])
	if err != nil {
		// Fallback to standard ChaCha20-Poly1305 if XChaCha20-Poly1305 fails
		aead, err = chacha20poly1305.New(hash[:])
		if err != nil {
			panic(fmt.Sprintf("failed to create cipher: %v", err))
		}
	}

	return &ChaChaEncryptor{
		aead: aead,
	}
}

// Encrypt encrypts data using ChaCha20-Poly1305
func (e *ChaChaEncryptor) Encrypt(data []byte) ([]byte, error) {
	// Generate nonce
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (nonce is prepended to ciphertext)
	ciphertext := e.aead.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using ChaCha20-Poly1305
func (e *ChaChaEncryptor) Decrypt(data []byte) ([]byte, error) {
	// Check minimum length
	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: got %d bytes, need at least %d", len(data), nonceSize)
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt data
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword is a convenience function for hashing passwords
func HashPassword(password string) (string, error) {
	hasher := NewBcryptHasher(12)
	return hasher.HashPassword(password)
}

// VerifyPassword is a convenience function for verifying passwords
func VerifyPassword(password, hash string) bool {
	hasher := NewBcryptHasher(12)
	return hasher.VerifyPassword(password, hash)
}

// ValidateEnvironment validates required environment variables
func ValidateEnvironment(requiredVars []string) error {
	if requiredVars == nil {
		// Default production environment variables
		requiredVars = []string{
			"JWT_SECRET",
			"ENCRYPTION_KEY",
			"CMP_DB_HOST",
			"CMP_DB_PORT",
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

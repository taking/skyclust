package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// BcryptHasher implements password hashing using bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt hasher
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{
		cost: bcrypt.DefaultCost,
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

// VerifyPassword verifies a password against a hash
func (h *BcryptHasher) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

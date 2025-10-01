package credentials

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"cmp/internal/encryption"
	"cmp/pkg/database"
)

// Credentials represents encrypted credentials for a cloud provider
type Credentials = database.Credentials

// Service defines the credentials service interface
type Service interface {
	CreateCredentials(ctx context.Context, workspaceID, provider string, config map[string]string) (*Credentials, error)
	GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error)
	ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error)
	UpdateCredentials(ctx context.Context, workspaceID, credID string, config map[string]string) (*Credentials, error)
	DeleteCredentials(ctx context.Context, workspaceID, credID string) error
	DecryptCredentials(ctx context.Context, cred *Credentials) (map[string]string, error)
}

type service struct {
	db         database.Service
	encryption encryption.Service
}

// NewService creates a new credentials service
func NewService(db database.Service, encryptionService encryption.Service) Service {
	return &service{
		db:         db,
		encryption: encryptionService,
	}
}

// CreateCredentials creates new encrypted credentials
func (s *service) CreateCredentials(ctx context.Context, workspaceID, provider string, config map[string]string) (*Credentials, error) {
	// Encrypt credentials
	encrypted, err := s.encrypt(config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	cred := &Credentials{
		ID:          generateID(),
		WorkspaceID: workspaceID,
		Provider:    provider,
		Encrypted:   encrypted,
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.CreateCredentials(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	return cred, nil
}

// GetCredentials retrieves credentials by ID
func (s *service) GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error) {
	cred, err := s.db.GetCredentials(ctx, workspaceID, credID)
	if err != nil {
		return nil, fmt.Errorf("credentials not found: %w", err)
	}
	return cred, nil
}

// ListCredentials lists all credentials for a workspace
func (s *service) ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error) {
	creds, err := s.db.ListCredentials(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	return creds, nil
}

// UpdateCredentials updates existing credentials
func (s *service) UpdateCredentials(ctx context.Context, workspaceID, credID string, config map[string]string) (*Credentials, error) {
	// Get existing credentials
	cred, err := s.GetCredentials(ctx, workspaceID, credID)
	if err != nil {
		return nil, err
	}

	// Encrypt new credentials
	encrypted, err := s.encrypt(config)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	// Update credentials
	cred.Encrypted = encrypted
	cred.UpdatedAt = time.Now()

	if err := s.db.UpdateCredentials(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to update credentials: %w", err)
	}

	return cred, nil
}

// DeleteCredentials deletes credentials
func (s *service) DeleteCredentials(ctx context.Context, workspaceID, credID string) error {
	if err := s.db.DeleteCredentials(ctx, workspaceID, credID); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// DecryptCredentials decrypts credentials
func (s *service) DecryptCredentials(ctx context.Context, cred *Credentials) (map[string]string, error) {
	config, err := s.decrypt(cred.Encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}
	return config, nil
}

// encrypt encrypts credentials using the encryption service
func (s *service) encrypt(config map[string]string) ([]byte, error) {
	// Convert map to JSON string
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	// Use encryption service
	encrypted, err := s.encryption.Encrypt(string(jsonData))
	if err != nil {
		return nil, err
	}

	return []byte(encrypted), nil
}

// decrypt decrypts credentials using the encryption service
func (s *service) decrypt(encrypted []byte) (map[string]string, error) {
	// Use encryption service
	decrypted, err := s.encryption.Decrypt(string(encrypted))
	if err != nil {
		return nil, err
	}

	// Parse JSON back to map
	var config map[string]string
	if err := json.Unmarshal([]byte(decrypted), &config); err != nil {
		return nil, err
	}

	return config, nil
}

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

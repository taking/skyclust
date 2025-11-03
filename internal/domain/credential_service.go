package domain

import (
	"context"

	"github.com/google/uuid"
)

// CredentialService defines the interface for credential business logic
type CredentialService interface {
	CreateCredential(ctx context.Context, workspaceID, createdBy uuid.UUID, req CreateCredentialRequest) (*Credential, error)
	GetCredentials(ctx context.Context, workspaceID uuid.UUID) ([]*Credential, error)
	GetCredentialByID(ctx context.Context, workspaceID, credentialID uuid.UUID) (*Credential, error)
	UpdateCredential(ctx context.Context, workspaceID, credentialID uuid.UUID, req UpdateCredentialRequest) (*Credential, error)
	DeleteCredential(ctx context.Context, workspaceID, credentialID uuid.UUID) error
	EncryptCredentialData(ctx context.Context, data map[string]interface{}) ([]byte, error)
	DecryptCredentialData(ctx context.Context, encryptedData []byte) (map[string]interface{}, error)
	// Deprecated: Use CreateCredential with workspaceID instead
	CreateCredentialByUser(ctx context.Context, userID uuid.UUID, req CreateCredentialRequest) (*Credential, error)
	// Deprecated: Use GetCredentials with workspaceID instead
	GetCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]*Credential, error)
	// Deprecated: Use GetCredentialByID with workspaceID instead
	GetCredentialByIDAndUser(ctx context.Context, userID, credentialID uuid.UUID) (*Credential, error)
	// Deprecated: Use UpdateCredential with workspaceID instead
	UpdateCredentialByUser(ctx context.Context, userID, credentialID uuid.UUID, req UpdateCredentialRequest) (*Credential, error)
	// Deprecated: Use DeleteCredential with workspaceID instead
	DeleteCredentialByUser(ctx context.Context, userID, credentialID uuid.UUID) error
}


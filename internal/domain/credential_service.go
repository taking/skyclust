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
	GetCredentialByIDDirect(ctx context.Context, credentialID uuid.UUID) (*Credential, error)
}

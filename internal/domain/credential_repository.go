package domain

import (
	"github.com/google/uuid"
)

// CredentialRepository defines the interface for credential data operations
type CredentialRepository interface {
	Create(credential *Credential) error
	GetByID(id uuid.UUID) (*Credential, error)
	GetByWorkspaceID(workspaceID uuid.UUID) ([]*Credential, error)
	GetByWorkspaceIDAndProvider(workspaceID uuid.UUID, provider string) ([]*Credential, error)
	Update(credential *Credential) error
	Delete(id uuid.UUID) error
	DeleteByWorkspaceID(workspaceID uuid.UUID) error
}

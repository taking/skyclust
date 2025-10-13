package database

import (
	"github.com/google/uuid"
	"skyclust/internal/domain"
	"time"
)

// GormCredentials represents encrypted credentials for a cloud provider with GORM tags
// This is kept separate from domain.Credential as it has different structure (workspace-based vs user-based)
type GormCredentials struct {
	ID          string            `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WorkspaceID string            `gorm:"not null;type:uuid" json:"workspace_id"`
	Provider    string            `gorm:"not null" json:"provider"`
	Encrypted   []byte            `gorm:"type:bytea;not null" json:"-"`
	Metadata    map[string]string `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Credentials
func (GormCredentials) TableName() string {
	return "credentials"
}

// Execution represents an OpenTofu execution with GORM tags
type GormExecution struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WorkspaceID string     `gorm:"not null;type:uuid" json:"workspace_id"`
	Command     string     `gorm:"not null" json:"command"`
	Status      string     `gorm:"not null" json:"status"`
	Output      string     `gorm:"type:text" json:"output"`
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   time.Time  `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// TableName specifies the table name for Execution
func (GormExecution) TableName() string {
	return "executions"
}

// Token represents JWT tokens for session management with GORM tags
type GormToken struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"column:token_hash;not null" json:"token"`
	UserID    string    `gorm:"not null;type:uuid" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for Token
func (GormToken) TableName() string {
	return "tokens"
}

// Common conversion utilities for domain entities
// Since domain entities now have GORM tags, direct usage is preferred
// These utilities are kept for backward compatibility and special cases

// ToCredentials converts GormCredentials to domain.Credential
// Note: This is a special case as GormCredentials has different structure than domain.Credential
func (g *GormCredentials) ToCredentials() *domain.Credential {
	credID, _ := uuid.Parse(g.ID)

	return &domain.Credential{
		ID:            credID,
		UserID:        uuid.Nil, // GormCredentials doesn't have UserID, it's workspace-based
		Provider:      g.Provider,
		Name:          g.Metadata["name"], // Extract name from metadata
		EncryptedData: g.Encrypted,
		IsActive:      true, // Default value
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

// ToGormCredentials converts domain.Credential to GormCredentials
// Note: This is a special case as they have different structures
func ToGormCredentials(c *domain.Credential, workspaceID string) *GormCredentials {
	metadata := make(map[string]string)
	if c.Name != "" {
		metadata["name"] = c.Name
	}

	return &GormCredentials{
		ID:          c.ID.String(),
		WorkspaceID: workspaceID,
		Provider:    c.Provider,
		Encrypted:   c.EncryptedData,
		Metadata:    metadata,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

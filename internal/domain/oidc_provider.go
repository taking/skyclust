package domain

import (
	"time"

	"github.com/google/uuid"
)

// OIDCProvider represents a user-registered OIDC provider configuration
type OIDCProvider struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:100"` // Display name for the provider
	ProviderType string    `json:"provider_type" gorm:"not null;size:50"` // google, github, azure, microsoft, custom
	ClientID     string    `json:"client_id" gorm:"not null;size:255"` // Encrypted
	ClientSecret string    `json:"-" gorm:"not null;size:255"`       // Encrypted, not returned in JSON
	RedirectURL  string    `json:"redirect_url" gorm:"not null;size:500"`
	AuthURL      string    `json:"auth_url" gorm:"size:500"`      // Custom OAuth2 authorization URL
	TokenURL     string    `json:"token_url" gorm:"size:500"`      // Custom OAuth2 token URL
	UserInfoURL  string    `json:"user_info_url" gorm:"size:500"`  // Custom user info endpoint
	Scopes       string    `json:"scopes" gorm:"size:500"`        // Comma-separated scopes
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for OIDCProvider
func (OIDCProvider) TableName() string {
	return "oidc_providers"
}

// IsCustomProvider checks if this is a custom provider (not predefined)
func (p *OIDCProvider) IsCustomProvider() bool {
	return p.ProviderType == "custom"
}

// IsEnabled checks if the provider is enabled
func (p *OIDCProvider) IsEnabled() bool {
	return p.Enabled
}

// Enable enables the provider
func (p *OIDCProvider) Enable() {
	p.Enabled = true
}

// Disable disables the provider
func (p *OIDCProvider) Disable() {
	p.Enabled = false
}

// OIDCProviderRepository defines the interface for OIDC provider persistence
type OIDCProviderRepository interface {
	Create(provider *OIDCProvider) error
	GetByID(id uuid.UUID) (*OIDCProvider, error)
	GetByUserID(userID uuid.UUID) ([]*OIDCProvider, error)
	GetByUserIDAndName(userID uuid.UUID, name string) (*OIDCProvider, error)
	Update(provider *OIDCProvider) error
	Delete(id uuid.UUID) error
	GetEnabledByUserID(userID uuid.UUID) ([]*OIDCProvider, error)
}


package domain

import (
	"time"

	"github.com/google/uuid"
)

// OIDCProvider: 사용자가 등록한 OIDC 제공자 구성을 나타내는 도메인 엔티티
type OIDCProvider struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:100"`         // Display name for the provider
	ProviderType string    `json:"provider_type" gorm:"not null;size:50"` // google, github, azure, microsoft, custom
	ClientID     string    `json:"client_id" gorm:"not null;size:255"`    // Encrypted
	ClientSecret string    `json:"-" gorm:"not null;size:255"`            // Encrypted, not returned in JSON
	RedirectURL  string    `json:"redirect_url" gorm:"not null;size:500"`
	AuthURL      string    `json:"auth_url" gorm:"size:500"`      // Custom OAuth2 authorization URL
	TokenURL     string    `json:"token_url" gorm:"size:500"`     // Custom OAuth2 token URL
	UserInfoURL  string    `json:"user_info_url" gorm:"size:500"` // Custom user info endpoint
	Scopes       string    `json:"scopes" gorm:"size:500"`        // Comma-separated scopes
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName: OIDCProvider의 테이블 이름을 반환합니다
func (OIDCProvider) TableName() string {
	return "oidc_providers"
}

// IsCustomProvider: 커스텀 제공자(미리 정의되지 않은)인지 확인합니다
func (p *OIDCProvider) IsCustomProvider() bool {
	return p.ProviderType == "custom"
}

// IsEnabled: 제공자가 활성화되어 있는지 확인합니다
func (p *OIDCProvider) IsEnabled() bool {
	return p.Enabled
}

// Enable: 제공자를 활성화합니다
func (p *OIDCProvider) Enable() {
	p.Enabled = true
}

// Disable: 제공자를 비활성화합니다
func (p *OIDCProvider) Disable() {
	p.Enabled = false
}

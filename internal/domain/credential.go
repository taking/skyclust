package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Credential represents a cloud provider credential
type Credential struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Provider      string     `json:"provider" gorm:"not null;size:20;index"` // aws, gcp, openstack, azure
	Name          string     `json:"name" gorm:"not null;size:100"`
	EncryptedData []byte     `json:"-" gorm:"type:bytea;not null"` // 암호화된 자격증명 데이터
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// CredentialRepository defines the interface for credential data operations
type CredentialRepository interface {
	Create(credential *Credential) error
	GetByID(id uuid.UUID) (*Credential, error)
	GetByUserID(userID uuid.UUID) ([]*Credential, error)
	GetByUserIDAndProvider(userID uuid.UUID, provider string) ([]*Credential, error)
	Update(credential *Credential) error
	Delete(id uuid.UUID) error
	DeleteByUserID(userID uuid.UUID) error
}

// CredentialService defines the interface for credential business logic
type CredentialService interface {
	CreateCredential(ctx context.Context, userID uuid.UUID, req CreateCredentialRequest) (*Credential, error)
	GetCredentials(ctx context.Context, userID uuid.UUID) ([]*Credential, error)
	GetCredentialByID(ctx context.Context, userID, credentialID uuid.UUID) (*Credential, error)
	UpdateCredential(ctx context.Context, userID, credentialID uuid.UUID, req UpdateCredentialRequest) (*Credential, error)
	DeleteCredential(ctx context.Context, userID, credentialID uuid.UUID) error
	EncryptCredentialData(ctx context.Context, data map[string]interface{}) ([]byte, error)
	DecryptCredentialData(ctx context.Context, encryptedData []byte) (map[string]interface{}, error)
}

// CreateCredentialRequest represents the request to create a credential
type CreateCredentialRequest struct {
	Provider string                 `json:"provider" validate:"required,oneof=aws gcp openstack azure"`
	Name     string                 `json:"name" validate:"required,min=1,max=100"`
	Data     map[string]interface{} `json:"data" validate:"required"`
}

// UpdateCredentialRequest represents the request to update a credential
type UpdateCredentialRequest struct {
	Name *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// CredentialData represents the structure of credential data for different providers
type CredentialData struct {
	// AWS
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Region    string `json:"region,omitempty"`
	RoleARN   string `json:"role_arn,omitempty"`

	// GCP
	ProjectID       string `json:"project_id,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty"`
	CredentialsJSON string `json:"credentials_json,omitempty"`

	// OpenStack
	AuthURL            string `json:"auth_url,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	OpenStackProjectID string `json:"openstack_project_id,omitempty"`

	// Azure
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
}

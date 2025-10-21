package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Credential represents a cloud provider credential
type Credential struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID        uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	Provider      string                 `json:"provider" gorm:"not null;size:20;index"` // aws, gcp, openstack, azure
	Name          string                 `json:"name" gorm:"not null;size:100"`
	EncryptedData []byte                 `json:"-" gorm:"type:bytea;not null"` // 암호화된 자격증명 데이터
	IsActive      bool                   `json:"is_active" gorm:"default:true"`
	MaskedData    map[string]interface{} `json:"masked_data,omitempty" gorm:"-"` // 마스킹된 데이터 (응답 전용)
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeletedAt     *time.Time             `json:"-" gorm:"index"`

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

// MaskString masks a string, showing only first and last few characters
func MaskString(s string, showFirst, showLast int) string {
	if len(s) <= showFirst+showLast {
		// Too short to mask meaningfully
		return "***"
	}

	masked := s[:showFirst] + "****" + s[len(s)-showLast:]
	return masked
}

// MaskCredentialData masks sensitive data in credential map
func MaskCredentialData(data map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})

	for key, value := range data {
		strValue, ok := value.(string)
		if !ok {
			masked[key] = value
			continue
		}

		switch key {
		case "access_key":
			// AWS Access Key: Show first 4 and last 4 (AKIA****XXXX)
			masked[key] = MaskString(strValue, 4, 4)
		case "secret_key", "password", "client_secret":
			// Secrets: Show first 4 and last 4
			masked[key] = MaskString(strValue, 4, 4)
		case "credentials_json":
			// JSON: Don't show at all
			masked[key] = "****"
		default:
			// Non-sensitive fields: Show as-is
			masked[key] = value
		}
	}

	return masked
}

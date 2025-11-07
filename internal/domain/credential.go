package domain

import (
	"time"

	"github.com/google/uuid"
)

// Credential: 클라우드 제공자 자격증명을 나타내는 도메인 엔티티
type Credential struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	WorkspaceID   uuid.UUID              `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Provider      string                 `json:"provider" gorm:"not null;size:20;index"` // aws, gcp, openstack, azure
	Name          string                 `json:"name" gorm:"not null;size:100"`
	EncryptedData []byte                 `json:"-" gorm:"type:bytea;not null"` // 암호화된 자격증명 데이터
	IsActive      bool                   `json:"is_active" gorm:"default:true"`
	MaskedData    map[string]interface{} `json:"masked_data,omitempty" gorm:"-"` // 마스킹된 데이터 (응답 전용)
	CreatedBy     uuid.UUID              `json:"created_by" gorm:"type:uuid;not null;index"` // 생성한 사용자
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`

	// 관계
	Workspace *Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE"`
}

// CreateCredentialRequest: 자격증명 생성 요청 DTO
type CreateCredentialRequest struct {
	WorkspaceID string                 `json:"workspace_id" validate:"required,uuid"`
	Provider    string                 `json:"provider" validate:"required,oneof=aws gcp openstack azure"`
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	Data        map[string]interface{} `json:"data" validate:"required"`
}

// UpdateCredentialRequest: 자격증명 업데이트 요청 DTO
type UpdateCredentialRequest struct {
	Name *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// CredentialData: 다양한 제공자별 자격증명 데이터 구조
type CredentialData struct {
	// AWS 자격증명 필드
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Region    string `json:"region,omitempty"`
	RoleARN   string `json:"role_arn,omitempty"`

	// GCP 자격증명 필드
	ProjectID       string `json:"project_id,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty"`
	CredentialsJSON string `json:"credentials_json,omitempty"`

	// OpenStack 자격증명 필드
	AuthURL            string `json:"auth_url,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	OpenStackProjectID string `json:"openstack_project_id,omitempty"`

	// Azure 자격증명 필드
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
}

// MaskString: 문자열을 마스킹하여 처음과 끝의 일부 문자만 표시합니다
func MaskString(s string, showFirst, showLast int) string {
	if len(s) <= showFirst+showLast {
		return "***"
	}

	masked := s[:showFirst] + "****" + s[len(s)-showLast:]
	return masked
}

// MaskCredentialData: 자격증명 맵의 민감한 데이터를 마스킹합니다
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
			// AWS Access Key: 처음 4자와 마지막 4자만 표시
			masked[key] = MaskString(strValue, 4, 4)
		case "secret_key", "password", "client_secret":
			// 비밀 정보: 처음 4자와 마지막 4자만 표시
			masked[key] = MaskString(strValue, 4, 4)
		case "private_key":
			// GCP Private Key: 처음 4자와 마지막 4자만 표시
			masked[key] = MaskString(strValue, 4, 4)
		case "private_key_id":
			// GCP Private Key ID: 처음 4자와 마지막 4자만 표시
			masked[key] = MaskString(strValue, 4, 4)
		case "credentials_json":
			// JSON: 전혀 표시하지 않음
			masked[key] = "****"
		default:
			// 민감하지 않은 필드: 그대로 표시
			masked[key] = value
		}
	}

	return masked
}

package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// Service: 자격증명 비즈니스 로직 구현체
type Service struct {
	credentialRepo domain.CredentialRepository
	auditLogRepo   domain.AuditLogRepository
	encryptor      security.Encryptor
	eventPublisher *messaging.Publisher
}

// NewService: 새로운 자격증명 서비스를 생성합니다
func NewService(
	credentialRepo domain.CredentialRepository,
	auditLogRepo domain.AuditLogRepository,
	encryptor security.Encryptor,
	eventPublisher *messaging.Publisher,
) domain.CredentialService {
	return &Service{
		credentialRepo: credentialRepo,
		auditLogRepo:   auditLogRepo,
		encryptor:      encryptor,
		eventPublisher: eventPublisher,
	}
}

// CreateCredential: 새로운 자격증명을 생성합니다 (워크스페이스 기반)
func (s *Service) CreateCredential(ctx context.Context, workspaceID, createdBy uuid.UUID, req domain.CreateCredentialRequest) (*domain.Credential, error) {
	// Validate provider
	if !s.isValidProvider(req.Provider) {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported provider", 400)
	}

	// Validate credential data based on provider
	if err := s.validateCredentialData(req.Provider, req.Data); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "invalid credential data: "+err.Error(), 400)
	}

	// Additional validation for GCP credentials
	if req.Provider == "gcp" {
		if err := s.validateGCPCredentialAccess(ctx, req.Data); err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "GCP credential validation failed: "+err.Error(), 400)
		}
	}

	// Encrypt credential data
	encryptedData, err := s.EncryptCredentialData(ctx, req.Data)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to encrypt credential data", 500)
	}

	// Create credential
	credential := &domain.Credential{
		WorkspaceID:   workspaceID,
		CreatedBy:     createdBy,
		Provider:      req.Provider,
		Name:          req.Name,
		EncryptedData: encryptedData,
		IsActive:      true,
	}

	if err := s.credentialRepo.Create(credential); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to create credential", 500)
	}

	// Log credential creation
	common.LogAction(ctx, s.auditLogRepo, &createdBy, domain.ActionCredentialCreate,
		"POST /api/v1/credentials",
		map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
			"name":          credential.Name,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		credentialData := map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
			"name":          credential.Name,
			"is_active":     credential.IsActive,
		}
		_ = s.eventPublisher.PublishCredentialEvent(ctx, workspaceID.String(), credential.Provider, "created", credentialData)
	}

	return credential, nil
}

// GetCredentials: 워크스페이스의 모든 자격증명을 조회합니다
func (s *Service) GetCredentials(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Credential, error) {
	credentials, err := s.credentialRepo.GetByWorkspaceID(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get credentials", 500)
	}

	// Decrypt and mask credential data for each credential
	for _, credential := range credentials {
		decryptedData, err := s.DecryptCredentialData(ctx, credential.EncryptedData)
		if err != nil {
			// Log error but don't fail the request, just skip masking
			logger.DefaultLogger.GetLogger().Warn("Failed to decrypt credential for masking",
				zap.String("credential_id", credential.ID.String()),
				zap.Error(err))
			continue
		}

		// Add masked data to credential
		credential.MaskedData = domain.MaskCredentialData(decryptedData)
	}

	return credentials, nil
}

// GetCredentialByID: ID로 특정 자격증명을 조회합니다 (워크스페이스 기반)
func (s *Service) GetCredentialByID(ctx context.Context, workspaceID, credentialID uuid.UUID) (*domain.Credential, error) {
	credential, err := s.credentialRepo.GetByID(credentialID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get credential", 500)
	}
	if credential == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "credential not found", 404)
	}

	// Check if credential belongs to workspace
	if credential.WorkspaceID != workspaceID {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "access denied", 403)
	}

	return credential, nil
}

// GetCredentialByIDDirect: ID로 자격증명을 조회합니다 (workspace 검증 없이, 내부 사용용)
func (s *Service) GetCredentialByIDDirect(ctx context.Context, credentialID uuid.UUID) (*domain.Credential, error) {
	credential, err := s.credentialRepo.GetByID(credentialID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get credential", 500)
	}
	if credential == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "credential not found", 404)
	}
	return credential, nil
}

// UpdateCredential: 자격증명을 업데이트합니다 (워크스페이스 기반)
func (s *Service) UpdateCredential(ctx context.Context, workspaceID, credentialID uuid.UUID, req domain.UpdateCredentialRequest) (*domain.Credential, error) {
	// Get existing credential
	credential, err := s.GetCredentialByID(ctx, workspaceID, credentialID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		credential.Name = *req.Name
	}

	if req.Data != nil {
		// Validate credential data
		if err := s.validateCredentialData(credential.Provider, req.Data); err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "invalid credential data: "+err.Error(), 400)
		}

		// Encrypt new data
		encryptedData, err := s.EncryptCredentialData(ctx, req.Data)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to encrypt credential data", 500)
		}
		credential.EncryptedData = encryptedData
	}

	// Save updated credential
	if err := s.credentialRepo.Update(credential); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to update credential", 500)
	}

	// Log credential update
	common.LogAction(ctx, s.auditLogRepo, &credential.CreatedBy, domain.ActionCredentialUpdate,
		"PUT /api/v1/credentials/"+credentialID.String(),
		map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		credentialData := map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
			"name":          credential.Name,
			"is_active":     credential.IsActive,
		}
		_ = s.eventPublisher.PublishCredentialEvent(ctx, workspaceID.String(), credential.Provider, "updated", credentialData)
	}

	return credential, nil
}

// DeleteCredential: 자격증명을 삭제합니다 (워크스페이스 기반)
func (s *Service) DeleteCredential(ctx context.Context, workspaceID, credentialID uuid.UUID) error {
	// Check if credential exists and belongs to workspace
	credential, err := s.GetCredentialByID(ctx, workspaceID, credentialID)
	if err != nil {
		return err
	}

	// Delete credential
	if err := s.credentialRepo.Delete(credentialID); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to delete credential", 500)
	}

	// Log credential deletion
	common.LogAction(ctx, s.auditLogRepo, &credential.CreatedBy, domain.ActionCredentialDelete,
		"DELETE /api/v1/credentials/"+credentialID.String(),
		map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		credentialData := map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
			"name":          credential.Name,
		}
		_ = s.eventPublisher.PublishCredentialEvent(ctx, workspaceID.String(), credential.Provider, "deleted", credentialData)
	}

	return nil
}

// EncryptCredentialData: 자격증명 데이터를 암호화합니다
func (s *Service) EncryptCredentialData(ctx context.Context, data map[string]interface{}) ([]byte, error) {
	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Encrypt using AES
	encrypted, err := s.encryptor.Encrypt(jsonData)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// DecryptCredentialData: 자격증명 데이터를 복호화합니다
func (s *Service) DecryptCredentialData(ctx context.Context, encryptedData []byte) (map[string]interface{}, error) {
	// Check if encrypted data is empty
	if len(encryptedData) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "encrypted data is empty", 400)
	}

	// Decrypt using AES
	decrypted, err := s.encryptor.Decrypt(encryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("decryption failed: %v", err), 500)
	}

	// Debug log (will be removed after fix)
	logger.DefaultLogger.GetLogger().Debug("Decrypted credential data",
		zap.Int("encrypted_length", len(encryptedData)),
		zap.Int("decrypted_length", len(decrypted)),
		zap.String("decrypted_preview", string(decrypted[:min(100, len(decrypted))])))

	// Convert from JSON
	var data map[string]interface{}
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("json unmarshal failed: %v (data: %s)", err, string(decrypted[:min(200, len(decrypted))])), 500)
	}

	return data, nil
}

// min: 두 정수 중 작은 값을 반환합니다
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isValidProvider: 프로바이더가 지원되는지 확인합니다
func (s *Service) isValidProvider(provider string) bool {
	validProviders := []string{"aws", "gcp", "openstack", "azure"}
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

// validateCredentialData: 프로바이더에 따라 자격증명 데이터를 검증합니다
func (s *Service) validateCredentialData(provider string, data map[string]interface{}) error {
	switch provider {
	case "aws":
		return s.validateAWSCredentials(data)
	case "gcp":
		return s.validateGCPCredentials(data)
	case "openstack":
		return s.validateOpenStackCredentials(data)
	case "azure":
		return s.validateAzureCredentials(data)
	default:
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported provider", 400)
	}
}

// validateAWSCredentials: AWS 자격증명 데이터를 검증합니다
func (s *Service) validateAWSCredentials(data map[string]interface{}) error {
	if _, ok := data["access_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key is required for AWS", 400)
	}
	if _, ok := data["secret_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key is required for AWS", 400)
	}
	return nil
}

// validateGCPCredentials: GCP 자격증명 데이터를 검증합니다
func (s *Service) validateGCPCredentials(data map[string]interface{}) error {
	// 필수 필드 검증
	requiredFields := []string{"type", "project_id", "private_key", "client_email", "client_id"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("%s is required for GCP service account", field), 400)
		}
	}

	// service_account 타입 확인
	if data["type"] != "service_account" {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("invalid service account type: %s", data["type"]), 400)
	}

	return nil
}

// validateOpenStackCredentials: OpenStack 자격증명 데이터를 검증합니다
func (s *Service) validateOpenStackCredentials(data map[string]interface{}) error {
	if _, ok := data["auth_url"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "auth_url is required for OpenStack", 400)
	}
	if _, ok := data["username"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "username is required for OpenStack", 400)
	}
	if _, ok := data["password"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "password is required for OpenStack", 400)
	}
	return nil
}

// validateGCPCredentialAccess: GCP 서비스 계정 접근 및 권한을 검증합니다
func (s *Service) validateGCPCredentialAccess(ctx context.Context, data map[string]interface{}) error {
	projectID := data["project_id"].(string)
	clientEmail := data["client_email"].(string)

	// JSON을 직접 사용하여 GCP 클라이언트 생성
	jsonData, err := json.Marshal(data)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal service account data: %v", err), 500)
	}

	// Create GCP clients using JSON credentials
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP container service: %v", err), 502)
	}

	iamService, err := iam.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP IAM service: %v", err), 502)
	}

	// Validate service account exists
	serviceAccountName := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, clientEmail)
	_, err = iamService.Projects.ServiceAccounts.Get(serviceAccountName).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("service account %s not found in project %s", clientEmail, projectID), 404)
	}

	// Validate required permissions by testing API access
	region := "asia-northeast3" // 기본값 또는 요청에서 가져오기
	if err := s.validateGCPPermissions(ctx, containerService, iamService, projectID, region); err != nil {
		return domain.NewDomainError(domain.ErrCodeForbidden, fmt.Sprintf("insufficient permissions: %v", err), 403)
	}

	return nil
}

// validateGCPPermissions: 서비스 계정이 필요한 권한을 가지고 있는지 검증합니다
func (s *Service) validateGCPPermissions(ctx context.Context, containerService *container.Service, iamService *iam.Service, projectID, region string) error {
	// Test GKE API access
	_, err := containerService.Projects.Locations.Clusters.List(fmt.Sprintf("projects/%s/locations/%s", projectID, region)).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("GKE API access failed: %v", err), 502)
	}

	// Test IAM API access
	_, err = iamService.Projects.ServiceAccounts.List(fmt.Sprintf("projects/%s", projectID)).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("IAM API access failed: %v", err), 502)
	}

	return nil
}

// validateAzureCredentials: Azure 자격증명 데이터를 검증합니다
func (s *Service) validateAzureCredentials(data map[string]interface{}) error {
	if _, ok := data["client_id"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "client_id is required for Azure", 400)
	}
	if _, ok := data["client_secret"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "client_secret is required for Azure", 400)
	}
	if _, ok := data["tenant_id"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "tenant_id is required for Azure", 400)
	}
	return nil
}

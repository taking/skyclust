package service

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// credentialService implements the credential business logic
type credentialService struct {
	credentialRepo domain.CredentialRepository
	auditLogRepo   domain.AuditLogRepository
	encryptor      security.Encryptor
}

// NewCredentialService creates a new credential service
func NewCredentialService(
	credentialRepo domain.CredentialRepository,
	auditLogRepo domain.AuditLogRepository,
	encryptor security.Encryptor,
) domain.CredentialService {
	return &credentialService{
		credentialRepo: credentialRepo,
		auditLogRepo:   auditLogRepo,
		encryptor:      encryptor,
	}
}

// CreateCredential creates a new credential
func (s *credentialService) CreateCredential(ctx context.Context, userID uuid.UUID, req domain.CreateCredentialRequest) (*domain.Credential, error) {
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
		UserID:        userID,
		Provider:      req.Provider,
		Name:          req.Name,
		EncryptedData: encryptedData,
		IsActive:      true,
	}

	if err := s.credentialRepo.Create(credential); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to create credential", 500)
	}

	// Log credential creation
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userID,
		Action:   domain.ActionCredentialCreate,
		Resource: "POST /api/v1/credentials",
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"provider":      credential.Provider,
			"name":          credential.Name,
		},
	})

	// TODO: Trigger plugin activation
	// This would be called by the plugin activation service
	// s.pluginActivationService.OnCredentialCreated(userID, credential.Provider)

	return credential, nil
}

// GetCredentials retrieves all credentials for a user
func (s *credentialService) GetCredentials(ctx context.Context, userID uuid.UUID) ([]*domain.Credential, error) {
	credentials, err := s.credentialRepo.GetByUserID(userID)
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

// GetCredentialByID retrieves a specific credential by ID
func (s *credentialService) GetCredentialByID(ctx context.Context, userID, credentialID uuid.UUID) (*domain.Credential, error) {
	credential, err := s.credentialRepo.GetByID(credentialID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get credential", 500)
	}
	if credential == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "credential not found", 404)
	}

	// Check if credential belongs to user
	if credential.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "access denied", 403)
	}

	return credential, nil
}

// UpdateCredential updates a credential
func (s *credentialService) UpdateCredential(ctx context.Context, userID, credentialID uuid.UUID, req domain.UpdateCredentialRequest) (*domain.Credential, error) {
	// Get existing credential
	credential, err := s.GetCredentialByID(ctx, userID, credentialID)
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
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userID,
		Action:   domain.ActionCredentialUpdate,
		Resource: "PUT /api/v1/credentials/" + credentialID.String(),
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"provider":      credential.Provider,
		},
	})

	return credential, nil
}

// DeleteCredential deletes a credential
func (s *credentialService) DeleteCredential(ctx context.Context, userID, credentialID uuid.UUID) error {
	// Check if credential exists and belongs to user
	credential, err := s.GetCredentialByID(ctx, userID, credentialID)
	if err != nil {
		return err
	}

	// Delete credential
	if err := s.credentialRepo.Delete(credentialID); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to delete credential", 500)
	}

	// Log credential deletion
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userID,
		Action:   domain.ActionCredentialDelete,
		Resource: "DELETE /api/v1/credentials/" + credentialID.String(),
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"provider":      credential.Provider,
		},
	})

	return nil
}

// EncryptCredentialData encrypts credential data
func (s *credentialService) EncryptCredentialData(ctx context.Context, data map[string]interface{}) ([]byte, error) {
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

// DecryptCredentialData decrypts credential data
func (s *credentialService) DecryptCredentialData(ctx context.Context, encryptedData []byte) (map[string]interface{}, error) {
	// Check if encrypted data is empty
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("encrypted data is empty")
	}

	// Decrypt using AES
	decrypted, err := s.encryptor.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Debug log (will be removed after fix)
	logger.DefaultLogger.GetLogger().Debug("Decrypted credential data",
		zap.Int("encrypted_length", len(encryptedData)),
		zap.Int("decrypted_length", len(decrypted)),
		zap.String("decrypted_preview", string(decrypted[:min(100, len(decrypted))])))

	// Convert from JSON
	var data map[string]interface{}
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w (data: %s)", err, string(decrypted[:min(200, len(decrypted))]))
	}

	return data, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isValidProvider checks if the provider is supported
func (s *credentialService) isValidProvider(provider string) bool {
	validProviders := []string{"aws", "gcp", "openstack", "azure"}
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

// validateCredentialData validates credential data based on provider
func (s *credentialService) validateCredentialData(provider string, data map[string]interface{}) error {
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

// validateAWSCredentials validates AWS credential data
func (s *credentialService) validateAWSCredentials(data map[string]interface{}) error {
	if _, ok := data["access_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key is required for AWS", 400)
	}
	if _, ok := data["secret_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key is required for AWS", 400)
	}
	return nil
}

// validateGCPCredentials validates GCP credential data
func (s *credentialService) validateGCPCredentials(data map[string]interface{}) error {
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

// validateOpenStackCredentials validates OpenStack credential data
func (s *credentialService) validateOpenStackCredentials(data map[string]interface{}) error {
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

// validateGCPCredentialAccess validates GCP service account access and permissions
func (s *credentialService) validateGCPCredentialAccess(ctx context.Context, data map[string]interface{}) error {
	projectID := data["project_id"].(string)
	clientEmail := data["client_email"].(string)

	// JSON을 직접 사용하여 GCP 클라이언트 생성
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal service account data: %w", err)
	}

	// Create GCP clients using JSON credentials
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create GCP container service: %w", err)
	}

	iamService, err := iam.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create GCP IAM service: %w", err)
	}

	// Validate service account exists
	serviceAccountName := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, clientEmail)
	_, err = iamService.Projects.ServiceAccounts.Get(serviceAccountName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("service account %s not found in project %s", clientEmail, projectID)
	}

	// Validate required permissions by testing API access
	region := "asia-northeast3" // 기본값 또는 요청에서 가져오기
	if err := s.validateGCPPermissions(ctx, containerService, iamService, projectID, region); err != nil {
		return fmt.Errorf("insufficient permissions: %w", err)
	}

	return nil
}

// validateGCPPermissions validates that the service account has required permissions
func (s *credentialService) validateGCPPermissions(ctx context.Context, containerService *container.Service, iamService *iam.Service, projectID, region string) error {
	// Test GKE API access
	_, err := containerService.Projects.Locations.Clusters.List(fmt.Sprintf("projects/%s/locations/%s", projectID, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GKE API access failed: %w", err)
	}

	// Test IAM API access
	_, err = iamService.Projects.ServiceAccounts.List(fmt.Sprintf("projects/%s", projectID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("IAM API access failed: %w", err)
	}

	return nil
}

// validateAzureCredentials validates Azure credential data
func (s *credentialService) validateAzureCredentials(data map[string]interface{}) error {
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

package credential

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

// Service implements the credential business logic
type Service struct {
	credentialRepo domain.CredentialRepository
	auditLogRepo   domain.AuditLogRepository
	encryptor      security.Encryptor
}

// NewService creates a new credential service
func NewService(
	credentialRepo domain.CredentialRepository,
	auditLogRepo domain.AuditLogRepository,
	encryptor security.Encryptor,
) domain.CredentialService {
	return &Service{
		credentialRepo: credentialRepo,
		auditLogRepo:   auditLogRepo,
		encryptor:      encryptor,
	}
}

// CreateCredential creates a new credential (Workspace-based)
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
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   createdBy,
		Action:   domain.ActionCredentialCreate,
		Resource: "POST /api/v1/credentials",
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
			"name":          credential.Name,
		},
	})

	return credential, nil
}

// Deprecated: Use CreateCredential with workspaceID instead
// CreateCredentialByUser creates a new credential (User-based - deprecated)
func (s *Service) CreateCredentialByUser(ctx context.Context, userID uuid.UUID, req domain.CreateCredentialRequest) (*domain.Credential, error) {
	// For backward compatibility, we need to find a workspace for the user
	// This is a temporary solution - in production, credentials should always be workspace-based
	// We'll create a default workspace or use the user's primary workspace
	// For now, return an error indicating this method is deprecated
	return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "CreateCredentialByUser is deprecated. Use CreateCredential with workspaceID instead", 400)
}

// GetCredentials retrieves all credentials for a workspace
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

// Deprecated: Use GetCredentials with workspaceID instead
// GetCredentialsByUser retrieves all credentials for a user (deprecated)
func (s *Service) GetCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Credential, error) {
	// For backward compatibility, search by created_by
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

// GetCredentialByID retrieves a specific credential by ID (Workspace-based)
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

// Deprecated: Use GetCredentialByID with workspaceID instead
// GetCredentialByIDAndUser retrieves a specific credential by ID (User-based - deprecated)
func (s *Service) GetCredentialByIDAndUser(ctx context.Context, userID, credentialID uuid.UUID) (*domain.Credential, error) {
	credential, err := s.credentialRepo.GetByID(credentialID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get credential", 500)
	}
	if credential == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "credential not found", 404)
	}

	// Check if credential was created by user (for backward compatibility)
	if credential.CreatedBy != userID {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "access denied", 403)
	}

	return credential, nil
}

// UpdateCredential updates a credential (Workspace-based)
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
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   credential.CreatedBy,
		Action:   domain.ActionCredentialUpdate,
		Resource: "PUT /api/v1/credentials/" + credentialID.String(),
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
		},
	})

	return credential, nil
}

// Deprecated: Use UpdateCredential with workspaceID instead
// UpdateCredentialByUser updates a credential (User-based - deprecated)
func (s *Service) UpdateCredentialByUser(ctx context.Context, userID, credentialID uuid.UUID, req domain.UpdateCredentialRequest) (*domain.Credential, error) {
	// For backward compatibility, get credential by user
	credential, err := s.GetCredentialByIDAndUser(ctx, userID, credentialID)
	if err != nil {
		return nil, err
	}

	// Update using workspace ID
	return s.UpdateCredential(ctx, credential.WorkspaceID, credentialID, req)
}

// DeleteCredential deletes a credential (Workspace-based)
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
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   credential.CreatedBy,
		Action:   domain.ActionCredentialDelete,
		Resource: "DELETE /api/v1/credentials/" + credentialID.String(),
		Details: map[string]interface{}{
			"credential_id": credential.ID,
			"workspace_id":  workspaceID,
			"provider":      credential.Provider,
		},
	})

	return nil
}

// Deprecated: Use DeleteCredential with workspaceID instead
// DeleteCredentialByUser deletes a credential (User-based - deprecated)
func (s *Service) DeleteCredentialByUser(ctx context.Context, userID, credentialID uuid.UUID) error {
	// For backward compatibility, get credential by user
	credential, err := s.GetCredentialByIDAndUser(ctx, userID, credentialID)
	if err != nil {
		return err
	}

	// Delete using workspace ID
	return s.DeleteCredential(ctx, credential.WorkspaceID, credentialID)
}

// EncryptCredentialData encrypts credential data
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

// DecryptCredentialData decrypts credential data
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

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isValidProvider checks if the provider is supported
func (s *Service) isValidProvider(provider string) bool {
	validProviders := []string{"aws", "gcp", "openstack", "azure"}
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

// validateCredentialData validates credential data based on provider
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

// validateAWSCredentials validates AWS credential data
func (s *Service) validateAWSCredentials(data map[string]interface{}) error {
	if _, ok := data["access_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key is required for AWS", 400)
	}
	if _, ok := data["secret_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key is required for AWS", 400)
	}
	return nil
}

// validateGCPCredentials validates GCP credential data
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

// validateOpenStackCredentials validates OpenStack credential data
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

// validateGCPCredentialAccess validates GCP service account access and permissions
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

// validateGCPPermissions validates that the service account has required permissions
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

// validateAzureCredentials validates Azure credential data
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

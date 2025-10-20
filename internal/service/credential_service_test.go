package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"skyclust/internal/domain"
)

// MockCredentialRepository for testing
type MockCredentialRepository struct {
	mock.Mock
}

func (m *MockCredentialRepository) Create(credential *domain.Credential) error {
	args := m.Called(credential)
	return args.Error(0)
}

func (m *MockCredentialRepository) GetByID(id uuid.UUID) (*domain.Credential, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Credential), args.Error(1)
}

func (m *MockCredentialRepository) GetByUserID(userID uuid.UUID) ([]*domain.Credential, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Credential), args.Error(1)
}

func (m *MockCredentialRepository) Update(credential *domain.Credential) error {
	args := m.Called(credential)
	return args.Error(0)
}

func (m *MockCredentialRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCredentialRepository) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockCredentialRepository) GetByUserIDAndProvider(userID uuid.UUID, provider string) ([]*domain.Credential, error) {
	args := m.Called(userID, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Credential), args.Error(1)
}

// MockEncryptor for testing
type MockEncryptor struct {
	mock.Mock
}

func (m *MockEncryptor) Encrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	args := m.Called(encryptedData)
	return args.Get(0).([]byte), args.Error(1)
}

func TestCredentialService_CreateCredential(t *testing.T) {
	tests := []struct {
		name        string
		userID      uuid.UUID
		request     domain.CreateCredentialRequest
		setupMocks  func(*MockCredentialRepository, *MockEncryptor, *MockAuditLogRepository)
		expectError bool
		errorType   string
	}{
		{
			name:   "successful credential creation",
			userID: uuid.New(),
			request: domain.CreateCredentialRequest{
				Provider: "aws",
				Name:     "AWS Production",
				Data: map[string]interface{}{
					"access_key": "AKIAIOSFODNN7EXAMPLE",
					"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor, auditRepo *MockAuditLogRepository) {
				encryptor.On("Encrypt", mock.AnythingOfType("[]uint8")).Return([]byte("encrypted_data"), nil)
				credRepo.On("Create", mock.AnythingOfType("*domain.Credential")).Return(nil)
				auditRepo.On("Create", mock.AnythingOfType("*domain.AuditLog")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "encryption failure",
			userID: uuid.New(),
			request: domain.CreateCredentialRequest{
				Provider: "aws",
				Name:     "AWS Production",
				Data: map[string]interface{}{
					"access_key": "AKIAIOSFODNN7EXAMPLE",
					"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor, auditRepo *MockAuditLogRepository) {
				encryptor.On("Encrypt", mock.AnythingOfType("[]uint8")).Return([]byte(nil), assert.AnError)
			},
			expectError: true,
			errorType:   "INTERNAL_ERROR",
		},
		{
			name:   "invalid provider",
			userID: uuid.New(),
			request: domain.CreateCredentialRequest{
				Provider: "",
				Name:     "AWS Production",
				Data: map[string]interface{}{
					"access_key": "AKIAIOSFODNN7EXAMPLE",
					"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor, auditRepo *MockAuditLogRepository) {
				// No mock setup needed for validation error
			},
			expectError: true,
			errorType:   "VALIDATION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			credRepo := new(MockCredentialRepository)
			encryptor := new(MockEncryptor)
			auditRepo := new(MockAuditLogRepository)
			tt.setupMocks(credRepo, encryptor, auditRepo)

			// Create service
			service := NewCredentialService(credRepo, auditRepo, encryptor)

			// Execute
			credential, err := service.CreateCredential(context.Background(), tt.userID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, credential)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, credential)
				assert.Equal(t, tt.request.Provider, credential.Provider)
				assert.Equal(t, tt.request.Name, credential.Name)
				assert.Equal(t, tt.userID, credential.UserID)
			}

			// Verify mocks
			credRepo.AssertExpectations(t)
			encryptor.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialService_GetCredentials(t *testing.T) {
	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func(*MockCredentialRepository, *MockEncryptor)
		expectError   bool
		expectedCount int
	}{
		{
			name:   "successful credentials retrieval",
			userID: uuid.New(),
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor) {
				credentials := []*domain.Credential{
					{
						ID:       uuid.New(),
						UserID:   uuid.New(),
						Provider: "aws",
						Name:     "AWS Production",
						IsActive: true,
					},
					{
						ID:       uuid.New(),
						UserID:   uuid.New(),
						Provider: "gcp",
						Name:     "GCP Development",
						IsActive: true,
					},
				}
				credRepo.On("GetByUserID", mock.AnythingOfType("uuid.UUID")).Return(credentials, nil)
				encryptor.On("Decrypt", mock.AnythingOfType("string")).Return("decrypted_data", nil)
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:   "no credentials found",
			userID: uuid.New(),
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor) {
				credRepo.On("GetByUserID", mock.AnythingOfType("uuid.UUID")).Return([]*domain.Credential{}, nil)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:   "database error",
			userID: uuid.New(),
			setupMocks: func(credRepo *MockCredentialRepository, encryptor *MockEncryptor) {
				credRepo.On("GetByUserID", mock.AnythingOfType("uuid.UUID")).Return(nil, assert.AnError)
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			credRepo := new(MockCredentialRepository)
			encryptor := new(MockEncryptor)
			tt.setupMocks(credRepo, encryptor)

			// Create service
			service := NewCredentialService(credRepo, nil, encryptor)

			// Execute
			credentials, err := service.GetCredentials(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, credentials)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, credentials)
				assert.Len(t, credentials, tt.expectedCount)
			}

			// Verify mocks
			credRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialService_DeleteCredential(t *testing.T) {
	tests := []struct {
		name         string
		userID       uuid.UUID
		credentialID uuid.UUID
		setupMocks   func(*MockCredentialRepository, *MockAuditLogRepository)
		expectError  bool
		errorType    string
	}{
		{
			name:         "successful credential deletion",
			userID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			credentialID: uuid.New(),
			setupMocks: func(credRepo *MockCredentialRepository, auditRepo *MockAuditLogRepository) {
				userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				credential := &domain.Credential{
					ID:       uuid.New(),
					UserID:   userID,
					Provider: "aws",
					Name:     "AWS Production",
				}
				credRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(credential, nil)
				credRepo.On("Delete", mock.AnythingOfType("uuid.UUID")).Return(nil)
				auditRepo.On("Create", mock.AnythingOfType("*domain.AuditLog")).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "credential not found",
			userID:       uuid.New(),
			credentialID: uuid.New(),
			setupMocks: func(credRepo *MockCredentialRepository, auditRepo *MockAuditLogRepository) {
				credRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return((*domain.Credential)(nil), assert.AnError)
			},
			expectError: true,
			errorType:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			credRepo := new(MockCredentialRepository)
			auditRepo := new(MockAuditLogRepository)
			tt.setupMocks(credRepo, auditRepo)

			// Create service
			service := NewCredentialService(credRepo, auditRepo, nil)

			// Execute
			err := service.DeleteCredential(context.Background(), tt.userID, tt.credentialID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			credRepo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
		})
	}
}

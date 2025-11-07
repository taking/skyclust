package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// credentialRepository: domain.CredentialRepository 인터페이스 구현체
type credentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository: 새로운 CredentialRepository를 생성합니다
func NewCredentialRepository(db *gorm.DB) domain.CredentialRepository {
	return &credentialRepository{db: db}
}

// Create: 새로운 자격증명을 생성합니다
func (r *credentialRepository) Create(credential *domain.Credential) error {
	if err := r.db.Create(credential).Error; err != nil {
		logger.Errorf("Failed to create credential: %v", err)
		return err
	}
	return nil
}

// GetByID: ID로 자격증명을 조회합니다
func (r *credentialRepository) GetByID(id uuid.UUID) (*domain.Credential, error) {
	var credential domain.Credential
	if err := r.db.Where("id = ?", id).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get credential by ID: %v", err)
		return nil, err
	}
	return &credential, nil
}

// GetByWorkspaceID: 워크스페이스의 모든 자격증명을 조회합니다
func (r *credentialRepository) GetByWorkspaceID(workspaceID uuid.UUID) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := r.db.Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by workspace ID: %v", err)
		return nil, err
	}
	return credentials, nil
}

// GetByWorkspaceIDAndProvider: 워크스페이스와 프로바이더로 자격증명을 조회합니다
func (r *credentialRepository) GetByWorkspaceIDAndProvider(workspaceID uuid.UUID, provider string) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := r.db.Where("workspace_id = ? AND provider = ?", workspaceID, provider).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by workspace ID and provider: %v", err)
		return nil, err
	}
	return credentials, nil
}

// DeleteByWorkspaceID: 워크스페이스의 모든 자격증명을 영구 삭제합니다
func (r *credentialRepository) DeleteByWorkspaceID(workspaceID uuid.UUID) error {
	if err := r.db.Unscoped().Where("workspace_id = ?", workspaceID).Delete(&domain.Credential{}).Error; err != nil {
		logger.Errorf("Failed to delete credentials by workspace ID: %v", err)
		return err
	}
	return nil
}

// Update: 자격증명 정보를 업데이트합니다
func (r *credentialRepository) Update(credential *domain.Credential) error {
	if err := r.db.Save(credential).Error; err != nil {
		logger.Errorf("Failed to update credential: %v", err)
		return err
	}
	return nil
}

// Delete: 자격증명을 영구 삭제합니다
func (r *credentialRepository) Delete(id uuid.UUID) error {
	if err := r.db.Unscoped().Where("id = ?", id).Delete(&domain.Credential{}).Error; err != nil {
		logger.Errorf("Failed to delete credential: %v", err)
		return err
	}
	return nil
}

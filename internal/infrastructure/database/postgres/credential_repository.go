package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// credentialRepository implements the CredentialRepository interface
type credentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository creates a new credential repository
func NewCredentialRepository(db *gorm.DB) domain.CredentialRepository {
	return &credentialRepository{db: db}
}

// Create creates a new credential
func (r *credentialRepository) Create(credential *domain.Credential) error {
	if err := r.db.Create(credential).Error; err != nil {
		logger.Errorf("Failed to create credential: %v", err)
		return err
	}
	return nil
}

// GetByID retrieves a credential by ID
func (r *credentialRepository) GetByID(id uuid.UUID) (*domain.Credential, error) {
	var credential domain.Credential
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get credential by ID: %v", err)
		return nil, err
	}
	return &credential, nil
}

// GetByWorkspaceID retrieves all credentials for a workspace
func (r *credentialRepository) GetByWorkspaceID(workspaceID uuid.UUID) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := r.db.Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by workspace ID: %v", err)
		return nil, err
	}
	return credentials, nil
}

// GetByWorkspaceIDAndProvider retrieves credentials for a workspace and provider
func (r *credentialRepository) GetByWorkspaceIDAndProvider(workspaceID uuid.UUID, provider string) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := r.db.Where("workspace_id = ? AND provider = ? AND deleted_at IS NULL", workspaceID, provider).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by workspace ID and provider: %v", err)
		return nil, err
	}
	return credentials, nil
}

// DeleteByWorkspaceID soft deletes all credentials for a workspace
func (r *credentialRepository) DeleteByWorkspaceID(workspaceID uuid.UUID) error {
	if err := r.db.Where("workspace_id = ?", workspaceID).Delete(&domain.Credential{}).Error; err != nil {
		logger.Errorf("Failed to delete credentials by workspace ID: %v", err)
		return err
	}
	return nil
}

// Deprecated: Use GetByWorkspaceID instead
// GetByUserID retrieves all credentials for a user (deprecated - kept for backward compatibility)
func (r *credentialRepository) GetByUserID(userID uuid.UUID) ([]*domain.Credential, error) {
	// Legacy: This method is deprecated but kept for backward compatibility
	// It now searches by created_by instead of user_id
	var credentials []*domain.Credential
	if err := r.db.Where("created_by = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by user ID (deprecated): %v", err)
		return nil, err
	}
	return credentials, nil
}

// Deprecated: Use GetByWorkspaceIDAndProvider instead
// GetByUserIDAndProvider retrieves credentials for a user and provider (deprecated - kept for backward compatibility)
func (r *credentialRepository) GetByUserIDAndProvider(userID uuid.UUID, provider string) ([]*domain.Credential, error) {
	// Legacy: This method is deprecated but kept for backward compatibility
	// It now searches by created_by instead of user_id
	var credentials []*domain.Credential
	if err := r.db.Where("created_by = ? AND provider = ? AND deleted_at IS NULL", userID, provider).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		logger.Errorf("Failed to get credentials by user ID and provider (deprecated): %v", err)
		return nil, err
	}
	return credentials, nil
}

// Update updates a credential
func (r *credentialRepository) Update(credential *domain.Credential) error {
	if err := r.db.Save(credential).Error; err != nil {
		logger.Errorf("Failed to update credential: %v", err)
		return err
	}
	return nil
}

// Delete soft deletes a credential
func (r *credentialRepository) Delete(id uuid.UUID) error {
	if err := r.db.Where("id = ?", id).Delete(&domain.Credential{}).Error; err != nil {
		logger.Errorf("Failed to delete credential: %v", err)
		return err
	}
	return nil
}

// Deprecated: Use DeleteByWorkspaceID instead
// DeleteByUserID soft deletes all credentials for a user (deprecated - kept for backward compatibility)
func (r *credentialRepository) DeleteByUserID(userID uuid.UUID) error {
	// Legacy: This method is deprecated but kept for backward compatibility
	// It now deletes by created_by instead of user_id
	if err := r.db.Where("created_by = ?", userID).Delete(&domain.Credential{}).Error; err != nil {
		logger.Errorf("Failed to delete credentials by user ID (deprecated): %v", err)
		return err
	}
	return nil
}

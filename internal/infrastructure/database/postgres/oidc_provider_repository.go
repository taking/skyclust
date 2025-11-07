package postgres

import (
	"encoding/base64"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// oidcProviderRepository: domain.OIDCProviderRepository 인터페이스 구현체
type oidcProviderRepository struct {
	db        *gorm.DB
	encryptor security.Encryptor
}

// NewOIDCProviderRepository: 새로운 OIDCProviderRepository를 생성합니다
func NewOIDCProviderRepository(db *gorm.DB, encryptor security.Encryptor) domain.OIDCProviderRepository {
	return &oidcProviderRepository{
		db:        db,
		encryptor: encryptor,
	}
}

// Create: 새로운 OIDC 프로바이더를 생성합니다
func (r *oidcProviderRepository) Create(provider *domain.OIDCProvider) error {
	// Encrypt client secret before saving
	if provider.ClientSecret != "" {
		encrypted, err := r.encryptor.Encrypt([]byte(provider.ClientSecret))
		if err != nil {
			logger.Errorf("Failed to encrypt client secret: %v", err)
			return fmt.Errorf("failed to encrypt client secret: %w", err)
		}
		// Store encrypted secret as base64-encoded string
		provider.ClientSecret = base64.StdEncoding.EncodeToString(encrypted)
	}

	if err := r.db.Create(provider).Error; err != nil {
		logger.Errorf("Failed to create OIDC provider: %v", err)
		return err
	}
	return nil
}

// GetByID: ID로 OIDC 프로바이더를 조회합니다
func (r *oidcProviderRepository) GetByID(id uuid.UUID) (*domain.OIDCProvider, error) {
	var provider domain.OIDCProvider
	if err := r.db.Where("id = ?", id).First(&provider).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get OIDC provider by ID: %v", err)
		return nil, err
	}

	// Decrypt client secret
	if provider.ClientSecret != "" {
		decrypted, err := r.decryptClientSecret(provider.ClientSecret)
		if err != nil {
			logger.Errorf("Failed to decrypt client secret: %v", err)
			return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
		}
		provider.ClientSecret = decrypted
	}

	return &provider, nil
}

// GetByUserID: 사용자의 모든 OIDC 프로바이더를 조회합니다
func (r *oidcProviderRepository) GetByUserID(userID uuid.UUID) ([]*domain.OIDCProvider, error) {
	var providers []*domain.OIDCProvider
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&providers).Error; err != nil {
		logger.Errorf("Failed to get OIDC providers by user ID: %v", err)
		return nil, err
	}

	// Decrypt client secrets for all providers
	for i := range providers {
		if providers[i].ClientSecret != "" {
			decrypted, err := r.decryptClientSecret(providers[i].ClientSecret)
			if err != nil {
				logger.Errorf("Failed to decrypt client secret for provider %s: %v", providers[i].ID, err)
				return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
			}
			providers[i].ClientSecret = decrypted
		}
	}

	return providers, nil
}

// GetByUserIDAndName: 사용자 ID와 이름으로 OIDC 프로바이더를 조회합니다
func (r *oidcProviderRepository) GetByUserIDAndName(userID uuid.UUID, name string) (*domain.OIDCProvider, error) {
	var provider domain.OIDCProvider
	if err := r.db.Where("user_id = ? AND name = ?", userID, name).
		First(&provider).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get OIDC provider by user ID and name: %v", err)
		return nil, err
	}

	// Decrypt client secret
	if provider.ClientSecret != "" {
		decrypted, err := r.decryptClientSecret(provider.ClientSecret)
		if err != nil {
			logger.Errorf("Failed to decrypt client secret: %v", err)
			return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
		}
		provider.ClientSecret = decrypted
	}

	return &provider, nil
}

// Update: OIDC 프로바이더 정보를 업데이트합니다
func (r *oidcProviderRepository) Update(provider *domain.OIDCProvider) error {
	// Encrypt client secret if it's being updated
	if provider.ClientSecret != "" {
		// Check if it's already encrypted (base64 format) by trying to decrypt
		// If decryption fails, it means it's a plain text that needs encryption
		if _, err := base64.StdEncoding.DecodeString(provider.ClientSecret); err == nil {
			// Already base64 encoded, try to decrypt to verify
			if _, err := r.decryptClientSecret(provider.ClientSecret); err != nil {
				// Not encrypted, encrypt it
				encrypted, err := r.encryptor.Encrypt([]byte(provider.ClientSecret))
				if err != nil {
					logger.Errorf("Failed to encrypt client secret: %v", err)
					return fmt.Errorf("failed to encrypt client secret: %w", err)
				}
				provider.ClientSecret = base64.StdEncoding.EncodeToString(encrypted)
			}
			// Already encrypted, keep as is
		} else {
			// Plain text, encrypt it
			encrypted, err := r.encryptor.Encrypt([]byte(provider.ClientSecret))
			if err != nil {
				logger.Errorf("Failed to encrypt client secret: %v", err)
				return fmt.Errorf("failed to encrypt client secret: %w", err)
			}
			provider.ClientSecret = base64.StdEncoding.EncodeToString(encrypted)
		}
	}

	if err := r.db.Save(provider).Error; err != nil {
		logger.Errorf("Failed to update OIDC provider: %v", err)
		return err
	}
	return nil
}

// Delete: OIDC 프로바이더를 영구 삭제합니다
func (r *oidcProviderRepository) Delete(id uuid.UUID) error {
	if err := r.db.Unscoped().Where("id = ?", id).
		Delete(&domain.OIDCProvider{}).Error; err != nil {
		logger.Errorf("Failed to delete OIDC provider: %v", err)
		return err
	}
	return nil
}

// GetEnabledByUserID: 사용자의 활성화된 OIDC 프로바이더를 조회합니다
func (r *oidcProviderRepository) GetEnabledByUserID(userID uuid.UUID) ([]*domain.OIDCProvider, error) {
	var providers []*domain.OIDCProvider
	if err := r.db.Where("user_id = ? AND enabled = true", userID).
		Order("created_at DESC").
		Find(&providers).Error; err != nil {
		logger.Errorf("Failed to get enabled OIDC providers by user ID: %v", err)
		return nil, err
	}

	// Decrypt client secrets for all providers
	for i := range providers {
		if providers[i].ClientSecret != "" {
			decrypted, err := r.decryptClientSecret(providers[i].ClientSecret)
			if err != nil {
				logger.Errorf("Failed to decrypt client secret for provider %s: %v", providers[i].ID, err)
				return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
			}
			providers[i].ClientSecret = decrypted
		}
	}

	return providers, nil
}

// decryptClientSecret: base64로 인코딩된 암호화된 클라이언트 시크릿을 복호화합니다
func (r *oidcProviderRepository) decryptClientSecret(encryptedSecret string) (string, error) {
	if encryptedSecret == "" {
		return "", fmt.Errorf("encrypted secret is empty")
	}

	// Decode from base64
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(encryptedBytes) == 0 {
		return "", fmt.Errorf("decoded secret is empty")
	}

	// Decrypt
	decryptedBytes, err := r.encryptor.Decrypt(encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(decryptedBytes), nil
}


/**
 * Notification Preferences Repository
 * 알림 설정 관련 PostgreSQL 저장소 구현
 */

package postgres

import (
	"context"
	"skyclust/internal/domain"

	"gorm.io/gorm"
)

type notificationPreferencesRepository struct {
	db *gorm.DB
}

// NewNotificationPreferencesRepository 알림 설정 저장소 생성
func NewNotificationPreferencesRepository(db *gorm.DB) domain.NotificationPreferencesRepository {
	return &notificationPreferencesRepository{db: db}
}

// GetByUserID 사용자별 알림 설정 조회
func (r *notificationPreferencesRepository) GetByUserID(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	var preferences domain.NotificationPreferences
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&preferences).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 기본 설정 생성
			defaultPreferences := &domain.NotificationPreferences{
				UserID:                userID,
				EmailEnabled:          true,
				PushEnabled:           true,
				BrowserEnabled:        true,
				InAppEnabled:          true,
				SystemNotifications:   true,
				VMNotifications:       true,
				CostNotifications:     true,
				SecurityNotifications: true,
				LowPriorityEnabled:    true,
				MediumPriorityEnabled: true,
				HighPriorityEnabled:   true,
				UrgentPriorityEnabled: true,
				Timezone:              "UTC",
			}

			if err := r.Create(ctx, defaultPreferences); err != nil {
				return nil, err
			}

			return defaultPreferences, nil
		}
		return nil, err
	}

	return &preferences, nil
}

// Create 알림 설정 생성
func (r *notificationPreferencesRepository) Create(ctx context.Context, preferences *domain.NotificationPreferences) error {
	return r.db.WithContext(ctx).Create(preferences).Error
}

// Update 알림 설정 업데이트
func (r *notificationPreferencesRepository) Update(ctx context.Context, preferences *domain.NotificationPreferences) error {
	return r.db.WithContext(ctx).Save(preferences).Error
}

// Upsert 알림 설정 생성 또는 업데이트
func (r *notificationPreferencesRepository) Upsert(ctx context.Context, preferences *domain.NotificationPreferences) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", preferences.UserID).
		Assign(preferences).
		FirstOrCreate(preferences).Error
}

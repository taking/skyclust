/**
 * Notification Repository
 * 알림 관련 PostgreSQL 저장소 구현
 */

package postgres

import (
	"skyclust/internal/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 알림 저장소 생성
func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

// Create 알림 생성
func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID 알림 ID로 조회
func (r *notificationRepository) GetByID(ctx context.Context, userID, notificationID string) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &notification, nil
}

// GetByUserID 사용자별 알림 목록 조회
func (r *notificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int, unreadOnly bool, category, priority string) ([]*domain.Notification, int, error) {
	var notifications []*domain.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)

	// 필터 적용
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// 총 개수 조회
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 데이터 조회
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	return notifications, int(total), err
}

// Update 알림 업데이트
func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete 알림 삭제
func (r *notificationRepository) Delete(ctx context.Context, userID, notificationID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&domain.Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteMultiple 여러 알림 삭제
func (r *notificationRepository) DeleteMultiple(ctx context.Context, userID string, notificationIDs []string) error {
	return r.db.WithContext(ctx).
		Where("id IN ? AND user_id = ?", notificationIDs, userID).
		Delete(&domain.Notification{}).Error
}

// MarkAsRead 알림 읽음 처리
func (r *notificationRepository) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// MarkAllAsRead 모든 알림 읽음 처리
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// GetStats 알림 통계 조회
func (r *notificationRepository) GetStats(ctx context.Context, userID string) (*domain.NotificationStats, error) {
	var stats domain.NotificationStats

	// 기본 통계
	var totalCount, unreadCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats.TotalNotifications = int(totalCount)

	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount).Error; err != nil {
		return nil, err
	}
	stats.UnreadNotifications = int(unreadCount)

	stats.ReadNotifications = stats.TotalNotifications - stats.UnreadNotifications

	// 카테고리별 통계
	categories := []string{"system", "vm", "cost", "security"}
	for _, category := range categories {
		var count int64
		if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
			Where("user_id = ? AND category = ?", userID, category).
			Count(&count).Error; err != nil {
			return nil, err
		}

		switch category {
		case "system":
			stats.SystemCount = int(count)
		case "vm":
			stats.VMCount = int(count)
		case "cost":
			stats.CostCount = int(count)
		case "security":
			stats.SecurityCount = int(count)
		}
	}

	// 우선순위별 통계
	priorities := []string{"low", "medium", "high", "urgent"}
	for _, priority := range priorities {
		var count int64
		if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
			Where("user_id = ? AND priority = ?", userID, priority).
			Count(&count).Error; err != nil {
			return nil, err
		}

		switch priority {
		case "low":
			stats.LowPriorityCount = int(count)
		case "medium":
			stats.MediumPriorityCount = int(count)
		case "high":
			stats.HighPriorityCount = int(count)
		case "urgent":
			stats.UrgentPriorityCount = int(count)
		}
	}

	// 최근 7일 통계
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var last7DaysCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, sevenDaysAgo).
		Count(&last7DaysCount).Error; err != nil {
		return nil, err
	}
	stats.Last7DaysCount = int(last7DaysCount)

	// 최근 30일 통계
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var last30DaysCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Count(&last30DaysCount).Error; err != nil {
		return nil, err
	}
	stats.Last30DaysCount = int(last30DaysCount)

	return &stats, nil
}

// CleanupOld 오래된 알림 정리
func (r *notificationRepository) CleanupOld(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&domain.Notification{}).Error
}

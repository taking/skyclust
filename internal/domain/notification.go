/**
 * Notification Domain Models
 * 알림 관련 도메인 모델
 */

package domain

import (
	"time"
)

// Notification 알림 모델
type Notification struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	UserID    string     `json:"user_id" gorm:"not null;index"`
	Type      string     `json:"type" gorm:"not null"` // info, warning, error, success
	Title     string     `json:"title" gorm:"not null"`
	Message   string     `json:"message" gorm:"not null"`
	Category  string     `json:"category"` // system, vm, cost, security, etc.
	Priority  string     `json:"priority"` // low, medium, high, urgent
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	Data      string     `json:"data"` // JSON metadata
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// NotificationPreferences 알림 설정 모델
type NotificationPreferences struct {
	ID             string `json:"id" gorm:"primaryKey"`
	UserID         string `json:"user_id" gorm:"not null;uniqueIndex"`
	EmailEnabled   bool   `json:"email_enabled" gorm:"default:true"`
	PushEnabled    bool   `json:"push_enabled" gorm:"default:true"`
	BrowserEnabled bool   `json:"browser_enabled" gorm:"default:true"`
	InAppEnabled   bool   `json:"in_app_enabled" gorm:"default:true"`

	// 카테고리별 설정
	SystemNotifications   bool `json:"system_notifications" gorm:"default:true"`
	VMNotifications       bool `json:"vm_notifications" gorm:"default:true"`
	CostNotifications     bool `json:"cost_notifications" gorm:"default:true"`
	SecurityNotifications bool `json:"security_notifications" gorm:"default:true"`

	// 우선순위별 설정
	LowPriorityEnabled    bool `json:"low_priority_enabled" gorm:"default:true"`
	MediumPriorityEnabled bool `json:"medium_priority_enabled" gorm:"default:true"`
	HighPriorityEnabled   bool `json:"high_priority_enabled" gorm:"default:true"`
	UrgentPriorityEnabled bool `json:"urgent_priority_enabled" gorm:"default:true"`

	// 시간 설정
	QuietHoursStart string `json:"quiet_hours_start"` // HH:MM format
	QuietHoursEnd   string `json:"quiet_hours_end"`   // HH:MM format
	Timezone        string `json:"timezone" gorm:"default:'UTC'"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationStats 알림 통계 모델
type NotificationStats struct {
	TotalNotifications  int `json:"total_notifications"`
	UnreadNotifications int `json:"unread_notifications"`
	ReadNotifications   int `json:"read_notifications"`

	// 카테고리별 통계
	SystemCount   int `json:"system_count"`
	VMCount       int `json:"vm_count"`
	CostCount     int `json:"cost_count"`
	SecurityCount int `json:"security_count"`

	// 우선순위별 통계
	LowPriorityCount    int `json:"low_priority_count"`
	MediumPriorityCount int `json:"medium_priority_count"`
	HighPriorityCount   int `json:"high_priority_count"`
	UrgentPriorityCount int `json:"urgent_priority_count"`

	// 최근 7일 통계
	Last7DaysCount  int `json:"last_7_days_count"`
	Last30DaysCount int `json:"last_30_days_count"`
}

// CreateNotificationRequest 알림 생성 요청
type CreateNotificationRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=info warning error success"`
	Title    string `json:"title" binding:"required"`
	Message  string `json:"message" binding:"required"`
	Category string `json:"category"`
	Priority string `json:"priority" binding:"oneof=low medium high urgent"`
	Data     string `json:"data"`
}

// UpdateNotificationPreferencesRequest 알림 설정 업데이트 요청
type UpdateNotificationPreferencesRequest struct {
	EmailEnabled          *bool  `json:"email_enabled"`
	PushEnabled           *bool  `json:"push_enabled"`
	BrowserEnabled        *bool  `json:"browser_enabled"`
	InAppEnabled          *bool  `json:"in_app_enabled"`
	SystemNotifications   *bool  `json:"system_notifications"`
	VMNotifications       *bool  `json:"vm_notifications"`
	CostNotifications     *bool  `json:"cost_notifications"`
	SecurityNotifications *bool  `json:"security_notifications"`
	LowPriorityEnabled    *bool  `json:"low_priority_enabled"`
	MediumPriorityEnabled *bool  `json:"medium_priority_enabled"`
	HighPriorityEnabled   *bool  `json:"high_priority_enabled"`
	UrgentPriorityEnabled *bool  `json:"urgent_priority_enabled"`
	QuietHoursStart       string `json:"quiet_hours_start"`
	QuietHoursEnd         string `json:"quiet_hours_end"`
	Timezone              string `json:"timezone"`
}


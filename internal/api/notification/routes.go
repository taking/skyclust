package notification

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up notification routes
func SetupRoutes(router *gin.RouterGroup, notificationService domain.NotificationService) {
	notificationHandler := NewHandler(notificationService)

	notifications := router.Group("/notifications")
	{
		// 알림 목록 조회
		notifications.GET("", notificationHandler.GetNotifications)

		// 알림 상세 조회
		notifications.GET("/:id", notificationHandler.GetNotification)

		// 알림 읽음 처리
		notifications.PUT("/:id/read", notificationHandler.MarkAsRead)

		// 알림 읽음 처리 (여러 개)
		notifications.PUT("/read", notificationHandler.MarkAllAsRead)

		// 알림 삭제
		notifications.DELETE("/:id", notificationHandler.DeleteNotification)

		// 알림 삭제 (여러 개)
		notifications.DELETE("", notificationHandler.DeleteNotifications)

		// 알림 설정 조회
		notifications.GET("/preferences", notificationHandler.GetNotificationPreferences)

		// 알림 설정 업데이트
		notifications.PUT("/preferences", notificationHandler.UpdateNotificationPreferences)

		// 알림 통계 조회
		notifications.GET("/stats", notificationHandler.GetNotificationStats)

		// 알림 테스트 (개발용)
		notifications.POST("/test", notificationHandler.SendTestNotification)
	}
}

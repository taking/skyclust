/**
 * Notification Routes
 * 알림 관련 HTTP 라우트 설정
 */

package routes

import (
	"github.com/gin-gonic/gin"
	"skyclust/internal/delivery/http"
)

// SetupNotificationRoutes 알림 관련 라우트 설정
func SetupNotificationRoutes(r *gin.RouterGroup, handler *http.NotificationHandler) {
	notifications := r.Group("/notifications")
	{
		// 알림 목록 조회
		notifications.GET("", handler.GetNotifications)

		// 알림 상세 조회
		notifications.GET("/:id", handler.GetNotification)

		// 알림 읽음 처리
		notifications.PUT("/:id/read", handler.MarkAsRead)

		// 알림 읽음 처리 (여러 개)
		notifications.PUT("/read", handler.MarkAllAsRead)

		// 알림 삭제
		notifications.DELETE("/:id", handler.DeleteNotification)

		// 알림 삭제 (여러 개)
		notifications.DELETE("", handler.DeleteNotifications)

		// 알림 설정 조회
		notifications.GET("/preferences", handler.GetNotificationPreferences)

		// 알림 설정 업데이트
		notifications.PUT("/preferences", handler.UpdateNotificationPreferences)

		// 알림 통계 조회
		notifications.GET("/stats", handler.GetNotificationStats)

		// 알림 테스트 (개발용)
		notifications.POST("/test", handler.SendTestNotification)
	}
}

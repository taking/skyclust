package dashboard

import (
	dashboardservice "skyclust/internal/application/services/dashboard"

	"github.com/gin-gonic/gin"
)

/**
 * Dashboard Routes
 * 대시보드 관련 라우트 설정
 */

// SetupRoutes 대시보드 라우트 설정
func SetupRoutes(router *gin.RouterGroup, dashboardService *dashboardservice.Service) {
	handler := NewHandler(dashboardService)

	// 대시보드 요약 정보 조회
	// GET /api/v1/dashboard/summary?workspace_id={workspace_id}&credential_id={credential_id}&region={region}
	router.GET("/summary", handler.GetDashboardSummary)
}

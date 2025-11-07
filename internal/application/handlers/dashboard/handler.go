package dashboard

import (
	dashboardservice "skyclust/internal/application/services/dashboard"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// Handler: 대시보드 관련 HTTP 요청을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	dashboardService *dashboardservice.Service
}

// NewHandler: 새로운 대시보드 핸들러를 생성합니다
func NewHandler(dashboardService *dashboardservice.Service) *Handler {
	return &Handler{
		BaseHandler:      handlers.NewBaseHandler("dashboard"),
		dashboardService: dashboardService,
	}
}

// GetDashboardSummary: 대시보드 요약 정보를 조회합니다
// GET /api/v1/dashboard/summary?workspace_id={workspace_id}&credential_id={credential_id}&region={region}
func (h *Handler) GetDashboardSummary(c *gin.Context) {
	handler := h.Compose(
		h.getDashboardSummaryHandler(),
		h.StandardCRUDDecorators("get_dashboard_summary")...,
	)

	handler(c)
}

// getDashboardSummaryHandler: 대시보드 요약 정보 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getDashboardSummaryHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// 워크스페이스 ID 조회 (필수)
		workspaceID := c.Query("workspace_id")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "workspace_id is required", 400), "get_dashboard_summary")
			return
		}

		// 자격 증명 ID 조회 (선택)
		credentialID := c.Query("credential_id")
		var credentialIDPtr *string
		if credentialID != "" {
			credentialIDPtr = &credentialID
		}

		// 리전 조회 (선택)
		region := c.Query("region")
		var regionPtr *string
		if region != "" {
			regionPtr = &region
		}

		// 대시보드 요약 정보 조회
		summary, err := h.dashboardService.GetDashboardSummary(c.Request.Context(), workspaceID, credentialIDPtr, regionPtr)
		if err != nil {
			h.HandleError(c, err, "get_dashboard_summary")
			return
		}

		h.OK(c, summary, "Dashboard summary retrieved successfully")
	}
}

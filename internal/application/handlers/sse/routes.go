package sse

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up SSE routes
func SetupRoutes(router *gin.RouterGroup, sseHandler *SSEHandler) {
	if sseHandler == nil {
		return
	}

	// SSE endpoint (인증 필요)
	router.GET("/events", sseHandler.HandleSSE)

	// 구독 관리 엔드포인트
	router.POST("/subscribe", sseHandler.HandleSubscribeToEvent)
	router.POST("/unsubscribe", sseHandler.HandleUnsubscribeFromEvent)
	
	// 연결 정보 조회 엔드포인트
	router.GET("/connection", sseHandler.HandleGetConnectionInfo)
}

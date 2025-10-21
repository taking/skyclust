package sse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up SSE routes
func SetupRoutes(router *gin.RouterGroup) {
	// SSE endpoint (인증 필요)
	router.GET("/events", func(c *gin.Context) {
		// SSE 핸들러는 별도 구현 필요
		// TODO: Add authentication middleware
		c.JSON(http.StatusNotImplemented, gin.H{"message": "SSE endpoint will be implemented"})
	})
}

package sse

import (
    "github.com/gin-gonic/gin"
    "skyclust/internal/shared/responses"
)

// SetupRoutes sets up SSE routes
func SetupRoutes(router *gin.RouterGroup) {
	// SSE endpoint (인증 필요)
	router.GET("/events", func(c *gin.Context) {
        responses.NewResponseBuilder(c).
            WithError("not_implemented", "SSE endpoint is not implemented").
            WithMessage("Not Implemented").
            Send(501)
	})
}

package handlers

import (
	"net/http"

	"cmp/pkg/realtime"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(realtimeService realtime.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		// Upgrade to WebSocket
		conn, err := realtimeService.UpgradeToWebSocket(c.Writer, c.Request, userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
			return
		}

		// Handle WebSocket connection
		realtimeService.HandleWebSocket(conn)
	}
}

// SSEHandler handles Server-Sent Events
func SSEHandler(realtimeService realtime.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		userID, _ := c.Get("user_id")

		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// Create SSE connection
		conn, err := realtimeService.CreateSSEConnection(c.Writer, c.Request, userID.(string), workspaceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create SSE connection"})
			return
		}

		// Handle SSE connection
		realtimeService.HandleSSE(conn)
	}
}


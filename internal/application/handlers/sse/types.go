package sse

import "time"

// SSEConnectionResponse represents an SSE connection response
type SSEConnectionResponse struct {
	ConnectionID string `json:"connection_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// SSEMessage represents a Server-Sent Event message
type SSEMessage struct {
	ID        string                 `json:"id"`
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// SSEEventTypes represents available SSE event types
type SSEEventTypes struct {
	EventTypes []string `json:"event_types"`
}

// SSEStatusResponse represents SSE connection status
type SSEStatusResponse struct {
	Connected    bool      `json:"connected"`
	ConnectionID string    `json:"connection_id"`
	LastPing     time.Time `json:"last_ping"`
	EventCount   int64     `json:"event_count"`
}

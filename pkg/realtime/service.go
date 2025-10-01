package realtime

import (
	"context"
	"fmt"
	"net/http"

	"cmp/internal/infrastructure/messaging"

	"github.com/gorilla/websocket"
)

// Service defines the realtime service interface
type Service interface {
	// WebSocket management
	UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, userID string) (*websocket.Conn, error)
	HandleWebSocket(conn *websocket.Conn)

	// SSE management
	CreateSSEConnection(w http.ResponseWriter, r *http.Request, userID, workspaceID string) (*SSEConnection, error)
	HandleSSE(conn *SSEConnection)

	// Event broadcasting
	BroadcastToWorkspace(workspaceID string, event *messaging.Event) error
	BroadcastToUser(userID string, event *messaging.Event) error
}

// SSEConnection represents a Server-Sent Events connection
type SSEConnection struct {
	UserID      string
	WorkspaceID string
	Writer      http.ResponseWriter
	Request     *http.Request
	Done        chan bool
}

// NewService creates a new realtime service
func NewService(eventBus messaging.Bus) Service {
	return &service{
		eventBus: eventBus,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

type service struct {
	eventBus messaging.Bus
	upgrader websocket.Upgrader
}

// UpgradeToWebSocket upgrades HTTP connection to WebSocket
func (s *service) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, userID string) (*websocket.Conn, error) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// HandleWebSocket handles WebSocket connections
func (s *service) HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Echo message back (in production, handle different message types)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

// CreateSSEConnection creates a Server-Sent Events connection
func (s *service) CreateSSEConnection(w http.ResponseWriter, r *http.Request, userID, workspaceID string) (*SSEConnection, error) {
	conn := &SSEConnection{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Writer:      w,
		Request:     r,
		Done:        make(chan bool),
	}

	return conn, nil
}

// HandleSSE handles Server-Sent Events connections
func (s *service) HandleSSE(conn *SSEConnection) {
	defer close(conn.Done)

	// Send initial connection event
	conn.SendEvent("connected", map[string]interface{}{
		"message": "Connected to real-time updates",
	})

	// Keep connection alive
	<-conn.Done
}

// BroadcastToWorkspace broadcasts an event to all users in a workspace
func (s *service) BroadcastToWorkspace(workspaceID string, event *messaging.Event) error {
	return s.eventBus.PublishToWorkspace(context.Background(), workspaceID, event)
}

// BroadcastToUser broadcasts an event to a specific user
func (s *service) BroadcastToUser(userID string, event *messaging.Event) error {
	return s.eventBus.PublishToUser(context.Background(), userID, event)
}

// SendEvent sends an event to the SSE connection
func (conn *SSEConnection) SendEvent(eventType string, data map[string]interface{}) {
	// Format as Server-Sent Events
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, data)
	//nolint:staticcheck // SA9003: intentionally ignore write errors to prevent SSE connection failure
	if _, err := conn.Writer.Write([]byte(event)); err != nil {
		// Log error but don't fail the connection
		// TODO: Add proper logging here
	}
	conn.Writer.(http.Flusher).Flush()
}

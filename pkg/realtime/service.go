package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"skyclust/internal/infrastructure/messaging"

	"go.uber.org/zap"
)

// Service defines the realtime service interface
// Realtime 서비스 인터페이스 정의
type Service interface {
	// SSE connection management
	// SSE 연결 관리
	CreateSSEConnection(w http.ResponseWriter, r *http.Request, userID, workspaceID string) (*SSEConnection, error)
	HandleSSE(conn *SSEConnection)
	RemoveConnection(connID string) error
	GetConnection(connID string) (*SSEConnection, error)
	GetConnectionCount() int

	// Event broadcasting
	// 이벤트 브로드캐스팅
	BroadcastToWorkspace(workspaceID string, event *messaging.Event) error
	BroadcastToUser(userID string, event *messaging.Event) error
	BroadcastToConnection(connID string, eventType string, data interface{}) error
	BroadcastToAll(eventType string, data interface{}) error

	// Connection subscription management
	// 연결 구독 관리
	SubscribeToEvent(connID, eventType string) error
	UnsubscribeFromEvent(connID, eventType string) error
	SubscribeToResource(connID, resourceType, resourceID string) error
	UnsubscribeFromResource(connID, resourceType, resourceID string) error

	// Cleanup
	// 정리
	CleanupInactiveConnections() error
	StartCleanupRoutine(ctx context.Context)
}

// SSEConnection represents a Server-Sent Events connection
// Server-Sent Events 연결을 나타냄
type SSEConnection struct {
	ID                  string
	UserID              string
	WorkspaceID         string
	Writer              http.ResponseWriter
	Flusher             http.Flusher
	Request             *http.Request
	Context             context.Context
	Cancel              context.CancelFunc
	LastSeen            time.Time
	SubscribedEvents    map[string]bool
	SubscribedResources map[string]map[string]bool // resourceType -> resourceID -> bool
	mu                  sync.RWMutex
}

// NewService creates a new realtime service
// 새로운 realtime 서비스 생성
func NewService(eventBus messaging.Bus, logger *zap.Logger) Service {
	return &service{
		eventBus:  eventBus,
		logger:    logger,
		connections: make(map[string]*SSEConnection),
		connectionsMux: sync.RWMutex{},
	}
}

type service struct {
	eventBus        messaging.Bus
	logger          *zap.Logger
	connections     map[string]*SSEConnection
	connectionsMux  sync.RWMutex
	cleanupInterval time.Duration
	clientTimeout time.Duration
}

// CreateSSEConnection creates a Server-Sent Events connection
// Server-Sent Events 연결 생성
func (s *service) CreateSSEConnection(w http.ResponseWriter, r *http.Request, userID, workspaceID string) (*SSEConnection, error) {
	// Flusher 확인
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("SSE not supported: ResponseWriter does not implement http.Flusher")
	}

	// 컨텍스트 생성
	ctx, cancel := context.WithCancel(r.Context())

	// 연결 ID 생성
	connID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

	// 연결 생성
	conn := &SSEConnection{
		ID:                  connID,
		UserID:              userID,
		WorkspaceID:         workspaceID,
		Writer:              w,
		Flusher:             flusher,
		Request:             r,
		Context:             ctx,
		Cancel:              cancel,
		LastSeen:            time.Now(),
		SubscribedEvents:    make(map[string]bool),
		SubscribedResources: make(map[string]map[string]bool),
	}

	// 연결 등록
	s.connectionsMux.Lock()
	s.connections[connID] = conn
	s.connectionsMux.Unlock()

	s.logger.Info("SSE connection created",
		zap.String("connection_id", connID),
		zap.String("user_id", userID),
		zap.String("workspace_id", workspaceID))

	return conn, nil
}

// HandleSSE handles Server-Sent Events connections
// Server-Sent Events 연결 처리
func (s *service) HandleSSE(conn *SSEConnection) {
	defer func() {
		conn.Cancel()
		s.RemoveConnection(conn.ID)
	}()

	// 초기 연결 이벤트 전송
	s.BroadcastToConnection(conn.ID, "connected", map[string]interface{}{
		"connection_id": conn.ID,
		"message":       "Connected to real-time updates",
		"timestamp":     time.Now().Unix(),
	})

	// Heartbeat 고루틴 시작
	go s.startHeartbeat(conn)

	// 연결이 유지되는 동안 대기
	<-conn.Context.Done()

	s.logger.Info("SSE connection closed",
		zap.String("connection_id", conn.ID),
		zap.String("user_id", conn.UserID))
}

// RemoveConnection removes a connection
// 연결 제거
func (s *service) RemoveConnection(connID string) error {
	s.connectionsMux.Lock()
	defer s.connectionsMux.Unlock()

	conn, exists := s.connections[connID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	conn.Cancel()
	delete(s.connections, connID)

	s.logger.Info("SSE connection removed",
		zap.String("connection_id", connID))

	return nil
}

// GetConnection retrieves a connection by ID
// ID로 연결 조회
func (s *service) GetConnection(connID string) (*SSEConnection, error) {
	s.connectionsMux.RLock()
	defer s.connectionsMux.RUnlock()

	conn, exists := s.connections[connID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	return conn, nil
}

// GetConnectionCount returns the number of active connections
// 활성 연결 수 반환
func (s *service) GetConnectionCount() int {
	s.connectionsMux.RLock()
	defer s.connectionsMux.RUnlock()
	return len(s.connections)
}

// BroadcastToWorkspace broadcasts an event to all users in a workspace
// 워크스페이스의 모든 사용자에게 이벤트 브로드캐스팅
func (s *service) BroadcastToWorkspace(workspaceID string, event *messaging.Event) error {
	return s.eventBus.PublishToWorkspace(context.Background(), workspaceID, event)
}

// BroadcastToUser broadcasts an event to a specific user
// 특정 사용자에게 이벤트 브로드캐스팅
func (s *service) BroadcastToUser(userID string, event *messaging.Event) error {
	return s.eventBus.PublishToUser(context.Background(), userID, event)
}

// BroadcastToConnection broadcasts an event to a specific connection
// 특정 연결에 이벤트 브로드캐스팅
func (s *service) BroadcastToConnection(connID string, eventType string, data interface{}) error {
	conn, err := s.GetConnection(connID)
	if err != nil {
		return err
	}

	return s.sendToConnection(conn, eventType, data)
}

// BroadcastToAll broadcasts an event to all connections
// 모든 연결에 이벤트 브로드캐스팅
func (s *service) BroadcastToAll(eventType string, data interface{}) error {
	s.connectionsMux.RLock()
	connections := make([]*SSEConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}
	s.connectionsMux.RUnlock()

	var lastErr error
	for _, conn := range connections {
		select {
		case <-conn.Context.Done():
			continue
		default:
			if err := s.sendToConnection(conn, eventType, data); err != nil {
				lastErr = err
				s.logger.Warn("Failed to send event to connection",
					zap.String("connection_id", conn.ID),
					zap.String("event_type", eventType),
					zap.Error(err))
			}
		}
	}

	return lastErr
}

// sendToConnection sends an event to a connection
// 연결에 이벤트 전송
func (s *service) sendToConnection(conn *SSEConnection, eventType string, data interface{}) error {
	// 이벤트 구독 확인
	conn.mu.RLock()
	subscribed := conn.SubscribedEvents[eventType] || len(conn.SubscribedEvents) == 0
	conn.mu.RUnlock()

	if !subscribed {
		return nil // 구독하지 않은 이벤트는 무시
	}

	// JSON 마샬링
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// SSE 형식으로 전송
	fmt.Fprintf(conn.Writer, "event: %s\n", eventType)
	fmt.Fprintf(conn.Writer, "data: %s\n\n", string(jsonData))
	conn.Flusher.Flush()

	// LastSeen 업데이트
	conn.mu.Lock()
	conn.LastSeen = time.Now()
	conn.mu.Unlock()

	return nil
}

// SubscribeToEvent subscribes a connection to an event type
// 연결을 이벤트 타입에 구독
func (s *service) SubscribeToEvent(connID, eventType string) error {
	conn, err := s.GetConnection(connID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	conn.SubscribedEvents[eventType] = true
	conn.mu.Unlock()

	return nil
}

// UnsubscribeFromEvent unsubscribes a connection from an event type
// 연결의 이벤트 타입 구독 해제
func (s *service) UnsubscribeFromEvent(connID, eventType string) error {
	conn, err := s.GetConnection(connID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	delete(conn.SubscribedEvents, eventType)
	conn.mu.Unlock()

	return nil
}

// SubscribeToResource subscribes a connection to a resource
// 연결을 리소스에 구독
func (s *service) SubscribeToResource(connID, resourceType, resourceID string) error {
	conn, err := s.GetConnection(connID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	if conn.SubscribedResources[resourceType] == nil {
		conn.SubscribedResources[resourceType] = make(map[string]bool)
	}
	conn.SubscribedResources[resourceType][resourceID] = true
	conn.mu.Unlock()

	return nil
}

// UnsubscribeFromResource unsubscribes a connection from a resource
// 연결의 리소스 구독 해제
func (s *service) UnsubscribeFromResource(connID, resourceType, resourceID string) error {
	conn, err := s.GetConnection(connID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	if conn.SubscribedResources[resourceType] != nil {
		delete(conn.SubscribedResources[resourceType], resourceID)
	}
	conn.mu.Unlock()

	return nil
}

// CleanupInactiveConnections removes inactive connections
// 비활성 연결 제거
func (s *service) CleanupInactiveConnections() error {
	if s.clientTimeout == 0 {
		s.clientTimeout = 5 * time.Minute
	}

	s.connectionsMux.Lock()
	defer s.connectionsMux.Unlock()

	now := time.Now()
	for connID, conn := range s.connections {
		conn.mu.RLock()
		lastSeen := conn.LastSeen
		conn.mu.RUnlock()

		if now.Sub(lastSeen) > s.clientTimeout {
			conn.Cancel()
			delete(s.connections, connID)
			s.logger.Info("Cleaned up inactive SSE connection",
				zap.String("connection_id", connID))
		}
	}

	return nil
}

// StartCleanupRoutine starts a background routine to clean up inactive connections
// 비활성 연결 정리를 위한 백그라운드 루틴 시작
func (s *service) StartCleanupRoutine(ctx context.Context) {
	if s.cleanupInterval == 0 {
		s.cleanupInterval = 30 * time.Second
	}

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CleanupInactiveConnections(); err != nil {
				s.logger.Warn("Failed to cleanup inactive connections",
					zap.Error(err))
			}
		}
	}
}

// startHeartbeat sends periodic heartbeat messages to keep the connection alive
// 연결을 유지하기 위해 주기적으로 heartbeat 메시지 전송
func (s *service) startHeartbeat(conn *SSEConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.Context.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(conn.Writer, ": heartbeat\n\n")
			conn.Flusher.Flush()

			conn.mu.Lock()
			conn.LastSeen = time.Now()
			conn.mu.Unlock()
		}
	}
}

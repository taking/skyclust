package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type SSEHandler struct {
	logger     *zap.Logger
	natsConn   *nats.Conn
	clients    map[string]*SSEClient
	clientsMux sync.RWMutex
}

type SSEClient struct {
	ID       string
	UserID   string
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Context  context.Context
	Cancel   context.CancelFunc
	LastSeen time.Time
	// 구독 중인 이벤트 타입들
	SubscribedEvents map[string]bool
	// 구독 중인 VM/Provider ID들
	SubscribedVMs       map[string]bool
	SubscribedProviders map[string]bool
}

func NewSSEHandler(logger *zap.Logger, natsConn *nats.Conn) *SSEHandler {
	handler := &SSEHandler{
		logger:   logger,
		natsConn: natsConn,
		clients:  make(map[string]*SSEClient),
	}

	// NATS 구독 설정
	handler.setupNATSSubscriptions()

	// 클라이언트 정리 고루틴
	go handler.cleanupClients()

	return handler
}

func (h *SSEHandler) setupNATSSubscriptions() {
	// VM 상태 업데이트 구독
	_, _ = h.natsConn.Subscribe("vm.status.update", func(m *nats.Msg) {
		h.broadcastToClients("vm-status", m.Data)
	})

	// VM 리소스 업데이트 구독
	_, _ = h.natsConn.Subscribe("vm.resource.update", func(m *nats.Msg) {
		h.broadcastToClients("vm-resource", m.Data)
	})

	// Provider 상태 업데이트 구독
	_, _ = h.natsConn.Subscribe("provider.status.update", func(m *nats.Msg) {
		h.broadcastToClients("provider-status", m.Data)
	})

	// Provider 인스턴스 업데이트 구독
	_, _ = h.natsConn.Subscribe("provider.instance.update", func(m *nats.Msg) {
		h.broadcastToClients("provider-instance", m.Data)
	})

	// 시스템 알림 구독
	_, _ = h.natsConn.Subscribe("system.notification", func(m *nats.Msg) {
		h.broadcastToClients("system-notification", m.Data)
	})

	// 시스템 알림 구독
	_, _ = h.natsConn.Subscribe("system.alert", func(m *nats.Msg) {
		h.broadcastToClients("system-alert", m.Data)
	})
}

func (h *SSEHandler) HandleSSE(c *gin.Context) {
	// 사용자 ID 추출 (JWT에서)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// SSE 헤더 설정
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 클라이언트 ID 생성
	clientID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

	// Flusher 확인
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE not supported"})
		return
	}

	// 컨텍스트 생성
	ctx, cancel := context.WithCancel(c.Request.Context())

	// 클라이언트 등록
	client := &SSEClient{
		ID:                  clientID,
		UserID:              userID.(string),
		Writer:              c.Writer,
		Flusher:             flusher,
		Context:             ctx,
		Cancel:              cancel,
		LastSeen:            time.Now(),
		SubscribedEvents:    make(map[string]bool),
		SubscribedVMs:       make(map[string]bool),
		SubscribedProviders: make(map[string]bool),
	}

	h.clientsMux.Lock()
	h.clients[clientID] = client
	h.clientsMux.Unlock()

	h.logger.Info("SSE client connected", zap.String("client_id", clientID), zap.String("user_id", userID.(string)))

	// 연결 확인 메시지 전송
	h.sendToClient(client, "connected", map[string]interface{}{
		"client_id": clientID,
		"timestamp": time.Now().Unix(),
	})

	// 클라이언트가 연결을 유지하는 동안 대기
	<-ctx.Done()

	// 클라이언트 정리
	h.clientsMux.Lock()
	delete(h.clients, clientID)
	h.clientsMux.Unlock()

	h.logger.Info("SSE client disconnected", zap.String("client_id", clientID))
}

func (h *SSEHandler) sendToClient(client *SSEClient, eventType string, data interface{}) {
	message := SSEMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Event:     eventType,
		Data:      map[string]interface{}{"data": data},
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal SSE message", zap.Error(err))
		return
	}

	// SSE 형식으로 전송
	fmt.Fprintf(client.Writer, "event: %s\n", eventType)
	fmt.Fprintf(client.Writer, "data: %s\n\n", string(jsonData))
	client.Flusher.Flush()

	client.LastSeen = time.Now()
}

func (h *SSEHandler) broadcastToClients(eventType string, data []byte) {
	var messageData interface{}
	if err := json.Unmarshal(data, &messageData); err != nil {
		h.logger.Error("Failed to unmarshal NATS message", zap.Error(err))
		return
	}

	h.clientsMux.RLock()
	clients := make([]*SSEClient, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.clientsMux.RUnlock()

	for _, client := range clients {
		select {
		case <-client.Context.Done():
			// 클라이언트가 연결 해제됨
			continue
		default:
			// 이벤트 타입 구독 확인
			if !client.SubscribedEvents[eventType] {
				continue
			}

			// VM/Provider 특정 구독 확인
			if !h.shouldSendToClient(client, eventType, messageData) {
				continue
			}

			h.sendToClient(client, eventType, messageData)
		}
	}
}

// 클라이언트에게 메시지를 보낼지 결정하는 스마트 필터링
func (h *SSEHandler) shouldSendToClient(client *SSEClient, eventType string, data interface{}) bool {
	// 시스템 이벤트는 항상 전송
	if eventType == "system-notification" || eventType == "system-alert" {
		return true
	}

	// VM 관련 이벤트
	if eventType == "vm-status" || eventType == "vm-resource" || eventType == "vm-error" {
		if vmData, ok := data.(map[string]interface{}); ok {
			if vmID, exists := vmData["vmId"]; exists {
				return client.SubscribedVMs[vmID.(string)]
			}
		}
		return false
	}

	// Provider 관련 이벤트
	if eventType == "provider-status" || eventType == "provider-instance" {
		if providerData, ok := data.(map[string]interface{}); ok {
			if provider, exists := providerData["provider"]; exists {
				return client.SubscribedProviders[provider.(string)]
			}
		}
		return false
	}

	// 기본적으로 전송
	return true
}

func (h *SSEHandler) cleanupClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.clientsMux.Lock()
		now := time.Now()
		for clientID, client := range h.clients {
			if now.Sub(client.LastSeen) > 5*time.Minute {
				client.Cancel()
				delete(h.clients, clientID)
				h.logger.Info("Cleaned up inactive SSE client", zap.String("client_id", clientID))
			}
		}
		h.clientsMux.Unlock()
	}
}

func (h *SSEHandler) GetClientCount() int {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	return len(h.clients)
}

// 구독 관리 메서드들
func (h *SSEHandler) SubscribeToEvent(clientID, eventType string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	client.SubscribedEvents[eventType] = true
	return nil
}

func (h *SSEHandler) UnsubscribeFromEvent(clientID, eventType string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	delete(client.SubscribedEvents, eventType)
	return nil
}

func (h *SSEHandler) SubscribeToVM(clientID, vmID string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	client.SubscribedVMs[vmID] = true
	return nil
}

func (h *SSEHandler) UnsubscribeFromVM(clientID, vmID string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	delete(client.SubscribedVMs, vmID)
	return nil
}

func (h *SSEHandler) SubscribeToProvider(clientID, provider string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	client.SubscribedProviders[provider] = true
	return nil
}

func (h *SSEHandler) UnsubscribeFromProvider(clientID, provider string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	delete(client.SubscribedProviders, provider)
	return nil
}

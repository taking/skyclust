package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/internal/shared/handlers"
	"skyclust/pkg/realtime"
)

// 오류 정의
var (
	ErrClientNotFound = errors.New("client not found")
)

// SSEHandler: Server-Sent Events 핸들러
type SSEHandler struct {
	*handlers.BaseHandler
	logger      *zap.Logger
	natsConn    *nats.Conn
	realtimeSvc realtime.Service
	clients     map[string]*SSEClient
	clientsMux  sync.RWMutex
	batchBuffer *BatchBuffer
}

// SSEClient: SSE 클라이언트를 나타내는 구조체
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
	// 구독 필터 (provider, credential_id, region 기반)
	Filters SSEClientFilters
}

// SSEClientFilters: SSE 클라이언트 구독을 위한 필터를 정의하는 구조체
type SSEClientFilters struct {
	Providers     map[string]bool // 구독 중인 provider 목록
	CredentialIDs map[string]bool // 구독 중인 credential_id 목록
	Regions       map[string]bool // 구독 중인 region 목록
	ResourceTypes map[string]bool // 구독 중인 resource type 목록
}

// NewSSEHandler: 새로운 SSE 핸들러를 생성합니다
func NewSSEHandler(logger *zap.Logger, natsConn *nats.Conn, eventBus interface{}) *SSEHandler {
	// realtime.Service 생성 (eventBus를 messaging.Bus로 변환)
	var messagingBus messaging.Bus
	if bus, ok := eventBus.(messaging.Bus); ok {
		messagingBus = bus
	} else {
		// Fallback: LocalBus 생성
		messagingBus = messaging.NewLocalBus()
	}

	realtimeSvc := realtime.NewService(messagingBus, logger)

	handler := &SSEHandler{
		BaseHandler: handlers.NewBaseHandler("sse"),
		logger:      logger,
		natsConn:    natsConn,
		realtimeSvc: realtimeSvc,
		clients:     make(map[string]*SSEClient),
	}

	// 배치 버퍼 초기화
	handler.batchBuffer = NewBatchBuffer(BatchMaxSize, BatchFlushInterval, handler.flushBatch)

	// NATS 구독 설정
	handler.setupNATSSubscriptions()

	// realtime.Service의 cleanup 루틴 시작
	ctx := context.Background()
	go realtimeSvc.StartCleanupRoutine(ctx)

	return handler
}

// flushBatch: 배치된 이벤트를 모든 구독 클라이언트에게 전송합니다
func (h *SSEHandler) flushBatch(events []BatchEvent) {
	if len(events) == 0 {
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
			continue
		default:
			// Send batch of events
			for _, event := range events {
				if h.shouldSendToClient(client, event.EventType, event.Data) {
					h.sendToClient(client, event.EventType, event.Data)
				}
			}
		}
	}
}

func (h *SSEHandler) setupNATSSubscriptions() {
	// VM 상태 업데이트 구독
	_, _ = h.natsConn.Subscribe("vm.status.update", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeVMStatus, m.Data)
	})

	// VM 리소스 업데이트 구독
	_, _ = h.natsConn.Subscribe("vm.resource.update", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeVMResource, m.Data)
	})

	// Provider 상태 업데이트 구독
	_, _ = h.natsConn.Subscribe("provider.status.update", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeProviderStatus, m.Data)
	})

	// Provider 인스턴스 업데이트 구독
	_, _ = h.natsConn.Subscribe("provider.instance.update", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeProviderInstance, m.Data)
	})

	// 시스템 알림 구독
	_, _ = h.natsConn.Subscribe("system.notification", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeSystemNotification, m.Data)
	})

	// 시스템 알림 구독
	_, _ = h.natsConn.Subscribe("system.alert", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeSystemAlert, m.Data)
	})

	// Kubernetes 클러스터 이벤트 구독 (와일드카드 패턴)
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.*.clusters.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesClusterCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.*.clusters.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesClusterUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.*.clusters.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesClusterDeleted, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.*.clusters.list", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesClusterList, m.Data)
	})

	// Kubernetes Node Pool 이벤트 구독
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodepools.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodePoolCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodepools.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodePoolUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodepools.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodePoolDeleted, m.Data)
	})

	// Kubernetes Node 이벤트 구독
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodes.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodeCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodes.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodeUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("kubernetes.*.*.clusters.*.nodes.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeKubernetesNodeDeleted, m.Data)
	})

	// Network VPC 이벤트 구독
	_, _ = h.natsConn.Subscribe("network.*.*.*.vpcs.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkVPCCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.vpcs.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkVPCUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.vpcs.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkVPCDeleted, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.vpcs.list", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkVPCList, m.Data)
	})

	// Network Subnet 이벤트 구독
	_, _ = h.natsConn.Subscribe("network.*.*.vpcs.*.subnets.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSubnetCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.vpcs.*.subnets.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSubnetUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.vpcs.*.subnets.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSubnetDeleted, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.vpcs.*.subnets.list", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSubnetList, m.Data)
	})

	// Network Security Group 이벤트 구독
	_, _ = h.natsConn.Subscribe("network.*.*.*.security-groups.created", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSecurityGroupCreated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.security-groups.updated", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSecurityGroupUpdated, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.security-groups.deleted", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSecurityGroupDeleted, m.Data)
	})
	_, _ = h.natsConn.Subscribe("network.*.*.*.security-groups.list", func(m *nats.Msg) {
		h.broadcastToClients(EventTypeNetworkSecurityGroupList, m.Data)
	})
}

// HandleSSE: Server-Sent Events 연결을 처리합니다
func (h *SSEHandler) HandleSSE(c *gin.Context) {
	handler := h.Compose(
		h.handleSSECore(),
		h.StandardCRUDDecorators("handle_sse")...,
	)

	handler(c)
}

// handleSSECore: SSE 연결의 핵심 비즈니스 로직을 처리합니다
func (h *SSEHandler) handleSSECore() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// 사용자 ID 추출 (JWT에서)
		userID, exists := c.Get("user_id")
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "handle_sse")
			return
		}

		// Workspace ID 추출 (선택적)
		workspaceID := ""
		if wsID, exists := c.Get("workspace_id"); exists {
			workspaceID = wsID.(string)
		}

		// SSE 헤더 설정
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Cache-Control")

		// realtime.Service를 사용하여 SSE 연결 생성
		conn, err := h.realtimeSvc.CreateSSEConnection(c.Writer, c.Request, userID.(string), workspaceID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to create SSE connection: %v", err), 500), "handle_sse")
			return
		}

		// 클라이언트 등록 (필터링 및 구독 관리를 위해)
		client := &SSEClient{
			ID:                  conn.ID,
			UserID:              conn.UserID,
			Writer:              conn.Writer,
			Flusher:             conn.Flusher,
			Context:             conn.Context,
			Cancel:              conn.Cancel,
			LastSeen:            conn.LastSeen,
			SubscribedEvents:    conn.SubscribedEvents,
			SubscribedVMs:       make(map[string]bool),
			SubscribedProviders: make(map[string]bool),
			Filters: SSEClientFilters{
				Providers:     make(map[string]bool),
				CredentialIDs: make(map[string]bool),
				Regions:       make(map[string]bool),
				ResourceTypes: make(map[string]bool),
			},
		}

		h.clientsMux.Lock()
		h.clients[conn.ID] = client
		h.clientsMux.Unlock()

		h.LogInfo(c, "SSE client connected", zap.String("client_id", conn.ID), zap.String("user_id", userID.(string)))

		// realtime.Service를 사용하여 SSE 연결 처리
		// 이는 연결 확인 메시지 전송 및 heartbeat 관리를 포함
		h.realtimeSvc.HandleSSE(conn)

		// 클라이언트 정리
		h.clientsMux.Lock()
		delete(h.clients, conn.ID)
		h.clientsMux.Unlock()

		h.LogInfo(c, "SSE client disconnected", zap.String("client_id", conn.ID))
	}
}

// sendToClient: 특정 클라이언트에게 SSE 이벤트를 전송합니다
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

	// 메시지 압축 시도
	compressedData, isCompressed, err := CompressMessage(jsonData)
	if err != nil {
		h.logger.Warn("Failed to compress SSE message, sending uncompressed",
			zap.Error(err),
			zap.String("event_type", eventType))
		compressedData = jsonData
		isCompressed = false
	}

	// SSE 형식으로 전송
	fmt.Fprintf(client.Writer, "event: %s\n", eventType)
	if isCompressed {
		// 압축된 메시지는 압축 플래그와 함께 전송
		fmt.Fprintf(client.Writer, "compressed: true\n")
		fmt.Fprintf(client.Writer, "data: %s\n\n", string(compressedData))
	} else {
		fmt.Fprintf(client.Writer, "data: %s\n\n", string(jsonData))
	}
	client.Flusher.Flush()

	client.LastSeen = time.Now()
}

// broadcastToClients: 모든 구독 클라이언트에게 이벤트를 브로드캐스트합니다
func (h *SSEHandler) broadcastToClients(eventType string, data []byte) {
	var messageData interface{}
	if err := json.Unmarshal(data, &messageData); err != nil {
		h.logger.Error("Failed to unmarshal NATS message", zap.Error(err))
		return
	}

	// realtime.Service를 사용하여 모든 연결에 브로드캐스팅
	// 필터링은 각 클라이언트별로 처리
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
			if len(client.SubscribedEvents) > 0 && !client.SubscribedEvents[eventType] {
				continue
			}

			// VM/Provider 특정 구독 확인
			if !h.shouldSendToClient(client, eventType, messageData) {
				continue
			}

			// realtime.Service를 통해 이벤트 전송
			if err := h.realtimeSvc.BroadcastToConnection(client.ID, eventType, messageData); err != nil {
				h.logger.Warn("Failed to broadcast to connection",
					zap.String("connection_id", client.ID),
					zap.String("event_type", eventType),
					zap.Error(err))
			}
		}
	}
}

// shouldSendToClient: 클라이언트에게 메시지를 보낼지 결정하는 스마트 필터링을 수행합니다
func (h *SSEHandler) shouldSendToClient(client *SSEClient, eventType string, data interface{}) bool {
	// 시스템 이벤트는 항상 전송
	if h.isSystemEvent(eventType) {
		return true
	}

	// 이벤트 타입 구독 확인
	if !client.SubscribedEvents[eventType] && len(client.SubscribedEvents) > 0 {
		return false
	}

	// VM 이벤트 필터링
	if h.isVMEvent(eventType) {
		return h.shouldSendVMEvent(client, data)
	}

	// Provider 이벤트 필터링
	if h.isProviderEvent(eventType) {
		return h.shouldSendProviderEvent(client, data)
	}

	// Kubernetes/Network 이벤트 필터링
	if h.isKubernetesEvent(eventType) || h.isNetworkEvent(eventType) {
		return h.shouldSendResourceEvent(client, eventType, data)
	}

	return true
}

// isKubernetesEvent: 이벤트 타입이 Kubernetes 관련 이벤트인지 확인합니다
func (h *SSEHandler) isKubernetesEvent(eventType string) bool {
	return eventType == EventTypeKubernetesClusterCreated ||
		eventType == EventTypeKubernetesClusterUpdated ||
		eventType == EventTypeKubernetesClusterDeleted ||
		eventType == EventTypeKubernetesClusterList ||
		eventType == EventTypeKubernetesNodePoolCreated ||
		eventType == EventTypeKubernetesNodePoolUpdated ||
		eventType == EventTypeKubernetesNodePoolDeleted ||
		eventType == EventTypeKubernetesNodeCreated ||
		eventType == EventTypeKubernetesNodeUpdated ||
		eventType == EventTypeKubernetesNodeDeleted
}

// isNetworkEvent: 이벤트 타입이 Network 관련 이벤트인지 확인합니다
func (h *SSEHandler) isNetworkEvent(eventType string) bool {
	return eventType == EventTypeNetworkVPCCreated ||
		eventType == EventTypeNetworkVPCUpdated ||
		eventType == EventTypeNetworkVPCDeleted ||
		eventType == EventTypeNetworkVPCList ||
		eventType == EventTypeNetworkSubnetCreated ||
		eventType == EventTypeNetworkSubnetUpdated ||
		eventType == EventTypeNetworkSubnetDeleted ||
		eventType == EventTypeNetworkSubnetList ||
		eventType == EventTypeNetworkSecurityGroupCreated ||
		eventType == EventTypeNetworkSecurityGroupUpdated ||
		eventType == EventTypeNetworkSecurityGroupDeleted ||
		eventType == EventTypeNetworkSecurityGroupList
}

// shouldSendResourceEvent: 필터를 기반으로 리소스 이벤트를 클라이언트에게 전송할지 결정합니다
func (h *SSEHandler) shouldSendResourceEvent(client *SSEClient, eventType string, data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return true
	}

	// Provider 필터 확인
	if len(client.Filters.Providers) > 0 {
		provider, exists := dataMap[FieldProvider]
		if !exists {
			return false
		}
		providerStr, ok := provider.(string)
		if !ok || !client.Filters.Providers[providerStr] {
			return false
		}
	}

	// Credential ID 필터 확인
	if len(client.Filters.CredentialIDs) > 0 {
		credentialID, exists := dataMap[FieldCredentialID]
		if !exists {
			return false
		}
		credentialIDStr, ok := credentialID.(string)
		if !ok || !client.Filters.CredentialIDs[credentialIDStr] {
			return false
		}
	}

	// Region 필터 확인
	if len(client.Filters.Regions) > 0 {
		region, exists := dataMap[FieldRegion]
		if !exists {
			return false
		}
		regionStr, ok := region.(string)
		if !ok || !client.Filters.Regions[regionStr] {
			return false
		}
	}

	return true
}

// isSystemEvent: 이벤트 타입이 시스템 이벤트인지 확인합니다
func (h *SSEHandler) isSystemEvent(eventType string) bool {
	return eventType == EventTypeSystemNotification || eventType == EventTypeSystemAlert
}

// isVMEvent: 이벤트 타입이 VM 관련 이벤트인지 확인합니다
func (h *SSEHandler) isVMEvent(eventType string) bool {
	return eventType == EventTypeVMStatus || eventType == EventTypeVMResource || eventType == EventTypeVMError
}

// isProviderEvent: 이벤트 타입이 Provider 관련 이벤트인지 확인합니다
func (h *SSEHandler) isProviderEvent(eventType string) bool {
	return eventType == EventTypeProviderStatus || eventType == EventTypeProviderInstance
}

// shouldSendVMEvent: VM 이벤트를 클라이언트에게 전송할지 결정합니다
func (h *SSEHandler) shouldSendVMEvent(client *SSEClient, data interface{}) bool {
	vmData, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	vmID, exists := vmData[FieldVMID]
	if !exists {
		return false
	}

	return client.SubscribedVMs[vmID.(string)]
}

// shouldSendProviderEvent: Provider 이벤트를 클라이언트에게 전송할지 결정합니다
func (h *SSEHandler) shouldSendProviderEvent(client *SSEClient, data interface{}) bool {
	providerData, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	provider, exists := providerData[FieldProvider]
	if !exists {
		return false
	}

	return client.SubscribedProviders[provider.(string)]
}

// cleanupClients는 realtime.Service의 CleanupInactiveConnections로 대체됨
// cleanupClients is replaced by realtime.Service's CleanupInactiveConnections

// GetClientCount: 현재 연결된 SSE 클라이언트 수를 반환합니다
func (h *SSEHandler) GetClientCount() int {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	return len(h.clients)
}

// SubscribeToEvent: 클라이언트를 특정 이벤트 타입에 구독시킵니다
func (h *SSEHandler) SubscribeToEvent(clientID, eventType string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	client.SubscribedEvents[eventType] = true
	return nil
}

// UnsubscribeFromEvent: 클라이언트의 특정 이벤트 타입 구독을 해제합니다
func (h *SSEHandler) UnsubscribeFromEvent(clientID, eventType string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	delete(client.SubscribedEvents, eventType)
	return nil
}

// SubscribeToVM: 클라이언트를 특정 VM에 구독시킵니다
func (h *SSEHandler) SubscribeToVM(clientID, vmID string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	client.SubscribedVMs[vmID] = true
	return nil
}

// UnsubscribeFromVM: 클라이언트의 특정 VM 구독을 해제합니다
func (h *SSEHandler) UnsubscribeFromVM(clientID, vmID string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	delete(client.SubscribedVMs, vmID)
	return nil
}

// SubscribeToProvider: 클라이언트를 특정 Provider에 구독시킵니다
func (h *SSEHandler) SubscribeToProvider(clientID, provider string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	client.SubscribedProviders[provider] = true
	return nil
}

// UnsubscribeFromProvider: 클라이언트의 특정 Provider 구독을 해제합니다
func (h *SSEHandler) UnsubscribeFromProvider(clientID, provider string) error {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	delete(client.SubscribedProviders, provider)
	return nil
}

// startHeartbeat는 realtime.Service의 HandleSSE 내부에서 처리됨
// startHeartbeat is handled internally by realtime.Service's HandleSSE

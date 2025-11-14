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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/internal/shared/handlers"
	"skyclust/pkg/cache"
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
	redisClient *redis.Client
	cache       cache.Cache
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
func NewSSEHandler(logger *zap.Logger, natsConn *nats.Conn, eventBus interface{}, cacheService cache.Cache) *SSEHandler {
	// realtime.Service 생성 (eventBus를 messaging.Bus로 변환)
	var messagingBus messaging.Bus
	if bus, ok := eventBus.(messaging.Bus); ok {
		messagingBus = bus
	} else {
		// Fallback: LocalBus 생성
		messagingBus = messaging.NewLocalBus()
	}

	realtimeSvc := realtime.NewService(messagingBus, logger)

	// Redis client 추출 (RedisService인 경우)
	var redisClient *redis.Client
	if redisService, ok := cacheService.(*cache.RedisService); ok {
		redisClient = redisService.GetClient()
		logger.Info("SSE handler initialized with Redis support")
	} else {
		logger.Warn("SSE handler initialized without Redis (event history will not be stored)")
	}

	handler := &SSEHandler{
		BaseHandler: handlers.NewBaseHandler("sse"),
		logger:      logger,
		natsConn:    natsConn,
		realtimeSvc: realtimeSvc,
		clients:     make(map[string]*SSEClient),
		redisClient: redisClient,
		cache:       cacheService,
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
	if h.natsConn == nil {
		h.logger.Warn("NATS connection not available, skipping subscription setup")
		return
	}

	// NATS 메시지 처리 헬퍼 함수 (압축 해제 포함)
	handleNATSMessage := func(eventType string) func(*nats.Msg) {
		return func(m *nats.Msg) {
			// NATSService가 압축된 메시지를 발행할 수 있으므로 압축 해제 시도
			data := m.Data

			// 압축 여부 확인 (gzip magic number: 0x1f 0x8b)
			if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
				// gzip 압축 해제
				decompressed, err := messaging.Decompress(data, messaging.CompressionGzip)
				if err != nil {
					h.logger.Warn("Failed to decompress gzip message, using raw data",
						zap.Error(err),
						zap.String("event_type", eventType))
					// 압축 해제 실패 시 원본 데이터 사용
					h.broadcastToClients(eventType, data)
					return
				}
				data = decompressed
			} else if len(data) >= 4 && data[0] == 's' && data[1] == 'N' && data[2] == 'a' && data[3] == 'P' {
				// snappy 압축 해제
				decompressed, err := messaging.Decompress(data, messaging.CompressionSnappy)
				if err != nil {
					h.logger.Warn("Failed to decompress snappy message, using raw data",
						zap.Error(err),
						zap.String("event_type", eventType))
					// 압축 해제 실패 시 원본 데이터 사용
					h.broadcastToClients(eventType, data)
					return
				}
				data = decompressed
			}

			h.broadcastToClients(eventType, data)
		}
	}

	// NATSService는 "cmp.events.{event.Type}" 형식으로 발행하므로 구독 패턴도 동일하게 맞춤
	// VM 상태 업데이트 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("vm.status.update", handleNATSMessage(EventTypeVMStatus))

	// VM 리소스 업데이트 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("vm.resource.update", handleNATSMessage(EventTypeVMResource))

	// Provider 상태 업데이트 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("provider.status.update", handleNATSMessage(EventTypeProviderStatus))

	// Provider 인스턴스 업데이트 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("provider.instance.update", handleNATSMessage(EventTypeProviderInstance))

	// 시스템 알림 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("system.notification", handleNATSMessage(EventTypeSystemNotification))

	// 시스템 알림 구독 (레거시 형식 지원)
	_, _ = h.natsConn.Subscribe("system.alert", handleNATSMessage(EventTypeSystemAlert))

	// Kubernetes 클러스터 이벤트 구독 (NATSService 형식: cmp.events.{topic})
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.created", handleNATSMessage(EventTypeKubernetesClusterCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.updated", handleNATSMessage(EventTypeKubernetesClusterUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.deleted", handleNATSMessage(EventTypeKubernetesClusterDeleted))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.list", handleNATSMessage(EventTypeKubernetesClusterList))

	// Kubernetes Node Pool 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodepools.created", handleNATSMessage(EventTypeKubernetesNodePoolCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodepools.updated", handleNATSMessage(EventTypeKubernetesNodePoolUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodepools.deleted", handleNATSMessage(EventTypeKubernetesNodePoolDeleted))

	// Kubernetes Node 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodes.created", handleNATSMessage(EventTypeKubernetesNodeCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodes.updated", handleNATSMessage(EventTypeKubernetesNodeUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.kubernetes.*.*.clusters.*.nodes.deleted", handleNATSMessage(EventTypeKubernetesNodeDeleted))

	// Network VPC 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.created", handleNATSMessage(EventTypeNetworkVPCCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.updated", handleNATSMessage(EventTypeNetworkVPCUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.deleted", handleNATSMessage(EventTypeNetworkVPCDeleted))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.list", handleNATSMessage(EventTypeNetworkVPCList))

	// Network Subnet 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.created", handleNATSMessage(EventTypeNetworkSubnetCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.updated", handleNATSMessage(EventTypeNetworkSubnetUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.deleted", handleNATSMessage(EventTypeNetworkSubnetDeleted))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.list", handleNATSMessage(EventTypeNetworkSubnetList))

	// Network Security Group 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.created", handleNATSMessage(EventTypeNetworkSecurityGroupCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.updated", handleNATSMessage(EventTypeNetworkSecurityGroupUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.deleted", handleNATSMessage(EventTypeNetworkSecurityGroupDeleted))
	_, _ = h.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.list", handleNATSMessage(EventTypeNetworkSecurityGroupList))

	// Azure Resource Group 이벤트 구독
	_, _ = h.natsConn.Subscribe("cmp.events.azure.*.*.resource-groups.created", handleNATSMessage(EventTypeAzureResourceGroupCreated))
	_, _ = h.natsConn.Subscribe("cmp.events.azure.*.*.resource-groups.updated", handleNATSMessage(EventTypeAzureResourceGroupUpdated))
	_, _ = h.natsConn.Subscribe("cmp.events.azure.*.*.resource-groups.deleted", handleNATSMessage(EventTypeAzureResourceGroupDeleted))
	_, _ = h.natsConn.Subscribe("cmp.events.azure.*.*.resource-groups.list", handleNATSMessage(EventTypeAzureResourceGroupList))

	// Dashboard Summary 이벤트 구독
	// dashboard-summary-updated 이벤트는 dashboard service에서 직접 발행되므로
	// messaging bus를 통해 전달됩니다 (NATS topic: dashboard.summary.updated)
	_, _ = h.natsConn.Subscribe("dashboard.summary.updated", handleNATSMessage(EventTypeDashboardSummaryUpdated))

	h.logger.Info("NATS subscriptions configured for SSE")
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
		userIDValue, exists := c.Get("user_id")
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "handle_sse")
			return
		}

		// user_id를 string으로 안전하게 변환
		var userID string
		switch v := userIDValue.(type) {
		case string:
			userID = v
		case fmt.Stringer:
			userID = v.String()
		default:
			userID = fmt.Sprintf("%v", v)
		}

		if userID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID", 401), "handle_sse")
			return
		}

		// Workspace ID 추출 (선택적)
		workspaceID := ""
		if wsIDValue, exists := c.Get("workspace_id"); exists {
			switch v := wsIDValue.(type) {
			case string:
				workspaceID = v
			case fmt.Stringer:
				workspaceID = v.String()
			default:
				workspaceID = fmt.Sprintf("%v", v)
			}
		}

		// Last-Event-ID 추출 (헤더 또는 쿼리 파라미터)
		lastEventID := c.GetHeader("Last-Event-ID")
		if lastEventID == "" {
			lastEventID = c.Query("last_event_id")
		}

		// SSE 헤더 설정
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Cache-Control, Last-Event-ID")

		// Retry 간격 설정 (3초)
		// SSE 표준에 따라 retry 이벤트로 전송
		h.sendRetryInterval(c.Writer, 3*time.Second)

		// realtime.Service를 사용하여 SSE 연결 생성
		conn, err := h.realtimeSvc.CreateSSEConnection(c.Writer, c.Request, userID, workspaceID)
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

		// Save connection state to Redis
		ctx := c.Request.Context()
		h.saveConnectionState(ctx, client)

		h.LogInfo(c, "SSE client connected", zap.String("client_id", conn.ID), zap.String("user_id", userID), zap.String("last_event_id", lastEventID))

		// Last-Event-ID가 있으면 누락된 이벤트 전송
		if lastEventID != "" {
			ctx := c.Request.Context()
			// 모든 이벤트 타입에 대해 누락된 이벤트 조회
			missedEvents, err := h.getMissedEvents(ctx, userID, lastEventID)
			if err == nil && len(missedEvents) > 0 {
				h.logger.Info("Sending missed events",
					zap.String("user_id", userID),
					zap.String("last_event_id", lastEventID),
					zap.Int("missed_count", len(missedEvents)))

				// 누락된 이벤트 전송 (타임스탬프 순으로 정렬)
				for _, event := range missedEvents {
					var eventData interface{}
					if err := json.Unmarshal(event.Data, &eventData); err == nil {
						// 이벤트 ID를 포함하여 전송
						h.sendToClientWithID(client, event.EventType, event.EventID, eventData)
					}
				}
			}
		}

		// realtime.Service를 사용하여 SSE 연결 처리
		// 이는 연결 확인 메시지 전송 및 heartbeat 관리를 포함
		h.realtimeSvc.HandleSSE(conn)

		// 클라이언트 정리
		h.clientsMux.Lock()
		clientToDelete, exists := h.clients[conn.ID]
		delete(h.clients, conn.ID)
		h.clientsMux.Unlock()

		// Delete connection state from Redis
		if exists && clientToDelete != nil {
			ctx := c.Request.Context()
			h.deleteConnectionState(ctx, clientToDelete)
		}

		h.LogInfo(c, "SSE client disconnected", zap.String("client_id", conn.ID))
	}
}

// sendToClient: 특정 클라이언트에게 SSE 이벤트를 전송합니다
func (h *SSEHandler) sendToClient(client *SSEClient, eventType string, data interface{}) {
	eventID := fmt.Sprintf("%d", time.Now().UnixNano())
	h.sendToClientWithID(client, eventType, eventID, data)
}

// sendToClientWithID: 특정 클라이언트에게 SSE 이벤트를 ID와 함께 전송합니다
func (h *SSEHandler) sendToClientWithID(client *SSEClient, eventType string, eventID string, data interface{}) {
	message := SSEMessage{
		ID:        eventID,
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

	// SSE 형식으로 전송 (id 필드 포함)
	fmt.Fprintf(client.Writer, "id: %s\n", eventID)
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

	// 이벤트 히스토리에 저장 (비동기)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.saveEventToHistory(ctx, client.UserID, eventType, eventID, data)
	}()

	// 연결 상태 업데이트 (LastSeen 갱신, 비동기)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		h.updateConnectionLastSeen(ctx, client)
	}()
}

// sendRetryInterval: SSE retry 간격을 전송합니다
func (h *SSEHandler) sendRetryInterval(w http.ResponseWriter, interval time.Duration) {
	// SSE 표준에 따라 retry 필드로 재연결 간격 전송
	fmt.Fprintf(w, "retry: %d\n\n", interval.Milliseconds())
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// broadcastToClients: 모든 구독 클라이언트에게 이벤트를 브로드캐스트합니다
func (h *SSEHandler) broadcastToClients(eventType string, data []byte) {
	// NATSService는 압축된 Event 형식으로 발행하므로 먼저 압축 해제 시도
	// 압축 해제는 NATSService에서 이미 처리되었을 수 있지만, 안전을 위해 확인
	var event messaging.Event
	if err := json.Unmarshal(data, &event); err != nil {
		// Event 형식이 아닌 경우 원본 데이터를 그대로 사용
		var messageData interface{}
		if err2 := json.Unmarshal(data, &messageData); err2 != nil {
			h.logger.Error("Failed to unmarshal NATS message",
				zap.Error(err),
				zap.Error(err2),
				zap.String("event_type", eventType))
			return
		}
		h.broadcastToClientsWithData(eventType, messageData)
		return
	}

	// Event 형식인 경우, Data 필드를 추출하여 전송
	// Event.Data는 map[string]interface{} 형식
	h.broadcastToClientsWithData(eventType, event.Data)
}

// broadcastToClientsWithData: 실제 데이터를 클라이언트에게 브로드캐스트합니다
func (h *SSEHandler) broadcastToClientsWithData(eventType string, messageData interface{}) {
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
	// Backend는 "credential_id"를 사용하지만, 필터링을 위해 "credentialId"도 확인
	if len(client.Filters.CredentialIDs) > 0 {
		var credentialID interface{}
		var exists bool
		// 먼저 "credential_id" 확인 (Backend 표준)
		credentialID, exists = dataMap["credential_id"]
		if !exists {
			// Fallback: "credentialId" 확인 (Frontend 표준)
			credentialID, exists = dataMap[FieldCredentialID]
		}
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

// HandleSubscribeToEvent: HTTP 핸들러로 이벤트 구독을 처리합니다
func (h *SSEHandler) HandleSubscribeToEvent(c *gin.Context) {
	handler := h.Compose(
		h.subscribeToEventHandler(),
		h.StandardCRUDDecorators("subscribe_to_event")...,
	)

	handler(c)
}

// subscribeToEventHandler: 이벤트 구독 핵심 로직
func (h *SSEHandler) subscribeToEventHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// 사용자 ID 추출
		userIDValue, exists := c.Get("user_id")
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "subscribe_to_event")
			return
		}

		// user_id를 string으로 안전하게 변환
		var userIDStr string
		switch v := userIDValue.(type) {
		case string:
			userIDStr = v
		case fmt.Stringer:
			userIDStr = v.String()
		default:
			userIDStr = fmt.Sprintf("%v", v)
		}

		// 요청 바디 파싱 및 검증
		var req struct {
			EventType string                 `json:"event_type" binding:"required"`
			Filters   map[string]interface{} `json:"filters,omitempty"`
		}

		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "subscribe_to_event")
			return
		}

		// 클라이언트 ID 찾기 (현재 사용자의 활성 연결)
		// 먼저 h.clients 맵에서 찾고, 없으면 realtime.Service의 connections 맵에서 찾기
		clientID := ""
		h.clientsMux.RLock()
		for id, client := range h.clients {
			if client.UserID == userIDStr {
				clientID = id
				break
			}
		}
		h.clientsMux.RUnlock()

		// h.clients에서 찾지 못했으면 realtime.Service의 connections에서 찾기
		if clientID == "" {
			conn, err := h.realtimeSvc.GetConnectionByUserID(userIDStr)
			if err == nil && conn != nil {
				clientID = conn.ID
				// realtime.Service의 connections에 있지만 h.clients에 없으면 등록
				h.clientsMux.Lock()
				if _, exists := h.clients[conn.ID]; !exists {
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
					h.clients[conn.ID] = client
				}
				h.clientsMux.Unlock()
			}
		}

		if clientID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "No active SSE connection found", 404), "subscribe_to_event")
			return
		}

		// 이벤트 구독
		if err := h.SubscribeToEvent(clientID, req.EventType); err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to subscribe: %v", err), 500), "subscribe_to_event")
			return
		}

		// 필터 설정 (있는 경우)
		if req.Filters != nil {
			h.clientsMux.Lock()
			client, exists := h.clients[clientID]
			if exists {
				if providers, ok := req.Filters["providers"].([]interface{}); ok {
					for _, p := range providers {
						if provider, ok := p.(string); ok {
							client.Filters.Providers[provider] = true
						}
					}
				}
				if credentialIDs, ok := req.Filters["credential_ids"].([]interface{}); ok {
					for _, cid := range credentialIDs {
						if credentialID, ok := cid.(string); ok {
							client.Filters.CredentialIDs[credentialID] = true
						}
					}
				}
				if regions, ok := req.Filters["regions"].([]interface{}); ok {
					for _, r := range regions {
						if region, ok := r.(string); ok {
							client.Filters.Regions[region] = true
						}
					}
				}
			}
			h.clientsMux.Unlock()
		}

		h.OK(c, gin.H{
			"client_id":  clientID,
			"event_type": req.EventType,
			"subscribed": true,
		}, "Successfully subscribed to event")
	}
}

// HandleUnsubscribeFromEvent: HTTP 핸들러로 이벤트 구독 해제를 처리합니다
func (h *SSEHandler) HandleUnsubscribeFromEvent(c *gin.Context) {
	handler := h.Compose(
		h.unsubscribeFromEventHandler(),
		h.StandardCRUDDecorators("unsubscribe_from_event")...,
	)

	handler(c)
}

// unsubscribeFromEventHandler: 이벤트 구독 해제 핵심 로직
func (h *SSEHandler) unsubscribeFromEventHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// 사용자 ID 추출
		userIDValue, exists := c.Get("user_id")
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "unsubscribe_from_event")
			return
		}

		// user_id를 string으로 안전하게 변환
		var userIDStr string
		switch v := userIDValue.(type) {
		case string:
			userIDStr = v
		case fmt.Stringer:
			userIDStr = v.String()
		default:
			userIDStr = fmt.Sprintf("%v", v)
		}

		// 요청 바디 파싱 및 검증
		var req struct {
			EventType string `json:"event_type" binding:"required"`
		}

		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "unsubscribe_from_event")
			return
		}

		// 클라이언트 ID 찾기 (현재 사용자의 활성 연결)
		// 먼저 h.clients 맵에서 찾고, 없으면 realtime.Service의 connections 맵에서 찾기
		clientID := ""
		h.clientsMux.RLock()
		for id, client := range h.clients {
			if client.UserID == userIDStr {
				clientID = id
				break
			}
		}
		h.clientsMux.RUnlock()

		// h.clients에서 찾지 못했으면 realtime.Service의 connections에서 찾기
		if clientID == "" {
			conn, err := h.realtimeSvc.GetConnectionByUserID(userIDStr)
			if err == nil && conn != nil {
				clientID = conn.ID
			}
		}

		if clientID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "No active SSE connection found", 404), "unsubscribe_from_event")
			return
		}

		// 이벤트 구독 해제
		if err := h.UnsubscribeFromEvent(clientID, req.EventType); err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to unsubscribe: %v", err), 500), "unsubscribe_from_event")
			return
		}

		h.OK(c, gin.H{
			"client_id":  clientID,
			"event_type": req.EventType,
			"subscribed": false,
		}, "Successfully unsubscribed from event")
	}
}

// HandleGetConnectionInfo: 현재 사용자의 SSE 연결 정보를 조회합니다
func (h *SSEHandler) HandleGetConnectionInfo(c *gin.Context) {
	handler := h.Compose(
		h.getConnectionInfoHandler(),
		h.StandardCRUDDecorators("get_connection_info")...,
	)

	handler(c)
}

// getConnectionInfoHandler: 연결 정보 조회 핵심 로직
func (h *SSEHandler) getConnectionInfoHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// 사용자 ID 추출
		userIDValue, exists := c.Get("user_id")
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "get_connection_info")
			return
		}

		// user_id를 string으로 안전하게 변환
		var userIDStr string
		switch v := userIDValue.(type) {
		case string:
			userIDStr = v
		case fmt.Stringer:
			userIDStr = v.String()
		default:
			userIDStr = fmt.Sprintf("%v", v)
		}

		// 클라이언트 찾기 (현재 사용자의 활성 연결)
		clientID := ""
		var client *SSEClient
		h.clientsMux.RLock()
		for id, cl := range h.clients {
			if cl.UserID == userIDStr {
				clientID = id
				client = cl
				break
			}
		}
		h.clientsMux.RUnlock()

		// h.clients에서 찾지 못했으면 realtime.Service의 connections에서 찾기
		if clientID == "" {
			conn, err := h.realtimeSvc.GetConnectionByUserID(userIDStr)
			if err == nil && conn != nil {
				clientID = conn.ID
				// realtime.Service의 connections에 있지만 h.clients에 없으면 등록
				h.clientsMux.Lock()
				if _, exists := h.clients[conn.ID]; !exists {
					client = &SSEClient{
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
					h.clients[conn.ID] = client
				} else {
					client = h.clients[conn.ID]
				}
				h.clientsMux.Unlock()
			}
		}

		// 연결이 없으면 404 반환
		if clientID == "" || client == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "No active SSE connection found", 404), "get_connection_info")
			return
		}

		// 구독 중인 이벤트 타입 목록
		subscribedEvents := make([]string, 0, len(client.SubscribedEvents))
		for eventType := range client.SubscribedEvents {
			subscribedEvents = append(subscribedEvents, eventType)
		}

		// 필터 정보
		filters := gin.H{}
		if len(client.Filters.Providers) > 0 {
			providers := make([]string, 0, len(client.Filters.Providers))
			for provider := range client.Filters.Providers {
				providers = append(providers, provider)
			}
			filters["providers"] = providers
		}
		if len(client.Filters.CredentialIDs) > 0 {
			credentialIDs := make([]string, 0, len(client.Filters.CredentialIDs))
			for credentialID := range client.Filters.CredentialIDs {
				credentialIDs = append(credentialIDs, credentialID)
			}
			filters["credential_ids"] = credentialIDs
		}
		if len(client.Filters.Regions) > 0 {
			regions := make([]string, 0, len(client.Filters.Regions))
			for region := range client.Filters.Regions {
				regions = append(regions, region)
			}
			filters["regions"] = regions
		}

		// 연결 정보 반환
		h.OK(c, gin.H{
			"connection_id":     clientID,
			"user_id":           client.UserID,
			"last_seen":         client.LastSeen,
			"subscribed_events": subscribedEvents,
			"filters":           filters,
			"connected":         true,
		}, "Connection information retrieved successfully")
	}
}

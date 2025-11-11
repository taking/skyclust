package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/pkg/logger"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL                  string
	Cluster              string
	Subject              string
	CompressionType      CompressionType
	CompressionThreshold int // Minimum size in bytes to compress
}

// QueueSubscription represents a queue subscription with management capabilities
type QueueSubscription struct {
	Subject      string
	Queue        string
	Subscription *nats.Subscription
	Handler      func([]byte) error
	Concurrency  int
	RetryMax     int
	RetryDelay   time.Duration
	mu           sync.RWMutex
	activeCount  int64
	processed    int64
	failed       int64
}

// NATSService implements messaging using NATS
type NATSService struct {
	conn          *nats.Conn
	config        NATSConfig
	subscriptions map[string]*QueueSubscription
	subsMux       sync.RWMutex
}

// NewNATSService creates a new NATS service
func NewNATSService(config NATSConfig) (*NATSService, error) {
	// Set default compression settings if not specified
	if config.CompressionType == "" {
		config.CompressionType = CompressionNone
	}
	if config.CompressionThreshold == 0 {
		config.CompressionThreshold = 1024 // Default: 1KB threshold
	}

	opts := []nats.Option{
		nats.Name("CMP-Server"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			logger.Warn("NATS disconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Warn("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info(fmt.Sprintf("Successfully connected to NATS (compression: %s, threshold: %d bytes)", config.CompressionType, config.CompressionThreshold))
	return &NATSService{
		conn:          conn,
		config:        config,
		subscriptions: make(map[string]*QueueSubscription),
	}, nil
}

// Publish publishes an event (implements Bus interface)
func (n *NATSService) Publish(ctx context.Context, event Event) error {
	subject := fmt.Sprintf("cmp.events.%s", event.Type)
	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 압축 적용
	compressed, err := n.compressMessage(message)
	if err != nil {
		return fmt.Errorf("failed to compress message: %w", err)
	}

	return n.conn.Publish(subject, compressed)
}

// PublishToWorkspace publishes an event to a specific workspace
func (n *NATSService) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return n.Publish(ctx, *event)
}

// PublishToUser publishes an event to a specific user
func (n *NATSService) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return n.Publish(ctx, *event)
}

// Subscribe subscribes to an event type (implements Bus interface)
func (n *NATSService) Subscribe(eventType string, handler EventHandler) error {
	subject := fmt.Sprintf("cmp.events.%s", eventType)
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		// 압축 해제
		decompressed, err := n.decompressMessage(msg.Data)
		if err != nil {
			logger.Errorf("Failed to decompress message: %v", err)
			return
		}

		var event Event
		if err := json.Unmarshal(decompressed, &event); err != nil {
			logger.Errorf("Failed to unmarshal event: %v", err)
			return
		}

		if err := handler.Handle(context.Background(), event); err != nil {
			logger.Errorf("Error processing event: %v", err)
		}
	})
	return err
}

// Health checks NATS health (implements Bus interface)
func (n *NATSService) Health(ctx context.Context) error {
	if !n.conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}
	return nil
}

// GetConnection returns the underlying NATS connection
func (n *NATSService) GetConnection() *nats.Conn {
	return n.conn
}

// PublishMessage publishes a message to a subject (legacy method)
func (n *NATSService) PublishMessage(ctx context.Context, subject string, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 압축 적용
	compressed, err := n.compressMessage(message)
	if err != nil {
		return fmt.Errorf("failed to compress message: %w", err)
	}

	return n.conn.Publish(subject, compressed)
}

// SubscribeToSubject subscribes to a subject with a handler
func (n *NATSService) SubscribeToSubject(ctx context.Context, subject string, handler func([]byte) error) error {
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		// 압축 해제
		decompressed, err := n.decompressMessage(msg.Data)
		if err != nil {
			logger.Errorf("Failed to decompress message: %v", err)
			return
		}

		if err := handler(decompressed); err != nil {
			logger.Errorf("Error processing message: %v", err)
		}
	})
	return err
}

// SubscribeWithQueue subscribes to a subject with queue group
func (n *NATSService) SubscribeWithQueue(ctx context.Context, subject, queue string, handler func([]byte) error) error {
	return n.SubscribeWithQueueAdvanced(ctx, subject, queue, handler, 1, 0, 0)
}

// SubscribeWithQueueAdvanced subscribes to a subject with queue group with advanced options
func (n *NATSService) SubscribeWithQueueAdvanced(
	ctx context.Context,
	subject, queue string,
	handler func([]byte) error,
	concurrency int,
	retryMax int,
	retryDelay time.Duration,
) error {
	if concurrency <= 0 {
		concurrency = 1
	}
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	// 동시성 제어를 위한 세마포어
	sem := make(chan struct{}, concurrency)

	sub := &QueueSubscription{
		Subject:     subject,
		Queue:       queue,
		Handler:     handler,
		Concurrency: concurrency,
		RetryMax:    retryMax,
		RetryDelay:  retryDelay,
	}

	subscription, err := n.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		// 동시성 제어: 세마포어 획득
		sem <- struct{}{}
		defer func() { <-sem }()

		sub.mu.Lock()
		sub.activeCount++
		sub.mu.Unlock()

		// 압축 해제
		decompressed, decompressErr := n.decompressMessage(msg.Data)
		if decompressErr != nil {
			logger.Errorf("Failed to decompress message: %v", decompressErr)
			sub.mu.Lock()
			sub.activeCount--
			sub.failed++
			sub.mu.Unlock()
			return
		}

		// 메시지 처리 (재시도 로직 포함)
		err := n.processMessageWithRetry(sub, decompressed)

		sub.mu.Lock()
		sub.activeCount--
		if err != nil {
			sub.failed++
		} else {
			sub.processed++
		}
		sub.mu.Unlock()

		// ACK/NAK 처리는 JetStream에서만 지원되므로 일반 NATS에서는 로깅만 수행
		if err != nil {
			logger.Errorf("Message processing failed after retries: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to queue %s on subject %s: %w", queue, subject, err)
	}

	sub.Subscription = subscription

	// 구독 관리 맵에 저장
	n.subsMux.Lock()
	key := fmt.Sprintf("%s:%s", subject, queue)
	n.subscriptions[key] = sub
	n.subsMux.Unlock()

	logger.Info(fmt.Sprintf("Subscribed to queue %s on subject %s with concurrency %d", queue, subject, concurrency))
	return nil
}

// processMessageWithRetry processes a message with retry logic
func (n *NATSService) processMessageWithRetry(sub *QueueSubscription, data []byte) error {
	var lastErr error
	for attempt := 0; attempt <= sub.RetryMax; attempt++ {
		if attempt > 0 {
			// 재시도 전 대기
			time.Sleep(sub.RetryDelay * time.Duration(attempt))
			logger.Warn(fmt.Sprintf("Retrying message processing (attempt %d/%d)", attempt, sub.RetryMax))
		}

		err := sub.Handler(data)
		if err == nil {
			return nil
		}

		lastErr = err
		logger.Errorf("Error processing message (attempt %d/%d): %v", attempt+1, sub.RetryMax+1, err)
	}

	return fmt.Errorf("failed after %d attempts: %w", sub.RetryMax+1, lastErr)
}

// UnsubscribeFromQueue unsubscribes from a queue subscription
func (n *NATSService) UnsubscribeFromQueue(subject, queue string) error {
	n.subsMux.Lock()
	defer n.subsMux.Unlock()

	key := fmt.Sprintf("%s:%s", subject, queue)
	sub, exists := n.subscriptions[key]
	if !exists {
		return fmt.Errorf("subscription not found for subject %s queue %s", subject, queue)
	}

	if sub.Subscription != nil {
		if err := sub.Subscription.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	delete(n.subscriptions, key)
	logger.Info(fmt.Sprintf("Unsubscribed from queue %s on subject %s", queue, subject))
	return nil
}

// GetQueueSubscriptionStats returns statistics for a queue subscription
func (n *NATSService) GetQueueSubscriptionStats(subject, queue string) (map[string]interface{}, error) {
	n.subsMux.RLock()
	defer n.subsMux.RUnlock()

	key := fmt.Sprintf("%s:%s", subject, queue)
	sub, exists := n.subscriptions[key]
	if !exists {
		return nil, fmt.Errorf("subscription not found for subject %s queue %s", subject, queue)
	}

	sub.mu.RLock()
	defer sub.mu.RUnlock()

	return map[string]interface{}{
		"subject":      sub.Subject,
		"queue":        sub.Queue,
		"concurrency":  sub.Concurrency,
		"active_count": sub.activeCount,
		"processed":    sub.processed,
		"failed":       sub.failed,
		"success_rate": func() float64 {
			total := sub.processed + sub.failed
			if total == 0 {
				return 0.0
			}
			return float64(sub.processed) / float64(total) * 100
		}(),
	}, nil
}

// GetAllQueueSubscriptionStats returns statistics for all queue subscriptions
func (n *NATSService) GetAllQueueSubscriptionStats() map[string]map[string]interface{} {
	n.subsMux.RLock()
	defer n.subsMux.RUnlock()

	result := make(map[string]map[string]interface{})
	for key, sub := range n.subscriptions {
		sub.mu.RLock()
		stats := map[string]interface{}{
			"subject":      sub.Subject,
			"queue":        sub.Queue,
			"concurrency":  sub.Concurrency,
			"active_count": sub.activeCount,
			"processed":    sub.processed,
			"failed":       sub.failed,
			"success_rate": func() float64 {
				total := sub.processed + sub.failed
				if total == 0 {
					return 0.0
				}
				return float64(sub.processed) / float64(total) * 100
			}(),
		}
		sub.mu.RUnlock()
		result[key] = stats
	}

	return result
}

// compressMessage compresses a message if compression is enabled and threshold is met
func (n *NATSService) compressMessage(data []byte) ([]byte, error) {
	if !ShouldCompress(data, n.config.CompressionThreshold, n.config.CompressionType) {
		return data, nil
	}

	compressed, err := Compress(data, n.config.CompressionType)
	if err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}

	// 압축이 효과적이지 않은 경우 원본 반환 (압축된 크기가 더 크거나 비슷한 경우)
	if len(compressed) >= len(data) {
		return data, nil
	}

	return compressed, nil
}

// decompressMessage decompresses a message, handling both compressed and uncompressed data
func (n *NATSService) decompressMessage(data []byte) ([]byte, error) {
	// 압축되지 않은 데이터인지 확인 (간단한 휴리스틱: gzip 헤더 또는 snappy 매직 바이트 확인)
	if isCompressed(data) {
		return Decompress(data, n.config.CompressionType)
	}

	// 압축되지 않은 데이터인 경우 그대로 반환
	return data, nil
}

// isCompressed checks if data appears to be compressed
func isCompressed(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// gzip magic number: 0x1f 0x8b
	if data[0] == 0x1f && data[1] == 0x8b {
		return true
	}

	// snappy magic number starts with specific bytes (simplified check)
	// 실제로는 더 정교한 검사가 필요할 수 있음
	if len(data) >= 4 && data[0] == 's' && data[1] == 'N' && data[2] == 'a' && data[3] == 'P' {
		return true
	}

	return false
}

// Request sends a request and waits for a response
func (n *NATSService) Request(ctx context.Context, subject string, data interface{}, timeout time.Duration) ([]byte, error) {
	message, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := n.conn.Request(subject, message, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return msg.Data, nil
}

// Close closes the NATS connection
func (n *NATSService) Close() {
	n.conn.Close()
}

// GetStats returns NATS connection statistics
func (n *NATSService) GetStats() map[string]interface{} {
	stats := n.conn.Statistics
	return map[string]interface{}{
		"in_msgs":    stats.InMsgs,
		"out_msgs":   stats.OutMsgs,
		"in_bytes":   stats.InBytes,
		"out_bytes":  stats.OutBytes,
		"reconnects": stats.Reconnects,
		"connected":  n.conn.IsConnected(),
	}
}

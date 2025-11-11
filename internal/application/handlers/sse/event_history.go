package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventHistory represents a stored SSE event in Redis Streams
type EventHistory struct {
	EventID    string
	EventType  string
	Data       []byte
	Timestamp  time.Time
	UserID     string
}

// Redis Streams key patterns
const (
	// sse:events:{user_id}:{event_type} - Redis Stream for event history
	eventStreamKeyPattern = "sse:events:%s:%s"
	// Maximum number of events to keep in stream (1000 events)
	maxStreamLength = 1000
	// TTL for event streams (1 hour)
	eventStreamTTL = 1 * time.Hour
)

// saveEventToHistory saves an event to Redis Streams for event history
func (h *SSEHandler) saveEventToHistory(ctx context.Context, userID, eventType string, eventID string, data interface{}) {
	if h.redisClient == nil {
		return
	}

	// Marshal event data
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Warn("Failed to marshal event data for history",
			zap.Error(err),
			zap.String("event_type", eventType),
			zap.String("event_id", eventID))
		return
	}

	// Redis Stream key
	streamKey := fmt.Sprintf(eventStreamKeyPattern, userID, eventType)

	// Add event to stream
	// XADD sse:events:{user_id}:{event_type} * event_id {id} data {data} timestamp {timestamp}
	args := map[string]interface{}{
		"event_id":  eventID,
		"data":      string(jsonData),
		"timestamp": time.Now().Unix(),
	}

	// Add to stream with automatic ID generation
	streamID, err := h.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: maxStreamLength,
		Approx: true, // Approximate trimming for better performance
		Values: args,
	}).Result()

	if err != nil {
		h.logger.Warn("Failed to save event to Redis Stream",
			zap.Error(err),
			zap.String("stream_key", streamKey),
			zap.String("event_type", eventType),
			zap.String("event_id", eventID))
		return
	}

	// Set TTL on stream
	h.redisClient.Expire(ctx, streamKey, eventStreamTTL)

	h.logger.Debug("Event saved to history",
		zap.String("stream_key", streamKey),
		zap.String("stream_id", streamID),
		zap.String("event_type", eventType),
		zap.String("event_id", eventID))
}

// getMissedEvents retrieves missed events from Redis Streams after lastEventID
func (h *SSEHandler) getMissedEvents(ctx context.Context, userID, lastEventID string) ([]EventHistory, error) {
	if h.redisClient == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	// Get all event types for this user
	// Pattern: sse:events:{user_id}:*
	pattern := fmt.Sprintf("sse:events:%s:*", userID)
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get event stream keys: %w", err)
	}

	var allEvents []EventHistory

	// Read from each stream
	for _, streamKey := range keys {
		// Extract event type from stream key
		// sse:events:{user_id}:{event_type}
		// streamKey format: "sse:events:{user_id}:{event_type}"
		prefix := fmt.Sprintf("sse:events:%s:", userID)
		if !strings.HasPrefix(streamKey, prefix) {
			h.logger.Warn("Stream key does not match expected pattern",
				zap.String("stream_key", streamKey),
				zap.String("expected_prefix", prefix))
			continue
		}
		eventType := strings.TrimPrefix(streamKey, prefix)

		// Read events after lastEventID
		// XREAD STREAMS {stream_key} {last_event_id}
		// lastEventID가 stream ID 형식이 아니면 "0"으로 시작 (모든 이벤트)
		startID := lastEventID
		if startID == "" || !strings.Contains(startID, "-") {
			// Stream ID 형식이 아니면 "0"으로 시작 (모든 이벤트)
			startID = "0"
		}
		streams, err := h.redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamKey, startID},
			Count:   1000, // Maximum events to read
			Block:   0,    // Non-blocking
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// No new events
				continue
			}
			h.logger.Warn("Failed to read from stream",
				zap.String("stream_key", streamKey),
				zap.Error(err))
			continue
		}

		// Parse events from stream
		for _, stream := range streams {
			for _, message := range stream.Messages {
				eventIDVal, ok := message.Values["event_id"]
				if !ok {
					continue
				}
				eventID, ok := eventIDVal.(string)
				if !ok {
					continue
				}

				dataStrVal, ok := message.Values["data"]
				if !ok {
					continue
				}
				dataStr, ok := dataStrVal.(string)
				if !ok {
					continue
				}

				timestampVal, ok := message.Values["timestamp"]
				if !ok {
					continue
				}
				timestampStr, ok := timestampVal.(string)
				if !ok {
					continue
				}

				var timestamp int64
				if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
					h.logger.Warn("Failed to parse timestamp",
						zap.String("timestamp", timestampStr),
						zap.Error(err))
					continue
				}

				allEvents = append(allEvents, EventHistory{
					EventID:   eventID,
					EventType: eventType,
					Data:      []byte(dataStr),
					Timestamp: time.Unix(timestamp, 0),
					UserID:    userID,
				})
			}
		}
	}

	// Sort events by timestamp
	// (Redis Streams already maintain order, but we ensure it)
	return allEvents, nil
}

// getMissedEventsByEventType retrieves missed events for a specific event type
func (h *SSEHandler) getMissedEventsByEventType(ctx context.Context, userID, eventType, lastEventID string) ([]EventHistory, error) {
	if h.redisClient == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	streamKey := fmt.Sprintf(eventStreamKeyPattern, userID, eventType)

	// Read events after lastEventID
	// lastEventID가 stream ID 형식이 아니면 "0"으로 시작 (모든 이벤트)
	startID := lastEventID
	if startID == "" || !strings.Contains(startID, "-") {
		// Stream ID 형식이 아니면 "0"으로 시작 (모든 이벤트)
		startID = "0"
	}
	streams, err := h.redisClient.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, startID},
		Count:   1000,
		Block:   0,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []EventHistory{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var events []EventHistory

	for _, stream := range streams {
		for _, message := range stream.Messages {
			eventIDVal, ok := message.Values["event_id"]
			if !ok {
				continue
			}
			eventID, ok := eventIDVal.(string)
			if !ok {
				continue
			}

			dataStrVal, ok := message.Values["data"]
			if !ok {
				continue
			}
			dataStr, ok := dataStrVal.(string)
			if !ok {
				continue
			}

			timestampVal, ok := message.Values["timestamp"]
			if !ok {
				continue
			}
			timestampStr, ok := timestampVal.(string)
			if !ok {
				continue
			}

			var timestamp int64
			if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
				h.logger.Warn("Failed to parse timestamp",
					zap.String("timestamp", timestampStr),
					zap.Error(err))
				continue
			}

			events = append(events, EventHistory{
				EventID:   eventID,
				EventType:  eventType,
				Data:       []byte(dataStr),
				Timestamp:  time.Unix(timestamp, 0),
				UserID:     userID,
			})
		}
	}

	return events, nil
}


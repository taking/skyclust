package sse

import (
	"sync"
	"time"
)

// BatchBuffer buffers multiple events and sends them together
type BatchBuffer struct {
	events    []BatchEvent
	mu        sync.Mutex
	timer     *time.Timer
	flushFunc func([]BatchEvent)
	maxSize   int
	interval  time.Duration
}

// BatchEvent represents an event in the batch buffer
type BatchEvent struct {
	EventType string
	Data      interface{}
	Timestamp time.Time
}

// NewBatchBuffer creates a new batch buffer
func NewBatchBuffer(maxSize int, interval time.Duration, flushFunc func([]BatchEvent)) *BatchBuffer {
	buffer := &BatchBuffer{
		events:    make([]BatchEvent, 0, maxSize),
		flushFunc: flushFunc,
		maxSize:   maxSize,
		interval:  interval,
		timer:     time.NewTimer(interval),
	}

	// Start flush timer
	go buffer.flushLoop()

	return buffer
}

// Add adds an event to the batch buffer
func (b *BatchBuffer) Add(eventType string, data interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, BatchEvent{
		EventType: eventType,
		Data:      data,
		Timestamp: time.Now(),
	})

	// If buffer is full, flush immediately
	if len(b.events) >= b.maxSize {
		b.flush()
	}
}

// Flush immediately flushes all buffered events
func (b *BatchBuffer) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
}

// flush sends all buffered events (must be called with lock held)
func (b *BatchBuffer) flush() {
	if len(b.events) == 0 {
		return
	}

	events := make([]BatchEvent, len(b.events))
	copy(events, b.events)
	b.events = b.events[:0]

	// Reset timer
	b.timer.Reset(b.interval)

	// Flush events (run in goroutine to avoid blocking)
	go b.flushFunc(events)
}

// flushLoop periodically flushes the buffer
func (b *BatchBuffer) flushLoop() {
	for range b.timer.C {
		b.Flush()
	}
}

// Close closes the batch buffer and flushes remaining events
func (b *BatchBuffer) Close() {
	b.timer.Stop()
	b.Flush()
}

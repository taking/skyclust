package telemetry

import (
	"context"
	"fmt"
	"time"
)

// Telemetry provides telemetry functionality
type Telemetry struct {
	enabled bool
}

// NewTelemetry creates a new telemetry instance
func NewTelemetry(enabled bool) *Telemetry {
	return &Telemetry{
		enabled: enabled,
	}
}

// StartSpan starts a new span
func (t *Telemetry) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if !t.enabled {
		return ctx, &NoOpSpan{}
	}

	// Simple implementation for now
	return ctx, &SimpleSpan{
		name:  name,
		start: time.Now(),
	}
}

// Span represents a telemetry span
type Span interface {
	End()
	SetAttribute(key, value string)
	SetStatus(code int, message string)
}

// SimpleSpan is a simple span implementation
type SimpleSpan struct {
	name  string
	start time.Time
}

// End ends the span
func (s *SimpleSpan) End() {
	// Simple implementation
}

// SetAttribute sets a span attribute
func (s *SimpleSpan) SetAttribute(key, value string) {
	// Simple implementation
}

// SetStatus sets the span status
func (s *SimpleSpan) SetStatus(code int, message string) {
	// Simple implementation
}

// NoOpSpan is a no-op span implementation
type NoOpSpan struct{}

// End ends the span
func (s *NoOpSpan) End() {}

// SetAttribute sets a span attribute
func (s *NoOpSpan) SetAttribute(key, value string) {}

// SetStatus sets the span status
func (s *NoOpSpan) SetStatus(code int, message string) {}

// StartSpan starts a new span (global function)
func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	// Simple implementation
	return ctx, &NoOpSpan{}
}

// EndSpan ends a span
func EndSpan(span Span) {
	if span != nil {
		span.End()
	}
}

// SpanFromContext retrieves a span from context
func SpanFromContext(ctx context.Context) Span {
	// Simple implementation
	return &NoOpSpan{}
}

// RecordError records an error in telemetry
func RecordError(span Span, err error) {
	if span != nil {
		span.SetAttribute("error", err.Error())
		span.SetStatus(500, "error")
	}
}

// AddEvent adds an event to telemetry
func AddEvent(span Span, name string, attrs map[string]interface{}) {
	if span != nil {
		for k, v := range attrs {
			span.SetAttribute(k, fmt.Sprint(v))
		}
	}
}

// SetAttributes sets attributes on a span
func SetAttributes(span Span, attrs map[string]interface{}) {
	if span != nil {
		for k, v := range attrs {
			span.SetAttribute(k, fmt.Sprint(v))
		}
	}
}

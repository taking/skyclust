package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new tracer
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(serviceName),
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new span with attributes
func (t *Tracer) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}) (context.Context, trace.Span) {
	spanOpts := []trace.SpanStartOption{
		trace.WithAttributes(t.convertAttributes(attrs)...),
	}
	return t.tracer.Start(ctx, name, spanOpts...)
}

// RecordError records an error in the current span
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetAttributes sets attributes on a span
func SetAttributes(span trace.Span, attrs map[string]interface{}) {
	tracer := &Tracer{}
	span.SetAttributes(tracer.convertAttributes(attrs)...)
}

// AddEvent adds an event to a span
func AddEvent(span trace.Span, name string, attrs map[string]interface{}) {
	tracer := &Tracer{}
	span.AddEvent(name, trace.WithAttributes(tracer.convertAttributes(attrs)...))
}

// convertAttributes converts map to OpenTelemetry attributes
func (t *Tracer) convertAttributes(attrs map[string]interface{}) []attribute.KeyValue {
	var result []attribute.KeyValue
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			result = append(result, attribute.String(key, v))
		case int:
			result = append(result, attribute.Int(key, v))
		case int64:
			result = append(result, attribute.Int64(key, v))
		case float64:
			result = append(result, attribute.Float64(key, v))
		case bool:
			result = append(result, attribute.Bool(key, v))
		case []string:
			result = append(result, attribute.StringSlice(key, v))
		case time.Time:
			result = append(result, attribute.String(key, v.Format(time.RFC3339)))
		default:
			result = append(result, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
	return result
}

// SpanFromContext extracts span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan adds span to context
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

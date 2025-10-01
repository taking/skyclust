package telemetry

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cmp/pkg/shared/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates a Gin middleware for distributed tracing
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := NewTracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract trace context from headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start span
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		ctx, span := tracer.StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
			"http.method":      c.Request.Method,
			"http.url":         c.Request.URL.String(),
			"http.user_agent":  c.Request.UserAgent(),
			"http.remote_addr": c.ClientIP(),
		})

		// Add request ID to context
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		type requestIDKey struct{}
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)
		c.Set("request_id", requestID)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Set response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		// Record error if status code indicates error
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
		}

		// Add trace headers to response
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		span.End()

		// Log request
		logger.Info(fmt.Sprintf("Request completed: %s %s - status: %d, duration: %dms, request_id: %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration.Milliseconds(), requestID))
	}
}

// DatabaseTracingMiddleware creates a middleware for database operations
func DatabaseTracingMiddleware(operation string) func(context.Context, string, ...interface{}) (context.Context, func(error)) {
	return func(ctx context.Context, query string, args ...interface{}) (context.Context, func(error)) {
		span := trace.SpanFromContext(ctx)
		if !span.IsRecording() {
			return ctx, func(error) {}
		}

		// Start database span
		ctx, dbSpan := otel.Tracer("database").Start(ctx, operation)
		dbSpan.SetAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.statement", query),
			attribute.Int("db.args_count", len(args)),
		)

		start := time.Now()

		return ctx, func(err error) {
			duration := time.Since(start)
			dbSpan.SetAttributes(
				attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
			)

			if err != nil {
				RecordError(dbSpan, err)
			}

			dbSpan.End()
		}
	}
}

// ExternalAPITracingMiddleware creates a middleware for external API calls
func ExternalAPITracingMiddleware(serviceName string) func(context.Context, string, string) (context.Context, func(error)) {
	return func(ctx context.Context, method, url string) (context.Context, func(error)) {
		span := trace.SpanFromContext(ctx)
		if !span.IsRecording() {
			return ctx, func(error) {}
		}

		// Start external API span
		ctx, apiSpan := otel.Tracer("external-api").Start(ctx, fmt.Sprintf("%s %s", method, serviceName))
		apiSpan.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
			attribute.String("external.service", serviceName),
		)

		start := time.Now()

		return ctx, func(err error) {
			duration := time.Since(start)
			apiSpan.SetAttributes(
				attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
			)

			if err != nil {
				RecordError(apiSpan, err)
			}

			apiSpan.End()
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

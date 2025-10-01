package middleware

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"cmp/pkg/ratelimit"
	"cmp/pkg/shared/errors"
	"cmp/pkg/shared/logger"
	"cmp/pkg/shared/security"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides comprehensive security middleware
type SecurityMiddleware struct {
	rateLimiter *ratelimit.SimpleRateLimiter
	validator   *security.InputValidator
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(rateLimiter *ratelimit.SimpleRateLimiter) *SecurityMiddleware {
	return &SecurityMiddleware{
		rateLimiter: rateLimiter,
		validator:   security.NewInputValidator(),
	}
}

// CORS middleware for cross-origin requests
func (s *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins in production
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://yourdomain.com",
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit middleware for rate limiting
func (s *SecurityMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		allowed, err := s.rateLimiter.AllowIP(c.Request.Context(), ip)
		if err != nil {
			logger.Error("Rate limit check failed")
			_ = c.Error(errors.NewInternalError("Rate limit check failed"))
			return
		}

		if !allowed {
			apiErr := errors.NewAPIError(
				errors.ErrCodeResourceExhausted,
				"Rate limit exceeded",
				http.StatusTooManyRequests,
			)
			_ = apiErr.WithDetails("retry_after", "60")
			_ = c.Error(apiErr)
			return
		}

		c.Next()
	}
}

// InputValidation middleware for input validation
func (s *SecurityMiddleware) InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request body if present
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// Read and validate JSON body
				body, err := c.GetRawData()
				if err != nil {
					_ = c.Error(errors.NewValidationError("Invalid request body"))
					return
				}

				// Restore body for further processing
				c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

				// Validate JSON structure
				if err := s.validator.ValidateJSON(body); err != nil {
					_ = c.Error(errors.NewValidationError("Invalid JSON structure"))
					return
				}
			}
		}

		// Validate query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if err := s.validator.ValidateQueryParam(key, value); err != nil {
					_ = c.Error(errors.NewValidationError("Invalid query parameter: " + key))
					return
				}
			}
		}

		c.Next()
	}
}

// XSSProtection middleware for XSS protection
func (s *SecurityMiddleware) XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// CSRFProtection middleware for CSRF protection
func (s *SecurityMiddleware) CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check CSRF token
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			_ = c.Error(errors.NewForbiddenError("CSRF token required"))
			return
		}

		// Validate CSRF token (implement your CSRF validation logic)
		if !s.validateCSRFToken(token) {
			_ = c.Error(errors.NewForbiddenError("Invalid CSRF token"))
			return
		}

		c.Next()
	}
}

// SecurityHeaders middleware for security headers
func (s *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// RequestSizeLimit middleware for request size limiting
func (s *SecurityMiddleware) RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			_ = c.Error(errors.NewAPIError(
				errors.ErrCodeInvalidInput,
				"Request too large",
				http.StatusRequestEntityTooLarge,
			))
			return
		}
		c.Next()
	}
}

// Timeout middleware for request timeout
func (s *SecurityMiddleware) Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// validateCSRFToken validates CSRF token (implement your logic)
func (s *SecurityMiddleware) validateCSRFToken(token string) bool {
	// Implement your CSRF token validation logic
	// This is a placeholder implementation
	return len(token) > 0
}

// getClientIP gets the real client IP address
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

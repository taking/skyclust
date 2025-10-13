package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Middleware provides comprehensive middleware functionality
type Middleware struct {
	logger      *zap.Logger
	config      *MiddlewareConfig
	rateLimiter RateLimiter
	authService AuthService
	rbacService RBACService
	auditLogger AuditLogger
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	// API Configuration
	APIVersion       string `json:"api_version"`
	DefaultPageSize  int    `json:"default_page_size"`
	MaxPageSize      int    `json:"max_page_size"`
	DefaultSortField string `json:"default_sort_field"`
	DefaultSortOrder string `json:"default_sort_order"`

	// Security Configuration
	SecurityHeaders   bool          `json:"security_headers"`
	CORSEnabled       bool          `json:"cors_enabled"`
	RateLimitEnabled  bool          `json:"rate_limit_enabled"`
	RateLimitRequests int           `json:"rate_limit_requests"`
	RateLimitWindow   time.Duration `json:"rate_limit_window"`
	MaxRequestSize    int64         `json:"max_request_size"`
	RequestTimeout    time.Duration `json:"request_timeout"`

	// Logging Configuration
	LoggingEnabled    bool   `json:"logging_enabled"`
	LogLevel          string `json:"log_level"`
	StructuredLogging bool   `json:"structured_logging"`

	// Monitoring Configuration
	MetricsEnabled     bool `json:"metrics_enabled"`
	HealthCheckEnabled bool `json:"health_check_enabled"`

	// CORS Configuration
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	CheckLimit(key string, limit int, window time.Duration) (bool, error)
	GetRemaining(key string) (int, error)
	GetResetTime(key string) (time.Time, error)
}

// AuthService interface for authentication
type AuthService interface {
	ValidateToken(token string) (*domain.User, error)
}

// RBACService interface for role-based access control
type RBACService interface {
	CheckPermission(userID string, resource, action string) (bool, error)
	GetUserRoles(userID string) ([]string, error)
}

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogEvent(eventType, userID, resource, action string, success bool, details map[string]interface{})
	LogSecurityEvent(eventType, userID, ip, userAgent string, details map[string]interface{})
}

// NewMiddleware creates a new middleware
func NewMiddleware(
	logger *zap.Logger,
	config *MiddlewareConfig,
	rateLimiter RateLimiter,
	authService AuthService,
	rbacService RBACService,
	auditLogger AuditLogger,
) *Middleware {
	if config == nil {
		config = GetDefaultMiddlewareConfig()
	}

	return &Middleware{
		logger:      logger,
		config:      config,
		rateLimiter: rateLimiter,
		authService: authService,
		rbacService: rbacService,
		auditLogger: auditLogger,
	}
}

// GetDefaultMiddlewareConfig returns default middleware configuration
func GetDefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		APIVersion:         "1.0",
		DefaultPageSize:    10,
		MaxPageSize:        100,
		DefaultSortField:   "created_at",
		DefaultSortOrder:   "desc",
		SecurityHeaders:    true,
		CORSEnabled:        true,
		RateLimitEnabled:   true,
		RateLimitRequests:  100,
		RateLimitWindow:    time.Minute,
		MaxRequestSize:     10 * 1024 * 1024, // 10MB
		RequestTimeout:     30 * time.Second,
		LoggingEnabled:     true,
		LogLevel:           "info",
		StructuredLogging:  true,
		MetricsEnabled:     true,
		HealthCheckEnabled: true,
		AllowedOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID", "X-API-Version"},
	}
}

// RequestIDMiddleware generates and sets request ID
func (m *Middleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// VersionMiddleware handles API versioning
func (m *Middleware) VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("X-API-Version")
		if version == "" {
			version = m.config.APIVersion
		}

		c.Set("api_version", version)
		c.Header("X-API-Version", version)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (m *Middleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.SecurityHeaders {
			c.Next()
			return
		}

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CORSMiddleware handles CORS
func (m *Middleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.CORSEnabled {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range m.config.AllowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware provides rate limiting
func (m *Middleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.RateLimitEnabled || m.rateLimiter == nil {
			c.Next()
			return
		}

		// Generate rate limit key
		key := m.generateRateLimitKey(c)

		// Check rate limit
		allowed, err := m.rateLimiter.CheckLimit(key, m.config.RateLimitRequests, m.config.RateLimitWindow)
		if err != nil {
			m.logger.Error("Rate limit check failed", zap.Error(err))
			c.Next() // Continue on error
			return
		}

		if !allowed {
			// Log rate limit hit
			if m.auditLogger != nil {
				m.auditLogger.LogSecurityEvent("rate_limit_hit", "", c.ClientIP(), c.Request.UserAgent(), map[string]interface{}{
					"key":    key,
					"limit":  m.config.RateLimitRequests,
					"window": m.config.RateLimitWindow,
				})
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		if remaining, err := m.rateLimiter.GetRemaining(key); err == nil {
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		}
		if resetTime, err := m.rateLimiter.GetResetTime(key); err == nil {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		}

		c.Next()
	}
}

// AuthMiddleware provides authentication
func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.authService == nil {
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.unauthorizedResponse(c, "Authorization header required")
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.unauthorizedResponse(c, "Invalid authorization header format")
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			m.unauthorizedResponse(c, "Token required")
			return
		}

		// Validate token
		user, err := m.authService.ValidateToken(token)
		if err != nil {
			m.unauthorizedResponse(c, "Invalid token")
			return
		}

		// Set user information in context
		c.Set("user", user)
		if userID := m.extractUserID(user); userID != "" {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// RBACMiddleware provides role-based access control
func (m *Middleware) RBACMiddleware(requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.rbacService == nil {
			c.Next()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			m.forbiddenResponse(c, "User not authenticated")
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.forbiddenResponse(c, "Invalid user ID")
			return
		}

		// Check permissions
		hasPermission := true
		for _, permission := range requiredPermissions {
			allowed, err := m.rbacService.CheckPermission(userIDStr, permission, "execute")
			if err != nil {
				m.logger.Error("Permission check failed", zap.Error(err), zap.String("permission", permission))
				hasPermission = false
				break
			}
			if !allowed {
				hasPermission = false
				break
			}
		}

		if !hasPermission {
			m.forbiddenResponse(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// LoggingMiddleware provides request/response logging
func (m *Middleware) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.LoggingEnabled {
			c.Next()
			return
		}

		start := time.Now()
		requestID := c.GetString("request_id")

		// Log request
		if m.config.StructuredLogging {
			m.logger.Info("Request started",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)
		}

		// Process request
		c.Next()

		// Log response
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		if m.config.StructuredLogging {
			m.logger.Info("Request completed",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.Int("status_code", statusCode),
				zap.Duration("duration", duration),
				zap.Int("response_size", c.Writer.Size()),
			)
		}
	}
}

// ErrorHandlingMiddleware provides centralized error handling
func (m *Middleware) ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("Request panicked",
					zap.String("request_id", c.GetString("request_id")),
					zap.String("method", c.Request.Method),
					zap.String("path", c.FullPath()),
					zap.Any("error", err),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal server error",
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ValidationMiddleware provides request validation
func (m *Middleware) ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request size
		if c.Request.ContentLength > m.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "Request too large",
				"max_size": m.config.MaxRequestSize,
			})
			c.Abort()
			return
		}

		// Validate required headers
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			if c.GetHeader("Content-Type") == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Content-Type header required",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// TimeoutMiddleware adds request timeout
func (m *Middleware) TimeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.RequestTimeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal completion
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			m.logger.Warn("Request timeout",
				zap.String("request_id", c.GetString("request_id")),
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.Duration("timeout", m.config.RequestTimeout),
			)

			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"timeout": m.config.RequestTimeout.Milliseconds(),
			})
			c.Abort()
		}
	}
}

// PaginationMiddleware handles pagination parameters
func (m *Middleware) PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := 1
		limit := m.config.DefaultPageSize

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= m.config.MaxPageSize {
				limit = l
			}
		}

		c.Set("page", page)
		c.Set("limit", limit)
		c.Next()
	}
}

// HealthCheckMiddleware provides health check endpoint
func (m *Middleware) HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.HealthCheckEnabled {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
				"version":   m.config.APIVersion,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ApplyAllMiddleware applies all middleware in the correct order
func (m *Middleware) ApplyAllMiddleware(router *gin.Engine) {
	// Apply middleware in order
	router.Use(m.RequestIDMiddleware())
	router.Use(m.VersionMiddleware())
	router.Use(m.SecurityHeadersMiddleware())
	router.Use(m.CORSMiddleware())
	router.Use(m.ValidationMiddleware())
	router.Use(m.TimeoutMiddleware())
	router.Use(m.RateLimitMiddleware())
	router.Use(m.LoggingMiddleware())
	router.Use(m.ErrorHandlingMiddleware())
	router.Use(m.PaginationMiddleware())
	router.Use(m.HealthCheckMiddleware())
}

// Helper methods

func (m *Middleware) generateRateLimitKey(c *gin.Context) string {
	// Use IP address as the primary key
	ip := c.ClientIP()

	// Optionally include user ID if available
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("%s:%s", ip, userID)
	}

	return ip
}

func (m *Middleware) extractUserID(user interface{}) string {
	// Handle *domain.User type
	if domainUser, ok := user.(*domain.User); ok {
		return domainUser.ID.String()
	}

	// Fallback for map[string]interface{} type
	if userMap, ok := user.(map[string]interface{}); ok {
		if id, exists := userMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				return idStr
			}
		}
	}
	return ""
}

func (m *Middleware) unauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error":      message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

func (m *Middleware) forbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":      message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

package handlers

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/responses"
	"skyclust/pkg/auth"
	"skyclust/pkg/logger"
	"skyclust/pkg/telemetry"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BaseHandler provides common functionality for all API handlers
type BaseHandler struct {
	logger             *zap.Logger
	tokenExtractor     *auth.TokenExtractor
	performanceTracker *PerformanceTracker
	requestLogger      *RequestLogger
	auditLogger        *AuditLogger
	validationRules    *ValidationRules
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(handlerName string) *BaseHandler {
	return &BaseHandler{
		logger:             logger.DefaultLogger.GetLogger(),
		tokenExtractor:     auth.NewTokenExtractor(),
		performanceTracker: NewPerformanceTracker(handlerName),
		requestLogger:      NewRequestLogger(nil),
		auditLogger:        NewAuditLogger(nil),
		validationRules:    NewValidationRules(),
	}
}

// GetUserIDFromToken extracts user ID from JWT token
func (h *BaseHandler) GetUserIDFromToken(c *gin.Context) (uuid.UUID, error) {
	return h.tokenExtractor.GetUserIDFromToken(c)
}

// GetUserRoleFromToken extracts user role from JWT token
func (h *BaseHandler) GetUserRoleFromToken(c *gin.Context) (domain.Role, error) {
	return h.tokenExtractor.GetUserRoleFromToken(c)
}

// GetBearerTokenFromHeader extracts Bearer token from Authorization header
func (h *BaseHandler) GetBearerTokenFromHeader(c *gin.Context) (string, error) {
	return h.tokenExtractor.GetBearerTokenFromHeader(c)
}

// GetCredentialFromRequest extracts and validates credential from request
// This is a common pattern across all cloud provider handlers
func (h *BaseHandler) GetCredentialFromRequest(c *gin.Context, credentialService domain.CredentialService, expectedProvider string) (*domain.Credential, error) {
	// 1. Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid token", 401)
	}

	// 2. Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400)
	}

	// 3. Parse credential UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid credential ID format", 400)
	}

	// 4. Get credential from service
	credential, err := credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		return nil, err
	}

	// 5. Verify credential matches expected provider
	if expectedProvider != "" && credential.Provider != expectedProvider {
		return nil, domain.NewDomainError(
			domain.ErrCodeBadRequest,
			"Credential provider does not match "+expectedProvider,
			400,
		)
	}

	return credential, nil
}

// ValidateRequest validates the request body against the provided struct with enhanced error handling
func (h *BaseHandler) ValidateRequest(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		// Enhanced validation with validation rules
		validationErrors := make(map[string]string)
		validationErrors["binding"] = err.Error()

		// Additional validation using validation rules
		if reqStruct, ok := req.(interface{ Validate() error }); ok {
			if validateErr := reqStruct.Validate(); validateErr != nil {
				validationErrors["validation"] = validateErr.Error()
			}
		}

		// Log validation error with context
		h.LogWarn(c, "Request validation failed",
			zap.Any("errors", validationErrors),
			zap.String("content_type", c.GetHeader("Content-Type")))

		// Return domain error instead of direct response
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "Request validation failed", 400)
	}
	return nil
}

// ValidateQueryParams validates query parameters
func (h *BaseHandler) ValidateQueryParams(c *gin.Context, params map[string]string) error {
	validationErrors := make(map[string]string)

	for paramName, paramValue := range params {
		if paramValue == "" {
			validationErrors[paramName] = "required parameter missing"
		}
	}

	if len(validationErrors) > 0 {
		h.LogWarn(c, "Query parameter validation failed", zap.Any("errors", validationErrors))
		responses.ValidationError(c, validationErrors)
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "query parameter validation failed", 400)
	}

	return nil
}

// ValidatePathParams validates path parameters
func (h *BaseHandler) ValidatePathParams(c *gin.Context, params map[string]string) error {
	validationErrors := make(map[string]string)

	for paramName, paramValue := range params {
		if paramValue == "" {
			validationErrors[paramName] = "required path parameter missing"
		}
	}

	if len(validationErrors) > 0 {
		h.LogWarn(c, "Path parameter validation failed", zap.Any("errors", validationErrors))
		responses.ValidationError(c, validationErrors)
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "path parameter validation failed", 400)
	}

	return nil
}

// HandleError handles errors and sends appropriate HTTP responses with enhanced logging
func (h *BaseHandler) HandleError(c *gin.Context, err error, operation string) {
	// Log the error with context
	h.LogError(c, err, "Handler error occurred",
		zap.String("operation", operation))

	// Handle domain errors
	if domain.IsDomainError(err) {
		domainErr := domain.GetDomainError(err)
		responses.DomainError(c, domainErr)
		return
	}

	// Handle different error types
	switch err {
	case domain.ErrUserNotFound, domain.ErrWorkspaceNotFound, domain.ErrVMNotFound, domain.ErrCredentialNotFound:
		responses.NotFound(c, "Resource not found")
	case domain.ErrUserAlreadyExists, domain.ErrWorkspaceExists, domain.ErrVMAlreadyExists:
		responses.Conflict(c, "Resource already exists")
	case domain.ErrInvalidCredentials:
		responses.Unauthorized(c, "Invalid credentials")
	default:
		responses.InternalServerError(c, "An unexpected error occurred")
	}
}

// LogBusinessEvent logs a business event
func (h *BaseHandler) LogBusinessEvent(c *gin.Context, eventType, userID, resourceID string, details map[string]interface{}) {
	LogBusinessEvent(c, eventType, userID, resourceID, details)
}

// LogAuditEvent logs an audit event
func (h *BaseHandler) LogAuditEvent(c *gin.Context, action, resource string, userID, resourceID string, details map[string]interface{}) {
	auditCtx := NewAuditContext(userID, action, resource, resourceID).
		WithDetails(details)
	auditCtx.LogSuccess(c, h.auditLogger)
}

// LogError logs an error with context
func (h *BaseHandler) LogError(c *gin.Context, err error, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("client_ip", c.ClientIP()),
	)
	h.logger.Error(message, allFields...)
}

// LogInfo logs an info message with context
func (h *BaseHandler) LogInfo(c *gin.Context, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("client_ip", c.ClientIP()),
	)
	h.logger.Info(message, allFields...)
}

// LogWarn logs a warning message with context
func (h *BaseHandler) LogWarn(c *gin.Context, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("client_ip", c.ClientIP()),
	)
	h.logger.Warn(message, allFields...)
}

// LogDebug logs a debug message with context
func (h *BaseHandler) LogDebug(c *gin.Context, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("client_ip", c.ClientIP()),
	)
	h.logger.Debug(message, allFields...)
}

// TrackRequest tracks request performance with automatic status code detection
func (h *BaseHandler) TrackRequest(c *gin.Context, operation string, expectedStatusCode int) {
	// Track the request with performance metrics
	h.performanceTracker.TrackRequest(c, operation, expectedStatusCode)

	// Log the operation start
	h.LogInfo(c, "Operation started",
		zap.String("operation", operation),
		zap.Int("expected_status", expectedStatusCode))
}

// TrackOperation tracks a specific operation with timing
func (h *BaseHandler) TrackOperation(c *gin.Context, operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Log performance metrics
	h.LogInfo(c, "Operation completed",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Bool("success", err == nil),
	)

	return err
}

// TrackAsyncOperation tracks an async operation
func (h *BaseHandler) TrackAsyncOperation(c *gin.Context, operation string, fn func()) {
	go func() {
		start := time.Now()
		fn()
		duration := time.Since(start)

		// Log performance metrics
		h.LogInfo(c, "Async operation completed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}()
}

// ValidateEmail validates email format
func (h *BaseHandler) ValidateEmail(email string) bool {
	return h.validationRules.ValidateEmail(email)
}

// ValidateUsername validates username format
func (h *BaseHandler) ValidateUsername(username string) bool {
	return h.validationRules.ValidateUsername(username)
}

// ValidatePassword validates password format
func (h *BaseHandler) ValidatePassword(password string) bool {
	return h.validationRules.ValidatePassword(password)
}

// ParseUUID parses a UUID from string parameter
func (h *BaseHandler) ParseUUID(c *gin.Context, paramName string) (uuid.UUID, error) {
	param := c.Param(paramName)
	id, err := uuid.Parse(param)
	if err != nil {
		responses.BadRequest(c, "Invalid ID format")
		return uuid.Nil, err
	}
	return id, nil
}

// ParsePaginationParams parses pagination parameters from query
func (h *BaseHandler) ParsePaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

// CheckPermission checks if user has permission to access resource
func (h *BaseHandler) CheckPermission(c *gin.Context, userID uuid.UUID, resourceID uuid.UUID, action string) bool {
	// This would integrate with RBAC service
	// For now, return true (implement based on your RBAC requirements)
	return true
}

// GetSpanFromContext gets telemetry span from context
func (h *BaseHandler) GetSpanFromContext(c *gin.Context) interface{} {
	return telemetry.SpanFromContext(c.Request.Context())
}

// AddSpanAttribute adds attribute to telemetry span
func (h *BaseHandler) AddSpanAttribute(c *gin.Context, key, value string) {
	// TODO: Implement span attribute addition
	// span := h.GetSpanFromContext(c)
	// This would add attribute to the span
	// Implementation depends on your telemetry setup
}

// Success sends a success response
func (h *BaseHandler) Success(c *gin.Context, statusCode int, data interface{}, message string) {
	responses.Success(c, statusCode, data, message)
}

// OK sends a 200 OK response
func (h *BaseHandler) OK(c *gin.Context, data interface{}, message string) {
	responses.OK(c, data, message)
}

// Created sends a 201 Created response
func (h *BaseHandler) Created(c *gin.Context, data interface{}, message string) {
	responses.Created(c, data, message)
}

// BadRequest sends a 400 Bad Request response
func (h *BaseHandler) BadRequest(c *gin.Context, message string) {
	responses.BadRequest(c, message)
}

// Unauthorized sends a 401 Unauthorized response
func (h *BaseHandler) Unauthorized(c *gin.Context, message string) {
	responses.Unauthorized(c, message)
}

// Forbidden sends a 403 Forbidden response
func (h *BaseHandler) Forbidden(c *gin.Context, message string) {
	responses.Forbidden(c, message)
}

// NotFound sends a 404 Not Found response
func (h *BaseHandler) NotFound(c *gin.Context, message string) {
	responses.NotFound(c, message)
}

// Conflict sends a 409 Conflict response
func (h *BaseHandler) Conflict(c *gin.Context, message string) {
	responses.Conflict(c, message)
}

// InternalServerError sends a 500 Internal Server Error response
func (h *BaseHandler) InternalServerError(c *gin.Context, message string) {
	responses.InternalServerError(c, message)
}

// DomainError sends a domain error response
func (h *BaseHandler) DomainError(c *gin.Context, err *domain.DomainError) {
	responses.DomainError(c, err)
}

// ValidationError sends a validation error response
func (h *BaseHandler) ValidationError(c *gin.Context, errors map[string]string) {
	responses.ValidationError(c, errors)
}

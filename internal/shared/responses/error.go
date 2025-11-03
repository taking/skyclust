package responses

import (
	"encoding/json"
	"net/http"
	"skyclust/internal/domain"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorBuilder provides methods for building error responses
type ErrorBuilder struct {
	logger *zap.Logger
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(logger *zap.Logger) *ErrorBuilder {
	return &ErrorBuilder{
		logger: logger,
	}
}

// BadRequest sends a 400 Bad Request response
func (eb *ErrorBuilder) BadRequest(c *gin.Context, message string) {
	eb.Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 Unauthorized response
func (eb *ErrorBuilder) Unauthorized(c *gin.Context, message string) {
	eb.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden response
func (eb *ErrorBuilder) Forbidden(c *gin.Context, message string) {
	eb.Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found response
func (eb *ErrorBuilder) NotFound(c *gin.Context, message string) {
	eb.Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict response
func (eb *ErrorBuilder) Conflict(c *gin.Context, message string) {
	eb.Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalServerError sends a 500 Internal Server Error response
func (eb *ErrorBuilder) InternalServerError(c *gin.Context, message string) {
	eb.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// Error sends an error response
func (eb *ErrorBuilder) Error(c *gin.Context, statusCode int, code, message string) {
	response := APIResponse{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// DomainError sends a domain error response
func (eb *ErrorBuilder) DomainError(c *gin.Context, err *domain.DomainError) {
	response := APIResponse{
		Success: false,
		Error: &Error{
			Code:    string(err.Code),
			Message: err.Message,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	}
	if err.Details != nil {
		response.Data = err.Details
	}
	c.JSON(err.StatusCode, response)
}

// ValidationError sends a validation error response
func (eb *ErrorBuilder) ValidationError(c *gin.Context, errors map[string]string) {
	response := APIResponse{
		Success: false,
		Error: &Error{
			Code:    string(domain.ErrCodeValidationFailed),
			Message: "Validation failed",
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
		Data:      errors,
	}
	c.JSON(http.StatusBadRequest, response)
}

// Global convenience functions for error handling
func BadRequest(c *gin.Context, message string) {
	NewErrorBuilder(nil).BadRequest(c, message)
}

func Unauthorized(c *gin.Context, message string) {
	NewErrorBuilder(nil).Unauthorized(c, message)
}

func Forbidden(c *gin.Context, message string) {
	NewErrorBuilder(nil).Forbidden(c, message)
}

func NotFound(c *gin.Context, message string) {
	NewErrorBuilder(nil).NotFound(c, message)
}

func Conflict(c *gin.Context, message string) {
	NewErrorBuilder(nil).Conflict(c, message)
}

func InternalServerError(c *gin.Context, message string) {
	NewErrorBuilder(nil).InternalServerError(c, message)
}

func DomainError(c *gin.Context, err *domain.DomainError) {
	NewErrorBuilder(nil).DomainError(c, err)
}

func ValidationError(c *gin.Context, errors map[string]string) {
	NewErrorBuilder(nil).ValidationError(c, errors)
}

// Handle handles errors and sends appropriate HTTP responses
func (eb *ErrorBuilder) Handle(c *gin.Context, err error) {
	// Extract request ID
	requestID := getRequestID(c)

	// Log the error
	eb.logger.Error("HTTP error occurred",
		zap.Error(err),
		zap.String("request_id", requestID),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	// Handle domain errors
	if domain.IsDomainError(err) {
		domainErr := domain.GetDomainError(err)
		eb.DomainError(c, domainErr)
		return
	}

	// Handle different error types
	switch err := err.(type) {
	case *json.SyntaxError:
		eb.BadRequest(c, "Invalid JSON syntax")
	case *json.UnmarshalTypeError:
		eb.BadRequest(c, "Invalid JSON type")
	default:
		// Check if it's a validation error
		if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "required") {
			eb.BadRequest(c, err.Error())
			return
		}

		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") {
			eb.NotFound(c, err.Error())
			return
		}

		// Default to internal server error
		eb.InternalServerError(c, "An unexpected error occurred")
	}
}

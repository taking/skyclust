package http

import (
	"skyclust/internal/domain"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

// Meta represents pagination and metadata information
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasNext    bool  `json:"has_next,omitempty"`
	HasPrev    bool  `json:"has_prev,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	}

	// Set response headers
	c.Header("X-Request-ID", response.RequestID)
	c.JSON(statusCode, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, error string, code string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   error,
		Code:    code,
	})
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, data interface{}, message string) {
	SuccessResponse(c, http.StatusCreated, data, message)
}

// OKResponse sends a 200 OK response
func OKResponse(c *gin.Context, data interface{}, message string) {
	SuccessResponse(c, http.StatusOK, data, message)
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusBadRequest, error, "BAD_REQUEST")
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusUnauthorized, error, "UNAUTHORIZED")
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusNotFound, error, "NOT_FOUND")
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusInternalServerError, error, "INTERNAL_SERVER_ERROR")
}

// ValidationErrorResponse sends a 422 Validation Error response
func ValidationErrorResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusUnprocessableEntity, error, "VALIDATION_ERROR")
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusForbidden, error, "FORBIDDEN")
}

// ConflictResponse sends a 409 Conflict response
func ConflictResponse(c *gin.Context, error string) {
	ErrorResponse(c, http.StatusConflict, error, "CONFLICT")
}

// DomainErrorResponse sends a response based on domain error
func DomainErrorResponse(c *gin.Context, domainErr *domain.DomainError) {
	response := APIResponse{
		Success:   false,
		Error:     domainErr.Message,
		Code:      string(domainErr.Code),
		RequestID: domainErr.RequestID,
		Timestamp: domainErr.Timestamp,
	}

	// Add details if available (but filter sensitive information)
	if len(domainErr.Details) > 0 {
		filteredDetails := filterSensitiveDetails(domainErr.Details)
		if len(filteredDetails) > 0 {
			response.Meta = &Meta{}
			// Store details in meta for now (could be separate field)
		}
	}

	// Set response headers
	c.Header("X-Request-ID", response.RequestID)
	c.JSON(domainErr.StatusCode, response)
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, data interface{}, page, limit int, total int64, message string) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	// Set response headers
	c.Header("X-Request-ID", response.RequestID)
	c.JSON(http.StatusOK, response)
}

// filterSensitiveDetails removes sensitive information from error details
func filterSensitiveDetails(details map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	sensitiveKeys := []string{
		"password", "token", "secret", "key", "credential",
		"authorization", "auth", "private", "sensitive",
	}

	for key, value := range details {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			filtered[key] = value
		} else {
			filtered[key] = "[REDACTED]"
		}
	}

	return filtered
}

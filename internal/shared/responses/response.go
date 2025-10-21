package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResponseBuilder provides a fluent interface for building responses
type ResponseBuilder struct {
	response *APIResponse
	context  *gin.Context
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(c *gin.Context) *ResponseBuilder {
	return &ResponseBuilder{
		response: &APIResponse{
			Success:   true,
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
		},
		context: c,
	}
}

// WithData sets the response data
func (rb *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	return rb
}

// WithMessage sets the response message
func (rb *ResponseBuilder) WithMessage(message string) *ResponseBuilder {
	rb.response.Message = message
	return rb
}

// WithError sets the response error
func (rb *ResponseBuilder) WithError(code, message string) *ResponseBuilder {
	rb.response.Success = false
	rb.response.Error = &Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		RequestID: rb.response.RequestID,
	}
	return rb
}

// WithFieldError sets a field-specific error
func (rb *ResponseBuilder) WithFieldError(code, message, field string, value interface{}) *ResponseBuilder {
	rb.response.Success = false
	rb.response.Error = &Error{
		Code:      code,
		Message:   message,
		Field:     field,
		Value:     value,
		Timestamp: time.Now(),
		RequestID: rb.response.RequestID,
	}
	return rb
}

// WithDetails adds error details
func (rb *ResponseBuilder) WithDetails(key string, value interface{}) *ResponseBuilder {
	if rb.response.Error != nil {
		if rb.response.Error.Details == nil {
			rb.response.Error.Details = make(map[string]interface{})
		}
		rb.response.Error.Details[key] = value
	}
	return rb
}

// WithPagination sets pagination metadata
func (rb *ResponseBuilder) WithPagination(page, limit int, total int64) *ResponseBuilder {
	rb.response.Meta = &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
		HasNext:    page < int((total+int64(limit)-1)/int64(limit)),
		HasPrev:    page > 1,
	}
	return rb
}

// Send sends the response with the specified status code
func (rb *ResponseBuilder) Send(statusCode int) {
	rb.context.Header("X-Request-ID", rb.response.RequestID)
	rb.context.Header("X-Response-Time", time.Now().Format(time.RFC3339))
	rb.context.JSON(statusCode, rb.response)
}

// SendOK sends a 200 OK response
func (rb *ResponseBuilder) SendOK() {
	rb.Send(http.StatusOK)
}

// SendCreated sends a 201 Created response
func (rb *ResponseBuilder) SendCreated() {
	rb.Send(http.StatusCreated)
}

// SendBadRequest sends a 400 Bad Request response
func (rb *ResponseBuilder) SendBadRequest() {
	rb.Send(http.StatusBadRequest)
}

// SendUnauthorized sends a 401 Unauthorized response
func (rb *ResponseBuilder) SendUnauthorized() {
	rb.Send(http.StatusUnauthorized)
}

// SendForbidden sends a 403 Forbidden response
func (rb *ResponseBuilder) SendForbidden() {
	rb.Send(http.StatusForbidden)
}

// SendNotFound sends a 404 Not Found response
func (rb *ResponseBuilder) SendNotFound() {
	rb.Send(http.StatusNotFound)
}

// SendConflict sends a 409 Conflict response
func (rb *ResponseBuilder) SendConflict() {
	rb.Send(http.StatusConflict)
}

// SendInternalServerError sends a 500 Internal Server Error response
func (rb *ResponseBuilder) SendInternalServerError() {
	rb.Send(http.StatusInternalServerError)
}

// getRequestID extracts or generates a request ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}

// Global convenience functions
func Success(c *gin.Context, statusCode int, data interface{}, message string) {
	NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		Send(statusCode)
}

func OK(c *gin.Context, data interface{}, message string) {
	NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		SendOK()
}

func Created(c *gin.Context, data interface{}, message string) {
	NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		SendCreated()
}

func Paginated(c *gin.Context, data interface{}, message string, meta *Meta) {
	NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		SendOK()
}

package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResponseBuilder provides methods for building success responses
type ResponseBuilder struct{}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

// Success sends a success response
func (rb *ResponseBuilder) Success(c *gin.Context, statusCode int, data interface{}, message string) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// OK sends a 200 OK response
func (rb *ResponseBuilder) OK(c *gin.Context, data interface{}, message string) {
	rb.Success(c, http.StatusOK, data, message)
}

// Created sends a 201 Created response
func (rb *ResponseBuilder) Created(c *gin.Context, data interface{}, message string) {
	rb.Success(c, http.StatusCreated, data, message)
}

// Paginated sends a paginated response
func (rb *ResponseBuilder) Paginated(c *gin.Context, data interface{}, message string, meta *Meta) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Meta:      meta,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
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
	NewResponseBuilder().Success(c, statusCode, data, message)
}

func OK(c *gin.Context, data interface{}, message string) {
	NewResponseBuilder().OK(c, data, message)
}

func Created(c *gin.Context, data interface{}, message string) {
	NewResponseBuilder().Created(c, data, message)
}

func Paginated(c *gin.Context, data interface{}, message string, meta *Meta) {
	NewResponseBuilder().Paginated(c, data, message, meta)
}

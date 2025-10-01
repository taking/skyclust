package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// SuccessResponse sends a successful response
func (h *Handler) SuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ErrorResponse sends an error response
func (h *Handler) ErrorResponse(c *gin.Context, statusCode int, error string, code string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   error,
		Code:    code,
	})
}

// CreatedResponse sends a 201 Created response
func (h *Handler) CreatedResponse(c *gin.Context, data interface{}, message string) {
	h.SuccessResponse(c, http.StatusCreated, data, message)
}

// OKResponse sends a 200 OK response
func (h *Handler) OKResponse(c *gin.Context, data interface{}, message string) {
	h.SuccessResponse(c, http.StatusOK, data, message)
}

// BadRequestResponse sends a 400 Bad Request response
func (h *Handler) BadRequestResponse(c *gin.Context, error string) {
	h.ErrorResponse(c, http.StatusBadRequest, error, "BAD_REQUEST")
}

// UnauthorizedResponse sends a 401 Unauthorized response
func (h *Handler) UnauthorizedResponse(c *gin.Context, error string) {
	h.ErrorResponse(c, http.StatusUnauthorized, error, "UNAUTHORIZED")
}

// NotFoundResponse sends a 404 Not Found response
func (h *Handler) NotFoundResponse(c *gin.Context, error string) {
	h.ErrorResponse(c, http.StatusNotFound, error, "NOT_FOUND")
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func (h *Handler) InternalServerErrorResponse(c *gin.Context, error string) {
	h.ErrorResponse(c, http.StatusInternalServerError, error, "INTERNAL_SERVER_ERROR")
}

// ValidationErrorResponse sends a 422 Validation Error response
func (h *Handler) ValidationErrorResponse(c *gin.Context, error string) {
	h.ErrorResponse(c, http.StatusUnprocessableEntity, error, "VALIDATION_ERROR")
}

package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response
// 표준화된 API 응답 형식
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *MetaInfo   `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo represents error information in API response
// API 응답의 에러 정보
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaInfo represents metadata in API response (pagination, etc.)
// API 응답의 메타데이터 (페이지네이션 등)
type MetaInfo struct {
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Version    string           `json:"version,omitempty"`
}

// PaginationInfo represents pagination information
// 페이지네이션 정보
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse sends a successful response
// 성공 응답 전송
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// SuccessResponseWithMeta sends a successful response with metadata
// 메타데이터가 포함된 성공 응답 전송
func SuccessResponseWithMeta(c *gin.Context, data interface{}, meta *MetaInfo) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// CreatedResponse sends a 201 Created response
// 201 Created 응답 전송
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// ErrorResponse sends an error response
// 에러 응답 전송
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// BadRequestResponse sends a 400 Bad Request response
// 400 Bad Request 응답 전송
func BadRequestResponse(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// UnauthorizedResponse sends a 401 Unauthorized response
// 401 Unauthorized 응답 전송
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// ForbiddenResponse sends a 403 Forbidden response
// 403 Forbidden 응답 전송
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFoundResponse sends a 404 Not Found response
// 404 Not Found 응답 전송
func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", resource+" not found", nil)
}

// ConflictResponse sends a 409 Conflict response
// 409 Conflict 응답 전송
func ConflictResponse(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, http.StatusConflict, "CONFLICT", message, details)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
// 500 Internal Server Error 응답 전송
func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, nil)
}

// PaginatedResponse sends a paginated response
// 페이지네이션된 응답 전송
func PaginatedResponse(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	SuccessResponseWithMeta(c, data, &MetaInfo{
		Pagination: &PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}


package handlers

import "github.com/gin-gonic/gin"

// HTTPHandler defines the interface for HTTP handlers
type HTTPHandler interface {
	// HTTP methods
	GET(path string, handlers ...gin.HandlerFunc)
	POST(path string, handlers ...gin.HandlerFunc)
	PUT(path string, handlers ...gin.HandlerFunc)
	DELETE(path string, handlers ...gin.HandlerFunc)
	PATCH(path string, handlers ...gin.HandlerFunc)

	// Route groups
	Group(relativePath string) *gin.RouterGroup
}

// BaseHandler defines the interface for common handler functionality
type BaseHandler interface {
	// Request handling
	ValidateRequest(c *gin.Context, req interface{}) error
	ValidateQueryParams(c *gin.Context, params map[string]string) error
	ValidatePathParams(c *gin.Context, params map[string]string) error

	// Authentication
	GetUserIDFromToken(c *gin.Context) (string, error)
	GetUserRoleFromToken(c *gin.Context) (string, error)
	GetBearerTokenFromHeader(c *gin.Context) (string, error)

	// Response handling
	Success(c *gin.Context, statusCode int, data interface{}, message string)
	Error(c *gin.Context, statusCode int, message string)
	ValidationError(c *gin.Context, errors map[string]string)

	// Logging
	LogInfo(c *gin.Context, message string, fields ...interface{})
	LogError(c *gin.Context, err error, message string)
	LogWarn(c *gin.Context, message string, fields ...interface{})
	LogDebug(c *gin.Context, message string, fields ...interface{})
}

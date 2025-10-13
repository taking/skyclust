package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents a standardized API response
type Response struct {
	Success   bool           `json:"success"`
	Data      interface{}    `json:"data,omitempty"`
	Message   string         `json:"message,omitempty"`
	Error     *Error         `json:"error,omitempty"`
	Meta      *ResponseMeta  `json:"meta,omitempty"`
	Links     *ResponseLinks `json:"links,omitempty"`
	Debug     *ResponseDebug `json:"debug,omitempty"`
	Cache     *ResponseCache `json:"cache,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
	Version   string         `json:"version,omitempty"`
}

// Error represents a unified error structure
type Error struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// ResponseMeta represents response metadata
type ResponseMeta struct {
	Pagination *PaginationMeta `json:"pagination,omitempty"`
	Sorting    *SortingMeta    `json:"sorting,omitempty"`
	Filtering  *FilteringMeta  `json:"filtering,omitempty"`
	Total      int64           `json:"total,omitempty"`
	Count      int             `json:"count,omitempty"`
	Duration   int64           `json:"duration,omitempty"` // milliseconds
	RateLimit  *RateLimitMeta  `json:"rate_limit,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
}

// SortingMeta represents sorting metadata
type SortingMeta struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// FilteringMeta represents filtering metadata
type FilteringMeta struct {
	Applied []FilterMeta `json:"applied"`
	Total   int          `json:"total"`
}

// FilterMeta represents a single filter
type FilterMeta struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// RateLimitMeta represents rate limiting metadata
type RateLimitMeta struct {
	Limit      int   `json:"limit"`
	Remaining  int   `json:"remaining"`
	Reset      int64 `json:"reset"`
	RetryAfter int   `json:"retry_after,omitempty"`
}

// ResponseLinks represents HATEOAS links
type ResponseLinks struct {
	Self    *Link  `json:"self,omitempty"`
	First   *Link  `json:"first,omitempty"`
	Last    *Link  `json:"last,omitempty"`
	Next    *Link  `json:"next,omitempty"`
	Prev    *Link  `json:"prev,omitempty"`
	Related []Link `json:"related,omitempty"`
}

// Link represents a HATEOAS link
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// ResponseDebug represents debug information
type ResponseDebug struct {
	QueryTime    int64                  `json:"query_time,omitempty"` // milliseconds
	CacheHit     bool                   `json:"cache_hit,omitempty"`
	CacheKey     string                 `json:"cache_key,omitempty"`
	DatabaseTime int64                  `json:"database_time,omitempty"` // milliseconds
	ExternalAPIs []ExternalAPIMeta      `json:"external_apis,omitempty"`
	MemoryUsage  *MemoryUsageMeta       `json:"memory_usage,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// ExternalAPIMeta represents external API call metadata
type ExternalAPIMeta struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	Method  string `json:"method"`
	Status  int    `json:"status"`
	Time    int64  `json:"time"` // milliseconds
	Success bool   `json:"success"`
}

// MemoryUsageMeta represents memory usage metadata
type MemoryUsageMeta struct {
	Allocated int64 `json:"allocated"` // bytes
	Used      int64 `json:"used"`      // bytes
	Free      int64 `json:"free"`      // bytes
}

// ResponseCache represents cache metadata
type ResponseCache struct {
	Hit       bool      `json:"hit"`
	Key       string    `json:"key,omitempty"`
	TTL       int64     `json:"ttl,omitempty"` // seconds
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Strategy  string    `json:"strategy,omitempty"` // memory, redis, etc.
}

// ResponseBuilder provides a fluent interface for building responses
type ResponseBuilder struct {
	response *Response
	context  *gin.Context
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(c *gin.Context) *ResponseBuilder {
	return &ResponseBuilder{
		response: &Response{
			Success:   true,
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
			Version:   c.GetString("api_version"),
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

// WithSuggestions adds error suggestions
func (rb *ResponseBuilder) WithSuggestions(suggestions []string) *ResponseBuilder {
	if rb.response.Error != nil {
		rb.response.Error.Suggestions = suggestions
	}
	return rb
}

// WithPagination sets pagination metadata
func (rb *ResponseBuilder) WithPagination(page, limit int, total int64) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	rb.response.Meta.Pagination = &PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	if page < totalPages {
		nextPage := page + 1
		rb.response.Meta.Pagination.NextPage = &nextPage
	}
	if page > 1 {
		prevPage := page - 1
		rb.response.Meta.Pagination.PrevPage = &prevPage
	}

	return rb
}

// WithSorting sets sorting metadata
func (rb *ResponseBuilder) WithSorting(field, order string) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.Sorting = &SortingMeta{
		Field: field,
		Order: order,
	}
	return rb
}

// WithFiltering sets filtering metadata
func (rb *ResponseBuilder) WithFiltering(filters []FilterMeta) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.Filtering = &FilteringMeta{
		Applied: filters,
		Total:   len(filters),
	}
	return rb
}

// WithTotal sets the total count
func (rb *ResponseBuilder) WithTotal(total int64) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.Total = total
	return rb
}

// WithCount sets the current count
func (rb *ResponseBuilder) WithCount(count int) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.Count = count
	return rb
}

// WithDuration sets the response duration
func (rb *ResponseBuilder) WithDuration(duration time.Duration) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.Duration = duration.Milliseconds()
	return rb
}

// WithRateLimit sets rate limiting metadata
func (rb *ResponseBuilder) WithRateLimit(limit, remaining int, reset time.Time) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &ResponseMeta{}
	}
	rb.response.Meta.RateLimit = &RateLimitMeta{
		Limit:     limit,
		Remaining: remaining,
		Reset:     reset.Unix(),
	}
	return rb
}

// WithLinks sets HATEOAS links
func (rb *ResponseBuilder) WithLinks(links *ResponseLinks) *ResponseBuilder {
	rb.response.Links = links
	return rb
}

// WithSelfLink sets the self link
func (rb *ResponseBuilder) WithSelfLink(href string) *ResponseBuilder {
	if rb.response.Links == nil {
		rb.response.Links = &ResponseLinks{}
	}
	rb.response.Links.Self = &Link{
		Href: href,
		Rel:  "self",
	}
	return rb
}

// WithNextLink sets the next page link
func (rb *ResponseBuilder) WithNextLink(href string) *ResponseBuilder {
	if rb.response.Links == nil {
		rb.response.Links = &ResponseLinks{}
	}
	rb.response.Links.Next = &Link{
		Href: href,
		Rel:  "next",
	}
	return rb
}

// WithPrevLink sets the previous page link
func (rb *ResponseBuilder) WithPrevLink(href string) *ResponseBuilder {
	if rb.response.Links == nil {
		rb.response.Links = &ResponseLinks{}
	}
	rb.response.Links.Prev = &Link{
		Href: href,
		Rel:  "prev",
	}
	return rb
}

// WithDebug sets debug information
func (rb *ResponseBuilder) WithDebug(debug *ResponseDebug) *ResponseBuilder {
	rb.response.Debug = debug
	return rb
}

// WithCache sets cache metadata
func (rb *ResponseBuilder) WithCache(cache *ResponseCache) *ResponseBuilder {
	rb.response.Cache = cache
	return rb
}

// WithTraceInfo sets trace information
func (rb *ResponseBuilder) WithTraceInfo(traceID, spanID string) *ResponseBuilder {
	if rb.response.Error != nil {
		rb.response.Error.TraceID = traceID
		rb.response.Error.SpanID = spanID
	}
	return rb
}

// Send sends the response with the specified status code
func (rb *ResponseBuilder) Send(statusCode int) {
	// Set response headers
	rb.context.Header("X-Request-ID", rb.response.RequestID)
	rb.context.Header("X-API-Version", rb.response.Version)
	rb.context.Header("X-Response-Time", time.Now().Format(time.RFC3339))

	// Set rate limit headers if available
	if rb.response.Meta != nil && rb.response.Meta.RateLimit != nil {
		rb.context.Header("X-RateLimit-Limit", strconv.Itoa(rb.response.Meta.RateLimit.Limit))
		rb.context.Header("X-RateLimit-Remaining", strconv.Itoa(rb.response.Meta.RateLimit.Remaining))
		rb.context.Header("X-RateLimit-Reset", strconv.FormatInt(rb.response.Meta.RateLimit.Reset, 10))
	}

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

// SendAccepted sends a 202 Accepted response
func (rb *ResponseBuilder) SendAccepted() {
	rb.Send(http.StatusAccepted)
}

// SendNoContent sends a 204 No Content response
func (rb *ResponseBuilder) SendNoContent() {
	rb.Send(http.StatusNoContent)
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

// SendUnprocessableEntity sends a 422 Unprocessable Entity response
func (rb *ResponseBuilder) SendUnprocessableEntity() {
	rb.Send(http.StatusUnprocessableEntity)
}

// SendTooManyRequests sends a 429 Too Many Requests response
func (rb *ResponseBuilder) SendTooManyRequests() {
	rb.Send(http.StatusTooManyRequests)
}

// SendInternalServerError sends a 500 Internal Server Error response
func (rb *ResponseBuilder) SendInternalServerError() {
	rb.Send(http.StatusInternalServerError)
}

// SendServiceUnavailable sends a 503 Service Unavailable response
func (rb *ResponseBuilder) SendServiceUnavailable() {
	rb.Send(http.StatusServiceUnavailable)
}

// Helper functions for common response patterns

// Success creates a success response builder
func Success(c *gin.Context) *ResponseBuilder {
	return NewResponseBuilder(c)
}

// ErrorResponse creates an error response builder
func ErrorResponse(c *gin.Context) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("UNKNOWN_ERROR", "An unknown error occurred")
}

// ValidationError creates a validation error response
func ValidationError(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("VALIDATION_ERROR", message)
}

// FieldValidationError creates a field validation error response
func FieldValidationError(c *gin.Context, field, message string, value interface{}) *ResponseBuilder {
	return NewResponseBuilder(c).WithFieldError("VALIDATION_ERROR", message, field, value)
}

// NotFound creates a not found error response
func NotFound(c *gin.Context, resource string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("NOT_FOUND", fmt.Sprintf("%s not found", resource))
}

// Unauthorized creates an unauthorized error response
func Unauthorized(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("UNAUTHORIZED", message)
}

// Forbidden creates a forbidden error response
func Forbidden(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("FORBIDDEN", message)
}

// Conflict creates a conflict error response
func Conflict(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("CONFLICT", message)
}

// InternalError creates an internal server error response
func InternalError(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("INTERNAL_ERROR", message)
}

// RateLimitExceeded creates a rate limit exceeded error response
func RateLimitExceeded(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).WithError("RATE_LIMIT_EXCEEDED", message)
}

// Paginated creates a paginated response
func Paginated(c *gin.Context, data interface{}, page, limit int, total int64, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		WithPagination(page, limit, total).
		WithTotal(total).
		WithCount(len(data.([]interface{})))
}

// List creates a list response
func List(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		WithCount(len(data.([]interface{})))
}

// Item creates a single item response
func Item(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message)
}

// Created creates a created response
func Created(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message)
}

// Updated creates an updated response
func Updated(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message)
}

// Deleted creates a deleted response
func Deleted(c *gin.Context, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithMessage(message)
}

// Health creates a health check response
func Health(c *gin.Context, status string, details map[string]interface{}) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(map[string]interface{}{
			"status":  status,
			"details": details,
		}).
		WithMessage("Health check completed")
}

// Info creates an info response
func Info(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message)
}

// Stats creates a statistics response
func Stats(c *gin.Context, data interface{}, message string) *ResponseBuilder {
	return NewResponseBuilder(c).
		WithData(data).
		WithMessage(message)
}

// ErrorResponseFromError creates an error response from a unified error
func ErrorResponseFromError(c *gin.Context, err *Error) *ResponseBuilder {
	builder := NewResponseBuilder(c)
	if err != nil {
		builder.response.Success = false
		builder.response.Error = err
	}
	return builder
}

// FromDomainError creates a response from a domain error
func FromDomainError(c *gin.Context, domainErr interface{}) *ResponseBuilder {
	// This would need to be implemented based on your domain error structure
	// For now, we'll create a generic error response
	return NewResponseBuilder(c).WithError("DOMAIN_ERROR", "A domain error occurred")
}

// ToJSON converts the response to JSON
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSONPretty converts the response to pretty-printed JSON
func (r *Response) ToJSONPretty() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// Legacy compatibility functions for existing code

// SuccessResponse sends a successful response (legacy compatibility)
func SuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	NewResponseBuilder(c).
		WithData(data).
		WithMessage(message).
		Send(statusCode)
}

// ErrorResponse sends an error response (legacy compatibility)
func ErrorResponseLegacy(c *gin.Context, statusCode int, error string, code string) {
	NewResponseBuilder(c).
		WithError(code, error).
		Send(statusCode)
}

// CreatedResponse sends a 201 Created response (legacy compatibility)
func CreatedResponse(c *gin.Context, data interface{}, message string) {
	Created(c, data, message).SendCreated()
}

// OKResponse sends a 200 OK response (legacy compatibility)
func OKResponse(c *gin.Context, data interface{}, message string) {
	Item(c, data, message).SendOK()
}

// BadRequestResponse sends a 400 Bad Request response (legacy compatibility)
func BadRequestResponse(c *gin.Context, error string) {
	NewResponseBuilder(c).WithError("BAD_REQUEST", error).SendBadRequest()
}

// UnauthorizedResponse sends a 401 Unauthorized response (legacy compatibility)
func UnauthorizedResponse(c *gin.Context, error string) {
	Unauthorized(c, error).SendUnauthorized()
}

// NotFoundResponse sends a 404 Not Found response (legacy compatibility)
func NotFoundResponse(c *gin.Context, error string) {
	NewResponseBuilder(c).WithError("NOT_FOUND", error).SendNotFound()
}

// InternalServerErrorResponse sends a 500 Internal Server Error response (legacy compatibility)
func InternalServerErrorResponse(c *gin.Context, error string) {
	InternalError(c, error).SendInternalServerError()
}

// ValidationErrorResponse sends a 422 Validation Error response (legacy compatibility)
func ValidationErrorResponse(c *gin.Context, error string) {
	ValidationError(c, error).SendUnprocessableEntity()
}

// ForbiddenResponse sends a 403 Forbidden response (legacy compatibility)
func ForbiddenResponse(c *gin.Context, error string) {
	Forbidden(c, error).SendForbidden()
}

// ConflictResponse sends a 409 Conflict response (legacy compatibility)
func ConflictResponse(c *gin.Context, error string) {
	Conflict(c, error).SendConflict()
}

// PaginatedResponse sends a paginated response (legacy compatibility)
func PaginatedResponse(c *gin.Context, data interface{}, page, limit int, total int64, message string) {
	Paginated(c, data, page, limit, total, message).SendOK()
}

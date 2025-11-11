package handlers

import (
	"strconv"
	"strings"
	"time"

	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StandardQueryParams represents standardized query parameters for list endpoints
type StandardQueryParams struct {
	// Pagination
	Page  int
	Limit int

	// Sorting
	SortBy    string
	SortOrder string // "asc" or "desc"

	// Filtering
	Search  string
	Filters map[string]interface{}
}

// ParseStandardQueryParams parses standardized query parameters from request
func (h *BaseHandler) ParseStandardQueryParams(c *gin.Context) *StandardQueryParams {
	params := &StandardQueryParams{
		Filters: make(map[string]interface{}),
	}

	// Parse pagination (page/limit)
	params.Page, params.Limit = h.ParsePageLimitParams(c)

	// Parse sorting
	params.SortBy = c.Query("sort_by")
	params.SortOrder = strings.ToLower(c.DefaultQuery("sort_order", "asc"))
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "asc"
	}

	// Parse search
	params.Search = c.Query("search")

	// Parse additional filters from query parameters
	// Common filter patterns: filter_<field>=<value>
	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "filter_") {
			field := strings.TrimPrefix(key, "filter_")
			if len(values) > 0 {
				params.Filters[field] = values[0]
			}
		}
	}

	return params
}

// ParseFilterParams parses filter parameters from query string
// Supports: filter_<field>=<value> pattern
func (h *BaseHandler) ParseFilterParams(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})

	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "filter_") {
			field := strings.TrimPrefix(key, "filter_")
			if len(values) > 0 {
				filters[field] = values[0]
			}
		}
	}

	return filters
}

// ParseSortParams parses sort parameters from query string
// Returns sort_by field and sort_order (asc/desc)
func (h *BaseHandler) ParseSortParams(c *gin.Context) (sortBy string, sortOrder string) {
	sortBy = c.Query("sort_by")
	sortOrder = strings.ToLower(c.DefaultQuery("sort_order", "asc"))

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	return sortBy, sortOrder
}

// ValidateSortParams validates sort parameters against allowed fields
func (h *BaseHandler) ValidateSortParams(sortBy string, allowedFields []string) bool {
	if sortBy == "" {
		return true // No sorting is valid
	}

	for _, field := range allowedFields {
		if field == sortBy {
			return true
		}
	}

	return false
}

// ParseUUIDFilter parses a UUID filter parameter from query string
func (h *BaseHandler) ParseUUIDFilter(c *gin.Context, paramName string) (*uuid.UUID, error) {
	param := c.Query(paramName)
	if param == "" {
		return nil, nil // Not provided is valid
	}

	id, err := uuid.Parse(param)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "Invalid "+paramName+" format", 400)
	}

	return &id, nil
}

// ParseIntFilter parses an integer filter parameter from query string
func (h *BaseHandler) ParseIntFilter(c *gin.Context, paramName string) (*int, error) {
	param := c.Query(paramName)
	if param == "" {
		return nil, nil // Not provided is valid
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "Invalid "+paramName+" format", 400)
	}

	return &value, nil
}

// ParseBoolFilter parses a boolean filter parameter from query string
func (h *BaseHandler) ParseBoolFilter(c *gin.Context, paramName string) (*bool, error) {
	param := c.Query(paramName)
	if param == "" {
		return nil, nil // Not provided is valid
	}

	value := strings.ToLower(param) == "true" || param == "1"
	return &value, nil
}

// ParseTimeFilter parses a time filter parameter from query string (RFC3339 format)
func (h *BaseHandler) ParseTimeFilter(c *gin.Context, paramName string) (*time.Time, error) {
	param := c.Query(paramName)
	if param == "" {
		return nil, nil // Not provided is valid
	}

	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, param)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "Invalid "+paramName+" format, expected RFC3339", 400)
	}

	return &t, nil
}

// BuildPaginatedResponse builds a standardized paginated response
func (h *BaseHandler) BuildPaginatedResponse(c *gin.Context, data interface{}, page, limit int, total int64, message string) {
	h.OKWithPagination(c, data, message, page, limit, total)
}

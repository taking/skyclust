package handlers

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidationRules defines validation rules for common fields
type ValidationRules struct {
	Email    *regexp.Regexp
	Username *regexp.Regexp
	Password *regexp.Regexp
}

// NewValidationRules creates a new validation rules instance
func NewValidationRules() *ValidationRules {
	return &ValidationRules{
		Email:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		Username: regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`),
		Password: regexp.MustCompile(`^.{8,}$`),
	}
}

// ValidateEmail validates email format
func (vr *ValidationRules) ValidateEmail(email string) bool {
	return vr.Email.MatchString(email)
}

// ValidateUsername validates username format
func (vr *ValidationRules) ValidateUsername(username string) bool {
	return vr.Username.MatchString(username)
}

// ValidatePassword validates password strength
func (vr *ValidationRules) ValidatePassword(password string) bool {
	return vr.Password.MatchString(password)
}

// ValidateRequired validates required fields
func ValidateRequired(fields map[string]string) map[string]string {
	errors := make(map[string]string)

	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			errors[field] = "This field is required"
		}
	}

	return errors
}

// ValidatePagination validates pagination parameters
func ValidatePagination(c *gin.Context) (limit, offset int, errors map[string]string) {
	errors = make(map[string]string)

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		errors["limit"] = "Limit must be a positive integer"
		limit = 10
	} else if limit > 100 {
		errors["limit"] = "Limit cannot exceed 100"
		limit = 100
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		errors["offset"] = "Offset must be a non-negative integer"
		offset = 0
	}

	return limit, offset, errors
}

// ValidateSort validates sort parameters
func ValidateSort(c *gin.Context, allowedFields []string) (sortBy, sortOrder string, errors map[string]string) {
	errors = make(map[string]string)

	sortBy = c.DefaultQuery("sort_by", "")
	sortOrder = c.DefaultQuery("sort_order", "asc")

	if sortBy != "" {
		validField := false
		for _, field := range allowedFields {
			if field == sortBy {
				validField = true
				break
			}
		}
		if !validField {
			errors["sort_by"] = "Invalid sort field"
		}
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		errors["sort_order"] = "Sort order must be 'asc' or 'desc'"
		sortOrder = "asc"
	}

	return sortBy, sortOrder, errors
}

// ValidateUUID validates UUID format
func ValidateUUID(uuidStr string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuidStr))
}

// ValidateSearch validates search query
func ValidateSearch(query string) bool {
	// Remove extra whitespace and check length
	trimmed := strings.TrimSpace(query)
	return len(trimmed) >= 2 && len(trimmed) <= 100
}

// ValidateFilters validates filter parameters
func ValidateFilters(c *gin.Context, allowedFilters []string) (filters map[string]interface{}, errors map[string]string) {
	errors = make(map[string]string)
	filters = make(map[string]interface{})

	for _, filter := range allowedFilters {
		if value := c.Query(filter); value != "" {
			// Basic validation - can be extended based on filter type
			if len(value) > 200 {
				errors[filter] = "Filter value too long"
			} else {
				filters[filter] = value
			}
		}
	}

	return filters, errors
}

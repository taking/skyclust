package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ValidationRule represents a validation rule
type ValidationRule interface {
	Validate(value interface{}) error
	GetField() string
	GetMessage() string
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

// UnifiedValidator provides comprehensive validation functionality
type UnifiedValidator struct {
	rules map[string][]ValidationRule
}

// NewUnifiedValidator creates a new unified validator
func NewUnifiedValidator() *UnifiedValidator {
	return &UnifiedValidator{
		rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule for a field
func (uv *UnifiedValidator) AddRule(field string, rule ValidationRule) {
	if uv.rules[field] == nil {
		uv.rules[field] = make([]ValidationRule, 0)
	}
	uv.rules[field] = append(uv.rules[field], rule)
}

// Validate validates a struct or map
func (uv *UnifiedValidator) Validate(data interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	// Convert to map for easier processing
	dataMap := uv.convertToMap(data)

	// Validate each field
	for field, rules := range uv.rules {
		value, exists := dataMap[field]
		if !exists {
			// Field doesn't exist, check if it's required
			for _, rule := range rules {
				if requiredRule, ok := rule.(*RequiredRule); ok {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   field,
						Message: requiredRule.GetMessage(),
						Code:    "REQUIRED",
					})
				}
			}
			continue
		}

		// Validate field value
		for _, rule := range rules {
			if err := rule.Validate(value); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field,
					Value:   value,
					Message: err.Error(),
					Code:    "VALIDATION_ERROR",
				})
			}
		}
	}

	return result
}

// convertToMap converts a struct or map to a map[string]interface{}
func (uv *UnifiedValidator) convertToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		return v
	case map[string]string:
		for k, val := range v {
			result[k] = val
		}
	default:
		// Use reflection for structs
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			typ := val.Type()
			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i)
				fieldValue := val.Field(i)

				// Get field name from tag or use struct field name
				fieldName := field.Name
				if tag := field.Tag.Get("json"); tag != "" {
					fieldName = strings.Split(tag, ",")[0]
				}

				result[fieldName] = fieldValue.Interface()
			}
		}
	}

	return result
}

// Built-in validation rules

// RequiredRule validates that a field is not empty
type RequiredRule struct {
	message string
}

// NewRequiredRule creates a new required rule
func NewRequiredRule(message string) *RequiredRule {
	if message == "" {
		message = "Field is required"
	}
	return &RequiredRule{message: message}
}

func (r *RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return errors.New(r.message)
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return errors.New(r.message)
		}
	case []interface{}:
		if len(v) == 0 {
			return errors.New(r.message)
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return errors.New(r.message)
		}
	}

	return nil
}

func (r *RequiredRule) GetField() string {
	return ""
}

func (r *RequiredRule) GetMessage() string {
	return r.message
}

// EmailRule validates email format
type EmailRule struct {
	message string
	pattern *regexp.Regexp
}

// NewEmailRule creates a new email rule
func NewEmailRule(message string) *EmailRule {
	if message == "" {
		message = "Invalid email format"
	}
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return &EmailRule{
		message: message,
		pattern: pattern,
	}
}

func (r *EmailRule) Validate(value interface{}) error {
	if value == nil {
		return nil // Let RequiredRule handle nil values
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("Email must be a string")
	}

	if !r.pattern.MatchString(str) {
		return errors.New(r.message)
	}

	return nil
}

func (r *EmailRule) GetField() string {
	return ""
}

func (r *EmailRule) GetMessage() string {
	return r.message
}

// MinLengthRule validates minimum length
type MinLengthRule struct {
	min     int
	message string
}

// NewMinLengthRule creates a new minimum length rule
func NewMinLengthRule(min int, message string) *MinLengthRule {
	if message == "" {
		message = fmt.Sprintf("Minimum length is %d", min)
	}
	return &MinLengthRule{
		min:     min,
		message: message,
	}
}

func (r *MinLengthRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) < r.min {
			return errors.New(r.message)
		}
	case []interface{}:
		if len(v) < r.min {
			return errors.New(r.message)
		}
	}

	return nil
}

func (r *MinLengthRule) GetField() string {
	return ""
}

func (r *MinLengthRule) GetMessage() string {
	return r.message
}

// MaxLengthRule validates maximum length
type MaxLengthRule struct {
	max     int
	message string
}

// NewMaxLengthRule creates a new maximum length rule
func NewMaxLengthRule(max int, message string) *MaxLengthRule {
	if message == "" {
		message = fmt.Sprintf("Maximum length is %d", max)
	}
	return &MaxLengthRule{
		max:     max,
		message: message,
	}
}

func (r *MaxLengthRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) > r.max {
			return errors.New(r.message)
		}
	case []interface{}:
		if len(v) > r.max {
			return errors.New(r.message)
		}
	}

	return nil
}

func (r *MaxLengthRule) GetField() string {
	return ""
}

func (r *MaxLengthRule) GetMessage() string {
	return r.message
}

// UUIDRule validates UUID format
type UUIDRule struct {
	message string
}

// NewUUIDRule creates a new UUID rule
func NewUUIDRule(message string) *UUIDRule {
	if message == "" {
		message = "Invalid UUID format"
	}
	return &UUIDRule{message: message}
}

func (r *UUIDRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("UUID must be a string")
	}

	if _, err := uuid.Parse(str); err != nil {
		return errors.New(r.message)
	}

	return nil
}

func (r *UUIDRule) GetField() string {
	return ""
}

func (r *UUIDRule) GetMessage() string {
	return r.message
}

// RangeRule validates numeric range
type RangeRule struct {
	min     float64
	max     float64
	message string
}

// NewRangeRule creates a new range rule
func NewRangeRule(min, max float64, message string) *RangeRule {
	if message == "" {
		message = fmt.Sprintf("Value must be between %f and %f", min, max)
	}
	return &RangeRule{
		min:     min,
		max:     max,
		message: message,
	}
}

func (r *RangeRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		return fmt.Errorf("Value must be a number")
	}

	if num < r.min || num > r.max {
		return errors.New(r.message)
	}

	return nil
}

func (r *RangeRule) GetField() string {
	return ""
}

func (r *RangeRule) GetMessage() string {
	return r.message
}

// ValidationManager manages multiple validators
type ValidationManager struct {
	validators map[string]*UnifiedValidator
}

// NewValidationManager creates a new validation manager
func NewValidationManager() *ValidationManager {
	return &ValidationManager{
		validators: make(map[string]*UnifiedValidator),
	}
}

// GetValidator returns a validator by name
func (vm *ValidationManager) GetValidator(name string) *UnifiedValidator {
	if validator, exists := vm.validators[name]; exists {
		return validator
	}

	validator := NewUnifiedValidator()
	vm.validators[name] = validator
	return validator
}

// AddValidator adds a validator to the manager
func (vm *ValidationManager) AddValidator(name string, validator *UnifiedValidator) {
	vm.validators[name] = validator
}

// Predefined validation rules for common entities

// UserValidationRules sets up validation rules for user entities
func UserValidationRules(validator *UnifiedValidator) {
	validator.AddRule("username", NewRequiredRule("Username is required"))
	validator.AddRule("username", NewMinLengthRule(3, "Username must be at least 3 characters"))
	validator.AddRule("username", NewMaxLengthRule(50, "Username must be at most 50 characters"))

	validator.AddRule("email", NewRequiredRule("Email is required"))
	validator.AddRule("email", NewEmailRule("Invalid email format"))

	validator.AddRule("password", NewRequiredRule("Password is required"))
	validator.AddRule("password", NewMinLengthRule(8, "Password must be at least 8 characters"))
	validator.AddRule("password", NewMaxLengthRule(128, "Password must be at most 128 characters"))
}

// CredentialValidationRules sets up validation rules for credential entities
func CredentialValidationRules(validator *UnifiedValidator) {
	validator.AddRule("name", NewRequiredRule("Credential name is required"))
	validator.AddRule("name", NewMinLengthRule(1, "Credential name must be at least 1 character"))
	validator.AddRule("name", NewMaxLengthRule(100, "Credential name must be at most 100 characters"))

	validator.AddRule("provider", NewRequiredRule("Provider is required"))
	validator.AddRule("provider", NewMinLengthRule(1, "Provider must be at least 1 character"))
	validator.AddRule("provider", NewMaxLengthRule(50, "Provider must be at most 50 characters"))

	validator.AddRule("data", NewRequiredRule("Credential data is required"))
}

// WorkspaceValidationRules sets up validation rules for workspace entities
func WorkspaceValidationRules(validator *UnifiedValidator) {
	validator.AddRule("name", NewRequiredRule("Workspace name is required"))
	validator.AddRule("name", NewMinLengthRule(1, "Workspace name must be at least 1 character"))
	validator.AddRule("name", NewMaxLengthRule(100, "Workspace name must be at most 100 characters"))

	validator.AddRule("description", NewMaxLengthRule(500, "Description must be at most 500 characters"))
}

// VMValidationRules sets up validation rules for VM entities
func VMValidationRules(validator *UnifiedValidator) {
	validator.AddRule("name", NewRequiredRule("VM name is required"))
	validator.AddRule("name", NewMinLengthRule(1, "VM name must be at least 1 character"))
	validator.AddRule("name", NewMaxLengthRule(100, "VM name must be at most 100 characters"))

	validator.AddRule("provider", NewRequiredRule("Provider is required"))
	validator.AddRule("region", NewRequiredRule("Region is required"))
	validator.AddRule("instance_type", NewRequiredRule("Instance type is required"))
}

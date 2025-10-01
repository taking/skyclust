package domain

import (
	"fmt"
	"net/http"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}

// NewFieldValidationError creates a new field-specific validation error
func NewFieldValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Predefined domain errors
var (
	ErrUserNotFound = &DomainError{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
		Status:  http.StatusNotFound,
	}

	ErrUserAlreadyExists = &DomainError{
		Code:    "USER_ALREADY_EXISTS",
		Message: "User already exists",
		Status:  http.StatusConflict,
	}

	ErrInvalidCredentials = &DomainError{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid credentials",
		Status:  http.StatusUnauthorized,
	}

	ErrWorkspaceNotFound = &DomainError{
		Code:    "WORKSPACE_NOT_FOUND",
		Message: "Workspace not found",
		Status:  http.StatusNotFound,
	}

	ErrWorkspaceAccessDenied = &DomainError{
		Code:    "WORKSPACE_ACCESS_DENIED",
		Message: "Access denied to workspace",
		Status:  http.StatusForbidden,
	}

	ErrVMNotFound = &DomainError{
		Code:    "VM_NOT_FOUND",
		Message: "Virtual machine not found",
		Status:  http.StatusNotFound,
	}

	ErrVMAlreadyExists = &DomainError{
		Code:    "VM_ALREADY_EXISTS",
		Message: "Virtual machine already exists",
		Status:  http.StatusConflict,
	}

	ErrInvalidVMStatus = &DomainError{
		Code:    "INVALID_VM_STATUS",
		Message: "Invalid virtual machine status",
		Status:  http.StatusBadRequest,
	}

	ErrProviderNotFound = &DomainError{
		Code:    "PROVIDER_NOT_FOUND",
		Message: "Cloud provider not found",
		Status:  http.StatusNotFound,
	}

	ErrProviderNotInitialized = &DomainError{
		Code:    "PROVIDER_NOT_INITIALIZED",
		Message: "Cloud provider not initialized",
		Status:  http.StatusBadRequest,
	}
)

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

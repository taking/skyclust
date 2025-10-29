package readability

import (
	"fmt"
	"strings"
	"time"
)

// Constants for better readability
const (
	// HTTP Status Codes
	StatusOK                  = 200
	StatusCreated             = 201
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusInternalServerError = 500

	// Default Values
	DefaultPageSize  = 20
	MaxPageSize      = 100
	DefaultTimeout   = 30 * time.Second
	MaxRetryAttempts = 3
	DefaultCacheTTL  = 5 * time.Minute

	// String Lengths
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MinEmailLength    = 5
	MaxEmailLength    = 100

	// Database Limits
	MaxDBConnections = 100
	MaxQueryTimeout  = 10 * time.Second

	// API Limits
	MaxRequestSize     = 1024 * 1024      // 1MB
	MaxResponseSize    = 10 * 1024 * 1024 // 10MB
	RateLimitPerMinute = 1000

	// File Limits
	MaxFileSize      = 50 * 1024 * 1024 // 50MB
	MaxUploadFiles   = 10
	AllowedFileTypes = "jpg,jpeg,png,pdf,doc,docx,txt"

	// Cache Keys
	UserCachePrefix    = "user:"
	SessionCachePrefix = "session:"
	TokenCachePrefix   = "token:"
	ConfigCachePrefix  = "config:"

	// Error Messages
	ErrMsgUserNotFound       = "User not found"
	ErrMsgInvalidCredentials = "Invalid credentials"
	ErrMsgAccessDenied       = "Access denied"
	ErrMsgValidationFailed   = "Validation failed"
	ErrMsgInternalError      = "Internal server error"

	// Success Messages
	SuccessMsgUserCreated   = "User created successfully"
	SuccessMsgUserUpdated   = "User updated successfully"
	SuccessMsgUserDeleted   = "User deleted successfully"
	SuccessMsgLoginSuccess  = "Login successful"
	SuccessMsgLogoutSuccess = "Logout successful"
)

// ReadabilityHelper provides helper functions for better code readability
type ReadabilityHelper struct{}

// NewReadabilityHelper creates a new readability helper
func NewReadabilityHelper() *ReadabilityHelper {
	return &ReadabilityHelper{}
}

// ValidateStringLength validates string length with clear parameters
func (rh *ReadabilityHelper) ValidateStringLength(value, fieldName string, minLength, maxLength int) error {
	length := len(strings.TrimSpace(value))

	if length < minLength {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, minLength)
	}

	if length > maxLength {
		return fmt.Errorf("%s must be at most %d characters long", fieldName, maxLength)
	}

	return nil
}

// ValidateEmail validates email format
func (rh *ReadabilityHelper) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if err := rh.ValidateStringLength(email, "email", MinEmailLength, MaxEmailLength); err != nil {
		return err
	}

	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateUsername validates username
func (rh *ReadabilityHelper) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	return rh.ValidateStringLength(username, "username", MinUsernameLength, MaxUsernameLength)
}

// ValidatePassword validates password
func (rh *ReadabilityHelper) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	return rh.ValidateStringLength(password, "password", MinPasswordLength, MaxPasswordLength)
}

// IsValidFileType checks if file type is allowed
func (rh *ReadabilityHelper) IsValidFileType(filename string) bool {
	extension := strings.ToLower(getFileExtension(filename))
	allowedTypes := strings.Split(AllowedFileTypes, ",")

	for _, allowedType := range allowedTypes {
		if extension == strings.TrimSpace(allowedType) {
			return true
		}
	}

	return false
}

// IsValidFileSize checks if file size is within limits
func (rh *ReadabilityHelper) IsValidFileSize(size int64) bool {
	return size > 0 && size <= MaxFileSize
}

// FormatDuration formats duration for better readability
func (rh *ReadabilityHelper) FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}

	if duration < time.Hour {
		return fmt.Sprintf("%.2fm", duration.Minutes())
	}

	return fmt.Sprintf("%.2fh", duration.Hours())
}

// FormatBytes formats bytes for better readability
func (rh *ReadabilityHelper) FormatBytes(bytes int64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper functions
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// FunctionDecomposer helps decompose complex functions
type FunctionDecomposer struct{}

// NewFunctionDecomposer creates a new function decomposer
func NewFunctionDecomposer() *FunctionDecomposer {
	return &FunctionDecomposer{}
}

// DecomposeComplexFunction breaks down a complex function into smaller parts
func (fd *FunctionDecomposer) DecomposeComplexFunction(complexFunc func() error) error {
	// Step 1: Validate inputs
	if err := fd.validateInputs(); err != nil {
		return err
	}

	// Step 2: Process data
	if err := fd.processData(); err != nil {
		return err
	}

	// Step 3: Save results
	if err := fd.saveResults(); err != nil {
		return err
	}

	// Step 4: Cleanup
	fd.cleanup()

	return nil
}

func (fd *FunctionDecomposer) validateInputs() error {
	// Implementation for input validation
	return nil
}

func (fd *FunctionDecomposer) processData() error {
	// Implementation for data processing
	return nil
}

func (fd *FunctionDecomposer) saveResults() error {
	// Implementation for saving results
	return nil
}

func (fd *FunctionDecomposer) cleanup() {
	// Implementation for cleanup
}

// ConditionalSimplifier helps simplify complex conditionals
type ConditionalSimplifier struct{}

// NewConditionalSimplifier creates a new conditional simplifier
func NewConditionalSimplifier() *ConditionalSimplifier {
	return &ConditionalSimplifier{}
}

// SimplifyComplexCondition simplifies complex conditional logic
func (cs *ConditionalSimplifier) SimplifyComplexCondition(userRole string, resourceOwner string, userID string) bool {
	// Extract conditions into named variables for clarity
	isAdmin := cs.isAdmin(userRole)
	isOwner := cs.isOwner(resourceOwner, userID)
	hasAccess := cs.hasAccess(userRole)

	// Combine conditions with clear logic
	return isAdmin || isOwner || hasAccess
}

func (cs *ConditionalSimplifier) isAdmin(role string) bool {
	return role == "admin"
}

func (cs *ConditionalSimplifier) isOwner(resourceOwner, userID string) bool {
	return resourceOwner == userID
}

func (cs *ConditionalSimplifier) hasAccess(role string) bool {
	allowedRoles := []string{"user", "moderator", "admin"}
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// MagicNumberReplacer helps replace magic numbers with named constants
type MagicNumberReplacer struct{}

// NewMagicNumberReplacer creates a new magic number replacer
func NewMagicNumberReplacer() *MagicNumberReplacer {
	return &MagicNumberReplacer{}
}

// ReplaceMagicNumbers replaces magic numbers with meaningful constants
func (mnr *MagicNumberReplacer) ReplaceMagicNumbers() map[string]int {
	return map[string]int{
		"max_retry_attempts":    MaxRetryAttempts,
		"default_page_size":     DefaultPageSize,
		"max_page_size":         MaxPageSize,
		"min_username_length":   MinUsernameLength,
		"max_username_length":   MaxUsernameLength,
		"min_password_length":   MinPasswordLength,
		"max_password_length":   MaxPasswordLength,
		"rate_limit_per_minute": RateLimitPerMinute,
		"max_file_size_mb":      MaxFileSize / (1024 * 1024),
		"max_upload_files":      MaxUploadFiles,
	}
}

// CodeStructureImprover helps improve code structure
type CodeStructureImprover struct{}

// NewCodeStructureImprover creates a new code structure improver
func NewCodeStructureImprover() *CodeStructureImprover {
	return &CodeStructureImprover{}
}

// ImproveCodeStructure provides guidelines for better code structure
func (csi *CodeStructureImprover) ImproveCodeStructure() map[string]string {
	return map[string]string{
		"function_length":       "Keep functions under 50 lines",
		"parameter_count":       "Limit function parameters to 5 or fewer",
		"nesting_depth":         "Limit nesting depth to 3 levels",
		"variable_naming":       "Use descriptive variable names",
		"function_naming":       "Use verb-noun pattern for function names",
		"constant_naming":       "Use UPPER_CASE for constants",
		"error_handling":        "Handle errors explicitly",
		"comment_style":         "Write comments that explain why, not what",
		"code_organization":     "Group related functionality together",
		"dependency_management": "Minimize dependencies and coupling",
	}
}

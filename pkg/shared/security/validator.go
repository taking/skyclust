package security

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// InputValidator validates user input
type InputValidator struct {
	emailRegex    *regexp.Regexp
	usernameRegex *regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`),
	}
}

// ValidateJSON validates JSON structure
func (v *InputValidator) ValidateJSON(data []byte) error {
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	return nil
}

// ValidateQueryParam validates query parameters
func (v *InputValidator) ValidateQueryParam(key, value string) error {
	// Check for SQL injection patterns
	sqlPatterns := []string{
		"'; DROP TABLE",
		"UNION SELECT",
		"OR 1=1",
		"AND 1=1",
		"<script>",
		"javascript:",
	}

	valueLower := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(valueLower, pattern) {
			return fmt.Errorf("potentially malicious input detected")
		}
	}

	// Check length
	if len(value) > 1000 {
		return fmt.Errorf("parameter value too long")
	}

	return nil
}

// ValidateEmail validates email format
func (v *InputValidator) ValidateEmail(email string) error {
	if !v.emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateUsername validates username format
func (v *InputValidator) ValidateUsername(username string) error {
	if !v.usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username format")
	}
	return nil
}

// SanitizeInput sanitizes user input
func (v *InputValidator) SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	re := regexp.MustCompile(`[<>'"&]`)
	return re.ReplaceAllString(input, "")
}

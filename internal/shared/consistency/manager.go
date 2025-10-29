package consistency

import (
	"skyclust/internal/shared/errors"
	"skyclust/internal/shared/logging"
	"skyclust/internal/shared/validation"
)

// UnifiedConsistencyManager provides unified consistency management
type UnifiedConsistencyManager struct {
	logger            *logging.UnifiedLogger
	errorHandler      *errors.ErrorHandler
	validationManager *validation.ValidationManager
	config            *ConsistencyConfig
}

// ConsistencyConfig holds consistency configuration
type ConsistencyConfig struct {
	Logging       LoggingConfig       `json:"logging"`
	ErrorHandling ErrorHandlingConfig `json:"error_handling"`
	Validation    ValidationConfig    `json:"validation"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	EnableBusinessLogging    bool `json:"enable_business_logging"`
	EnableAuditLogging       bool `json:"enable_audit_logging"`
	EnableSecurityLogging    bool `json:"enable_security_logging"`
	EnablePerformanceLogging bool `json:"enable_performance_logging"`
}

// ErrorHandlingConfig holds error handling configuration
type ErrorHandlingConfig struct {
	EnableDetailedErrors bool `json:"enable_detailed_errors"`
	EnableStackTrace     bool `json:"enable_stack_trace"`
	EnableErrorContext   bool `json:"enable_error_context"`
}

// ValidationConfig holds validation configuration
type ValidationConfig struct {
	EnableFieldValidation  bool `json:"enable_field_validation"`
	EnableEntityValidation bool `json:"enable_entity_validation"`
	EnableCustomRules      bool `json:"enable_custom_rules"`
}

// NewUnifiedConsistencyManager creates a new unified consistency manager
func NewUnifiedConsistencyManager(logger *logging.UnifiedLogger, errorHandler *errors.ErrorHandler, validationManager *validation.ValidationManager, config *ConsistencyConfig) *UnifiedConsistencyManager {
	if config == nil {
		config = DefaultConsistencyConfig()
	}

	return &UnifiedConsistencyManager{
		logger:            logger,
		errorHandler:      errorHandler,
		validationManager: validationManager,
		config:            config,
	}
}

// DefaultConsistencyConfig returns default consistency configuration
func DefaultConsistencyConfig() *ConsistencyConfig {
	return &ConsistencyConfig{
		Logging: LoggingConfig{
			EnableBusinessLogging:    true,
			EnableAuditLogging:       true,
			EnableSecurityLogging:    true,
			EnablePerformanceLogging: true,
		},
		ErrorHandling: ErrorHandlingConfig{
			EnableDetailedErrors: true,
			EnableStackTrace:     true,
			EnableErrorContext:   true,
		},
		Validation: ValidationConfig{
			EnableFieldValidation:  true,
			EnableEntityValidation: true,
			EnableCustomRules:      true,
		},
	}
}

// LogInfo logs an info message
func (ucm *UnifiedConsistencyManager) LogInfo(message string, fields ...interface{}) {
	ucm.logger.LogInfo(message)
}

// LogWarn logs a warning message
func (ucm *UnifiedConsistencyManager) LogWarn(message string, fields ...interface{}) {
	ucm.logger.LogWarn(message)
}

// LogError logs an error message
func (ucm *UnifiedConsistencyManager) LogError(message string, fields ...interface{}) {
	ucm.logger.LogError(message)
}

// LogDebug logs a debug message
func (ucm *UnifiedConsistencyManager) LogDebug(message string, fields ...interface{}) {
	ucm.logger.LogDebug(message)
}

// LogBusinessEvent logs a business event
func (ucm *UnifiedConsistencyManager) LogBusinessEvent(eventType, userID, resourceID string, details map[string]interface{}) {
	if ucm.config.Logging.EnableBusinessLogging {
		ucm.logger.LogBusinessEvent(eventType, userID, resourceID, details)
	}
}

// LogAuditEvent logs an audit event
func (ucm *UnifiedConsistencyManager) LogAuditEvent(eventType, resourceType, resourceID, userID string, details map[string]interface{}) {
	if ucm.config.Logging.EnableAuditLogging {
		ucm.logger.LogAuditEvent(eventType, resourceType, resourceID, userID, details)
	}
}

// LogSecurityEvent logs a security event
func (ucm *UnifiedConsistencyManager) LogSecurityEvent(eventType, userID string, details map[string]interface{}) {
	if ucm.config.Logging.EnableSecurityLogging {
		ucm.logger.LogSecurityEvent(eventType, userID, details)
	}
}

// LogPerformanceEvent logs a performance event
func (ucm *UnifiedConsistencyManager) LogPerformanceEvent(operation string, duration interface{}, details map[string]interface{}) {
	if ucm.config.Logging.EnablePerformanceLogging {
		ucm.logger.LogPerformanceEvent(operation, 0, details) // Simplified for now
	}
}

// HandleError handles an error with consistent processing
func (ucm *UnifiedConsistencyManager) HandleError(err error, context map[string]interface{}) *errors.UnifiedError {
	return ucm.errorHandler.HandleError(err, context)
}

// CreateValidationError creates a validation error
func (ucm *UnifiedConsistencyManager) CreateValidationError(message string, details map[string]interface{}) *errors.UnifiedError {
	return ucm.errorHandler.CreateValidationError(message, details)
}

// CreateAuthenticationError creates an authentication error
func (ucm *UnifiedConsistencyManager) CreateAuthenticationError(message string) *errors.UnifiedError {
	return ucm.errorHandler.CreateAuthenticationError(message)
}

// CreateAuthorizationError creates an authorization error
func (ucm *UnifiedConsistencyManager) CreateAuthorizationError(message string) *errors.UnifiedError {
	return ucm.errorHandler.CreateAuthorizationError(message)
}

// CreateBusinessError creates a business error
func (ucm *UnifiedConsistencyManager) CreateBusinessError(message string, details map[string]interface{}) *errors.UnifiedError {
	return ucm.errorHandler.CreateBusinessError(message, details)
}

// CreateInfrastructureError creates an infrastructure error
func (ucm *UnifiedConsistencyManager) CreateInfrastructureError(message string, cause error) *errors.UnifiedError {
	return ucm.errorHandler.CreateInfrastructureError(message, cause)
}

// CreateExternalError creates an external service error
func (ucm *UnifiedConsistencyManager) CreateExternalError(message string, cause error) *errors.UnifiedError {
	return ucm.errorHandler.CreateExternalError(message, cause)
}

// CreateSystemError creates a system error
func (ucm *UnifiedConsistencyManager) CreateSystemError(message string, cause error) *errors.UnifiedError {
	return ucm.errorHandler.CreateSystemError(message, cause)
}

// ValidateField validates a single field
func (ucm *UnifiedConsistencyManager) ValidateField(field string, value interface{}, rules []validation.ValidationRule) error {
	if !ucm.config.Validation.EnableFieldValidation {
		return nil
	}

	validator := ucm.validationManager.GetValidator("field")
	for _, rule := range rules {
		validator.AddRule(field, rule)
	}

	result := validator.Validate(map[string]interface{}{field: value})
	if !result.Valid {
		if len(result.Errors) > 0 {
			return &errors.UnifiedError{
				Code:    "VALIDATION_ERROR",
				Message: result.Errors[0].Message,
			}
		}
	}
	return nil
}

// ValidateEntity validates an entire entity
func (ucm *UnifiedConsistencyManager) ValidateEntity(entityName string, data interface{}) *validation.ValidationResult {
	if !ucm.config.Validation.EnableEntityValidation {
		return &validation.ValidationResult{Valid: true}
	}

	validator := ucm.validationManager.GetValidator(entityName)
	return validator.Validate(data)
}

// TrackRequest tracks request performance
func (ucm *UnifiedConsistencyManager) TrackRequest(operation string, startTime interface{}, details map[string]interface{}) {
	if ucm.config.Logging.EnablePerformanceLogging {
		ucm.LogPerformanceEvent(operation, startTime, details)
	}
}

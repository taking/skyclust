package quality

import (
	"skyclust/internal/shared/errors"
	"skyclust/internal/shared/logging"
	"skyclust/internal/shared/validation"
	"time"
)

// QualityManager provides comprehensive quality management functionality
type QualityManager struct {
	logger            *logging.UnifiedLogger
	errorHandler      *errors.ErrorHandler
	validationManager *validation.ValidationManager
	config            *QualityConfig
}

// QualityConfig holds quality management configuration
type QualityConfig struct {
	EnableLogging       bool `json:"enable_logging"`
	EnableErrorHandling bool `json:"enable_error_handling"`
	EnableValidation    bool `json:"enable_validation"`
	EnablePerformance   bool `json:"enable_performance"`
	EnableAudit         bool `json:"enable_audit"`
	EnableSecurity      bool `json:"enable_security"`
}

// NewQualityManager creates a new quality manager
func NewQualityManager(logger *logging.UnifiedLogger, errorHandler *errors.ErrorHandler, validationManager *validation.ValidationManager, config *QualityConfig) *QualityManager {
	if config == nil {
		config = DefaultQualityConfig()
	}

	return &QualityManager{
		logger:            logger,
		errorHandler:      errorHandler,
		validationManager: validationManager,
		config:            config,
	}
}

// DefaultQualityConfig returns default quality configuration
func DefaultQualityConfig() *QualityConfig {
	return &QualityConfig{
		EnableLogging:       true,
		EnableErrorHandling: true,
		EnableValidation:    true,
		EnablePerformance:   true,
		EnableAudit:         true,
		EnableSecurity:      true,
	}
}

// LogInfo logs an info message
func (qm *QualityManager) LogInfo(message string, fields ...interface{}) {
	if qm.config.EnableLogging {
		qm.logger.LogInfo(message)
	}
}

// LogWarn logs a warning message
func (qm *QualityManager) LogWarn(message string, fields ...interface{}) {
	if qm.config.EnableLogging {
		qm.logger.LogWarn(message)
	}
}

// LogError logs an error message
func (qm *QualityManager) LogError(message string, fields ...interface{}) {
	if qm.config.EnableLogging {
		qm.logger.LogError(message)
	}
}

// LogDebug logs a debug message
func (qm *QualityManager) LogDebug(message string, fields ...interface{}) {
	if qm.config.EnableLogging {
		qm.logger.LogDebug(message)
	}
}

// LogBusinessEvent logs a business event
func (qm *QualityManager) LogBusinessEvent(eventType, userID, resourceID string, details map[string]interface{}) {
	if qm.config.EnableAudit {
		qm.logger.LogBusinessEvent(eventType, userID, resourceID, details)
	}
}

// LogAuditEvent logs an audit event
func (qm *QualityManager) LogAuditEvent(eventType, resourceType, resourceID, userID string, details map[string]interface{}) {
	if qm.config.EnableAudit {
		qm.logger.LogAuditEvent(eventType, resourceType, resourceID, userID, details)
	}
}

// LogSecurityEvent logs a security event
func (qm *QualityManager) LogSecurityEvent(eventType, userID string, details map[string]interface{}) {
	if qm.config.EnableSecurity {
		qm.logger.LogSecurityEvent(eventType, userID, details)
	}
}

// LogPerformanceEvent logs a performance event
func (qm *QualityManager) LogPerformanceEvent(operation string, duration time.Duration, details map[string]interface{}) {
	if qm.config.EnablePerformance {
		qm.logger.LogPerformanceEvent(operation, duration, details)
	}
}

// HandleError handles an error
func (qm *QualityManager) HandleError(err error, context map[string]interface{}) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.HandleError(err, context)
	}
	return nil
}

// CreateValidationError creates a validation error
func (qm *QualityManager) CreateValidationError(message string, details map[string]interface{}) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateValidationError(message, details)
	}
	return nil
}

// CreateAuthenticationError creates an authentication error
func (qm *QualityManager) CreateAuthenticationError(message string) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateAuthenticationError(message)
	}
	return nil
}

// CreateAuthorizationError creates an authorization error
func (qm *QualityManager) CreateAuthorizationError(message string) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateAuthorizationError(message)
	}
	return nil
}

// CreateBusinessError creates a business error
func (qm *QualityManager) CreateBusinessError(message string, details map[string]interface{}) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateBusinessError(message, details)
	}
	return nil
}

// CreateInfrastructureError creates an infrastructure error
func (qm *QualityManager) CreateInfrastructureError(message string, cause error) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateInfrastructureError(message, cause)
	}
	return nil
}

// CreateExternalError creates an external service error
func (qm *QualityManager) CreateExternalError(message string, cause error) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateExternalError(message, cause)
	}
	return nil
}

// CreateSystemError creates a system error
func (qm *QualityManager) CreateSystemError(message string, cause error) *errors.UnifiedError {
	if qm.config.EnableErrorHandling {
		return qm.errorHandler.CreateSystemError(message, cause)
	}
	return nil
}

// Validate validates data using the specified validator
func (qm *QualityManager) Validate(validatorName string, data interface{}) *validation.ValidationResult {
	if qm.config.EnableValidation {
		validator := qm.validationManager.GetValidator(validatorName)
		return validator.Validate(data)
	}
	return &validation.ValidationResult{Valid: true}
}

// GetValidator returns a validator by name
func (qm *QualityManager) GetValidator(name string) *validation.UnifiedValidator {
	return qm.validationManager.GetValidator(name)
}

// QualityMiddleware provides middleware for quality management
type QualityMiddleware struct {
	qualityManager *QualityManager
}

// NewQualityMiddleware creates a new quality middleware
func NewQualityMiddleware(qualityManager *QualityManager) *QualityMiddleware {
	return &QualityMiddleware{
		qualityManager: qualityManager,
	}
}

// TrackRequest tracks request performance
func (qm *QualityMiddleware) TrackRequest(operation string, startTime time.Time) {
	if qm.qualityManager.config.EnablePerformance {
		duration := time.Since(startTime)
		qm.qualityManager.LogPerformanceEvent(operation, duration, map[string]interface{}{
			"operation": operation,
		})
	}
}

// QualityService provides per-service quality management
type QualityService struct {
	*QualityManager
	serviceName string
}

// NewQualityService creates a new quality service
func NewQualityService(qualityManager *QualityManager, serviceName string) *QualityService {
	return &QualityService{
		QualityManager: qualityManager,
		serviceName:    serviceName,
	}
}

// LogServiceInfo logs service-specific info
func (qs *QualityService) LogServiceInfo(message string, fields ...interface{}) {
	qs.LogInfo(qs.serviceName+": "+message, fields...)
}

// LogServiceWarn logs service-specific warning
func (qs *QualityService) LogServiceWarn(message string, fields ...interface{}) {
	qs.LogWarn(qs.serviceName+": "+message, fields...)
}

// LogServiceError logs service-specific error
func (qs *QualityService) LogServiceError(message string, fields ...interface{}) {
	qs.LogError(qs.serviceName+": "+message, fields...)
}

// LogServiceDebug logs service-specific debug
func (qs *QualityService) LogServiceDebug(message string, fields ...interface{}) {
	qs.LogDebug(qs.serviceName+": "+message, fields...)
}

// LogServiceBusinessEvent logs service-specific business event
func (qs *QualityService) LogServiceBusinessEvent(eventType, userID, resourceID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	qs.LogBusinessEvent(eventType, userID, resourceID, details)
}

// LogServiceAuditEvent logs service-specific audit event
func (qs *QualityService) LogServiceAuditEvent(eventType, resourceType, resourceID, userID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	qs.LogAuditEvent(eventType, resourceType, resourceID, userID, details)
}

// LogServiceSecurityEvent logs service-specific security event
func (qs *QualityService) LogServiceSecurityEvent(eventType, userID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	qs.LogSecurityEvent(eventType, userID, details)
}

// LogServicePerformanceEvent logs service-specific performance event
func (qs *QualityService) LogServicePerformanceEvent(operation string, duration time.Duration, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	qs.LogPerformanceEvent(operation, duration, details)
}

// HandleServiceError handles service-specific error
func (qs *QualityService) HandleServiceError(err error, context map[string]interface{}) *errors.UnifiedError {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["service"] = qs.serviceName
	return qs.HandleError(err, context)
}

// CreateServiceValidationError creates service-specific validation error
func (qs *QualityService) CreateServiceValidationError(message string, details map[string]interface{}) *errors.UnifiedError {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	return qs.CreateValidationError(message, details)
}

// CreateServiceBusinessError creates service-specific business error
func (qs *QualityService) CreateServiceBusinessError(message string, details map[string]interface{}) *errors.UnifiedError {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = qs.serviceName
	return qs.CreateBusinessError(message, details)
}

// CreateServiceInfrastructureError creates service-specific infrastructure error
func (qs *QualityService) CreateServiceInfrastructureError(message string, cause error) *errors.UnifiedError {
	return qs.CreateInfrastructureError(qs.serviceName+": "+message, cause)
}

// CreateServiceExternalError creates service-specific external error
func (qs *QualityService) CreateServiceExternalError(message string, cause error) *errors.UnifiedError {
	return qs.CreateExternalError(qs.serviceName+": "+message, cause)
}

// CreateServiceSystemError creates service-specific system error
func (qs *QualityService) CreateServiceSystemError(message string, cause error) *errors.UnifiedError {
	return qs.CreateSystemError(qs.serviceName+": "+message, cause)
}

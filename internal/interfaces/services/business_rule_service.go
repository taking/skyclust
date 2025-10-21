package services

// BusinessRuleService defines the interface for business rule operations
type BusinessRuleService interface {
	// GetBusinessRules retrieves all business rules
	GetBusinessRules() ([]interface{}, error)

	// GetBusinessRule retrieves a specific business rule
	GetBusinessRule(id string) (interface{}, error)

	// CreateBusinessRule creates a new business rule
	CreateBusinessRule(rule interface{}) error

	// UpdateBusinessRule updates an existing business rule
	UpdateBusinessRule(rule interface{}) error

	// DeleteBusinessRule deletes a business rule
	DeleteBusinessRule(id string) error

	// EvaluateBusinessRule evaluates a business rule against given context
	EvaluateBusinessRule(ruleID string, context map[string]interface{}) (bool, error)

	// GetBusinessRuleViolations retrieves violations of business rules
	GetBusinessRuleViolations(workspaceID string) ([]interface{}, error)

	// ValidateWorkspace validates a workspace against business rules
	ValidateWorkspace(workspaceID string) ([]interface{}, error)

	// GetBusinessRuleMetrics retrieves metrics for business rules
	GetBusinessRuleMetrics() (interface{}, error)
}

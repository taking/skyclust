package credential

import "time"

// CreateCredentialRequest represents a credential creation request
type CreateCredentialRequest struct {
	Provider string                 `json:"provider" validate:"required,oneof=aws gcp openstack azure"`
	Name     string                 `json:"name" validate:"required,min=1,max=100"`
	Data     map[string]interface{} `json:"data" validate:"required"`
}

// UpdateCredentialRequest represents a credential update request
type UpdateCredentialRequest struct {
	Name *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// CredentialResponse represents a credential in API responses
type CredentialResponse struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Note: Data field is intentionally omitted for security
}

// CredentialListResponse represents a list of credentials
type CredentialListResponse struct {
	Credentials []*CredentialResponse `json:"credentials"`
	Total       int64                 `json:"total"`
}

// CredentialByProviderResponse represents credentials grouped by provider
type CredentialByProviderResponse struct {
	Provider    string                `json:"provider"`
	Credentials []*CredentialResponse `json:"credentials"`
	Count       int                   `json:"count"`
}

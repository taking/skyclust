package oidc

// AuthURLRequest represents an OIDC auth URL request
type AuthURLRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google github azure microsoft"`
}

// AuthURLResponse represents an OIDC auth URL response
type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// CallbackRequest represents an OIDC callback request
type CallbackRequest struct {
	Provider string `json:"provider" validate:"required"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}

// CallbackResponse represents an OIDC callback response
type CallbackResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

// UserResponse represents a user from OIDC provider
type UserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture,omitempty"`
}

// LogoutRequest represents an OIDC logout request
type LogoutRequest struct {
	UserID                string `json:"user_id" validate:"required"`
	Provider              string `json:"provider" validate:"required"`
	IDToken               string `json:"id_token" validate:"required"`
	PostLogoutRedirectURI string `json:"post_logout_redirect_uri,omitempty"`
}

// LogoutURLRequest represents a logout URL request
type LogoutURLRequest struct {
	Provider              string `json:"provider" validate:"required"`
	PostLogoutRedirectURI string `json:"post_logout_redirect_uri,omitempty"`
}

// LogoutURLResponse represents a logout URL response
type LogoutURLResponse struct {
	LogoutURL string `json:"logout_url"`
}

// ProviderResponse represents an OIDC provider
type ProviderResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

// ProviderListResponse represents a list of OIDC providers
type ProviderListResponse struct {
	Providers []*ProviderResponse `json:"providers"`
	Total     int                 `json:"total"`
}

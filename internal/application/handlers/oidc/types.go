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

// CreateProviderRequest represents a request to create an OIDC provider
type CreateProviderRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	ProviderType string `json:"provider_type" validate:"required,oneof=google github azure microsoft custom"`
	ClientID     string `json:"client_id" validate:"required,min=1"`
	ClientSecret string `json:"client_secret" validate:"required,min=1"`
	RedirectURL  string `json:"redirect_url" validate:"required,url"`
	AuthURL      string `json:"auth_url,omitempty" validate:"omitempty,url"`
	TokenURL     string `json:"token_url,omitempty" validate:"omitempty,url"`
	UserInfoURL  string `json:"user_info_url,omitempty" validate:"omitempty,url"`
	Scopes       string `json:"scopes,omitempty" validate:"omitempty"`
	Enabled      bool   `json:"enabled,omitempty"`
}

// UpdateProviderRequest represents a request to update an OIDC provider
type UpdateProviderRequest struct {
	Name         string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ClientID     string `json:"client_id,omitempty" validate:"omitempty,min=1"`
	ClientSecret string `json:"client_secret,omitempty" validate:"omitempty,min=1"`
	RedirectURL  string `json:"redirect_url,omitempty" validate:"omitempty,url"`
	AuthURL      string `json:"auth_url,omitempty" validate:"omitempty,url"`
	TokenURL     string `json:"token_url,omitempty" validate:"omitempty,url"`
	UserInfoURL  string `json:"user_info_url,omitempty" validate:"omitempty,url"`
	Scopes       string `json:"scopes,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
}

// ProviderDetailResponse represents detailed OIDC provider information
type ProviderDetailResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	ClientID     string `json:"client_id"`
	RedirectURL  string `json:"redirect_url"`
	AuthURL      string `json:"auth_url,omitempty"`
	TokenURL     string `json:"token_url,omitempty"`
	UserInfoURL  string `json:"user_info_url,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
	Enabled      bool   `json:"enabled"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

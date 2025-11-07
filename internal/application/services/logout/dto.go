package logout

import "github.com/google/uuid"

// LogoutRequest represents a logout request
type LogoutRequest struct {
	UserID                uuid.UUID `json:"user_id"`
	Token                 string    `json:"token"`
	AuthType              string    `json:"auth_type"`                          // "jwt" or "oidc"
	Provider              string    `json:"provider,omitempty"`                 // For OIDC
	IDToken               string    `json:"id_token,omitempty"`                 // For OIDC
	PostLogoutRedirectURI string    `json:"post_logout_redirect_uri,omitempty"` // For OIDC
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Success               bool   `json:"success"`
	Message               string `json:"message"`
	LogoutURL             string `json:"logout_url,omitempty"` // For OIDC front-channel logout
	PostLogoutRedirectURI string `json:"post_logout_redirect_uri,omitempty"`
}

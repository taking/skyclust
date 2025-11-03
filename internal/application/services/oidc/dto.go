package oidc

import (
	"time"

	"golang.org/x/oauth2"
)

// OIDCState represents stored OIDC state information
type OIDCState struct {
	Provider    string    `json:"provider"`
	Timestamp   time.Time `json:"timestamp"`
	UserSession string    `json:"user_session,omitempty"`
}

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Config       *oauth2.Config
}

// OIDCUserInfo represents user information from OIDC provider
type OIDCUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"login"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar_url"`
}


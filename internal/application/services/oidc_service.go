package service

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"net/http"
	"net/url"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
)

// OIDCService handles OIDC authentication
type OIDCService struct {
	userRepo     domain.UserRepository
	auditLogRepo domain.AuditLogRepository
	authService  domain.AuthService
	configs      map[string]*OIDCConfig
	httpClient   *http.Client
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

// NewOIDCService creates a new OIDC service
func NewOIDCService(
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
	authService domain.AuthService,
) domain.OIDCService {
	service := &OIDCService{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
		authService:  authService,
		configs:      make(map[string]*OIDCConfig),
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Initialize OIDC configurations
	service.initializeConfigs()

	return service
}

// GetAuthURL returns the OAuth authorization URL for the specified provider
func (s *OIDCService) GetAuthURL(ctx context.Context, provider string, state string) (string, error) {
	config, exists := s.configs[provider]
	if !exists {
		return "", domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
	}

	// Add state parameter
	authURL := config.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return authURL, nil
}

// ExchangeCode exchanges authorization code for access token and user info
func (s *OIDCService) ExchangeCode(ctx context.Context, provider, code, state string) (*domain.User, string, error) {
	config, exists := s.configs[provider]
	if !exists {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
	}

	// Exchange code for token
	token, err := config.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to exchange code for token", 500)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfo(provider, token)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user info", 500)
	}

	// Check if user exists
	user, err := s.userRepo.GetByOIDC(provider, userInfo.ID)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing user", 500)
	}

	// Create user if doesn't exist
	if user == nil {
		user = &domain.User{
			Username:     userInfo.Username,
			Email:        userInfo.Email,
			OIDCProvider: provider,
			OIDCSubject:  userInfo.ID,
			Active:       true,
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to create user", 500)
		}

		// Log OIDC registration
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:   user.ID,
			Action:   domain.ActionUserRegister,
			Resource: "POST /api/v1/auth/oidc/login",
			Details: map[string]interface{}{
				"provider": provider,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}

	// Generate JWT token using auth service
	_, jwtToken, err := s.authService.Login(user.Username, "OIDC_USER") // Use dummy password for OIDC
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate JWT token", 500)
	}

	// Log OIDC login
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   user.ID,
		Action:   domain.ActionOIDCLogin,
		Resource: "POST /api/v1/auth/oidc/login",
		Details: map[string]interface{}{
			"provider": provider,
		},
	})

	return user, jwtToken, nil
}

// getUserInfo fetches user information from the OIDC provider
func (s *OIDCService) getUserInfo(provider string, token *oauth2.Token) (*OIDCUserInfo, error) {
	client := s.configs[provider].Config.Client(context.Background(), token)

	var userInfo OIDCUserInfo
	var apiURL string

	switch provider {
	case "google":
		apiURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		apiURL = "https://api.github.com/user"
	case "azure":
		apiURL = "https://graph.microsoft.com/v1.0/me"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// EndSession initiates OIDC logout by calling the provider's end_session_endpoint
func (s *OIDCService) EndSession(ctx context.Context, userID uuid.UUID, provider, idToken, postLogoutRedirectURI string) error {
	config, exists := s.configs[provider]
	if !exists {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
	}

	// Get the end_session_endpoint URL for the provider
	endSessionURL, err := s.getEndSessionEndpoint(provider)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to get end session endpoint", 500)
	}

	// Build logout URL with required parameters
	logoutURL, err := s.buildLogoutURL(endSessionURL, idToken, postLogoutRedirectURI, config.ClientID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to build logout URL", 500)
	}

	// Call the end_session_endpoint
	resp, err := s.httpClient.Get(logoutURL)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to call end session endpoint", 500)
	}
	defer resp.Body.Close()

	// Log OIDC logout
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userID,
		Action:   domain.ActionOIDCLogout,
		Resource: "POST /api/v1/auth/oidc/logout",
		Details: map[string]interface{}{
			"provider": provider,
			"status":   resp.StatusCode,
		},
	})

	return nil
}

// getEndSessionEndpoint returns the end_session_endpoint URL for the provider
func (s *OIDCService) getEndSessionEndpoint(provider string) (string, error) {
	switch provider {
	case "google":
		return "https://accounts.google.com/logout", nil
	case "github":
		// GitHub doesn't have a standard end_session_endpoint
		// We'll use a custom logout flow
		return "https://github.com/logout", nil
	case "azure":
		return "https://login.microsoftonline.com/common/oauth2/v2.0/logout", nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// buildLogoutURL constructs the logout URL with required parameters
func (s *OIDCService) buildLogoutURL(endSessionURL, idToken, postLogoutRedirectURI, clientID string) (string, error) {
	u, err := url.Parse(endSessionURL)
	if err != nil {
		return "", err
	}

	params := u.Query()
	params.Set("id_token_hint", idToken)
	params.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	params.Set("client_id", clientID)
	u.RawQuery = params.Encode()

	return u.String(), nil
}

// HandleBackChannelLogout handles back-channel logout notifications from OIDC providers
func (s *OIDCService) HandleBackChannelLogout(ctx context.Context, logoutToken string) error {
	// Parse and validate the logout token
	// This is a simplified implementation - in production, you should properly validate the JWT
	// and extract the user information from the token claims

	// For now, we'll just log the back-channel logout
	// In a real implementation, you would:
	// 1. Validate the logout token signature
	// 2. Extract user information from the token
	// 3. Invalidate local sessions for that user
	// 4. Notify the user's active sessions

	return nil
}

// GetLogoutURL returns a logout URL that can be used for front-channel logout
func (s *OIDCService) GetLogoutURL(ctx context.Context, provider, postLogoutRedirectURI string) (string, error) {
	config, exists := s.configs[provider]
	if !exists {
		return "", domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
	}

	endSessionURL, err := s.getEndSessionEndpoint(provider)
	if err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get end session endpoint", 500)
	}

	// Build logout URL without id_token_hint (for cases where we don't have the token)
	u, err := url.Parse(endSessionURL)
	if err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to parse logout URL", 500)
	}

	params := u.Query()
	params.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	params.Set("client_id", config.ClientID)
	u.RawQuery = params.Encode()

	return u.String(), nil
}

// initializeConfigs initializes OIDC provider configurations
func (s *OIDCService) initializeConfigs() {
	// Google OAuth2
	s.configs["google"] = &OIDCConfig{
		ClientID:     "your-google-client-id",
		ClientSecret: "your-google-client-secret",
		RedirectURL:  "http://localhost:8080/auth/oidc/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Config: &oauth2.Config{
			ClientID:     "your-google-client-id",
			ClientSecret: "your-google-client-secret",
			RedirectURL:  "http://localhost:8080/auth/oidc/callback",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
	}

	// GitHub OAuth2
	s.configs["github"] = &OIDCConfig{
		ClientID:     "your-github-client-id",
		ClientSecret: "your-github-client-secret",
		RedirectURL:  "http://localhost:8080/auth/oidc/callback",
		Scopes:       []string{"user:email"},
		Config: &oauth2.Config{
			ClientID:     "your-github-client-id",
			ClientSecret: "your-github-client-secret",
			RedirectURL:  "http://localhost:8080/auth/oidc/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		},
	}

	// Azure AD OAuth2
	s.configs["azure"] = &OIDCConfig{
		ClientID:     "your-azure-client-id",
		ClientSecret: "your-azure-client-secret",
		RedirectURL:  "http://localhost:8080/auth/oidc/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Config: &oauth2.Config{
			ClientID:     "your-azure-client-id",
			ClientSecret: "your-azure-client-secret",
			RedirectURL:  "http://localhost:8080/auth/oidc/callback",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     microsoft.AzureADEndpoint("common"),
		},
	}
}

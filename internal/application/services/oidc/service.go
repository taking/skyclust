package oidc

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
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service handles OIDC authentication
type Service struct {
	userRepo         domain.UserRepository
	auditLogRepo     domain.AuditLogRepository
	authService      domain.AuthService
	cacheService     domain.CacheService
	oidcProviderRepo domain.OIDCProviderRepository
	configs          map[string]*OIDCConfig
	httpClient       *http.Client
}

// NewService creates a new OIDC service
func NewService(
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
	authService domain.AuthService,
	cacheService domain.CacheService,
	oidcProviderRepo domain.OIDCProviderRepository,
) domain.OIDCService {
	service := &Service{
		userRepo:         userRepo,
		auditLogRepo:     auditLogRepo,
		authService:      authService,
		cacheService:     cacheService,
		oidcProviderRepo: oidcProviderRepo,
		configs:          make(map[string]*OIDCConfig),
		httpClient:       &http.Client{Timeout: DefaultHTTPClientTimeout},
	}

	// Initialize OIDC configurations
	service.initializeConfigs()

	return service
}

// GetAuthURL returns the OAuth authorization URL for the specified provider
func (s *Service) GetAuthURL(ctx context.Context, provider string, state string) (string, error) {
	var config *OIDCConfig

	// Try to parse provider as UUID (user-registered provider)
	if providerID, err := uuid.Parse(provider); err == nil {
		// User-registered provider
		userProvider, err := s.oidcProviderRepo.GetByID(providerID)
		if err != nil {
			return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user provider", 500)
		}
		if userProvider == nil {
			return "", domain.NewDomainError(domain.ErrCodeNotFound, "OIDC provider not found", 404)
		}
		if !userProvider.Enabled {
			return "", domain.NewDomainError(domain.ErrCodeValidationFailed, "OIDC provider is disabled", 400)
		}

		// Create OAuth2 config from user provider
		config, err = s.createConfigFromProvider(userProvider)
		if err != nil {
			return "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create config: %v", err), 500)
		}
		provider = userProvider.ProviderType // Use provider type for state storage
	} else {
		// System provider
		var exists bool
		config, exists = s.configs[provider]
		if !exists {
			return "", domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
		}
	}

	// Store state in cache for validation (10 minutes TTL)
	stateData := OIDCState{
		Provider:  provider,
		Timestamp: time.Now(),
	}
	stateKey := fmt.Sprintf("oidc:state:%s", state)
	if err := s.cacheService.Set(ctx, stateKey, stateData, StateCacheTTL); err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to store state", 500)
	}

	// Generate auth URL
	authURL := config.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return authURL, nil
}

// ExchangeCode exchanges authorization code for access token and user info
func (s *Service) ExchangeCode(ctx context.Context, provider, code, state string) (*domain.User, string, error) {
	var config *OIDCConfig
	var userProvider *domain.OIDCProvider
	var providerType string

	// Try to parse provider as UUID (user-registered provider)
	if providerID, err := uuid.Parse(provider); err == nil {
		// User-registered provider
		userProvider, err = s.oidcProviderRepo.GetByID(providerID)
		if err != nil {
			return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user provider", 500)
		}
		if userProvider == nil {
			return nil, "", domain.NewDomainError(domain.ErrCodeNotFound, "OIDC provider not found", 404)
		}
		if !userProvider.Enabled {
			return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "OIDC provider is disabled", 400)
		}

		// Create OAuth2 config from user provider
		config, err = s.createConfigFromProvider(userProvider)
		if err != nil {
			return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create config: %v", err), 500)
		}
		providerType = userProvider.ProviderType
	} else {
		// System provider
		var exists bool
		config, exists = s.configs[provider]
		if !exists {
			return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported OIDC provider", 400)
		}
		providerType = provider
	}

	// Validate state parameter
	if err := s.validateState(ctx, state, providerType); err != nil {
		return nil, "", err
	}

	// Exchange code for token
	token, err := config.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to exchange code for token", 500)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfoFromProvider(providerType, token, userProvider)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user info", 500)
	}

	// Check if user exists
	user, err := s.userRepo.GetByOIDC(providerType, userInfo.ID)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing user", 500)
	}

	// Create user if doesn't exist
	if user == nil {
		user = &domain.User{
			Username:     userInfo.Username,
			Email:        userInfo.Email,
			OIDCProvider: providerType,
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
				"provider":      providerType,
				"provider_id":   provider,
				"user_provider": userProvider != nil,
				"username":      user.Username,
				"email":         user.Email,
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
			"provider":      providerType,
			"provider_id":   provider,
			"user_provider": userProvider != nil,
		},
	})

	// Delete state after successful exchange (prevent reuse)
	stateKey := fmt.Sprintf("oidc:state:%s", state)
	_ = s.cacheService.Delete(ctx, stateKey)

	return user, jwtToken, nil
}

// validateState validates the OIDC state parameter
func (s *Service) validateState(ctx context.Context, state, provider string) error {
	stateKey := fmt.Sprintf("oidc:state:%s", state)

	// Get state from cache
	value, err := s.cacheService.Get(ctx, stateKey)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid or expired state parameter", 401)
	}

	// Type assert to OIDCState
	stateData, ok := value.(OIDCState)
	if !ok {
		// Try to unmarshal if it's stored as map
		if stateMap, ok := value.(map[string]interface{}); ok {
			stateData = OIDCState{
				Provider:  stateMap["provider"].(string),
				Timestamp: stateMap["timestamp"].(time.Time),
			}
		} else {
			return domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid state data format", 401)
		}
	}

	// Validate provider matches
	if stateData.Provider != provider {
		return domain.NewDomainError(domain.ErrCodeUnauthorized, "state provider mismatch", 401)
	}

	// Validate state is not too old (should be within 10 minutes)
	if time.Since(stateData.Timestamp) > StateMaxAge {
		_ = s.cacheService.Delete(ctx, stateKey)
		return domain.NewDomainError(domain.ErrCodeUnauthorized, "state parameter expired", 401)
	}

	return nil
}

// getUserInfoFromProvider fetches user information from the OIDC provider
func (s *Service) getUserInfoFromProvider(providerType string, token *oauth2.Token, userProvider *domain.OIDCProvider) (*OIDCUserInfo, error) {
	var client *http.Client
	var apiURL string

	if userProvider != nil && userProvider.IsCustomProvider() {
		// Custom provider - use user-provided endpoint
		if userProvider.UserInfoURL == "" {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "user_info_url is required for custom providers", 400)
		}
		apiURL = userProvider.UserInfoURL
		// Create OAuth2 client with custom config
		config, err := s.createConfigFromProvider(userProvider)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create config: %v", err), 500)
		}
		client = config.Config.Client(context.Background(), token)
	} else {
		// System provider - use predefined endpoints
		config, exists := s.configs[providerType]
		if !exists {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", providerType), 400)
		}
		client = config.Config.Client(context.Background(), token)

		switch providerType {
		case "google":
			apiURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		case "github":
			apiURL = "https://api.github.com/user"
		case "azure", "microsoft":
			apiURL = "https://graph.microsoft.com/v1.0/me"
		default:
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", providerType), 400)
		}
	}

	var userInfo OIDCUserInfo

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get user info: status %d", resp.StatusCode), 502)
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// createConfigFromProvider creates an OAuth2 config from a user-registered OIDC provider
func (s *Service) createConfigFromProvider(provider *domain.OIDCProvider) (*OIDCConfig, error) {
	// Parse scopes
	var scopes []string
	if provider.Scopes != "" {
		scopes = []string{}
		scopeParts := strings.Split(provider.Scopes, ",")
		for _, part := range scopeParts {
			scope := strings.TrimSpace(part)
			if scope != "" {
				scopes = append(scopes, scope)
			}
		}
	}

	var endpoint oauth2.Endpoint

	if provider.IsCustomProvider() {
		// Custom provider - use custom endpoints
		if provider.AuthURL == "" || provider.TokenURL == "" {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "auth_url and token_url are required for custom providers", 400)
		}
		endpoint = oauth2.Endpoint{
			AuthURL:  provider.AuthURL,
			TokenURL: provider.TokenURL,
		}
	} else {
		// System provider - use predefined endpoints
		switch provider.ProviderType {
		case "google":
			endpoint = google.Endpoint
		case "github":
			endpoint = github.Endpoint
		case "azure", "microsoft":
			endpoint = microsoft.AzureADEndpoint("common")
		default:
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider type: %s", provider.ProviderType), 400)
		}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}

	return &OIDCConfig{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		Scopes:       scopes,
		Config:       oauthConfig,
	}, nil
}

// EndSession initiates OIDC logout by calling the provider's end_session_endpoint
func (s *Service) EndSession(ctx context.Context, userID uuid.UUID, provider, idToken, postLogoutRedirectURI string) error {
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
func (s *Service) getEndSessionEndpoint(provider string) (string, error) {
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
		return "", domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", provider), 400)
	}
}

// buildLogoutURL constructs the logout URL with required parameters
func (s *Service) buildLogoutURL(endSessionURL, idToken, postLogoutRedirectURI, clientID string) (string, error) {
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
func (s *Service) HandleBackChannelLogout(ctx context.Context, logoutToken string) error {
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
func (s *Service) GetLogoutURL(ctx context.Context, provider, postLogoutRedirectURI string) (string, error) {
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
func (s *Service) initializeConfigs() {
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

// CreateProvider creates a new OIDC provider for a user
func (s *Service) CreateProvider(ctx context.Context, userID uuid.UUID, provider *domain.OIDCProvider) (*domain.OIDCProvider, error) {
	// Check if provider name already exists for this user
	existing, err := s.oidcProviderRepo.GetByUserIDAndName(userID, provider.Name)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check existing provider: %v", err), 500)
	}
	if existing != nil {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "provider with this name already exists", 400)
	}

	// Set user ID
	provider.UserID = userID

	// Create provider (encryption is handled in repository)
	if err := s.oidcProviderRepo.Create(provider); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create provider: %v", err), 500)
	}

	return provider, nil
}

// GetUserProviders retrieves all OIDC providers for a user
func (s *Service) GetUserProviders(ctx context.Context, userID uuid.UUID) ([]*domain.OIDCProvider, error) {
	providers, err := s.oidcProviderRepo.GetByUserID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user providers: %v", err), 500)
	}

	return providers, nil
}

// GetProvider retrieves a specific OIDC provider for a user
func (s *Service) GetProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) (*domain.OIDCProvider, error) {
	provider, err := s.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get provider: %v", err), 500)
	}
	if provider == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404)
	}

	// Verify ownership
	if provider.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403)
	}

	return provider, nil
}

// UpdateProvider updates an OIDC provider for a user
func (s *Service) UpdateProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, req *domain.OIDCProvider) (*domain.OIDCProvider, error) {
	// Get existing provider
	provider, err := s.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get provider: %v", err), 500)
	}
	if provider == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404)
	}

	// Verify ownership
	if provider.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403)
	}

	// Check name uniqueness if name is being updated
	if req.Name != "" && req.Name != provider.Name {
		existing, err := s.oidcProviderRepo.GetByUserIDAndName(userID, req.Name)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check existing provider: %v", err), 500)
		}
		if existing != nil && existing.ID != providerID {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "provider with this name already exists", 400)
		}
		provider.Name = req.Name
	}

	// Update fields (encryption is handled in repository)
	if req.ClientID != "" {
		provider.ClientID = req.ClientID
	}
	if req.ClientSecret != "" {
		provider.ClientSecret = req.ClientSecret // Will be encrypted in repository
	}
	if req.RedirectURL != "" {
		provider.RedirectURL = req.RedirectURL
	}
	if req.AuthURL != "" {
		provider.AuthURL = req.AuthURL
	}
	if req.TokenURL != "" {
		provider.TokenURL = req.TokenURL
	}
	if req.UserInfoURL != "" {
		provider.UserInfoURL = req.UserInfoURL
	}
	if req.Scopes != "" {
		provider.Scopes = req.Scopes
	}
	if req.Enabled != provider.Enabled {
		provider.Enabled = req.Enabled
	}

	if err := s.oidcProviderRepo.Update(provider); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update provider: %v", err), 500)
	}

	return provider, nil
}

// DeleteProvider deletes an OIDC provider for a user
func (s *Service) DeleteProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) error {
	// Get existing provider
	provider, err := s.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get provider: %v", err), 500)
	}
	if provider == nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404)
	}

	// Verify ownership
	if provider.UserID != userID {
		return domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403)
	}

	if err := s.oidcProviderRepo.Delete(providerID); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete provider: %v", err), 500)
	}

	return nil
}

package oidc

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles OIDC authentication requests
type Handler struct {
	*handlers.BaseHandler
	oidcService       domain.OIDCService
	oidcProviderRepo  domain.OIDCProviderRepository
}

// NewHandler creates a new OIDC handler
func NewHandler(oidcService domain.OIDCService) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("oidc"),
		oidcService: oidcService,
	}
}

// SetOIDCProviderRepository sets the OIDC provider repository (for dependency injection)
func (h *Handler) SetOIDCProviderRepository(repo domain.OIDCProviderRepository) {
	h.oidcProviderRepo = repo
}

// GetAuthURL returns the OAuth authorization URL
func (h *Handler) GetAuthURL(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_auth_url", 200)

	provider := c.Param("provider")

	// Log operation start
	h.LogInfo(c, "Getting OIDC auth URL",
		zap.String("operation", "get_auth_url"),
		zap.String("provider", provider))

	if provider == "" {
		h.LogWarn(c, "Provider is required")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider is required", 400), "get_auth_url")
		return
	}

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		h.LogError(c, err, "Failed to generate state parameter")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Failed to generate state parameter", 500), "get_auth_url")
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state in session or cache for validation
	// State is stored in cache by OIDCService.GetAuthURL
	h.LogInfo(c, "Generated state parameter",
		zap.String("state", state))

	// Log business event
	h.LogBusinessEvent(c, "oidc_auth_url_requested", "", "", map[string]interface{}{
		"provider": provider,
		"state":    state,
	})

	authURL, err := h.oidcService.GetAuthURL(c.Request.Context(), provider, state)
	if err != nil {
		h.LogError(c, err, "Failed to get auth URL")
		h.HandleError(c, err, "get_auth_url")
		return
	}

	h.LogInfo(c, "Auth URL generated successfully",
		zap.String("provider", provider))

	h.OK(c, gin.H{
		"auth_url": authURL,
		"state":    state,
	}, "Auth URL generated successfully")
}

// Callback handles OAuth callback
func (h *Handler) Callback(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "oidc_callback", 200)

	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	// Log operation start
	h.LogInfo(c, "Processing OIDC callback",
		zap.String("operation", "oidc_callback"),
		zap.String("provider", provider))

	if provider == "" || code == "" {
		h.LogWarn(c, "Provider and code are required",
			zap.String("provider", provider),
			zap.String("code", code))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider and code are required", 400), "oidc_callback")
		return
	}

	// State validation is handled in OIDCService.ExchangeCode
	h.LogInfo(c, "Processing OIDC callback with state",
		zap.String("state", state))

	// Log business event
	h.LogBusinessEvent(c, "oidc_callback_received", "", "", map[string]interface{}{
		"provider": provider,
		"state":    state,
	})

	user, token, err := h.oidcService.ExchangeCode(c.Request.Context(), provider, code, state)
	if err != nil {
		h.LogError(c, err, "Failed to process OIDC callback")
		h.HandleError(c, err, "oidc_callback")
		return
	}

	// Log successful authentication
	h.LogBusinessEvent(c, "oidc_authentication_successful", user.ID.String(), "", map[string]interface{}{
		"provider": provider,
		"user_id":  user.ID.String(),
	})

	h.LogInfo(c, "OIDC authentication successful",
		zap.String("provider", provider),
		zap.String("user_id", user.ID.String()))

	h.OK(c, gin.H{
		"user":  user,
		"token": token,
	}, "Authentication successful")
}

// Login handles OIDC login
func (h *Handler) Login(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "oidc_login", 200)

	// Log operation start
	h.LogInfo(c, "Processing OIDC login",
		zap.String("operation", "oidc_login"))

	var req struct {
		Provider string `json:"provider" binding:"required"`
		Code     string `json:"code" binding:"required"`
		State    string `json:"state" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "oidc_callback")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_login_attempted", "", "", map[string]interface{}{
		"provider": req.Provider,
		"state":    req.State,
	})

	user, token, err := h.oidcService.ExchangeCode(c.Request.Context(), req.Provider, req.Code, req.State)
	if err != nil {
		h.LogError(c, err, "Failed to process OIDC login")
		h.HandleError(c, err, "oidc_login")
		return
	}

	// Log successful login
	h.LogBusinessEvent(c, "oidc_login_successful", user.ID.String(), "", map[string]interface{}{
		"provider": req.Provider,
		"user_id":  user.ID.String(),
	})

	h.LogInfo(c, "OIDC login successful",
		zap.String("provider", req.Provider),
		zap.String("user_id", user.ID.String()))

	h.OK(c, gin.H{
		"user":  user,
		"token": token,
	}, "Login successful")
}

// Logout handles OIDC logout
func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		UserID                string `json:"user_id" binding:"required"`
		Provider              string `json:"provider" binding:"required"`
		IDToken               string `json:"id_token" binding:"required"`
		PostLogoutRedirectURI string `json:"post_logout_redirect_uri"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "logout")
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid user ID format", 400), "logout")
		return
	}

	err = h.oidcService.EndSession(c.Request.Context(), userID, req.Provider, req.IDToken, req.PostLogoutRedirectURI)
	if err != nil {
		h.HandleError(c, err, "logout")
		return
	}

	h.OK(c, gin.H{"message": "Logout successful"}, "Logout successful")
}

// GetLogoutURL returns the OAuth logout URL
func (h *Handler) GetLogoutURL(c *gin.Context) {
	provider := c.Param("provider")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")

	if provider == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider is required", 400), "get_logout_url")
		return
	}

	logoutURL, err := h.oidcService.GetLogoutURL(c.Request.Context(), provider, postLogoutRedirectURI)
	if err != nil {
		h.HandleError(c, err, "get_logout_url")
		return
	}

	h.OK(c, gin.H{
		"logout_url": logoutURL,
	}, "Logout URL generated successfully")
}

// GetProviders returns available system OIDC providers (public)
// This returns the list of supported provider types that users can register
func (h *Handler) GetProviders(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_oidc_providers", 200)

	// Log operation start
	h.LogInfo(c, "Getting system OIDC providers",
		zap.String("operation", "get_oidc_providers"))

	// Log business event
	h.LogBusinessEvent(c, "oidc_providers_requested", "", "", map[string]interface{}{
		"operation": "get_system_oidc_providers",
	})

	// Return list of supported provider types
	// These are the provider types that the system supports
	// Users can register their own OIDC providers with these types
	providers := []string{"google", "github", "azure", "microsoft", "custom"}

	h.LogInfo(c, "System OIDC providers retrieved successfully",
		zap.Int("count", len(providers)))

	h.OK(c, gin.H{
		"providers": providers,
	}, "System OIDC providers retrieved successfully")
}

// CreateProvider creates a new OIDC provider for the user
func (h *Handler) CreateProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "create_oidc_provider", 201)

	// Log operation start
	h.LogInfo(c, "Creating OIDC provider",
		zap.String("operation", "create_oidc_provider"))

	userID := h.extractUserIDFromContext(c)
	if userID == uuid.Nil {
		return
	}

	var req CreateProviderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_oidc_provider")
		return
	}

	if h.oidcProviderRepo == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "OIDC provider repository not initialized", 500), "create_oidc_provider")
		return
	}

	// Check if provider name already exists for this user
	existing, err := h.oidcProviderRepo.GetByUserIDAndName(userID, req.Name)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing provider", 500), "create_oidc_provider")
		return
	}
	if existing != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeValidationFailed, "provider with this name already exists", 400), "create_oidc_provider")
		return
	}

	// Create provider (encryption is handled in repository)
	provider := &domain.OIDCProvider{
		UserID:       userID,
		Name:         req.Name,
		ProviderType: req.ProviderType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret, // Will be encrypted in repository
		RedirectURL:  req.RedirectURL,
		AuthURL:      req.AuthURL,
		TokenURL:     req.TokenURL,
		UserInfoURL:  req.UserInfoURL,
		Scopes:       req.Scopes,
		Enabled:      req.Enabled,
	}

	if err := h.oidcProviderRepo.Create(provider); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create provider: %v", err), 500), "create_oidc_provider")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_provider_created", userID.String(), provider.ID.String(), map[string]interface{}{
		"provider_id":   provider.ID.String(),
		"provider_name": provider.Name,
		"provider_type": provider.ProviderType,
	})

	h.LogInfo(c, "OIDC provider created successfully",
		zap.String("provider_id", provider.ID.String()),
		zap.String("provider_name", provider.Name))

	h.Created(c, gin.H{
		"id":           provider.ID.String(),
		"name":         provider.Name,
		"provider_type": provider.ProviderType,
		"enabled":      provider.Enabled,
		"created_at":   provider.CreatedAt,
	}, "OIDC provider created successfully")
}

// GetUserProviders retrieves all OIDC providers for the current user
func (h *Handler) GetUserProviders(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_user_oidc_providers", 200)

	// Log operation start
	h.LogInfo(c, "Getting user OIDC providers",
		zap.String("operation", "get_user_oidc_providers"))

	userID := h.extractUserIDFromContext(c)
	if userID == uuid.Nil {
		return
	}

	if h.oidcProviderRepo == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "OIDC provider repository not initialized", 500), "get_user_oidc_providers")
		return
	}

	providers, err := h.oidcProviderRepo.GetByUserID(userID)
	if err != nil {
		h.LogError(c, err, "Failed to get user OIDC providers from repository",
			zap.String("user_id", userID.String()))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get providers: %v", err), 500), "get_user_oidc_providers")
		return
	}

	// Convert to response format (without sensitive data)
	response := make([]gin.H, 0, len(providers))
	for _, p := range providers {
		response = append(response, gin.H{
			"id":           p.ID.String(),
			"name":         p.Name,
			"provider_type": p.ProviderType,
			"redirect_url": p.RedirectURL,
			"enabled":      p.Enabled,
			"created_at":   p.CreatedAt,
			"updated_at":   p.UpdatedAt,
		})
	}

	h.LogInfo(c, "User OIDC providers retrieved successfully",
		zap.Int("count", len(providers)))

	h.OK(c, gin.H{
		"providers": response,
		"total":     len(response),
	}, "OIDC providers retrieved successfully")
}

// GetProvider retrieves a specific OIDC provider
func (h *Handler) GetProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Getting OIDC provider",
		zap.String("operation", "get_oidc_provider"))

	userID := h.extractUserIDFromContext(c)
	if userID == uuid.Nil {
		return
	}

	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid provider ID", 400), "get_oidc_provider")
		return
	}

	if h.oidcProviderRepo == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "OIDC provider repository not initialized", 500), "get_oidc_provider")
		return
	}

	provider, err := h.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get provider", 500), "get_oidc_provider")
		return
	}
	if provider == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404), "get_oidc_provider")
		return
	}

	// Verify ownership
	if provider.UserID != userID {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403), "get_oidc_provider")
		return
	}

	h.OK(c, gin.H{
		"id":           provider.ID.String(),
		"name":         provider.Name,
		"provider_type": provider.ProviderType,
		"client_id":    provider.ClientID,
		"redirect_url": provider.RedirectURL,
		"auth_url":     provider.AuthURL,
		"token_url":    provider.TokenURL,
		"user_info_url": provider.UserInfoURL,
		"scopes":       provider.Scopes,
		"enabled":      provider.Enabled,
		"created_at":   provider.CreatedAt,
		"updated_at":   provider.UpdatedAt,
	}, "OIDC provider retrieved successfully")
}

// UpdateProvider updates an OIDC provider
func (h *Handler) UpdateProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "update_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Updating OIDC provider",
		zap.String("operation", "update_oidc_provider"))

	userID := h.extractUserIDFromContext(c)
	if userID == uuid.Nil {
		return
	}

	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid provider ID", 400), "update_oidc_provider")
		return
	}

	if h.oidcProviderRepo == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "OIDC provider repository not initialized", 500), "update_oidc_provider")
		return
	}

	provider, err := h.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get provider", 500), "update_oidc_provider")
		return
	}
	if provider == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404), "update_oidc_provider")
		return
	}

	// Verify ownership
	if provider.UserID != userID {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403), "update_oidc_provider")
		return
	}

	var req UpdateProviderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_oidc_provider")
		return
	}

	// Check name uniqueness if name is being updated
	if req.Name != "" && req.Name != provider.Name {
		existing, err := h.oidcProviderRepo.GetByUserIDAndName(userID, req.Name)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing provider", 500), "update_oidc_provider")
			return
		}
		if existing != nil && existing.ID != provider.ID {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeValidationFailed, "provider with this name already exists", 400), "update_oidc_provider")
			return
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
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}

	if err := h.oidcProviderRepo.Update(provider); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update provider: %v", err), 500), "update_oidc_provider")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_provider_updated", userID.String(), provider.ID.String(), map[string]interface{}{
		"provider_id": provider.ID.String(),
	})

	h.LogInfo(c, "OIDC provider updated successfully",
		zap.String("provider_id", provider.ID.String()))

	h.OK(c, gin.H{
		"id":           provider.ID.String(),
		"name":         provider.Name,
		"provider_type": provider.ProviderType,
		"enabled":      provider.Enabled,
		"updated_at":   provider.UpdatedAt,
	}, "OIDC provider updated successfully")
}

// DeleteProvider deletes an OIDC provider
func (h *Handler) DeleteProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "delete_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Deleting OIDC provider",
		zap.String("operation", "delete_oidc_provider"))

	userID := h.extractUserIDFromContext(c)
	if userID == uuid.Nil {
		return
	}

	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid provider ID", 400), "delete_oidc_provider")
		return
	}

	if h.oidcProviderRepo == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "OIDC provider repository not initialized", 500), "delete_oidc_provider")
		return
	}

	provider, err := h.oidcProviderRepo.GetByID(providerID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get provider", 500), "delete_oidc_provider")
		return
	}
	if provider == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "provider not found", 404), "delete_oidc_provider")
		return
	}

	// Verify ownership
	if provider.UserID != userID {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "you don't have access to this provider", 403), "delete_oidc_provider")
		return
	}

	if err := h.oidcProviderRepo.Delete(providerID); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to delete provider", 500), "delete_oidc_provider")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_provider_deleted", userID.String(), providerID.String(), map[string]interface{}{
		"provider_id": providerID.String(),
	})

	h.LogInfo(c, "OIDC provider deleted successfully",
		zap.String("provider_id", providerID.String()))

	h.OK(c, gin.H{
		"message": "OIDC provider deleted successfully",
	}, "OIDC provider deleted successfully")
}

// extractUserIDFromContext extracts user ID from context
func (h *Handler) extractUserIDFromContext(c *gin.Context) uuid.UUID {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}

	// Convert to uuid.UUID (handle both string and uuid.UUID types)
	switch v := userIDValue.(type) {
	case uuid.UUID:
		return v
	case string:
		parsedUserID, err := uuid.Parse(v)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID format", 401), "extract_user_id")
			return uuid.Nil
		}
		return parsedUserID
	default:
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID type", 401), "extract_user_id")
		return uuid.Nil
	}
}

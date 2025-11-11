package oidc

import (
	"crypto/rand"
	"encoding/hex"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler: OIDC 인증 요청을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	oidcService domain.OIDCService
}

// NewHandler: 새로운 OIDC 핸들러를 생성합니다
func NewHandler(oidcService domain.OIDCService) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("oidc"),
		oidcService: oidcService,
	}
}

// GetAuthURL: OAuth 인증 URL을 반환합니다
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

// Callback: OAuth 콜백을 처리합니다
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

	ctx := h.EnrichContextWithRequestMetadata(c)
	user, token, err := h.oidcService.ExchangeCode(ctx, provider, code, state)
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

// CreateSession: OIDC 세션 생성을 처리합니다 (POST /auth/oidc/sessions)
// RESTful: 새로운 OIDC 세션 생성
func (h *Handler) CreateSession(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "oidc_create_session", 201)

	// Log operation start
	h.LogInfo(c, "Creating OIDC session",
		zap.String("operation", "oidc_create_session"))

	var req struct {
		Provider string `json:"provider" binding:"required"`
		Code     string `json:"code" binding:"required"`
		State    string `json:"state" binding:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "oidc_create_session")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_session_creation_attempted", "", "", map[string]interface{}{
		"provider": req.Provider,
		"state":    req.State,
	})

	ctx := h.EnrichContextWithRequestMetadata(c)
	user, token, err := h.oidcService.ExchangeCode(ctx, req.Provider, req.Code, req.State)
	if err != nil {
		h.LogError(c, err, "Failed to create OIDC session")
		h.HandleError(c, err, "oidc_create_session")
		return
	}

	// Log successful session creation
	h.LogBusinessEvent(c, "oidc_session_created", user.ID.String(), "", map[string]interface{}{
		"provider": req.Provider,
		"user_id":  user.ID.String(),
	})

	h.LogInfo(c, "OIDC session created successfully",
		zap.String("provider", req.Provider),
		zap.String("user_id", user.ID.String()))

	h.Created(c, gin.H{
		"user":  user,
		"token": token,
	}, "Session created successfully")
}

// DeleteSession: OIDC 세션 삭제를 처리합니다 (DELETE /auth/oidc/sessions/me)
// RESTful: 현재 OIDC 세션 삭제
func (h *Handler) DeleteSession(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "oidc_delete_session", 200)

	// Log operation start
	h.LogInfo(c, "Deleting OIDC session",
		zap.String("operation", "oidc_delete_session"))

	var req struct {
		Provider              string `json:"provider" binding:"required"`
		IDToken               string `json:"id_token" binding:"required"`
		PostLogoutRedirectURI string `json:"post_logout_redirect_uri"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_session_deletion_attempted", userID.String(), "", map[string]interface{}{
		"provider": req.Provider,
	})

	ctx := h.EnrichContextWithRequestMetadata(c)
	err = h.oidcService.EndSession(ctx, userID, req.Provider, req.IDToken, req.PostLogoutRedirectURI)
	if err != nil {
		h.LogError(c, err, "Failed to delete OIDC session")
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	// Log successful session deletion
	h.LogBusinessEvent(c, "oidc_session_deleted", userID.String(), "", map[string]interface{}{
		"provider": req.Provider,
	})

	h.LogInfo(c, "OIDC session deleted successfully",
		zap.String("provider", req.Provider),
		zap.String("user_id", userID.String()))

	h.OK(c, gin.H{
		"message": "Session terminated successfully",
	}, "Session deleted successfully")
}

// GetLogoutURL returns the OAuth logout URL (GET /auth/oidc/:provider/logout-url)
// RESTful: Get logout URL for a provider
func (h *Handler) GetLogoutURL(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "oidc_get_logout_url", 200)

	provider := c.Param("provider")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")

	if provider == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider is required", 400), "oidc_get_logout_url")
		return
	}

	// Log operation start
	h.LogInfo(c, "Getting OIDC logout URL",
		zap.String("operation", "oidc_get_logout_url"),
		zap.String("provider", provider))

	logoutURL, err := h.oidcService.GetLogoutURL(c.Request.Context(), provider, postLogoutRedirectURI)
	if err != nil {
		h.LogError(c, err, "Failed to get OIDC logout URL")
		h.HandleError(c, err, "oidc_get_logout_url")
		return
	}

	h.LogInfo(c, "OIDC logout URL generated successfully",
		zap.String("provider", provider))

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

	// Always include meta information for consistency (direct array: data[])
	page, limit := h.ParsePageLimitParams(c)
	total := int64(len(providers))
	h.BuildPaginatedResponse(c, providers, page, limit, total, "System OIDC providers retrieved successfully")
}

// CreateProvider creates a new OIDC provider for the user
func (h *Handler) CreateProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "create_oidc_provider", 201)

	// Log operation start
	h.LogInfo(c, "Creating OIDC provider",
		zap.String("operation", "create_oidc_provider"))

	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	var req CreateProviderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_oidc_provider")
		return
	}

	// Create provider domain object
	provider := &domain.OIDCProvider{
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

	// Create provider via service
	provider, err = h.oidcService.CreateProvider(c.Request.Context(), userID, provider)
	if err != nil {
		h.HandleError(c, err, "create_oidc_provider")
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
		"id":            provider.ID.String(),
		"name":          provider.Name,
		"provider_type": provider.ProviderType,
		"enabled":       provider.Enabled,
		"created_at":    provider.CreatedAt,
	}, "OIDC provider created successfully")
}

// GetUserProviders retrieves all OIDC providers for the current user
func (h *Handler) GetUserProviders(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_user_oidc_providers", 200)

	// Log operation start
	h.LogInfo(c, "Getting user OIDC providers",
		zap.String("operation", "get_user_oidc_providers"))

	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "get_user_oidc_providers")
		return
	}

	providers, err := h.oidcService.GetUserProviders(c.Request.Context(), userID)
	if err != nil {
		h.LogError(c, err, "Failed to get user OIDC providers",
			zap.String("user_id", userID.String()))
		h.HandleError(c, err, "get_user_oidc_providers")
		return
	}

	// Convert to response format (without sensitive data)
	response := make([]gin.H, 0, len(providers))
	for _, p := range providers {
		response = append(response, gin.H{
			"id":            p.ID.String(),
			"name":          p.Name,
			"provider_type": p.ProviderType,
			"redirect_url":  p.RedirectURL,
			"enabled":       p.Enabled,
			"created_at":    p.CreatedAt,
			"updated_at":    p.UpdatedAt,
		})
	}

	h.LogInfo(c, "User OIDC providers retrieved successfully",
		zap.Int("count", len(providers)))

	// Always include meta information for consistency (direct array: data[])
	page, limit := h.ParsePageLimitParams(c)
	total := int64(len(response))
	h.BuildPaginatedResponse(c, response, page, limit, total, "OIDC providers retrieved successfully")
}

// GetProvider retrieves a specific OIDC provider
func (h *Handler) GetProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Getting OIDC provider",
		zap.String("operation", "get_oidc_provider"))

	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	providerID, err := h.ExtractPathParam(c, "id")
	if err != nil {
		h.HandleError(c, err, "get_oidc_provider")
		return
	}

	provider, err := h.oidcService.GetProvider(c.Request.Context(), userID, providerID)
	if err != nil {
		h.HandleError(c, err, "get_oidc_provider")
		return
	}

	h.OK(c, gin.H{
		"id":            provider.ID.String(),
		"name":          provider.Name,
		"provider_type": provider.ProviderType,
		"client_id":     provider.ClientID,
		"redirect_url":  provider.RedirectURL,
		"auth_url":      provider.AuthURL,
		"token_url":     provider.TokenURL,
		"user_info_url": provider.UserInfoURL,
		"scopes":        provider.Scopes,
		"enabled":       provider.Enabled,
		"created_at":    provider.CreatedAt,
		"updated_at":    provider.UpdatedAt,
	}, "OIDC provider retrieved successfully")
}

// UpdateProvider updates an OIDC provider
func (h *Handler) UpdateProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "update_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Updating OIDC provider",
		zap.String("operation", "update_oidc_provider"))

	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "oidc_delete_session")
		return
	}

	providerID, err := h.ExtractPathParam(c, "id")
	if err != nil {
		h.HandleError(c, err, "update_oidc_provider")
		return
	}

	var req UpdateProviderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_oidc_provider")
		return
	}

	// Create update request domain object
	updateProvider := &domain.OIDCProvider{
		Name:         req.Name,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret, // Will be encrypted in repository
		RedirectURL:  req.RedirectURL,
		AuthURL:      req.AuthURL,
		TokenURL:     req.TokenURL,
		UserInfoURL:  req.UserInfoURL,
		Scopes:       req.Scopes,
	}
	if req.Enabled != nil {
		updateProvider.Enabled = *req.Enabled
	}

	provider, err := h.oidcService.UpdateProvider(c.Request.Context(), userID, providerID, updateProvider)
	if err != nil {
		h.HandleError(c, err, "update_oidc_provider")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "oidc_provider_updated", userID.String(), provider.ID.String(), map[string]interface{}{
		"provider_id": provider.ID.String(),
	})

	h.LogInfo(c, "OIDC provider updated successfully",
		zap.String("provider_id", provider.ID.String()))

	h.OK(c, gin.H{
		"id":            provider.ID.String(),
		"name":          provider.Name,
		"provider_type": provider.ProviderType,
		"enabled":       provider.Enabled,
		"updated_at":    provider.UpdatedAt,
	}, "OIDC provider updated successfully")
}

// DeleteProvider deletes an OIDC provider
func (h *Handler) DeleteProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "delete_oidc_provider", 200)

	// Log operation start
	h.LogInfo(c, "Deleting OIDC provider",
		zap.String("operation", "delete_oidc_provider"))

	userID, err := h.ExtractUserIDFromContext(c)
	if err != nil {
		h.HandleError(c, err, "delete_oidc_provider")
		return
	}

	providerID, err := h.ExtractPathParam(c, "id")
	if err != nil {
		h.HandleError(c, err, "delete_oidc_provider")
		return
	}

	if err := h.oidcService.DeleteProvider(c.Request.Context(), userID, providerID); err != nil {
		h.HandleError(c, err, "delete_oidc_provider")
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

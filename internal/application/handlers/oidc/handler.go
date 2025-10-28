package oidc

import (
	"crypto/rand"
	"encoding/hex"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles OIDC authentication requests
type Handler struct {
	*handlers.BaseHandler
	oidcService domain.OIDCService
}

// NewHandler creates a new OIDC handler
func NewHandler(oidcService domain.OIDCService) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("oidc"),
		oidcService: oidcService,
	}
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
	// TODO: Implement state storage
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

	// TODO: Validate state parameter
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

	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid request body", 400), "logout")
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

// GetProviders returns available OIDC providers
func (h *Handler) GetProviders(c *gin.Context) {
	// TODO: Implement proper OIDC provider listing
	// For now, return a list of supported providers
	providers := []string{"google", "github", "azure", "microsoft"}

	h.OK(c, gin.H{
		"providers": providers,
	}, "Providers retrieved successfully")
}

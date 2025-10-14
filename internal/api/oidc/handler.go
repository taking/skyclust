package oidc

import (
	"crypto/rand"
	"encoding/hex"
	"skyclust/internal/api/common"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles OIDC authentication requests
type Handler struct {
	oidcService domain.OIDCService
}

// NewHandler creates a new OIDC handler
func NewHandler(oidcService domain.OIDCService) *Handler {
	return &Handler{
		oidcService: oidcService,
	}
}

// GetAuthURL returns the OAuth authorization URL
func (h *Handler) GetAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		common.BadRequest(c, "Provider is required")
		return
	}

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		common.InternalServerError(c, "Failed to generate state parameter")
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state in session or cache for validation
	// TODO: Implement state storage

	authURL, err := h.oidcService.GetAuthURL(c.Request.Context(), provider, state)
	if err != nil {
		common.InternalServerError(c, "Failed to get auth URL")
		return
	}

	common.OK(c, gin.H{
		"auth_url": authURL,
		"state":    state,
	}, "Auth URL generated successfully")
}

// Callback handles OAuth callback
func (h *Handler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if provider == "" || code == "" {
		common.BadRequest(c, "Provider and code are required")
		return
	}

	// TODO: Validate state parameter

	user, token, err := h.oidcService.ExchangeCode(c.Request.Context(), provider, code, state)
	if err != nil {
		common.InternalServerError(c, "Failed to process callback")
		return
	}

	common.OK(c, gin.H{
		"user":  user,
		"token": token,
	}, "Authentication successful")
}

// Login handles OIDC login
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		Code     string `json:"code" binding:"required"`
		State    string `json:"state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	user, token, err := h.oidcService.ExchangeCode(c.Request.Context(), req.Provider, req.Code, req.State)
	if err != nil {
		common.InternalServerError(c, "Failed to login")
		return
	}

	common.OK(c, gin.H{
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
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		common.BadRequest(c, "Invalid user ID format")
		return
	}

	err = h.oidcService.EndSession(c.Request.Context(), userID, req.Provider, req.IDToken, req.PostLogoutRedirectURI)
	if err != nil {
		common.InternalServerError(c, "Failed to logout")
		return
	}

	common.OK(c, gin.H{"message": "Logout successful"}, "Logout successful")
}

// GetLogoutURL returns the OAuth logout URL
func (h *Handler) GetLogoutURL(c *gin.Context) {
	provider := c.Param("provider")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")

	if provider == "" {
		common.BadRequest(c, "Provider is required")
		return
	}

	logoutURL, err := h.oidcService.GetLogoutURL(c.Request.Context(), provider, postLogoutRedirectURI)
	if err != nil {
		common.InternalServerError(c, "Failed to get logout URL")
		return
	}

	common.OK(c, gin.H{
		"logout_url": logoutURL,
	}, "Logout URL generated successfully")
}

// GetProviders returns available OIDC providers
func (h *Handler) GetProviders(c *gin.Context) {
	// TODO: Implement proper OIDC provider listing
	// For now, return a list of supported providers
	providers := []string{"google", "github", "azure", "microsoft"}

	common.OK(c, gin.H{
		"providers": providers,
	}, "Providers retrieved successfully")
}

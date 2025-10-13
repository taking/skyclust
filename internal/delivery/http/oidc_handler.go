package http

import (
	"skyclust/internal/domain"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OIDCHandler handles OIDC authentication requests
type OIDCHandler struct {
	oidcService domain.OIDCService
}

// NewOIDCHandler creates a new OIDC handler
func NewOIDCHandler(oidcService domain.OIDCService) *OIDCHandler {
	return &OIDCHandler{
		oidcService: oidcService,
	}
}

// GetAuthURL returns the OAuth authorization URL
func (h *OIDCHandler) GetAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		BadRequestResponse(c, "Provider is required")
		return
	}

	// Generate random state for CSRF protection
	state, err := h.generateState()
	if err != nil {
		InternalServerErrorResponse(c, "Failed to generate state")
		return
	}

	// Store state in session or cache (for production, use Redis)
	c.SetCookie("oidc_state", state, 600, "/", "", false, true)

	// Get authorization URL
	authURL, err := h.oidcService.GetAuthURL(c.Request.Context(), provider, state)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Unsupported provider")
			return
		}
		InternalServerErrorResponse(c, "Failed to get auth URL")
		return
	}

	OKResponse(c, gin.H{
		"auth_url": authURL,
		"provider": provider,
		"state":    state,
	}, "Authorization URL generated successfully")
}

// Callback handles the OAuth callback
func (h *OIDCHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if provider == "" || code == "" || state == "" {
		BadRequestResponse(c, "Missing required parameters")
		return
	}

	// Verify state parameter for CSRF protection
	cookieState, err := c.Cookie("oidc_state")
	if err != nil || cookieState != state {
		BadRequestResponse(c, "Invalid state parameter")
		return
	}

	// Clear state cookie
	c.SetCookie("oidc_state", "", -1, "/", "", false, true)

	// Exchange code for user and token
	user, token, err := h.oidcService.ExchangeCode(c.Request.Context(), provider, code, state)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Invalid request")
			return
		}
		InternalServerErrorResponse(c, "Failed to exchange code")
		return
	}

	// Redirect to frontend with token
	// In production, you might want to redirect to a specific frontend URL
	frontendURL := "http://localhost:3000/auth/callback"
	redirectURL := frontendURL + "?token=" + token + "&user_id=" + user.ID.String()

	c.Redirect(http.StatusFound, redirectURL)
}

// Login initiates OIDC login
func (h *OIDCHandler) Login(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Generate random state for CSRF protection
	state, err := h.generateState()
	if err != nil {
		InternalServerErrorResponse(c, "Failed to generate state")
		return
	}

	// Store state in session or cache
	c.SetCookie("oidc_state", state, 600, "/", "", false, true)

	// Get authorization URL
	authURL, err := h.oidcService.GetAuthURL(c.Request.Context(), req.Provider, state)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Unsupported provider")
			return
		}
		InternalServerErrorResponse(c, "Failed to get auth URL")
		return
	}

	OKResponse(c, gin.H{
		"auth_url": authURL,
		"provider": req.Provider,
		"state":    state,
	}, "OIDC login initiated successfully")
}

// Logout handles OIDC logout
func (h *OIDCHandler) Logout(c *gin.Context) {
	var req struct {
		Provider              string `json:"provider" binding:"required"`
		IDToken               string `json:"id_token" binding:"required"`
		PostLogoutRedirectURI string `json:"post_logout_redirect_uri"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		InternalServerErrorResponse(c, "Invalid user ID")
		return
	}

	// Set default redirect URI if not provided
	if req.PostLogoutRedirectURI == "" {
		req.PostLogoutRedirectURI = "http://localhost:3000"
	}

	// Call OIDC logout
	if err := h.oidcService.EndSession(c.Request.Context(), userUUID, req.Provider, req.IDToken, req.PostLogoutRedirectURI); err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Invalid request")
			return
		}
		InternalServerErrorResponse(c, "Failed to logout from OIDC provider")
		return
	}

	// Get logout URL for front-channel logout
	logoutURL, err := h.oidcService.GetLogoutURL(c.Request.Context(), req.Provider, req.PostLogoutRedirectURI)
	if err != nil {
		// Don't fail the logout if we can't get the URL
		logoutURL = ""
	}

	OKResponse(c, gin.H{
		"message":                  "OIDC logout successful",
		"provider":                 req.Provider,
		"logout_url":               logoutURL,
		"post_logout_redirect_uri": req.PostLogoutRedirectURI,
	}, "OIDC logout successful")
}

// GetProviders returns the list of available OIDC providers
func (h *OIDCHandler) GetProviders(c *gin.Context) {
	providers := []gin.H{
		{
			"id":          "google",
			"name":        "Google",
			"description": "Sign in with Google",
			"icon":        "https://developers.google.com/identity/images/g-logo.png",
		},
		{
			"id":          "github",
			"name":        "GitHub",
			"description": "Sign in with GitHub",
			"icon":        "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
		},
		{
			"id":          "azure",
			"name":        "Microsoft Azure",
			"description": "Sign in with Microsoft Azure",
			"icon":        "https://azure.microsoft.com/svghandler/identity/",
		},
	}

	OKResponse(c, gin.H{
		"providers": providers,
		"total":     len(providers),
	}, "OIDC providers retrieved successfully")
}

// GetLogoutURL returns the OIDC logout URL
func (h *OIDCHandler) GetLogoutURL(c *gin.Context) {
	provider := c.Param("provider")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")

	if provider == "" {
		BadRequestResponse(c, "Provider is required")
		return
	}

	if postLogoutRedirectURI == "" {
		postLogoutRedirectURI = "http://localhost:3000"
	}

	logoutURL, err := h.oidcService.GetLogoutURL(c.Request.Context(), provider, postLogoutRedirectURI)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Unsupported provider")
			return
		}
		InternalServerErrorResponse(c, "Failed to get logout URL")
		return
	}

	OKResponse(c, gin.H{
		"logout_url":               logoutURL,
		"provider":                 provider,
		"post_logout_redirect_uri": postLogoutRedirectURI,
	}, "Logout URL generated successfully")
}

// generateState generates a random state string for CSRF protection
func (h *OIDCHandler) generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

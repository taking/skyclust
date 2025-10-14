package http

import (
	"net/http"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogoutHandler handles unified logout requests
type LogoutHandler struct {
	logoutService *usecase.LogoutService
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(logoutService *usecase.LogoutService) *LogoutHandler {
	return &LogoutHandler{
		logoutService: logoutService,
	}
}

// UnifiedLogout handles unified logout for both JWT and OIDC
func (h *LogoutHandler) UnifiedLogout(c *gin.Context) {
	var req usecase.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil {
		BadRequestResponse(c, "user_id is required")
		return
	}

	if req.Token == "" {
		BadRequestResponse(c, "token is required")
		return
	}

	if req.AuthType == "" {
		BadRequestResponse(c, "auth_type is required")
		return
	}

	// Validate auth type
	if req.AuthType != "jwt" && req.AuthType != "oidc" {
		BadRequestResponse(c, "auth_type must be 'jwt' or 'oidc'")
		return
	}

	// For OIDC, validate additional fields
	if req.AuthType == "oidc" {
		if req.Provider == "" {
			BadRequestResponse(c, "provider is required for OIDC logout")
			return
		}
		if req.IDToken == "" {
			BadRequestResponse(c, "id_token is required for OIDC logout")
			return
		}
	}

	// Process logout
	resp, err := h.logoutService.Logout(c.Request.Context(), req)
	if err != nil {
		if domain.IsDomainError(err) {
			domainErr := domain.GetDomainError(err)
			c.JSON(domainErr.StatusCode, gin.H{
				"error": domainErr,
			})
			return
		}
		InternalServerErrorResponse(c, "Failed to logout")
		return
	}

	// Return appropriate response based on auth type
	if req.AuthType == "oidc" && resp.LogoutURL != "" {
		// For OIDC, return logout URL for front-channel logout
		c.JSON(http.StatusOK, gin.H{
			"success":                  resp.Success,
			"message":                  resp.Message,
			"logout_url":               resp.LogoutURL,
			"post_logout_redirect_uri": resp.PostLogoutRedirectURI,
		})
	} else {
		// For JWT, return simple success response
		c.JSON(http.StatusOK, gin.H{
			"success": resp.Success,
			"message": resp.Message,
		})
	}
}

// BatchLogout handles logout for multiple tokens
func (h *LogoutHandler) BatchLogout(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Tokens []string  `json:"tokens" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	if len(req.Tokens) == 0 {
		BadRequestResponse(c, "tokens array cannot be empty")
		return
	}

	if err := h.logoutService.BatchLogout(c.Request.Context(), req.UserID, req.Tokens); err != nil {
		InternalServerErrorResponse(c, "Failed to logout from all devices")
		return
	}

	OKResponse(c, gin.H{
		"message":      "Batch logout successful",
		"tokens_count": len(req.Tokens),
	}, "Batch logout successful")
}

// GetLogoutStats returns logout statistics
func (h *LogoutHandler) GetLogoutStats(c *gin.Context) {
	stats, err := h.logoutService.GetLogoutStats(c.Request.Context())
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get logout statistics")
		return
	}

	OKResponse(c, stats, "Logout statistics retrieved successfully")
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (h *LogoutHandler) CleanupExpiredTokens(c *gin.Context) {
	if err := h.logoutService.CleanupExpiredTokens(c.Request.Context()); err != nil {
		InternalServerErrorResponse(c, "Failed to cleanup expired tokens")
		return
	}

	OKResponse(c, gin.H{
		"message": "Expired tokens cleanup completed",
	}, "Cleanup completed successfully")
}

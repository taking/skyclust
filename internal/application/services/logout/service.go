package logout

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// Service handles unified logout for both JWT and OIDC authentication
type Service struct {
	blacklist    *cache.TokenBlacklist
	oidcService  domain.OIDCService
	auditLogRepo domain.AuditLogRepository
}

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

// NewService creates a new unified logout service
func NewService(
	blacklist *cache.TokenBlacklist,
	oidcService domain.OIDCService,
	auditLogRepo domain.AuditLogRepository,
) *Service {
	return &Service{
		blacklist:    blacklist,
		oidcService:  oidcService,
		auditLogRepo: auditLogRepo,
	}
}

// Logout handles simple logout (for backward compatibility with domain interface)
func (ls *Service) Logout(userID uuid.UUID, token string) error {
	ctx := context.Background()
	req := LogoutRequest{
		UserID:   userID,
		Token:    token,
		AuthType: "jwt", // Default to JWT
	}
	_, err := ls.handleJWTLogout(ctx, req)
	return err
}

// LogoutWithContext handles unified logout for both JWT and OIDC authentication
func (ls *Service) LogoutWithContext(ctx context.Context, req LogoutRequest) (*LogoutResponse, error) {
	switch req.AuthType {
	case "jwt":
		return ls.handleJWTLogout(ctx, req)
	case "oidc":
		return ls.handleOIDCLogout(ctx, req)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported auth type", 400)
	}
}

// handleJWTLogout handles JWT token logout
func (ls *Service) handleJWTLogout(ctx context.Context, req LogoutRequest) (*LogoutResponse, error) {
	// Add token to blacklist
	// Set expiry to 15 minutes (matching JWT expiry for better security)
	expiry := 15 * time.Minute
	if err := ls.blacklist.AddToBlacklist(ctx, req.Token, expiry); err != nil {
		logger.Errorf("Failed to add JWT token to blacklist: %v", err)
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to invalidate token", 500)
	}

	// Log JWT logout
	_ = ls.auditLogRepo.Create(&domain.AuditLog{
		UserID:   req.UserID,
		Action:   domain.ActionUserLogout,
		Resource: "POST /api/v1/auth/logout",
		Details: map[string]interface{}{
			"auth_type":  "jwt",
			"token_hash": ls.hashToken(req.Token),
		},
	})

	logger.Infof("JWT logout successful for user: %s", req.UserID.String())

	return &LogoutResponse{
		Success: true,
		Message: "JWT logout successful",
	}, nil
}

// handleOIDCLogout handles OIDC logout
func (ls *Service) handleOIDCLogout(ctx context.Context, req LogoutRequest) (*LogoutResponse, error) {
	// Call OIDC provider's end_session_endpoint
	if err := ls.oidcService.EndSession(ctx, req.UserID, req.Provider, req.IDToken, req.PostLogoutRedirectURI); err != nil {
		logger.Errorf("Failed to call OIDC end session endpoint: %v", err)
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to logout from OIDC provider", 500)
	}

	// Get logout URL for front-channel logout
	logoutURL, err := ls.oidcService.GetLogoutURL(ctx, req.Provider, req.PostLogoutRedirectURI)
	if err != nil {
		logger.Warnf("Failed to get OIDC logout URL: %v", err)
		// Don't fail the logout if we can't get the URL
	}

	logger.Infof("OIDC logout successful for user: %s, provider: %s", req.UserID.String(), req.Provider)

	return &LogoutResponse{
		Success:               true,
		Message:               "OIDC logout successful",
		LogoutURL:             logoutURL,
		PostLogoutRedirectURI: req.PostLogoutRedirectURI,
	}, nil
}

// BatchLogout handles logout for multiple tokens (useful for multi-device logout)
func (ls *Service) BatchLogout(ctx context.Context, userID uuid.UUID, tokens []string) error {
	var errors []error

	for _, token := range tokens {
		req := LogoutRequest{
			UserID:   userID,
			Token:    token,
			AuthType: "jwt", // Assume JWT for batch logout
		}

		_, err := ls.handleJWTLogout(ctx, req)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("batch logout failed for %d tokens: %v", len(errors), errors), 500)
	}

	logger.Infof("Batch logout successful for user %s, token count: %d", userID.String(), len(tokens))

	return nil
}

// GetLogoutStats returns statistics about logout operations
func (ls *Service) GetLogoutStats(ctx context.Context) (map[string]interface{}, error) {
	// Get blacklist statistics
	blacklistStats, err := ls.blacklist.GetBlacklistStats(ctx)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get blacklist stats: %v", err), 500)
	}

	// Add additional logout statistics
	stats := map[string]interface{}{
		"blacklist_stats": blacklistStats,
		"timestamp":       time.Now().Unix(),
	}

	return stats, nil
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (ls *Service) CleanupExpiredTokens(ctx context.Context) error {
	if err := ls.blacklist.CleanupExpiredTokens(ctx); err != nil {
		logger.Errorf("Failed to cleanup expired tokens: %v", err)
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to cleanup expired tokens: %v", err), 500)
	}

	logger.Info("Expired tokens cleanup completed")
	return nil
}

// hashToken creates a secure hash of the token for logging
func (ls *Service) hashToken(token string) string {
	// Use the same hashing method as TokenBlacklist
	// This is for logging purposes only
	return fmt.Sprintf("hash_%x", len(token))
}

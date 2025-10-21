package auth

import (
	"net/http"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TokenExtractor provides utilities for extracting information from JWT tokens
type TokenExtractor struct{}

// NewTokenExtractor creates a new token extractor
func NewTokenExtractor() *TokenExtractor {
	return &TokenExtractor{}
}

// GetUserIDFromToken extracts user ID from the token in the request context
func (te *TokenExtractor) GetUserIDFromToken(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"User ID not found in token",
			http.StatusUnauthorized,
		)
	}

	userID, ok := userIDStr.(string)
	if !ok {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Invalid user ID format in token",
			http.StatusUnauthorized,
		)
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Invalid user ID format",
			http.StatusUnauthorized,
		)
	}

	return parsedUserID, nil
}

// GetUserRoleFromToken extracts user role from the token in the request context
func (te *TokenExtractor) GetUserRoleFromToken(c *gin.Context) (domain.Role, error) {
	userRoleStr, exists := c.Get("user_role")
	if !exists {
		return "", domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"User role not found in token",
			http.StatusUnauthorized,
		)
	}

	userRole, ok := userRoleStr.(string)
	if !ok {
		return "", domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Invalid user role format in token",
			http.StatusUnauthorized,
		)
	}

	return domain.Role(userRole), nil
}

// GetWorkspaceIDFromToken extracts workspace ID from the token in the request context
func (te *TokenExtractor) GetWorkspaceIDFromToken(c *gin.Context) (uuid.UUID, error) {
	workspaceIDStr, exists := c.Get("workspace_id")
	if !exists {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Workspace ID not found in token",
			http.StatusUnauthorized,
		)
	}

	workspaceID, ok := workspaceIDStr.(string)
	if !ok {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Invalid workspace ID format in token",
			http.StatusUnauthorized,
		)
	}

	parsedWorkspaceID, err := uuid.Parse(workspaceID)
	if err != nil {
		return uuid.Nil, domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Invalid workspace ID format",
			http.StatusUnauthorized,
		)
	}

	return parsedWorkspaceID, nil
}

// GetUserInfoFromToken extracts all user information from the token
func (te *TokenExtractor) GetUserInfoFromToken(c *gin.Context) (userID uuid.UUID, userRole domain.Role, workspaceID uuid.UUID, err error) {
	userID, err = te.GetUserIDFromToken(c)
	if err != nil {
		return uuid.Nil, "", uuid.Nil, err
	}

	userRole, err = te.GetUserRoleFromToken(c)
	if err != nil {
		return uuid.Nil, "", uuid.Nil, err
	}

	workspaceID, err = te.GetWorkspaceIDFromToken(c)
	if err != nil {
		return uuid.Nil, "", uuid.Nil, err
	}

	return userID, userRole, workspaceID, nil
}

// CheckAdminPermission checks if the user has admin role
func (te *TokenExtractor) CheckAdminPermission(c *gin.Context) (bool, error) {
	userRole, err := te.GetUserRoleFromToken(c)
	if err != nil {
		return false, err
	}

	return userRole == domain.AdminRoleType, nil
}

// CheckUserPermission checks if the user can access the resource (own resource or admin)
func (te *TokenExtractor) CheckUserPermission(c *gin.Context, resourceUserID uuid.UUID) (bool, error) {
	currentUserID, err := te.GetUserIDFromToken(c)
	if err != nil {
		return false, err
	}

	userRole, err := te.GetUserRoleFromToken(c)
	if err != nil {
		return false, err
	}

	// Can access if it's own resource or user is admin
	return currentUserID == resourceUserID || userRole == domain.AdminRoleType, nil
}

// GetBearerTokenFromHeader extracts Bearer token from Authorization header
func (te *TokenExtractor) GetBearerTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", domain.NewDomainError(
			domain.ErrCodeUnauthorized,
			"Authorization header is required",
			http.StatusUnauthorized,
		)
	}

	// Extract Bearer token
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}

	return "", domain.NewDomainError(
		domain.ErrCodeUnauthorized,
		"Invalid authorization header format. Expected 'Bearer <token>'",
		http.StatusUnauthorized,
	)
}

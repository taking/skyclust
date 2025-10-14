package http

import (
	"skyclust/internal/domain"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CredentialHandler handles credential-related HTTP requests
type CredentialHandler struct {
	credentialService domain.CredentialService
}

// NewCredentialHandler creates a new credential handler
func NewCredentialHandler(credentialService domain.CredentialService) *CredentialHandler {
	return &CredentialHandler{
		credentialService: credentialService,
	}
}

// CreateCredential handles credential creation
func (h *CredentialHandler) CreateCredential(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Handle both string and uuid.UUID types
	var userUUID uuid.UUID
	var err error

	if userIDStr, ok := userID.(string); ok {
		userUUID, err = uuid.Parse(userIDStr)
		if err != nil {
			InternalServerErrorResponse(c, "Invalid user ID format")
			return
		}
	} else if userUUID, ok = userID.(uuid.UUID); !ok {
		InternalServerErrorResponse(c, "Invalid user ID type")
		return
	}

	var req domain.CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userUUID, req)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Validation failed")
			return
		}
		InternalServerErrorResponse(c, "Failed to create credential")
		return
	}

	CreatedResponse(c, credential, "Credential created successfully")
}

// GetCredentials handles credential listing
func (h *CredentialHandler) GetCredentials(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Handle both string and uuid.UUID types
	var userUUID uuid.UUID
	var err error

	if userIDStr, ok := userID.(string); ok {
		userUUID, err = uuid.Parse(userIDStr)
		if err != nil {
			InternalServerErrorResponse(c, "Invalid user ID format")
			return
		}
	} else if userUUID, ok = userID.(uuid.UUID); !ok {
		InternalServerErrorResponse(c, "Invalid user ID type")
		return
	}

	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userUUID)
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get credentials")
		return
	}

	OKResponse(c, credentials, "Credentials retrieved successfully")
}

// GetCredential handles single credential retrieval
func (h *CredentialHandler) GetCredential(c *gin.Context) {
	// Get user ID from context
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

	// Get credential ID from URL parameter
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid credential ID")
		return
	}

	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userUUID, credentialID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			NotFoundResponse(c, "Credential not found")
			return
		}
		if domain.IsUnauthorizedError(err) {
			ForbiddenResponse(c, "Access denied")
			return
		}
		InternalServerErrorResponse(c, "Failed to get credential")
		return
	}

	OKResponse(c, credential, "Credential retrieved successfully")
}

// UpdateCredential handles credential updates
func (h *CredentialHandler) UpdateCredential(c *gin.Context) {
	// Get user ID from context
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

	// Get credential ID from URL parameter
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid credential ID")
		return
	}

	var req domain.UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	credential, err := h.credentialService.UpdateCredential(c.Request.Context(), userUUID, credentialID, req)
	if err != nil {
		if domain.IsValidationError(err) {
			BadRequestResponse(c, "Validation failed")
			return
		}
		if domain.IsNotFoundError(err) {
			NotFoundResponse(c, "Credential not found")
			return
		}
		if domain.IsUnauthorizedError(err) {
			ForbiddenResponse(c, "Access denied")
			return
		}
		InternalServerErrorResponse(c, "Failed to update credential")
		return
	}

	OKResponse(c, credential, "Credential updated successfully")
}

// DeleteCredential handles credential deletion
func (h *CredentialHandler) DeleteCredential(c *gin.Context) {
	// Get user ID from context
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

	// Get credential ID from URL parameter
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid credential ID")
		return
	}

	err = h.credentialService.DeleteCredential(c.Request.Context(), userUUID, credentialID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			NotFoundResponse(c, "Credential not found")
			return
		}
		if domain.IsUnauthorizedError(err) {
			ForbiddenResponse(c, "Access denied")
			return
		}
		InternalServerErrorResponse(c, "Failed to delete credential")
		return
	}

	OKResponse(c, nil, "Credential deleted successfully")
}

// GetCredentialsByProvider handles credential listing by provider
func (h *CredentialHandler) GetCredentialsByProvider(c *gin.Context) {
	// Get user ID from context
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

	// Get provider from URL parameter
	provider := c.Param("provider")
	if provider == "" {
		BadRequestResponse(c, "Provider is required")
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get credentials by provider (this would need to be implemented in the service)
	// For now, we'll get all credentials and filter by provider
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userUUID)
	if err != nil {
		InternalServerErrorResponse(c, "Failed to get credentials")
		return
	}

	// Filter by provider
	var filteredCredentials []*domain.Credential
	for _, cred := range credentials {
		if cred.Provider == provider {
			filteredCredentials = append(filteredCredentials, cred)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(filteredCredentials) {
		filteredCredentials = []*domain.Credential{}
	} else {
		if end > len(filteredCredentials) {
			end = len(filteredCredentials)
		}
		filteredCredentials = filteredCredentials[start:end]
	}

	OKResponse(c, gin.H{
		"credentials": filteredCredentials,
		"total":       len(filteredCredentials),
		"limit":       limit,
		"offset":      offset,
	}, "Credentials retrieved successfully")
}

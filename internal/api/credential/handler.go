package credential

import (
	"net/http"
	"skyclust/internal/api/common"
	"skyclust/internal/domain"
	"skyclust/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles credential management operations
type Handler struct {
	credentialService  domain.CredentialService
	tokenExtractor     *utils.TokenExtractor
	performanceTracker *common.PerformanceTracker
	requestLogger      *common.RequestLogger
	validationRules    *common.ValidationRules
	queryOptimizer     *common.QueryOptimizer
}

// NewHandler creates a new credential handler
func NewHandler(credentialService domain.CredentialService) *Handler {
	return &Handler{
		credentialService:  credentialService,
		tokenExtractor:     utils.NewTokenExtractor(),
		performanceTracker: common.NewPerformanceTracker("credential"),
		requestLogger:      common.NewRequestLogger(nil),
		validationRules:    common.NewValidationRules(),
		queryOptimizer:     nil, // Will be set by dependency injection
	}
}

// CreateCredential handles credential creation
func (h *Handler) CreateCredential(c *gin.Context) {
	var req domain.CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
	if err != nil {
		common.InternalServerError(c, "Failed to create credential")
		return
	}

	common.Success(c, http.StatusCreated, credential, "Credential created successfully")
}

// GetCredentials handles credential listing
func (h *Handler) GetCredentials(c *gin.Context) {
	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		common.InternalServerError(c, "Failed to get credentials")
		return
	}

	common.OK(c, gin.H{
		"credentials": credentials,
	}, "Credentials retrieved successfully")
}

// GetCredential handles single credential retrieval
func (h *Handler) GetCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get all credentials and find the specific one
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		common.InternalServerError(c, "Failed to get credentials")
		return
	}

	// Find the specific credential
	var credential *domain.Credential
	for _, cred := range credentials {
		if cred.ID == credentialID {
			credential = cred
			break
		}
	}

	if credential == nil {
		common.NotFound(c, "Credential not found")
		return
	}

	common.OK(c, credential, "Credential retrieved successfully")
}

// UpdateCredential handles credential updates
func (h *Handler) UpdateCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	var req domain.UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request body")
		return
	}

	credential, err := h.credentialService.UpdateCredential(c.Request.Context(), userID, credentialID, req)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "Credential not found")
			return
		}
		common.InternalServerError(c, "Failed to update credential")
		return
	}

	common.OK(c, credential, "Credential updated successfully")
}

// DeleteCredential handles credential deletion
func (h *Handler) DeleteCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	err = h.credentialService.DeleteCredential(c.Request.Context(), userID, credentialID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			common.NotFound(c, "Credential not found")
			return
		}
		common.InternalServerError(c, "Failed to delete credential")
		return
	}

	common.OK(c, gin.H{"message": "Credential deleted successfully"}, "Credential deleted successfully")
}

// GetCredentialsByProvider handles getting credentials by provider
func (h *Handler) GetCredentialsByProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		common.BadRequest(c, "Provider parameter is required")
		return
	}

	// Get user ID from token
	userID, err := h.tokenExtractor.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			common.DomainError(c, domainErr)
		} else {
			common.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get all credentials and filter by provider
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		common.InternalServerError(c, "Failed to get credentials")
		return
	}

	// Filter credentials by provider
	var filteredCredentials []*domain.Credential
	for _, cred := range credentials {
		if cred.Provider == provider {
			filteredCredentials = append(filteredCredentials, cred)
		}
	}

	common.OK(c, gin.H{
		"credentials": filteredCredentials,
		"provider":    provider,
	}, "Credentials retrieved successfully")
}

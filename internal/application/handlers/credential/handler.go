package credential

import (
	"encoding/json"
	"io"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles credential management operations
type Handler struct {
	*handlers.BaseHandler
	credentialService domain.CredentialService
}

// NewHandler creates a new credential handler
func NewHandler(credentialService domain.CredentialService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("credential"),
		credentialService: credentialService,
	}
}

// CreateCredential handles credential creation
func (h *Handler) CreateCredential(c *gin.Context) {
	var req domain.CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
	if err != nil {
		responses.InternalServerError(c, "Failed to create credential")
		return
	}

	responses.Created(c, credential, "Credential created successfully")
}

// GetCredentials handles credential listing
func (h *Handler) GetCredentials(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		responses.InternalServerError(c, "Failed to get credentials")
		return
	}

	responses.OK(c, gin.H{
		"credentials": credentials,
	}, "Credentials retrieved successfully")
}

// GetCredential handles single credential retrieval
func (h *Handler) GetCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get all credentials and find the specific one
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		responses.InternalServerError(c, "Failed to get credentials")
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
		responses.NotFound(c, "Credential not found")
		return
	}

	responses.OK(c, credential, "Credential retrieved successfully")
}

// UpdateCredential handles credential updates
func (h *Handler) UpdateCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	var req domain.UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body")
		return
	}

	credential, err := h.credentialService.UpdateCredential(c.Request.Context(), userID, credentialID, req)
	if err != nil {
		if domain.IsNotFoundError(err) {
			responses.NotFound(c, "Credential not found")
			return
		}
		responses.InternalServerError(c, "Failed to update credential")
		return
	}

	responses.OK(c, credential, "Credential updated successfully")
}

// DeleteCredential handles credential deletion
func (h *Handler) DeleteCredential(c *gin.Context) {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID format")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	err = h.credentialService.DeleteCredential(c.Request.Context(), userID, credentialID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			responses.NotFound(c, "Credential not found")
			return
		}
		responses.InternalServerError(c, "Failed to delete credential")
		return
	}

	responses.OK(c, gin.H{"message": "Credential deleted successfully"}, "Credential deleted successfully")
}

// CreateCredentialFromFile handles credential creation from uploaded file
func (h *Handler) CreateCredentialFromFile(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Parse form data
	name := c.PostForm("name")
	provider := c.PostForm("provider")

	if name == "" || provider == "" {
		responses.BadRequest(c, "name and provider are required")
		return
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		responses.BadRequest(c, "Failed to get uploaded file")
		return
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		responses.BadRequest(c, "Failed to read file content")
		return
	}

	// Parse JSON content
	var credentialData map[string]interface{}
	if err := json.Unmarshal(fileContent, &credentialData); err != nil {
		responses.BadRequest(c, "Invalid JSON file format")
		return
	}

	// Create credential request
	req := domain.CreateCredentialRequest{
		Name:     name,
		Provider: provider,
		Data:     credentialData,
	}

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
	if err != nil {
		responses.InternalServerError(c, "Failed to create credential")
		return
	}

	responses.Created(c, credential, "Credential created successfully from file")
}

// GetCredentialsByProvider handles getting credentials by provider
func (h *Handler) GetCredentialsByProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		responses.BadRequest(c, "Provider parameter is required")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Get all credentials and filter by provider
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		responses.InternalServerError(c, "Failed to get credentials")
		return
	}

	// Filter credentials by provider
	var filteredCredentials []*domain.Credential
	for _, cred := range credentials {
		if cred.Provider == provider {
			filteredCredentials = append(filteredCredentials, cred)
		}
	}

	responses.OK(c, gin.H{
		"credentials": filteredCredentials,
		"provider":    provider,
	}, "Credentials retrieved successfully")
}

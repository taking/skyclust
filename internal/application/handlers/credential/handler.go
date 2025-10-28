package credential

import (
	"encoding/json"
	"io"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	// Start performance tracking
	defer h.TrackRequest(c, "create_credential", 201)

	// Log operation start
	h.LogInfo(c, "Creating credential",
		zap.String("operation", "create_credential"))

	var req domain.CreateCredentialRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_credential")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credential_creation_attempted", userID.String(), "", map[string]interface{}{
		"provider": req.Provider,
		"name":     req.Name,
	})

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
	if err != nil {
		h.LogError(c, err, "Failed to create credential")
		h.HandleError(c, err, "create_credential")
		return
	}

	// Log successful creation
	h.LogBusinessEvent(c, "credential_created", userID.String(), credential.ID.String(), map[string]interface{}{
		"credential_id": credential.ID.String(),
		"provider":      credential.Provider,
		"name":          credential.Name,
	})

	h.LogInfo(c, "Credential created successfully",
		zap.String("credential_id", credential.ID.String()),
		zap.String("provider", credential.Provider))

	h.Created(c, credential, "Credential created successfully")
}

// GetCredentials handles credential listing
func (h *Handler) GetCredentials(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_credentials", 200)

	// Log operation start
	h.LogInfo(c, "Getting credentials",
		zap.String("operation", "get_credentials"))

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_credentials")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credentials_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_credentials",
	})

	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		h.LogError(c, err, "Failed to get credentials")
		h.HandleError(c, err, "get_credentials")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Credentials retrieved successfully",
		zap.Int("credentials_count", len(credentials)))

	h.OK(c, gin.H{
		"credentials": credentials,
	}, "Credentials retrieved successfully")
}

// GetCredential handles single credential retrieval
func (h *Handler) GetCredential(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_credential", 200)

	credentialIDStr := c.Param("id")

	// Log operation start
	h.LogInfo(c, "Getting specific credential",
		zap.String("operation", "get_credential"),
		zap.String("credential_id", credentialIDStr))

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid credential ID format",
			zap.String("credential_id", credentialIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID format", 400), "get_credential")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_credential")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credential_requested", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})

	// Get all credentials and find the specific one
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		h.LogError(c, err, "Failed to get credentials")
		h.HandleError(c, err, "get_credential")
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
		h.LogWarn(c, "Credential not found",
			zap.String("credential_id", credentialID.String()))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "get_credential")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Credential retrieved successfully",
		zap.String("credential_id", credentialID.String()))

	h.OK(c, credential, "Credential retrieved successfully")
}

// UpdateCredential handles credential updates
func (h *Handler) UpdateCredential(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "update_credential", 200)

	credentialIDStr := c.Param("id")

	// Log operation start
	h.LogInfo(c, "Updating credential",
		zap.String("operation", "update_credential"),
		zap.String("credential_id", credentialIDStr))

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid credential ID format",
			zap.String("credential_id", credentialIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID format", 400), "update_credential")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_credential")
		return
	}

	var req domain.UpdateCredentialRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credential_update_attempted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})

	credential, err := h.credentialService.UpdateCredential(c.Request.Context(), userID, credentialID, req)
	if err != nil {
		h.LogError(c, err, "Failed to update credential")
		h.HandleError(c, err, "update_credential")
		return
	}

	// Log successful update
	h.LogBusinessEvent(c, "credential_updated", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})

	h.LogInfo(c, "Credential updated successfully",
		zap.String("credential_id", credentialID.String()))

	h.OK(c, credential, "Credential updated successfully")
}

// DeleteCredential handles credential deletion
func (h *Handler) DeleteCredential(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "delete_credential", 200)

	credentialIDStr := c.Param("id")

	// Log operation start
	h.LogInfo(c, "Deleting credential",
		zap.String("operation", "delete_credential"),
		zap.String("credential_id", credentialIDStr))

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid credential ID format",
			zap.String("credential_id", credentialIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID format", 400), "delete_credential")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_credential")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credential_deletion_attempted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})

	err = h.credentialService.DeleteCredential(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.LogError(c, err, "Failed to delete credential")
		h.HandleError(c, err, "delete_credential")
		return
	}

	// Log successful deletion
	h.LogBusinessEvent(c, "credential_deleted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})

	h.LogInfo(c, "Credential deleted successfully",
		zap.String("credential_id", credentialID.String()))

	h.OK(c, gin.H{"message": "Credential deleted successfully"}, "Credential deleted successfully")
}

// CreateCredentialFromFile handles credential creation from uploaded file
func (h *Handler) CreateCredentialFromFile(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "create_credential_from_file", 201)

	// Log operation start
	h.LogInfo(c, "Creating credential from file",
		zap.String("operation", "create_credential_from_file"))

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_credential_from_file")
		return
	}

	// Parse form data
	name := c.PostForm("name")
	provider := c.PostForm("provider")

	if name == "" || provider == "" {
		h.LogWarn(c, "Missing required form fields",
			zap.String("name", name),
			zap.String("provider", provider))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "name and provider are required", 400), "create_credential_from_file")
		return
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		h.LogWarn(c, "Failed to get uploaded file", zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Failed to get uploaded file", 400), "create_credential_from_file")
		return
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.LogError(c, err, "Failed to read file content")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Failed to read file content", 400), "create_credential_from_file")
		return
	}

	// Parse JSON content
	var credentialData map[string]interface{}
	if err := json.Unmarshal(fileContent, &credentialData); err != nil {
		h.LogWarn(c, "Invalid JSON file format", zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid JSON file format", 400), "create_credential_from_file")
		return
	}

	// Create credential request
	req := domain.CreateCredentialRequest{
		Name:     name,
		Provider: provider,
		Data:     credentialData,
	}

	// Log business event
	h.LogBusinessEvent(c, "credential_from_file_creation_attempted", userID.String(), "", map[string]interface{}{
		"provider": provider,
		"name":     name,
	})

	credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
	if err != nil {
		h.LogError(c, err, "Failed to create credential from file")
		h.HandleError(c, err, "create_credential_from_file")
		return
	}

	// Log successful creation
	h.LogBusinessEvent(c, "credential_from_file_created", userID.String(), credential.ID.String(), map[string]interface{}{
		"credential_id": credential.ID.String(),
		"provider":      credential.Provider,
		"name":          credential.Name,
	})

	h.LogInfo(c, "Credential created successfully from file",
		zap.String("credential_id", credential.ID.String()),
		zap.String("provider", credential.Provider))

	h.Created(c, credential, "Credential created successfully from file")
}

// GetCredentialsByProvider handles getting credentials by provider
func (h *Handler) GetCredentialsByProvider(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_credentials_by_provider", 200)

	provider := c.Param("provider")

	// Log operation start
	h.LogInfo(c, "Getting credentials by provider",
		zap.String("operation", "get_credentials_by_provider"),
		zap.String("provider", provider))

	if provider == "" {
		h.LogWarn(c, "Provider parameter is required")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider parameter is required", 400), "get_credentials_by_provider")
		return
	}

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_credentials_by_provider")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "credentials_by_provider_requested", userID.String(), "", map[string]interface{}{
		"provider": provider,
	})

	// Get all credentials and filter by provider
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		h.LogError(c, err, "Failed to get credentials")
		h.HandleError(c, err, "get_credentials_by_provider")
		return
	}

	// Filter credentials by provider
	var filteredCredentials []*domain.Credential
	for _, cred := range credentials {
		if cred.Provider == provider {
			filteredCredentials = append(filteredCredentials, cred)
		}
	}

	// Log successful operation
	h.LogInfo(c, "Credentials by provider retrieved successfully",
		zap.String("provider", provider),
		zap.Int("credentials_count", len(filteredCredentials)))

	h.OK(c, gin.H{
		"credentials": filteredCredentials,
		"provider":    provider,
	}, "Credentials retrieved successfully")
}

package credential

import (
	"encoding/json"
	"io"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles credential management operations using improved patterns
type Handler struct {
	*handlers.BaseHandler
	credentialService domain.CredentialService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new credential handler
func NewHandler(credentialService domain.CredentialService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("credential"),
		credentialService: credentialService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// CreateCredential handles credential creation using decorator pattern
func (h *Handler) CreateCredential(c *gin.Context) {
	var req domain.CreateCredentialRequest

	handler := h.Compose(
		h.createCredentialHandler(req),
		h.StandardCRUDDecorators("create_credential")...,
	)

	handler(c)
}

// createCredentialHandler is the core business logic for credential creation
func (h *Handler) createCredentialHandler(req domain.CreateCredentialRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		req = h.extractValidatedRequest(c)
		userID := h.extractUserID(c)

		h.logCredentialCreationAttempt(c, userID, req)

		credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
		if err != nil {
			h.HandleError(c, err, "create_credential")
			return
		}

		h.logCredentialCreationSuccess(c, userID, credential)
		h.Created(c, credential, readability.SuccessMsgUserCreated)
	}
}

// GetCredentials handles credential listing using decorator pattern
func (h *Handler) GetCredentials(c *gin.Context) {
	handler := h.Compose(
		h.getCredentialsHandler(),
		h.StandardCRUDDecorators("get_credentials")...,
	)

	handler(c)
}

// getCredentialsHandler is the core business logic for getting credentials
func (h *Handler) getCredentialsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		h.logCredentialsRequest(c, userID)

		credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
		if err != nil {
			h.HandleError(c, err, "get_credentials")
			return
		}

		h.OK(c, gin.H{"credentials": credentials}, "Credentials retrieved successfully")
	}
}

// GetCredential handles single credential retrieval using decorator pattern
func (h *Handler) GetCredential(c *gin.Context) {
	handler := h.Compose(
		h.getCredentialHandler(),
		h.StandardCRUDDecorators("get_credential")...,
	)

	handler(c)
}

// getCredentialHandler is the core business logic for getting a single credential
func (h *Handler) getCredentialHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credentialID := h.parseCredentialID(c)
		userID := h.extractUserID(c)

		h.logCredentialRequest(c, userID, credentialID)

		credential, err := h.findCredentialByID(c, userID, credentialID)
		if err != nil {
			h.HandleError(c, err, "get_credential")
			return
		}

		h.OK(c, credential, "Credential retrieved successfully")
	}
}

// UpdateCredential handles credential updates using decorator pattern
func (h *Handler) UpdateCredential(c *gin.Context) {
	var req domain.UpdateCredentialRequest

	handler := h.Compose(
		h.updateCredentialHandler(req),
		h.StandardCRUDDecorators("update_credential")...,
	)

	handler(c)
}

// updateCredentialHandler is the core business logic for updating credentials
func (h *Handler) updateCredentialHandler(req domain.UpdateCredentialRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		credentialID := h.parseCredentialID(c)
		req = h.extractValidatedUpdateRequest(c)
		userID := h.extractUserID(c)

		h.logCredentialUpdateAttempt(c, userID, credentialID)

		credential, err := h.credentialService.UpdateCredential(c.Request.Context(), userID, credentialID, req)
		if err != nil {
			h.HandleError(c, err, "update_credential")
			return
		}

		h.logCredentialUpdateSuccess(c, userID, credentialID)
		h.OK(c, credential, readability.SuccessMsgUserUpdated)
	}
}

// DeleteCredential handles credential deletion using decorator pattern
func (h *Handler) DeleteCredential(c *gin.Context) {
	handler := h.Compose(
		h.deleteCredentialHandler(),
		h.StandardCRUDDecorators("delete_credential")...,
	)

	handler(c)
}

// deleteCredentialHandler is the core business logic for deleting credentials
func (h *Handler) deleteCredentialHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credentialID := h.parseCredentialID(c)
		userID := h.extractUserID(c)

		h.logCredentialDeletionAttempt(c, userID, credentialID)

		err := h.credentialService.DeleteCredential(c.Request.Context(), userID, credentialID)
		if err != nil {
			h.HandleError(c, err, "delete_credential")
			return
		}

		h.logCredentialDeletionSuccess(c, userID, credentialID)
		h.OK(c, gin.H{"message": "Credential deleted successfully"}, readability.SuccessMsgUserDeleted)
	}
}

// CreateCredentialFromFile handles credential creation from uploaded file using decorator pattern
func (h *Handler) CreateCredentialFromFile(c *gin.Context) {
	handler := h.Compose(
		h.createCredentialFromFileHandler(),
		h.StandardCRUDDecorators("create_credential_from_file")...,
	)

	handler(c)
}

// createCredentialFromFileHandler is the core business logic for creating credentials from file
func (h *Handler) createCredentialFromFileHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

		formData, err := h.parseFormData(c)
		if err != nil {
			h.HandleError(c, err, "create_credential_from_file")
			return
		}

		fileContent, err := h.readUploadedFile(c)
		if err != nil {
			h.HandleError(c, err, "create_credential_from_file")
			return
		}

		credentialData, err := h.parseJSONContent(fileContent)
		if err != nil {
			h.HandleError(c, err, "create_credential_from_file")
			return
		}

		req := h.buildCredentialRequest(formData, credentialData)
		h.logCredentialFromFileCreationAttempt(c, userID, formData)

		credential, err := h.credentialService.CreateCredential(c.Request.Context(), userID, req)
		if err != nil {
			h.HandleError(c, err, "create_credential_from_file")
			return
		}

		h.logCredentialFromFileCreationSuccess(c, userID, credential)
		h.Created(c, credential, "Credential created successfully from file")
	}
}

// GetCredentialsByProvider handles getting credentials by provider using decorator pattern
func (h *Handler) GetCredentialsByProvider(c *gin.Context) {
	handler := h.Compose(
		h.getCredentialsByProviderHandler(),
		h.StandardCRUDDecorators("get_credentials_by_provider")...,
	)

	handler(c)
}

// getCredentialsByProviderHandler is the core business logic for getting credentials by provider
func (h *Handler) getCredentialsByProviderHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		provider := h.extractProviderParam(c)
		userID := h.extractUserID(c)

		h.logCredentialsByProviderRequest(c, userID, provider)

		credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
		if err != nil {
			h.HandleError(c, err, "get_credentials_by_provider")
			return
		}

		filteredCredentials := h.filterCredentialsByProvider(credentials, provider)
		h.OK(c, gin.H{
			"credentials": filteredCredentials,
			"provider":    provider,
		}, "Credentials retrieved successfully")
	}
}

// Helper methods for better readability

func (h *Handler) extractValidatedRequest(c *gin.Context) domain.CreateCredentialRequest {
	if validatedReq, exists := c.Get("validated_request"); exists {
		return validatedReq.(domain.CreateCredentialRequest)
	}
	return domain.CreateCredentialRequest{}
}

func (h *Handler) extractValidatedUpdateRequest(c *gin.Context) domain.UpdateCredentialRequest {
	if validatedReq, exists := c.Get("validated_request"); exists {
		return validatedReq.(domain.UpdateCredentialRequest)
	}
	return domain.UpdateCredentialRequest{}
}

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func (h *Handler) parseCredentialID(c *gin.Context) uuid.UUID {
	credentialIDStr := c.Param("id")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID format", 400), "parse_credential_id")
		return uuid.Nil
	}
	return credentialID
}

func (h *Handler) extractProviderParam(c *gin.Context) string {
	provider := c.Param("provider")
	if provider == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Provider parameter is required", 400), "extract_provider")
		return ""
	}
	return provider
}

func (h *Handler) findCredentialByID(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) (*domain.Credential, error) {
	credentials, err := h.credentialService.GetCredentials(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}

	for _, cred := range credentials {
		if cred.ID == credentialID {
			return cred, nil
		}
	}

	return nil, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404)
}

func (h *Handler) filterCredentialsByProvider(credentials []*domain.Credential, provider string) []*domain.Credential {
	var filteredCredentials []*domain.Credential
	for _, cred := range credentials {
		if cred.Provider == provider {
			filteredCredentials = append(filteredCredentials, cred)
		}
	}
	return filteredCredentials
}

func (h *Handler) parseFormData(c *gin.Context) (map[string]string, error) {
	name := c.PostForm("name")
	provider := c.PostForm("provider")

	if name == "" || provider == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "name and provider are required", 400)
	}

	return map[string]string{
		"name":     name,
		"provider": provider,
	}, nil
}

func (h *Handler) readUploadedFile(c *gin.Context) ([]byte, error) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "Failed to get uploaded file", 400)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "Failed to read file content", 400)
	}

	return fileContent, nil
}

func (h *Handler) parseJSONContent(fileContent []byte) (map[string]interface{}, error) {
	var credentialData map[string]interface{}
	if err := json.Unmarshal(fileContent, &credentialData); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid JSON file format", 400)
	}
	return credentialData, nil
}

func (h *Handler) buildCredentialRequest(formData map[string]string, credentialData map[string]interface{}) domain.CreateCredentialRequest {
	return domain.CreateCredentialRequest{
		Name:     formData["name"],
		Provider: formData["provider"],
		Data:     credentialData,
	}
}

// Logging helper methods

func (h *Handler) logCredentialCreationAttempt(c *gin.Context, userID uuid.UUID, req domain.CreateCredentialRequest) {
	h.LogBusinessEvent(c, "credential_creation_attempted", userID.String(), "", map[string]interface{}{
		"provider": req.Provider,
		"name":     req.Name,
	})
}

func (h *Handler) logCredentialCreationSuccess(c *gin.Context, userID uuid.UUID, credential *domain.Credential) {
	h.LogBusinessEvent(c, "credential_created", userID.String(), credential.ID.String(), map[string]interface{}{
		"credential_id": credential.ID.String(),
		"provider":      credential.Provider,
		"name":          credential.Name,
	})
}

func (h *Handler) logCredentialsRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "credentials_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_credentials",
	})
}

func (h *Handler) logCredentialRequest(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "credential_requested", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})
}

func (h *Handler) logCredentialUpdateAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "credential_update_attempted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})
}

func (h *Handler) logCredentialUpdateSuccess(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "credential_updated", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})
}

func (h *Handler) logCredentialDeletionAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "credential_deletion_attempted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})
}

func (h *Handler) logCredentialDeletionSuccess(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "credential_deleted", userID.String(), credentialID.String(), map[string]interface{}{
		"credential_id": credentialID.String(),
	})
}

func (h *Handler) logCredentialFromFileCreationAttempt(c *gin.Context, userID uuid.UUID, formData map[string]string) {
	h.LogBusinessEvent(c, "credential_from_file_creation_attempted", userID.String(), "", map[string]interface{}{
		"provider": formData["provider"],
		"name":     formData["name"],
	})
}

func (h *Handler) logCredentialFromFileCreationSuccess(c *gin.Context, userID uuid.UUID, credential *domain.Credential) {
	h.LogBusinessEvent(c, "credential_from_file_created", userID.String(), credential.ID.String(), map[string]interface{}{
		"credential_id": credential.ID.String(),
		"provider":      credential.Provider,
		"name":          credential.Name,
	})
}

func (h *Handler) logCredentialsByProviderRequest(c *gin.Context, userID uuid.UUID, provider string) {
	h.LogBusinessEvent(c, "credentials_by_provider_requested", userID.String(), "", map[string]interface{}{
		"provider": provider,
	})
}

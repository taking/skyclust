package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"skyclust/internal/domain"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
)

// AzureCredentials: Azure 자격증명 정보
type AzureCredentials struct {
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string
	ResourceGroup  string
}

// extractAzureCredentials: 복호화된 자격증명 데이터에서 Azure 자격증명을 추출합니다
func (s *Service) extractAzureCredentials(ctx context.Context, credential *domain.Credential) (*AzureCredentials, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	subscriptionID, ok := credData["subscription_id"].(string)
	if !ok || subscriptionID == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "subscription_id not found in credential", 400)
	}

	clientID, ok := credData["client_id"].(string)
	if !ok || clientID == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "client_id not found in credential", 400)
	}

	clientSecret, ok := credData["client_secret"].(string)
	if !ok || clientSecret == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "client_secret not found in credential", 400)
	}

	tenantID, ok := credData["tenant_id"].(string)
	if !ok || tenantID == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "tenant_id not found in credential", 400)
	}

	// Resource group은 선택적 (요청에서 받을 수 있음)
	resourceGroup := ""
	if rg, ok := credData["resource_group"].(string); ok {
		resourceGroup = rg
	}

	return &AzureCredentials{
		SubscriptionID: subscriptionID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
		ResourceGroup:  resourceGroup,
	}, nil
}

// createAzureContainerServiceClient: Azure Container Service (AKS) 클라이언트를 생성합니다
func (s *Service) createAzureContainerServiceClient(ctx context.Context, creds *AzureCredentials) (*armcontainerservice.ClientFactory, error) {
	cred, err := azidentity.NewClientSecretCredential(
		creds.TenantID,
		creds.ClientID,
		creds.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create Azure credential: %v", err), 502)
	}

	clientFactory, err := armcontainerservice.NewClientFactory(creds.SubscriptionID, cred, nil)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create Azure Container Service client: %v", err), 502)
	}

	return clientFactory, nil
}

// handleAzureError: Azure SDK 에러를 적절한 도메인 에러로 변환합니다
func (s *Service) handleAzureError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check if it's an Azure ResponseError
	if strings.Contains(err.Error(), "ResponseError") {
		// Try to extract error details
		errorMsg := err.Error()

		// Check for specific Azure error codes
		switch {
		case strings.Contains(errorMsg, "InvalidAuthenticationToken") || strings.Contains(errorMsg, "AuthenticationFailed"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid Azure credentials", 400)
		case strings.Contains(errorMsg, "AuthorizationFailed") || strings.Contains(errorMsg, "Forbidden"):
			return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to Azure resources. Please check your Azure RBAC permissions.", 403)
		case strings.Contains(errorMsg, "ResourceNotFound") || strings.Contains(errorMsg, "NotFound"):
			return domain.NewDomainError(domain.ErrCodeNotFound, "Azure resource not found", 404)
		case strings.Contains(errorMsg, "InvalidParameter") || strings.Contains(errorMsg, "BadRequest"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Invalid Azure parameter: %s", errorMsg), 400)
		case strings.Contains(errorMsg, "TooManyRequests") || strings.Contains(errorMsg, "Throttled"):
			return domain.NewDomainError(domain.ErrCodeProviderQuota, "Azure API rate limit exceeded", 429)
		case strings.Contains(errorMsg, "Conflict"):
			return domain.NewDomainError(domain.ErrCodeConflict, "Azure resource conflict", 409)
		default:
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Azure API error: %s", errorMsg), 400)
		}
	}

	// For other errors, return as internal server error
	return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to %s: %v", operation, err), 500)
}


package network

import (
	"context"
	"fmt"

	"skyclust/internal/domain"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
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

// createAzureNetworkClient: Azure Network Management 클라이언트를 생성합니다
func (s *Service) createAzureNetworkClient(ctx context.Context, creds *AzureCredentials) (*armnetwork.ClientFactory, error) {
	cred, err := azidentity.NewClientSecretCredential(
		creds.TenantID,
		creds.ClientID,
		creds.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create Azure credential: %v", err), 502)
	}

	clientFactory, err := armnetwork.NewClientFactory(creds.SubscriptionID, cred, nil)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create Azure Network client: %v", err), 502)
	}

	return clientFactory, nil
}

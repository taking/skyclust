package cost_analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"skyclust/internal/domain"

	billingv1 "cloud.google.com/go/billing/apiv1"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"google.golang.org/api/option"
)

// getAWSCostExplorerClient: 자격증명으로부터 AWS Cost Explorer 클라이언트를 생성합니다
func (s *Service) getAWSCostExplorerClient(ctx context.Context, credential *domain.Credential, defaultRegion string) (*costexplorer.Client, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	region := defaultRegion
	if region == "" {
		region = AWSDefaultRegion
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	ceClient := costexplorer.NewFromConfig(cfg)
	return ceClient, region, nil
}

// setupGCPBillingClient: GCP Billing 클라이언트를 생성합니다
func (s *Service) setupGCPBillingClient(ctx context.Context, credential *domain.Credential) (*billingv1.CloudBillingClient, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	projectID, ok := credData["project_id"].(string)
	if !ok || projectID == "" {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "project_id not found in credential", 400)
	}

	// Create service account key from credential data
	serviceAccountKey := map[string]interface{}{
		"type":                        credData["type"],
		"project_id":                  credData["project_id"],
		"private_key_id":              credData["private_key_id"],
		"private_key":                 credData["private_key"],
		"client_email":                credData["client_email"],
		"client_id":                   credData["client_id"],
		"auth_uri":                    credData["auth_uri"],
		"token_uri":                   credData["token_uri"],
		"auth_provider_x509_cert_url": credData["auth_provider_x509_cert_url"],
		"client_x509_cert_url":        credData["client_x509_cert_url"],
	}

	keyBytes, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal service account key: %v", err), 500)
	}

	billingClient, err := billingv1.NewCloudBillingClient(ctx, option.WithCredentialsJSON(keyBytes))
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create billing client: %v", err), 502)
	}

	return billingClient, projectID, nil
}

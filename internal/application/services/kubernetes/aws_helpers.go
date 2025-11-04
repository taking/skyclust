package kubernetes

import (
	"context"
	"fmt"

	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	// AWSDefaultRegionForEKS is the default AWS region for EKS operations
	AWSDefaultRegionForEKS = "us-east-1"
)

// AWSCredentials contains extracted AWS credentials
type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

// extractAWSCredentials extracts AWS credentials from decrypted credential data
func (s *Service) extractAWSCredentials(ctx context.Context, credential *domain.Credential, defaultRegion string) (*AWSCredentials, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	region := defaultRegion
	if r, ok := credData["region"].(string); ok && r != "" {
		region = r
	}
	if region == "" {
		region = AWSDefaultRegionForEKS
	}

	return &AWSCredentials{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}, nil
}

// createAWSConfig creates AWS config from credentials
func (s *Service) createAWSConfig(ctx context.Context, creds *AWSCredentials) (aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(creds.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			creds.AccessKey,
			creds.SecretKey,
			"",
		)),
	)
	if err != nil {
		return aws.Config{}, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	return cfg, nil
}

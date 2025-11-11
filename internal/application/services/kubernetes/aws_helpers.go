package kubernetes

import (
	"context"
	"fmt"

	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

// extractAWSCredentials: 복호화된 자격 증명 데이터에서 AWS 자격 증명을 추출합니다
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

// createAWSConfig: 자격 증명으로부터 AWS 설정을 생성합니다
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

// getAWSAccountID: AWS STS를 통해 Account ID를 조회합니다
func (s *Service) getAWSAccountID(ctx context.Context, cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)

	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", domain.NewDomainError(
			domain.ErrCodeProviderError,
			fmt.Sprintf("failed to get AWS account ID: %v", err),
			502,
		)
	}

	if result.Account == nil {
		return "", domain.NewDomainError(
			domain.ErrCodeProviderError,
			"account ID not found in STS response",
			502,
		)
	}

	return *result.Account, nil
}

// generateDefaultRoleARN: 기본 Role ARN을 생성합니다
func (s *Service) generateDefaultRoleARN(ctx context.Context, cfg aws.Config, roleName string) (string, error) {
	accountID, err := s.getAWSAccountID(ctx, cfg)
	if err != nil {
		return "", err
	}

	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)

	if s.logger != nil {
		s.logger.Info(ctx, "Auto-generated Role ARN",
			domain.NewLogField("role_arn", roleARN),
			domain.NewLogField("account_id", accountID),
			domain.NewLogField("role_name", roleName))
	}

	return roleARN, nil
}

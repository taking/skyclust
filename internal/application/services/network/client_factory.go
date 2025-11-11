package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"skyclust/internal/domain"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// createEC2Client: AWS EC2 클라이언트를 생성합니다
func (s *Service) createEC2Client(ctx context.Context, credential *domain.Credential, region string) (*ec2.Client, error) {
	// Validate region - region이 비어있거나 VPC ID 형식(vpc-로 시작)이면 에러
	if region == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgRegionRequired, 400)
	}

	// VPC ID가 region으로 잘못 전달되는 경우 방지
	if strings.HasPrefix(region, "vpc-") {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf(ErrMsgInvalidRegionFormat, region), 400)
	}

	// Decrypt credential
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Extract AWS credentials (same as kubernetes service)
	accessKey, ok := decryptedData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgAccessKeyNotFound, 400)
	}

	secretKey, ok := decryptedData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgSecretKeyNotFound, 400)
	}

	// Debug: Log the extracted credentials and region
	s.logger.Info(ctx, "Creating AWS EC2 client",
		domain.NewLogField("access_key", accessKey),
		domain.NewLogField("secret_key", secretKey[:10]+"..."),
		domain.NewLogField("region", region))

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "load AWS config")
	}

	return ec2.NewFromConfig(cfg), nil
}

// createGCPComputeClient: GCP Compute 클라이언트를 생성합니다
func (s *Service) createGCPComputeClient(ctx context.Context, credential *domain.Credential) (*compute.Service, error) {
	// Decrypt credential
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Create GCP service account key from decrypted data
	serviceAccountKey := map[string]interface{}{
		"type":                        decryptedData["type"],
		"project_id":                  decryptedData["project_id"],
		"private_key_id":              decryptedData["private_key_id"],
		"private_key":                 decryptedData["private_key"],
		"client_email":                decryptedData["client_email"],
		"client_id":                   decryptedData["client_id"],
		"auth_uri":                    decryptedData["auth_uri"],
		"token_uri":                   decryptedData["token_uri"],
		"auth_provider_x509_cert_url": decryptedData["auth_provider_x509_cert_url"],
		"client_x509_cert_url":        decryptedData["client_x509_cert_url"],
		"universe_domain":             decryptedData["universe_domain"],
	}

	// Convert to JSON
	keyBytes, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalServiceAccountKey, err), 500)
	}

	// Create credentials from service account key
	creds, err := google.CredentialsFromJSON(ctx, keyBytes, compute.CloudPlatformScope)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create credentials")
	}

	// Create compute service
	computeService, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create compute service")
	}

	return computeService, nil
}


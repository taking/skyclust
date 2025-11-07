package common

import (
	"fmt"
	"skyclust/internal/domain"
)

// CredentialValidator: 자격증명 검증을 위한 공통 유틸리티
type CredentialValidator struct{}

// NewCredentialValidator: 새로운 자격증명 검증기를 생성합니다
func NewCredentialValidator() *CredentialValidator {
	return &CredentialValidator{}
}

// ValidateCredentialData: 프로바이더에 따라 자격증명 데이터를 검증합니다
func (v *CredentialValidator) ValidateCredentialData(provider string, data map[string]interface{}) error {
	switch provider {
	case domain.ProviderAWS:
		return v.validateAWSCredentials(data)
	case domain.ProviderGCP:
		return v.validateGCPCredentials(data)
	case domain.ProviderAzure:
		return v.validateAzureCredentials(data)
	case "openstack":
		return v.validateOpenStackCredentials(data)
	default:
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "unsupported provider", 400)
	}
}

// validateAWSCredentials: AWS 자격증명 데이터를 검증합니다
func (v *CredentialValidator) validateAWSCredentials(data map[string]interface{}) error {
	if _, ok := data["access_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key is required for AWS", 400)
	}
	if _, ok := data["secret_key"]; !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key is required for AWS", 400)
	}
	return nil
}

// validateGCPCredentials: GCP 자격증명 데이터를 검증합니다
func (v *CredentialValidator) validateGCPCredentials(data map[string]interface{}) error {
	// 필수 필드 검증
	requiredFields := []string{"type", "project_id", "private_key", "client_email", "client_id"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("%s is required for GCP service account", field), 400)
		}
	}

	// service_account 타입 확인
	if data["type"] != "service_account" {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("invalid service account type: %s", data["type"]), 400)
	}

	return nil
}

// validateAzureCredentials: Azure 자격증명 데이터를 검증합니다
func (v *CredentialValidator) validateAzureCredentials(data map[string]interface{}) error {
	requiredFields := []string{"subscription_id", "client_id", "client_secret", "tenant_id"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("%s is required for Azure", field), 400)
		}
	}
	return nil
}

// validateOpenStackCredentials: OpenStack 자격증명 데이터를 검증합니다
func (v *CredentialValidator) validateOpenStackCredentials(data map[string]interface{}) error {
	requiredFields := []string{"auth_url", "username", "password"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("%s is required for OpenStack", field), 400)
		}
	}
	return nil
}

// ValidateProvider: 지원되는 프로바이더인지 검증합니다
func (v *CredentialValidator) ValidateProvider(provider string) bool {
	switch provider {
	case domain.ProviderAWS, domain.ProviderGCP, domain.ProviderAzure, "openstack", domain.ProviderNCP:
		return true
	default:
		return false
	}
}

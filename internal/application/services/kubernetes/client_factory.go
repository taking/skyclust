package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"skyclust/internal/domain"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// setupGCPContainerService: GCP Container 서비스 클라이언트와 프로젝트 ID를 설정합니다
// 이 함수는 getGCPContainerServiceAndProjectID의 래퍼로, 일관된 네이밍을 제공합니다
func (s *Service) setupGCPContainerService(ctx context.Context, credential *domain.Credential) (*container.Service, string, error) {
	return s.getGCPContainerServiceAndProjectID(ctx, credential)
}

// getGCPContainerServiceAndProjectID: 자격 증명으로부터 GCP Container 서비스 클라이언트와 프로젝트 ID를 조회합니다
func (s *Service) getGCPContainerServiceAndProjectID(ctx context.Context, credential *domain.Credential) (*container.Service, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredentialData, err), 500)
	}

	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf(ErrMsgFailedToCreateGCPContainerService, err), 502)
	}

	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgProjectIDNotFound, 400)
	}

	return containerService, projectID, nil
}


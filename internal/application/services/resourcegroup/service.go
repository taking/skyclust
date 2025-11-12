package resourcegroup

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"skyclust/internal/domain"
	providererrors "skyclust/internal/shared/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// Service: Azure Resource Group 작업을 처리하는 서비스
type Service struct {
	credentialService      domain.CredentialService
	cacheService           domain.CacheService
	logger                 domain.LoggerService
	providerErrorConverter *providererrors.ProviderErrorConverter
}

// NewService: 새로운 Resource Group 서비스를 생성합니다
func NewService(
	credentialService domain.CredentialService,
	cacheService domain.CacheService,
	loggerService domain.LoggerService,
) *Service {
	// Use provided logger or fallback to default logger
	logger := loggerService
	if logger == nil {
		// Create a wrapper around default logger if domain.LoggerService is needed
		// For now, we'll use nil and handle it in methods
		logger = nil
	}
	return &Service{
		credentialService:      credentialService,
		cacheService:           cacheService,
		logger:                 logger,
		providerErrorConverter: providererrors.NewProviderErrorConverter(),
	}
}

// ListResourceGroups: Azure Resource Group 목록을 조회합니다
func (s *Service) ListResourceGroups(ctx context.Context, credential *domain.Credential, req ListResourceGroupsRequest) (*ListResourceGroupsResponse, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("azure:resource-groups:%s:%s", credentialID, req.Location)

	var resourceGroups []ResourceGroupInfo

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			var cachedResponse *ListResourceGroupsResponse
			if resp, ok := cachedValue.(*ListResourceGroupsResponse); ok {
				cachedResponse = resp
			} else if resp, ok := cachedValue.(ListResourceGroupsResponse); ok {
				cachedResponse = &resp
			}

			if cachedResponse != nil {
				if s.logger != nil {
					s.logger.Debug(ctx, "Resource groups retrieved from cache",
						domain.NewLogField("credential_id", credentialID),
						domain.NewLogField("location", req.Location))
				}
				resourceGroups = cachedResponse.ResourceGroups
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	if resourceGroups == nil {
		response, err := s.listAzureResourceGroups(ctx, credential, req)
		if err != nil {
			return nil, err
		}

		if response != nil {
			resourceGroups = response.ResourceGroups
		}

		// 캐시에 저장
		if s.cacheService != nil && resourceGroups != nil {
			cacheResponse := &ListResourceGroupsResponse{
				ResourceGroups: resourceGroups,
			}
			if err := s.cacheService.Set(ctx, cacheKey, cacheResponse, 300); err != nil {
				if s.logger != nil {
					s.logger.Warn(ctx, "Failed to cache resource groups",
						domain.NewLogField("error", err.Error()))
				}
			}
		}
	}

	// 필터링 및 정렬
	filtered := s.filterResourceGroups(resourceGroups, req)
	sorted := s.sortResourceGroups(filtered, req.SortBy, req.SortOrder)

	// 페이지네이션
	total := int64(len(sorted))
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}

	// 클라이언트 사이드 페이징 지원: limit이 0이거나 매우 큰 값(1000 이상)이면 모든 데이터 반환
	// limit이 0이면 클라이언트 사이드 페이징을 위한 것으로 간주
	if limit == 0 || limit >= 1000 {
		// 모든 데이터 반환 (클라이언트 사이드 페이징)
		return &ListResourceGroupsResponse{
			ResourceGroups: sorted,
			Total:          total,
		}, nil
	}

	// 서버 사이드 페이징: limit 기본값 및 최대값 설정
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(sorted) {
		start = len(sorted)
	}
	if end > len(sorted) {
		end = len(sorted)
	}

	paginated := sorted[start:end]

	return &ListResourceGroupsResponse{
		ResourceGroups: paginated,
		Total:          total,
	}, nil
}

// GetResourceGroup: 특정 Resource Group을 조회합니다
func (s *Service) GetResourceGroup(ctx context.Context, credential *domain.Credential, name string) (*ResourceGroupInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("azure:resource-group:%s:%s", credentialID, name)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if rg, ok := cachedValue.(*ResourceGroupInfo); ok {
				if s.logger != nil {
					s.logger.Debug(ctx, "Resource group retrieved from cache",
						domain.NewLogField("credential_id", credentialID),
						domain.NewLogField("name", name))
				}
				return rg, nil
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	rg, err := s.getAzureResourceGroup(ctx, credential, name)
	if err != nil {
		return nil, err
	}

	// 캐시에 저장
	if s.cacheService != nil && rg != nil {
		if err := s.cacheService.Set(ctx, cacheKey, rg, 300); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to cache resource group",
					domain.NewLogField("error", err.Error()))
			}
		}
	}

	return rg, nil
}

// CreateResourceGroup: 새로운 Resource Group을 생성합니다
func (s *Service) CreateResourceGroup(ctx context.Context, credential *domain.Credential, req CreateResourceGroupRequest) (*ResourceGroupInfo, error) {
	rg, err := s.createAzureResourceGroup(ctx, credential, req)
	if err != nil {
		return nil, err
	}

	// 캐시 무효화
	if s.cacheService != nil {
		credentialID := credential.ID.String()
		cacheKey := fmt.Sprintf("azure:resource-groups:%s:%s", credentialID, req.Location)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to invalidate cache",
					domain.NewLogField("error", err.Error()))
			}
		}
	}

	return rg, nil
}

// UpdateResourceGroup: Resource Group을 업데이트합니다 (태그만 업데이트 가능)
func (s *Service) UpdateResourceGroup(ctx context.Context, credential *domain.Credential, name string, req UpdateResourceGroupRequest) (*ResourceGroupInfo, error) {
	rg, err := s.updateAzureResourceGroup(ctx, credential, name, req)
	if err != nil {
		return nil, err
	}

	// 캐시 무효화
	if s.cacheService != nil {
		credentialID := credential.ID.String()
		cacheKey := fmt.Sprintf("azure:resource-group:%s:%s", credentialID, name)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to invalidate cache",
					domain.NewLogField("error", err.Error()))
			}
		}
		// 목록 캐시도 무효화
		listCacheKey := fmt.Sprintf("azure:resource-groups:%s:", credentialID)
		if err := s.cacheService.Delete(ctx, listCacheKey); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to invalidate list cache",
					domain.NewLogField("error", err.Error()))
			}
		}
	}

	return rg, nil
}

// DeleteResourceGroup: Resource Group을 삭제합니다
func (s *Service) DeleteResourceGroup(ctx context.Context, credential *domain.Credential, name string) error {
	err := s.deleteAzureResourceGroup(ctx, credential, name)
	if err != nil {
		return err
	}

	// 캐시 무효화
	if s.cacheService != nil {
		credentialID := credential.ID.String()
		cacheKey := fmt.Sprintf("azure:resource-group:%s:%s", credentialID, name)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to invalidate cache",
					domain.NewLogField("error", err.Error()))
			}
		}
		// 목록 캐시도 무효화
		listCacheKey := fmt.Sprintf("azure:resource-groups:%s:", credentialID)
		if err := s.cacheService.Delete(ctx, listCacheKey); err != nil {
			if s.logger != nil {
				s.logger.Warn(ctx, "Failed to invalidate list cache",
					domain.NewLogField("error", err.Error()))
			}
		}
	}

	return nil
}

// listAzureResourceGroups: Azure Resource Group 목록을 조회합니다
func (s *Service) listAzureResourceGroups(ctx context.Context, credential *domain.Credential, req ListResourceGroupsRequest) (*ListResourceGroupsResponse, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Resource Manager client
	clientFactory, err := s.createAzureResourceManagerClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get resource groups client
	resourceGroupsClient := clientFactory.NewResourceGroupsClient()

	var resourceGroups []ResourceGroupInfo

	// List all resource groups
	pager := resourceGroupsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "list Azure resource groups")
		}

		for _, rg := range page.Value {
			// Filter by location if provided
			if req.Location != "" && rg.Location != nil && *rg.Location != req.Location {
				continue
			}

			rgInfo := s.buildResourceGroupInfoFromAzure(rg)
			resourceGroups = append(resourceGroups, rgInfo)
		}
	}

	if s.logger != nil {
		s.logger.Info(ctx, "Azure resource groups listed successfully",
			domain.NewLogField("count", len(resourceGroups)),
			domain.NewLogField("location", req.Location))
	}

	return &ListResourceGroupsResponse{ResourceGroups: resourceGroups}, nil
}

// getAzureResourceGroup: 특정 Azure Resource Group을 조회합니다
func (s *Service) getAzureResourceGroup(ctx context.Context, credential *domain.Credential, name string) (*ResourceGroupInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Resource Manager client
	clientFactory, err := s.createAzureResourceManagerClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get resource groups client
	resourceGroupsClient := clientFactory.NewResourceGroupsClient()

	// Get resource group
	resp, err := resourceGroupsClient.Get(ctx, name, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("get Azure resource group: %s", name))
	}

	rgInfo := s.buildResourceGroupInfoFromAzure(&resp.ResourceGroup)

	if s.logger != nil {
		s.logger.Info(ctx, "Azure resource group retrieved successfully",
			domain.NewLogField("name", name))
	}

	return &rgInfo, nil
}

// createAzureResourceGroup: Azure Resource Group을 생성합니다
func (s *Service) createAzureResourceGroup(ctx context.Context, credential *domain.Credential, req CreateResourceGroupRequest) (*ResourceGroupInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Resource Manager client
	clientFactory, err := s.createAzureResourceManagerClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get resource groups client
	resourceGroupsClient := clientFactory.NewResourceGroupsClient()

	// Prepare resource group parameters
	parameters := armresources.ResourceGroup{
		Location: to.Ptr(req.Location),
	}

	// Add tags if provided
	if len(req.Tags) > 0 {
		tags := make(map[string]*string)
		for k, v := range req.Tags {
			tags[k] = to.Ptr(v)
		}
		parameters.Tags = tags
	}

	// Create resource group
	resp, err := resourceGroupsClient.CreateOrUpdate(ctx, req.Name, parameters, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("create Azure resource group: %s", req.Name))
	}

	rgInfo := s.buildResourceGroupInfoFromAzure(&resp.ResourceGroup)

	if s.logger != nil {
		s.logger.Info(ctx, "Azure resource group created successfully",
			domain.NewLogField("name", req.Name),
			domain.NewLogField("location", req.Location))
	}

	return &rgInfo, nil
}

// updateAzureResourceGroup: Azure Resource Group을 업데이트합니다 (태그만)
func (s *Service) updateAzureResourceGroup(ctx context.Context, credential *domain.Credential, name string, req UpdateResourceGroupRequest) (*ResourceGroupInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Resource Manager client
	clientFactory, err := s.createAzureResourceManagerClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get resource groups client
	resourceGroupsClient := clientFactory.NewResourceGroupsClient()

	// Get existing resource group
	existing, err := resourceGroupsClient.Get(ctx, name, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("get Azure resource group for update: %s", name))
	}

	// Prepare update parameters (tags only)
	parameters := armresources.ResourceGroupPatchable{
		Tags: existing.ResourceGroup.Tags,
	}

	// Update tags if provided
	if len(req.Tags) > 0 {
		if parameters.Tags == nil {
			parameters.Tags = make(map[string]*string)
		}
		for k, v := range req.Tags {
			parameters.Tags[k] = to.Ptr(v)
		}
	}

	// Update resource group
	resp, err := resourceGroupsClient.Update(ctx, name, parameters, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("update Azure resource group: %s", name))
	}

	rgInfo := s.buildResourceGroupInfoFromAzure(&resp.ResourceGroup)

	if s.logger != nil {
		s.logger.Info(ctx, "Azure resource group updated successfully",
			domain.NewLogField("name", name))
	}

	return &rgInfo, nil
}

// deleteAzureResourceGroup: Azure Resource Group을 삭제합니다
func (s *Service) deleteAzureResourceGroup(ctx context.Context, credential *domain.Credential, name string) error {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return err
	}

	// Create Azure Resource Manager client
	clientFactory, err := s.createAzureResourceManagerClient(ctx, creds)
	if err != nil {
		return err
	}

	// Get resource groups client
	resourceGroupsClient := clientFactory.NewResourceGroupsClient()

	// Delete resource group (async operation)
	poller, err := resourceGroupsClient.BeginDelete(ctx, name, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("delete Azure resource group: %s", name))
	}

	// Wait for deletion to complete (optional, can be async)
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, fmt.Sprintf("wait for Azure resource group deletion: %s", name))
	}

	if s.logger != nil {
		s.logger.Info(ctx, "Azure resource group deletion initiated",
			domain.NewLogField("name", name))
	}

	return nil
}

// buildResourceGroupInfoFromAzure: Azure ResourceGroup에서 ResourceGroupInfo를 생성합니다
func (s *Service) buildResourceGroupInfoFromAzure(rg *armresources.ResourceGroup) ResourceGroupInfo {
	rgInfo := ResourceGroupInfo{
		ID:   "",
		Name: "",
	}

	if rg.ID != nil {
		rgInfo.ID = *rg.ID
	}
	if rg.Name != nil {
		rgInfo.Name = *rg.Name
	}
	if rg.Location != nil {
		rgInfo.Location = *rg.Location
	}
	if rg.Properties != nil && rg.Properties.ProvisioningState != nil {
		rgInfo.ProvisioningState = *rg.Properties.ProvisioningState
	}
	if rg.Tags != nil {
		tags := make(map[string]string)
		for k, v := range rg.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		rgInfo.Tags = tags
	}

	return rgInfo
}

// filterResourceGroups: Resource Group 목록을 필터링합니다
func (s *Service) filterResourceGroups(rgs []ResourceGroupInfo, req ListResourceGroupsRequest) []ResourceGroupInfo {
	if req.Search == "" && req.Location == "" {
		return rgs
	}

	filtered := make([]ResourceGroupInfo, 0)
	for _, rg := range rgs {
		// Location 필터링
		if req.Location != "" && rg.Location != req.Location {
			continue
		}

		// 검색어 필터링 (이름, ID)
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			nameMatch := strings.Contains(strings.ToLower(rg.Name), searchLower)
			idMatch := strings.Contains(strings.ToLower(rg.ID), searchLower)

			if !nameMatch && !idMatch {
				continue
			}
		}

		filtered = append(filtered, rg)
	}

	return filtered
}

// sortResourceGroups: Resource Group 목록을 정렬합니다
func (s *Service) sortResourceGroups(rgs []ResourceGroupInfo, sortBy, sortOrder string) []ResourceGroupInfo {
	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	sorted := make([]ResourceGroupInfo, len(rgs))
	copy(sorted, rgs)

	switch sortBy {
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].Name > sorted[j].Name
			}
			return sorted[i].Name < sorted[j].Name
		})
	case "location":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].Location > sorted[j].Location
			}
			return sorted[i].Location < sorted[j].Location
		})
	case "provisioning_state":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].ProvisioningState > sorted[j].ProvisioningState
			}
			return sorted[i].ProvisioningState < sorted[j].ProvisioningState
		})
	}

	return sorted
}

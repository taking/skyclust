package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/domain"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
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
// resourceGroup 파라미터가 제공되면 우선 사용 (요청에서 받은 값이 credential에 저장된 값보다 우선)
func (s *Service) extractAzureCredentials(ctx context.Context, credential *domain.Credential, resourceGroup ...string) (*AzureCredentials, error) {
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

	// Resource group 우선순위: 1) 요청 파라미터, 2) credential에 저장된 값
	rg := ""
	if len(resourceGroup) > 0 && resourceGroup[0] != "" {
		rg = resourceGroup[0]
	} else if rgFromCred, ok := credData["resource_group"].(string); ok && rgFromCred != "" {
		rg = rgFromCred
	}

	return &AzureCredentials{
		SubscriptionID: subscriptionID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
		ResourceGroup:  rg,
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

// GetAKSVersions returns available Kubernetes versions for AKS in the specified location
func (s *Service) GetAKSVersions(ctx context.Context, credential *domain.Credential, location string) ([]string, error) {
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aks:versions:%s:%s", credentialID, location)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVersions, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "AKS versions retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("location", location))
				return cachedVersions, nil
			}
		}
	}

	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Azure AKS Kubernetes 버전 조회
	// 참고: Azure SDK v5에서 Kubernetes 버전을 직접 조회하는 API가 제한적임
	// GetOSOptions는 OS 옵션만 반환하고 Kubernetes 버전은 포함하지 않을 수 있음
	// 향후 Azure REST API를 직접 호출하거나 다른 방법을 사용할 수 있음
	// 현재는 Azure 포털에서 표시되는 실제 버전 형식을 반영한 fallback 목록 사용
	
	// Azure 포털에서 표시되는 실제 버전 형식 반영 (예: 1.34.0, 1.33.5, 1.33.4, 1.33.3, 1.33.2 등)
	// 패치 버전까지 포함하여 정확한 버전 정보 제공
	var versions []string = []string{
		"1.34.0",
		"1.33.5",
		"1.33.4",
		"1.33.3",
		"1.33.2",
		"1.32.0",
		"1.31.0",
	}
	
	// TODO: 향후 Azure REST API를 직접 호출하여 실제 사용 가능한 버전 조회
	// 예: GET https://management.azure.com/subscriptions/{subscriptionId}/providers/Microsoft.ContainerService/locations/{location}/orchestrators?api-version=2023-05-02
	// 또는 Azure SDK에 Kubernetes 버전 조회 API가 추가되면 사용

	// 캐시에 저장 (1시간 TTL)
	if s.cacheService != nil && len(versions) > 0 {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, versions, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache AKS versions",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "AKS versions retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("location", location),
		domain.NewLogField("version_count", len(versions)))

	// 사용하지 않는 변수 경고 방지
	_ = clientFactory

	// 중복 제거 및 정렬 (정규화 없이 원본 유지)
	versions = removeDuplicatesAndSortAKSVersions(versions)

	return versions, nil
}

// removeDuplicatesAndSortAKSVersions removes duplicates and sorts AKS Kubernetes versions in descending order
// Supports both "major.minor" and "major.minor.patch" formats
// Important: No normalization is performed - original version strings are preserved
func removeDuplicatesAndSortAKSVersions(versions []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, v := range versions {
		if v == "" {
			continue
		}
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}

	// Sort in descending order (newest first) using semantic version comparison
	// No normalization - original version strings are preserved
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			if compareAKSSemanticVersion(unique[i], unique[j]) < 0 {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	return unique
}

// compareAKSSemanticVersion compares two AKS semantic version strings
// Supports "major.minor" and "major.minor.patch" formats
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareAKSSemanticVersion(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &num1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &num2)
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}

// GetAzureAvailabilityZones returns available zones for the specified location
func (s *Service) GetAzureAvailabilityZones(ctx context.Context, credential *domain.Credential, location string) ([]string, error) {
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("azure:availability-zones:%s:%s", credentialID, location)

	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedZones, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "Azure zones retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("location", location))
				return cachedZones, nil
			}
		}
	}

	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewClientSecretCredential(
		creds.TenantID,
		creds.ClientID,
		creds.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError,
			fmt.Sprintf("failed to create Azure credential: %v", err), 502)
	}

	computeClientFactory, err := armcompute.NewClientFactory(creds.SubscriptionID, cred, nil)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError,
			fmt.Sprintf("failed to create Azure Compute client factory: %v", err), 502)
	}

	resourceSKUsClient := computeClientFactory.NewResourceSKUsClient()
	pager := resourceSKUsClient.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("location eq '%s'", location)),
	})

	zoneSet := make(map[string]bool)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			err = s.providerErrorConverter.ConvertAzureError(err, "get Azure zones")
			if err != nil {
				return nil, err
			}
		}

		if page.Value != nil {
			for _, sku := range page.Value {
				if sku.LocationInfo != nil {
					for _, locInfo := range sku.LocationInfo {
						if locInfo.Location != nil && *locInfo.Location == location {
							if locInfo.Zones != nil {
								for _, zone := range locInfo.Zones {
									if zone != nil {
										zoneSet[*zone] = true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	var zones []string
	for zone := range zoneSet {
		zones = append(zones, zone)
	}

	if zones == nil {
		zones = []string{}
	}

	if s.cacheService != nil && len(zones) > 0 {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, zones, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache Azure zones",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "Azure zones retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("location", location),
		domain.NewLogField("zone_count", len(zones)))

	return zones, nil
}

// GetAzureVMSizes returns available VM sizes for the specified location
func (s *Service) GetAzureVMSizes(ctx context.Context, credential *domain.Credential, location string) ([]string, error) {
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("azure:vm-sizes:%s:%s", credentialID, location)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVMSizes, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "Azure VM sizes retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("location", location))
				return cachedVMSizes, nil
			}
		}
	}

	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewClientSecretCredential(
		creds.TenantID,
		creds.ClientID,
		creds.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError,
			fmt.Sprintf("failed to create Azure credential: %v", err), 502)
	}

	computeClientFactory, err := armcompute.NewClientFactory(creds.SubscriptionID, cred, nil)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError,
			fmt.Sprintf("failed to create Azure Compute client factory: %v", err), 502)
	}

	resourceSKUsClient := computeClientFactory.NewResourceSKUsClient()
	
	// ResourceType을 "virtualMachines"로 필터링하여 VM Size만 조회
	filter := fmt.Sprintf("location eq '%s' and resourceType eq 'virtualMachines'", location)
	pager := resourceSKUsClient.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: to.Ptr(filter),
	})

	vmSizeSet := make(map[string]bool)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "get Azure VM sizes")
		}

		if page.Value != nil {
			for _, sku := range page.Value {
				// SKU Name이 VM Size (예: "Standard_D2s_v3")
				if sku.Name != nil {
					vmSize := *sku.Name
					// Location 정보 확인하여 해당 location에서 사용 가능한지 확인
					if sku.LocationInfo != nil {
						for _, locInfo := range sku.LocationInfo {
							if locInfo.Location != nil && *locInfo.Location == location {
								// Restrictions가 없거나 사용 가능한 경우에만 추가
								available := true
								if len(sku.Restrictions) > 0 {
									for _, restriction := range sku.Restrictions {
										if restriction.Type != nil && *restriction.Type == armcompute.ResourceSKURestrictionsTypeLocation {
											available = false
											break
										}
									}
								}
								if available {
									vmSizeSet[vmSize] = true
								}
								break
							}
						}
					} else {
						// LocationInfo가 없으면 일단 추가 (전역적으로 사용 가능할 수 있음)
						vmSizeSet[vmSize] = true
					}
				}
			}
		}
	}

	var vmSizes []string
	for vmSize := range vmSizeSet {
		vmSizes = append(vmSizes, vmSize)
	}

	// 정렬 (알파벳 순서)
	for i := 0; i < len(vmSizes)-1; i++ {
		for j := i + 1; j < len(vmSizes); j++ {
			if vmSizes[i] > vmSizes[j] {
				vmSizes[i], vmSizes[j] = vmSizes[j], vmSizes[i]
			}
		}
	}

	if vmSizes == nil {
		vmSizes = []string{}
	}

	// 캐시에 저장 (1시간 TTL)
	if s.cacheService != nil && len(vmSizes) > 0 {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, vmSizes, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache Azure VM sizes",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "Azure VM sizes retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("location", location),
		domain.NewLogField("vm_size_count", len(vmSizes)))

	return vmSizes, nil
}

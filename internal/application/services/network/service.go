package network

import (
	"context"
	"fmt"
	"strings"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	providererrors "skyclust/internal/shared/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// Service: 네트워크 리소스 작업을 처리하는 서비스
type Service struct {
	credentialService      domain.CredentialService
	cacheService           domain.CacheService
	eventService           domain.EventService
	auditLogRepo           domain.AuditLogRepository
	logger                 domain.LoggerService
	providerErrorConverter *providererrors.ProviderErrorConverter
}

// NewService: 새로운 네트워크 서비스를 생성합니다
func NewService(
	credentialService domain.CredentialService,
	cacheService domain.CacheService,
	eventService domain.EventService,
	auditLogRepo domain.AuditLogRepository,
	logger domain.LoggerService,
) *Service {
	return &Service{
		credentialService:      credentialService,
		cacheService:           cacheService,
		eventService:           eventService,
		auditLogRepo:           auditLogRepo,
		logger:                 logger,
		providerErrorConverter: providererrors.NewProviderErrorConverter(),
	}
}

// ListVPCs: 주어진 자격증명과 리전에 대한 VPC 목록을 조회합니다
func (s *Service) ListVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// 캐시 키 생성 (필터링/정렬/페이지네이션 파라미터는 제외)
	credentialID := credential.ID.String()
	cacheKey := buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)

	var allVPCs []VPCInfo

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			var cachedResponse *ListVPCsResponse
			if resp, ok := cachedValue.(*ListVPCsResponse); ok {
				cachedResponse = resp
			} else if resp, ok := cachedValue.(ListVPCsResponse); ok {
				cachedResponse = &resp
			}

			if cachedResponse != nil {
				s.logger.Debug(ctx, "VPCs retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", req.Region))
				allVPCs = cachedResponse.VPCs
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	if allVPCs == nil {
		var response *ListVPCsResponse
		var err error

		switch credential.Provider {
		case domain.ProviderAWS:
			response, err = s.listAWSVPCs(ctx, credential, req)
		case domain.ProviderGCP:
			response, err = s.listGCPVPCs(ctx, credential, req)
		case domain.ProviderAzure:
			response, err = s.listAzureVPCs(ctx, credential, req)
		case domain.ProviderNCP:
			response, err = s.listNCPVPCs(ctx, credential, req)
		default:
			return nil, domain.NewDomainError(
				domain.ErrCodeNotSupported,
				fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
				400,
			)
		}

		if err != nil {
			return nil, err
		}

		if response != nil {
			allVPCs = response.VPCs
		} else {
			allVPCs = []VPCInfo{}
		}

		// 전체 데이터를 캐시에 저장 (필터링/정렬/페이지네이션 전)
		if s.cacheService != nil {
			fullResponse := &ListVPCsResponse{VPCs: allVPCs}
			if err := s.cacheService.Set(ctx, cacheKey, fullResponse, defaultNetworkTTL); err != nil {
				s.logger.Warn(ctx, "Failed to cache VPCs, continuing without cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", req.Region),
					domain.NewLogField("error", err))
			}
		}
	}

	// 필터링 적용
	filteredVPCs := applyVPCFiltering(allVPCs, req.Search)
	totalCount := int64(len(filteredVPCs))

	// 정렬 적용
	applyVPCSorting(filteredVPCs, req.SortBy, req.SortOrder)

	// 페이지네이션 적용
	paginatedVPCs := applyVPCPagination(filteredVPCs, req.Page, req.Limit)

	return &ListVPCsResponse{
		VPCs:  paginatedVPCs,
		Total: totalCount,
	}, nil
}

// GetVPC: ID로 특정 VPC를 조회합니다
func (s *Service) GetVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := buildNetworkVPCItemKey(credential.Provider, credentialID, req.VPCID)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVPC, ok := cachedValue.(*VPCInfo); ok {
				s.logger.Debug(ctx, "VPC retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("vpc_id", req.VPCID))
				return cachedVPC, nil
			} else if cachedVPC, ok := cachedValue.(VPCInfo); ok {
				s.logger.Debug(ctx, "VPC retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("vpc_id", req.VPCID))
				return &cachedVPC, nil
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	var vpc *VPCInfo
	var err error

	switch credential.Provider {
	case "aws":
		vpc, err = s.getAWSVPC(ctx, credential, req)
	case "gcp":
		vpc, err = s.getGCPVPC(ctx, credential, req)
	case "azure":
		vpc, err = s.getAzureVPC(ctx, credential, req)
	case "ncp":
		vpc, err = s.getNCPVPC(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장
	if s.cacheService != nil && vpc != nil {
		if err := s.cacheService.Set(ctx, cacheKey, vpc, defaultNetworkTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache VPC",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", req.VPCID),
				domain.NewLogField("error", err))
		}
	}

	return vpc, nil
}

// CreateVPC: 새로운 VPC를 생성합니다
func (s *Service) CreateVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	// Route to provider-specific implementation
	var vpc *VPCInfo
	var err error

	switch credential.Provider {
	case "aws":
		vpc, err = s.createAWSVPC(ctx, credential, req)
	case "gcp":
		vpc, err = s.createGCPVPC(ctx, credential, req)
	case "azure":
		vpc, err = s.createAzureVPC(ctx, credential, req)
	case "ncp":
		vpc, err = s.createNCPVPC(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 캐시 무효화: VPC 목록 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: VPC 생성 이벤트
	if s.eventService != nil {
		vpcData := map[string]interface{}{
			"vpc_id":        vpc.ID,
			"name":          vpc.Name,
			"state":         vpc.State,
			"region":        vpc.Region,
			"provider":      credential.Provider,
			"credential_id": credentialID,
		}
		eventType := fmt.Sprintf("network.vpc.%s.created", credential.Provider)
		if err := s.eventService.Publish(ctx, eventType, vpcData); err != nil {
			s.logger.Warn(ctx, "Failed to publish VPC created event",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", vpc.ID),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs", credential.Provider),
		map[string]interface{}{
			"vpc_id":        vpc.ID,
			"name":          vpc.Name,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        vpc.Region,
		},
	)

	return vpc, nil
}

// createGCPVPC: GCP VPC를 생성합니다

// createGCPVPCWithAdvanced: 고급 설정으로 GCP VPC를 생성합니다

// buildGCPNetworkObject: GCP 네트워크 객체를 생성합니다

// logNetworkConfiguration: 네트워크 구성을 로깅합니다

// createGCPNetworkOperation: GCP 네트워크 생성 작업을 시작합니다

// logOperationInitiated: 작업 시작을 로깅합니다

// buildVPCInfoFromRequest: 요청으로부터 VPC 정보를 생성합니다

// listGCPVPCs: GCP VPC 목록을 조회합니다

// getGCPVPC: 특정 GCP VPC를 조회합니다

// extractNetworkNameFromVPCID: VPC ID에서 네트워크 이름을 추출합니다

// deleteGCPVPC: GCP VPC를 삭제합니다

// cleanupVPCResources: VPC 리소스를 정리합니다

// deleteNetworkFirewallRules: 네트워크 방화벽 규칙을 삭제합니다

// deleteNetworkSubnets: 네트워크 서브넷을 삭제합니다

// checkNetworkInstances: 네트워크에 연결된 인스턴스를 확인합니다

// checkVPCDeletionDependencies checks if VPC can be safely deleted
// Stub implementations for Azure, NCP, and GCP update functions

// listAzureVPCs: Azure Virtual Network 목록을 조회합니다
func (s *Service) listAzureVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get virtual networks client
	virtualNetworksClient := clientFactory.NewVirtualNetworksClient()

	// List all virtual networks in the subscription
	var vpcs []VPCInfo
	pager := virtualNetworksClient.NewListAllPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "list Azure virtual networks")
		}

		for _, vnet := range page.Value {
			// Filter by location if provided
			if req.Region != "" && vnet.Location != nil && *vnet.Location != req.Region {
				continue
			}

			// Filter by VPC ID if provided
			if req.VPCID != "" && vnet.ID != nil && *vnet.ID != req.VPCID {
				continue
			}

			vpcInfo := VPCInfo{
				ID:        "",
				Name:      "",
				State:     "Succeeded",
				IsDefault: false,
			}

			if vnet.ID != nil {
				vpcInfo.ID = *vnet.ID
			}
			if vnet.Name != nil {
				vpcInfo.Name = *vnet.Name
			}
			if vnet.Location != nil {
				vpcInfo.Region = *vnet.Location
			}

			// Extract CIDR blocks from address space
			if vnet.Properties != nil && vnet.Properties.AddressSpace != nil && len(vnet.Properties.AddressSpace.AddressPrefixes) > 0 {
				// Use first CIDR block
				if vnet.Properties.AddressSpace.AddressPrefixes[0] != nil {
					// Note: VPCInfo doesn't have CIDRBlock field, but we can add it if needed
				}
			}

			// Extract tags
			if vnet.Tags != nil {
				tags := make(map[string]string)
				for k, v := range vnet.Tags {
					if v != nil {
						tags[k] = *v
					}
				}
				vpcInfo.Tags = tags
			}

			vpcs = append(vpcs, vpcInfo)
		}
	}

	return &ListVPCsResponse{VPCs: vpcs}, nil
}

// listNCPVPCs: NCP VPC 목록을 조회합니다
func (s *Service) listNCPVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	s.logger.Info(ctx, "NCP VPC listing not yet implemented")
	return &ListVPCsResponse{VPCs: []VPCInfo{}}, nil
}

// getAzureVPC: 특정 Azure Virtual Network를 조회합니다
func (s *Service) getAzureVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group from VPC ID
	// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}
	resourceGroup := ""
	vnetName := ""
	parts := strings.Split(req.VPCID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" {
		// If VPCID is just the name, try to find it by listing
		if creds.ResourceGroup != "" {
			resourceGroup = creds.ResourceGroup
			vnetName = req.VPCID
		} else {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, ErrMsgInvalidVPCIDOrResourceGroup, 400)
		}
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get virtual networks client
	virtualNetworksClient := clientFactory.NewVirtualNetworksClient()

	// Get virtual network details
	vnet, err := virtualNetworksClient.Get(ctx, resourceGroup, vnetName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get Azure virtual network")
	}

	vpcInfo := VPCInfo{
		ID:        "",
		Name:      "",
		State:     "Succeeded",
		IsDefault: false,
	}

	if vnet.ID != nil {
		vpcInfo.ID = *vnet.ID
	}
	if vnet.Name != nil {
		vpcInfo.Name = *vnet.Name
	}
	if vnet.Location != nil {
		vpcInfo.Region = *vnet.Location
	}

	// Extract tags
	if vnet.Tags != nil {
		tags := make(map[string]string)
		for k, v := range vnet.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		vpcInfo.Tags = tags
	}

	return &vpcInfo, nil
}

// getNCPVPC: 특정 NCP VPC를 조회합니다
func (s *Service) getNCPVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	s.logger.Info(ctx, "NCP VPC retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC retrieval not yet implemented", 501)
}

// createAzureVPC: Azure Virtual Network를 생성합니다
func (s *Service) createAzureVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Resource group is required for Azure
	if creds.ResourceGroup == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgResourceGroupRequired, 400)
	}

	// Location is required for Azure
	if req.Region == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgRegionRequiredAzure, 400)
	}

	// CIDR block is required for Azure
	if req.CIDRBlock == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgCIDRBlockRequired, 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get virtual networks client
	virtualNetworksClient := clientFactory.NewVirtualNetworksClient()

	// Build virtual network
	virtualNetwork := armnetwork.VirtualNetwork{
		Location: to.Ptr(req.Region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{to.Ptr(req.CIDRBlock)},
			},
		},
	}

	// Add tags
	if len(req.Tags) > 0 {
		tags := make(map[string]*string)
		for k, v := range req.Tags {
			tags[k] = to.Ptr(v)
		}
		virtualNetwork.Tags = tags
	}

	// Create virtual network
	poller, err := virtualNetworksClient.BeginCreateOrUpdate(ctx, creds.ResourceGroup, req.Name, virtualNetwork, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure virtual network")
	}

	// Wait for completion
	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure virtual network")
	}

	vpcInfo := VPCInfo{
		ID:        "",
		Name:      req.Name,
		State:     "Succeeded",
		IsDefault: false,
		Region:    req.Region,
	}

	if result.ID != nil {
		vpcInfo.ID = *result.ID
	}

	if result.Tags != nil {
		tags := make(map[string]string)
		for k, v := range result.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		vpcInfo.Tags = tags
	}

	s.logger.Info(ctx, "Azure Virtual Network creation completed",
		domain.NewLogField("vpc_name", req.Name),
		domain.NewLogField("resource_group", creds.ResourceGroup),
		domain.NewLogField("location", req.Region))

	// 캐시 무효화: VPC 목록 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs", credential.Provider),
		map[string]interface{}{
			"vpc_id":         vpcInfo.ID,
			"name":           req.Name,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
			"resource_group": creds.ResourceGroup,
		},
	)

	// 이벤트 발행
	if s.eventService != nil {
		vpcData := map[string]interface{}{
			"vpc_id":        vpcInfo.ID,
			"name":          req.Name,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			vpcData := vpcData
			vpcData["provider"] = credential.Provider
			vpcData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.vpc.%s.created", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, vpcData)
		}
	}

	return &vpcInfo, nil
}

// createNCPVPC: NCP VPC를 생성합니다
func (s *Service) createNCPVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	s.logger.Info(ctx, "NCP VPC creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC creation not yet implemented", 501)
}

// updateGCPVPC: GCP VPC를 업데이트합니다

// updateAzureVPC: Azure Virtual Network를 업데이트합니다
func (s *Service) updateAzureVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group and VNet name from VPC ID
	resourceGroup := ""
	vnetName := ""
	parts := strings.Split(vpcID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" {
		if creds.ResourceGroup != "" {
			resourceGroup = creds.ResourceGroup
			vnetName = vpcID
		} else {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, ErrMsgInvalidVPCIDOrResourceGroup, 400)
		}
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get virtual networks client
	virtualNetworksClient := clientFactory.NewVirtualNetworksClient()

	// Get existing virtual network
	existingVNet, err := virtualNetworksClient.Get(ctx, resourceGroup, vnetName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get Azure virtual network")
	}

	// Update tags if provided
	if len(req.Tags) > 0 {
		if existingVNet.VirtualNetwork.Tags == nil {
			existingVNet.VirtualNetwork.Tags = make(map[string]*string)
		}
		for k, v := range req.Tags {
			existingVNet.VirtualNetwork.Tags[k] = to.Ptr(v)
		}
	}

	// Update virtual network
	poller, err := virtualNetworksClient.BeginCreateOrUpdate(ctx, resourceGroup, vnetName, existingVNet.VirtualNetwork, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "update Azure virtual network")
	}

	// Wait for completion
	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "update Azure virtual network")
	}

	vpcInfo := VPCInfo{
		ID:        "",
		Name:      vnetName,
		State:     "Succeeded",
		IsDefault: false,
		Region:    region,
	}

	if result.VirtualNetwork.ID != nil {
		vpcInfo.ID = *result.VirtualNetwork.ID
	}
	if result.VirtualNetwork.Location != nil {
		vpcInfo.Region = *result.VirtualNetwork.Location
	}

	if result.VirtualNetwork.Tags != nil {
		tags := make(map[string]string)
		for k, v := range result.VirtualNetwork.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		vpcInfo.Tags = tags
	}

	return &vpcInfo, nil
}

// updateNCPVPC: NCP VPC를 업데이트합니다
func (s *Service) updateNCPVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	s.logger.Info(ctx, "NCP VPC update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC update not yet implemented", 501)
}

// deleteAzureVPC: Azure Virtual Network를 삭제합니다
func (s *Service) deleteAzureVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return err
	}

	// Extract resource group and VNet name from VPC ID
	resourceGroup := ""
	vnetName := ""
	parts := strings.Split(req.VPCID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" {
		if creds.ResourceGroup != "" {
			resourceGroup = creds.ResourceGroup
			vnetName = req.VPCID
		} else {
			return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid VPC ID format or resource group not found", 400)
		}
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return err
	}

	// Get virtual networks client
	virtualNetworksClient := clientFactory.NewVirtualNetworksClient()

	// Delete virtual network
	poller, err := virtualNetworksClient.BeginDelete(ctx, resourceGroup, vnetName, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete Azure virtual network")
	}

	// Wait for completion
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete Azure virtual network")
	}

	s.logger.Info(ctx, "Azure Virtual Network deletion completed",
		domain.NewLogField("vpc_name", vnetName),
		domain.NewLogField("resource_group", resourceGroup))

	// 캐시 무효화: VPC 목록 및 개별 VPC 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
		itemKey := buildNetworkVPCItemKey(credential.Provider, credentialID, req.VPCID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", req.VPCID),
				domain.NewLogField("error", err))
		}
		listKey = buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/vpcs/%s", credential.Provider, req.VPCID),
		map[string]interface{}{
			"vpc_id":         req.VPCID,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
			"resource_group": resourceGroup,
		},
	)

	// 이벤트 발행
	if s.eventService != nil {
		vpcData := map[string]interface{}{
			"vpc_id":        req.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			vpcData := vpcData
			vpcData["provider"] = credential.Provider
			vpcData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.vpc.%s.deleted", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, vpcData)
		}
	}

	return nil
}

// deleteNCPVPC: NCP VPC를 삭제합니다
func (s *Service) deleteNCPVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	s.logger.Info(ctx, "NCP VPC deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC deletion not yet implemented", 501)
}

// UpdateVPC updates a VPC
func (s *Service) UpdateVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.updateAWSVPC(ctx, credential, req, vpcID, region)
	case "gcp":
		return s.updateGCPVPC(ctx, credential, req, vpcID, region)
	case "azure":
		return s.updateAzureVPC(ctx, credential, req, vpcID, region)
	case "ncp":
		return s.updateNCPVPC(ctx, credential, req, vpcID, region)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// DeleteVPC: VPC를 삭제합니다
func (s *Service) DeleteVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	// Route to provider-specific implementation
	var err error

	switch credential.Provider {
	case "aws":
		err = s.deleteAWSVPC(ctx, credential, req)
	case "gcp":
		err = s.deleteGCPVPC(ctx, credential, req)
	case "azure":
		err = s.deleteAzureVPC(ctx, credential, req)
	case "ncp":
		err = s.deleteNCPVPC(ctx, credential, req)
	default:
		return domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return err
	}

	// 캐시 무효화: VPC 목록 및 개별 VPC 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildNetworkVPCListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
		itemKey := buildNetworkVPCItemKey(credential.Provider, credentialID, req.VPCID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", req.VPCID),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: VPC 삭제 이벤트
	vpcData := map[string]interface{}{
		"vpc_id": req.VPCID,
		"region": req.Region,
	}
	if s.eventService != nil {
		vpcData["provider"] = credential.Provider
		vpcData["credential_id"] = credentialID
		eventType := fmt.Sprintf("network.vpc.%s.deleted", credential.Provider)
		if err := s.eventService.Publish(ctx, eventType, vpcData); err != nil {
			s.logger.Warn(ctx, "Failed to publish VPC deleted event",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", req.VPCID),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/vpcs/%s", credential.Provider, req.VPCID),
		map[string]interface{}{
			"vpc_id":        req.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	return nil
}

// ListSubnets: 서브넷 목록을 조회합니다
func (s *Service) ListSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	// 캐시 키 생성 (필터링/정렬/페이지네이션 파라미터는 제외)
	credentialID := credential.ID.String()
	cacheKey := buildNetworkSubnetListKey(credential.Provider, credentialID, req.VPCID)

	var allSubnets []SubnetInfo

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			var cachedResponse *ListSubnetsResponse
			if resp, ok := cachedValue.(*ListSubnetsResponse); ok {
				cachedResponse = resp
			} else if resp, ok := cachedValue.(ListSubnetsResponse); ok {
				cachedResponse = &resp
			}

			if cachedResponse != nil {
				s.logger.Debug(ctx, "Subnets retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("vpc_id", req.VPCID))
				allSubnets = cachedResponse.Subnets
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	if allSubnets == nil {
		var response *ListSubnetsResponse
		var err error

		switch credential.Provider {
		case "aws":
			response, err = s.listAWSSubnets(ctx, credential, req)
		case "gcp":
			response, err = s.listGCPSubnets(ctx, credential, req)
		case "azure":
			response, err = s.listAzureSubnets(ctx, credential, req)
		case "ncp":
			response, err = s.listNCPSubnets(ctx, credential, req)
		default:
			return nil, domain.NewDomainError(
				domain.ErrCodeNotSupported,
				fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
				400,
			)
		}

		if err != nil {
			return nil, err
		}

		if response != nil {
			allSubnets = response.Subnets
		} else {
			allSubnets = []SubnetInfo{}
		}

		// 전체 데이터를 캐시에 저장 (필터링/정렬/페이지네이션 전)
		if s.cacheService != nil {
			fullResponse := &ListSubnetsResponse{Subnets: allSubnets}
			if err := s.cacheService.Set(ctx, cacheKey, fullResponse, defaultNetworkTTL); err != nil {
				s.logger.Warn(ctx, "Failed to cache subnets, continuing without cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("vpc_id", req.VPCID),
					domain.NewLogField("error", err))
			}
		}
	}

	// 필터링 적용
	filteredSubnets := applySubnetFiltering(allSubnets, req.Search)
	totalCount := int64(len(filteredSubnets))

	// 정렬 적용
	applySubnetSorting(filteredSubnets, req.SortBy, req.SortOrder)

	// 페이지네이션 적용
	paginatedSubnets := applySubnetPagination(filteredSubnets, req.Page, req.Limit)

	return &ListSubnetsResponse{
		Subnets: paginatedSubnets,
		Total:   totalCount,
	}, nil
}

// GetSubnet: 특정 서브넷을 조회합니다
func (s *Service) GetSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := buildNetworkSubnetItemKey(credential.Provider, credentialID, req.SubnetID)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedSubnet, ok := cachedValue.(*SubnetInfo); ok {
				s.logger.Debug(ctx, "Subnet retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("subnet_id", req.SubnetID))
				return cachedSubnet, nil
			} else if cachedSubnet, ok := cachedValue.(SubnetInfo); ok {
				s.logger.Debug(ctx, "Subnet retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("subnet_id", req.SubnetID))
				return &cachedSubnet, nil
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	var subnet *SubnetInfo
	var err error

	switch credential.Provider {
	case "aws":
		subnet, err = s.getAWSSubnet(ctx, credential, req)
	case "gcp":
		subnet, err = s.getGCPSubnet(ctx, credential, req)
	case "azure":
		subnet, err = s.getAzureSubnet(ctx, credential, req)
	case "ncp":
		subnet, err = s.getNCPSubnet(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장
	if s.cacheService != nil && subnet != nil {
		if err := s.cacheService.Set(ctx, cacheKey, subnet, defaultNetworkTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache subnet",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("subnet_id", req.SubnetID),
				domain.NewLogField("error", err))
		}
	}

	return subnet, nil
}

// CreateSubnet: 새로운 서브넷을 생성합니다
func (s *Service) CreateSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	var subnet *SubnetInfo
	var err error

	switch credential.Provider {
	case "aws":
		subnet, err = s.createAWSSubnet(ctx, credential, req)
	case "gcp":
		subnet, err = s.createGCPSubnet(ctx, credential, req)
	case "azure":
		subnet, err = s.createAzureSubnet(ctx, credential, req)
	case "ncp":
		subnet, err = s.createNCPSubnet(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 캐시 무효화: Subnet 목록 및 개별 Subnet 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil && subnet != nil {
		// Subnet 목록 캐시 무효화
		listKey := buildNetworkSubnetListKey(credential.Provider, credentialID, subnet.VPCID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", subnet.VPCID),
				domain.NewLogField("error", err))
		}
		// 개별 Subnet 캐시 무효화
		itemKey := buildNetworkSubnetItemKey(credential.Provider, credentialID, subnet.ID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("subnet_id", subnet.ID),
				domain.NewLogField("error", err))
		}
	}

	return subnet, nil
}

// UpdateSubnet: 서브넷을 업데이트합니다
func (s *Service) UpdateSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	// 먼저 기존 Subnet 정보 조회 (VPCID 확인용)
	getReq := GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}
	existingSubnet, err := s.GetSubnet(ctx, credential, getReq)
	if err != nil {
		return nil, err
	}

	var subnet *SubnetInfo
	switch credential.Provider {
	case "aws":
		subnet, err = s.updateAWSSubnet(ctx, credential, req, subnetID, region)
	case "gcp":
		subnet, err = s.updateGCPSubnet(ctx, credential, req, subnetID, region)
	case "azure":
		subnet, err = s.updateAzureSubnet(ctx, credential, req, subnetID, region)
	case "ncp":
		subnet, err = s.updateNCPSubnet(ctx, credential, req, subnetID, region)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 캐시 무효화: Subnet 목록 및 개별 Subnet 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		// Subnet 목록 캐시 무효화
		vpcID := existingSubnet.VPCID
		if subnet != nil && subnet.VPCID != "" {
			vpcID = subnet.VPCID
		}
		listKey := buildNetworkSubnetListKey(credential.Provider, credentialID, vpcID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", vpcID),
				domain.NewLogField("error", err))
		}
		// 개별 Subnet 캐시 무효화
		itemKey := buildNetworkSubnetItemKey(credential.Provider, credentialID, subnetID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("subnet_id", subnetID),
				domain.NewLogField("error", err))
		}
	}

	return subnet, nil
}

// DeleteSubnet: 서브넷을 삭제합니다
func (s *Service) DeleteSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	// 먼저 기존 Subnet 정보 조회 (VPCID 확인용)
	getReq := GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     req.SubnetID,
		Region:       req.Region,
	}
	existingSubnet, err := s.GetSubnet(ctx, credential, getReq)
	if err != nil {
		// Subnet이 없어도 삭제는 진행 (이미 삭제된 경우)
	}

	var deleteErr error
	switch credential.Provider {
	case "aws":
		deleteErr = s.deleteAWSSubnet(ctx, credential, req)
	case "gcp":
		deleteErr = s.deleteGCPSubnet(ctx, credential, req)
	case "azure":
		deleteErr = s.deleteAzureSubnet(ctx, credential, req)
	case "ncp":
		deleteErr = s.deleteNCPSubnet(ctx, credential, req)
	default:
		return domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if deleteErr != nil {
		return deleteErr
	}

	// 캐시 무효화: Subnet 목록 및 개별 Subnet 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil && existingSubnet != nil {
		// Subnet 목록 캐시 무효화
		listKey := buildNetworkSubnetListKey(credential.Provider, credentialID, existingSubnet.VPCID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", existingSubnet.VPCID),
				domain.NewLogField("error", err))
		}
		// 개별 Subnet 캐시 무효화
		itemKey := buildNetworkSubnetItemKey(credential.Provider, credentialID, req.SubnetID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate subnet item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("subnet_id", req.SubnetID),
				domain.NewLogField("error", err))
		}
	}

	return nil
}

// ListSecurityGroups: 보안 그룹 목록을 조회합니다
func (s *Service) ListSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	// 캐시 키 생성 (필터링/정렬/페이지네이션 파라미터는 제외)
	credentialID := credential.ID.String()
	cacheKey := buildNetworkSecurityGroupListKey(credential.Provider, credentialID, req.Region)

	var allSecurityGroups []SecurityGroupInfo

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			var cachedResponse *ListSecurityGroupsResponse
			if resp, ok := cachedValue.(*ListSecurityGroupsResponse); ok {
				cachedResponse = resp
			} else if resp, ok := cachedValue.(ListSecurityGroupsResponse); ok {
				cachedResponse = &resp
			}

			if cachedResponse != nil {
				s.logger.Debug(ctx, "Security groups retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", req.Region))
				allSecurityGroups = cachedResponse.SecurityGroups
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	if allSecurityGroups == nil {
		var response *ListSecurityGroupsResponse
		var err error

		switch credential.Provider {
		case "aws":
			response, err = s.listAWSSecurityGroups(ctx, credential, req)
		case "gcp":
			response, err = s.listGCPSecurityGroups(ctx, credential, req)
		case "azure":
			return nil, domain.NewDomainError(
				domain.ErrCodeNotImplemented,
				fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
				501,
			)
		case "ncp":
			return nil, domain.NewDomainError(
				domain.ErrCodeNotImplemented,
				fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
				501,
			)
		default:
			return nil, domain.NewDomainError(
				domain.ErrCodeNotSupported,
				fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
				400,
			)
		}

		if err != nil {
			return nil, err
		}

		if response != nil {
			allSecurityGroups = response.SecurityGroups
		} else {
			allSecurityGroups = []SecurityGroupInfo{}
		}

		// 전체 데이터를 캐시에 저장 (필터링/정렬/페이지네이션 전)
		if s.cacheService != nil {
			fullResponse := &ListSecurityGroupsResponse{SecurityGroups: allSecurityGroups}
			if err := s.cacheService.Set(ctx, cacheKey, fullResponse, defaultNetworkTTL); err != nil {
				s.logger.Warn(ctx, "Failed to cache security groups, continuing without cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", req.Region),
					domain.NewLogField("error", err))
			}
		}
	}

	// 필터링 적용
	filteredSecurityGroups := applySecurityGroupFiltering(allSecurityGroups, req.Search)
	totalCount := int64(len(filteredSecurityGroups))

	// 정렬 적용
	applySecurityGroupSorting(filteredSecurityGroups, req.SortBy, req.SortOrder)

	// 페이지네이션 적용
	paginatedSecurityGroups := applySecurityGroupPagination(filteredSecurityGroups, req.Page, req.Limit)

	return &ListSecurityGroupsResponse{
		SecurityGroups: paginatedSecurityGroups,
		Total:          totalCount,
	}, nil
}

// listGCPSecurityGroups: GCP 보안 그룹 목록을 조회합니다

// convertGCPFirewallRules: GCP 방화벽 규칙을 보안 그룹 규칙 정보로 변환합니다

// GetSecurityGroup: 특정 보안 그룹을 조회합니다
func (s *Service) GetSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := buildNetworkSecurityGroupItemKey(credential.Provider, credentialID, req.SecurityGroupID)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedSG, ok := cachedValue.(*SecurityGroupInfo); ok {
				s.logger.Debug(ctx, "Security group retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("security_group_id", req.SecurityGroupID))
				return cachedSG, nil
			} else if cachedSG, ok := cachedValue.(SecurityGroupInfo); ok {
				s.logger.Debug(ctx, "Security group retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("security_group_id", req.SecurityGroupID))
				return &cachedSG, nil
			}
		}
	}

	// 캐시 미스 시 실제 API 호출
	var sg *SecurityGroupInfo
	var err error

	switch credential.Provider {
	case "aws":
		sg, err = s.getAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		sg, err = s.getGCPSecurityGroup(ctx, credential, req)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장
	if s.cacheService != nil && sg != nil {
		if err := s.cacheService.Set(ctx, cacheKey, sg, defaultNetworkTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache security group",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("security_group_id", req.SecurityGroupID),
				domain.NewLogField("error", err))
		}
	}

	return sg, nil
}

// getGCPSecurityGroup: 특정 GCP 보안 그룹을 조회합니다

// CreateSecurityGroup: 새로운 보안 그룹을 생성합니다
func (s *Service) CreateSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	var sg *SecurityGroupInfo
	var err error

	switch credential.Provider {
	case "aws":
		sg, err = s.createAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		sg, err = s.createGCPSecurityGroup(ctx, credential, req)
	case "azure":
		sg, err = s.createAzureSecurityGroup(ctx, credential, req)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 캐시 무효화: Security Group 목록 및 개별 Security Group 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil && sg != nil {
		// Security Group 목록 캐시 무효화
		listKey := buildNetworkSecurityGroupListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
		// 개별 Security Group 캐시 무효화
		itemKey := buildNetworkSecurityGroupItemKey(credential.Provider, credentialID, sg.ID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("security_group_id", sg.ID),
				domain.NewLogField("error", err))
		}
	}

	return sg, nil
}

// createGCPSecurityGroup: GCP 보안 그룹을 생성합니다

// createAzureSecurityGroup: Azure Network Security Group을 생성합니다
func (s *Service) createAzureSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Extract Azure credentials
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Resource group is required for Azure NSG
	if creds.ResourceGroup == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "resource_group is required for Azure Network Security Group", 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get Network Security Groups client
	nsgClient := clientFactory.NewSecurityGroupsClient()

	// Build Network Security Group
	nsg := armnetwork.SecurityGroup{
		Location:   to.Ptr(req.Region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			// NSG properties can be set here if needed
		},
	}

	// Add tags
	if len(req.Tags) > 0 {
		tags := make(map[string]*string)
		for k, v := range req.Tags {
			tags[k] = to.Ptr(v)
		}
		nsg.Tags = tags
	}

	// Create NSG
	poller, err := nsgClient.BeginCreateOrUpdate(ctx, creds.ResourceGroup, req.Name, nsg, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure Network Security Group")
	}

	// Wait for completion
	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure Network Security Group")
	}

	// Build response
	sgInfo := &SecurityGroupInfo{
		ID:          "",
		Name:        req.Name,
		Description: req.Description,
		VPCID:       req.VPCID, // Azure에서는 VNet ID를 VPCID로 사용
		Region:      req.Region,
		Rules:       []SecurityGroupRuleInfo{}, // Empty initially
		Tags:        req.Tags,
	}

	if result.ID != nil {
		sgInfo.ID = *result.ID
	}

	if result.Tags != nil {
		tags := make(map[string]string)
		for k, v := range result.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		sgInfo.Tags = tags
	}

	s.logger.Info(ctx, "Azure Network Security Group created successfully",
		domain.NewLogField("nsg_name", req.Name),
		domain.NewLogField("resource_group", creds.ResourceGroup),
		domain.NewLogField("location", req.Region))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/security-groups", credential.Provider),
		map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.created", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return sgInfo, nil
}

// convertGCPFirewallRulesFromRequest: 요청으로부터 GCP 방화벽 규칙을 보안 그룹 규칙 정보로 변환합니다

// UpdateSecurityGroup: 보안 그룹을 업데이트합니다
func (s *Service) UpdateSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, securityGroupID, region string) (*SecurityGroupInfo, error) {
	var sg *SecurityGroupInfo
	var err error

	switch credential.Provider {
	case "aws":
		sg, err = s.updateAWSSecurityGroup(ctx, credential, req, securityGroupID, region)
	case "gcp":
		sg, err = s.updateGCPSecurityGroup(ctx, credential, req, securityGroupID, region)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if err != nil {
		return nil, err
	}

	// 캐시 무효화: Security Group 목록 및 개별 Security Group 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		// Security Group 목록 캐시 무효화
		listKey := buildNetworkSecurityGroupListKey(credential.Provider, credentialID, region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
		// 개별 Security Group 캐시 무효화
		itemKey := buildNetworkSecurityGroupItemKey(credential.Provider, credentialID, securityGroupID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("security_group_id", securityGroupID),
				domain.NewLogField("error", err))
		}
	}

	return sg, nil
}

// updateGCPSecurityGroup: GCP 보안 그룹을 업데이트합니다

// DeleteSecurityGroup: 보안 그룹을 삭제합니다
func (s *Service) DeleteSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	var deleteErr error

	switch credential.Provider {
	case "aws":
		deleteErr = s.deleteAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		deleteErr = s.deleteGCPSecurityGroup(ctx, credential, req)
	case "azure":
		return domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}

	if deleteErr != nil {
		return deleteErr
	}

	// 캐시 무효화: Security Group 목록 및 개별 Security Group 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		// Security Group 목록 캐시 무효화
		listKey := buildNetworkSecurityGroupListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
		// 개별 Security Group 캐시 무효화
		itemKey := buildNetworkSecurityGroupItemKey(credential.Provider, credentialID, req.SecurityGroupID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate security group item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("security_group_id", req.SecurityGroupID),
				domain.NewLogField("error", err))
		}
	}

	return nil
}

// deleteGCPSecurityGroup: GCP 보안 그룹을 삭제합니다

// RemoveFirewallRule removes a specific firewall rule from a GCP firewall (security group)
func (s *Service) RemoveFirewallRule(ctx context.Context, credential *domain.Credential, req RemoveFirewallRuleRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return nil, domain.NewDomainError(domain.ErrCodeNotSupported, "AWS does not support individual rule removal - use RemoveSecurityGroupRule instead", 400)
	case "gcp":
		return s.removeGCPFirewallRule(ctx, credential, req)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// removeGCPFirewallRule: GCP 방화벽 규칙을 제거합니다

// AddFirewallRule adds a specific firewall rule to a GCP firewall (security group)
func (s *Service) AddFirewallRule(ctx context.Context, credential *domain.Credential, req AddFirewallRuleRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return nil, domain.NewDomainError(domain.ErrCodeNotSupported, "AWS does not support individual rule addition - use AddSecurityGroupRule instead", 400)
	case "gcp":
		return s.addGCPFirewallRule(ctx, credential, req)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// addGCPFirewallRule: GCP 방화벽 규칙을 추가합니다

// Helper methods

// setupGCPComputeService: GCP Compute 서비스를 설정합니다

// buildGCPFirewall: 요청으로부터 GCP 방화벽 객체를 생성합니다

// buildAllowedRules: 허용 규칙을 생성합니다

// buildDeniedRules: 거부 규칙을 생성합니다

// buildSecurityGroupInfo: 요청으로부터 보안 그룹 정보를 생성합니다

// cloneFirewall: 방화벽 객체를 복제합니다

// removePortsFromAllowed: 허용 규칙에서 포트를 제거합니다

// removePortsFromDenied: 거부 규칙에서 포트를 제거합니다

// addPortsToAllowed: 허용 규칙에 포트를 추가합니다

// addPortsToDenied: 거부 규칙에 포트를 추가합니다

// filterPorts: 포트 목록에서 특정 포트를 필터링합니다

// mergePorts: 기존 포트와 새 포트를 병합합니다

// waitForGCPOperation: GCP 작업이 완료될 때까지 대기합니다

// AddSecurityGroupRule adds a rule to a security group
func (s *Service) AddSecurityGroupRule(ctx context.Context, credential *domain.Credential, req AddSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.addAWSSecurityGroupRule(ctx, credential, req)
	case "gcp":
		// Convert AddSecurityGroupRuleRequest to AddFirewallRuleRequest
		firewallReq := AddFirewallRuleRequest{
			CredentialID:    req.CredentialID,
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Protocol:        req.Protocol,
			Ports:           []string{fmt.Sprintf("%d-%d", req.FromPort, req.ToPort)},
			SourceRanges:    req.CIDRBlocks,
		}
		return s.addGCPFirewallRule(ctx, credential, firewallReq)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// RemoveSecurityGroupRule removes a rule from a security group
func (s *Service) RemoveSecurityGroupRule(ctx context.Context, credential *domain.Credential, req RemoveSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.removeAWSSecurityGroupRule(ctx, credential, req)
	case "gcp":
		// Convert RemoveSecurityGroupRuleRequest to RemoveFirewallRuleRequest
		firewallReq := RemoveFirewallRuleRequest{
			CredentialID:    req.CredentialID,
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Protocol:        req.Protocol,
			Ports:           []string{fmt.Sprintf("%d-%d", req.FromPort, req.ToPort)},
		}
		return s.removeGCPFirewallRule(ctx, credential, firewallReq)
	case "azure":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderAzure),
			501,
		)
	case "ncp":
		return nil, domain.NewDomainError(
			domain.ErrCodeNotImplemented,
			fmt.Sprintf(ErrMsgProviderNotImplemented, domain.ProviderNCP),
			501,
		)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// UpdateSecurityGroupRules updates all rules for a security group
func (s *Service) UpdateSecurityGroupRules(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRulesRequest) (*SecurityGroupInfo, error) {

	// Get current security group
	getReq := GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	currentSG, err := s.GetSecurityGroup(ctx, credential, getReq)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "get current security group")
	}

	// Remove all existing rules
	for _, rule := range currentSG.Rules {
		removeReq := RemoveSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            rule.Type,
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.RemoveSecurityGroupRule(ctx, credential, removeReq)
		if err != nil {
			s.logger.Warn(ctx, "Failed to remove existing rule", domain.NewLogField("error", err))
		}
	}

	// Add new ingress rules
	for _, rule := range req.IngressRules {
		addReq := AddSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            "ingress",
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.AddSecurityGroupRule(ctx, credential, addReq)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAWSError(err, "add ingress rule")
		}
	}

	// Add new egress rules
	for _, rule := range req.EgressRules {
		addReq := AddSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            "egress",
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.AddSecurityGroupRule(ctx, credential, addReq)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAWSError(err, "add egress rule")
		}
	}

	// Get updated security group info
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// GCP Subnet Functions

// listGCPSubnets: GCP 서브넷 목록을 조회합니다

// getGCPSubnet: 특정 GCP 서브넷을 조회합니다

// createGCPSubnet: 새로운 GCP 서브넷을 생성합니다

// updateGCPSubnet: GCP 서브넷을 업데이트합니다

// deleteGCPSubnet: GCP 서브넷을 삭제합니다

// extractSubnetNameFromSubnetID: GCP 서브넷 ID에서 서브넷 이름을 추출합니다

// extractCleanSubnetID: GCP URL에서 서브넷 ID를 추출합니다

// extractCleanVPCID: GCP URL에서 VPC ID를 추출합니다

// extractCleanRegionID: GCP URL에서 리전 ID를 추출합니다

// extractProjectID extracts project ID from GCP URL

// extractRegionName: GCP URL에서 리전 이름을 추출합니다

// getFirewallRulesCount: 특정 네트워크의 방화벽 규칙 개수를 조회합니다

// getGatewayInfo: 특정 네트워크의 게이트웨이 정보를 조회합니다

// listRouters: 모든 리전의 라우터 목록을 조회합니다

// findGatewayForNetwork: 특정 네트워크의 게이트웨이 정보를 찾습니다

// isRouterConnectedToNetwork: 라우터가 특정 네트워크에 연결되어 있는지 확인합니다

// checkRouterForGateway: 라우터에 NAT 또는 인터넷 게이트웨이가 있는지 확인합니다

// checkForNATGateway: 라우터에 NAT 게이트웨이가 있는지 확인합니다

// checkForInternetGateway: 라우터에 인터넷 게이트웨이가 있는지 확인합니다

// AWS Subnet Functions (existing implementations)

// Stub implementations for Azure and NCP

// listAzureSubnets lists Azure subnets (stub)
func (s *Service) listAzureSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group and VNet name from VPC ID
	resourceGroup := ""
	vnetName := ""
	parts := strings.Split(req.VPCID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" {
		if creds.ResourceGroup != "" {
			resourceGroup = creds.ResourceGroup
			// Try to extract VNet name from VPCID or use it as-is
			if strings.Contains(req.VPCID, "/") {
				// It's a full resource ID, extract name
				for i, part := range parts {
					if part == "virtualNetworks" && i+1 < len(parts) {
						vnetName = parts[i+1]
						break
					}
				}
			} else {
				vnetName = req.VPCID
			}
		} else {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, ErrMsgInvalidVPCIDOrResourceGroup, 400)
		}
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get subnets client
	subnetsClient := clientFactory.NewSubnetsClient()

	// List subnets
	var subnets []SubnetInfo
	pager := subnetsClient.NewListPager(resourceGroup, vnetName, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "list Azure subnets")
		}

		for _, subnet := range page.Value {
			subnetInfo := SubnetInfo{
				ID:     "",
				Name:   "",
				VPCID:  req.VPCID,
				Region: req.Region,
				State:  "Succeeded",
			}

			if subnet.ID != nil {
				subnetInfo.ID = *subnet.ID
			}
			if subnet.Name != nil {
				subnetInfo.Name = *subnet.Name
			}

			if subnet.Properties != nil {
				if subnet.Properties.AddressPrefix != nil {
					subnetInfo.CIDRBlock = *subnet.Properties.AddressPrefix
				}
				if subnet.Properties.ProvisioningState != nil {
					subnetInfo.State = string(*subnet.Properties.ProvisioningState)
				}
			}

			subnets = append(subnets, subnetInfo)
		}
	}

	return &ListSubnetsResponse{Subnets: subnets}, nil
}

// getAzureSubnet: Azure 서브넷 상세 정보를 조회합니다
func (s *Service) getAzureSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group, VNet name, and subnet name from subnet ID
	resourceGroup := ""
	vnetName := ""
	subnetName := ""
	parts := strings.Split(req.SubnetID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
		if part == "subnets" && i+1 < len(parts) {
			subnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" || subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid subnet ID format", 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get subnets client
	subnetsClient := clientFactory.NewSubnetsClient()

	// Get subnet details
	subnet, err := subnetsClient.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get Azure subnet")
	}

	// Extract VPC ID from subnet ID
	// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/{subnet}
	vpcID := ""
	for i, part := range parts {
		if part == "subnets" && i > 0 {
			// Reconstruct VPC ID up to virtualNetworks/{vnet}
			vpcID = strings.Join(parts[:i], "/")
			break
		}
	}
	if vpcID == "" {
		// Fallback: construct VPC ID from resource group and VNet name
		vpcID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", creds.SubscriptionID, resourceGroup, vnetName)
	}

	subnetInfo := SubnetInfo{
		ID:     "",
		Name:   subnetName,
		VPCID:  vpcID,
		Region: req.Region,
		State:  "Succeeded",
	}

	if subnet.Subnet.ID != nil {
		subnetInfo.ID = *subnet.Subnet.ID
	}

	if subnet.Subnet.Properties != nil {
		if subnet.Subnet.Properties.AddressPrefix != nil {
			subnetInfo.CIDRBlock = *subnet.Subnet.Properties.AddressPrefix
		}
		if subnet.Subnet.Properties.ProvisioningState != nil {
			subnetInfo.State = string(*subnet.Subnet.Properties.ProvisioningState)
		}
	}

	return &subnetInfo, nil
}

// createAzureSubnet: Azure 서브넷을 생성합니다
func (s *Service) createAzureSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group and VNet name from VPC ID
	resourceGroup := ""
	vnetName := ""
	parts := strings.Split(req.VPCID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" {
		if creds.ResourceGroup != "" {
			resourceGroup = creds.ResourceGroup
			if strings.Contains(req.VPCID, "/") {
				for i, part := range parts {
					if part == "virtualNetworks" && i+1 < len(parts) {
						vnetName = parts[i+1]
						break
					}
				}
			} else {
				vnetName = req.VPCID
			}
		} else {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, ErrMsgInvalidVPCIDOrResourceGroup, 400)
		}
	}

	// CIDR block is required
	if req.CIDRBlock == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "cidr_block is required for Azure subnet", 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get subnets client
	subnetsClient := clientFactory.NewSubnetsClient()

	// Build subnet
	subnet := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr(req.CIDRBlock),
		},
	}

	// Create subnet
	poller, err := subnetsClient.BeginCreateOrUpdate(ctx, resourceGroup, vnetName, req.Name, subnet, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure subnet")
	}

	// Wait for completion
	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create Azure subnet")
	}

	subnetInfo := SubnetInfo{
		ID:        "",
		Name:      req.Name,
		VPCID:     req.VPCID,
		CIDRBlock: req.CIDRBlock,
		Region:    req.Region,
		State:     "Succeeded",
	}

	if result.Subnet.ID != nil {
		subnetInfo.ID = *result.Subnet.ID
	}

	if result.Subnet.Properties != nil && result.Subnet.Properties.ProvisioningState != nil {
		subnetInfo.State = string(*result.Subnet.Properties.ProvisioningState)
	}

	s.logger.Info(ctx, "Azure subnet creation completed",
		domain.NewLogField("subnet_name", req.Name),
		domain.NewLogField("resource_group", resourceGroup),
		domain.NewLogField("vnet_name", vnetName))

	// Construct VPC ID for cache invalidation
	vpcID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", creds.SubscriptionID, resourceGroup, vnetName)

	// 캐시 무효화: 서브넷 목록 캐시 삭제 (메서드가 없으면 로그만 남김)
	credentialID := credential.ID.String()
	s.logger.Debug(ctx, "Subnet list cache invalidation skipped (method not available)",
		domain.NewLogField("provider", credential.Provider),
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("vpc_id", vpcID))

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs/%s/subnets", credential.Provider, req.VPCID),
		map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	// 이벤트 발행
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"cidr_block":    subnetInfo.CIDRBlock,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.created", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return &subnetInfo, nil
}

// updateAzureSubnet: Azure 서브넷을 업데이트합니다
func (s *Service) updateAzureSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract resource group, VNet name, and subnet name from subnet ID
	resourceGroup := ""
	vnetName := ""
	subnetName := ""
	parts := strings.Split(subnetID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
		if part == "subnets" && i+1 < len(parts) {
			subnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" || subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid subnet ID format", 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get subnets client
	subnetsClient := clientFactory.NewSubnetsClient()

	// Get existing subnet
	existingSubnet, err := subnetsClient.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get Azure subnet")
	}

	// Extract VPC ID from subnet ID
	// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/{subnet}
	vpcID := ""
	for i, part := range parts {
		if part == "subnets" && i > 0 {
			// Reconstruct VPC ID up to virtualNetworks/{vnet}
			vpcID = strings.Join(parts[:i], "/")
			break
		}
	}
	if vpcID == "" {
		// Fallback: construct VPC ID from resource group and VNet name
		vpcID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", creds.SubscriptionID, resourceGroup, vnetName)
	}

	// Update subnet (Azure subnets are mostly immutable, but we can update service endpoints, etc.)
	// For now, we'll just return the existing subnet as Azure doesn't support many subnet updates
	subnetInfo := SubnetInfo{
		ID:     subnetID,
		Name:   subnetName,
		VPCID:  vpcID,
		Region: region,
		State:  "Succeeded",
	}

	if existingSubnet.Subnet.Properties != nil {
		if existingSubnet.Subnet.Properties.AddressPrefix != nil {
			subnetInfo.CIDRBlock = *existingSubnet.Subnet.Properties.AddressPrefix
		}
		if existingSubnet.Subnet.Properties.ProvisioningState != nil {
			subnetInfo.State = string(*existingSubnet.Subnet.Properties.ProvisioningState)
		}
	}

	return &subnetInfo, nil
}

// deleteAzureSubnet: Azure 서브넷을 삭제합니다
func (s *Service) deleteAzureSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return err
	}

	// Extract resource group, VNet name, and subnet name from subnet ID
	resourceGroup := ""
	vnetName := ""
	subnetName := ""
	parts := strings.Split(req.SubnetID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if part == "virtualNetworks" && i+1 < len(parts) {
			vnetName = parts[i+1]
		}
		if part == "subnets" && i+1 < len(parts) {
			subnetName = parts[i+1]
		}
	}

	if resourceGroup == "" || vnetName == "" || subnetName == "" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid subnet ID format", 400)
	}

	// Create Azure Network client
	clientFactory, err := s.createAzureNetworkClient(ctx, creds)
	if err != nil {
		return err
	}

	// Get subnets client
	subnetsClient := clientFactory.NewSubnetsClient()

	// Delete subnet
	poller, err := subnetsClient.BeginDelete(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete Azure subnet")
	}

	// Wait for completion
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete Azure subnet")
	}

	s.logger.Info(ctx, "Azure subnet deletion completed",
		domain.NewLogField("subnet_name", subnetName),
		domain.NewLogField("resource_group", resourceGroup),
		domain.NewLogField("vnet_name", vnetName))

	// Extract VPC ID from subnet ID for cache invalidation
	vpcID := strings.Join(parts[:strings.Index(strings.Join(parts, "/"), "/subnets")], "/")
	if !strings.Contains(vpcID, "/virtualNetworks/") {
		vpcID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", creds.SubscriptionID, resourceGroup, vnetName)
	}

	// 캐시 무효화: 서브넷 목록 및 개별 서브넷 캐시 삭제 (메서드가 없으면 로그만 남김)
	credentialID := credential.ID.String()
	s.logger.Debug(ctx, "Subnet cache invalidation skipped (method not available)",
		domain.NewLogField("provider", credential.Provider),
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("vpc_id", vpcID),
		domain.NewLogField("subnet_id", req.SubnetID))

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/subnets/%s", credential.Provider, req.SubnetID),
		map[string]interface{}{
			"subnet_id":      req.SubnetID,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
			"resource_group": resourceGroup,
		},
	)

	// 이벤트 발행
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.deleted", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return nil
}

// listNCPSubnets lists NCP subnets (stub)
func (s *Service) listNCPSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	s.logger.Info(ctx, "NCP subnet listing not yet implemented")
	return &ListSubnetsResponse{Subnets: []SubnetInfo{}}, nil
}

// getNCPSubnet gets NCP subnet (stub)
func (s *Service) getNCPSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info(ctx, "NCP subnet retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet retrieval not yet implemented", 501)
}

// createNCPSubnet creates NCP subnet (stub)
func (s *Service) createNCPSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info(ctx, "NCP subnet creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet creation not yet implemented", 501)
}

// updateNCPSubnet updates NCP subnet (stub)
func (s *Service) updateNCPSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	s.logger.Info(ctx, "NCP subnet update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet update not yet implemented", 501)
}

// deleteNCPSubnet deletes NCP subnet (stub)
func (s *Service) deleteNCPSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	s.logger.Info(ctx, "NCP subnet deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet deletion not yet implemented", 501)
}

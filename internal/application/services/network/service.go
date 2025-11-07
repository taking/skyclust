package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// Service: 네트워크 리소스 작업을 처리하는 서비스
type Service struct {
	credentialService domain.CredentialService
	cache             cache.Cache
	keyBuilder        *cache.CacheKeyBuilder
	invalidator       *cache.Invalidator
	eventPublisher    *messaging.Publisher
	auditLogRepo      domain.AuditLogRepository
	logger            *zap.Logger
}

// NewService: 새로운 네트워크 서비스를 생성합니다
func NewService(credentialService domain.CredentialService, cacheService cache.Cache, eventBus messaging.Bus, auditLogRepo domain.AuditLogRepository, logger *zap.Logger) *Service {
	eventPublisher := messaging.NewPublisher(eventBus, logger)
	return &Service{
		credentialService: credentialService,
		cache:             cacheService,
		keyBuilder:        cache.NewCacheKeyBuilder(),
		invalidator:       cache.NewInvalidatorWithEvents(cacheService, eventPublisher),
		eventPublisher:    eventPublisher,
		auditLogRepo:      auditLogRepo,
		logger:            logger,
	}
}

// ListVPCs: 주어진 자격증명과 리전에 대한 VPC 목록을 조회합니다
func (s *Service) ListVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := s.keyBuilder.BuildNetworkVPCListKey(credential.Provider, credentialID, req.Region)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedResponse ListVPCsResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
			s.logger.Debug("VPCs retrieved from cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", req.Region))
			return &cachedResponse, nil
		}
	}

	// 캐시 미스 시 실제 API 호출
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

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cache != nil && response != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceNetwork)
		if err := s.cache.Set(ctx, cacheKey, response, ttl); err != nil {
			s.logger.Warn("Failed to cache VPCs, continuing without cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", req.Region),
				zap.Error(err))
			// 캐시 실패는 치명적이지 않으므로 계속 진행
		}
	}

	return response, nil
}

// listAWSVPCs: AWS VPC 목록을 조회합니다
func (s *Service) listAWSVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe VPCs
	input := &ec2.DescribeVpcsInput{}
	if req.VPCID != "" {
		input.VpcIds = []string{req.VPCID}
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe VPCs: %v", err), 502)
	}

	// Convert to DTOs
	vpcs := make([]VPCInfo, 0, len(result.Vpcs))
	for _, vpc := range result.Vpcs {
		vpcInfo := VPCInfo{
			ID:        aws.ToString(vpc.VpcId),
			Name:      s.getTagValue(vpc.Tags, "Name"),
			State:     string(vpc.State),
			IsDefault: aws.ToBool(vpc.IsDefault),
			Region:    req.Region,
		}
		vpcs = append(vpcs, vpcInfo)
	}

	return &ListVPCsResponse{VPCs: vpcs}, nil
}

// GetVPC: ID로 특정 VPC를 조회합니다
func (s *Service) GetVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := s.keyBuilder.BuildNetworkVPCItemKey(credential.Provider, credentialID, req.VPCID)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedVPC VPCInfo
		if err := s.cache.Get(ctx, cacheKey, &cachedVPC); err == nil {
			s.logger.Debug("VPC retrieved from cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("vpc_id", req.VPCID))
			return &cachedVPC, nil
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
	if s.cache != nil && vpc != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceNetwork)
		if err := s.cache.Set(ctx, cacheKey, vpc, ttl); err != nil {
			s.logger.Warn("Failed to cache VPC",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("vpc_id", req.VPCID),
				zap.Error(err))
		}
	}

	return vpc, nil
}

// getAWSVPC: 특정 AWS VPC를 조회합니다
func (s *Service) getAWSVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe specific VPC
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{req.VPCID},
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe VPC: %v", err), 502)
	}

	if len(result.Vpcs) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("VPC not found: %s", req.VPCID), 404)
	}

	vpc := result.Vpcs[0]
	vpcInfo := &VPCInfo{
		ID:        aws.ToString(vpc.VpcId),
		Name:      s.getTagValue(vpc.Tags, "Name"),
		State:     string(vpc.State),
		IsDefault: aws.ToBool(vpc.IsDefault),
		Region:    req.Region,
		Tags:      s.convertTags(vpc.Tags),
	}

	return vpcInfo, nil
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
	if err := s.invalidator.InvalidateNetworkVPCList(ctx, credential.Provider, credentialID, req.Region); err != nil {
		s.logger.Warn("Failed to invalidate VPC list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", req.Region),
			zap.Error(err))
	}

	// 이벤트 발행: VPC 생성 이벤트
	vpcData := map[string]interface{}{
		"vpc_id": vpc.ID,
		"name":   vpc.Name,
		"state":  vpc.State,
		"region": vpc.Region,
	}
	if err := s.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, req.Region, "created", vpcData); err != nil {
		s.logger.Warn("Failed to publish VPC created event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("vpc_id", vpc.ID),
			zap.Error(err))
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
func (s *Service) createGCPVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Marshal credential data for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Extract project ID from credential data
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Convert to GCP-specific request
	// GCP SDK 제한으로 인해 auto_create_subnets는 항상 true로 강제 설정
	autoCreateSubnets := true // GCP SDK 제한으로 인한 강제 설정

	routingMode := "REGIONAL" // Default value
	if req.RoutingMode != "" {
		routingMode = req.RoutingMode
	}

	mtu := int64(1460) // Default value
	if req.MTU > 0 {
		mtu = req.MTU
	}

	gcpReq := CreateGCPVPCRequest{
		CredentialID:      req.CredentialID,
		Name:              req.Name,
		Description:       req.Description,
		Region:            req.Region, // Optional for VPC (Global resource)
		ProjectID:         projectID,
		AutoCreateSubnets: autoCreateSubnets, // GCP SDK 제한으로 인한 강제 설정
		RoutingMode:       routingMode,       // Use user's preference
		MTU:               mtu,               // Use user's preference
		Tags:              req.Tags,
	}

	return s.createGCPVPCWithAdvanced(ctx, credential, gcpReq, computeService)
}

// createGCPVPCWithAdvanced: 고급 설정으로 GCP VPC를 생성합니다
func (s *Service) createGCPVPCWithAdvanced(ctx context.Context, credential *domain.Credential, req CreateGCPVPCRequest, computeService *compute.Service) (*VPCInfo, error) {
	network := s.buildGCPNetworkObject(req)

	s.logNetworkConfiguration(network)

	operation, err := s.createGCPNetworkOperation(ctx, computeService, req.ProjectID, network)
	if err != nil {
		return nil, err
	}

	s.logOperationInitiated(req, operation)

	return s.buildVPCInfoFromRequest(req), nil
}

// buildGCPNetworkObject: GCP 네트워크 객체를 생성합니다
func (s *Service) buildGCPNetworkObject(req CreateGCPVPCRequest) *compute.Network {
	return &compute.Network{
		Name:                  req.Name,
		Description:           req.Description,
		AutoCreateSubnetworks: req.AutoCreateSubnets,
		RoutingConfig: &compute.NetworkRoutingConfig{
			RoutingMode: req.RoutingMode,
		},
		Mtu: req.MTU,
		// IPv4Range field is intentionally omitted to ensure subnet mode
	}
}

// logNetworkConfiguration: 네트워크 구성을 로깅합니다
func (s *Service) logNetworkConfiguration(network *compute.Network) {
	routingMode := "REGIONAL"
	if network.RoutingConfig != nil {
		routingMode = network.RoutingConfig.RoutingMode
	}
	s.logger.Info("Creating GCP network with configuration",
		zap.String("name", network.Name),
		zap.String("description", network.Description),
		zap.Bool("auto_create_subnetworks", network.AutoCreateSubnetworks),
		zap.String("routing_mode", routingMode),
		zap.Int64("mtu", network.Mtu))
}

// createGCPNetworkOperation: GCP 네트워크 생성 작업을 시작합니다
func (s *Service) createGCPNetworkOperation(ctx context.Context, computeService *compute.Service, projectID string, network *compute.Network) (*compute.Operation, error) {
	operation, err := computeService.Networks.Insert(projectID, network).Context(ctx).Do()
	if err != nil {
		s.logger.Error("Failed to create GCP network",
			zap.String("error", err.Error()),
			zap.String("network_name", network.Name),
			zap.Bool("auto_create_subnetworks", network.AutoCreateSubnetworks))
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP network: %v", err), 502)
	}
	return operation, nil
}

// logOperationInitiated: 작업 시작을 로깅합니다
func (s *Service) logOperationInitiated(req CreateGCPVPCRequest, operation *compute.Operation) {
	s.logger.Info("GCP VPC creation initiated",
		zap.String("vpc_name", req.Name),
		zap.String("project_id", req.ProjectID),
		zap.String("operation_id", operation.Name))
}

// buildVPCInfoFromRequest: 요청으로부터 VPC 정보를 생성합니다
func (s *Service) buildVPCInfoFromRequest(req CreateGCPVPCRequest) *VPCInfo {
	return &VPCInfo{
		ID:          fmt.Sprintf(ResourcePrefixVPC, req.ProjectID, req.Name),
		Name:        req.Name,
		State:       StateCreating,
		NetworkMode: NetworkModeSubnet,
		RoutingMode: req.RoutingMode,
		MTU:         req.MTU,
		AutoSubnets: req.AutoCreateSubnets,
		Description: req.Description,
		Tags:        req.Tags,
	}
}

// listGCPVPCs: GCP VPC 목록을 조회합니다
func (s *Service) listGCPVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// List all networks
	networks, err := computeService.Networks.List(projectID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list GCP networks: %v", err), 502)
	}

	// Convert to VPCInfo
	vpcs := make([]VPCInfo, 0, len(networks.Items))
	for _, networkItem := range networks.Items {
		// Determine network mode
		networkMode := "subnet"
		if networkItem.IPv4Range != "" {
			networkMode = "legacy"
		}

		// Get routing mode
		routingMode := "REGIONAL"
		if networkItem.RoutingConfig != nil {
			routingMode = networkItem.RoutingConfig.RoutingMode
		}

		// Get MTU
		mtu := int64(1460) // Default MTU
		if networkItem.Mtu > 0 {
			mtu = networkItem.Mtu
		}

		// Extract clean format IDs
		vpcIDClean := s.extractCleanVPCID(networkItem.SelfLink)

		// Get firewall rules count for this network
		firewallCount, err := s.getFirewallRulesCount(ctx, computeService, projectID, networkItem.Name)
		if err != nil {
			s.logger.Warn("Failed to get firewall rules count",
				zap.String("network_name", networkItem.Name),
				zap.Error(err))
			firewallCount = 0
		}

		// Get gateway information for this network
		gatewayInfo, err := s.getGatewayInfo(ctx, computeService, projectID, networkItem.Name)
		if err != nil {
			s.logger.Warn("Failed to get gateway info",
				zap.String("network_name", networkItem.Name),
				zap.Error(err))
			gatewayInfo = nil
		}

		vpcInfo := VPCInfo{
			ID:                vpcIDClean, // Clean format: projects/{project}/global/networks/{name}
			Name:              networkItem.Name,
			State:             "available",
			IsDefault:         networkItem.Name == "default",
			NetworkMode:       networkMode,
			RoutingMode:       routingMode,
			MTU:               mtu,
			AutoSubnets:       networkItem.AutoCreateSubnetworks,
			Description:       networkItem.Description,
			FirewallRuleCount: firewallCount,
			Gateway:           gatewayInfo,
			CreationTimestamp: networkItem.CreationTimestamp,
			Tags:              map[string]string{}, // User-defined tags would be populated here
		}
		vpcs = append(vpcs, vpcInfo)
	}

	s.logger.Info("GCP VPCs listed successfully",
		zap.String("project_id", projectID),
		zap.Int("count", len(vpcs)))

	return &ListVPCsResponse{VPCs: vpcs}, nil
}

// getGCPVPC: 특정 GCP VPC를 조회합니다
func (s *Service) getGCPVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract network name from VPC ID
	networkName := s.extractNetworkNameFromVPCID(req.VPCID)
	if networkName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid VPC ID format: %s", req.VPCID), 400)
	}

	// Get specific network
	network, err := computeService.Networks.Get(projectID, networkName).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get GCP network: %v", err), 502)
	}

	// Determine network mode
	networkMode := "subnet"
	if network.IPv4Range != "" {
		networkMode = "legacy"
	}

	// Get routing mode
	routingMode := "REGIONAL"
	if network.RoutingConfig != nil {
		routingMode = network.RoutingConfig.RoutingMode
	}

	// Get MTU
	mtu := int64(1460) // Default MTU
	if network.Mtu > 0 {
		mtu = network.Mtu
	}

	// Extract clean format IDs
	vpcIDClean := s.extractCleanVPCID(network.SelfLink)

	// Get firewall rules count for this network
	firewallCount, err := s.getFirewallRulesCount(ctx, computeService, projectID, network.Name)
	if err != nil {
		s.logger.Warn("Failed to get firewall rules count",
			zap.String("network_name", network.Name),
			zap.Error(err))
		firewallCount = 0
	}

	// Get gateway information for this network
	gatewayInfo, err := s.getGatewayInfo(ctx, computeService, projectID, network.Name)
	if err != nil {
		s.logger.Warn("Failed to get gateway info",
			zap.String("network_name", network.Name),
			zap.Error(err))
		gatewayInfo = nil
	}

	vpcInfo := &VPCInfo{
		ID:                vpcIDClean, // Clean format: projects/{project}/global/networks/{name}
		Name:              network.Name,
		State:             "available",
		IsDefault:         network.Name == "default",
		NetworkMode:       networkMode,
		RoutingMode:       routingMode,
		MTU:               mtu,
		AutoSubnets:       network.AutoCreateSubnetworks,
		Description:       network.Description,
		FirewallRuleCount: firewallCount,
		Gateway:           gatewayInfo,
		CreationTimestamp: network.CreationTimestamp,
		Tags:              map[string]string{}, // User-defined tags would be populated here
	}

	s.logger.Info("GCP VPC retrieved successfully",
		zap.String("vpc_name", network.Name),
		zap.String("project_id", projectID))

	return vpcInfo, nil
}

// extractNetworkNameFromVPCID: VPC ID에서 네트워크 이름을 추출합니다
func (s *Service) extractNetworkNameFromVPCID(vpcID string) string {
	// Support two formats:
	// 1. Full format: projects/{project}/global/networks/{network_name}
	// 2. Simple format: {network_name}

	// Check if it's a full format
	parts := strings.Split(vpcID, "/")
	if len(parts) >= 4 && parts[len(parts)-2] == "networks" {
		return parts[len(parts)-1]
	}

	// If it's a simple format, return as is
	return vpcID
}

// deleteGCPVPC: GCP VPC를 삭제합니다
func (s *Service) deleteGCPVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract network name from VPC ID
	networkName := s.extractNetworkNameFromVPCID(req.VPCID)
	if networkName == "" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid VPC ID format: %s", req.VPCID), 400)
	}

	// Check if VPC exists
	_, err = computeService.Networks.Get(projectID, networkName).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get VPC: %v", err), 502)
	}

	// Check and clean up dependencies
	s.logger.Info("Starting VPC dependency cleanup",
		zap.String("vpc_name", networkName),
		zap.String("project_id", projectID))

	err = s.cleanupVPCResources(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn("Failed to clean up VPC resources, proceeding with deletion",
			zap.String("vpc_name", networkName),
			zap.Error(err))
		// Continue with deletion - GCP will handle validation
	} else {
		s.logger.Info("VPC dependency cleanup completed successfully",
			zap.String("vpc_name", networkName))
	}

	// Delete the network
	s.logger.Info("Initiating VPC deletion",
		zap.String("vpc_name", networkName),
		zap.String("project_id", projectID))

	operation, err := computeService.Networks.Delete(projectID, networkName).Context(ctx).Do()
	if err != nil {
		s.logger.Error("Failed to delete GCP network",
			zap.String("vpc_name", networkName),
			zap.String("project_id", projectID),
			zap.Error(err))
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete GCP network: %v", err), 502)
	}

	s.logger.Info("GCP VPC deletion initiated successfully",
		zap.String("vpc_name", networkName),
		zap.String("project_id", projectID),
		zap.String("operation_id", operation.Name),
		zap.String("operation_status", operation.Status))

	return nil
}

// cleanupVPCResources: VPC 리소스를 정리합니다
func (s *Service) cleanupVPCResources(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info("Starting VPC resource cleanup",
		zap.String("vpc_name", networkName),
		zap.String("project_id", projectID))

	// 1. Delete firewall rules associated with this network
	err := s.deleteNetworkFirewallRules(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn("Failed to delete firewall rules",
			zap.String("vpc_name", networkName),
			zap.Error(err))
	}

	// 2. Delete subnets in this network
	err = s.deleteNetworkSubnets(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn("Failed to delete subnets",
			zap.String("vpc_name", networkName),
			zap.Error(err))
	}

	// 3. Check for instances using this network
	err = s.checkNetworkInstances(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn("Found instances using this network",
			zap.String("vpc_name", networkName),
			zap.Error(err))
		return domain.NewDomainError(domain.ErrCodeConflict, "cannot delete VPC: instances are still using this network", 409)
	}

	s.logger.Info("VPC resource cleanup completed",
		zap.String("vpc_name", networkName))

	return nil
}

// deleteNetworkFirewallRules: 네트워크 방화벽 규칙을 삭제합니다
func (s *Service) deleteNetworkFirewallRules(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info("Listing firewall rules for cleanup",
		zap.String("network", networkName),
		zap.String("project_id", projectID))

	// List all firewall rules
	firewalls, err := computeService.Firewalls.List(projectID).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list firewall rules: %v", err), 502)
	}

	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
	deletedCount := 0

	for _, firewall := range firewalls.Items {
		if firewall.Network == networkURL {
			s.logger.Info("Deleting firewall rule",
				zap.String("firewall_name", firewall.Name),
				zap.String("network", networkName))

			_, err := computeService.Firewalls.Delete(projectID, firewall.Name).Context(ctx).Do()
			if err != nil {
				s.logger.Warn("Failed to delete firewall rule",
					zap.String("firewall_name", firewall.Name),
					zap.Error(err))
				// Continue with other firewall rules
			} else {
				deletedCount++
				s.logger.Info("Firewall rule deleted successfully",
					zap.String("firewall_name", firewall.Name))
			}
		}
	}

	s.logger.Info("Firewall rules cleanup completed",
		zap.String("network", networkName),
		zap.Int("deleted_count", deletedCount))

	return nil
}

// deleteNetworkSubnets: 네트워크 서브넷을 삭제합니다
func (s *Service) deleteNetworkSubnets(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info("Listing subnets for cleanup",
		zap.String("network", networkName),
		zap.String("project_id", projectID))

	// List all subnets
	subnets, err := computeService.Subnetworks.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list subnets: %v", err), 502)
	}

	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
	deletedCount := 0

	for _, subnetList := range subnets.Items {
		for _, subnet := range subnetList.Subnetworks {
			if subnet.Network == networkURL {
				s.logger.Info("Deleting subnet",
					zap.String("subnet_name", subnet.Name),
					zap.String("region", subnet.Region),
					zap.String("network", networkName))

				_, err := computeService.Subnetworks.Delete(projectID, subnet.Region, subnet.Name).Context(ctx).Do()
				if err != nil {
					s.logger.Warn("Failed to delete subnet",
						zap.String("subnet_name", subnet.Name),
						zap.String("region", subnet.Region),
						zap.Error(err))
					// Continue with other subnets
				} else {
					deletedCount++
					s.logger.Info("Subnet deleted successfully",
						zap.String("subnet_name", subnet.Name),
						zap.String("region", subnet.Region))
				}
			}
		}
	}

	s.logger.Info("Subnets cleanup completed",
		zap.String("network", networkName),
		zap.Int("deleted_count", deletedCount))

	return nil
}

// checkNetworkInstances: 네트워크에 연결된 인스턴스를 확인합니다
func (s *Service) checkNetworkInstances(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	// List all instances
	instances, err := computeService.Instances.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list instances: %v", err), 502)
	}

	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)

	for _, instanceList := range instances.Items {
		for _, instance := range instanceList.Instances {
			for _, networkInterface := range instance.NetworkInterfaces {
				if networkInterface.Network == networkURL {
					return domain.NewDomainError(domain.ErrCodeConflict, fmt.Sprintf("instance %s is using this network", instance.Name), 409)
				}
			}
		}
	}

	return nil
}

// checkVPCDeletionDependencies checks if VPC can be safely deleted
// Stub implementations for Azure, NCP, and GCP update functions

// listAzureVPCs: Azure VPC 목록을 조회합니다
func (s *Service) listAzureVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	s.logger.Info("Azure VPC listing not yet implemented")
	return &ListVPCsResponse{VPCs: []VPCInfo{}}, nil
}

// listNCPVPCs: NCP VPC 목록을 조회합니다
func (s *Service) listNCPVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	s.logger.Info("NCP VPC listing not yet implemented")
	return &ListVPCsResponse{VPCs: []VPCInfo{}}, nil
}

// getAzureVPC: 특정 Azure VPC를 조회합니다
func (s *Service) getAzureVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	s.logger.Info("Azure VPC retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure VPC retrieval not yet implemented", 501)
}

// getNCPVPC: 특정 NCP VPC를 조회합니다
func (s *Service) getNCPVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	s.logger.Info("NCP VPC retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC retrieval not yet implemented", 501)
}

// createAzureVPC: Azure VPC를 생성합니다
func (s *Service) createAzureVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	s.logger.Info("Azure VPC creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure VPC creation not yet implemented", 501)
}

// createNCPVPC: NCP VPC를 생성합니다
func (s *Service) createNCPVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	s.logger.Info("NCP VPC creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC creation not yet implemented", 501)
}

// updateGCPVPC: GCP VPC를 업데이트합니다
func (s *Service) updateGCPVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	s.logger.Info("GCP VPC update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "GCP VPC update not yet implemented", 501)
}

// updateAzureVPC: Azure VPC를 업데이트합니다
func (s *Service) updateAzureVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	s.logger.Info("Azure VPC update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure VPC update not yet implemented", 501)
}

// updateNCPVPC: NCP VPC를 업데이트합니다
func (s *Service) updateNCPVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	s.logger.Info("NCP VPC update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC update not yet implemented", 501)
}

// deleteAzureVPC: Azure VPC를 삭제합니다
func (s *Service) deleteAzureVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	s.logger.Info("Azure VPC deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure VPC deletion not yet implemented", 501)
}

// deleteNCPVPC: NCP VPC를 삭제합니다
func (s *Service) deleteNCPVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	s.logger.Info("NCP VPC deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP VPC deletion not yet implemented", 501)
}

// createAWSVPC: AWS VPC를 생성합니다
func (s *Service) createAWSVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Create VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String(req.CIDRBlock),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeVpc,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateVpc(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create VPC: %v", err), 502)
	}

	vpcInfo := &VPCInfo{
		ID:        aws.ToString(result.Vpc.VpcId),
		Name:      req.Name,
		State:     string(result.Vpc.State),
		IsDefault: aws.ToBool(result.Vpc.IsDefault),
		Region:    req.Region,
		Tags:      req.Tags,
	}

	return vpcInfo, nil
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

// updateAWSVPC: AWS VPC를 업데이트합니다
func (s *Service) updateAWSVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Update VPC tags if provided
	if req.Name != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{vpcID},
			Tags:      tags,
		})
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update VPC tags: %v", err), 502)
		}
	}

	// Get updated VPC info
	getReq := GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}
	vpc, err := s.GetVPC(ctx, credential, getReq)
	if err != nil {
		return nil, err
	}

	// 캐시 무효화: VPC 목록 및 개별 VPC 캐시 삭제
	credentialID := credential.ID.String()
	if err := s.invalidator.InvalidateNetworkVPCList(ctx, credential.Provider, credentialID, region); err != nil {
		s.logger.Warn("Failed to invalidate VPC list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", region),
			zap.Error(err))
	}
	if err := s.invalidator.InvalidateNetworkVPCItem(ctx, credential.Provider, credentialID, vpcID); err != nil {
		s.logger.Warn("Failed to invalidate VPC item cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("vpc_id", vpcID),
			zap.Error(err))
	}

	// 이벤트 발행: VPC 업데이트 이벤트
	vpcData := map[string]interface{}{
		"vpc_id": vpc.ID,
		"name":   vpc.Name,
		"state":  vpc.State,
		"region": vpc.Region,
	}
	if err := s.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, region, "updated", vpcData); err != nil {
		s.logger.Warn("Failed to publish VPC updated event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("vpc_id", vpcID),
			zap.Error(err))
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/vpcs/%s", credential.Provider, vpcID),
		map[string]interface{}{
			"vpc_id":        vpc.ID,
			"name":          vpc.Name,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	return vpc, nil
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
	if err := s.invalidator.InvalidateNetworkVPCList(ctx, credential.Provider, credentialID, req.Region); err != nil {
		s.logger.Warn("Failed to invalidate VPC list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", req.Region),
			zap.Error(err))
	}
	if err := s.invalidator.InvalidateNetworkVPCItem(ctx, credential.Provider, credentialID, req.VPCID); err != nil {
		s.logger.Warn("Failed to invalidate VPC item cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("vpc_id", req.VPCID),
			zap.Error(err))
	}

	// 이벤트 발행: VPC 삭제 이벤트
	vpcData := map[string]interface{}{
		"vpc_id": req.VPCID,
		"region": req.Region,
	}
	if err := s.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, req.Region, "deleted", vpcData); err != nil {
		s.logger.Warn("Failed to publish VPC deleted event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("vpc_id", req.VPCID),
			zap.Error(err))
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

// deleteAWSVPC: AWS VPC를 삭제합니다
func (s *Service) deleteAWSVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Delete VPC
	_, err = ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(req.VPCID),
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete VPC: %v", err), 502)
	}

	return nil
}

// ListSubnets: 서브넷 목록을 조회합니다
func (s *Service) ListSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	s.logger.Info("ListSubnets called",
		zap.String("provider", credential.Provider),
		zap.String("credential_id", credential.ID.String()))

	switch credential.Provider {
	case "aws":
		return s.listAWSSubnets(ctx, credential, req)
	case "gcp":
		return s.listGCPSubnets(ctx, credential, req)
	case "azure":
		return s.listAzureSubnets(ctx, credential, req)
	case "ncp":
		return s.listNCPSubnets(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// GetSubnet: 특정 서브넷을 조회합니다
func (s *Service) GetSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.getAWSSubnet(ctx, credential, req)
	case "gcp":
		return s.getGCPSubnet(ctx, credential, req)
	case "azure":
		return s.getAzureSubnet(ctx, credential, req)
	case "ncp":
		return s.getNCPSubnet(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// CreateSubnet: 새로운 서브넷을 생성합니다
func (s *Service) CreateSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.createAWSSubnet(ctx, credential, req)
	case "gcp":
		return s.createGCPSubnet(ctx, credential, req)
	case "azure":
		return s.createAzureSubnet(ctx, credential, req)
	case "ncp":
		return s.createNCPSubnet(ctx, credential, req)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// UpdateSubnet: 서브넷을 업데이트합니다
func (s *Service) UpdateSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.updateAWSSubnet(ctx, credential, req, subnetID, region)
	case "gcp":
		return s.updateGCPSubnet(ctx, credential, req, subnetID, region)
	case "azure":
		return s.updateAzureSubnet(ctx, credential, req, subnetID, region)
	case "ncp":
		return s.updateNCPSubnet(ctx, credential, req, subnetID, region)
	default:
		return nil, domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// DeleteSubnet: 서브넷을 삭제합니다
func (s *Service) DeleteSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	switch credential.Provider {
	case "aws":
		return s.deleteAWSSubnet(ctx, credential, req)
	case "gcp":
		return s.deleteGCPSubnet(ctx, credential, req)
	case "azure":
		return s.deleteAzureSubnet(ctx, credential, req)
	case "ncp":
		return s.deleteNCPSubnet(ctx, credential, req)
	default:
		return domain.NewDomainError(
			domain.ErrCodeNotSupported,
			fmt.Sprintf(ErrMsgUnsupportedProvider, credential.Provider),
			400,
		)
	}
}

// ListSecurityGroups: 보안 그룹 목록을 조회합니다
func (s *Service) ListSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	switch credential.Provider {
	case "aws":
		return s.listAWSSecurityGroups(ctx, credential, req)
	case "gcp":
		return s.listGCPSecurityGroups(ctx, credential, req)
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

// listAWSSecurityGroups: AWS 보안 그룹 목록을 조회합니다
func (s *Service) listAWSSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe security groups
	input := &ec2.DescribeSecurityGroupsInput{}
	if req.SecurityGroupID != "" {
		input.GroupIds = []string{req.SecurityGroupID}
	}
	if req.VPCID != "" {
		input.Filters = []ec2Types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{req.VPCID}},
		}
	}

	result, err := ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe security groups: %v", err), 502)
	}

	// Convert to DTOs
	securityGroups := make([]SecurityGroupInfo, 0, len(result.SecurityGroups))
	for _, sg := range result.SecurityGroups {
		sgInfo := SecurityGroupInfo{
			ID:          aws.ToString(sg.GroupId),
			Name:        aws.ToString(sg.GroupName),
			Description: aws.ToString(sg.Description),
			VPCID:       aws.ToString(sg.VpcId),
			Region:      req.Region,
			Rules:       s.convertSecurityGroupRules(sg.IpPermissions, sg.IpPermissionsEgress),
			Tags:        s.convertTags(sg.Tags),
		}
		securityGroups = append(securityGroups, sgInfo)
	}

	return &ListSecurityGroupsResponse{SecurityGroups: securityGroups}, nil
}

// listGCPSecurityGroups: GCP 보안 그룹 목록을 조회합니다
func (s *Service) listGCPSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// List firewall rules
	firewalls, err := computeService.Firewalls.List(projectID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list GCP firewall rules: %v", err), 502)
	}

	// Convert to DTOs
	securityGroups := make([]SecurityGroupInfo, 0, len(firewalls.Items))
	for _, firewall := range firewalls.Items {
		// Filter by VPC if specified
		if req.VPCID != "" {
			networkName := s.extractNetworkNameFromVPCID(req.VPCID)
			if firewall.Network != "" {
				firewallNetworkURL := firewall.Network
				if strings.Contains(firewallNetworkURL, "/networks/") {
					parts := strings.Split(firewallNetworkURL, "/networks/")
					if len(parts) == 2 {
						firewallNetworkName := parts[1]
						if firewallNetworkName != networkName {
							continue
						}
					}
				}
			}
		}

		// Convert GCP firewall rule to SecurityGroupInfo
		sgInfo := SecurityGroupInfo{
			ID:          firewall.Name,
			Name:        firewall.Name,
			Description: firewall.Description,
			VPCID:       s.extractCleanVPCID(firewall.Network),
			Region:      req.Region,
			Rules:       s.convertGCPFirewallRules(firewall),
			Tags:        make(map[string]string),
		}

		// Add GCP-specific fields to tags
		if firewall.Priority != 0 {
			sgInfo.Tags["priority"] = fmt.Sprintf("%d", firewall.Priority)
		}
		if firewall.Direction != "" {
			sgInfo.Tags["direction"] = firewall.Direction
		}
		// GCP firewall doesn't have Action field, it's determined by Allowed/Denied

		securityGroups = append(securityGroups, sgInfo)
	}

	s.logger.Info("GCP firewall rules listed successfully",
		zap.String("project_id", projectID),
		zap.Int("count", len(securityGroups)))

	return &ListSecurityGroupsResponse{SecurityGroups: securityGroups}, nil
}

// convertGCPFirewallRules: GCP 방화벽 규칙을 보안 그룹 규칙 정보로 변환합니다
func (s *Service) convertGCPFirewallRules(firewall *compute.Firewall) []SecurityGroupRuleInfo {
	var rules []SecurityGroupRuleInfo

	// Convert allowed rules
	for _, allowed := range firewall.Allowed {
		for _, portRange := range allowed.Ports {
			rule := SecurityGroupRuleInfo{
				Type:        "ingress",
				Protocol:    allowed.IPProtocol,
				FromPort:    int32(s.parsePort(portRange)),
				ToPort:      int32(s.parsePort(portRange)),
				CIDRBlocks:  firewall.SourceRanges,
				Description: firewall.Description,
			}
			rules = append(rules, rule)
		}
	}

	// Convert denied rules
	for _, denied := range firewall.Denied {
		for _, portRange := range denied.Ports {
			rule := SecurityGroupRuleInfo{
				Type:        "egress",
				Protocol:    denied.IPProtocol,
				FromPort:    int32(s.parsePort(portRange)),
				ToPort:      int32(s.parsePort(portRange)),
				CIDRBlocks:  firewall.DestinationRanges,
				Description: firewall.Description,
			}
			rules = append(rules, rule)
		}
	}

	return rules
}

// parsePort: 포트 범위 문자열에서 포트 번호를 파싱합니다
func (s *Service) parsePort(portRange string) int {
	if portRange == "" {
		return 0
	}
	if strings.Contains(portRange, "-") {
		parts := strings.Split(portRange, "-")
		if len(parts) == 2 {
			if port, err := strconv.Atoi(parts[0]); err == nil {
				return port
			}
		}
	}
	if port, err := strconv.Atoi(portRange); err == nil {
		return port
	}
	return 0
}

// GetSecurityGroup: 특정 보안 그룹을 조회합니다
func (s *Service) GetSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.getAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		return s.getGCPSecurityGroup(ctx, credential, req)
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

// getAWSSecurityGroup: 특정 AWS 보안 그룹을 조회합니다
func (s *Service) getAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe specific security group
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{req.SecurityGroupID},
	}

	result, err := ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe security group: %v", err), 502)
	}

	if len(result.SecurityGroups) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("security group not found: %s", req.SecurityGroupID), 404)
	}

	sg := result.SecurityGroups[0]
	sgInfo := &SecurityGroupInfo{
		ID:          aws.ToString(sg.GroupId),
		Name:        aws.ToString(sg.GroupName),
		Description: aws.ToString(sg.Description),
		VPCID:       aws.ToString(sg.VpcId),
		Region:      req.Region,
		Rules:       s.convertSecurityGroupRules(sg.IpPermissions, sg.IpPermissionsEgress),
		Tags:        s.convertTags(sg.Tags),
	}

	return sgInfo, nil
}

// getGCPSecurityGroup: 특정 GCP 보안 그룹을 조회합니다
func (s *Service) getGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Get firewall rule
	firewall, err := computeService.Firewalls.Get(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get GCP firewall rule: %v", err), 502)
	}

	// Convert to SecurityGroupInfo
	sgInfo := &SecurityGroupInfo{
		ID:          firewall.Name,
		Name:        firewall.Name,
		Description: firewall.Description,
		VPCID:       s.extractCleanVPCID(firewall.Network),
		Region:      req.Region,
		Rules:       s.convertGCPFirewallRules(firewall),
		Tags:        make(map[string]string),
	}

	// Add GCP-specific fields to tags
	if firewall.Priority != 0 {
		sgInfo.Tags["priority"] = fmt.Sprintf("%d", firewall.Priority)
	}
	if firewall.Direction != "" {
		sgInfo.Tags["direction"] = firewall.Direction
	}

	return sgInfo, nil
}

// CreateSecurityGroup: 새로운 보안 그룹을 생성합니다
func (s *Service) CreateSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.createAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		return s.createGCPSecurityGroup(ctx, credential, req)
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

// createAWSSecurityGroup: AWS 보안 그룹을 생성합니다
func (s *Service) createAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Create security group
	input := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(req.Name),
		Description: aws.String(req.Description),
		VpcId:       aws.String(req.VPCID),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeSecurityGroup,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateSecurityGroup(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create security group: %v", err), 502)
	}

	sgInfo := &SecurityGroupInfo{
		ID:          aws.ToString(result.GroupId),
		Name:        req.Name,
		Description: req.Description,
		VPCID:       req.VPCID,
		Region:      req.Region,
		Rules:       []SecurityGroupRuleInfo{}, // Empty initially
		Tags:        req.Tags,
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs/%s/security-groups", credential.Provider, req.VPCID),
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
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, req.Region, "created", sgData)
	}

	return sgInfo, nil
}

// createGCPSecurityGroup: GCP 보안 그룹을 생성합니다
func (s *Service) createGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	firewall := s.buildGCPFirewall(req)

	operation, err := computeService.Firewalls.Insert(projectID, firewall).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP firewall rule: %v", err), 502)
	}

	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall creation"); err != nil {
		return nil, err
	}

	sgInfo := s.buildSecurityGroupInfo(req, projectID)

	s.logger.Info("GCP firewall rule created successfully",
		zap.String("firewall_name", req.Name),
		zap.String("project_id", projectID))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs/%s/security-groups", credential.Provider, req.VPCID),
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
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, req.Region, "created", sgData)
	}

	return sgInfo, nil
}

// convertGCPFirewallRulesFromRequest: 요청으로부터 GCP 방화벽 규칙을 보안 그룹 규칙 정보로 변환합니다
func (s *Service) convertGCPFirewallRulesFromRequest(req CreateSecurityGroupRequest) []SecurityGroupRuleInfo {
	var rules []SecurityGroupRuleInfo

	for _, port := range req.Ports {
		rule := SecurityGroupRuleInfo{
			Type:        strings.ToLower(req.Direction),
			Protocol:    req.Protocol,
			FromPort:    int32(s.parsePort(port)),
			ToPort:      int32(s.parsePort(port)),
			CIDRBlocks:  req.SourceRanges,
			Description: req.Description,
		}
		rules = append(rules, rule)
	}

	return rules
}

// UpdateSecurityGroup: 보안 그룹을 업데이트합니다
func (s *Service) UpdateSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, securityGroupID, region string) (*SecurityGroupInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.updateAWSSecurityGroup(ctx, credential, req, securityGroupID, region)
	case "gcp":
		return s.updateGCPSecurityGroup(ctx, credential, req, securityGroupID, region)
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

// updateAWSSecurityGroup: AWS 보안 그룹을 업데이트합니다
func (s *Service) updateAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, securityGroupID, region string) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Update security group tags if provided
	if req.Name != "" || req.Description != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{securityGroupID},
			Tags:      tags,
		})
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update security group tags: %v", err), 502)
		}
	}

	// Get updated security group info
	sgInfo, err := s.GetSecurityGroup(ctx, credential, GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/security-groups/%s", credential.Provider, securityGroupID),
		map[string]interface{}{
			"security_group_id": securityGroupID,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": securityGroupID,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, region, "updated", sgData)
	}

	return sgInfo, nil
}

// updateGCPSecurityGroup: GCP 보안 그룹을 업데이트합니다
func (s *Service) updateGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, firewallName, region string) (*SecurityGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Get current firewall rule
	currentFirewall, err := computeService.Firewalls.Get(projectID, firewallName).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get current firewall rule: %v", err), 502)
	}

	// Update firewall rule fields
	updatedFirewall := &compute.Firewall{
		Name:         currentFirewall.Name,
		Description:  currentFirewall.Description,
		Network:      currentFirewall.Network,
		Direction:    currentFirewall.Direction,
		Priority:     currentFirewall.Priority,
		Allowed:      currentFirewall.Allowed,
		Denied:       currentFirewall.Denied,
		SourceRanges: currentFirewall.SourceRanges,
		TargetTags:   currentFirewall.TargetTags,
	}

	// Update description if provided
	if req.Description != "" {
		updatedFirewall.Description = req.Description
	}

	// Update firewall rule
	operation, err := computeService.Firewalls.Update(projectID, firewallName, updatedFirewall).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update GCP firewall rule: %v", err), 502)
	}

	// Wait for operation to complete
	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall update"); err != nil {
		return nil, err
	}

	// Get updated firewall rule info
	sgInfo, err := s.GetSecurityGroup(ctx, credential, GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: firewallName,
		Region:          region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/security-groups/%s", credential.Provider, firewallName),
		map[string]interface{}{
			"security_group_id": firewallName,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": firewallName,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, region, "updated", sgData)
	}

	return sgInfo, nil
}

// DeleteSecurityGroup: 보안 그룹을 삭제합니다
func (s *Service) DeleteSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	switch credential.Provider {
	case "aws":
		return s.deleteAWSSecurityGroup(ctx, credential, req)
	case "gcp":
		return s.deleteGCPSecurityGroup(ctx, credential, req)
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
}

// deleteAWSSecurityGroup: AWS 보안 그룹을 삭제합니다
func (s *Service) deleteAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Delete security group
	_, err = ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(req.SecurityGroupID),
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete security group: %v", err), 502)
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/security-groups/%s", credential.Provider, req.SecurityGroupID),
		map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, req.Region, "deleted", sgData)
	}

	return nil
}

// deleteGCPSecurityGroup: GCP 보안 그룹을 삭제합니다
func (s *Service) deleteGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	// Get project ID from credential
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Delete firewall rule
	operation, err := computeService.Firewalls.Delete(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete GCP firewall rule: %v", err), 502)
	}

	// Wait for operation to complete
	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall deletion"); err != nil {
		return err
	}

	s.logger.Info("GCP firewall rule deleted successfully",
		zap.String("firewall_name", req.SecurityGroupID),
		zap.String("project_id", projectID))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/security-groups/%s", credential.Provider, req.SecurityGroupID),
		map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		sgData := map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		_ = s.eventPublisher.PublishSecurityGroupEvent(ctx, credential.Provider, credentialID, req.Region, "deleted", sgData)
	}

	return nil
}

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
func (s *Service) removeGCPFirewallRule(ctx context.Context, credential *domain.Credential, req RemoveFirewallRuleRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	currentFirewall, err := computeService.Firewalls.Get(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get current firewall rule: %v", err), 502)
	}

	updatedFirewall := s.cloneFirewall(currentFirewall)
	updatedFirewall.Allowed = s.removePortsFromAllowed(currentFirewall.Allowed, req.Protocol, req.Ports)
	updatedFirewall.Denied = s.removePortsFromDenied(currentFirewall.Denied, req.Protocol, req.Ports)

	operation, err := computeService.Firewalls.Update(projectID, req.SecurityGroupID, updatedFirewall).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update GCP firewall rule: %v", err), 502)
	}

	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall rule removal"); err != nil {
		return nil, err
	}

	getReq := GetSecurityGroupRequest{
		CredentialID:    req.CredentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

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
func (s *Service) addGCPFirewallRule(ctx context.Context, credential *domain.Credential, req AddFirewallRuleRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	currentFirewall, err := computeService.Firewalls.Get(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get current firewall rule: %v", err), 502)
	}

	updatedFirewall := s.cloneFirewall(currentFirewall)
	updatedFirewall.Allowed = s.addPortsToAllowed(currentFirewall.Allowed, req.Protocol, req.Ports, req.Action)
	updatedFirewall.Denied = s.addPortsToDenied(currentFirewall.Denied, req.Protocol, req.Ports, req.Action)

	operation, err := computeService.Firewalls.Update(projectID, req.SecurityGroupID, updatedFirewall).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update GCP firewall rule: %v", err), 502)
	}

	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall rule addition"); err != nil {
		return nil, err
	}

	getReq := GetSecurityGroupRequest{
		CredentialID:    req.CredentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// Helper methods

// setupGCPComputeService: GCP Compute 서비스를 설정합니다
func (s *Service) setupGCPComputeService(ctx context.Context, credential *domain.Credential) (*compute.Service, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute service: %v", err), 502)
	}

	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	return computeService, projectID, nil
}

// buildGCPFirewall: 요청으로부터 GCP 방화벽 객체를 생성합니다
func (s *Service) buildGCPFirewall(req CreateSecurityGroupRequest) *compute.Firewall {
	firewall := &compute.Firewall{
		Name:        req.Name,
		Description: req.Description,
		Network:     req.VPCID,
		Direction:   req.Direction,
		Priority:    req.Priority,
	}

	if req.Action == ActionAllow {
		firewall.Allowed = s.buildAllowedRules(req)
		firewall.SourceRanges = req.SourceRanges
	} else if req.Action == ActionDeny {
		firewall.Denied = s.buildDeniedRules(req)
		firewall.SourceRanges = req.SourceRanges
	}

	if len(req.TargetTags) > 0 {
		firewall.TargetTags = req.TargetTags
	}

	return firewall
}

// buildAllowedRules: 허용 규칙을 생성합니다
func (s *Service) buildAllowedRules(req CreateSecurityGroupRequest) []*compute.FirewallAllowed {
	if len(req.Allowed) > 0 {
		var allowedRules []*compute.FirewallAllowed
		for _, allowed := range req.Allowed {
			allowedRule := &compute.FirewallAllowed{
				IPProtocol: allowed.Protocol,
			}
			if len(allowed.Ports) > 0 && allowed.Protocol != ProtocolICMP {
				allowedRule.Ports = allowed.Ports
			}
			allowedRules = append(allowedRules, allowedRule)
		}
		return allowedRules
	}

	// Fallback to old method for backward compatibility
	if req.Protocol != "" && len(req.Ports) > 0 {
		var allowedRules []*compute.FirewallAllowed
		for _, port := range req.Ports {
			allowedRule := &compute.FirewallAllowed{
				IPProtocol: req.Protocol,
				Ports:      []string{port},
			}
			allowedRules = append(allowedRules, allowedRule)
		}
		return allowedRules
	}

	return nil
}

// buildDeniedRules: 거부 규칙을 생성합니다
func (s *Service) buildDeniedRules(req CreateSecurityGroupRequest) []*compute.FirewallDenied {
	if len(req.Denied) > 0 {
		var deniedRules []*compute.FirewallDenied
		for _, denied := range req.Denied {
			deniedRule := &compute.FirewallDenied{
				IPProtocol: denied.Protocol,
			}
			if len(denied.Ports) > 0 && denied.Protocol != ProtocolICMP {
				deniedRule.Ports = denied.Ports
			}
			deniedRules = append(deniedRules, deniedRule)
		}
		return deniedRules
	}

	// Fallback to old method for backward compatibility
	if req.Protocol != "" && len(req.Ports) > 0 {
		var deniedRules []*compute.FirewallDenied
		for _, port := range req.Ports {
			deniedRule := &compute.FirewallDenied{
				IPProtocol: req.Protocol,
				Ports:      []string{port},
			}
			deniedRules = append(deniedRules, deniedRule)
		}
		return deniedRules
	}

	return nil
}

// buildSecurityGroupInfo: 요청으로부터 보안 그룹 정보를 생성합니다
func (s *Service) buildSecurityGroupInfo(req CreateSecurityGroupRequest, projectID string) *SecurityGroupInfo {
	sgInfo := &SecurityGroupInfo{
		ID:          req.Name,
		Name:        req.Name,
		Description: req.Description,
		VPCID:       req.VPCID,
		Region:      req.Region,
		Rules:       s.convertGCPFirewallRulesFromRequest(req),
		Tags:        req.Tags,
	}

	if sgInfo.Tags == nil {
		sgInfo.Tags = make(map[string]string)
	}

	sgInfo.Tags["direction"] = req.Direction
	sgInfo.Tags["priority"] = fmt.Sprintf("%d", req.Priority)
	sgInfo.Tags["action"] = req.Action

	return sgInfo
}

// cloneFirewall: 방화벽 객체를 복제합니다
func (s *Service) cloneFirewall(firewall *compute.Firewall) *compute.Firewall {
	return &compute.Firewall{
		Name:         firewall.Name,
		Description:  firewall.Description,
		Network:      firewall.Network,
		Direction:    firewall.Direction,
		Priority:     firewall.Priority,
		SourceRanges: firewall.SourceRanges,
		TargetTags:   firewall.TargetTags,
	}
}

// removePortsFromAllowed: 허용 규칙에서 포트를 제거합니다
func (s *Service) removePortsFromAllowed(allowed []*compute.FirewallAllowed, protocol string, portsToRemove []string) []*compute.FirewallAllowed {
	if len(allowed) == 0 {
		return allowed
	}

	var updatedAllowed []*compute.FirewallAllowed
	for _, rule := range allowed {
		if rule.IPProtocol == protocol {
			remainingPorts := s.filterPorts(rule.Ports, portsToRemove)
			if len(remainingPorts) > 0 {
				updatedAllowed = append(updatedAllowed, &compute.FirewallAllowed{
					IPProtocol: rule.IPProtocol,
					Ports:      remainingPorts,
				})
			}
		} else {
			updatedAllowed = append(updatedAllowed, rule)
		}
	}
	return updatedAllowed
}

// removePortsFromDenied: 거부 규칙에서 포트를 제거합니다
func (s *Service) removePortsFromDenied(denied []*compute.FirewallDenied, protocol string, portsToRemove []string) []*compute.FirewallDenied {
	if len(denied) == 0 {
		return denied
	}

	var updatedDenied []*compute.FirewallDenied
	for _, rule := range denied {
		if rule.IPProtocol == protocol {
			remainingPorts := s.filterPorts(rule.Ports, portsToRemove)
			if len(remainingPorts) > 0 {
				updatedDenied = append(updatedDenied, &compute.FirewallDenied{
					IPProtocol: rule.IPProtocol,
					Ports:      remainingPorts,
				})
			}
		} else {
			updatedDenied = append(updatedDenied, rule)
		}
	}
	return updatedDenied
}

// addPortsToAllowed: 허용 규칙에 포트를 추가합니다
func (s *Service) addPortsToAllowed(allowed []*compute.FirewallAllowed, protocol string, portsToAdd []string, action string) []*compute.FirewallAllowed {
	if action == ActionDeny {
		return allowed
	}

	if len(allowed) == 0 {
		return []*compute.FirewallAllowed{
			{
				IPProtocol: protocol,
				Ports:      portsToAdd,
			},
		}
	}

	var updatedAllowed []*compute.FirewallAllowed
	foundProtocol := false

	for _, rule := range allowed {
		if rule.IPProtocol == protocol {
			mergedPorts := s.mergePorts(rule.Ports, portsToAdd)
			updatedAllowed = append(updatedAllowed, &compute.FirewallAllowed{
				IPProtocol: rule.IPProtocol,
				Ports:      mergedPorts,
			})
			foundProtocol = true
		} else {
			updatedAllowed = append(updatedAllowed, rule)
		}
	}

	if !foundProtocol {
		updatedAllowed = append(updatedAllowed, &compute.FirewallAllowed{
			IPProtocol: protocol,
			Ports:      portsToAdd,
		})
	}

	return updatedAllowed
}

// addPortsToDenied: 거부 규칙에 포트를 추가합니다
func (s *Service) addPortsToDenied(denied []*compute.FirewallDenied, protocol string, portsToAdd []string, action string) []*compute.FirewallDenied {
	if action != ActionDeny {
		return denied
	}

	if len(denied) == 0 {
		return []*compute.FirewallDenied{
			{
				IPProtocol: protocol,
				Ports:      portsToAdd,
			},
		}
	}

	var updatedDenied []*compute.FirewallDenied
	foundProtocol := false

	for _, rule := range denied {
		if rule.IPProtocol == protocol {
			mergedPorts := s.mergePorts(rule.Ports, portsToAdd)
			updatedDenied = append(updatedDenied, &compute.FirewallDenied{
				IPProtocol: rule.IPProtocol,
				Ports:      mergedPorts,
			})
			foundProtocol = true
		} else {
			updatedDenied = append(updatedDenied, rule)
		}
	}

	if !foundProtocol {
		updatedDenied = append(updatedDenied, &compute.FirewallDenied{
			IPProtocol: protocol,
			Ports:      portsToAdd,
		})
	}

	return updatedDenied
}

// filterPorts: 포트 목록에서 특정 포트를 필터링합니다
func (s *Service) filterPorts(ports []string, portsToRemove []string) []string {
	removeSet := make(map[string]bool)
	for _, port := range portsToRemove {
		removeSet[port] = true
	}

	var remaining []string
	for _, port := range ports {
		if !removeSet[port] {
			remaining = append(remaining, port)
		}
	}
	return remaining
}

// mergePorts: 기존 포트와 새 포트를 병합합니다
func (s *Service) mergePorts(existingPorts, newPorts []string) []string {
	portSet := make(map[string]bool)
	for _, port := range existingPorts {
		portSet[port] = true
	}

	var merged []string
	merged = append(merged, existingPorts...)

	for _, port := range newPorts {
		if !portSet[port] {
			merged = append(merged, port)
		}
	}
	return merged
}

// waitForGCPOperation: GCP 작업이 완료될 때까지 대기합니다
func (s *Service) waitForGCPOperation(ctx context.Context, computeService *compute.Service, projectID, operationName, operationType string) error {
	if operationName == "" {
		return nil
	}

	ticker := time.NewTicker(OperationPollInterval)
	defer ticker.Stop()

	timeout := time.After(OperationTimeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return domain.NewDomainError(
				domain.ErrCodeTimeout,
				fmt.Sprintf("GCP %s operation timed out after %v", operationType, OperationTimeout),
				408,
			)
		case <-ticker.C:
			op, err := computeService.GlobalOperations.Get(projectID, operationName).Context(ctx).Do()
			if err != nil {
				return domain.NewDomainError(
					domain.ErrCodeInternalError,
					fmt.Sprintf("failed to check operation status: %v", err),
					500,
				)
			}

			if op.Status == OperationStatusDone {
				if op.Error != nil {
					return domain.NewDomainError(
						domain.ErrCodeInternalError,
						fmt.Sprintf("GCP %s operation failed: %v", operationType, op.Error),
						500,
					)
				}
				return nil
			}
		}
	}
}

// createGCPComputeClient: GCP Compute 클라이언트를 생성합니다
func (s *Service) createGCPComputeClient(ctx context.Context, credential *domain.Credential) (*compute.Service, error) {
	// Decrypt credential
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
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
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal service account key: %v", err), 500)
	}

	// Create credentials from service account key
	creds, err := google.CredentialsFromJSON(ctx, keyBytes, compute.CloudPlatformScope)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create credentials: %v", err), 502)
	}

	// Create compute service
	computeService, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create compute service: %v", err), 502)
	}

	return computeService, nil
}

// createEC2Client: AWS EC2 클라이언트를 생성합니다
func (s *Service) createEC2Client(ctx context.Context, credential *domain.Credential, region string) (*ec2.Client, error) {
	// Decrypt credential
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Extract AWS credentials (same as kubernetes service)
	accessKey, ok := decryptedData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := decryptedData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	// Debug: Log the extracted credentials
	s.logger.Info("Extracted AWS credentials",
		zap.String("access_key", accessKey),
		zap.String("secret_key", secretKey[:10]+"..."))

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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	return ec2.NewFromConfig(cfg), nil
}

// getTagValue: 태그 목록에서 특정 키의 값을 조회합니다
func (s *Service) getTagValue(tags []ec2Types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// convertTags: 태그 목록을 맵으로 변환합니다
func (s *Service) convertTags(tags []ec2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// isSubnetPublic: 서브넷이 공개 서브넷인지 확인합니다
func (s *Service) isSubnetPublic(ctx context.Context, ec2Client *ec2.Client, subnetID string) bool {
	// Get route tables for the subnet
	input := &ec2.DescribeRouteTablesInput{
		Filters: []ec2Types.Filter{
			{Name: aws.String("association.subnet-id"), Values: []string{subnetID}},
		},
	}

	result, err := ec2Client.DescribeRouteTables(ctx, input)
	if err != nil {
		s.logger.Warn("Failed to describe route tables", zap.String("subnet_id", subnetID), zap.Error(err))
		return false
	}

	// Check if any route table has a route to internet gateway
	for _, rt := range result.RouteTables {
		for _, route := range rt.Routes {
			if route.GatewayId != nil && aws.ToString(route.GatewayId) != "local" {
				// Check if this is an internet gateway
				igwInput := &ec2.DescribeInternetGatewaysInput{
					InternetGatewayIds: []string{aws.ToString(route.GatewayId)},
				}
				igwResult, err := ec2Client.DescribeInternetGateways(ctx, igwInput)
				if err == nil && len(igwResult.InternetGateways) > 0 {
					return true
				}
			}
		}
	}

	return false
}

// convertSecurityGroupRules: AWS IP 권한을 보안 그룹 규칙 정보로 변환합니다
func (s *Service) convertSecurityGroupRules(ingress, egress []ec2Types.IpPermission) []SecurityGroupRuleInfo {
	rules := make([]SecurityGroupRuleInfo, 0)

	// Convert ingress rules
	for _, perm := range ingress {
		rule := SecurityGroupRuleInfo{
			Type:         "ingress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	// Convert egress rules
	for _, perm := range egress {
		rule := SecurityGroupRuleInfo{
			Type:         "egress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

// AddSecurityGroupRule adds a rule to a security group
func (s *Service) AddSecurityGroupRule(ctx context.Context, credential *domain.Credential, req AddSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Build IP permissions
	ipPermission := ec2Types.IpPermission{
		IpProtocol: aws.String(req.Protocol),
		FromPort:   aws.Int32(req.FromPort),
		ToPort:     aws.Int32(req.ToPort),
	}

	// Add CIDR blocks
	if len(req.CIDRBlocks) > 0 {
		for _, cidr := range req.CIDRBlocks {
			ipPermission.IpRanges = append(ipPermission.IpRanges, ec2Types.IpRange{
				CidrIp:      aws.String(cidr),
				Description: aws.String(req.Description),
			})
		}
	}

	// Add source groups
	if len(req.SourceGroups) > 0 {
		for _, groupID := range req.SourceGroups {
			ipPermission.UserIdGroupPairs = append(ipPermission.UserIdGroupPairs, ec2Types.UserIdGroupPair{
				GroupId:     aws.String(groupID),
				Description: aws.String(req.Description),
			})
		}
	}

	// Add rule based on type
	if req.Type == "ingress" {
		_, err = ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	} else {
		_, err = ec2Client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	}

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to add security group rule: %v", err), 502)
	}

	// Get updated security group info
	getReq := GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// RemoveSecurityGroupRule removes a rule from a security group
func (s *Service) RemoveSecurityGroupRule(ctx context.Context, credential *domain.Credential, req RemoveSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Build IP permissions
	ipPermission := ec2Types.IpPermission{
		IpProtocol: aws.String(req.Protocol),
		FromPort:   aws.Int32(req.FromPort),
		ToPort:     aws.Int32(req.ToPort),
	}

	// Add CIDR blocks
	if len(req.CIDRBlocks) > 0 {
		for _, cidr := range req.CIDRBlocks {
			ipPermission.IpRanges = append(ipPermission.IpRanges, ec2Types.IpRange{
				CidrIp: aws.String(cidr),
			})
		}
	}

	// Add source groups
	if len(req.SourceGroups) > 0 {
		for _, groupID := range req.SourceGroups {
			ipPermission.UserIdGroupPairs = append(ipPermission.UserIdGroupPairs, ec2Types.UserIdGroupPair{
				GroupId: aws.String(groupID),
			})
		}
	}

	// Remove rule based on type
	if req.Type == "ingress" {
		_, err = ec2Client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	} else {
		_, err = ec2Client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	}

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to remove security group rule: %v", err), 502)
	}

	// Get updated security group info
	getReq := GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get current security group: %v", err), 502)
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
			s.logger.Warn("Failed to remove existing rule", zap.Error(err))
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
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to add ingress rule: %v", err), 502)
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
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to add egress rule: %v", err), 502)
		}
	}

	// Get updated security group info
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// GCP Subnet Functions

// listGCPSubnets: GCP 서브넷 목록을 조회합니다
func (s *Service) listGCPSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	// Create GCP Compute client
	computeService, err := s.createGCPComputeClient(ctx, credential)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute client: %v", err), 502)
	}

	// Decrypt credential to get project ID
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID from decrypted data
	projectID, ok := decryptedData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// List subnets
	subnets, err := computeService.Subnetworks.List(projectID, req.Region).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list GCP subnets: %v", err), 502)
	}

	// Convert to DTOs
	subnetInfos := make([]SubnetInfo, 0, len(subnets.Items))
	for _, subnet := range subnets.Items {
		// Extract clean format IDs
		subnetIDClean := s.extractCleanSubnetID(subnet.SelfLink)
		vpcIDClean := s.extractCleanVPCID(subnet.Network)
		regionClean := s.extractCleanRegionID(subnet.Region)
		regionName := s.extractRegionName(subnet.Region)

		subnetInfo := SubnetInfo{
			ID:                    subnetIDClean, // Clean format: projects/{project}/regions/{region}/subnetworks/{name}
			Name:                  subnet.Name,
			VPCID:                 vpcIDClean, // Clean format: projects/{project}/global/networks/{name}
			CIDRBlock:             subnet.IpCidrRange,
			AvailabilityZone:      regionClean, // Clean format: projects/{project}/regions/{region}
			State:                 "READY",     // GCP subnets are always ready when listed
			IsPublic:              false,       // GCP doesn't have public/private concept like AWS
			Region:                regionName,  // Region name only (e.g., asia-northeast3)
			Description:           subnet.Description,
			GatewayAddress:        subnet.GatewayAddress,
			PrivateIPGoogleAccess: subnet.PrivateIpGoogleAccess,
			FlowLogs:              subnet.EnableFlowLogs,
			CreationTimestamp:     subnet.CreationTimestamp,
			Tags:                  make(map[string]string), // GCP subnets don't have labels in the same way
		}

		// Add GCP-specific fields to tags for backward compatibility
		if subnet.PrivateIpGoogleAccess {
			subnetInfo.Tags["private_ip_google_access"] = "true"
		}
		if subnet.EnableFlowLogs {
			subnetInfo.Tags["flow_logs"] = "true"
		}

		subnetInfos = append(subnetInfos, subnetInfo)
	}

	return &ListSubnetsResponse{Subnets: subnetInfos}, nil
}

// getGCPSubnet: 특정 GCP 서브넷을 조회합니다
func (s *Service) getGCPSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, err := s.createGCPComputeClient(ctx, credential)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute client: %v", err), 502)
	}

	// Decrypt credential to get project ID
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID from decrypted data
	projectID, ok := decryptedData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(req.SubnetID)
	if subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", req.SubnetID), 400)
	}

	// Get subnet
	subnet, err := computeService.Subnetworks.Get(projectID, req.Region, subnetName).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get GCP subnet: %v", err), 502)
	}

	// Extract clean format IDs
	subnetIDClean := s.extractCleanSubnetID(subnet.SelfLink)
	vpcIDClean := s.extractCleanVPCID(subnet.Network)
	regionClean := s.extractCleanRegionID(subnet.Region)
	regionName := s.extractRegionName(subnet.Region)

	// Convert to DTO
	subnetInfo := &SubnetInfo{
		ID:                    subnetIDClean, // Clean format: projects/{project}/regions/{region}/subnetworks/{name}
		Name:                  subnet.Name,
		VPCID:                 vpcIDClean, // Clean format: projects/{project}/global/networks/{name}
		CIDRBlock:             subnet.IpCidrRange,
		AvailabilityZone:      regionClean, // Clean format: projects/{project}/regions/{region}
		State:                 "READY",
		IsPublic:              false,
		Region:                regionName, // Region name only (e.g., asia-northeast3)
		Description:           subnet.Description,
		GatewayAddress:        subnet.GatewayAddress,
		PrivateIPGoogleAccess: subnet.PrivateIpGoogleAccess,
		FlowLogs:              subnet.EnableFlowLogs,
		CreationTimestamp:     subnet.CreationTimestamp,
		Tags:                  make(map[string]string), // GCP subnets don't have labels in the same way
	}

	// Add GCP-specific fields to tags for backward compatibility
	if subnet.PrivateIpGoogleAccess {
		subnetInfo.Tags["private_ip_google_access"] = "true"
	}
	if subnet.EnableFlowLogs {
		subnetInfo.Tags["flow_logs"] = "true"
	}

	return subnetInfo, nil
}

// createGCPSubnet: 새로운 GCP 서브넷을 생성합니다
func (s *Service) createGCPSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, err := s.createGCPComputeClient(ctx, credential)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute client: %v", err), 502)
	}

	// Decrypt credential to get project ID
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID from decrypted data
	projectID, ok := decryptedData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract network name from VPC ID
	networkName := s.extractNetworkNameFromVPCID(req.VPCID)
	if networkName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid VPC ID format: %s", req.VPCID), 400)
	}

	// Create subnet
	subnet := &compute.Subnetwork{
		Name:                  req.Name,
		IpCidrRange:           req.CIDRBlock,
		Network:               req.VPCID,
		Region:                req.Region,
		PrivateIpGoogleAccess: req.PrivateIPGoogleAccess,
		EnableFlowLogs:        req.FlowLogs,
		Description:           req.Description,
		// Labels not supported in GCP subnets
	}

	operation, err := computeService.Subnetworks.Insert(projectID, req.Region, subnet).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP subnet: %v", err), 502)
	}

	s.logger.Info("GCP subnet creation initiated",
		zap.String("subnet_name", req.Name),
		zap.String("project_id", projectID),
		zap.String("region", req.Region),
		zap.String("operation_id", operation.Name))

	// Wait for operation to complete (optional)
	// For now, return a mock response
	subnetInfo := &SubnetInfo{
		ID:               fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectID, req.Region, req.Name),
		Name:             req.Name,
		VPCID:            req.VPCID,
		CIDRBlock:        req.CIDRBlock,
		AvailabilityZone: req.Region,
		State:            "CREATING",
		IsPublic:         false,
		Region:           req.Region,
		Tags:             req.Tags,
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
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

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"cidr_block":    subnetInfo.CIDRBlock,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		_ = s.eventPublisher.PublishSubnetEvent(ctx, credential.Provider, credentialID, req.VPCID, "created", subnetData)
	}

	return subnetInfo, nil
}

// updateGCPSubnet: GCP 서브넷을 업데이트합니다
func (s *Service) updateGCPSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, err := s.createGCPComputeClient(ctx, credential)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute client: %v", err), 502)
	}

	// Decrypt credential to get project ID
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID from decrypted data
	projectID, ok := decryptedData["project_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(subnetID)
	if subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", subnetID), 400)
	}

	// Get current subnet
	currentSubnet, err := computeService.Subnetworks.Get(projectID, region, subnetName).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get current subnet: %v", err), 502)
	}

	// Update subnet
	subnet := &compute.Subnetwork{
		Name:                  currentSubnet.Name,
		IpCidrRange:           currentSubnet.IpCidrRange,
		Network:               currentSubnet.Network,
		Region:                currentSubnet.Region,
		PrivateIpGoogleAccess: currentSubnet.PrivateIpGoogleAccess,
		EnableFlowLogs:        currentSubnet.EnableFlowLogs,
		Description:           currentSubnet.Description,
		// Labels not supported in GCP subnets
	}

	// Update fields if provided
	if req.Description != "" {
		subnet.Description = req.Description
	}
	if req.PrivateIPGoogleAccess != nil {
		subnet.PrivateIpGoogleAccess = *req.PrivateIPGoogleAccess
	}
	if req.FlowLogs != nil {
		subnet.EnableFlowLogs = *req.FlowLogs
	}
	// Labels not supported in GCP subnets
	// Tags are ignored for GCP subnets

	operation, err := computeService.Subnetworks.Patch(projectID, region, subnetName, subnet).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update GCP subnet: %v", err), 502)
	}

	s.logger.Info("GCP subnet update initiated",
		zap.String("subnet_name", subnetName),
		zap.String("project_id", projectID),
		zap.String("region", region),
		zap.String("operation_id", operation.Name))

	// Get updated subnet info
	subnetInfo, err := s.getGCPSubnet(ctx, credential, GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/vpcs/%s/subnets/%s", credential.Provider, subnetInfo.VPCID, subnetID),
		map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		}
		_ = s.eventPublisher.PublishSubnetEvent(ctx, credential.Provider, credentialID, subnetInfo.VPCID, "updated", subnetData)
	}

	return subnetInfo, nil
}

// deleteGCPSubnet: GCP 서브넷을 삭제합니다
func (s *Service) deleteGCPSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	// Create GCP Compute client
	computeService, err := s.createGCPComputeClient(ctx, credential)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP compute client: %v", err), 502)
	}

	// Decrypt credential to get project ID
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID from decrypted data
	projectID, ok := decryptedData["project_id"].(string)
	if !ok {
		return domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			ErrMsgProjectIDNotFound,
			400,
		)
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(req.SubnetID)
	if subnetName == "" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", req.SubnetID), 400)
	}

	// Delete subnet
	operation, err := computeService.Subnetworks.Delete(projectID, req.Region, subnetName).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete GCP subnet: %v", err), 502)
	}

	s.logger.Info("GCP subnet deletion initiated",
		zap.String("subnet_name", subnetName),
		zap.String("project_id", projectID),
		zap.String("region", req.Region),
		zap.String("operation_id", operation.Name))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/subnets/%s", credential.Provider, req.SubnetID),
		map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	return nil
}

// extractSubnetNameFromSubnetID: GCP 서브넷 ID에서 서브넷 이름을 추출합니다
func (s *Service) extractSubnetNameFromSubnetID(subnetID string) string {
	// Handle full GCP subnet path: projects/{project}/regions/{region}/subnetworks/{subnet_name}
	if strings.Contains(subnetID, "/subnetworks/") {
		parts := strings.Split(subnetID, "/subnetworks/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	// Handle simple subnet name
	return subnetID
}

// extractCleanSubnetID: GCP URL에서 서브넷 ID를 추출합니다
func (s *Service) extractCleanSubnetID(selfLink string) string {
	// From: https://www.googleapis.com/compute/v1/projects/leafy-environs-445206-d2/regions/asia-northeast3/subnetworks/default
	// To:   projects/leafy-environs-445206-d2/regions/asia-northeast3/subnetworks/default
	if strings.Contains(selfLink, "/projects/") {
		parts := strings.Split(selfLink, "/projects/")
		if len(parts) == 2 {
			return "projects/" + parts[1]
		}
	}
	return selfLink
}

// extractCleanVPCID: GCP URL에서 VPC ID를 추출합니다
func (s *Service) extractCleanVPCID(networkURL string) string {
	// From: https://www.googleapis.com/compute/v1/projects/leafy-environs-445206-d2/global/networks/default
	// To:   projects/leafy-environs-445206-d2/global/networks/default
	if strings.Contains(networkURL, "/projects/") {
		parts := strings.Split(networkURL, "/projects/")
		if len(parts) == 2 {
			return "projects/" + parts[1]
		}
	}
	return networkURL
}

// extractCleanRegionID: GCP URL에서 리전 ID를 추출합니다
func (s *Service) extractCleanRegionID(regionURL string) string {
	// From: https://www.googleapis.com/compute/v1/projects/leafy-environs-445206-d2/regions/asia-northeast3
	// To:   projects/leafy-environs-445206-d2/regions/asia-northeast3
	if strings.Contains(regionURL, "/projects/") {
		parts := strings.Split(regionURL, "/projects/")
		if len(parts) == 2 {
			return "projects/" + parts[1]
		}
	}
	return regionURL
}

// extractProjectID extracts project ID from GCP URL

// extractRegionName: GCP URL에서 리전 이름을 추출합니다
func (s *Service) extractRegionName(regionURL string) string {
	// From: https://www.googleapis.com/compute/v1/projects/leafy-environs-445206-d2/regions/asia-northeast3
	// To:   asia-northeast3
	if strings.Contains(regionURL, "/regions/") {
		parts := strings.Split(regionURL, "/regions/")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

// getFirewallRulesCount: 특정 네트워크의 방화벽 규칙 개수를 조회합니다
func (s *Service) getFirewallRulesCount(ctx context.Context, computeService *compute.Service, projectID, networkName string) (int, error) {
	// List firewall rules for the specific network
	firewalls, err := computeService.Firewalls.List(projectID).Context(ctx).Do()
	if err != nil {
		return 0, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list firewall rules: %v", err), 502)
	}

	count := 0
	for _, firewall := range firewalls.Items {
		// Check if firewall rule applies to this network
		if firewall.Network != "" {
			// Extract network name from network URL
			networkURL := firewall.Network
			if strings.Contains(networkURL, "/networks/") {
				parts := strings.Split(networkURL, "/networks/")
				if len(parts) == 2 {
					firewallNetworkName := parts[1]
					if firewallNetworkName == networkName {
						count++
					}
				}
			}
		}
	}

	return count, nil
}

// getGatewayInfo: 특정 네트워크의 게이트웨이 정보를 조회합니다
func (s *Service) getGatewayInfo(ctx context.Context, computeService *compute.Service, projectID, networkName string) (*GatewayInfo, error) {
	routers, err := s.listRouters(ctx, computeService, projectID)
	if err != nil {
		return nil, err
	}

	return s.findGatewayForNetwork(routers, networkName), nil
}

// listRouters: 모든 리전의 라우터 목록을 조회합니다
func (s *Service) listRouters(ctx context.Context, computeService *compute.Service, projectID string) (*compute.RouterAggregatedList, error) {
	routers, err := computeService.Routers.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list routers: %v", err), 502)
	}
	return routers, nil
}

// findGatewayForNetwork: 특정 네트워크의 게이트웨이 정보를 찾습니다
func (s *Service) findGatewayForNetwork(routers *compute.RouterAggregatedList, networkName string) *GatewayInfo {
	for _, routerList := range routers.Items {
		for _, router := range routerList.Routers {
			if !s.isRouterConnectedToNetwork(router, networkName) {
				continue
			}

			if gatewayInfo := s.checkRouterForGateway(router); gatewayInfo != nil {
				return gatewayInfo
			}
		}
	}
	return nil
}

// isRouterConnectedToNetwork: 라우터가 특정 네트워크에 연결되어 있는지 확인합니다
func (s *Service) isRouterConnectedToNetwork(router *compute.Router, networkName string) bool {
	if router.Network == "" {
		return false
	}

	networkURL := router.Network
	if !strings.Contains(networkURL, "/networks/") {
		return false
	}

	parts := strings.Split(networkURL, "/networks/")
	if len(parts) != 2 {
		return false
	}

	return parts[1] == networkName
}

// checkRouterForGateway: 라우터에 NAT 또는 인터넷 게이트웨이가 있는지 확인합니다
func (s *Service) checkRouterForGateway(router *compute.Router) *GatewayInfo {
	// Check for NAT gateway
	if natGateway := s.checkForNATGateway(router); natGateway != nil {
		return natGateway
	}

	// Check for Internet Gateway
	if internetGateway := s.checkForInternetGateway(router); internetGateway != nil {
		return internetGateway
	}

	return nil
}

// checkForNATGateway: 라우터에 NAT 게이트웨이가 있는지 확인합니다
func (s *Service) checkForNATGateway(router *compute.Router) *GatewayInfo {
	if len(router.Nats) == 0 {
		return nil
	}

	for _, nat := range router.Nats {
		if nat.NatIpAllocateOption == "AUTO_ONLY" || nat.NatIpAllocateOption == "MANUAL_ONLY" {
			return &GatewayInfo{
				Type: "NAT",
				Name: router.Name,
			}
		}
	}
	return nil
}

// checkForInternetGateway: 라우터에 인터넷 게이트웨이가 있는지 확인합니다
func (s *Service) checkForInternetGateway(router *compute.Router) *GatewayInfo {
	if router.Bgp == nil || router.Bgp.Asn <= 0 {
		return nil
	}

	return &GatewayInfo{
		Type: "Internet Gateway",
		Name: router.Name,
	}
}

// AWS Subnet Functions (existing implementations)

// listAWSSubnets: AWS 서브넷 목록을 조회합니다
func (s *Service) listAWSSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe subnets
	input := &ec2.DescribeSubnetsInput{}
	if req.SubnetID != "" {
		input.SubnetIds = []string{req.SubnetID}
	}
	if req.VPCID != "" {
		input.Filters = []ec2Types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{req.VPCID}},
		}
	}

	result, err := ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe subnets: %v", err), 502)
	}

	// Convert to DTOs
	subnets := make([]SubnetInfo, 0, len(result.Subnets))
	for _, subnet := range result.Subnets {
		subnetInfo := SubnetInfo{
			ID:               aws.ToString(subnet.SubnetId),
			Name:             s.getTagValue(subnet.Tags, "Name"),
			VPCID:            aws.ToString(subnet.VpcId),
			CIDRBlock:        aws.ToString(subnet.CidrBlock),
			AvailabilityZone: aws.ToString(subnet.AvailabilityZone),
			State:            string(subnet.State),
			IsPublic:         s.isSubnetPublic(ctx, ec2Client, aws.ToString(subnet.SubnetId)),
			Region:           req.Region,
			Tags:             s.convertTags(subnet.Tags),
		}
		subnets = append(subnets, subnetInfo)
	}

	return &ListSubnetsResponse{Subnets: subnets}, nil
}

// getAWSSubnet gets a specific AWS subnet
func (s *Service) getAWSSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Describe specific subnet
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{req.SubnetID},
	}

	result, err := ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe subnet: %v", err), 502)
	}

	if len(result.Subnets) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("subnet not found: %s", req.SubnetID), 404)
	}

	subnet := result.Subnets[0]
	subnetInfo := &SubnetInfo{
		ID:               aws.ToString(subnet.SubnetId),
		Name:             s.getTagValue(subnet.Tags, "Name"),
		VPCID:            aws.ToString(subnet.VpcId),
		CIDRBlock:        aws.ToString(subnet.CidrBlock),
		AvailabilityZone: aws.ToString(subnet.AvailabilityZone),
		State:            string(subnet.State),
		IsPublic:         s.isSubnetPublic(ctx, ec2Client, aws.ToString(subnet.SubnetId)),
		Region:           req.Region,
		Tags:             s.convertTags(subnet.Tags),
	}

	return subnetInfo, nil
}

// createAWSSubnet creates a new AWS subnet
func (s *Service) createAWSSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Create subnet
	input := &ec2.CreateSubnetInput{
		VpcId:            aws.String(req.VPCID),
		CidrBlock:        aws.String(req.CIDRBlock),
		AvailabilityZone: aws.String(req.AvailabilityZone),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeSubnet,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateSubnet(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create subnet: %v", err), 502)
	}

	subnetInfo := &SubnetInfo{
		ID:               aws.ToString(result.Subnet.SubnetId),
		Name:             req.Name,
		VPCID:            aws.ToString(result.Subnet.VpcId),
		CIDRBlock:        aws.ToString(result.Subnet.CidrBlock),
		AvailabilityZone: aws.ToString(result.Subnet.AvailabilityZone),
		State:            string(result.Subnet.State),
		IsPublic:         false, // Will be determined later
		Region:           req.Region,
		Tags:             req.Tags,
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
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

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"cidr_block":    subnetInfo.CIDRBlock,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		_ = s.eventPublisher.PublishSubnetEvent(ctx, credential.Provider, credentialID, req.VPCID, "created", subnetData)
	}

	return subnetInfo, nil
}

// updateAWSSubnet updates an AWS subnet
func (s *Service) updateAWSSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Update subnet tags if provided
	if req.Name != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{subnetID},
			Tags:      tags,
		})
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to update subnet tags: %v", err), 502)
		}
	}

	// Get updated subnet info
	subnetInfo, err := s.getAWSSubnet(ctx, credential, GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/vpcs/%s/subnets/%s", credential.Provider, subnetInfo.VPCID, subnetID),
		map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		}
		_ = s.eventPublisher.PublishSubnetEvent(ctx, credential.Provider, credentialID, subnetInfo.VPCID, "updated", subnetData)
	}

	return subnetInfo, nil
}

// deleteAWSSubnet deletes an AWS subnet
func (s *Service) deleteAWSSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EC2 client: %v", err), 502)
	}

	// Delete subnet
	_, err = ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(req.SubnetID),
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete subnet: %v", err), 502)
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/subnets/%s", credential.Provider, req.SubnetID),
		map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	// NATS 이벤트 발행 (VPCID는 req에 없으므로 빈 문자열로 처리)
	if s.eventPublisher != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		_ = s.eventPublisher.PublishSubnetEvent(ctx, credential.Provider, credentialID, "", "deleted", subnetData)
	}

	return nil
}

// Stub implementations for Azure and NCP

// listAzureSubnets lists Azure subnets (stub)
func (s *Service) listAzureSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	s.logger.Info("Azure subnet listing not yet implemented")
	return &ListSubnetsResponse{Subnets: []SubnetInfo{}}, nil
}

// getAzureSubnet gets Azure subnet (stub)
func (s *Service) getAzureSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info("Azure subnet retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure subnet retrieval not yet implemented", 501)
}

// createAzureSubnet creates Azure subnet (stub)
func (s *Service) createAzureSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info("Azure subnet creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure subnet creation not yet implemented", 501)
}

// updateAzureSubnet updates Azure subnet (stub)
func (s *Service) updateAzureSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	s.logger.Info("Azure subnet update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure subnet update not yet implemented", 501)
}

// deleteAzureSubnet deletes Azure subnet (stub)
func (s *Service) deleteAzureSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	s.logger.Info("Azure subnet deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure subnet deletion not yet implemented", 501)
}

// listNCPSubnets lists NCP subnets (stub)
func (s *Service) listNCPSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	s.logger.Info("NCP subnet listing not yet implemented")
	return &ListSubnetsResponse{Subnets: []SubnetInfo{}}, nil
}

// getNCPSubnet gets NCP subnet (stub)
func (s *Service) getNCPSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info("NCP subnet retrieval not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet retrieval not yet implemented", 501)
}

// createNCPSubnet creates NCP subnet (stub)
func (s *Service) createNCPSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	s.logger.Info("NCP subnet creation not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet creation not yet implemented", 501)
}

// updateNCPSubnet updates NCP subnet (stub)
func (s *Service) updateNCPSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	s.logger.Info("NCP subnet update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet update not yet implemented", 501)
}

// deleteNCPSubnet deletes NCP subnet (stub)
func (s *Service) deleteNCPSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	s.logger.Info("NCP subnet deletion not yet implemented")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP subnet deletion not yet implemented", 501)
}

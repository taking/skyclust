package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// GCP Network Functions
// All GCP-specific network operations are implemented in this file

func (s *Service) createGCPVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Marshal credential data for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
func (s *Service) createGCPVPCWithAdvanced(ctx context.Context, credential *domain.Credential, req CreateGCPVPCRequest, computeService *compute.Service) (*VPCInfo, error) {
	network := s.buildGCPNetworkObject(req)

	s.logNetworkConfiguration(ctx, network)

	operation, err := s.createGCPNetworkOperation(ctx, computeService, req.ProjectID, network)
	if err != nil {
		return nil, err
	}

	s.logOperationInitiated(ctx, req, operation)

	return s.buildVPCInfoFromRequest(req), nil
}
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
func (s *Service) logNetworkConfiguration(ctx context.Context, network *compute.Network) {
	routingMode := "REGIONAL"
	if network.RoutingConfig != nil {
		routingMode = network.RoutingConfig.RoutingMode
	}
	s.logger.Info(ctx, "Creating GCP network with configuration",
		domain.NewLogField("name", network.Name),
		domain.NewLogField("description", network.Description),
		domain.NewLogField("auto_create_subnetworks", network.AutoCreateSubnetworks),
		domain.NewLogField("routing_mode", routingMode),
		domain.NewLogField("mtu", network.Mtu))
}
func (s *Service) createGCPNetworkOperation(ctx context.Context, computeService *compute.Service, projectID string, network *compute.Network) (*compute.Operation, error) {
	operation, err := computeService.Networks.Insert(projectID, network).Context(ctx).Do()
	if err != nil {
		s.logger.Error(ctx, "Failed to create GCP network", err,
			domain.NewLogField("network_name", network.Name),
			domain.NewLogField("auto_create_subnetworks", network.AutoCreateSubnetworks))
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP network")
	}
	return operation, nil
}
func (s *Service) logOperationInitiated(ctx context.Context, req CreateGCPVPCRequest, operation *compute.Operation) {
	s.logger.Info(ctx, "GCP VPC creation initiated",
		domain.NewLogField("vpc_name", req.Name),
		domain.NewLogField("project_id", req.ProjectID),
		domain.NewLogField("operation_id", operation.Name))
}
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
func (s *Service) listGCPVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "list GCP networks")
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
			s.logger.Warn(ctx, "Failed to get firewall rules count",
				domain.NewLogField("network_name", networkItem.Name),
				domain.NewLogField("error", err))
			firewallCount = 0
		}

		// Get gateway information for this network
		gatewayInfo, err := s.getGatewayInfo(ctx, computeService, projectID, networkItem.Name)
		if err != nil {
			s.logger.Warn(ctx, "Failed to get gateway info",
				domain.NewLogField("network_name", networkItem.Name),
				domain.NewLogField("error", err))
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

	s.logger.Info(ctx, "GCP VPCs listed successfully",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("count", len(vpcs)))

	return &ListVPCsResponse{VPCs: vpcs}, nil
}
func (s *Service) getGCPVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf(ErrMsgInvalidVPCIDFormatGCP, req.VPCID), 400)
	}

	// Get specific network
	network, err := computeService.Networks.Get(projectID, networkName).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get GCP network")
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
		s.logger.Warn(ctx, "Failed to get firewall rules count",
			domain.NewLogField("network_name", network.Name),
			domain.NewLogField("error", err))
		firewallCount = 0
	}

	// Get gateway information for this network
	gatewayInfo, err := s.getGatewayInfo(ctx, computeService, projectID, network.Name)
	if err != nil {
		s.logger.Warn(ctx, "Failed to get gateway info",
			domain.NewLogField("network_name", network.Name),
			domain.NewLogField("error", err))
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

	s.logger.Info(ctx, "GCP VPC retrieved successfully",
		domain.NewLogField("vpc_name", network.Name),
		domain.NewLogField("project_id", projectID))

	return vpcInfo, nil
}
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
func (s *Service) deleteGCPVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return s.providerErrorConverter.ConvertGCPError(err, "get GCP VPC")
	}

	// Check and clean up dependencies
	s.logger.Info(ctx, "Starting VPC dependency cleanup",
		domain.NewLogField("vpc_name", networkName),
		domain.NewLogField("project_id", projectID))

	err = s.cleanupVPCResources(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn(ctx, "Failed to clean up VPC resources, proceeding with deletion",
			domain.NewLogField("vpc_name", networkName),
			domain.NewLogField("error", err))
		// Continue with deletion - GCP will handle validation
	} else {
		s.logger.Info(ctx, "VPC dependency cleanup completed successfully",
			domain.NewLogField("vpc_name", networkName))
	}

	// Delete the network
	s.logger.Info(ctx, "Initiating VPC deletion",
		domain.NewLogField("vpc_name", networkName),
		domain.NewLogField("project_id", projectID))

	operation, err := computeService.Networks.Delete(projectID, networkName).Context(ctx).Do()
	if err != nil {
		s.logger.Error(ctx, "Failed to delete GCP network", err,
			domain.NewLogField("vpc_name", networkName),
			domain.NewLogField("project_id", projectID))
		return s.providerErrorConverter.ConvertGCPError(err, "delete GCP network")
	}

	s.logger.Info(ctx, "GCP VPC deletion initiated successfully",
		domain.NewLogField("vpc_name", networkName),
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("operation_id", operation.Name),
		domain.NewLogField("operation_status", operation.Status))

	return nil
}
func (s *Service) cleanupVPCResources(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info(ctx, "Starting VPC resource cleanup",
		domain.NewLogField("vpc_name", networkName),
		domain.NewLogField("project_id", projectID))

	// 1. Delete firewall rules associated with this network
	err := s.deleteNetworkFirewallRules(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn(ctx, "Failed to delete firewall rules",
			domain.NewLogField("vpc_name", networkName),
			domain.NewLogField("error", err))
	}

	// 2. Delete subnets in this network
	err = s.deleteNetworkSubnets(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn(ctx, "Failed to delete subnets",
			domain.NewLogField("vpc_name", networkName),
			domain.NewLogField("error", err))
	}

	// 3. Check for instances using this network
	err = s.checkNetworkInstances(ctx, computeService, projectID, networkName)
	if err != nil {
		s.logger.Warn(ctx, "Found instances using this network",
			domain.NewLogField("vpc_name", networkName),
			domain.NewLogField("error", err))
		return domain.NewDomainError(domain.ErrCodeConflict, "cannot delete VPC: instances are still using this network", 409)
	}

	s.logger.Info(ctx, "VPC resource cleanup completed",
		domain.NewLogField("vpc_name", networkName))

	return nil
}
func (s *Service) deleteNetworkFirewallRules(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info(ctx, "Listing firewall rules for cleanup",
		domain.NewLogField("network", networkName),
		domain.NewLogField("project_id", projectID))

	// List all firewall rules
	firewalls, err := computeService.Firewalls.List(projectID).Context(ctx).Do()
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "list firewall rules")
	}

	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
	deletedCount := 0

	for _, firewall := range firewalls.Items {
		if firewall.Network == networkURL {
			s.logger.Info(ctx, "Deleting firewall rule",
				domain.NewLogField("firewall_name", firewall.Name),
				domain.NewLogField("network", networkName))

			_, err := computeService.Firewalls.Delete(projectID, firewall.Name).Context(ctx).Do()
			if err != nil {
				s.logger.Warn(ctx, "Failed to delete firewall rule",
					domain.NewLogField("firewall_name", firewall.Name),
					domain.NewLogField("error", err))
				// Continue with other firewall rules
			} else {
				deletedCount++
				s.logger.Info(ctx, "Firewall rule deleted successfully",
					domain.NewLogField("firewall_name", firewall.Name))
			}
		}
	}

	s.logger.Info(ctx, "Firewall rules cleanup completed",
		domain.NewLogField("network", networkName),
		domain.NewLogField("deleted_count", deletedCount))

	return nil
}
func (s *Service) deleteNetworkSubnets(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	s.logger.Info(ctx, "Listing subnets for cleanup",
		domain.NewLogField("network", networkName),
		domain.NewLogField("project_id", projectID))

	// List all subnets
	subnets, err := computeService.Subnetworks.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "list subnets")
	}

	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
	deletedCount := 0

	for _, subnetList := range subnets.Items {
		for _, subnet := range subnetList.Subnetworks {
			if subnet.Network == networkURL {
				s.logger.Info(ctx, "Deleting subnet",
					domain.NewLogField("subnet_name", subnet.Name),
					domain.NewLogField("region", subnet.Region),
					domain.NewLogField("network", networkName))

				_, err := computeService.Subnetworks.Delete(projectID, subnet.Region, subnet.Name).Context(ctx).Do()
				if err != nil {
					s.logger.Warn(ctx, "Failed to delete subnet",
						domain.NewLogField("subnet_name", subnet.Name),
						domain.NewLogField("region", subnet.Region),
						domain.NewLogField("error", err))
					// Continue with other subnets
				} else {
					deletedCount++
					s.logger.Info(ctx, "Subnet deleted successfully",
						domain.NewLogField("subnet_name", subnet.Name),
						domain.NewLogField("region", subnet.Region))
				}
			}
		}
	}

	s.logger.Info(ctx, "Subnets cleanup completed",
		domain.NewLogField("network", networkName),
		domain.NewLogField("deleted_count", deletedCount))

	return nil
}
func (s *Service) checkNetworkInstances(ctx context.Context, computeService *compute.Service, projectID, networkName string) error {
	// List all instances
	instances, err := computeService.Instances.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "list instances")
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
func (s *Service) updateGCPVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	s.logger.Info(ctx, "GCP VPC update not yet implemented")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "GCP VPC update not yet implemented", 501)
}
func (s *Service) listGCPSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "list GCP firewall rules")
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

	s.logger.Info(ctx, "GCP firewall rules listed successfully",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("count", len(securityGroups)))

	return &ListSecurityGroupsResponse{SecurityGroups: securityGroups}, nil
}
func (s *Service) getGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get GCP firewall rule")
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
func (s *Service) createGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	firewall := s.buildGCPFirewall(req)

	operation, err := computeService.Firewalls.Insert(projectID, firewall).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP firewall rule")
	}

	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall creation"); err != nil {
		return nil, err
	}

	sgInfo := s.buildSecurityGroupInfo(req, projectID)

	s.logger.Info(ctx, "GCP firewall rule created successfully",
		domain.NewLogField("firewall_name", req.Name),
		domain.NewLogField("project_id", projectID))

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
func (s *Service) updateGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, firewallName, region string) (*SecurityGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get current firewall rule")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "update GCP firewall rule")
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
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": firewallName,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.updated", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return sgInfo, nil
}
func (s *Service) deleteGCPSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	// Convert to JSON for GCP SDK
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	// Create GCP Compute service
	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
		return s.providerErrorConverter.ConvertGCPError(err, "delete GCP firewall rule")
	}

	// Wait for operation to complete
	if err := s.waitForGCPOperation(ctx, computeService, projectID, operation.Name, "firewall deletion"); err != nil {
		return err
	}

	s.logger.Info(ctx, "GCP firewall rule deleted successfully",
		domain.NewLogField("firewall_name", req.SecurityGroupID),
		domain.NewLogField("project_id", projectID))

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
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.deleted", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return nil
}
func (s *Service) removeGCPFirewallRule(ctx context.Context, credential *domain.Credential, req RemoveFirewallRuleRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	currentFirewall, err := computeService.Firewalls.Get(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get current firewall rule")
	}

	updatedFirewall := s.cloneFirewall(currentFirewall)
	updatedFirewall.Allowed = s.removePortsFromAllowed(currentFirewall.Allowed, req.Protocol, req.Ports)
	updatedFirewall.Denied = s.removePortsFromDenied(currentFirewall.Denied, req.Protocol, req.Ports)

	operation, err := computeService.Firewalls.Update(projectID, req.SecurityGroupID, updatedFirewall).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "update GCP firewall rule")
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
func (s *Service) addGCPFirewallRule(ctx context.Context, credential *domain.Credential, req AddFirewallRuleRequest) (*SecurityGroupInfo, error) {
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	currentFirewall, err := computeService.Firewalls.Get(projectID, req.SecurityGroupID).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get current firewall rule")
	}

	updatedFirewall := s.cloneFirewall(currentFirewall)
	updatedFirewall.Allowed = s.addPortsToAllowed(currentFirewall.Allowed, req.Protocol, req.Ports, req.Action)
	updatedFirewall.Denied = s.addPortsToDenied(currentFirewall.Denied, req.Protocol, req.Ports, req.Action)

	operation, err := computeService.Firewalls.Update(projectID, req.SecurityGroupID, updatedFirewall).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "update GCP firewall rule")
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
func (s *Service) setupGCPComputeService(ctx context.Context, credential *domain.Credential) (*compute.Service, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToDecryptCredential, err), 500)
	}

	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf(ErrMsgFailedToMarshalCredential, err), 500)
	}

	computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, "", s.providerErrorConverter.ConvertGCPError(err, "create GCP compute service")
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
func (s *Service) listGCPSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	// Create GCP Compute client
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	// List subnets
	subnets, err := computeService.Subnetworks.List(projectID, req.Region).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "list GCP subnets")
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
func (s *Service) getGCPSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(req.SubnetID)
	if subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", req.SubnetID), 400)
	}

	// Get subnet
	subnet, err := computeService.Subnetworks.Get(projectID, req.Region, subnetName).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get GCP subnet")
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
func (s *Service) createGCPSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract network name from VPC ID
	networkName := s.extractNetworkNameFromVPCID(req.VPCID)
	if networkName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf(ErrMsgInvalidVPCIDFormatGCP, req.VPCID), 400)
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "create GCP subnet")
	}

	s.logger.Info(ctx, "GCP subnet creation initiated",
		domain.NewLogField("subnet_name", req.Name),
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("operation_id", operation.Name))

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

	return subnetInfo, nil
}
func (s *Service) updateGCPSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	// Create GCP Compute client
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(subnetID)
	if subnetName == "" {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", subnetID), 400)
	}

	// Get current subnet
	currentSubnet, err := computeService.Subnetworks.Get(projectID, region, subnetName).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "get current subnet")
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
		return nil, s.providerErrorConverter.ConvertGCPError(err, "update GCP subnet")
	}

	s.logger.Info(ctx, "GCP subnet update initiated",
		domain.NewLogField("subnet_name", subnetName),
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("region", region),
		domain.NewLogField("operation_id", operation.Name))

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
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.updated", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return subnetInfo, nil
}
func (s *Service) deleteGCPSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	// Create GCP Compute client
	computeService, projectID, err := s.setupGCPComputeService(ctx, credential)
	if err != nil {
		return err
	}

	// Extract subnet name from subnet ID
	subnetName := s.extractSubnetNameFromSubnetID(req.SubnetID)
	if subnetName == "" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid subnet ID format: %s", req.SubnetID), 400)
	}

	// Delete subnet
	operation, err := computeService.Subnetworks.Delete(projectID, req.Region, subnetName).Context(ctx).Do()
	if err != nil {
		return s.providerErrorConverter.ConvertGCPError(err, "delete GCP subnet")
	}

	s.logger.Info(ctx, "GCP subnet deletion initiated",
		domain.NewLogField("subnet_name", subnetName),
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("operation_id", operation.Name))

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
func (s *Service) getFirewallRulesCount(ctx context.Context, computeService *compute.Service, projectID, networkName string) (int, error) {
	// List firewall rules for the specific network
	firewalls, err := computeService.Firewalls.List(projectID).Context(ctx).Do()
	if err != nil {
		return 0, s.providerErrorConverter.ConvertGCPError(err, "list firewall rules")
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
func (s *Service) getGatewayInfo(ctx context.Context, computeService *compute.Service, projectID, networkName string) (*GatewayInfo, error) {
	routers, err := s.listRouters(ctx, computeService, projectID)
	if err != nil {
		return nil, err
	}

	return s.findGatewayForNetwork(routers, networkName), nil
}
func (s *Service) listRouters(ctx context.Context, computeService *compute.Service, projectID string) (*compute.RouterAggregatedList, error) {
	routers, err := computeService.Routers.AggregatedList(projectID).Context(ctx).Do()
	if err != nil {
		return nil, s.providerErrorConverter.ConvertGCPError(err, "list routers")
	}
	return routers, nil
}
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
func (s *Service) checkForInternetGateway(router *compute.Router) *GatewayInfo {
	if router.Bgp == nil || router.Bgp.Asn <= 0 {
		return nil
	}

	return &GatewayInfo{
		Type: "Internet Gateway",
		Name: router.Name,
	}
}

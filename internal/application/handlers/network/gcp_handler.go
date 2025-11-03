package network

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GCPHandler handles GCP network resource HTTP requests
type GCPHandler struct {
	*handlers.BaseHandler
	networkService    *networkservice.Service
	credentialService domain.CredentialService
	logger            *zap.Logger
}

// NewGCPHandler creates a new GCP network handler
func NewGCPHandler(networkService *networkservice.Service, credentialService domain.CredentialService, logger *zap.Logger) *GCPHandler {
	return &GCPHandler{
		BaseHandler:       handlers.NewBaseHandler("gcp-network"),
		networkService:    networkService,
		credentialService: credentialService,
		logger:            logger,
	}
}

// VPC Handlers for GCP

// ListGCPVPCs handles VPC listing requests for GCP
func (h *GCPHandler) ListGCPVPCs(c *gin.Context) {
	// Get and validate credential
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	// Create handler request (no region needed for VPC - Global resource)
	handlerReq := ListVPCsRequest{
		Region: "", // VPC is Global, no region needed
	}
	serviceReq := ToServiceListVPCsRequest(handlerReq, credential.ID.String())

	// List VPCs
	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	handlerResp := FromServiceListVPCsResponse(vpcs)
	h.OK(c, handlerResp, "GCP VPCs retrieved successfully")
}

// ListGCPSubnets handles subnet listing requests for GCP
func (h *GCPHandler) ListGCPSubnets(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	// Get VPC ID from query parameter
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "list_subnets")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler request
	handlerReq := ListSubnetsRequest{
		VPCID:  vpcID,
		Region: region,
	}
	serviceReq := ToServiceListSubnetsRequest(handlerReq, credential.ID.String())

	// List subnets
	subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	handlerResp := FromServiceListSubnetsResponse(subnets)
	h.OK(c, handlerResp, "GCP subnets retrieved successfully")
}

// ListGCPSecurityGroups handles security group listing requests for GCP
func (h *GCPHandler) ListGCPSecurityGroups(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	// Get VPC ID from query parameter
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "list_security_groups")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler request
	handlerReq := ListSecurityGroupsRequest{
		VPCID:  vpcID,
		Region: region,
	}
	serviceReq := ToServiceListSecurityGroupsRequest(handlerReq, credential.ID.String())

	// List security groups
	securityGroups, err := h.networkService.ListSecurityGroups(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	handlerResp := FromServiceListSecurityGroupsResponse(securityGroups)
	h.OK(c, handlerResp, "GCP security groups retrieved successfully")
}

// GetGCPVPC handles VPC detail requests for GCP
func (h *GCPHandler) GetGCPVPC(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "get_vpc")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler request
	handlerReq := GetVPCRequest{
		VPCID:  vpcName, // Using vpcName as VPCID for service call
		Region: region,
	}
	serviceReq := ToServiceGetVPCRequest(handlerReq, credential.ID.String())

	// Get VPC
	vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	handlerResp := FromServiceVPCInfo(vpc)
	h.OK(c, handlerResp, "GCP VPC retrieved successfully")
}

// CreateGCPVPC handles VPC creation requests for GCP
func (h *GCPHandler) CreateGCPVPC(c *gin.Context) {
	// Parse request body
	var req CreateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	// Convert to service request
	serviceReq := ToServiceCreateVPCRequest(req, credential.ID.String())

	// Create VPC
	vpc, err := h.networkService.CreateVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	handlerResp := FromServiceVPCInfo(vpc)
	h.Created(c, handlerResp, "GCP VPC created successfully")
}

// UpdateGCPVPC handles VPC update requests for GCP
func (h *GCPHandler) UpdateGCPVPC(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "update_vpc")
		return
	}

	// Parse request body
	var req UpdateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Convert to service request
	serviceReq := ToServiceUpdateVPCRequest(req)

	// Update VPC
	vpc, err := h.networkService.UpdateVPC(c.Request.Context(), credential, serviceReq, vpcName, region)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	handlerResp := FromServiceVPCInfo(vpc)
	h.OK(c, handlerResp, "GCP VPC updated successfully")
}

// DeleteGCPVPC handles VPC deletion requests for GCP
func (h *GCPHandler) DeleteGCPVPC(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "delete_vpc")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler delete request
	handlerReq := DeleteVPCRequest{
		VPCID:  vpcName, // Using vpcName as VPCID for service call
		Region: region,
	}
	serviceReq := ToServiceDeleteVPCRequest(handlerReq, credential.ID.String())

	// Delete VPC
	err = h.networkService.DeleteVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	h.OK(c, nil, "GCP VPC deleted successfully")
}

// GetGCPSubnet handles subnet detail requests for GCP
func (h *GCPHandler) GetGCPSubnet(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "get_subnet")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler request
	handlerReq := GetSubnetRequest{
		SubnetID: subnetID,
		Region:   region,
	}
	serviceReq := ToServiceGetSubnetRequest(handlerReq, credential.ID.String())

	// Get subnet
	subnet, err := h.networkService.GetSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	handlerResp := FromServiceSubnetInfo(subnet)
	h.OK(c, handlerResp, "GCP subnet retrieved successfully")
}

// CreateGCPSubnet handles subnet creation requests for GCP
func (h *GCPHandler) CreateGCPSubnet(c *gin.Context) {
	// Parse request body
	var req CreateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	// Convert to service request
	serviceReq := ToServiceCreateSubnetRequest(req, credential.ID.String())

	// Create subnet
	subnet, err := h.networkService.CreateSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	handlerResp := FromServiceSubnetInfo(subnet)
	h.Created(c, handlerResp, "GCP subnet created successfully")
}

// UpdateGCPSubnet handles subnet update requests for GCP
func (h *GCPHandler) UpdateGCPSubnet(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "update_subnet")
		return
	}

	// Parse request body
	var req UpdateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Convert to service request
	serviceReq := ToServiceUpdateSubnetRequest(req)

	// Update subnet
	subnet, err := h.networkService.UpdateSubnet(c.Request.Context(), credential, serviceReq, subnetID, region)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	handlerResp := FromServiceSubnetInfo(subnet)
	h.OK(c, handlerResp, "GCP subnet updated successfully")
}

// DeleteGCPSubnet handles subnet deletion requests for GCP
func (h *GCPHandler) DeleteGCPSubnet(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "delete_subnet")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler delete request
	handlerReq := DeleteSubnetRequest{
		SubnetID: subnetID,
		Region:   region,
	}
	serviceReq := ToServiceDeleteSubnetRequest(handlerReq, credential.ID.String())

	// Delete subnet
	err = h.networkService.DeleteSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	h.OK(c, nil, "GCP subnet deleted successfully")
}

// GetGCPSecurityGroup handles security group detail requests for GCP
func (h *GCPHandler) GetGCPSecurityGroup(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_security_group")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "get_security_group")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler request
	handlerReq := GetSecurityGroupRequest{
		SecurityGroupID: securityGroupID,
		Region:          region,
	}
	serviceReq := ToServiceGetSecurityGroupRequest(handlerReq, credential.ID.String())

	// Get security group
	securityGroup, err := h.networkService.GetSecurityGroup(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_security_group")
		return
	}

	handlerResp := FromServiceSecurityGroupInfo(securityGroup)
	h.OK(c, handlerResp, "GCP security group retrieved successfully")
}

// CreateGCPSecurityGroup handles security group creation requests for GCP
func (h *GCPHandler) CreateGCPSecurityGroup(c *gin.Context) {
	// Parse request body
	var req CreateSecurityGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	// Convert to service request
	serviceReq := ToServiceCreateSecurityGroupRequest(req, credential.ID.String())

	// Create security group
	securityGroup, err := h.networkService.CreateSecurityGroup(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	handlerResp := FromServiceSecurityGroupInfo(securityGroup)
	h.Created(c, handlerResp, "GCP security group created successfully")
}

// UpdateGCPSecurityGroup handles security group update requests for GCP
func (h *GCPHandler) UpdateGCPSecurityGroup(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "update_security_group")
		return
	}

	// Parse request body
	var req UpdateSecurityGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Convert to service request
	serviceReq := ToServiceUpdateSecurityGroupRequest(req)

	// Update security group
	securityGroup, err := h.networkService.UpdateSecurityGroup(c.Request.Context(), credential, serviceReq, securityGroupID, region)
	if err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	handlerResp := FromServiceSecurityGroupInfo(securityGroup)
	h.OK(c, handlerResp, "GCP security group updated successfully")
}

// DeleteGCPSecurityGroup handles security group deletion requests for GCP
func (h *GCPHandler) DeleteGCPSecurityGroup(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "delete_security_group")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "delete_security_group")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

	// Create handler delete request
	handlerReq := DeleteSecurityGroupRequest{
		SecurityGroupID: securityGroupID,
		Region:          region,
	}
	serviceReq := ToServiceDeleteSecurityGroupRequest(handlerReq, credential.ID.String())

	// Delete security group
	err = h.networkService.DeleteSecurityGroup(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_security_group")
		return
	}

	h.OK(c, nil, "GCP security group deleted successfully")
}

// AddGCPSecurityGroupRule adds a rule to a GCP security group
func (h *GCPHandler) AddGCPSecurityGroupRule(c *gin.Context) {
	// Parse request
	var req AddSecurityGroupRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	// Convert to service request
	serviceReq := ToServiceAddSecurityGroupRuleRequest(req, credential.ID.String())

	// Add security group rule
	result, err := h.networkService.AddSecurityGroupRule(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "GCP security group rule added successfully")
	h.OK(c, handlerResp, "GCP security group rule added successfully")
}

// RemoveGCPSecurityGroupRule removes a rule from a GCP security group
func (h *GCPHandler) RemoveGCPSecurityGroupRule(c *gin.Context) {
	// Parse request
	var req RemoveSecurityGroupRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	// Convert to service request
	serviceReq := ToServiceRemoveSecurityGroupRuleRequest(req, credential.ID.String())

	// Remove security group rule
	result, err := h.networkService.RemoveSecurityGroupRule(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "GCP security group rule removed successfully")
	h.OK(c, handlerResp, "GCP security group rule removed successfully")
}

// UpdateGCPSecurityGroupRules updates all rules for a GCP security group
func (h *GCPHandler) UpdateGCPSecurityGroupRules(c *gin.Context) {
	// Parse request
	var req UpdateSecurityGroupRulesRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	// Get credential from query or body
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	// Convert to service request
	serviceReq := ToServiceUpdateSecurityGroupRulesRequest(req, credential.ID.String())

	// Update security group rules
	result, err := h.networkService.UpdateSecurityGroupRules(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "GCP security group rules updated successfully")
	h.OK(c, handlerResp, "GCP security group rules updated successfully")
}

// RemoveGCPFirewallRule handles removal of specific firewall rules for GCP
func (h *GCPHandler) RemoveGCPFirewallRule(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "remove_firewall_rule")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "remove_firewall_rule")
		return
	}

	// Parse request body
	var req RemoveFirewallRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "remove_firewall_rule")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "remove_firewall_rule")
		return
	}

	// Set request fields
	req.SecurityGroupID = securityGroupID
	req.Region = region

	// Convert to service request
	serviceReq := ToServiceRemoveFirewallRuleRequest(req, credential.ID.String())

	// Remove firewall rule
	result, err := h.networkService.RemoveFirewallRule(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "remove_firewall_rule")
		return
	}

	handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "GCP firewall rule removed successfully")
	h.OK(c, handlerResp, "GCP firewall rule removed successfully")
}

// AddGCPFirewallRule handles addition of specific firewall rules for GCP
func (h *GCPHandler) AddGCPFirewallRule(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "add_firewall_rule")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "add_firewall_rule")
		return
	}

	// Parse request body
	var req AddFirewallRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "add_firewall_rule")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "add_firewall_rule")
		return
	}

	// Set request fields
	req.SecurityGroupID = securityGroupID
	req.Region = region

	// Convert to service request
	serviceReq := ToServiceAddFirewallRuleRequest(req, credential.ID.String())

	// Add firewall rule
	result, err := h.networkService.AddFirewallRule(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "add_firewall_rule")
		return
	}

	handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "GCP firewall rule added successfully")
	h.OK(c, handlerResp, "GCP firewall rule added successfully")
}

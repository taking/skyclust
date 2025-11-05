package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GCPHandler handles GCP network resource HTTP requests
type GCPHandler struct {
	*BaseHandler
	logger *zap.Logger
}

// NewGCPHandler creates a new GCP network handler
func NewGCPHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
	logger interface{},
) *GCPHandler {
	var zapLogger *zap.Logger
	if l, ok := logger.(*zap.Logger); ok {
		zapLogger = l
	}
	return &GCPHandler{
		BaseHandler: NewBaseHandler(networkService, credentialService, domain.ProviderGCP, "gcp-network"),
		logger:      zapLogger,
	}
}

// ListVPCs handles VPC listing requests for GCP
func (h *GCPHandler) ListVPCs(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	serviceReq := networkservice.ListVPCsRequest{
		CredentialID: credential.ID.String(),
		Region:        "", // VPC is Global, no region needed
	}

	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	h.OK(c, vpcs, "GCP VPCs retrieved successfully")
}

// CreateVPC handles VPC creation requests for GCP
func (h *GCPHandler) CreateVPC(c *gin.Context) {
	var req networkservice.CreateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	req.CredentialID = credential.ID.String()

	vpc, err := h.networkService.CreateVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	h.Created(c, vpc, "GCP VPC created successfully")
}

// GetVPC handles VPC detail requests for GCP
func (h *GCPHandler) GetVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "get_vpc")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcName,
		Region:       region,
	}

	vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	h.OK(c, vpc, "GCP VPC retrieved successfully")
}

// UpdateVPC handles VPC update requests for GCP
func (h *GCPHandler) UpdateVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "update_vpc")
		return
	}

	var req networkservice.UpdateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	region := c.Query("region")

	vpc, err := h.networkService.UpdateVPC(c.Request.Context(), credential, req, vpcName, region)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	h.OK(c, vpc, "GCP VPC updated successfully")
}

// DeleteVPC handles VPC deletion requests for GCP
func (h *GCPHandler) DeleteVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	vpcName := c.Param("id")
	if vpcName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC name is required", 400), "delete_vpc")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.DeleteVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcName,
		Region:       region,
	}

	err = h.networkService.DeleteVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	h.OK(c, nil, "GCP VPC deleted successfully")
}

// ListSubnets handles subnet listing requests for GCP
func (h *GCPHandler) ListSubnets(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "list_subnets")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.ListSubnetsRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	h.OK(c, subnets, "GCP subnets retrieved successfully")
}

// CreateSubnet handles subnet creation requests for GCP
func (h *GCPHandler) CreateSubnet(c *gin.Context) {
	var req networkservice.CreateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	req.CredentialID = credential.ID.String()

	subnet, err := h.networkService.CreateSubnet(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	h.Created(c, subnet, "GCP subnet created successfully")
}

// GetSubnet handles subnet detail requests for GCP
func (h *GCPHandler) GetSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "get_subnet")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}

	subnet, err := h.networkService.GetSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	h.OK(c, subnet, "GCP subnet retrieved successfully")
}

// UpdateSubnet handles subnet update requests for GCP
func (h *GCPHandler) UpdateSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "update_subnet")
		return
	}

	var req networkservice.UpdateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	region := c.Query("region")

	subnet, err := h.networkService.UpdateSubnet(c.Request.Context(), credential, req, subnetID, region)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	h.OK(c, subnet, "GCP subnet updated successfully")
}

// DeleteSubnet handles subnet deletion requests for GCP
func (h *GCPHandler) DeleteSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	subnetID := c.Param("id")
	if subnetID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Subnet ID is required", 400), "delete_subnet")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.DeleteSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}

	err = h.networkService.DeleteSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	h.OK(c, nil, "GCP subnet deleted successfully")
}

// ListSecurityGroups handles security group listing requests for GCP
func (h *GCPHandler) ListSecurityGroups(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "list_security_groups")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.ListSecurityGroupsRequest{
		CredentialID:    credential.ID.String(),
		VPCID:           vpcID,
		Region:          region,
	}

	securityGroups, err := h.networkService.ListSecurityGroups(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	h.OK(c, securityGroups, "GCP security groups retrieved successfully")
}

// CreateSecurityGroup handles security group creation requests for GCP
func (h *GCPHandler) CreateSecurityGroup(c *gin.Context) {
	var req networkservice.CreateSecurityGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	req.CredentialID = credential.ID.String()

	securityGroup, err := h.networkService.CreateSecurityGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	h.Created(c, securityGroup, "GCP security group created successfully")
}

// GetSecurityGroup handles security group detail requests for GCP
func (h *GCPHandler) GetSecurityGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_security_group")
		return
	}

	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "get_security_group")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	}

	securityGroup, err := h.networkService.GetSecurityGroup(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_security_group")
		return
	}

	h.OK(c, securityGroup, "GCP security group retrieved successfully")
}

// UpdateSecurityGroup handles security group update requests for GCP
func (h *GCPHandler) UpdateSecurityGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "update_security_group")
		return
	}

	var req networkservice.UpdateSecurityGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	region := c.Query("region")

	securityGroup, err := h.networkService.UpdateSecurityGroup(c.Request.Context(), credential, req, securityGroupID, region)
	if err != nil {
		h.HandleError(c, err, "update_security_group")
		return
	}

	h.OK(c, securityGroup, "GCP security group updated successfully")
}

// DeleteSecurityGroup handles security group deletion requests for GCP
func (h *GCPHandler) DeleteSecurityGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "delete_security_group")
		return
	}

	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Security Group ID is required", 400), "delete_security_group")
		return
	}

	region := c.Query("region")

	serviceReq := networkservice.DeleteSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	}

	err = h.networkService.DeleteSecurityGroup(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_security_group")
		return
	}

	h.OK(c, nil, "GCP security group deleted successfully")
}

// AddSecurityGroupRule adds a rule to a GCP security group
func (h *GCPHandler) AddSecurityGroupRule(c *gin.Context) {
	var req networkservice.AddSecurityGroupRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	req.CredentialID = credential.ID.String()

	result, err := h.networkService.AddSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "add_security_group_rule")
		return
	}

	h.OK(c, result, "GCP security group rule added successfully")
}

// RemoveSecurityGroupRule removes a rule from a GCP security group
func (h *GCPHandler) RemoveSecurityGroupRule(c *gin.Context) {
	var req networkservice.RemoveSecurityGroupRuleRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	req.CredentialID = credential.ID.String()

	result, err := h.networkService.RemoveSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "remove_security_group_rule")
		return
	}

	h.OK(c, result, "GCP security group rule removed successfully")
}

// UpdateSecurityGroupRules updates all rules for a GCP security group
func (h *GCPHandler) UpdateSecurityGroupRules(c *gin.Context) {
	var req networkservice.UpdateSecurityGroupRulesRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	req.CredentialID = credential.ID.String()

	result, err := h.networkService.UpdateSecurityGroupRules(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "update_security_group_rules")
		return
	}

	h.OK(c, result, "GCP security group rules updated successfully")
}


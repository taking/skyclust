package providers

import (
	"fmt"
	"strings"

	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// AWSHandler handles AWS network resource HTTP requests
type AWSHandler struct {
	*BaseHandler
}

// NewAWSHandler creates a new AWS network handler
func NewAWSHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
) *AWSHandler {
	return &AWSHandler{
		BaseHandler: NewBaseHandler(networkService, credentialService, domain.ProviderAWS, "aws-network"),
	}
}

// ListVPCs handles VPC listing requests
func (h *AWSHandler) ListVPCs(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		return
	}

	serviceReq := networkservice.ListVPCsRequest{
		CredentialID: credential.ID.String(),
		Region:       region,
	}

	vpcs, err := h.GetNetworkService().ListVPCs(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	// Always include meta information for consistency (direct array: data[])
	page, limit := h.ParsePageLimitParams(c)
	total := int64(0)
	vpcList := []networkservice.VPCInfo{}
	if vpcs != nil {
		total = vpcs.Total
		vpcList = vpcs.VPCs
	}
	h.OKWithPagination(c, vpcList, "VPCs retrieved successfully", page, limit, total)
}

// CreateVPC handles VPC creation
func (h *AWSHandler) CreateVPC(c *gin.Context) {
	var req networkservice.CreateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	// credential_id는 body 또는 query parameter에서 가져올 수 있음
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = c.Query("credential_id")
	}

	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "create_vpc")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	req.CredentialID = credential.ID.String()

	ctx := h.EnrichContextWithRequestMetadata(c)
	vpc, err := h.GetNetworkService().CreateVPC(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	h.Created(c, vpc, "AWS VPC created successfully")
}

// GetVPC handles VPC detail requests
func (h *AWSHandler) GetVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	vpcID := c.Param("id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC ID is required", 400), "get_vpc")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		return
	}

	serviceReq := networkservice.GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	vpc, err := h.GetNetworkService().GetVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	h.OK(c, vpc, "AWS VPC retrieved successfully")
}

// UpdateVPC handles VPC update requests
func (h *AWSHandler) UpdateVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	vpcID := c.Param("id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC ID is required", 400), "update_vpc")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		return
	}

	var req networkservice.UpdateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	vpc, err := h.GetNetworkService().UpdateVPC(ctx, credential, req, vpcID, region)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	h.OK(c, vpc, "AWS VPC updated successfully")
}

// DeleteVPC handles VPC deletion requests
func (h *AWSHandler) DeleteVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	vpcID := c.Param("id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "VPC ID is required", 400), "delete_vpc")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		return
	}

	serviceReq := networkservice.DeleteVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	err = h.GetNetworkService().DeleteVPC(ctx, credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	h.OK(c, nil, "AWS VPC deleted successfully")
}

// ListSubnets handles subnet listing requests
func (h *AWSHandler) ListSubnets(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)

	if vpcID == "" || region == "" {
		return
	}

	serviceReq := networkservice.ListSubnetsRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	subnets, err := h.GetNetworkService().ListSubnets(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	// Always include meta information for consistency (direct array: data[])
	page, limit := h.ParsePageLimitParams(c)
	total := int64(0)
	subnetList := []networkservice.SubnetInfo{}
	if subnets != nil {
		total = subnets.Total
		subnetList = subnets.Subnets
	}
	h.OKWithPagination(c, subnetList, "Subnets retrieved successfully", page, limit, total)
}

// CreateSubnet handles subnet creation
func (h *AWSHandler) CreateSubnet(c *gin.Context) {
	var req networkservice.CreateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	// credential_id는 body 또는 query parameter에서 가져올 수 있음
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = c.Query("credential_id")
	}

	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "create_subnet")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	req.CredentialID = credential.ID.String()

	ctx := h.EnrichContextWithRequestMetadata(c)
	subnet, err := h.GetNetworkService().CreateSubnet(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	h.Created(c, subnet, "AWS subnet created successfully")
}

// GetSubnet handles subnet detail requests
func (h *AWSHandler) GetSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)

	if subnetID == "" || region == "" {
		return
	}

	serviceReq := networkservice.GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}

	subnet, err := h.GetNetworkService().GetSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	h.OK(c, subnet, "AWS subnet retrieved successfully")
}

// UpdateSubnet handles subnet update requests
func (h *AWSHandler) UpdateSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)

	if subnetID == "" || region == "" {
		return
	}

	var req networkservice.UpdateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	subnet, err := h.GetNetworkService().UpdateSubnet(ctx, credential, req, subnetID, region)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	h.OK(c, subnet, "AWS subnet updated successfully")
}

// DeleteSubnet handles subnet deletion requests
func (h *AWSHandler) DeleteSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)

	if subnetID == "" || region == "" {
		return
	}

	serviceReq := networkservice.DeleteSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	err = h.GetNetworkService().DeleteSubnet(ctx, credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	h.OK(c, nil, "AWS subnet deleted successfully")
}

// ListSecurityGroups handles security group listing requests
func (h *AWSHandler) ListSecurityGroups(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	vpcID := c.Query("vpc_id")
	region := c.Query("region")

	// Validate required parameters
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "list_security_groups")
		return
	}

	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "list_security_groups")
		return
	}

	// Validate region is not a VPC ID
	if strings.HasPrefix(region, "vpc-") {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid region: '%s' appears to be a VPC ID, not a region", region), 400), "list_security_groups")
		return
	}

	serviceReq := networkservice.ListSecurityGroupsRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	securityGroups, err := h.GetNetworkService().ListSecurityGroups(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	// Always include meta information for consistency (direct array: data[])
	page, limit := h.ParsePageLimitParams(c)
	total := int64(0)
	sgList := []networkservice.SecurityGroupInfo{}
	if securityGroups != nil {
		total = securityGroups.Total
		sgList = securityGroups.SecurityGroups
	}
	h.OKWithPagination(c, sgList, "Security groups retrieved successfully", page, limit, total)
}

// CreateSecurityGroup handles security group creation
func (h *AWSHandler) CreateSecurityGroup(c *gin.Context) {
	var req networkservice.CreateSecurityGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	// credential_id는 body 또는 query parameter에서 가져올 수 있음
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = c.Query("credential_id")
	}

	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "create_security_group")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	req.CredentialID = credential.ID.String()

	// Validate required fields
	if req.Name == "" || req.Description == "" || req.VPCID == "" || req.Region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "name, description, vpc_id, and region are required", 400), "create_security_group")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	securityGroup, err := h.GetNetworkService().CreateSecurityGroup(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	h.Created(c, securityGroup, "Security group created successfully")
}

// GetSecurityGroup handles security group detail requests
func (h *AWSHandler) GetSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "get_security_group")
}

// UpdateSecurityGroup handles security group update requests
func (h *AWSHandler) UpdateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "update_security_group")
}

// DeleteSecurityGroup handles security group deletion requests
func (h *AWSHandler) DeleteSecurityGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
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
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "delete_security_group")
		return
	}

	serviceReq := networkservice.DeleteSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	err = h.GetNetworkService().DeleteSecurityGroup(ctx, credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_security_group")
		return
	}

	h.OK(c, nil, "AWS security group deleted successfully")
}

// AddSecurityGroupRule handles adding a security group rule
func (h *AWSHandler) AddSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "add_security_group_rule")
}

// RemoveSecurityGroupRule handles removing a security group rule
func (h *AWSHandler) RemoveSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "remove_security_group_rule")
}

// UpdateSecurityGroupRules handles updating all security group rules
func (h *AWSHandler) UpdateSecurityGroupRules(c *gin.Context) {
	h.NotImplemented(c, "update_security_group_rules")
}

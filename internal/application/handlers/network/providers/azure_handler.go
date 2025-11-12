package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// AzureHandler: Azure 네트워크 리소스 HTTP 요청을 처리하는 핸들러
type AzureHandler struct {
	*BaseHandler
}

// NewAzureHandler: 새로운 Azure 네트워크 핸들러를 생성합니다
func NewAzureHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
) *AzureHandler {
	return &AzureHandler{
		BaseHandler: NewBaseHandler(networkService, credentialService, domain.ProviderAzure, "azure-network"),
	}
}

// ListVPCs: Virtual Network 목록 조회를 처리합니다
func (h *AzureHandler) ListVPCs(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	resourceGroup := c.Query("resource_group")

	serviceReq := networkservice.ListVPCsRequest{
		CredentialID:  credential.ID.String(),
		Region:        region,
		ResourceGroup: resourceGroup,
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
	h.OKWithPagination(c, vpcList, "Virtual Networks retrieved successfully", page, limit, total)
}

// CreateVPC handles Virtual Network creation
func (h *AzureHandler) CreateVPC(c *gin.Context) {
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

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAzure)
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

	h.Created(c, vpc, "Virtual Network created successfully")
}

// GetVPC handles Virtual Network detail requests
func (h *AzureHandler) GetVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	if vpcID == "" || region == "" {
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

	h.OK(c, vpc, "Virtual Network retrieved successfully")
}

// UpdateVPC handles Virtual Network update requests
func (h *AzureHandler) UpdateVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	if vpcID == "" || region == "" {
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

	h.OK(c, vpc, "Virtual Network updated successfully")
}

// DeleteVPC handles Virtual Network deletion requests
func (h *AzureHandler) DeleteVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	if vpcID == "" || region == "" {
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

	h.OK(c, nil, "Virtual Network deleted successfully")
}

// ListSubnets handles subnet listing requests
func (h *AzureHandler) ListSubnets(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	resourceGroup := c.Query("resource_group")

	if vpcID == "" || region == "" {
		return
	}

	serviceReq := networkservice.ListSubnetsRequest{
		CredentialID:  credential.ID.String(),
		VPCID:         vpcID,
		Region:        region,
		ResourceGroup: resourceGroup,
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
func (h *AzureHandler) CreateSubnet(c *gin.Context) {
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

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAzure)
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

	h.Created(c, subnet, "Subnet created successfully")
}

// GetSubnet handles subnet detail requests
func (h *AzureHandler) GetSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

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

	h.OK(c, subnet, "Subnet retrieved successfully")
}

// UpdateSubnet handles subnet update requests
func (h *AzureHandler) UpdateSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

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

	h.OK(c, subnet, "Subnet updated successfully")
}

// DeleteSubnet handles subnet deletion requests
func (h *AzureHandler) DeleteSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

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

	h.OK(c, nil, "Subnet deleted successfully")
}

// ListSecurityGroups: Network Security Group 목록 조회를 처리합니다
func (h *AzureHandler) ListSecurityGroups(c *gin.Context) {
	h.NotImplemented(c, "list_security_groups")
}

// CreateSecurityGroup: Network Security Group 생성을 처리합니다
func (h *AzureHandler) CreateSecurityGroup(c *gin.Context) {
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

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAzure)
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

	h.Created(c, securityGroup, "Azure Network Security Group created successfully")
}

// GetSecurityGroup: Network Security Group 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "get_security_group")
}

// UpdateSecurityGroup: Network Security Group 업데이트를 처리합니다
func (h *AzureHandler) UpdateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "update_security_group")
}

// DeleteSecurityGroup: Network Security Group 삭제를 처리합니다
func (h *AzureHandler) DeleteSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_security_group")
}

// AddSecurityGroupRule: Network Security Group 규칙 추가를 처리합니다
func (h *AzureHandler) AddSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "add_security_group_rule")
}

// RemoveSecurityGroupRule: Network Security Group 규칙 제거를 처리합니다
func (h *AzureHandler) RemoveSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "remove_security_group_rule")
}

// UpdateSecurityGroupRules: Network Security Group 규칙 전체 업데이트를 처리합니다
func (h *AzureHandler) UpdateSecurityGroupRules(c *gin.Context) {
	h.NotImplemented(c, "update_security_group_rules")
}

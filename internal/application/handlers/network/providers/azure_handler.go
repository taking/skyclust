package providers

import (
	"fmt"

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
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	region := h.parseRegion(c)
	if region == "" {
		region = c.Query("location") // Azure uses "location" instead of "region"
	}

	serviceReq := networkservice.ListVPCsRequest{
		CredentialID: credential.ID.String(),
		Region:       region,
	}

	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	h.OK(c, vpcs, "Virtual Networks retrieved successfully")
}

// CreateVPC: Virtual Network 생성을 처리합니다
func (h *AzureHandler) CreateVPC(c *gin.Context) {
	var req networkservice.CreateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	vpc, err := h.networkService.CreateVPC(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_vpc")
		return
	}

	h.Created(c, vpc, "Virtual Network creation initiated")
}

// GetVPC: Virtual Network 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)

	if vpcID == "" || region == "" {
		return
	}

	serviceReq := networkservice.GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_vpc")
		return
	}

	h.OK(c, vpc, "Virtual Network retrieved successfully")
}

// UpdateVPC: Virtual Network 업데이트를 처리합니다
func (h *AzureHandler) UpdateVPC(c *gin.Context) {
	var req networkservice.UpdateVPCRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	credentialID := c.Query("credential_id")
	if credentialID == "" {
		credentialID = c.Param("credential_id")
	}
	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "update_vpc")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)

	if vpcID == "" || region == "" {
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	vpc, err := h.networkService.UpdateVPC(ctx, credential, req, vpcID, region)
	if err != nil {
		h.HandleError(c, err, "update_vpc")
		return
	}

	h.OK(c, vpc, "Virtual Network updated successfully")
}

// DeleteVPC: Virtual Network 삭제를 처리합니다
func (h *AzureHandler) DeleteVPC(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)

	if vpcID == "" || region == "" {
		return
	}

	serviceReq := networkservice.DeleteVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}

	err = h.networkService.DeleteVPC(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_vpc")
		return
	}

	h.OK(c, nil, "Virtual Network deletion initiated")
}

// ListSubnets: 서브넷 목록 조회를 처리합니다
func (h *AzureHandler) ListSubnets(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
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

	subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	h.OK(c, subnets, "Subnets retrieved successfully")
}

// CreateSubnet: 서브넷 생성을 처리합니다
func (h *AzureHandler) CreateSubnet(c *gin.Context) {
	var req networkservice.CreateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	subnet, err := h.networkService.CreateSubnet(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_subnet")
		return
	}

	h.Created(c, subnet, "Subnet creation initiated")
}

// GetSubnet: 서브넷 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
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

	subnet, err := h.networkService.GetSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "get_subnet")
		return
	}

	h.OK(c, subnet, "Subnet retrieved successfully")
}

// UpdateSubnet: 서브넷 업데이트를 처리합니다
func (h *AzureHandler) UpdateSubnet(c *gin.Context) {
	var req networkservice.UpdateSubnetRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	subnetID := h.parseSubnetID(c)
	region := h.parseRegion(c)

	if subnetID == "" || region == "" {
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	subnet, err := h.networkService.UpdateSubnet(ctx, credential, req, subnetID, region)
	if err != nil {
		h.HandleError(c, err, "update_subnet")
		return
	}

	h.OK(c, subnet, "Subnet updated successfully")
}

// DeleteSubnet: 서브넷 삭제를 처리합니다
func (h *AzureHandler) DeleteSubnet(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
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

	err = h.networkService.DeleteSubnet(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "delete_subnet")
		return
	}

	h.OK(c, nil, "Subnet deletion initiated")
}

// ListSecurityGroups: Network Security Group 목록 조회를 처리합니다
func (h *AzureHandler) ListSecurityGroups(c *gin.Context) {
	h.NotImplemented(c, "list_security_groups")
}

// CreateSecurityGroup: Network Security Group 생성을 처리합니다
func (h *AzureHandler) CreateSecurityGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "create_security_group")
		return
	}

	var req networkservice.CreateSecurityGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid request body: %v", err), 400), "create_security_group")
		return
	}

	// credential_id는 body 또는 query에서 가져올 수 있음
	if req.CredentialID == "" {
		req.CredentialID = credential.ID.String()
	}

	// Validate required fields
	if req.Name == "" || req.Description == "" || req.VPCID == "" || req.Region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("name, description, vpc_id, and region are required"), 400), "create_security_group")
		return
	}

	ctx := c.Request.Context()
	securityGroup, err := h.networkService.CreateSecurityGroup(ctx, credential, req)
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

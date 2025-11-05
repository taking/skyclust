package providers

import (
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
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAWS)
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
		Region:        region,
	}

	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_vpcs")
		return
	}

	h.OK(c, vpcs, "VPCs retrieved successfully")
}

// CreateVPC handles VPC creation
func (h *AWSHandler) CreateVPC(c *gin.Context) {
	h.NotImplemented(c, "create_vpc")
}

// GetVPC handles VPC detail requests
func (h *AWSHandler) GetVPC(c *gin.Context) {
	h.NotImplemented(c, "get_vpc")
}

// UpdateVPC handles VPC update requests
func (h *AWSHandler) UpdateVPC(c *gin.Context) {
	h.NotImplemented(c, "update_vpc")
}

// DeleteVPC handles VPC deletion requests
func (h *AWSHandler) DeleteVPC(c *gin.Context) {
	h.NotImplemented(c, "delete_vpc")
}

// ListSubnets handles subnet listing requests
func (h *AWSHandler) ListSubnets(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAWS)
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
		VPCID:         vpcID,
		Region:        region,
	}

	subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, serviceReq)
	if err != nil {
		h.HandleError(c, err, "list_subnets")
		return
	}

	h.OK(c, subnets, "Subnets retrieved successfully")
}

// CreateSubnet handles subnet creation
func (h *AWSHandler) CreateSubnet(c *gin.Context) {
	h.NotImplemented(c, "create_subnet")
}

// GetSubnet handles subnet detail requests
func (h *AWSHandler) GetSubnet(c *gin.Context) {
	h.NotImplemented(c, "get_subnet")
}

// UpdateSubnet handles subnet update requests
func (h *AWSHandler) UpdateSubnet(c *gin.Context) {
	h.NotImplemented(c, "update_subnet")
}

// DeleteSubnet handles subnet deletion requests
func (h *AWSHandler) DeleteSubnet(c *gin.Context) {
	h.NotImplemented(c, "delete_subnet")
}

// ListSecurityGroups handles security group listing requests
func (h *AWSHandler) ListSecurityGroups(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "list_security_groups")
		return
	}

	vpcID := h.parseVPCID(c)
	region := h.parseRegion(c)

	if vpcID == "" || region == "" {
		return
	}

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

	h.OK(c, securityGroups, "Security groups retrieved successfully")
}

// CreateSecurityGroup handles security group creation
func (h *AWSHandler) CreateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "create_security_group")
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
	h.NotImplemented(c, "delete_security_group")
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


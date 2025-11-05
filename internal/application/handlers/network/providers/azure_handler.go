package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// AzureHandler handles Azure network resource HTTP requests
type AzureHandler struct {
	*BaseHandler
}

// NewAzureHandler creates a new Azure network handler
func NewAzureHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
) *AzureHandler {
	return &AzureHandler{
		BaseHandler: NewBaseHandler(networkService, credentialService, domain.ProviderAzure, "azure-network"),
	}
}

// ListVPCs handles VPC listing requests
func (h *AzureHandler) ListVPCs(c *gin.Context) {
	h.NotImplemented(c, "list_vpcs")
}

// CreateVPC handles VPC creation
func (h *AzureHandler) CreateVPC(c *gin.Context) {
	h.NotImplemented(c, "create_vpc")
}

// GetVPC handles VPC detail requests
func (h *AzureHandler) GetVPC(c *gin.Context) {
	h.NotImplemented(c, "get_vpc")
}

// UpdateVPC handles VPC update requests
func (h *AzureHandler) UpdateVPC(c *gin.Context) {
	h.NotImplemented(c, "update_vpc")
}

// DeleteVPC handles VPC deletion requests
func (h *AzureHandler) DeleteVPC(c *gin.Context) {
	h.NotImplemented(c, "delete_vpc")
}

// ListSubnets handles subnet listing requests
func (h *AzureHandler) ListSubnets(c *gin.Context) {
	h.NotImplemented(c, "list_subnets")
}

// CreateSubnet handles subnet creation
func (h *AzureHandler) CreateSubnet(c *gin.Context) {
	h.NotImplemented(c, "create_subnet")
}

// GetSubnet handles subnet detail requests
func (h *AzureHandler) GetSubnet(c *gin.Context) {
	h.NotImplemented(c, "get_subnet")
}

// UpdateSubnet handles subnet update requests
func (h *AzureHandler) UpdateSubnet(c *gin.Context) {
	h.NotImplemented(c, "update_subnet")
}

// DeleteSubnet handles subnet deletion requests
func (h *AzureHandler) DeleteSubnet(c *gin.Context) {
	h.NotImplemented(c, "delete_subnet")
}

// ListSecurityGroups handles security group listing requests
func (h *AzureHandler) ListSecurityGroups(c *gin.Context) {
	h.NotImplemented(c, "list_security_groups")
}

// CreateSecurityGroup handles security group creation
func (h *AzureHandler) CreateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "create_security_group")
}

// GetSecurityGroup handles security group detail requests
func (h *AzureHandler) GetSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "get_security_group")
}

// UpdateSecurityGroup handles security group update requests
func (h *AzureHandler) UpdateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "update_security_group")
}

// DeleteSecurityGroup handles security group deletion requests
func (h *AzureHandler) DeleteSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_security_group")
}

// AddSecurityGroupRule handles adding a security group rule
func (h *AzureHandler) AddSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "add_security_group_rule")
}

// RemoveSecurityGroupRule handles removing a security group rule
func (h *AzureHandler) RemoveSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "remove_security_group_rule")
}

// UpdateSecurityGroupRules handles updating all security group rules
func (h *AzureHandler) UpdateSecurityGroupRules(c *gin.Context) {
	h.NotImplemented(c, "update_security_group_rules")
}


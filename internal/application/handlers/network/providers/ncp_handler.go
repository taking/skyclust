package providers

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// NCPHandler handles NCP network resource HTTP requests
type NCPHandler struct {
	*BaseHandler
}

// NewNCPHandler creates a new NCP network handler
func NewNCPHandler(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
) *NCPHandler {
	return &NCPHandler{
		BaseHandler: NewBaseHandler(networkService, credentialService, domain.ProviderNCP, "ncp-network"),
	}
}

// ListVPCs handles VPC listing requests
func (h *NCPHandler) ListVPCs(c *gin.Context) {
	h.NotImplemented(c, "list_vpcs")
}

// CreateVPC handles VPC creation
func (h *NCPHandler) CreateVPC(c *gin.Context) {
	h.NotImplemented(c, "create_vpc")
}

// GetVPC handles VPC detail requests
func (h *NCPHandler) GetVPC(c *gin.Context) {
	h.NotImplemented(c, "get_vpc")
}

// UpdateVPC handles VPC update requests
func (h *NCPHandler) UpdateVPC(c *gin.Context) {
	h.NotImplemented(c, "update_vpc")
}

// DeleteVPC handles VPC deletion requests
func (h *NCPHandler) DeleteVPC(c *gin.Context) {
	h.NotImplemented(c, "delete_vpc")
}

// ListSubnets handles subnet listing requests
func (h *NCPHandler) ListSubnets(c *gin.Context) {
	h.NotImplemented(c, "list_subnets")
}

// CreateSubnet handles subnet creation
func (h *NCPHandler) CreateSubnet(c *gin.Context) {
	h.NotImplemented(c, "create_subnet")
}

// GetSubnet handles subnet detail requests
func (h *NCPHandler) GetSubnet(c *gin.Context) {
	h.NotImplemented(c, "get_subnet")
}

// UpdateSubnet handles subnet update requests
func (h *NCPHandler) UpdateSubnet(c *gin.Context) {
	h.NotImplemented(c, "update_subnet")
}

// DeleteSubnet handles subnet deletion requests
func (h *NCPHandler) DeleteSubnet(c *gin.Context) {
	h.NotImplemented(c, "delete_subnet")
}

// ListSecurityGroups handles security group listing requests
func (h *NCPHandler) ListSecurityGroups(c *gin.Context) {
	h.NotImplemented(c, "list_security_groups")
}

// CreateSecurityGroup handles security group creation
func (h *NCPHandler) CreateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "create_security_group")
}

// GetSecurityGroup handles security group detail requests
func (h *NCPHandler) GetSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "get_security_group")
}

// UpdateSecurityGroup handles security group update requests
func (h *NCPHandler) UpdateSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "update_security_group")
}

// DeleteSecurityGroup handles security group deletion requests
func (h *NCPHandler) DeleteSecurityGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_security_group")
}

// AddSecurityGroupRule handles adding a security group rule
func (h *NCPHandler) AddSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "add_security_group_rule")
}

// RemoveSecurityGroupRule handles removing a security group rule
func (h *NCPHandler) RemoveSecurityGroupRule(c *gin.Context) {
	h.NotImplemented(c, "remove_security_group_rule")
}

// UpdateSecurityGroupRules handles updating all security group rules
func (h *NCPHandler) UpdateSecurityGroupRules(c *gin.Context) {
	h.NotImplemented(c, "update_security_group_rules")
}

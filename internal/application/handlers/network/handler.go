package network

import (
	"skyclust/internal/application/dto"
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles network resource HTTP requests
type Handler struct {
	*handlers.BaseHandler
	networkService    *service.NetworkService
	credentialService domain.CredentialService
	logger            *zap.Logger
	provider          string
}

// NewHandler creates a new network handler
func NewHandler(networkService *service.NetworkService, credentialService domain.CredentialService, logger *zap.Logger, provider string) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("network"),
		networkService:    networkService,
		credentialService: credentialService,
		logger:            logger,
		provider:          provider,
	}
}

// VPC Handlers

// ListVPCs handles VPC listing requests
func (h *Handler) ListVPCs(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.ListVPCsRequest{
		CredentialID: credentialID,
		Region:       region,
	}

	// List VPCs
	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list VPCs")
		responses.InternalServerError(c, "Failed to list VPCs: "+err.Error())
		return
	}

	responses.OK(c, vpcs, "VPCs retrieved successfully")
}

// ListSubnets handles subnet listing requests
func (h *Handler) ListSubnets(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get VPC ID from query parameter
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		responses.BadRequest(c, "vpc_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.ListSubnetsRequest{
		CredentialID: credentialID,
		VPCID:        vpcID,
		Region:       region,
	}

	// List subnets
	subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list subnets")
		responses.InternalServerError(c, "Failed to list subnets: "+err.Error())
		return
	}

	responses.OK(c, subnets, "Subnets retrieved successfully")
}

// ListSecurityGroups handles security group listing requests
func (h *Handler) ListSecurityGroups(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get VPC ID from query parameter
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		responses.BadRequest(c, "vpc_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.ListSecurityGroupsRequest{
		CredentialID: credentialID,
		VPCID:        vpcID,
		Region:       region,
	}

	// List security groups
	securityGroups, err := h.networkService.ListSecurityGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list security groups")
		responses.InternalServerError(c, "Failed to list security groups: "+err.Error())
		return
	}

	responses.OK(c, securityGroups, "Security groups retrieved successfully")
}

// GetVPC handles VPC detail requests
func (h *Handler) GetVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC ID from path parameter
	vpcID := c.Param("id")
	if vpcID == "" {
		responses.BadRequest(c, "VPC ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.GetVPCRequest{
		CredentialID: credentialID,
		VPCID:        vpcID,
		Region:       region,
	}

	// Get VPC
	vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get VPC")
		responses.InternalServerError(c, "Failed to get VPC: "+err.Error())
		return
	}

	responses.OK(c, vpc, "VPC retrieved successfully")
}

// CreateVPC handles VPC creation requests
func (h *Handler) CreateVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Parse request body
	var req dto.CreateVPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create VPC
	vpc, err := h.networkService.CreateVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create VPC")
		responses.InternalServerError(c, "Failed to create VPC: "+err.Error())
		return
	}

	responses.Created(c, vpc, "VPC created successfully")
}

// UpdateVPC handles VPC update requests
func (h *Handler) UpdateVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC ID from path parameter
	vpcID := c.Param("id")
	if vpcID == "" {
		responses.BadRequest(c, "VPC ID is required")
		return
	}

	// Parse request body
	var req dto.UpdateVPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Update VPC
	vpc, err := h.networkService.UpdateVPC(c.Request.Context(), credential, req, vpcID, region)
	if err != nil {
		h.LogError(c, err, "Failed to update VPC")
		responses.InternalServerError(c, "Failed to update VPC: "+err.Error())
		return
	}

	responses.OK(c, vpc, "VPC updated successfully")
}

// DeleteVPC handles VPC deletion requests
func (h *Handler) DeleteVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC ID from path parameter
	vpcID := c.Param("id")
	if vpcID == "" {
		responses.BadRequest(c, "VPC ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create delete request
	req := dto.DeleteVPCRequest{
		CredentialID: credentialID,
		VPCID:        vpcID,
		Region:       region,
	}

	// Delete VPC
	err = h.networkService.DeleteVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to delete VPC")
		responses.InternalServerError(c, "Failed to delete VPC: "+err.Error())
		return
	}

	responses.OK(c, nil, "VPC deleted successfully")
}

// GetSubnet handles subnet detail requests
func (h *Handler) GetSubnet(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		responses.BadRequest(c, "Subnet ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.GetSubnetRequest{
		CredentialID: credentialID,
		SubnetID:     subnetID,
		Region:       region,
	}

	// Get subnet
	subnet, err := h.networkService.GetSubnet(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get subnet")
		responses.InternalServerError(c, "Failed to get subnet: "+err.Error())
		return
	}

	responses.OK(c, subnet, "Subnet retrieved successfully")
}

// CreateSubnet handles subnet creation requests
func (h *Handler) CreateSubnet(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Parse request body
	var req dto.CreateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create subnet
	subnet, err := h.networkService.CreateSubnet(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create subnet")
		responses.InternalServerError(c, "Failed to create subnet: "+err.Error())
		return
	}

	responses.Created(c, subnet, "Subnet created successfully")
}

// UpdateSubnet handles subnet update requests
func (h *Handler) UpdateSubnet(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		responses.BadRequest(c, "Subnet ID is required")
		return
	}

	// Parse request body
	var req dto.UpdateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Update subnet
	subnet, err := h.networkService.UpdateSubnet(c.Request.Context(), credential, req, subnetID, region)
	if err != nil {
		h.LogError(c, err, "Failed to update subnet")
		responses.InternalServerError(c, "Failed to update subnet: "+err.Error())
		return
	}

	responses.OK(c, subnet, "Subnet updated successfully")
}

// DeleteSubnet handles subnet deletion requests
func (h *Handler) DeleteSubnet(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get subnet ID from path parameter
	subnetID := c.Param("id")
	if subnetID == "" {
		responses.BadRequest(c, "Subnet ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create delete request
	req := dto.DeleteSubnetRequest{
		CredentialID: credentialID,
		SubnetID:     subnetID,
		Region:       region,
	}

	// Delete subnet
	err = h.networkService.DeleteSubnet(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to delete subnet")
		responses.InternalServerError(c, "Failed to delete subnet: "+err.Error())
		return
	}

	responses.OK(c, nil, "Subnet deleted successfully")
}

// GetSecurityGroup handles security group detail requests
func (h *Handler) GetSecurityGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		responses.BadRequest(c, "Security Group ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create request
	req := dto.GetSecurityGroupRequest{
		CredentialID:    credentialID,
		SecurityGroupID: securityGroupID,
		Region:          region,
	}

	// Get security group
	securityGroup, err := h.networkService.GetSecurityGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get security group")
		responses.InternalServerError(c, "Failed to get security group: "+err.Error())
		return
	}

	responses.OK(c, securityGroup, "Security group retrieved successfully")
}

// CreateSecurityGroup handles security group creation requests
func (h *Handler) CreateSecurityGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Parse request body
	var req dto.CreateSecurityGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create security group
	securityGroup, err := h.networkService.CreateSecurityGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create security group")
		responses.InternalServerError(c, "Failed to create security group: "+err.Error())
		return
	}

	responses.Created(c, securityGroup, "Security group created successfully")
}

// UpdateSecurityGroup handles security group update requests
func (h *Handler) UpdateSecurityGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		responses.BadRequest(c, "Security Group ID is required")
		return
	}

	// Parse request body
	var req dto.UpdateSecurityGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Update security group
	securityGroup, err := h.networkService.UpdateSecurityGroup(c.Request.Context(), credential, req, securityGroupID, region)
	if err != nil {
		h.LogError(c, err, "Failed to update security group")
		responses.InternalServerError(c, "Failed to update security group: "+err.Error())
		return
	}

	responses.OK(c, securityGroup, "Security group updated successfully")
}

// DeleteSecurityGroup handles security group deletion requests
func (h *Handler) DeleteSecurityGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get security group ID from path parameter
	securityGroupID := c.Param("id")
	if securityGroupID == "" {
		responses.BadRequest(c, "Security Group ID is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter
	region := c.Query("region")
	if region == "" {
		responses.BadRequest(c, "region is required")
		return
	}

	// Parse credential ID to UUID
	credentialUUID, err := uuid.Parse(credentialID)
	if err != nil {
		responses.BadRequest(c, "invalid credential ID format")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialUUID)
	if err != nil {
		responses.NotFound(c, "credential not found")
		return
	}

	// Create delete request
	req := dto.DeleteSecurityGroupRequest{
		CredentialID:    credentialID,
		SecurityGroupID: securityGroupID,
		Region:          region,
	}

	// Delete security group
	err = h.networkService.DeleteSecurityGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to delete security group")
		responses.InternalServerError(c, "Failed to delete security group: "+err.Error())
		return
	}

	responses.OK(c, nil, "Security group deleted successfully")
}

// AddSecurityGroupRule adds a rule to a security group
func (h *Handler) AddSecurityGroupRule(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid user ID")
		return
	}

	// Parse request
	var req dto.AddSecurityGroupRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request")
		return
	}

	// Parse credential ID
	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Add security group rule
	result, err := h.networkService.AddSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to add security group rule")
		responses.InternalServerError(c, "Failed to add security group rule: "+err.Error())
		return
	}

	responses.OK(c, result, "Security group rule added successfully")
}

// RemoveSecurityGroupRule removes a rule from a security group
func (h *Handler) RemoveSecurityGroupRule(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid user ID")
		return
	}

	// Parse request
	var req dto.RemoveSecurityGroupRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request")
		return
	}

	// Parse credential ID
	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Remove security group rule
	result, err := h.networkService.RemoveSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to remove security group rule")
		responses.InternalServerError(c, "Failed to remove security group rule: "+err.Error())
		return
	}

	responses.OK(c, result, "Security group rule removed successfully")
}

// UpdateSecurityGroupRules updates all rules for a security group
func (h *Handler) UpdateSecurityGroupRules(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid user ID")
		return
	}

	// Parse request
	var req dto.UpdateSecurityGroupRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request")
		return
	}

	// Parse credential ID
	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Get credential
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Update security group rules
	result, err := h.networkService.UpdateSecurityGroupRules(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to update security group rules")
		responses.InternalServerError(c, "Failed to update security group rules: "+err.Error())
		return
	}

	responses.OK(c, result, "Security group rules updated successfully")
}

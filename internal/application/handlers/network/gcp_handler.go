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

// GCPHandler handles GCP network resource HTTP requests
type GCPHandler struct {
	*handlers.BaseHandler
	networkService    *service.NetworkService
	credentialService domain.CredentialService
	logger            *zap.Logger
}

// NewGCPHandler creates a new GCP network handler
func NewGCPHandler(networkService *service.NetworkService, credentialService domain.CredentialService, logger *zap.Logger) *GCPHandler {
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

	// Create request (no region needed for VPC - Global resource)
	req := dto.ListVPCsRequest{
		CredentialID: credential.ID.String(),
		Region:       "", // VPC is Global, no region needed
	}

	// List VPCs
	vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list GCP VPCs")
		responses.InternalServerError(c, "Failed to list VPCs: "+err.Error())
		return
	}

	responses.OK(c, vpcs, "GCP VPCs retrieved successfully")
}

// ListGCPSubnets handles subnet listing requests for GCP
func (h *GCPHandler) ListGCPSubnets(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to list GCP subnets")
		responses.InternalServerError(c, "Failed to list subnets: "+err.Error())
		return
	}

	responses.OK(c, subnets, "GCP subnets retrieved successfully")
}

// ListGCPSecurityGroups handles security group listing requests for GCP
func (h *GCPHandler) ListGCPSecurityGroups(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to list GCP security groups")
		responses.InternalServerError(c, "Failed to list security groups: "+err.Error())
		return
	}

	responses.OK(c, securityGroups, "GCP security groups retrieved successfully")
}

// GetGCPVPC handles VPC detail requests for GCP
func (h *GCPHandler) GetGCPVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		responses.BadRequest(c, "VPC name is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create request
	req := dto.GetVPCRequest{
		CredentialID: credentialID,
		VPCID:        vpcName, // Using vpcName as VPCID for service call
		Region:       region,
	}

	// Get VPC
	vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get GCP VPC")
		responses.InternalServerError(c, "Failed to get VPC: "+err.Error())
		return
	}

	responses.OK(c, vpc, "GCP VPC retrieved successfully")
}

// CreateGCPVPC handles VPC creation requests for GCP
func (h *GCPHandler) CreateGCPVPC(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create VPC
	vpc, err := h.networkService.CreateVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create GCP VPC")
		responses.InternalServerError(c, "Failed to create VPC: "+err.Error())
		return
	}

	responses.Created(c, vpc, "GCP VPC created successfully")
}

// UpdateGCPVPC handles VPC update requests for GCP
func (h *GCPHandler) UpdateGCPVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		responses.BadRequest(c, "VPC name is required")
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Update VPC
	vpc, err := h.networkService.UpdateVPC(c.Request.Context(), credential, req, vpcName, region)
	if err != nil {
		h.LogError(c, err, "Failed to update GCP VPC")
		responses.InternalServerError(c, "Failed to update VPC: "+err.Error())
		return
	}

	responses.OK(c, vpc, "GCP VPC updated successfully")
}

// DeleteGCPVPC handles VPC deletion requests for GCP
func (h *GCPHandler) DeleteGCPVPC(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get VPC name from path parameter
	vpcName := c.Param("id")
	if vpcName == "" {
		responses.BadRequest(c, "VPC name is required")
		return
	}

	// Get credential ID from query parameter
	credentialID := c.Query("credential_id")
	if credentialID == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create delete request
	req := dto.DeleteVPCRequest{
		CredentialID: credentialID,
		VPCID:        vpcName, // Using vpcName as VPCID for service call
		Region:       region,
	}

	// Delete VPC
	err = h.networkService.DeleteVPC(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to delete GCP VPC")
		responses.InternalServerError(c, "Failed to delete VPC: "+err.Error())
		return
	}

	responses.OK(c, nil, "GCP VPC deleted successfully")
}

// GetGCPSubnet handles subnet detail requests for GCP
func (h *GCPHandler) GetGCPSubnet(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to get GCP subnet")
		responses.InternalServerError(c, "Failed to get subnet: "+err.Error())
		return
	}

	responses.OK(c, subnet, "GCP subnet retrieved successfully")
}

// CreateGCPSubnet handles subnet creation requests for GCP
func (h *GCPHandler) CreateGCPSubnet(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create subnet
	subnet, err := h.networkService.CreateSubnet(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create GCP subnet")
		responses.InternalServerError(c, "Failed to create subnet: "+err.Error())
		return
	}

	responses.Created(c, subnet, "GCP subnet created successfully")
}

// UpdateGCPSubnet handles subnet update requests for GCP
func (h *GCPHandler) UpdateGCPSubnet(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Update subnet
	subnet, err := h.networkService.UpdateSubnet(c.Request.Context(), credential, req, subnetID, region)
	if err != nil {
		h.LogError(c, err, "Failed to update GCP subnet")
		responses.InternalServerError(c, "Failed to update subnet: "+err.Error())
		return
	}

	responses.OK(c, subnet, "GCP subnet updated successfully")
}

// DeleteGCPSubnet handles subnet deletion requests for GCP
func (h *GCPHandler) DeleteGCPSubnet(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to delete GCP subnet")
		responses.InternalServerError(c, "Failed to delete subnet: "+err.Error())
		return
	}

	responses.OK(c, nil, "GCP subnet deleted successfully")
}

// GetGCPSecurityGroup handles security group detail requests for GCP
func (h *GCPHandler) GetGCPSecurityGroup(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to get GCP security group")
		responses.InternalServerError(c, "Failed to get security group: "+err.Error())
		return
	}

	responses.OK(c, securityGroup, "GCP security group retrieved successfully")
}

// CreateGCPSecurityGroup handles security group creation requests for GCP
func (h *GCPHandler) CreateGCPSecurityGroup(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create security group
	securityGroup, err := h.networkService.CreateSecurityGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create GCP security group")
		responses.InternalServerError(c, "Failed to create security group: "+err.Error())
		return
	}

	responses.Created(c, securityGroup, "GCP security group created successfully")
}

// UpdateGCPSecurityGroup handles security group update requests for GCP
func (h *GCPHandler) UpdateGCPSecurityGroup(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Update security group
	securityGroup, err := h.networkService.UpdateSecurityGroup(c.Request.Context(), credential, req, securityGroupID, region)
	if err != nil {
		h.LogError(c, err, "Failed to update GCP security group")
		responses.InternalServerError(c, "Failed to update security group: "+err.Error())
		return
	}

	responses.OK(c, securityGroup, "GCP security group updated successfully")
}

// DeleteGCPSecurityGroup handles security group deletion requests for GCP
func (h *GCPHandler) DeleteGCPSecurityGroup(c *gin.Context) {
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

	// Get region from query parameter (optional for VPC - Global resource)
	region := c.Query("region")
	// Note: VPC is a Global resource, so region is optional

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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
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
		h.LogError(c, err, "Failed to delete GCP security group")
		responses.InternalServerError(c, "Failed to delete security group: "+err.Error())
		return
	}

	responses.OK(c, nil, "GCP security group deleted successfully")
}

// AddGCPSecurityGroupRule adds a rule to a GCP security group
func (h *GCPHandler) AddGCPSecurityGroupRule(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Add security group rule
	result, err := h.networkService.AddSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to add GCP security group rule")
		responses.InternalServerError(c, "Failed to add security group rule: "+err.Error())
		return
	}

	responses.OK(c, result, "GCP security group rule added successfully")
}

// RemoveGCPSecurityGroupRule removes a rule from a GCP security group
func (h *GCPHandler) RemoveGCPSecurityGroupRule(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Remove security group rule
	result, err := h.networkService.RemoveSecurityGroupRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to remove GCP security group rule")
		responses.InternalServerError(c, "Failed to remove security group rule: "+err.Error())
		return
	}

	responses.OK(c, result, "GCP security group rule removed successfully")
}

// UpdateGCPSecurityGroupRules updates all rules for a GCP security group
func (h *GCPHandler) UpdateGCPSecurityGroupRules(c *gin.Context) {
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Update security group rules
	result, err := h.networkService.UpdateSecurityGroupRules(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to update GCP security group rules")
		responses.InternalServerError(c, "Failed to update security group rules: "+err.Error())
		return
	}

	responses.OK(c, result, "GCP security group rules updated successfully")
}

// RemoveGCPFirewallRule handles removal of specific firewall rules for GCP
func (h *GCPHandler) RemoveGCPFirewallRule(c *gin.Context) {
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
	var req dto.RemoveFirewallRuleRequest
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Set request fields
	req.CredentialID = credentialID
	req.SecurityGroupID = securityGroupID
	req.Region = region

	// Remove firewall rule
	result, err := h.networkService.RemoveFirewallRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to remove GCP firewall rule")
		responses.InternalServerError(c, "Failed to remove firewall rule: "+err.Error())
		return
	}

	responses.OK(c, result, "GCP firewall rule removed successfully")
}

// AddGCPFirewallRule handles addition of specific firewall rules for GCP
func (h *GCPHandler) AddGCPFirewallRule(c *gin.Context) {
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
	var req dto.AddFirewallRuleRequest
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

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Set request fields
	req.CredentialID = credentialID
	req.SecurityGroupID = securityGroupID
	req.Region = region

	// Add firewall rule
	result, err := h.networkService.AddFirewallRule(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to add GCP firewall rule")
		responses.InternalServerError(c, "Failed to add firewall rule: "+err.Error())
		return
	}

	responses.OK(c, result, "GCP firewall rule added successfully")
}

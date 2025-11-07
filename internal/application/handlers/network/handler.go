package network

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles network resource HTTP requests using improved patterns
type Handler struct {
	*handlers.BaseHandler
	networkService    *networkservice.Service
	credentialService domain.CredentialService
	provider          string
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new network handler
func NewHandler(networkService *networkservice.Service, credentialService domain.CredentialService, provider string) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("network"),
		networkService:    networkService,
		credentialService: credentialService,
		provider:          provider,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// VPC Handlers

// ListVPCs handles VPC listing requests using decorator pattern
func (h *Handler) ListVPCs(c *gin.Context) {
	handler := h.Compose(
		h.listVPCsHandler(),
		h.StandardCRUDDecorators("list_vpcs")...,
	)

	handler(c)
}

// listVPCsHandler is the core business logic for listing VPCs
func (h *Handler) listVPCsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		region := h.parseRegion(c)

		if region == "" {
			return
		}

		h.logVPCsListAttempt(c, userID, credential.ID, region)

		handlerReq := ListVPCsRequest{
			Region: region,
		}
		serviceReq := ToServiceListVPCsRequest(handlerReq, credential.ID.String())

		vpcs, err := h.networkService.ListVPCs(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		handlerResp := FromServiceListVPCsResponse(vpcs)
		h.logVPCsListSuccess(c, userID, len(handlerResp.VPCs))
		h.OK(c, handlerResp, "VPCs retrieved successfully")
	}
}

// ListSubnets handles subnet listing requests using decorator pattern
func (h *Handler) ListSubnets(c *gin.Context) {
	handler := h.Compose(
		h.listSubnetsHandler(),
		h.StandardCRUDDecorators("list_subnets")...,
	)

	handler(c)
}

// listSubnetsHandler is the core business logic for listing subnets
func (h *Handler) listSubnetsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "list_subnets")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		vpcID := h.parseVPCID(c)
		region := h.parseRegion(c)

		if vpcID == "" || region == "" {
			return
		}

		h.logSubnetsListAttempt(c, userID, credential.ID, vpcID, region)

		handlerReq := ListSubnetsRequest{
			VPCID:  vpcID,
			Region: region,
		}
		serviceReq := ToServiceListSubnetsRequest(handlerReq, credential.ID.String())

		subnets, err := h.networkService.ListSubnets(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "list_subnets")
			return
		}

		handlerResp := FromServiceListSubnetsResponse(subnets)
		h.logSubnetsListSuccess(c, userID, vpcID, len(handlerResp.Subnets))
		h.OK(c, handlerResp, "Subnets retrieved successfully")
	}
}

// ListSecurityGroups handles security group listing requests using decorator pattern
func (h *Handler) ListSecurityGroups(c *gin.Context) {
	handler := h.Compose(
		h.listSecurityGroupsHandler(),
		h.StandardCRUDDecorators("list_security_groups")...,
	)

	handler(c)
}

// listSecurityGroupsHandler is the core business logic for listing security groups
func (h *Handler) listSecurityGroupsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "list_security_groups")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		vpcID := h.parseVPCID(c)
		region := h.parseRegion(c)

		if vpcID == "" || region == "" {
			return
		}

		h.logSecurityGroupsListAttempt(c, userID, credential.ID, vpcID, region)

		handlerReq := ListSecurityGroupsRequest{
			VPCID:  vpcID,
			Region: region,
		}
		serviceReq := ToServiceListSecurityGroupsRequest(handlerReq, credential.ID.String())

		securityGroups, err := h.networkService.ListSecurityGroups(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "list_security_groups")
			return
		}

		handlerResp := FromServiceListSecurityGroupsResponse(securityGroups)
		h.logSecurityGroupsListSuccess(c, userID, vpcID, len(handlerResp.SecurityGroups))
		h.OK(c, handlerResp, "Security groups retrieved successfully")
	}
}

// GetVPC handles VPC detail requests using decorator pattern
func (h *Handler) GetVPC(c *gin.Context) {
	handler := h.Compose(
		h.getVPCHandler(),
		h.StandardCRUDDecorators("get_vpc")...,
	)

	handler(c)
}

// getVPCHandler is the core business logic for getting a VPC
func (h *Handler) getVPCHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "get_vpc")
			return
		}

		vpcID := h.parseResourceID(c)
		region := c.Query("region") // Optional for VPC

		if vpcID == "" {
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		h.logVPCGetAttempt(c, userID, vpcID, credential.ID, region)

		handlerReq := GetVPCRequest{
			VPCID:  vpcID,
			Region: region,
		}
		serviceReq := ToServiceGetVPCRequest(handlerReq, credential.ID.String())

		vpc, err := h.networkService.GetVPC(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "get_vpc")
			return
		}

		handlerResp := FromServiceVPCInfo(vpc)
		h.logVPCGetSuccess(c, userID, vpcID)
		h.OK(c, handlerResp, "VPC retrieved successfully")
	}
}

// CreateVPC handles VPC creation requests using decorator pattern
func (h *Handler) CreateVPC(c *gin.Context) {
	var req CreateVPCRequest

	handler := h.Compose(
		h.createVPCHandler(req),
		h.StandardCRUDDecorators("create_vpc")...,
	)

	handler(c)
}

// createVPCHandler is the core business logic for creating a VPC
func (h *Handler) createVPCHandler(req CreateVPCRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_vpc")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_vpc")
			return
		}

		serviceReq := ToServiceCreateVPCRequest(req, credential.ID.String())
		h.logVPCCreationAttempt(c, userID, serviceReq)

		ctx := h.EnrichContextWithRequestMetadata(c)
		vpc, err := h.networkService.CreateVPC(ctx, credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "create_vpc")
			return
		}

		handlerResp := FromServiceVPCInfo(vpc)
		h.logVPCCreationSuccess(c, userID, handlerResp)
		h.Created(c, handlerResp, "VPC created successfully")
	}
}

// UpdateVPC handles VPC update requests using decorator pattern
func (h *Handler) UpdateVPC(c *gin.Context) {
	var req UpdateVPCRequest

	handler := h.Compose(
		h.updateVPCHandler(req),
		h.StandardCRUDDecorators("update_vpc")...,
	)

	handler(c)
}

// updateVPCHandler is the core business logic for updating a VPC
func (h *Handler) updateVPCHandler(req UpdateVPCRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_vpc")
			return
		}

		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "update_vpc")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		vpcID := h.parseResourceID(c)

		if vpcID == "" {
			return
		}

		serviceReq := ToServiceUpdateVPCRequest(req)
		h.logVPCUpdateAttempt(c, userID, vpcID, serviceReq)

		region := c.Query("region") // Optional for VPC

		ctx := h.EnrichContextWithRequestMetadata(c)
		vpc, err := h.networkService.UpdateVPC(ctx, credential, serviceReq, vpcID, region)
		if err != nil {
			h.HandleError(c, err, "update_vpc")
			return
		}

		handlerResp := FromServiceVPCInfo(vpc)
		h.logVPCUpdateSuccess(c, userID, vpcID)
		h.OK(c, handlerResp, "VPC updated successfully")
	}
}

// DeleteVPC handles VPC deletion requests using decorator pattern
func (h *Handler) DeleteVPC(c *gin.Context) {
	handler := h.Compose(
		h.deleteVPCHandler(),
		h.StandardCRUDDecorators("delete_vpc")...,
	)

	handler(c)
}

// deleteVPCHandler is the core business logic for deleting a VPC
func (h *Handler) deleteVPCHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "delete_vpc")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		vpcID := h.parseResourceID(c)

		if vpcID == "" {
			return
		}

		h.logVPCDeletionAttempt(c, userID, vpcID)

		region := c.Query("region") // Optional for VPC

		handlerReq := DeleteVPCRequest{
			VPCID:  vpcID,
			Region: region,
		}
		serviceReq := ToServiceDeleteVPCRequest(handlerReq, credential.ID.String())

		ctx := h.EnrichContextWithRequestMetadata(c)
		if err := h.networkService.DeleteVPC(ctx, credential, serviceReq); err != nil {
			h.HandleError(c, err, "delete_vpc")
			return
		}

		h.logVPCDeletionSuccess(c, userID, vpcID)
		h.OK(c, nil, "VPC deleted successfully")
	}
}

// GetSubnet handles subnet detail requests using decorator pattern
func (h *Handler) GetSubnet(c *gin.Context) {
	handler := h.Compose(
		h.getSubnetHandler(),
		h.StandardCRUDDecorators("get_subnet")...,
	)

	handler(c)
}

// getSubnetHandler is the core business logic for getting a subnet
func (h *Handler) getSubnetHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "get_subnet")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		subnetID := h.parseResourceID(c)

		if subnetID == "" {
			return
		}

		h.logSubnetGetAttempt(c, userID, subnetID)

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		handlerReq := GetSubnetRequest{
			SubnetID: subnetID,
			Region:   region,
		}
		serviceReq := ToServiceGetSubnetRequest(handlerReq, credential.ID.String())

		subnet, err := h.networkService.GetSubnet(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "get_subnet")
			return
		}

		handlerResp := FromServiceSubnetInfo(subnet)
		h.logSubnetGetSuccess(c, userID, subnetID)
		h.OK(c, handlerResp, "Subnet retrieved successfully")
	}
}

// CreateSubnet handles subnet creation requests using decorator pattern
func (h *Handler) CreateSubnet(c *gin.Context) {
	var req CreateSubnetRequest

	handler := h.Compose(
		h.createSubnetHandler(req),
		h.StandardCRUDDecorators("create_subnet")...,
	)

	handler(c)
}

// createSubnetHandler is the core business logic for creating a subnet
func (h *Handler) createSubnetHandler(req CreateSubnetRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_subnet")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_subnet")
			return
		}

		serviceReq := ToServiceCreateSubnetRequest(req, credential.ID.String())
		h.logSubnetCreationAttempt(c, userID, serviceReq)

		ctx := h.EnrichContextWithRequestMetadata(c)
		subnet, err := h.networkService.CreateSubnet(ctx, credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "create_subnet")
			return
		}

		handlerResp := FromServiceSubnetInfo(subnet)
		h.logSubnetCreationSuccess(c, userID, handlerResp)
		h.Created(c, handlerResp, "Subnet created successfully")
	}
}

// UpdateSubnet handles subnet update requests using decorator pattern
func (h *Handler) UpdateSubnet(c *gin.Context) {
	var req UpdateSubnetRequest

	handler := h.Compose(
		h.updateSubnetHandler(req),
		h.StandardCRUDDecorators("update_subnet")...,
	)

	handler(c)
}

// updateSubnetHandler is the core business logic for updating a subnet
func (h *Handler) updateSubnetHandler(req UpdateSubnetRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_subnet")
			return
		}

		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "update_subnet")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		subnetID := h.parseResourceID(c)

		if subnetID == "" {
			return
		}

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		serviceReq := ToServiceUpdateSubnetRequest(req)
		h.logSubnetUpdateAttempt(c, userID, subnetID, serviceReq)

		ctx := h.EnrichContextWithRequestMetadata(c)
		subnet, err := h.networkService.UpdateSubnet(ctx, credential, serviceReq, subnetID, region)
		if err != nil {
			h.HandleError(c, err, "update_subnet")
			return
		}

		handlerResp := FromServiceSubnetInfo(subnet)
		h.logSubnetUpdateSuccess(c, userID, subnetID)
		h.OK(c, handlerResp, "Subnet updated successfully")
	}
}

// DeleteSubnet handles subnet deletion requests using decorator pattern
func (h *Handler) DeleteSubnet(c *gin.Context) {
	handler := h.Compose(
		h.deleteSubnetHandler(),
		h.StandardCRUDDecorators("delete_subnet")...,
	)

	handler(c)
}

// deleteSubnetHandler is the core business logic for deleting a subnet
func (h *Handler) deleteSubnetHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "delete_subnet")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		subnetID := h.parseResourceID(c)

		if subnetID == "" {
			return
		}

		h.logSubnetDeletionAttempt(c, userID, subnetID)

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		handlerReq := DeleteSubnetRequest{
			SubnetID: subnetID,
			Region:   region,
		}
		serviceReq := ToServiceDeleteSubnetRequest(handlerReq, credential.ID.String())

		ctx := h.EnrichContextWithRequestMetadata(c)
		if err := h.networkService.DeleteSubnet(ctx, credential, serviceReq); err != nil {
			h.HandleError(c, err, "delete_subnet")
			return
		}

		h.logSubnetDeletionSuccess(c, userID, subnetID)
		h.OK(c, nil, "Subnet deleted successfully")
	}
}

// GetSecurityGroup handles security group detail requests using decorator pattern
func (h *Handler) GetSecurityGroup(c *gin.Context) {
	handler := h.Compose(
		h.getSecurityGroupHandler(),
		h.StandardCRUDDecorators("get_security_group")...,
	)

	handler(c)
}

// getSecurityGroupHandler is the core business logic for getting a security group
func (h *Handler) getSecurityGroupHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "get_security_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		securityGroupID := h.parseResourceID(c)

		if securityGroupID == "" {
			return
		}

		h.logSecurityGroupGetAttempt(c, userID, securityGroupID)

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		handlerReq := GetSecurityGroupRequest{
			SecurityGroupID: securityGroupID,
			Region:          region,
		}
		serviceReq := ToServiceGetSecurityGroupRequest(handlerReq, credential.ID.String())

		securityGroup, err := h.networkService.GetSecurityGroup(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "get_security_group")
			return
		}

		handlerResp := FromServiceSecurityGroupInfo(securityGroup)
		h.logSecurityGroupGetSuccess(c, userID, securityGroupID)
		h.OK(c, handlerResp, "Security group retrieved successfully")
	}
}

// CreateSecurityGroup handles security group creation requests using decorator pattern
func (h *Handler) CreateSecurityGroup(c *gin.Context) {
	var req CreateSecurityGroupRequest

	handler := h.Compose(
		h.createSecurityGroupHandler(req),
		h.StandardCRUDDecorators("create_security_group")...,
	)

	handler(c)
}

// createSecurityGroupHandler is the core business logic for creating a security group
func (h *Handler) createSecurityGroupHandler(req CreateSecurityGroupRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_security_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_security_group")
			return
		}

		serviceReq := ToServiceCreateSecurityGroupRequest(req, credential.ID.String())
		h.logSecurityGroupCreationAttempt(c, userID, serviceReq)

		ctx := h.EnrichContextWithRequestMetadata(c)
		securityGroup, err := h.networkService.CreateSecurityGroup(ctx, credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "create_security_group")
			return
		}

		handlerResp := FromServiceSecurityGroupInfo(securityGroup)
		h.logSecurityGroupCreationSuccess(c, userID, handlerResp)
		h.Created(c, handlerResp, "Security group created successfully")
	}
}

// UpdateSecurityGroup handles security group update requests using decorator pattern
func (h *Handler) UpdateSecurityGroup(c *gin.Context) {
	var req UpdateSecurityGroupRequest

	handler := h.Compose(
		h.updateSecurityGroupHandler(req),
		h.StandardCRUDDecorators("update_security_group")...,
	)

	handler(c)
}

// updateSecurityGroupHandler is the core business logic for updating a security group
func (h *Handler) updateSecurityGroupHandler(req UpdateSecurityGroupRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_security_group")
			return
		}

		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "update_security_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		securityGroupID := h.parseResourceID(c)

		if securityGroupID == "" {
			return
		}

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		serviceReq := ToServiceUpdateSecurityGroupRequest(req)
		h.logSecurityGroupUpdateAttempt(c, userID, securityGroupID, serviceReq)

		ctx := h.EnrichContextWithRequestMetadata(c)
		securityGroup, err := h.networkService.UpdateSecurityGroup(ctx, credential, serviceReq, securityGroupID, region)
		if err != nil {
			h.HandleError(c, err, "update_security_group")
			return
		}

		handlerResp := FromServiceSecurityGroupInfo(securityGroup)
		h.logSecurityGroupUpdateSuccess(c, userID, securityGroupID)
		h.OK(c, handlerResp, "Security group updated successfully")
	}
}

// DeleteSecurityGroup handles security group deletion requests using decorator pattern
func (h *Handler) DeleteSecurityGroup(c *gin.Context) {
	handler := h.Compose(
		h.deleteSecurityGroupHandler(),
		h.StandardCRUDDecorators("delete_security_group")...,
	)

	handler(c)
}

// deleteSecurityGroupHandler is the core business logic for deleting a security group
func (h *Handler) deleteSecurityGroupHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (no provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, "")
		if err != nil {
			h.HandleError(c, err, "delete_security_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}
		securityGroupID := h.parseResourceID(c)

		if securityGroupID == "" {
			return
		}

		h.logSecurityGroupDeletionAttempt(c, userID, securityGroupID)

		region := h.parseRegion(c)

		if region == "" {
			return
		}

		handlerReq := DeleteSecurityGroupRequest{
			SecurityGroupID: securityGroupID,
			Region:          region,
		}
		serviceReq := ToServiceDeleteSecurityGroupRequest(handlerReq, credential.ID.String())

		ctx := h.EnrichContextWithRequestMetadata(c)
		if err := h.networkService.DeleteSecurityGroup(ctx, credential, serviceReq); err != nil {
			h.HandleError(c, err, "delete_security_group")
			return
		}

		h.logSecurityGroupDeletionSuccess(c, userID, securityGroupID)
		h.OK(c, nil, "Security group deleted successfully")
	}
}

// AddSecurityGroupRule adds a rule to a security group using decorator pattern
func (h *Handler) AddSecurityGroupRule(c *gin.Context) {
	var req AddSecurityGroupRuleRequest

	handler := h.Compose(
		h.addSecurityGroupRuleHandler(req),
		h.StandardCRUDDecorators("add_security_group_rule")...,
	)

	handler(c)
}

// addSecurityGroupRuleHandler is the core business logic for adding a security group rule
func (h *Handler) addSecurityGroupRuleHandler(req AddSecurityGroupRuleRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "add_security_group_rule")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "add_security_group_rule")
			return
		}

		serviceReq := ToServiceAddSecurityGroupRuleRequest(req, credential.ID.String())
		h.logSecurityGroupRuleAdditionAttempt(c, userID, serviceReq)

		result, err := h.networkService.AddSecurityGroupRule(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "add_security_group_rule")
			return
		}

		handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "Security group rule added successfully")
		h.logSecurityGroupRuleAdditionSuccess(c, userID, serviceReq)
		h.OK(c, handlerResp, "Security group rule added successfully")
	}
}

// RemoveSecurityGroupRule removes a rule from a security group using decorator pattern
func (h *Handler) RemoveSecurityGroupRule(c *gin.Context) {
	var req RemoveSecurityGroupRuleRequest

	handler := h.Compose(
		h.removeSecurityGroupRuleHandler(req),
		h.StandardCRUDDecorators("remove_security_group_rule")...,
	)

	handler(c)
}

// removeSecurityGroupRuleHandler is the core business logic for removing a security group rule
func (h *Handler) removeSecurityGroupRuleHandler(req RemoveSecurityGroupRuleRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "remove_security_group_rule")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "remove_security_group_rule")
			return
		}

		serviceReq := ToServiceRemoveSecurityGroupRuleRequest(req, credential.ID.String())
		h.logSecurityGroupRuleRemovalAttempt(c, userID, serviceReq)

		result, err := h.networkService.RemoveSecurityGroupRule(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "remove_security_group_rule")
			return
		}

		handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "Security group rule removed successfully")
		h.logSecurityGroupRuleRemovalSuccess(c, userID, serviceReq)
		h.OK(c, handlerResp, "Security group rule removed successfully")
	}
}

// UpdateSecurityGroupRules updates all rules for a security group using decorator pattern
func (h *Handler) UpdateSecurityGroupRules(c *gin.Context) {
	var req UpdateSecurityGroupRulesRequest

	handler := h.Compose(
		h.updateSecurityGroupRulesHandler(req),
		h.StandardCRUDDecorators("update_security_group_rules")...,
	)

	handler(c)
}

// updateSecurityGroupRulesHandler is the core business logic for updating security group rules
func (h *Handler) updateSecurityGroupRulesHandler(req UpdateSecurityGroupRulesRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_security_group_rules")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_vpcs")
			return
		}

		// Get credential from query or body
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "update_security_group_rules")
			return
		}

		serviceReq := ToServiceUpdateSecurityGroupRulesRequest(req, credential.ID.String())
		h.logSecurityGroupRulesUpdateAttempt(c, userID, serviceReq)

		result, err := h.networkService.UpdateSecurityGroupRules(c.Request.Context(), credential, serviceReq)
		if err != nil {
			h.HandleError(c, err, "update_security_group_rules")
			return
		}

		handlerResp := FromServiceSecurityGroupInfoToRuleResponse(result, true, "Security group rules updated successfully")
		h.logSecurityGroupRulesUpdateSuccess(c, userID, serviceReq)
		h.OK(c, handlerResp, "Security group rules updated successfully")
	}
}

// Helper methods for better readability

func (h *Handler) parseRegion(c *gin.Context) string {
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "parse_region")
		return ""
	}
	return region
}

func (h *Handler) parseVPCID(c *gin.Context) string {
	vpcID := c.Query("vpc_id")
	if vpcID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "vpc_id is required", 400), "parse_vpc_id")
		return ""
	}
	return vpcID
}

func (h *Handler) parseResourceID(c *gin.Context) string {
	resourceID := c.Param("id")
	if resourceID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "resource ID is required", 400), "parse_resource_id")
		return ""
	}
	return resourceID
}

// Logging helper methods

func (h *Handler) logVPCsListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "vpcs_list_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "list_vpcs",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logVPCsListSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "vpcs_listed", userID.String(), "", map[string]interface{}{
		"operation": "list_vpcs",
		"provider":  h.provider,
		"count":     count,
	})
}

func (h *Handler) logSubnetsListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, vpcID, region string) {
	h.LogBusinessEvent(c, "subnets_list_attempted", userID.String(), vpcID, map[string]interface{}{
		"operation":     "list_subnets",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logSubnetsListSuccess(c *gin.Context, userID uuid.UUID, vpcID string, count int) {
	h.LogBusinessEvent(c, "subnets_listed", userID.String(), vpcID, map[string]interface{}{
		"operation": "list_subnets",
		"provider":  h.provider,
		"count":     count,
	})
}

func (h *Handler) logSecurityGroupsListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, vpcID, region string) {
	h.LogBusinessEvent(c, "security_groups_list_attempted", userID.String(), vpcID, map[string]interface{}{
		"operation":     "list_security_groups",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logSecurityGroupsListSuccess(c *gin.Context, userID uuid.UUID, vpcID string, count int) {
	h.LogBusinessEvent(c, "security_groups_listed", userID.String(), vpcID, map[string]interface{}{
		"operation": "list_security_groups",
		"provider":  h.provider,
		"count":     count,
	})
}

func (h *Handler) logVPCGetAttempt(c *gin.Context, userID uuid.UUID, vpcID string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "vpc_get_attempted", userID.String(), vpcID, map[string]interface{}{
		"operation":     "get_vpc",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logVPCGetSuccess(c *gin.Context, userID uuid.UUID, vpcID string) {
	h.LogBusinessEvent(c, "vpc_retrieved", userID.String(), vpcID, map[string]interface{}{
		"operation": "get_vpc",
		"provider":  h.provider,
	})
}

// Additional helper methods for VPC operations

// Deprecated helpers removed: validation now handled via h.ValidateRequest in handlers

func (h *Handler) logVPCCreationAttempt(c *gin.Context, userID uuid.UUID, req networkservice.CreateVPCRequest) {
	h.LogBusinessEvent(c, "vpc_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_vpc",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logVPCCreationSuccess(c *gin.Context, userID uuid.UUID, vpc interface{}) {
	h.LogBusinessEvent(c, "vpc_created", userID.String(), "", map[string]interface{}{
		"operation": "create_vpc",
		"provider":  h.provider,
	})
}

func (h *Handler) logVPCUpdateAttempt(c *gin.Context, userID uuid.UUID, vpcID string, req networkservice.UpdateVPCRequest) {
	h.LogBusinessEvent(c, "vpc_update_attempted", userID.String(), vpcID, map[string]interface{}{
		"operation": "update_vpc",
		"provider":  h.provider,
	})
}

func (h *Handler) logVPCUpdateSuccess(c *gin.Context, userID uuid.UUID, vpcID string) {
	h.LogBusinessEvent(c, "vpc_updated", userID.String(), vpcID, map[string]interface{}{
		"operation": "update_vpc",
		"provider":  h.provider,
	})
}

func (h *Handler) logVPCDeletionAttempt(c *gin.Context, userID uuid.UUID, vpcID string) {
	h.LogBusinessEvent(c, "vpc_deletion_attempted", userID.String(), vpcID, map[string]interface{}{
		"operation": "delete_vpc",
		"provider":  h.provider,
	})
}

func (h *Handler) logVPCDeletionSuccess(c *gin.Context, userID uuid.UUID, vpcID string) {
	h.LogBusinessEvent(c, "vpc_deleted", userID.String(), vpcID, map[string]interface{}{
		"operation": "delete_vpc",
		"provider":  h.provider,
	})
}

// Additional helper methods for Subnet operations

// Deprecated helpers removed: validation now handled via h.ValidateRequest in handlers

func (h *Handler) logSubnetGetAttempt(c *gin.Context, userID uuid.UUID, subnetID string) {
	h.LogBusinessEvent(c, "subnet_get_attempted", userID.String(), subnetID, map[string]interface{}{
		"operation": "get_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetGetSuccess(c *gin.Context, userID uuid.UUID, subnetID string) {
	h.LogBusinessEvent(c, "subnet_retrieved", userID.String(), subnetID, map[string]interface{}{
		"operation": "get_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetCreationAttempt(c *gin.Context, userID uuid.UUID, req networkservice.CreateSubnetRequest) {
	h.LogBusinessEvent(c, "subnet_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_subnet",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logSubnetCreationSuccess(c *gin.Context, userID uuid.UUID, subnet interface{}) {
	h.LogBusinessEvent(c, "subnet_created", userID.String(), "", map[string]interface{}{
		"operation": "create_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetUpdateAttempt(c *gin.Context, userID uuid.UUID, subnetID string, req networkservice.UpdateSubnetRequest) {
	h.LogBusinessEvent(c, "subnet_update_attempted", userID.String(), subnetID, map[string]interface{}{
		"operation": "update_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetUpdateSuccess(c *gin.Context, userID uuid.UUID, subnetID string) {
	h.LogBusinessEvent(c, "subnet_updated", userID.String(), subnetID, map[string]interface{}{
		"operation": "update_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetDeletionAttempt(c *gin.Context, userID uuid.UUID, subnetID string) {
	h.LogBusinessEvent(c, "subnet_deletion_attempted", userID.String(), subnetID, map[string]interface{}{
		"operation": "delete_subnet",
		"provider":  h.provider,
	})
}

func (h *Handler) logSubnetDeletionSuccess(c *gin.Context, userID uuid.UUID, subnetID string) {
	h.LogBusinessEvent(c, "subnet_deleted", userID.String(), subnetID, map[string]interface{}{
		"operation": "delete_subnet",
		"provider":  h.provider,
	})
}

// Additional helper methods for Security Group operations

// Deprecated helpers removed: validation now handled via h.ValidateRequest in handlers

func (h *Handler) logSecurityGroupGetAttempt(c *gin.Context, userID uuid.UUID, securityGroupID string) {
	h.LogBusinessEvent(c, "security_group_get_attempted", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "get_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupGetSuccess(c *gin.Context, userID uuid.UUID, securityGroupID string) {
	h.LogBusinessEvent(c, "security_group_retrieved", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "get_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupCreationAttempt(c *gin.Context, userID uuid.UUID, req networkservice.CreateSecurityGroupRequest) {
	h.LogBusinessEvent(c, "security_group_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_security_group",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logSecurityGroupCreationSuccess(c *gin.Context, userID uuid.UUID, securityGroup interface{}) {
	h.LogBusinessEvent(c, "security_group_created", userID.String(), "", map[string]interface{}{
		"operation": "create_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupUpdateAttempt(c *gin.Context, userID uuid.UUID, securityGroupID string, req networkservice.UpdateSecurityGroupRequest) {
	h.LogBusinessEvent(c, "security_group_update_attempted", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "update_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupUpdateSuccess(c *gin.Context, userID uuid.UUID, securityGroupID string) {
	h.LogBusinessEvent(c, "security_group_updated", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "update_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupDeletionAttempt(c *gin.Context, userID uuid.UUID, securityGroupID string) {
	h.LogBusinessEvent(c, "security_group_deletion_attempted", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "delete_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupDeletionSuccess(c *gin.Context, userID uuid.UUID, securityGroupID string) {
	h.LogBusinessEvent(c, "security_group_deleted", userID.String(), securityGroupID, map[string]interface{}{
		"operation": "delete_security_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupRuleAdditionAttempt(c *gin.Context, userID uuid.UUID, req networkservice.AddSecurityGroupRuleRequest) {
	h.LogBusinessEvent(c, "security_group_rule_addition_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "add_security_group_rule",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logSecurityGroupRuleAdditionSuccess(c *gin.Context, userID uuid.UUID, req networkservice.AddSecurityGroupRuleRequest) {
	h.LogBusinessEvent(c, "security_group_rule_added", userID.String(), "", map[string]interface{}{
		"operation": "add_security_group_rule",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupRuleRemovalAttempt(c *gin.Context, userID uuid.UUID, req networkservice.RemoveSecurityGroupRuleRequest) {
	h.LogBusinessEvent(c, "security_group_rule_removal_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "remove_security_group_rule",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logSecurityGroupRuleRemovalSuccess(c *gin.Context, userID uuid.UUID, req networkservice.RemoveSecurityGroupRuleRequest) {
	h.LogBusinessEvent(c, "security_group_rule_removed", userID.String(), "", map[string]interface{}{
		"operation": "remove_security_group_rule",
		"provider":  h.provider,
	})
}

func (h *Handler) logSecurityGroupRulesUpdateAttempt(c *gin.Context, userID uuid.UUID, req networkservice.UpdateSecurityGroupRulesRequest) {
	h.LogBusinessEvent(c, "security_group_rules_update_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "update_security_group_rules",
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logSecurityGroupRulesUpdateSuccess(c *gin.Context, userID uuid.UUID, req networkservice.UpdateSecurityGroupRulesRequest) {
	h.LogBusinessEvent(c, "security_group_rules_updated", userID.String(), "", map[string]interface{}{
		"operation": "update_security_group_rules",
		"provider":  h.provider,
	})
}

package resourcegroup

import (
	"strconv"

	resourcegroupservice "skyclust/internal/application/services/resourcegroup"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// Handler: Azure Resource Group HTTP 요청을 처리하는 핸들러
type Handler struct {
	*handlers.ProviderBaseHandler[*resourcegroupservice.Service]
}

// NewHandler: 새로운 Resource Group 핸들러를 생성합니다
func NewHandler(
	resourceGroupService *resourcegroupservice.Service,
	credentialService domain.CredentialService,
) *Handler {
	return &Handler{
		ProviderBaseHandler: handlers.NewProviderBaseHandler(
			resourceGroupService,
			credentialService,
			domain.ProviderAzure,
			"azure-resource-group",
		),
	}
}

// ListResourceGroups: Resource Group 목록 조회를 처리합니다
func (h *Handler) ListResourceGroups(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_resource_groups")
		return
	}

	// Parse query parameters manually (credential_id는 제외)
	// limit이 전달되지 않으면 0으로 설정하여 클라이언트 사이드 페이징 지원
	limitStr := c.Query("limit")
	var limit int
	if limitStr == "" {
		limit = 0 // 클라이언트 사이드 페이징을 위한 모든 데이터 반환
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			limit = 0 // 잘못된 값이면 모든 데이터 반환
		}
	}

	// page는 항상 파싱 (클라이언트 사이드 페이징에서도 메타데이터에 필요)
	pageStr := c.Query("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}

	req := resourcegroupservice.ListResourceGroupsRequest{
		CredentialID: credential.ID.String(),
		Location:     c.Query("location"),
		Search:       c.Query("search"),
		SortBy:       c.Query("sort_by"),
		SortOrder:    c.Query("sort_order"),
		Page:         page,
		Limit:        limit,
	}

	// Validate and set defaults
	if req.SortOrder != "" && req.SortOrder != "asc" && req.SortOrder != "desc" {
		req.SortOrder = "asc"
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}

	resourceGroups, err := h.GetService().ListResourceGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "list_resource_groups")
		return
	}

	total := int64(0)
	rgList := []resourcegroupservice.ResourceGroupInfo{}
	if resourceGroups != nil {
		total = resourceGroups.Total
		rgList = resourceGroups.ResourceGroups
	}

	// 클라이언트 사이드 페이징인 경우 limit을 total로 설정하여 메타데이터 일관성 유지
	metaLimit := limit
	if limit == 0 || limit >= 1000 {
		metaLimit = int(total)
		if metaLimit == 0 {
			metaLimit = 1 // 최소값 1
		}
	}

	h.OKWithPagination(c, rgList, "Resource groups retrieved successfully", page, metaLimit, total)
}

// GetResourceGroup: Resource Group 상세 조회를 처리합니다
func (h *Handler) GetResourceGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_resource_group")
		return
	}

	name := c.Param("name")
	if name == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "resource group name is required", 400), "get_resource_group")
		return
	}

	rg, err := h.GetService().GetResourceGroup(c.Request.Context(), credential, name)
	if err != nil {
		h.HandleError(c, err, "get_resource_group")
		return
	}

	h.OK(c, rg, "Resource group retrieved successfully")
}

// CreateResourceGroup: Resource Group 생성을 처리합니다
func (h *Handler) CreateResourceGroup(c *gin.Context) {
	var req resourcegroupservice.CreateResourceGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_resource_group")
		return
	}

	// credential_id는 body 또는 query parameter에서 가져올 수 있음
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = c.Query("credential_id")
	}

	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "create_resource_group")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "create_resource_group")
		return
	}

	req.CredentialID = credential.ID.String()

	rg, err := h.GetService().CreateResourceGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_resource_group")
		return
	}

	h.Created(c, rg, "Resource group created successfully")
}

// UpdateResourceGroup: Resource Group 수정을 처리합니다
func (h *Handler) UpdateResourceGroup(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "resource group name is required", 400), "update_resource_group")
		return
	}

	var req resourcegroupservice.UpdateResourceGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_resource_group")
		return
	}

	// credential_id는 body 또는 query parameter에서 가져올 수 있음
	credentialID := req.CredentialID
	if credentialID == "" {
		credentialID = c.Query("credential_id")
	}

	if credentialID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "update_resource_group")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), credentialID, domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "update_resource_group")
		return
	}

	req.CredentialID = credential.ID.String()

	rg, err := h.GetService().UpdateResourceGroup(c.Request.Context(), credential, name, req)
	if err != nil {
		h.HandleError(c, err, "update_resource_group")
		return
	}

	h.OK(c, rg, "Resource group updated successfully")
}

// DeleteResourceGroup: Resource Group 삭제를 처리합니다
func (h *Handler) DeleteResourceGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "delete_resource_group")
		return
	}

	name := c.Param("name")
	if name == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "resource group name is required", 400), "delete_resource_group")
		return
	}

	err = h.GetService().DeleteResourceGroup(c.Request.Context(), credential, name)
	if err != nil {
		h.HandleError(c, err, "delete_resource_group")
		return
	}

	h.OK(c, nil, "Resource group deletion initiated")
}

// GetService returns the resource group service instance
func (h *Handler) GetService() *resourcegroupservice.Service {
	return h.ProviderBaseHandler.GetService()
}

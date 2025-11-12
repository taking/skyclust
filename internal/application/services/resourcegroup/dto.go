package resourcegroup

// ListResourceGroupsRequest represents a request to list resource groups
type ListResourceGroupsRequest struct {
	CredentialID string `json:"credential_id"`                      // Set by handler, not from query
	Location     string `json:"location,omitempty" form:"location"` // Azure region filter
	Search       string `json:"search,omitempty" form:"search"`
	Page         int    `json:"page,omitempty" form:"page"`
	Limit        int    `json:"limit,omitempty" form:"limit"`
	SortBy       string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder    string `json:"sort_order,omitempty" form:"sort_order"` // "asc" or "desc"
}

// ListResourceGroupsResponse represents the response after listing resource groups
type ListResourceGroupsResponse struct {
	ResourceGroups []ResourceGroupInfo `json:"resource_groups"`
	Total          int64               `json:"total"`
}

// ResourceGroupInfo represents Azure Resource Group information
type ResourceGroupInfo struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Location          string            `json:"location"`
	ProvisioningState string            `json:"provisioning_state"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// CreateResourceGroupRequest represents a request to create a resource group
type CreateResourceGroupRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=90"`
	Location     string            `json:"location" validate:"required"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// UpdateResourceGroupRequest represents a request to update a resource group (tags only)
type UpdateResourceGroupRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Tags         map[string]string `json:"tags,omitempty"`
}

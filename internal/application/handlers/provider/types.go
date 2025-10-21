package provider

import "time"

// ProviderResponse represents a cloud provider in API responses
type ProviderResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

// ProviderListResponse represents a list of providers
type ProviderListResponse struct {
	Providers []*ProviderResponse `json:"providers"`
	Total     int                 `json:"total"`
}

// InstanceResponse represents a cloud instance in API responses
type InstanceResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Region    string            `json:"region"`
	PublicIP  string            `json:"public_ip,omitempty"`
	PrivateIP string            `json:"private_ip,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// InstanceListResponse represents a list of instances
type InstanceListResponse struct {
	Instances []*InstanceResponse `json:"instances"`
	Total     int                 `json:"total"`
}

// CreateInstanceRequest represents an instance creation request
type CreateInstanceRequest struct {
	Name   string                 `json:"name" validate:"required"`
	Type   string                 `json:"type" validate:"required"`
	Region string                 `json:"region" validate:"required"`
	Tags   map[string]string      `json:"tags,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// RegionResponse represents a cloud region in API responses
type RegionResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

// RegionListResponse represents a list of regions
type RegionListResponse struct {
	Regions []*RegionResponse `json:"regions"`
	Total   int               `json:"total"`
}

// CostEstimateRequest represents a cost estimation request
type CostEstimateRequest struct {
	InstanceType string                 `json:"instance_type" validate:"required"`
	Region       string                 `json:"region" validate:"required"`
	Duration     string                 `json:"duration" validate:"required"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// CostEstimateResponse represents a cost estimate in API responses
type CostEstimateResponse struct {
	InstanceType  string                 `json:"instance_type"`
	Region        string                 `json:"region"`
	Duration      string                 `json:"duration"`
	EstimatedCost float64                `json:"estimated_cost"`
	Currency      string                 `json:"currency"`
	Breakdown     map[string]interface{} `json:"breakdown,omitempty"`
}

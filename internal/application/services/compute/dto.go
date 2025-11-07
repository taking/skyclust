package compute

// ComputeInstance represents a cloud compute instance (EC2, GCE, Azure VM, etc.)
type ComputeInstance struct {
	ID       string                 `json:"id"`
	Status   string                 `json:"status"`
	Type     string                 `json:"type"`
	Region   string                 `json:"region"`
	ImageID  string                 `json:"image_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateInstanceRequest represents a request to create a compute instance
type CreateInstanceRequest struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Region   string                 `json:"region"`
	ImageID  string                 `json:"image_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"skyclust/internal/plugin/interfaces"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// GCPExampleProvider implements the CloudProvider interface for Google Cloud Platform
// This is a simplified example showing how to create a basic GCP provider
type GCPExampleProvider struct {
	config          map[string]interface{}
	instancesClient *compute.InstancesClient
	zonesClient     *compute.ZonesClient
	projectID       string
	region          string
}

// New creates a new GCP provider instance
// This function is required for plugin loading
func New() interfaces.CloudProvider {
	return &GCPExampleProvider{}
}

// GetName returns the provider name
func (p *GCPExampleProvider) GetName() string {
	return "GCP Example"
}

// GetVersion returns the provider version
func (p *GCPExampleProvider) GetVersion() string {
	return "1.0.0"
}

// Initialize initializes the GCP provider with configuration
func (p *GCPExampleProvider) Initialize(config map[string]interface{}) error {
	p.config = config

	// Validate required configuration
	if _, ok := config["project_id"]; !ok {
		return fmt.Errorf("GCP project_id is required")
	}
	if _, ok := config["region"]; !ok {
		config["region"] = "us-central1" // Default region
	}

	p.projectID = config["project_id"].(string)
	p.region = config["region"].(string)

	ctx := context.Background()

	// Create instances client
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instances client: %w", err)
	}
	p.instancesClient = instancesClient

	// Create zones client
	zonesClient, err := compute.NewZonesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create zones client: %w", err)
	}
	p.zonesClient = zonesClient

	fmt.Printf("GCP Example provider initialized for project: %s, region: %s\n", p.projectID, p.region)
	return nil
}

// ListInstances returns a list of GCP Compute Engine instances
func (p *GCPExampleProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
	// Get zones in the region
	zones, err := p.getZonesInRegion(ctx, p.region)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	var instances []interfaces.Instance

	// List instances in each zone
	for _, zone := range zones {
		req := &computepb.ListInstancesRequest{
			Project: p.projectID,
			Zone:    zone,
		}

		it := p.instancesClient.List(ctx, req)
		for {
			instance, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				continue // Skip zones with errors
			}

			// Get instance name
			name := instance.GetName()

			// Get instance status
			status := "UNKNOWN"
			if instance.Status != nil {
				status = *instance.Status
			}

			// Get machine type
			machineType := "unknown"
			if instance.MachineType != nil {
				machineType = (*instance.MachineType)[strings.LastIndex(*instance.MachineType, "/")+1:]
			}

			// Get creation time
			createdAt := time.Now().Format(time.RFC3339)
			if instance.CreationTimestamp != nil {
				// Parse the timestamp string
				if t, err := time.Parse(time.RFC3339, *instance.CreationTimestamp); err == nil {
					createdAt = t.Format(time.RFC3339)
				}
			}

			// Get network interfaces
			var publicIP, privateIP string
			if len(instance.NetworkInterfaces) > 0 {
				ni := instance.NetworkInterfaces[0]
				if ni.NetworkIP != nil {
					privateIP = *ni.NetworkIP
				}
				if len(ni.AccessConfigs) > 0 && ni.AccessConfigs[0].NatIP != nil {
					publicIP = *ni.AccessConfigs[0].NatIP
				}
			}

			// Convert labels to tags
			tags := make(map[string]string)
			if instance.Labels != nil {
				for key, value := range instance.Labels {
					tags[key] = value
				}
			}
			tags["Project"] = p.projectID
			tags["Zone"] = zone

			instances = append(instances, interfaces.Instance{
				ID:        instance.GetName(),
				Name:      name,
				Status:    status,
				Type:      machineType,
				Region:    p.region,
				CreatedAt: createdAt,
				Tags:      tags,
				PublicIP:  publicIP,
				PrivateIP: privateIP,
			})
		}
	}

	return instances, nil
}

// CreateInstance creates a new GCP Compute Engine instance
func (p *GCPExampleProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error) {
	// Get zones in the region
	zones, err := p.getZonesInRegion(ctx, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("no zones found in region %s", req.Region)
	}

	// Use the first zone
	zone := zones[0]

	// Prepare labels
	labels := make(map[string]string)
	for key, value := range req.Tags {
		labels[key] = value
	}

	// Create instance configuration
	instanceConfig := &computepb.Instance{
		Name:        &req.Name,
		MachineType: &req.Type,
		Labels:      labels,
	}

	// Add user data if provided
	if req.UserData != "" {
		instanceConfig.Metadata = &computepb.Metadata{
			Items: []*computepb.Items{
				{
					Key:   &[]string{"startup-script"}[0],
					Value: &req.UserData,
				},
			},
		}
	}

	// Create the instance
	req_gcp := &computepb.InsertInstanceRequest{
		Project:          p.projectID,
		Zone:             zone,
		InstanceResource: instanceConfig,
	}

	op, err := p.instancesClient.Insert(ctx, req_gcp)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Wait for operation to complete (optional)
	_ = op

	fmt.Printf("Creating GCP instance: %s (%s) in %s\n", req.Name, req.Type, req.Region)

	return &interfaces.Instance{
		ID:        req.Name,
		Name:      req.Name,
		Status:    "PROVISIONING",
		Type:      req.Type,
		Region:    req.Region,
		CreatedAt: time.Now().Format(time.RFC3339),
		Tags:      req.Tags,
	}, nil
}

// DeleteInstance deletes a GCP Compute Engine instance
func (p *GCPExampleProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	// Get zones in the region
	zones, err := p.getZonesInRegion(ctx, p.region)
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	for _, zone := range zones {
		req := &computepb.DeleteInstanceRequest{
			Project:  p.projectID,
			Zone:     zone,
			Instance: instanceID,
		}

		_, err := p.instancesClient.Delete(ctx, req)
		if err == nil {
			fmt.Printf("Deleting GCP instance: %s from zone %s\n", instanceID, zone)
			return nil
		}
	}

	return fmt.Errorf("instance not found: %s", instanceID)
}

// GetInstanceStatus returns the status of a GCP instance
func (p *GCPExampleProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
	// Get zones in the region
	zones, err := p.getZonesInRegion(ctx, p.region)
	if err != nil {
		return "", fmt.Errorf("failed to get zones: %w", err)
	}

	for _, zone := range zones {
		req := &computepb.GetInstanceRequest{
			Project:  p.projectID,
			Zone:     zone,
			Instance: instanceID,
		}

		instance, err := p.instancesClient.Get(ctx, req)
		if err == nil {
			if instance.Status != nil {
				return *instance.Status, nil
			}
			return "UNKNOWN", nil
		}
	}

	return "", fmt.Errorf("instance not found: %s", instanceID)
}

// ListRegions returns available GCP regions
func (p *GCPExampleProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error) {
	req := &computepb.ListZonesRequest{
		Project: p.projectID,
	}

	it := p.zonesClient.List(ctx, req)
	regionMap := make(map[string]string) // region -> display name

	for {
		zone, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		zoneName := zone.GetName()
		// Extract region from zone name (e.g., "us-central1-a" -> "us-central1")
		parts := strings.Split(zoneName, "-")
		if len(parts) >= 2 {
			region := strings.Join(parts[:len(parts)-1], "-")
			if regionMap[region] == "" {
				regionMap[region] = region
			}
		}
	}

	var regions []interfaces.Region
	for region, description := range regionMap {
		regions = append(regions, interfaces.Region{
			ID:          region,
			Name:        region,
			DisplayName: description,
			Status:      "available",
		})
	}

	return regions, nil
}

// GetCostEstimate returns cost estimate for GCP resources
func (p *GCPExampleProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
	// Mock cost calculation
	var costPerHour float64
	switch req.InstanceType {
	case "e2-micro":
		costPerHour = 0.006
	case "e2-small":
		costPerHour = 0.012
	case "e2-medium":
		costPerHour = 0.024
	default:
		costPerHour = 0.05
	}

	// Simple duration calculation
	var multiplier float64
	switch req.Duration {
	case "1h":
		multiplier = 1
	case "1d":
		multiplier = 24
	case "1m":
		multiplier = 24 * 30
	default:
		multiplier = 1
	}

	return &interfaces.CostEstimate{
		InstanceType: req.InstanceType,
		Region:       req.Region,
		Duration:     req.Duration,
		Cost:         costPerHour * multiplier,
		Currency:     "USD",
	}, nil
}

// GetNetworkProvider returns the network provider (not implemented for GCP in this example)
func (p *GCPExampleProvider) GetNetworkProvider() interfaces.NetworkProvider {
	return nil
}

// GetIAMProvider returns the IAM provider (not implemented for GCP in this example)
func (p *GCPExampleProvider) GetIAMProvider() interfaces.IAMProvider {
	return nil
}

// Helper function to get zones in a specific region
func (p *GCPExampleProvider) getZonesInRegion(ctx context.Context, region string) ([]string, error) {
	req := &computepb.ListZonesRequest{
		Project: p.projectID,
	}

	it := p.zonesClient.List(ctx, req)
	var zones []string

	for {
		zone, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Check if zone is in the specified region
		if strings.HasPrefix(zone.GetName(), region+"-") {
			zones = append(zones, zone.GetName())
		}
	}

	return zones, nil
}

// Example usage function
func main() {
	// This is just for demonstration - in a real plugin, this wouldn't be needed
	log.Println("GCP Example Provider - This is a template for creating GCP plugins")
}

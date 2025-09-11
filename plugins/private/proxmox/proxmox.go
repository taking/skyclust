package main

import (
	"context"
	"fmt"
	"time"

	"cmp/pkg/interfaces"

	"github.com/luthermonson/go-proxmox"
)

// ProxmoxProvider implements the CloudProvider interface for Proxmox
type ProxmoxProvider struct {
	config   map[string]interface{}
	client   *proxmox.Client
	host     string
	username string
	password string
	realm    string
}

// New creates a new Proxmox provider instance
// This function is required for plugin loading
func New() interfaces.CloudProvider {
	return &ProxmoxProvider{}
}

// GetName returns the provider name
func (p *ProxmoxProvider) GetName() string {
	return "Proxmox"
}

// GetVersion returns the provider version
func (p *ProxmoxProvider) GetVersion() string {
	return "1.0.0"
}

// Initialize initializes the Proxmox provider with configuration
func (p *ProxmoxProvider) Initialize(config map[string]interface{}) error {
	p.config = config

	// Validate required configuration
	if _, ok := config["host"]; !ok {
		return fmt.Errorf("Proxmox host is required")
	}
	if _, ok := config["username"]; !ok {
		return fmt.Errorf("Proxmox username is required")
	}
	if _, ok := config["password"]; !ok {
		return fmt.Errorf("Proxmox password is required")
	}
	if _, ok := config["realm"]; !ok {
		config["realm"] = "pve" // Default realm
	}

	p.host = config["host"].(string)
	p.username = config["username"].(string)
	p.password = config["password"].(string)
	p.realm = config["realm"].(string)

	// Create Proxmox client
	p.client = proxmox.NewClient(fmt.Sprintf("https://%s:8006", p.host))

	// Authenticate
	ctx := context.Background()
	err := p.client.Login(ctx, p.username, p.password)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Proxmox: %w", err)
	}

	fmt.Printf("Proxmox provider initialized for host: %s\n", p.host)
	return nil
}

// GetNetworkProvider returns the network provider (not implemented for Proxmox in this example)
func (p *ProxmoxProvider) GetNetworkProvider() interfaces.NetworkProvider {
	return nil
}

// GetIAMProvider returns the IAM provider (not implemented for Proxmox in this example)
func (p *ProxmoxProvider) GetIAMProvider() interfaces.IAMProvider {
	return nil
}

// ListInstances returns a list of Proxmox VMs
func (p *ProxmoxProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
	// Get all nodes
	nodes, err := p.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var instances []interfaces.Instance
	for _, node := range nodes {
		// Get VMs for each node using the correct API
		nodeObj, err := p.client.Node(ctx, node.Node)
		if err != nil {
			continue // Skip nodes with errors
		}

		vms, err := nodeObj.VirtualMachines(ctx)
		if err != nil {
			continue // Skip nodes with errors
		}

		for _, vm := range vms {
			// Get VM name
			name := vm.Name
			if name == "" {
				name = fmt.Sprintf("VM-%d", vm.VMID)
			}

			// Get VM status string
			statusStr := "unknown"
			if vm.Status != "" {
				statusStr = vm.Status
			}

			// Get VM type
			vmType := "qemu"
			if vm.Template {
				vmType = "template"
			}

			// Get creation time (Proxmox doesn't provide this, use current time)
			createdAt := time.Now().Format(time.RFC3339)

			// Get network interfaces (simplified)
			var publicIP, privateIP string
			// In a real implementation, you would parse the network configuration
			// For now, we'll leave these empty

			// Convert tags
			tags := make(map[string]string)
			if vm.Tags != "" {
				// Parse tags string (comma-separated)
				tags["Tags"] = vm.Tags
			}
			tags["Node"] = node.Node
			tags["VMID"] = fmt.Sprintf("%d", vm.VMID)

			instances = append(instances, interfaces.Instance{
				ID:        fmt.Sprintf("%d", vm.VMID),
				Name:      name,
				Status:    statusStr,
				Type:      vmType,
				Region:    node.Node,
				CreatedAt: createdAt,
				Tags:      tags,
				PublicIP:  publicIP,
				PrivateIP: privateIP,
			})
		}
	}

	return instances, nil
}

// CreateInstance creates a new Proxmox VM
func (p *ProxmoxProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error) {
	// Get all nodes to find the first available one
	nodes, err := p.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Use the first node
	node := nodes[0]
	nodeObj, err := p.client.Node(ctx, node.Node)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Find next available VMID
	vmid := 100 // Start from 100
	for {
		_, err := nodeObj.VirtualMachine(ctx, vmid)
		if err != nil {
			break // VMID is available
		}
		vmid++
		if vmid > 999999 {
			return nil, fmt.Errorf("no available VMID")
		}
	}

	// Create VM configuration options
	options := []proxmox.VirtualMachineOption{
		proxmox.VirtualMachineOption{
			Name:  "name",
			Value: req.Name,
		},
		proxmox.VirtualMachineOption{
			Name:  "memory",
			Value: "1024", // Default memory
		},
		proxmox.VirtualMachineOption{
			Name:  "cores",
			Value: "1", // Default cores
		},
		proxmox.VirtualMachineOption{
			Name:  "net0",
			Value: "virtio,bridge=vmbr0", // Default network
		},
		proxmox.VirtualMachineOption{
			Name:  "scsi0",
			Value: "local-lvm:8", // Default storage
		},
	}

	// Create the VM
	_, err = nodeObj.NewVirtualMachine(ctx, vmid, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	fmt.Printf("Creating Proxmox VM: %s (ID: %d) on node %s\n", req.Name, vmid, node.Node)

	return &interfaces.Instance{
		ID:        fmt.Sprintf("%d", vmid),
		Name:      req.Name,
		Status:    "stopped",
		Type:      "qemu",
		Region:    node.Node,
		CreatedAt: time.Now().Format(time.RFC3339),
		Tags:      req.Tags,
	}, nil
}

// DeleteInstance deletes a Proxmox VM
func (p *ProxmoxProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	// Get all nodes to find the VM
	nodes, err := p.client.Nodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	for _, node := range nodes {
		nodeObj, err := p.client.Node(ctx, node.Node)
		if err != nil {
			continue
		}

		// Try to get the VM to see if it exists on this node
		var vmid int
		if _, err := fmt.Sscanf(instanceID, "%d", &vmid); err != nil {
			continue // Invalid VMID format
		}

		vm, err := nodeObj.VirtualMachine(ctx, vmid)
		if err != nil {
			continue // VM not on this node
		}

		// Delete the VM
		_, err = vm.Delete(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete VM: %w", err)
		}

		fmt.Printf("Deleting Proxmox VM: %s from node %s\n", instanceID, node.Node)
		return nil
	}

	return fmt.Errorf("VM not found: %s", instanceID)
}

// GetInstanceStatus returns the status of a Proxmox VM
func (p *ProxmoxProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
	// Get all nodes to find the VM
	nodes, err := p.client.Nodes(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get nodes: %w", err)
	}

	for _, node := range nodes {
		nodeObj, err := p.client.Node(ctx, node.Node)
		if err != nil {
			continue
		}

		// Try to get the VM from this node
		var vmid int
		if _, err := fmt.Sscanf(instanceID, "%d", &vmid); err != nil {
			continue // Invalid VMID format
		}

		vm, err := nodeObj.VirtualMachine(ctx, vmid)
		if err == nil {
			return vm.Status, nil
		}
	}

	return "", fmt.Errorf("VM not found: %s", instanceID)
}

// ListRegions returns available Proxmox nodes as regions
func (p *ProxmoxProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error) {
	nodes, err := p.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var regions []interfaces.Region
	for _, node := range nodes {
		regions = append(regions, interfaces.Region{
			ID:          node.Node,
			Name:        node.Node,
			DisplayName: node.Node,
			Status:      "available",
		})
	}

	return regions, nil
}

// GetCostEstimate returns cost estimate for Proxmox resources (mock implementation)
func (p *ProxmoxProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
	// Mock cost calculation for Proxmox (usually free for self-hosted)
	var costPerHour float64
	switch req.InstanceType {
	case "qemu":
		costPerHour = 0.0 // Self-hosted, no cost
	default:
		costPerHour = 0.0
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

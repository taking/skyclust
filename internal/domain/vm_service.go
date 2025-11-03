package domain

import (
	"context"
)

// VMService defines the business logic interface for VMs
type VMService interface {
	CreateVM(ctx context.Context, req CreateVMRequest) (*VM, error)
	GetVM(ctx context.Context, id string) (*VM, error)
	UpdateVM(ctx context.Context, id string, req UpdateVMRequest) (*VM, error)
	DeleteVM(ctx context.Context, id string) error
	GetVMs(ctx context.Context, workspaceID string) ([]*VM, error)
	StartVM(ctx context.Context, id string) error
	StopVM(ctx context.Context, id string) error
	RestartVM(ctx context.Context, id string) error
	GetVMStatus(ctx context.Context, id string) (VMStatus, error)
}


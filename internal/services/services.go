package services

import (
	"fmt"

	"cmp/pkg/auth"
	"cmp/pkg/credentials"
	"cmp/pkg/database"
	"cmp/pkg/encryption"
	"cmp/pkg/events"
	"cmp/pkg/iac"
	"cmp/pkg/interfaces"
	"cmp/pkg/kubernetes"
	"cmp/pkg/realtime"
	"cmp/pkg/workspace"

	"github.com/spf13/viper"
)

// Services holds all service dependencies
type Services struct {
	Auth        auth.Service
	Workspace   workspace.Service
	Cloud       CloudService
	Credentials credentials.Service
	IaC         iac.Service
	Realtime    realtime.Service
	Kubernetes  kubernetes.Service
	Database    database.Service
	EventBus    events.Bus
	Encryption  encryption.Service
}

// CloudService defines the cloud management service
type CloudService interface {
	// Provider management
	ListProviders() ([]interfaces.CloudProvider, error)
	GetProvider(name string) (interfaces.CloudProvider, error)
	InitializeProvider(name string, config map[string]interface{}) error

	// VM management
	ListVMs(workspaceID, provider string) ([]interfaces.Instance, error)
	CreateVM(workspaceID, provider string, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error)
	GetVM(workspaceID, provider, vmID string) (*interfaces.Instance, error)
	DeleteVM(workspaceID, provider, vmID string) error
	StartVM(workspaceID, provider, vmID string) error
	StopVM(workspaceID, provider, vmID string) error

	// Region management
	ListRegions(provider string) ([]interfaces.Region, error)

	// Cost estimation
	GetCostEstimate(provider string, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error)
}

// NewServices creates a new Services instance
func NewServices(db database.Service, eventBus events.Bus, encryptionKey string) *Services {
	// Initialize encryption service
	encryptionService, err := encryption.NewService(encryptionKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize encryption service: %v", err))
	}

	// Initialize JWT secret
	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}

	authService := auth.NewService(db, jwtSecret)
	workspaceService := workspace.NewService(db)
	credentialsService := credentials.NewService(db, encryptionService)
	iacService := iac.NewService(db, eventBus)
	realtimeService := realtime.NewService(eventBus)
	cloudService := NewCloudService(credentialsService, eventBus)

	return &Services{
		Auth:        authService,
		Workspace:   workspaceService,
		Cloud:       cloudService,
		Credentials: credentialsService,
		IaC:         iacService,
		Realtime:    realtimeService,
		Kubernetes:  kubernetes.NewService(),
		Database:    db,
		EventBus:    eventBus,
		Encryption:  encryptionService,
	}
}

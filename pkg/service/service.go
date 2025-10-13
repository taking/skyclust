package service

import (
	"context"
	"fmt"
	"time"
)

// Service represents a basic service interface
type Service interface {
	// Initialize initializes the service
	Initialize(ctx context.Context) error

	// Start starts the service
	Start(ctx context.Context) error

	// Stop stops the service
	Stop(ctx context.Context) error

	// Health checks the service health
	Health(ctx context.Context) error

	// Name returns the service name
	Name() string
}

// ServiceManager manages multiple services
type ServiceManager struct {
	services map[string]Service
	status   map[string]ServiceStatus
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	Error     string    `json:"error,omitempty"`
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]Service),
		status:   make(map[string]ServiceStatus),
	}
}

// RegisterService registers a service
func (sm *ServiceManager) RegisterService(service Service) {
	sm.services[service.Name()] = service
	sm.status[service.Name()] = ServiceStatus{
		Name:   service.Name(),
		Status: "registered",
	}
}

// StartAll starts all registered services
func (sm *ServiceManager) StartAll(ctx context.Context) error {
	for name, service := range sm.services {
		if err := service.Initialize(ctx); err != nil {
			sm.status[name] = ServiceStatus{
				Name:   name,
				Status: "failed",
				Error:  err.Error(),
			}
			return err
		}

		if err := service.Start(ctx); err != nil {
			sm.status[name] = ServiceStatus{
				Name:   name,
				Status: "failed",
				Error:  err.Error(),
			}
			return err
		}

		sm.status[name] = ServiceStatus{
			Name:      name,
			Status:    "running",
			StartedAt: time.Now(),
		}
	}

	return nil
}

// StopAll stops all registered services
func (sm *ServiceManager) StopAll(ctx context.Context) error {
	for name, service := range sm.services {
		if err := service.Stop(ctx); err != nil {
			sm.status[name] = ServiceStatus{
				Name:   name,
				Status: "error",
				Error:  err.Error(),
			}
			return err
		}

		sm.status[name] = ServiceStatus{
			Name:   name,
			Status: "stopped",
		}
	}

	return nil
}

// GetStatus returns the status of all services
func (sm *ServiceManager) GetStatus() map[string]ServiceStatus {
	return sm.status
}

// GetServiceStatus returns the status of a specific service
func (sm *ServiceManager) GetServiceStatus(name string) (ServiceStatus, bool) {
	status, exists := sm.status[name]
	return status, exists
}

// Health checks the health of all services
func (sm *ServiceManager) Health(ctx context.Context) map[string]error {
	health := make(map[string]error)

	for name, service := range sm.services {
		if err := service.Health(ctx); err != nil {
			health[name] = err
		}
	}

	return health
}

// BasicService provides a basic service implementation
type BasicService struct {
	name    string
	running bool
	started time.Time
}

// NewBasicService creates a new basic service
func NewBasicService(name string) *BasicService {
	return &BasicService{
		name: name,
	}
}

// Initialize initializes the service
func (s *BasicService) Initialize(ctx context.Context) error {
	return nil
}

// Start starts the service
func (s *BasicService) Start(ctx context.Context) error {
	s.running = true
	s.started = time.Now()
	return nil
}

// Stop stops the service
func (s *BasicService) Stop(ctx context.Context) error {
	s.running = false
	return nil
}

// Health checks the service health
func (s *BasicService) Health(ctx context.Context) error {
	if !s.running {
		return fmt.Errorf("service is not running")
	}
	return nil
}

// Name returns the service name
func (s *BasicService) Name() string {
	return s.name
}

// IsRunning returns true if the service is running
func (s *BasicService) IsRunning() bool {
	return s.running
}

// GetStartedAt returns when the service was started
func (s *BasicService) GetStartedAt() time.Time {
	return s.started
}

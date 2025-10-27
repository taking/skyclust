package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ProviderManager manages gRPC connections to cloud provider services
type ProviderManager struct {
	connections map[string]*grpc.ClientConn
	configs     map[string]*ProviderConfig
	mu          sync.RWMutex

	// Connection settings
	dialTimeout   time.Duration
	maxRetries    int
	retryInterval time.Duration
}

// ProviderConfig holds configuration for a single provider
type ProviderConfig struct {
	Name        string            // Provider name (e.g., "aws", "gcp")
	Address     string            // gRPC server address (e.g., "localhost:50051")
	Enabled     bool              // Whether this provider is enabled
	Credentials map[string]string // Provider-specific credentials
	Metadata    map[string]string // Additional metadata
}

// NewProviderManager creates a new gRPC provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		connections:   make(map[string]*grpc.ClientConn),
		configs:       make(map[string]*ProviderConfig),
		dialTimeout:   10 * time.Second,
		maxRetries:    3,
		retryInterval: 2 * time.Second,
	}
}

// RegisterProvider registers a provider configuration
func (pm *ProviderManager) RegisterProvider(config *ProviderConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if config.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if config.Address == "" {
		return fmt.Errorf("provider address cannot be empty")
	}

	pm.configs[config.Name] = config
	return nil
}

// ConnectProvider establishes a gRPC connection to a provider
func (pm *ProviderManager) ConnectProvider(ctx context.Context, providerName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	config, ok := pm.configs[providerName]
	if !ok {
		return fmt.Errorf("provider %s not registered", providerName)
	}

	if !config.Enabled {
		return fmt.Errorf("provider %s is disabled", providerName)
	}

	// Check if already connected
	if conn, exists := pm.connections[providerName]; exists {
		if conn.GetState().String() == "READY" {
			return nil // Already connected
		}
		// Close stale connection
		conn.Close()
	}

	// Create connection with keepalive and retry
	conn, err := pm.dialWithRetry(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to provider %s: %w", providerName, err)
	}

	pm.connections[providerName] = conn
	return nil
}

// dialWithRetry attempts to establish a gRPC connection with retries
func (pm *ProviderManager) dialWithRetry(ctx context.Context, config *ProviderConfig) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	for attempt := 0; attempt < pm.maxRetries; attempt++ {
		conn, err = grpc.NewClient(config.Address, dialOpts...)

		if err == nil {
			return conn, nil
		}

		if attempt < pm.maxRetries-1 {
			time.Sleep(pm.retryInterval)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", pm.maxRetries, err)
}

// GetConnection returns the gRPC connection for a provider
func (pm *ProviderManager) GetConnection(providerName string) (*grpc.ClientConn, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	conn, ok := pm.connections[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not connected", providerName)
	}

	// Check connection state
	state := conn.GetState()
	if state.String() != "READY" && state.String() != "IDLE" {
		return nil, fmt.Errorf("provider %s connection not ready (state: %s)", providerName, state)
	}

	return conn, nil
}

// DisconnectProvider closes the gRPC connection to a provider
func (pm *ProviderManager) DisconnectProvider(providerName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	conn, ok := pm.connections[providerName]
	if !ok {
		return fmt.Errorf("provider %s not connected", providerName)
	}

	err := conn.Close()
	delete(pm.connections, providerName)

	return err
}

// DisconnectAll closes all provider connections
func (pm *ProviderManager) DisconnectAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errs []error
	for name, conn := range pm.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close %s: %w", name, err))
		}
	}

	pm.connections = make(map[string]*grpc.ClientConn)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// ListProviders returns a list of registered provider names
func (pm *ProviderManager) ListProviders() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]string, 0, len(pm.configs))
	for name := range pm.configs {
		providers = append(providers, name)
	}

	return providers
}

// ListConnectedProviders returns a list of connected provider names
func (pm *ProviderManager) ListConnectedProviders() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]string, 0, len(pm.connections))
	for name, conn := range pm.connections {
		if conn.GetState().String() == "READY" {
			providers = append(providers, name)
		}
	}

	return providers
}

// GetProviderConfig returns the configuration for a provider
func (pm *ProviderManager) GetProviderConfig(providerName string) (*ProviderConfig, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	config, ok := pm.configs[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", providerName)
	}

	return config, nil
}

// IsConnected checks if a provider is connected and ready
func (pm *ProviderManager) IsConnected(providerName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	conn, ok := pm.connections[providerName]
	if !ok {
		return false
	}

	return conn.GetState().String() == "READY"
}

// HealthCheck performs a health check on all connected providers
func (pm *ProviderManager) HealthCheck(ctx context.Context) map[string]bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	results := make(map[string]bool)
	for name, conn := range pm.connections {
		state := conn.GetState()
		results[name] = state.String() == "READY" || state.String() == "IDLE"
	}

	return results
}

// SetDialTimeout sets the connection dial timeout
func (pm *ProviderManager) SetDialTimeout(timeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.dialTimeout = timeout
}

// SetMaxRetries sets the maximum number of connection retries
func (pm *ProviderManager) SetMaxRetries(retries int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.maxRetries = retries
}

// SetRetryInterval sets the interval between connection retries
func (pm *ProviderManager) SetRetryInterval(interval time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.retryInterval = interval
}

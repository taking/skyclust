package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthChecker performs periodic health checks on provider connections
type HealthChecker struct {
	manager       *ProviderManager
	interval      time.Duration
	timeout       time.Duration
	mu            sync.RWMutex
	statuses      map[string]HealthStatus
	stopCh        chan struct{}
	running       bool
	reconnectMode bool
}

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Healthy          bool
	LastCheck        time.Time
	LastError        error
	ConnState        string
	CheckCount       int
	FailureCount     int
	ConsecutiveFails int
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(manager *ProviderManager, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		manager:       manager,
		interval:      interval,
		timeout:       timeout,
		statuses:      make(map[string]HealthStatus),
		stopCh:        make(chan struct{}),
		reconnectMode: true,
	}
}

// Start begins periodic health checking
func (hc *HealthChecker) Start() {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.mu.Unlock()

	go hc.healthCheckLoop()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if !hc.running {
		return
	}

	hc.running = false
	close(hc.stopCh)
}

// healthCheckLoop runs the health check periodically
func (hc *HealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllProviders()
		case <-hc.stopCh:
			return
		}
	}
}

// checkAllProviders checks health of all registered providers
func (hc *HealthChecker) checkAllProviders() {
	providers := hc.manager.ListProviders()

	for _, providerName := range providers {
		hc.checkProvider(providerName)
	}
}

// checkProvider checks health of a single provider
func (hc *HealthChecker) checkProvider(providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	hc.mu.Lock()
	status := hc.statuses[providerName]
	status.CheckCount++
	status.LastCheck = time.Now()
	hc.mu.Unlock()

	// Get connection
	conn, err := hc.manager.GetConnection(providerName)
	if err != nil {
		hc.updateStatus(providerName, false, err, "UNKNOWN")
		hc.handleUnhealthy(providerName)
		return
	}

	// Check connection state
	connState := conn.GetState()
	status.ConnState = connState.String()

	// Perform gRPC health check
	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})

	if err != nil {
		hc.updateStatus(providerName, false, err, connState.String())
		hc.handleUnhealthy(providerName)
		return
	}

	healthy := resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
	hc.updateStatus(providerName, healthy, nil, connState.String())

	if !healthy {
		hc.handleUnhealthy(providerName)
	}
}

// updateStatus updates the health status of a provider
func (hc *HealthChecker) updateStatus(providerName string, healthy bool, err error, connState string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	status := hc.statuses[providerName]
	status.Healthy = healthy
	status.LastCheck = time.Now()
	status.LastError = err
	status.ConnState = connState

	if !healthy {
		status.FailureCount++
		status.ConsecutiveFails++
	} else {
		status.ConsecutiveFails = 0
	}

	hc.statuses[providerName] = status
}

// handleUnhealthy handles an unhealthy provider
func (hc *HealthChecker) handleUnhealthy(providerName string) {
	if !hc.reconnectMode {
		return
	}

	hc.mu.RLock()
	status := hc.statuses[providerName]
	hc.mu.RUnlock()

	// Attempt reconnection after 3 consecutive failures
	if status.ConsecutiveFails >= 3 {
		go hc.attemptReconnect(providerName)
	}
}

// attemptReconnect attempts to reconnect to a provider
func (hc *HealthChecker) attemptReconnect(providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Disconnect first
	_ = hc.manager.DisconnectProvider(providerName)

	// Wait a bit before reconnecting
	time.Sleep(2 * time.Second)

	// Attempt reconnection
	err := hc.manager.ConnectProvider(ctx, providerName)
	if err != nil {
		fmt.Printf("Failed to reconnect to provider %s: %v\n", providerName, err)
		return
	}

	// Reset consecutive failures on successful reconnection
	hc.mu.Lock()
	status := hc.statuses[providerName]
	status.ConsecutiveFails = 0
	hc.statuses[providerName] = status
	hc.mu.Unlock()

	fmt.Printf("Successfully reconnected to provider %s\n", providerName)
}

// GetStatus returns the health status of a provider
func (hc *HealthChecker) GetStatus(providerName string) (HealthStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, ok := hc.statuses[providerName]
	return status, ok
}

// GetAllStatuses returns all provider health statuses
func (hc *HealthChecker) GetAllStatuses() map[string]HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	statuses := make(map[string]HealthStatus)
	for name, status := range hc.statuses {
		statuses[name] = status
	}

	return statuses
}

// IsHealthy returns whether a provider is currently healthy
func (hc *HealthChecker) IsHealthy(providerName string) bool {
	status, ok := hc.GetStatus(providerName)
	if !ok {
		return false
	}
	return status.Healthy
}

// SetReconnectMode enables or disables automatic reconnection
func (hc *HealthChecker) SetReconnectMode(enabled bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.reconnectMode = enabled
}

// ConnectionStateMonitor monitors connection state changes
type ConnectionStateMonitor struct {
	manager *ProviderManager
	stopCh  chan struct{}
	running bool
	mu      sync.Mutex
}

// NewConnectionStateMonitor creates a new connection state monitor
func NewConnectionStateMonitor(manager *ProviderManager) *ConnectionStateMonitor {
	return &ConnectionStateMonitor{
		manager: manager,
		stopCh:  make(chan struct{}),
	}
}

// Start begins monitoring connection states
func (csm *ConnectionStateMonitor) Start() {
	csm.mu.Lock()
	if csm.running {
		csm.mu.Unlock()
		return
	}
	csm.running = true
	csm.mu.Unlock()

	go csm.monitorLoop()
}

// Stop stops the connection state monitor
func (csm *ConnectionStateMonitor) Stop() {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.running {
		return
	}

	csm.running = false
	close(csm.stopCh)
}

// monitorLoop monitors connection states
func (csm *ConnectionStateMonitor) monitorLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			csm.checkConnectionStates()
		case <-csm.stopCh:
			return
		}
	}
}

// checkConnectionStates checks the state of all connections
func (csm *ConnectionStateMonitor) checkConnectionStates() {
	providers := csm.manager.ListConnectedProviders()

	for _, providerName := range providers {
		conn, err := csm.manager.GetConnection(providerName)
		if err != nil {
			continue
		}

		state := conn.GetState()

		// Handle specific states
		switch state {
		case connectivity.Shutdown:
			fmt.Printf("Provider %s connection is shutdown, attempting reconnect\n", providerName)
			go csm.attemptReconnect(providerName)
		case connectivity.TransientFailure:
			fmt.Printf("Provider %s experiencing transient failure\n", providerName)
		}
	}
}

// attemptReconnect attempts to reconnect a provider
func (csm *ConnectionStateMonitor) attemptReconnect(providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = csm.manager.DisconnectProvider(providerName)
	time.Sleep(2 * time.Second)

	err := csm.manager.ConnectProvider(ctx, providerName)
	if err != nil {
		fmt.Printf("Failed to reconnect provider %s: %v\n", providerName, err)
		return
	}

	fmt.Printf("Successfully reconnected provider %s\n", providerName)
}

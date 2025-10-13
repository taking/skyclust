package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"skyclust/pkg/logger"
)

// Connection represents a connection in the pool
type Connection interface {
	Close() error
	Health() error
	ID() string
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxConnections      int
	MinConnections      int
	MaxIdleTime         time.Duration
	AcquireTimeout      time.Duration
	HealthCheckInterval time.Duration
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConnections:      10,
		MinConnections:      2,
		MaxIdleTime:         5 * time.Minute,
		AcquireTimeout:      30 * time.Second,
		HealthCheckInterval: 1 * time.Minute,
	}
}

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	config      PoolConfig
	factory     ConnectionFactory
	connections chan Connection
	active      map[string]Connection
	mu          sync.RWMutex
	closed      bool
	ctx         context.Context
	cancel      context.CancelFunc
	stats       PoolStats
}

// ConnectionFactory creates new connections
type ConnectionFactory interface {
	CreateConnection(ctx context.Context) (Connection, error)
}

// PoolStats holds pool statistics
type PoolStats struct {
	TotalConnections  int
	ActiveConnections int
	IdleConnections   int
	AcquiredCount     int64
	ReleasedCount     int64
	ErrorCount        int64
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config PoolConfig, factory ConnectionFactory) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:      config,
		factory:     factory,
		connections: make(chan Connection, config.MaxConnections),
		active:      make(map[string]Connection),
		ctx:         ctx,
		cancel:      cancel,
		stats:       PoolStats{},
	}

	// Start health check goroutine
	go pool.healthCheck()

	// Pre-populate with minimum connections
	go pool.prePopulate()

	return pool
}

// Acquire gets a connection from the pool
func (p *ConnectionPool) Acquire(ctx context.Context) (Connection, error) {
	select {
	case conn := <-p.connections:
		// Check if connection is still healthy
		if err := conn.Health(); err != nil {
			logger.Warnf("Unhealthy connection %s, creating new one", conn.ID())
			return p.createNewConnection(ctx)
		}

		p.mu.Lock()
		p.active[conn.ID()] = conn
		p.stats.ActiveConnections++
		p.stats.AcquiredCount++
		p.mu.Unlock()

		return conn, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-time.After(p.config.AcquireTimeout):
		return nil, fmt.Errorf("timeout acquiring connection")
	}
}

// Release returns a connection to the pool
func (p *ConnectionPool) Release(conn Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return conn.Close()
	}

	// Check if connection is still healthy
	if err := conn.Health(); err != nil {
		logger.Warnf("Unhealthy connection %s, closing", conn.ID())
		p.stats.ErrorCount++
		return conn.Close()
	}

	// Remove from active connections
	delete(p.active, conn.ID())
	p.stats.ActiveConnections--
	p.stats.ReleasedCount++

	// Try to return to pool
	select {
	case p.connections <- conn:
		p.stats.IdleConnections++
		return nil
	default:
		// Pool is full, close the connection
		return conn.Close()
	}
}

// Close closes the connection pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	p.cancel()

	// Close all active connections
	for _, conn := range p.active {
		conn.Close()
	}

	// Close all idle connections
	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}

	return nil
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	stats.TotalConnections = len(p.connections) + len(p.active)
	stats.IdleConnections = len(p.connections)
	stats.ActiveConnections = len(p.active)

	return stats
}

// createNewConnection creates a new connection
func (p *ConnectionPool) createNewConnection(ctx context.Context) (Connection, error) {
	conn, err := p.factory.CreateConnection(ctx)
	if err != nil {
		p.mu.Lock()
		p.stats.ErrorCount++
		p.mu.Unlock()
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	p.mu.Lock()
	p.stats.TotalConnections++
	p.mu.Unlock()

	return conn, nil
}

// prePopulate creates minimum connections
func (p *ConnectionPool) prePopulate() {
	for i := 0; i < p.config.MinConnections; i++ {
		conn, err := p.createNewConnection(p.ctx)
		if err != nil {
			logger.Errorf("Failed to pre-populate connection: %v", err)
			continue
		}

		select {
		case p.connections <- conn:
			p.mu.Lock()
			p.stats.IdleConnections++
			p.mu.Unlock()
		default:
			conn.Close()
		}
	}
}

// healthCheck performs periodic health checks
func (p *ConnectionPool) healthCheck() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck checks the health of all connections
func (p *ConnectionPool) performHealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check active connections
	for id, conn := range p.active {
		if err := conn.Health(); err != nil {
			logger.Warnf("Unhealthy active connection %s: %v", id, err)
			conn.Close()
			delete(p.active, id)
			p.stats.ActiveConnections--
			p.stats.ErrorCount++
		}
	}

	// Check idle connections
	unhealthy := make([]Connection, 0)
	healthy := make([]Connection, 0)

	// Drain all connections
	for {
		select {
		case conn := <-p.connections:
			if err := conn.Health(); err != nil {
				unhealthy = append(unhealthy, conn)
			} else {
				healthy = append(healthy, conn)
			}
		default:
			goto done
		}
	}
done:

	// Close unhealthy connections
	for _, conn := range unhealthy {
		conn.Close()
		p.stats.ErrorCount++
	}

	// Return healthy connections
	for _, conn := range healthy {
		select {
		case p.connections <- conn:
		default:
			conn.Close()
		}
	}

	// Ensure minimum connections
	for len(p.connections) < p.config.MinConnections {
		conn, err := p.createNewConnection(p.ctx)
		if err != nil {
			logger.Errorf("Failed to create connection during health check: %v", err)
			break
		}

		select {
		case p.connections <- conn:
		default:
			conn.Close()
		}
	}
}

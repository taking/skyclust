package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"
	"time"

	"skyclust/internal/plugin/interfaces"
	"skyclust/pkg/logger"
)

// AsyncManager handles asynchronous plugin loading and management
type AsyncManager struct {
	plugins     map[string]interfaces.CloudProvider
	configs     map[string]map[string]interface{}
	loading     map[string]bool
	errors      map[string]error
	mu          sync.RWMutex
	loadTimeout time.Duration
}

// NewAsyncManager creates a new async plugin manager
func NewAsyncManager() *AsyncManager {
	return &AsyncManager{
		plugins:     make(map[string]interfaces.CloudProvider),
		configs:     make(map[string]map[string]interface{}),
		loading:     make(map[string]bool),
		errors:      make(map[string]error),
		loadTimeout: 30 * time.Second,
	}
}

// LoadPluginsAsync loads plugins asynchronously
func (m *AsyncManager) LoadPluginsAsync(pluginDir string) <-chan LoadResult {
	resultChan := make(chan LoadResult, 10)

	go func() {
		defer close(resultChan)

		// Load plugins from public directory
		publicDir := filepath.Join(pluginDir, "public")
		if _, err := os.Stat(publicDir); !os.IsNotExist(err) {
			m.loadPluginsFromDirAsync(publicDir, "public", resultChan)
		}

		// Load plugins from private directory
		privateDir := filepath.Join(pluginDir, "private")
		if _, err := os.Stat(privateDir); !os.IsNotExist(err) {
			m.loadPluginsFromDirAsync(privateDir, "private", resultChan)
		}
	}()

	return resultChan
}

// LoadResult represents the result of a plugin load operation
type LoadResult struct {
	PluginName string
	Success    bool
	Error      error
	Provider   interfaces.CloudProvider
}

// loadPluginsFromDirAsync loads plugins from a directory asynchronously
func (m *AsyncManager) loadPluginsFromDirAsync(dir, category string, resultChan chan<- LoadResult) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.so files
		if info.IsDir() || !strings.HasSuffix(path, ".so") {
			return nil
		}

		// Load the plugin asynchronously
		pluginName := strings.TrimSuffix(filepath.Base(path), ".so")
		if category != "" {
			pluginName = category + "/" + pluginName
		}

		go m.loadPluginAsync(path, pluginName, resultChan)

		return nil
	})
}

// loadPluginAsync loads a single plugin asynchronously
func (m *AsyncManager) loadPluginAsync(pluginPath, pluginName string, resultChan chan<- LoadResult) {
	// Set loading state
	m.mu.Lock()
	m.loading[pluginName] = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.loading, pluginName)
		m.mu.Unlock()
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.loadTimeout)
	defer cancel()

	// Load the plugin
	provider, err := m.loadPluginWithContext(ctx, pluginPath, pluginName)

	result := LoadResult{
		PluginName: pluginName,
		Success:    err == nil,
		Error:      err,
		Provider:   provider,
	}

	// Store result
	if err == nil {
		m.mu.Lock()
		m.plugins[pluginName] = provider
		delete(m.errors, pluginName)
		m.mu.Unlock()

		logger.Infof("Successfully loaded plugin: %s", pluginName)
	} else {
		m.mu.Lock()
		m.errors[pluginName] = err
		m.mu.Unlock()

		logger.Errorf("Failed to load plugin %s: %v", pluginName, err)
	}

	// Send result
	select {
	case resultChan <- result:
	case <-ctx.Done():
		logger.Warnf("Plugin load result for %s not sent due to timeout", pluginName)
	}
}

// loadPluginWithContext loads a plugin with context support
func (m *AsyncManager) loadPluginWithContext(ctx context.Context, pluginPath, pluginName string) (interfaces.CloudProvider, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Open the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", pluginName, err)
	}

	// Look for the New function
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("plugin %s does not export 'New' function: %w", pluginName, err)
	}

	// Type assert to the expected function signature
	newProvider, ok := newFunc.(func() interfaces.CloudProvider)
	if !ok {
		return nil, fmt.Errorf("plugin %s 'New' function has wrong signature", pluginName)
	}

	// Create the provider instance
	provider := newProvider()

	return provider, nil
}

// GetProvider returns a cloud provider by name
func (m *AsyncManager) GetProvider(name string) (interfaces.CloudProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.plugins[name]
	if !exists {
		// Check if it's still loading
		if m.loading[name] {
			return nil, fmt.Errorf("plugin %s is still loading", name)
		}

		// Check if there was an error
		if err, hasError := m.errors[name]; hasError {
			return nil, fmt.Errorf("plugin %s failed to load: %w", name, err)
		}

		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return provider, nil
}

// ListProviders returns a list of loaded provider names
func (m *AsyncManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		providers = append(providers, name)
	}
	return providers
}

// InitializeProvider initializes a provider with configuration
func (m *AsyncManager) InitializeProvider(name string, config map[string]interface{}) error {
	provider, err := m.GetProvider(name)
	if err != nil {
		return err
	}

	// Store the configuration
	m.mu.Lock()
	m.configs[name] = config
	m.mu.Unlock()

	// Initialize the provider
	return provider.Initialize(config)
}

// GetProviderInfo returns information about a provider
func (m *AsyncManager) GetProviderInfo(name string) (map[string]string, error) {
	provider, err := m.GetProvider(name)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"name":    provider.GetName(),
		"version": provider.GetVersion(),
	}, nil
}

// GetLoadingStatus returns the loading status of all plugins
func (m *AsyncManager) GetLoadingStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)

	// Check loaded plugins
	for name := range m.plugins {
		status[name] = "loaded"
	}

	// Check loading plugins
	for name := range m.loading {
		status[name] = "loading"
	}

	// Check failed plugins
	for name := range m.errors {
		status[name] = "failed"
	}

	return status
}

// GetErrors returns all plugin loading errors
func (m *AsyncManager) GetErrors() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errors := make(map[string]error)
	for name, err := range m.errors {
		errors[name] = err
	}
	return errors
}

// SetLoadTimeout sets the timeout for plugin loading
func (m *AsyncManager) SetLoadTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadTimeout = timeout
}

// WaitForPlugin waits for a specific plugin to load
func (m *AsyncManager) WaitForPlugin(name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for plugin %s", name)
		case <-ticker.C:
			m.mu.RLock()
			_, loaded := m.plugins[name]
			_, loading := m.loading[name]
			err, hasError := m.errors[name]
			m.mu.RUnlock()

			if loaded {
				return nil
			}
			if hasError {
				return fmt.Errorf("plugin %s failed to load: %w", name, err)
			}
			if !loading {
				return fmt.Errorf("plugin %s not found", name)
			}
		}
	}
}

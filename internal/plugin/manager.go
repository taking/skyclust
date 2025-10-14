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

	plugininterfaces "skyclust/pkg/plugin"
)

// Manager handles loading and managing cloud provider plugins
type Manager struct {
	plugins     map[string]plugininterfaces.CloudProvider
	configs     map[string]map[string]interface{}
	loading     map[string]bool
	errors      map[string]error
	mu          sync.RWMutex
	loadTimeout time.Duration
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins:     make(map[string]plugininterfaces.CloudProvider),
		configs:     make(map[string]map[string]interface{}),
		loading:     make(map[string]bool),
		errors:      make(map[string]error),
		loadTimeout: 30 * time.Second,
	}
}

// LoadPlugins loads all plugins from the specified directory (synchronous)
func (m *Manager) LoadPlugins(pluginDir string) error {
	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory %s does not exist", pluginDir)
	}

	// Load plugins from public directory
	publicDir := filepath.Join(pluginDir, "public")
	if _, err := os.Stat(publicDir); !os.IsNotExist(err) {
		if err := m.loadPluginsFromDir(publicDir, "public"); err != nil {
			fmt.Printf("Warning: failed to load public plugins: %v\n", err)
		}
	}

	// Load plugins from private directory
	privateDir := filepath.Join(pluginDir, "private")
	if _, err := os.Stat(privateDir); !os.IsNotExist(err) {
		if err := m.loadPluginsFromDir(privateDir, "private"); err != nil {
			fmt.Printf("Warning: failed to load private plugins: %v\n", err)
		}
	}

	return nil
}

// LoadPluginsAsync loads all plugins asynchronously
func (m *Manager) LoadPluginsAsync(pluginDir string) <-chan LoadResult {
	resultChan := make(chan LoadResult, 10)

	go func() {
		defer close(resultChan)

		// Check if plugin directory exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			resultChan <- LoadResult{
				PluginName: "all",
				Success:    false,
				Error:      fmt.Errorf("plugin directory %s does not exist", pluginDir),
			}
			return
		}

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
	Provider   plugininterfaces.CloudProvider
}

// loadPluginsFromDir loads plugins from a specific directory (synchronous)
func (m *Manager) loadPluginsFromDir(dir, category string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.so files
		if info.IsDir() || !strings.HasSuffix(path, ".so") {
			return nil
		}

		// Load the plugin
		pluginName := strings.TrimSuffix(filepath.Base(path), ".so")
		if category != "" {
			pluginName = category + "/" + pluginName
		}

		// Check if plugin is already loaded to prevent duplicates
		if _, exists := m.plugins[pluginName]; exists {
			fmt.Printf("Plugin %s already loaded, skipping\n", pluginName)
			return nil
		}

		if err := m.loadPlugin(path, pluginName); err != nil {
			fmt.Printf("Warning: failed to load plugin %s: %v\n", pluginName, err)
			return nil // Continue loading other plugins
		}

		return nil
	})
}

// loadPluginsFromDirAsync loads plugins from a specific directory asynchronously
func (m *Manager) loadPluginsFromDirAsync(dir, category string, resultChan chan<- LoadResult) {
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
func (m *Manager) loadPluginAsync(pluginPath, pluginName string, resultChan chan<- LoadResult) {
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

		fmt.Printf("Successfully loaded plugin: %s\n", pluginName)
	} else {
		m.mu.Lock()
		m.errors[pluginName] = err
		m.mu.Unlock()

		fmt.Printf("Failed to load plugin %s: %v\n", pluginName, err)
	}

	// Send result
	select {
	case resultChan <- result:
	case <-ctx.Done():
		fmt.Printf("Timeout loading plugin %s\n", pluginName)
	}
}

// loadPluginWithContext loads a plugin with context support
func (m *Manager) loadPluginWithContext(ctx context.Context, pluginPath, pluginName string) (plugininterfaces.CloudProvider, error) {
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
	newProvider, ok := newFunc.(func() plugininterfaces.CloudProvider)
	if !ok {
		return nil, fmt.Errorf("plugin %s 'New' function has wrong signature", pluginName)
	}

	// Create the provider instance
	provider := newProvider()

	return provider, nil
}

// loadPlugin loads a single plugin from the given path
func (m *Manager) loadPlugin(pluginPath, pluginName string) error {
	// Open the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", pluginName, err)
	}

	// Look for the New function
	newFunc, err := p.Lookup("New")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'New' function: %w", pluginName, err)
	}

	// Type assert to the expected function signature
	newProvider, ok := newFunc.(func() plugininterfaces.CloudProvider)
	if !ok {
		return fmt.Errorf("plugin %s 'New' function has wrong signature", pluginName)
	}

	// Create the provider instance
	provider := newProvider()

	// Store the plugin
	m.plugins[pluginName] = provider

	fmt.Printf("Successfully loaded plugin: %s\n", pluginName)
	return nil
}

// GetProvider returns a cloud provider by name
func (m *Manager) GetProvider(name string) (plugininterfaces.CloudProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try exact match first
	if provider, exists := m.plugins[name]; exists {
		return provider, nil
	}

	// Check if it's still loading
	if m.loading[name] {
		return nil, fmt.Errorf("plugin %s is still loading", name)
	}

	// Check if there was an error
	if err, hasError := m.errors[name]; hasError {
		return nil, fmt.Errorf("plugin %s failed to load: %w", name, err)
	}

	// Try to find by base name (e.g., "aws" -> "public/aws")
	for pluginName, provider := range m.plugins {
		baseName := filepath.Base(pluginName)
		if baseName == name {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

// ListProviders returns a list of loaded provider names
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		providers = append(providers, name)
	}
	return providers
}

// InitializeProvider initializes a provider with configuration
func (m *Manager) InitializeProvider(name string, config map[string]interface{}) error {
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
func (m *Manager) GetProviderInfo(name string) (map[string]string, error) {
	provider, err := m.GetProvider(name)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"name":    provider.GetName(),
		"version": provider.GetVersion(),
	}, nil
}

// ExecuteProviderMethod executes a method on a provider
func (m *Manager) ExecuteProviderMethod(ctx context.Context, providerName, method string, args ...interface{}) (interface{}, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// This is a simplified version - in a real implementation, you'd want
	// a more sophisticated method dispatch system
	switch method {
	case "ListInstances":
		return provider.ListInstances(ctx)
	case "ListRegions":
		return provider.ListRegions(ctx)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

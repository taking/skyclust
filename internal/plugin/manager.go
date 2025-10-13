package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"skyclust/internal/plugin/interfaces"
)

// Manager handles loading and managing cloud provider plugins
type Manager struct {
	plugins map[string]interfaces.CloudProvider
	configs map[string]map[string]interface{}
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]interfaces.CloudProvider),
		configs: make(map[string]map[string]interface{}),
	}
}

// LoadPlugins loads all plugins from the specified directory
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

	// Load plugins from root directory (for backward compatibility)
	// Skip root directory loading to prevent duplicates
	// return m.loadPluginsFromDir(pluginDir, "")
	return nil
}

// loadPluginsFromDir loads plugins from a specific directory
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
	newProvider, ok := newFunc.(func() interfaces.CloudProvider)
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
func (m *Manager) GetProvider(name string) (interfaces.CloudProvider, error) {
	// Try exact match first
	if provider, exists := m.plugins[name]; exists {
		return provider, nil
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
	m.configs[name] = config

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

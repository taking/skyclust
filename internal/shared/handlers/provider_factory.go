package handlers

import (
	"fmt"
)

// ProviderFactory defines the common interface for all provider handler factories
type ProviderFactory[T any] struct {
	handlers map[string]T
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory[T any]() *ProviderFactory[T] {
	return &ProviderFactory[T]{
		handlers: make(map[string]T),
	}
}

// Register registers a provider handler
func (f *ProviderFactory[T]) Register(provider string, handler T) {
	f.handlers[provider] = handler
}

// GetHandler returns the handler for a specific provider
func (f *ProviderFactory[T]) GetHandler(provider string) (T, error) {
	handler, exists := f.handlers[provider]
	if !exists {
		var zero T
		return zero, fmt.Errorf("unsupported provider: %s", provider)
	}
	return handler, nil
}

// GetAllProviders returns a list of all registered providers
func (f *ProviderFactory[T]) GetAllProviders() []string {
	providers := make([]string, 0, len(f.handlers))
	for provider := range f.handlers {
		providers = append(providers, provider)
	}
	return providers
}

// IsProviderSupported checks if a provider is supported
func (f *ProviderFactory[T]) IsProviderSupported(provider string) bool {
	_, exists := f.handlers[provider]
	return exists
}

// GetHandlerCount returns the number of registered handlers
func (f *ProviderFactory[T]) GetHandlerCount() int {
	return len(f.handlers)
}

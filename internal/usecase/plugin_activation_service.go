package usecase

import (
	"context"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"

	"github.com/google/uuid"
)

// pluginActivationService implements the plugin activation business logic
type pluginActivationService struct {
	credentialRepo domain.CredentialRepository
	eventBus       messaging.Bus
}

// NewPluginActivationService creates a new plugin activation service
func NewPluginActivationService(credentialRepo domain.CredentialRepository, eventBus messaging.Bus) domain.PluginActivationService {
	return &pluginActivationService{
		credentialRepo: credentialRepo,
		eventBus:       eventBus,
	}
}

// ActivatePlugin activates a plugin for a user
func (s *pluginActivationService) ActivatePlugin(ctx context.Context, userID uuid.UUID, provider string) error {
	// Check if user has credentials for this provider
	credentials, err := s.credentialRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	hasCredentials := false
	for _, cred := range credentials {
		if cred.Provider == provider {
			hasCredentials = true
			break
		}
	}

	if !hasCredentials {
		return domain.ErrNoActiveCredentials
	}

	// Publish plugin activation event
	event := messaging.Event{
		Type:   "plugin_activated",
		UserID: userID.String(),
		Data: map[string]interface{}{
			"provider": provider,
			"user_id":  userID.String(),
		},
	}

	return s.eventBus.Publish(context.Background(), event)
}

// DeactivatePlugin deactivates a plugin for a user
func (s *pluginActivationService) DeactivatePlugin(ctx context.Context, userID uuid.UUID, provider string) error {
	// Publish plugin deactivation event
	event := messaging.Event{
		Type:   "plugin_deactivated",
		UserID: userID.String(),
		Data: map[string]interface{}{
			"provider": provider,
			"user_id":  userID.String(),
		},
	}

	return s.eventBus.Publish(context.Background(), event)
}

// GetActivePlugins returns the list of active plugins for a user
func (s *pluginActivationService) GetActivePlugins(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Get user's credentials to determine active providers
	credentials, err := s.credentialRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	providers := make([]string, 0)
	providerMap := make(map[string]bool)

	for _, cred := range credentials {
		if !providerMap[cred.Provider] {
			providers = append(providers, cred.Provider)
			providerMap[cred.Provider] = true
		}
	}

	return providers, nil
}

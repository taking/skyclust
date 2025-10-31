package service

import (
	"context"
	"skyclust/internal/domain"

	"github.com/google/uuid"
)

// pluginActivationService implements the plugin activation business logic
type pluginActivationService struct {
	credentialRepo domain.CredentialRepository
	eventService   domain.EventService
}

// NewPluginActivationService creates a new plugin activation service
func NewPluginActivationService(credentialRepo domain.CredentialRepository, eventService domain.EventService) domain.PluginActivationService {
	return &pluginActivationService{
		credentialRepo: credentialRepo,
		eventService:   eventService,
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
	if s.eventService != nil {
		if err := s.eventService.Publish(context.Background(), "plugin.activated", map[string]interface{}{
			"provider": provider,
			"user_id":  userID.String(),
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeactivatePlugin deactivates a plugin for a user
func (s *pluginActivationService) DeactivatePlugin(ctx context.Context, userID uuid.UUID, provider string) error {
	// Publish plugin deactivation event
	if s.eventService != nil {
		if err := s.eventService.Publish(context.Background(), "plugin.deactivated", map[string]interface{}{
			"provider": provider,
			"user_id":  userID.String(),
		}); err != nil {
			return err
		}
	}
	return nil
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

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Publisher provides high-level event publishing functionality
type Publisher struct {
	bus    Bus
	logger *zap.Logger
}

// NewPublisher creates a new event publisher
func NewPublisher(bus Bus, logger *zap.Logger) *Publisher {
	return &Publisher{
		bus:    bus,
		logger: logger,
	}
}

// PublishKubernetesEvent publishes a Kubernetes resource event
func (p *Publisher) PublishKubernetesEvent(ctx context.Context, provider, credentialID, region, resource, action string, data map[string]interface{}) error {
	topic := BuildKubernetesTopic(provider, credentialID, region, resource, action)

	event := Event{
		Type:      topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Kubernetes event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.Error(err))
		// Don't fail the operation if event publishing fails
		// This allows graceful degradation when NATS is unavailable
		return nil
	}

	p.logger.Debug("Published Kubernetes event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("action", action))

	return nil
}

// PublishKubernetesClusterEvent publishes a Kubernetes cluster event
func (p *Publisher) PublishKubernetesClusterEvent(ctx context.Context, provider, credentialID, region, action string, clusterData map[string]interface{}) error {
	data := map[string]interface{}{
		"provider":      provider,
		"credential_id": credentialID,
		"region":        region,
		"action":        action,
	}

	for k, v := range clusterData {
		data[k] = v
	}

	return p.PublishKubernetesEvent(ctx, provider, credentialID, region, "clusters", action, data)
}

// PublishNetworkEvent publishes a Network resource event
func (p *Publisher) PublishNetworkEvent(ctx context.Context, provider, credentialID, region, resource, action string, data map[string]interface{}) error {
	topic := BuildNetworkTopic(provider, credentialID, region, resource, action)

	event := Event{
		Type:      topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Network event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.Error(err))
		// Don't fail the operation if event publishing fails
		// This allows graceful degradation when NATS is unavailable
		return nil
	}

	p.logger.Debug("Published Network event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("action", action))

	return nil
}

// PublishVPCEvent publishes a VPC event
func (p *Publisher) PublishVPCEvent(ctx context.Context, provider, credentialID, region, action string, vpcData map[string]interface{}) error {
	data := map[string]interface{}{
		"provider":      provider,
		"credential_id": credentialID,
		"region":        region,
		"action":        action,
	}

	for k, v := range vpcData {
		data[k] = v
	}

	return p.PublishNetworkEvent(ctx, provider, credentialID, region, "vpcs", action, data)
}

// PublishVMEvent publishes a VM event
func (p *Publisher) PublishVMEvent(ctx context.Context, provider, workspaceID, vmID, action string, vmData map[string]interface{}) error {
	topic := BuildVMTopic(provider, workspaceID, "", action)

	event := Event{
		Type:      topic,
		Data:      vmData,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish VM event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.String("vm_id", vmID),
			zap.Error(err))
		// Don't fail the operation if event publishing fails
		// This allows graceful degradation when NATS is unavailable
		return nil
	}

	p.logger.Debug("Published VM event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("vm_id", vmID),
		zap.String("action", action))

	return nil
}

// PublishToNATS publishes an event directly to NATS (for NATSService)
func (p *Publisher) PublishToNATS(ctx context.Context, subject string, data interface{}) error {
	if natsService, ok := p.bus.(*NATSService); ok {
		return natsService.PublishMessage(ctx, subject, data)
	}

	// For other bus types, convert to Event format
	eventData, err := p.convertToEventData(data)
	if err != nil {
		return fmt.Errorf("failed to convert data to event format: %w", err)
	}

	event := Event{
		Type:      subject,
		Data:      eventData,
		Timestamp: time.Now().Unix(),
	}

	return p.bus.Publish(ctx, event)
}

// convertToEventData converts any data to map[string]interface{}
func (p *Publisher) convertToEventData(data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var eventData map[string]interface{}
	if err := json.Unmarshal(jsonData, &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return eventData, nil
}

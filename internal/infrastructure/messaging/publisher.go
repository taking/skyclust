package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"skyclust/internal/domain"

	"github.com/google/uuid"
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

// PublishKubernetesNodePoolEvent publishes a Kubernetes node pool event
func (p *Publisher) PublishKubernetesNodePoolEvent(ctx context.Context, provider, credentialID, clusterName, action string, nodePoolData map[string]interface{}) error {
	topic := BuildKubernetesNodePoolTopic(provider, credentialID, clusterName, action)

	data := map[string]interface{}{
		"provider":      provider,
		"credential_id": credentialID,
		"cluster_name":  clusterName,
		"action":        action,
	}

	for k, v := range nodePoolData {
		data[k] = v
	}

	event := Event{
		Type:      topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Kubernetes node pool event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.String("cluster_name", clusterName),
			zap.Error(err))
		return nil
	}

	p.logger.Debug("Published Kubernetes node pool event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("cluster_name", clusterName),
		zap.String("action", action))

	return nil
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

// PublishSubnetEvent publishes a Subnet event
func (p *Publisher) PublishSubnetEvent(ctx context.Context, provider, credentialID, vpcID, action string, subnetData map[string]interface{}) error {
	topic := BuildNetworkSubnetTopic(provider, credentialID, vpcID, action)

	data := map[string]interface{}{
		"provider":      provider,
		"credential_id": credentialID,
		"vpc_id":        vpcID,
		"action":        action,
	}

	for k, v := range subnetData {
		data[k] = v
	}

	event := Event{
		Type:      topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Subnet event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.String("vpc_id", vpcID),
			zap.Error(err))
		return nil
	}

	p.logger.Debug("Published Subnet event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("vpc_id", vpcID),
		zap.String("action", action))

	return nil
}

// PublishSecurityGroupEvent publishes a Security Group event
func (p *Publisher) PublishSecurityGroupEvent(ctx context.Context, provider, credentialID, region, action string, securityGroupData map[string]interface{}) error {
	topic := BuildNetworkSecurityGroupTopic(provider, credentialID, region, action)

	data := map[string]interface{}{
		"provider":      provider,
		"credential_id": credentialID,
		"region":        region,
		"action":        action,
	}

	for k, v := range securityGroupData {
		data[k] = v
	}

	event := Event{
		Type:      topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Security Group event, continuing without event",
			zap.String("topic", topic),
			zap.String("provider", provider),
			zap.String("region", region),
			zap.Error(err))
		return nil
	}

	p.logger.Debug("Published Security Group event",
		zap.String("topic", topic),
		zap.String("provider", provider),
		zap.String("region", region),
		zap.String("action", action))

	return nil
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

// PublishWorkspaceEvent publishes a Workspace event
func (p *Publisher) PublishWorkspaceEvent(ctx context.Context, workspaceID, action string, workspaceData map[string]interface{}) error {
	topic := BuildWorkspaceTopic(workspaceID, action)

	data := map[string]interface{}{
		"workspace_id": workspaceID,
		"action":       action,
	}

	for k, v := range workspaceData {
		data[k] = v
	}

	event := Event{
		Type:        topic,
		Data:        data,
		Timestamp:   time.Now().Unix(),
		WorkspaceID: workspaceID,
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Workspace event, continuing without event",
			zap.String("topic", topic),
			zap.String("workspace_id", workspaceID),
			zap.Error(err))
		// Don't fail the operation if event publishing fails
		// This allows graceful degradation when NATS is unavailable
		return nil
	}

	p.logger.Debug("Published Workspace event",
		zap.String("topic", topic),
		zap.String("workspace_id", workspaceID),
		zap.String("action", action))

	return nil
}

// PublishCredentialEvent publishes a Credential event
func (p *Publisher) PublishCredentialEvent(ctx context.Context, workspaceID, provider, action string, credentialData map[string]interface{}) error {
	topic := BuildCredentialTopic(workspaceID, provider, action)

	data := map[string]interface{}{
		"workspace_id": workspaceID,
		"provider":     provider,
		"action":       action,
	}

	for k, v := range credentialData {
		data[k] = v
	}

	event := Event{
		Type:        topic,
		Data:        data,
		Timestamp:   time.Now().Unix(),
		WorkspaceID: workspaceID,
	}

	if err := p.bus.Publish(ctx, event); err != nil {
		p.logger.Warn("Failed to publish Credential event, continuing without event",
			zap.String("topic", topic),
			zap.String("workspace_id", workspaceID),
			zap.String("provider", provider),
			zap.Error(err))
		// Don't fail the operation if event publishing fails
		// This allows graceful degradation when NATS is unavailable
		return nil
	}

	p.logger.Debug("Published Credential event",
		zap.String("topic", topic),
		zap.String("workspace_id", workspaceID),
		zap.String("provider", provider),
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

// PublishToOutbox stores an event in the outbox table for transactional publishing
// 트랜잭션 발행을 위해 이벤트를 outbox 테이블에 저장
func (p *Publisher) PublishToOutbox(ctx context.Context, outboxRepo domain.OutboxRepository, topic, eventType string, data map[string]interface{}, workspaceID *string) error {
	// Convert data to JSONBMap
	// data를 JSONBMap으로 변환
	jsonbData := domain.JSONBMap(data)

	// Create OutboxEvent
	// OutboxEvent 생성
	event := &domain.OutboxEvent{
		Topic:     topic,
		EventType: eventType,
		Data:      jsonbData,
		Status:    domain.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	// Set workspace ID if provided
	// workspace ID가 제공되면 설정
	if workspaceID != nil {
		wsID, err := uuid.Parse(*workspaceID)
		if err == nil {
			event.WorkspaceID = &wsID
		}
	}

	// Store in outbox
	// outbox에 저장
	if err := outboxRepo.Create(ctx, event); err != nil {
		p.logger.Warn("Failed to store event in outbox",
			zap.String("topic", topic),
			zap.String("event_type", eventType),
			zap.Error(err))
		return fmt.Errorf("failed to store event in outbox: %w", err)
	}

	p.logger.Debug("Stored event in outbox",
		zap.String("topic", topic),
		zap.String("event_type", eventType),
		zap.String("id", event.ID.String()))

	return nil
}

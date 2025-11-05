package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
}

// DomainEventHandler handles domain events
type DomainEventHandler interface {
	Handle(event DomainEvent) error
}

// DomainEventPublisher publishes domain events
type DomainEventPublisher interface {
	Publish(event DomainEvent) error
}

// DomainEventSubscriber subscribes to domain events
type DomainEventSubscriber interface {
	Subscribe(eventType string, handler DomainEventHandler) error
}

// Event types
const (
	// User events
	EventUserCreated     = "user.created"
	EventUserUpdated     = "user.updated"
	EventUserDeleted     = "user.deleted"
	EventUserActivated   = "user.activated"
	EventUserDeactivated = "user.deactivated"

	// Workspace events
	EventWorkspaceCreated = "workspace.created"
	EventWorkspaceUpdated = "workspace.updated"
	EventWorkspaceDeleted = "workspace.deleted"

	// VM events
	EventVMCreated       = "vm.created"
	EventVMUpdated       = "vm.updated"
	EventVMDeleted       = "vm.deleted"
	EventVMStarted       = "vm.started"
	EventVMStopped       = "vm.stopped"
	EventVMRestarted     = "vm.restarted"
	EventVMStatusChanged = "vm.status_changed"

	// Credential events
	EventCredentialCreated = "credential.created"
	EventCredentialUpdated = "credential.updated"
	EventCredentialDeleted = "credential.deleted"

	// Audit events
	EventAuditLogCreated = "audit_log.created"

	// Kubernetes events
	EventKubernetesClusterCreated   = "kubernetes.cluster.created"
	EventKubernetesClusterUpdated   = "kubernetes.cluster.updated"
	EventKubernetesClusterDeleted   = "kubernetes.cluster.deleted"
	EventKubernetesClusterStatusChanged = "kubernetes.cluster.status_changed"
	EventKubernetesNodePoolCreated  = "kubernetes.node_pool.created"
	EventKubernetesNodePoolUpdated  = "kubernetes.node_pool.updated"
	EventKubernetesNodePoolDeleted  = "kubernetes.node_pool.deleted"
	EventKubernetesNodeCreated      = "kubernetes.node.created"
	EventKubernetesNodeUpdated      = "kubernetes.node.updated"
	EventKubernetesNodeDeleted      = "kubernetes.node.deleted"

	// Network events
	EventNetworkVPCCreated          = "network.vpc.created"
	EventNetworkVPCUpdated          = "network.vpc.updated"
	EventNetworkVPCDeleted          = "network.vpc.deleted"
	EventNetworkSubnetCreated       = "network.subnet.created"
	EventNetworkSubnetUpdated       = "network.subnet.updated"
	EventNetworkSubnetDeleted       = "network.subnet.deleted"
	EventNetworkSecurityGroupCreated = "network.security_group.created"
	EventNetworkSecurityGroupUpdated = "network.security_group.updated"
	EventNetworkSecurityGroupDeleted = "network.security_group.deleted"
)

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType string, data map[string]interface{}) DomainEvent {
	return DomainEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Version:   1,
	}
}

// UserCreatedEvent creates a user created event
func UserCreatedEvent(user *User) DomainEvent {
	return NewDomainEvent(EventUserCreated, map[string]interface{}{
		"user_id":   user.ID.String(),
		"username":  user.Username,
		"email":     user.Email,
		"is_active": user.IsActive,
	})
}

// UserUpdatedEvent creates a user updated event
func UserUpdatedEvent(user *User, changes map[string]interface{}) DomainEvent {
	data := map[string]interface{}{
		"user_id":   user.ID.String(),
		"username":  user.Username,
		"email":     user.Email,
		"is_active": user.IsActive,
	}

	// Add changes
	for key, value := range changes {
		data["changed_"+key] = value
	}

	return NewDomainEvent(EventUserUpdated, data)
}

// UserDeletedEvent creates a user deleted event
func UserDeletedEvent(userID uuid.UUID, username string) DomainEvent {
	return NewDomainEvent(EventUserDeleted, map[string]interface{}{
		"user_id":  userID.String(),
		"username": username,
	})
}

// WorkspaceCreatedEvent creates a workspace created event
func WorkspaceCreatedEvent(workspace *Workspace) DomainEvent {
	return NewDomainEvent(EventWorkspaceCreated, map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner_id":     workspace.OwnerID,
		"is_active":    workspace.IsActive,
	})
}

// WorkspaceUpdatedEvent creates a workspace updated event
func WorkspaceUpdatedEvent(workspace *Workspace, changes map[string]interface{}) DomainEvent {
	data := map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner_id":     workspace.OwnerID,
		"is_active":    workspace.IsActive,
	}

	// Add changes
	for key, value := range changes {
		data["changed_"+key] = value
	}

	return NewDomainEvent(EventWorkspaceUpdated, data)
}

// WorkspaceDeletedEvent creates a workspace deleted event
func WorkspaceDeletedEvent(workspaceID string, name string, ownerID string) DomainEvent {
	return NewDomainEvent(EventWorkspaceDeleted, map[string]interface{}{
		"workspace_id": workspaceID,
		"name":         name,
		"owner_id":     ownerID,
	})
}

// VMCreatedEvent creates a VM created event
func VMCreatedEvent(vm *VM) DomainEvent {
	return NewDomainEvent(EventVMCreated, map[string]interface{}{
		"vm_id":        vm.ID,
		"name":         vm.Name,
		"workspace_id": vm.WorkspaceID,
		"provider":     vm.Provider,
		"instance_id":  vm.InstanceID,
		"status":       string(vm.Status),
		"type":         vm.Type,
		"region":       vm.Region,
	})
}

// VMUpdatedEvent creates a VM updated event
func VMUpdatedEvent(vm *VM, changes map[string]interface{}) DomainEvent {
	data := map[string]interface{}{
		"vm_id":        vm.ID,
		"name":         vm.Name,
		"workspace_id": vm.WorkspaceID,
		"provider":     vm.Provider,
		"instance_id":  vm.InstanceID,
		"status":       string(vm.Status),
		"type":         vm.Type,
		"region":       vm.Region,
	}

	// Add changes
	for key, value := range changes {
		data["changed_"+key] = value
	}

	return NewDomainEvent(EventVMUpdated, data)
}

// VMStatusChangedEvent creates a VM status changed event
func VMStatusChangedEvent(vmID string, oldStatus VMStatus, newStatus VMStatus, reason string) DomainEvent {
	return NewDomainEvent(EventVMStatusChanged, map[string]interface{}{
		"vm_id":      vmID,
		"old_status": string(oldStatus),
		"new_status": string(newStatus),
		"reason":     reason,
	})
}

// VMDeletedEvent creates a VM deleted event
func VMDeletedEvent(vmID string, name string, workspaceID string, provider string) DomainEvent {
	return NewDomainEvent(EventVMDeleted, map[string]interface{}{
		"vm_id":        vmID,
		"name":         name,
		"workspace_id": workspaceID,
		"provider":     provider,
	})
}

// CredentialCreatedEvent creates a credential created event
func CredentialCreatedEvent(credential *Credential) DomainEvent {
	return NewDomainEvent(EventCredentialCreated, map[string]interface{}{
		"credential_id": credential.ID,
		"workspace_id":  credential.WorkspaceID.String(),
		"created_by":    credential.CreatedBy.String(),
		"provider":      credential.Provider,
	})
}

// CredentialUpdatedEvent creates a credential updated event
func CredentialUpdatedEvent(credential *Credential, changes map[string]interface{}) DomainEvent {
	data := map[string]interface{}{
		"credential_id": credential.ID,
		"workspace_id":  credential.WorkspaceID.String(),
		"created_by":    credential.CreatedBy.String(),
		"provider":      credential.Provider,
	}

	// Add changes
	for key, value := range changes {
		data["changed_"+key] = value
	}

	return NewDomainEvent(EventCredentialUpdated, data)
}

// CredentialDeletedEvent creates a credential deleted event
func CredentialDeletedEvent(credentialID string, userID string, provider string) DomainEvent {
	return NewDomainEvent(EventCredentialDeleted, map[string]interface{}{
		"credential_id": credentialID,
		"user_id":       userID,
		"provider":      provider,
	})
}

// AuditLogCreatedEvent creates an audit log created event
func AuditLogCreatedEvent(auditLog *AuditLog) DomainEvent {
	return NewDomainEvent(EventAuditLogCreated, map[string]interface{}{
		"audit_log_id": auditLog.ID,
		"user_id":      auditLog.UserID.String(),
		"action":       string(auditLog.Action),
		"resource":     auditLog.Resource,
		"ip_address":   auditLog.IPAddress,
		"user_agent":   auditLog.UserAgent,
	})
}

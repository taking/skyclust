package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent: 도메인 이벤트를 나타내는 타입
type DomainEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
}

// DomainEventHandler: 도메인 이벤트를 처리하는 인터페이스
type DomainEventHandler interface {
	Handle(event DomainEvent) error
}

// DomainEventPublisher: 도메인 이벤트를 발행하는 인터페이스
type DomainEventPublisher interface {
	Publish(event DomainEvent) error
}

// DomainEventSubscriber: 도메인 이벤트를 구독하는 인터페이스
type DomainEventSubscriber interface {
	Subscribe(eventType string, handler DomainEventHandler) error
}

// 이벤트 타입 상수
const (
	// 사용자 관련 이벤트
	EventUserCreated     = "user.created"
	EventUserUpdated     = "user.updated"
	EventUserDeleted     = "user.deleted"
	EventUserActivated   = "user.activated"
	EventUserDeactivated = "user.deactivated"

	// 워크스페이스 관련 이벤트
	EventWorkspaceCreated = "workspace.created"
	EventWorkspaceUpdated = "workspace.updated"
	EventWorkspaceDeleted = "workspace.deleted"

	// VM 관련 이벤트
	EventVMCreated       = "vm.created"
	EventVMUpdated       = "vm.updated"
	EventVMDeleted       = "vm.deleted"
	EventVMStarted       = "vm.started"
	EventVMStopped       = "vm.stopped"
	EventVMRestarted     = "vm.restarted"
	EventVMStatusChanged = "vm.status_changed"

	// 자격증명 관련 이벤트
	EventCredentialCreated = "credential.created"
	EventCredentialUpdated = "credential.updated"
	EventCredentialDeleted = "credential.deleted"

	// 감사 로그 관련 이벤트
	EventAuditLogCreated = "audit_log.created"

	// Kubernetes 관련 이벤트
	EventKubernetesClusterCreated       = "kubernetes.cluster.created"
	EventKubernetesClusterUpdated       = "kubernetes.cluster.updated"
	EventKubernetesClusterDeleted       = "kubernetes.cluster.deleted"
	EventKubernetesClusterStatusChanged = "kubernetes.cluster.status_changed"
	EventKubernetesNodePoolCreated      = "kubernetes.node_pool.created"
	EventKubernetesNodePoolUpdated      = "kubernetes.node_pool.updated"
	EventKubernetesNodePoolDeleted      = "kubernetes.node_pool.deleted"
	EventKubernetesNodeCreated          = "kubernetes.node.created"
	EventKubernetesNodeUpdated          = "kubernetes.node.updated"
	EventKubernetesNodeDeleted          = "kubernetes.node.deleted"

	// 네트워크 관련 이벤트
	EventNetworkVPCCreated           = "network.vpc.created"
	EventNetworkVPCUpdated           = "network.vpc.updated"
	EventNetworkVPCDeleted           = "network.vpc.deleted"
	EventNetworkSubnetCreated        = "network.subnet.created"
	EventNetworkSubnetUpdated        = "network.subnet.updated"
	EventNetworkSubnetDeleted        = "network.subnet.deleted"
	EventNetworkSecurityGroupCreated = "network.security_group.created"
	EventNetworkSecurityGroupUpdated = "network.security_group.updated"
	EventNetworkSecurityGroupDeleted = "network.security_group.deleted"
)

// NewDomainEvent: 새로운 도메인 이벤트를 생성합니다
func NewDomainEvent(eventType string, data map[string]interface{}) DomainEvent {
	return DomainEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Version:   1,
	}
}

// UserCreatedEvent: 사용자 생성 이벤트를 생성합니다
func UserCreatedEvent(user *User) DomainEvent {
	return NewDomainEvent(EventUserCreated, map[string]interface{}{
		"user_id":   user.ID.String(),
		"username":  user.Username,
		"email":     user.Email,
		"is_active": user.IsActive,
	})
}

// UserUpdatedEvent: 사용자 업데이트 이벤트를 생성합니다
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

// UserDeletedEvent: 사용자 삭제 이벤트를 생성합니다
func UserDeletedEvent(userID uuid.UUID, username string) DomainEvent {
	return NewDomainEvent(EventUserDeleted, map[string]interface{}{
		"user_id":  userID.String(),
		"username": username,
	})
}

// WorkspaceCreatedEvent: 워크스페이스 생성 이벤트를 생성합니다
func WorkspaceCreatedEvent(workspace *Workspace) DomainEvent {
	return NewDomainEvent(EventWorkspaceCreated, map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner_id":     workspace.OwnerID,
		"is_active":    workspace.IsActive,
	})
}

// WorkspaceUpdatedEvent: 워크스페이스 업데이트 이벤트를 생성합니다
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

// WorkspaceDeletedEvent: 워크스페이스 삭제 이벤트를 생성합니다
func WorkspaceDeletedEvent(workspaceID string, name string, ownerID string) DomainEvent {
	return NewDomainEvent(EventWorkspaceDeleted, map[string]interface{}{
		"workspace_id": workspaceID,
		"name":         name,
		"owner_id":     ownerID,
	})
}

// VMCreatedEvent: VM 생성 이벤트를 생성합니다
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

// VMUpdatedEvent: VM 업데이트 이벤트를 생성합니다
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

// VMStatusChangedEvent: VM 상태 변경 이벤트를 생성합니다
func VMStatusChangedEvent(vmID string, oldStatus VMStatus, newStatus VMStatus, reason string) DomainEvent {
	return NewDomainEvent(EventVMStatusChanged, map[string]interface{}{
		"vm_id":      vmID,
		"old_status": string(oldStatus),
		"new_status": string(newStatus),
		"reason":     reason,
	})
}

// VMDeletedEvent: VM 삭제 이벤트를 생성합니다
func VMDeletedEvent(vmID string, name string, workspaceID string, provider string) DomainEvent {
	return NewDomainEvent(EventVMDeleted, map[string]interface{}{
		"vm_id":        vmID,
		"name":         name,
		"workspace_id": workspaceID,
		"provider":     provider,
	})
}

// CredentialCreatedEvent: 자격증명 생성 이벤트를 생성합니다
func CredentialCreatedEvent(credential *Credential) DomainEvent {
	return NewDomainEvent(EventCredentialCreated, map[string]interface{}{
		"credential_id": credential.ID,
		"workspace_id":  credential.WorkspaceID.String(),
		"created_by":    credential.CreatedBy.String(),
		"provider":      credential.Provider,
	})
}

// CredentialUpdatedEvent: 자격증명 업데이트 이벤트를 생성합니다
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

// CredentialDeletedEvent: 자격증명 삭제 이벤트를 생성합니다
func CredentialDeletedEvent(credentialID string, userID string, provider string) DomainEvent {
	return NewDomainEvent(EventCredentialDeleted, map[string]interface{}{
		"credential_id": credentialID,
		"user_id":       userID,
		"provider":      provider,
	})
}

// AuditLogCreatedEvent: 감사 로그 생성 이벤트를 생성합니다
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

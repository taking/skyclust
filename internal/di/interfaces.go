package di

import (
	"skyclust/internal/domain"
)

// ContainerInterface defines the interface for dependency injection container
type ContainerInterface interface {
	// Repository interfaces
	GetUserRepository() domain.UserRepository
	GetWorkspaceRepository() domain.WorkspaceRepository
	GetVMRepository() domain.VMRepository
	GetCredentialRepository() domain.CredentialRepository
	GetAuditLogRepository() domain.AuditLogRepository
	GetOIDCProviderRepository() domain.OIDCProviderRepository
	GetOutboxRepository() domain.OutboxRepository

	// Service interfaces
	GetUserService() domain.UserService
	GetWorkspaceService() domain.WorkspaceService
	GetVMService() domain.VMService
	GetAuthService() domain.AuthService
	GetCredentialService() domain.CredentialService
	GetRBACService() domain.RBACService
	GetAuditLogService() domain.AuditLogService
	GetOIDCService() domain.OIDCService
	GetLogoutService() domain.LogoutService
	GetNotificationService() domain.NotificationService
	GetSystemMonitoringService() interface{}
	GetKubernetesService() interface{}
	GetNetworkService() interface{}
	GetExportService() interface{}
	GetCostAnalysisService() interface{}
	GetComputeService() interface{}
	GetDashboardService() interface{}
	GetBusinessRuleService() interface{}

	// Domain services
	GetDomainService() *domain.DomainService
	GetUserDomainService() *domain.UserDomainService
	GetWorkspaceDomainService() *domain.WorkspaceDomainService
	GetVMDomainService() *domain.VMDomainService

	// Infrastructure
	GetDatabase() interface{}
	GetCache() interface{}
	GetMessaging() interface{}
	GetLogger() interface{}

	// Cleanup
	Close() error
}

// RepositoryContainer holds repository dependencies
type RepositoryContainer struct {
	UserRepository                    domain.UserRepository
	WorkspaceRepository               domain.WorkspaceRepository
	VMRepository                      domain.VMRepository
	CredentialRepository              domain.CredentialRepository
	AuditLogRepository                domain.AuditLogRepository
	NotificationRepository            domain.NotificationRepository
	NotificationPreferencesRepository domain.NotificationPreferencesRepository
	OIDCProviderRepository            domain.OIDCProviderRepository
	RBACRepository                    domain.RBACRepository
	OutboxRepository                  domain.OutboxRepository
}

// ServiceContainer holds service dependencies
type ServiceContainer struct {
	UserService             domain.UserService
	WorkspaceService        domain.WorkspaceService
	VMService               domain.VMService
	AuthService             domain.AuthService
	CredentialService       domain.CredentialService
	RBACService             domain.RBACService
	AuditLogService         domain.AuditLogService
	OIDCService             domain.OIDCService
	LogoutService           domain.LogoutService
	NotificationService     domain.NotificationService
	SystemMonitoringService interface{} // SystemMonitoringService for system health and metrics
	KubernetesService       interface{} // KubernetesService for K8s cluster management
	NetworkService          interface{} // NetworkService for VPC, Subnet, Security Group management
	ExportService           interface{} // TODO: Define ExportService interface in domain
	CostAnalysisService     interface{} // TODO: Define CostAnalysisService interface in domain
	ComputeService          interface{} // TODO: Define ComputeService interface in domain
	DashboardService        interface{} // DashboardService for dashboard summary data
	BusinessRuleService     interface{} // TODO: Define BusinessRuleService interface in domain
}

// DomainContainer holds domain service dependencies
type DomainContainer struct {
	DomainService          *domain.DomainService
	UserDomainService      *domain.UserDomainService
	WorkspaceDomainService *domain.WorkspaceDomainService
	VMDomainService        *domain.VMDomainService
	BusinessRuleService    *domain.BusinessRuleService
}

// InfrastructureContainer holds infrastructure dependencies
type InfrastructureContainer struct {
	Database           interface{}
	Cache              interface{}
	Messaging          interface{}
	Logger             interface{}
	TransactionManager domain.TransactionManager
}

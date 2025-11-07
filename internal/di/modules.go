package di

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	auditlogservice "skyclust/internal/application/services/audit_log"
	authservice "skyclust/internal/application/services/auth"
	cacheservice "skyclust/internal/application/services/cache"
	computeservice "skyclust/internal/application/services/compute"
	costanalysisservice "skyclust/internal/application/services/cost_analysis"
	credentialservice "skyclust/internal/application/services/credential"
	dashboardservice "skyclust/internal/application/services/dashboard"
	eventservice "skyclust/internal/application/services/event"
	exportservice "skyclust/internal/application/services/export"
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	logoutservice "skyclust/internal/application/services/logout"
	networkservice "skyclust/internal/application/services/network"
	notificationservice "skyclust/internal/application/services/notification"
	oidcservice "skyclust/internal/application/services/oidc"
	rbacservice "skyclust/internal/application/services/rbac"
	systemmonitoringservice "skyclust/internal/application/services/system_monitoring"
	userservice "skyclust/internal/application/services/user"
	vmservice "skyclust/internal/application/services/vm"
	workspaceservice "skyclust/internal/application/services/workspace"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/database/postgres"
	"skyclust/internal/infrastructure/messaging"
	k8sworker "skyclust/internal/workers/kubernetes"
	networkworker "skyclust/internal/workers/network"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"
)

// RepositoryModule initializes repository dependencies
type RepositoryModule struct {
	repositories *RepositoryContainer
}

// NewRepositoryModule creates a new repository module
func NewRepositoryModule(db *gorm.DB, redisClient *redis.Client) *RepositoryModule {
	logger.Info("Initializing repository module...")

	// Create repositories
	// Note: OIDCProviderRepository needs encryptor, which will be created later in ServiceModule
	// We'll create it separately after encryptor is available
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	vmRepo := postgres.NewVMRepository(db)
	credentialRepo := postgres.NewCredentialRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	notificationPreferencesRepo := postgres.NewNotificationPreferencesRepository(db)
	rbacRepo := postgres.NewRBACRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)

	logger.Info("Repository module initialized")

	return &RepositoryModule{
		repositories: &RepositoryContainer{
			UserRepository:                    userRepo,
			WorkspaceRepository:               workspaceRepo,
			VMRepository:                      vmRepo,
			CredentialRepository:              credentialRepo,
			AuditLogRepository:                auditLogRepo,
			NotificationRepository:            notificationRepo,
			NotificationPreferencesRepository: notificationPreferencesRepo,
			OIDCProviderRepository:            nil, // Will be set later after encryptor is available
			RBACRepository:                    rbacRepo,
			OutboxRepository:                  outboxRepo,
		},
	}
}

// GetContainer returns the repository container
func (m *RepositoryModule) GetContainer() *RepositoryContainer {
	return m.repositories
}

// ServiceModule initializes service dependencies
type ServiceModule struct {
	services     *ServiceContainer
	messagingBus messaging.Bus
}

// GetMessagingBus returns the messaging bus used by services
func (m *ServiceModule) GetMessagingBus() messaging.Bus {
	return m.messagingBus
}

// NewServiceModule creates a new service module with actual service implementations
func NewServiceModule(repos *RepositoryContainer, db *gorm.DB, config ServiceConfig) *ServiceModule {
	logger.Info("Starting service module initialization...")

	// Get Redis client from config
	var redisClient *redis.Client
	if config.RedisClient != nil {
		if rc, ok := config.RedisClient.(*redis.Client); ok {
			redisClient = rc
		}
	}

	// Note: Security components (hasher, blacklist) are created later after config is available

	// Create System Monitoring service
	systemMonitoringService := systemmonitoringservice.NewService(
		logger.DefaultLogger.GetLogger(),
		nil,          // Config will be set later or use defaults
		config.Cache, // Cache interface for health checks
	)

	// Create security components
	hasher := security.NewBcryptHasher(12) // Use bcrypt with cost 12
	encryptor := security.NewAESEncryptor([]byte(config.EncryptionKey))
	blacklist := cache.NewTokenBlacklist(redisClient)

	// Create messaging bus (shared across services)
	messagingBus := messaging.NewLocalBus()

	// Create EventService (requires messaging bus)
	eventService := eventservice.NewService(messagingBus)

	// Create RBACService first (needed by AuthService)
	rbacService := rbacservice.NewService(repos.RBACRepository)

	// Create AuthService
	authService := authservice.NewService(
		repos.UserRepository,
		repos.AuditLogRepository,
		rbacService,
		hasher,
		blacklist,
		config.JWTSecret,
		config.JWTExpiry,
	)

	// Create UserService
	userService := userservice.NewService(repos.UserRepository, hasher, repos.AuditLogRepository)

	// Create CredentialService with event publisher
	credentialEventPublisher := messaging.NewPublisher(messagingBus, logger.DefaultLogger.GetLogger())
	credentialService := credentialservice.NewService(repos.CredentialRepository, repos.AuditLogRepository, encryptor, credentialEventPublisher)

	// Create Kubernetes service
	k8sService := kubernetesservice.NewService(credentialService, config.Cache, messagingBus, repos.AuditLogRepository, logger.DefaultLogger.GetLogger())

	// Create Network service (after credentialService is created)
	networkService := networkservice.NewService(credentialService, config.Cache, messagingBus, repos.AuditLogRepository, logger.DefaultLogger.GetLogger())

	// Create AuditLogService
	auditLogService := auditlogservice.NewService(repos.AuditLogRepository)

	// Create CacheService for OIDC state storage
	cacheService := cacheservice.NewService(config.Cache)

	// Create OIDCProviderRepository (needs encryptor)
	oidcProviderRepo := postgres.NewOIDCProviderRepository(db, encryptor)
	repos.OIDCProviderRepository = oidcProviderRepo

	// Create OIDCService
	oidcService := oidcservice.NewService(repos.UserRepository, repos.AuditLogRepository, authService, cacheService, oidcProviderRepo)

	// Create WorkspaceService with event publisher
	workspaceEventPublisher := messaging.NewPublisher(messagingBus, logger.DefaultLogger.GetLogger())
	workspaceService := workspaceservice.NewService(repos.WorkspaceRepository, repos.UserRepository, eventService, repos.AuditLogRepository, workspaceEventPublisher)

	// Create NotificationService
	notificationService := notificationservice.NewService(
		logger.DefaultLogger.GetLogger(),
		repos.NotificationRepository,
		repos.NotificationPreferencesRepository,
		repos.AuditLogRepository,
		repos.UserRepository,
		repos.WorkspaceRepository,
		eventService,
	)

	// Create ComputeService first (needed by VMService)
	computeService := computeservice.NewService()

	// Create VMService
	vmEventPublisher := messaging.NewPublisher(messagingBus, logger.DefaultLogger.GetLogger())
	vmService := vmservice.NewService(
		repos.VMRepository,
		repos.WorkspaceRepository,
		computeService,
		eventService,
		repos.AuditLogRepository,
		config.Cache,
		cache.NewCacheKeyBuilder(),
		cache.NewInvalidatorWithEvents(config.Cache, vmEventPublisher),
		logger.DefaultLogger.GetLogger(),
	)

	// Create ExportService
	exportService := exportservice.NewService(
		logger.DefaultLogger.GetLogger(),
		repos.VMRepository,
		repos.WorkspaceRepository,
		repos.CredentialRepository,
		repos.AuditLogRepository,
	)

	// Create CostAnalysisService
	costAnalysisService := costanalysisservice.NewService(
		repos.VMRepository,
		repos.CredentialRepository,
		repos.WorkspaceRepository,
		repos.AuditLogRepository,
		credentialService,
		k8sService,   // Inject KubernetesService for cluster cost calculation
		config.Cache, // Inject cache for cost analysis result caching
	)

	// Create LogoutService
	logoutService := logoutservice.NewService(
		blacklist,
		oidcService,
		repos.AuditLogRepository,
	)

	// Create DashboardService
	dashboardService := dashboardservice.NewService(
		repos.WorkspaceRepository,
		repos.VMRepository,
		repos.CredentialRepository,
		k8sService,
		networkService,
		logger.DefaultLogger.GetLogger(),
	)

	return &ServiceModule{
		services: &ServiceContainer{
			AuthService:             authService,
			UserService:             userService,
			CredentialService:       credentialService,
			RBACService:             rbacService,
			AuditLogService:         auditLogService,
			KubernetesService:       k8sService,
			NetworkService:          networkService,
			SystemMonitoringService: systemMonitoringService,
			OIDCService:             oidcService,
			LogoutService:           logoutService,
			WorkspaceService:        workspaceService,
			VMService:               vmService,
			NotificationService:     notificationService,
			ExportService:           exportService,
			CostAnalysisService:     costAnalysisService,
			ComputeService:          computeService,
			DashboardService:        dashboardService,
			BusinessRuleService:     nil, // BusinessRuleService is in DomainContainer, not ServiceContainer
		},
		messagingBus: messagingBus,
	}
}

// GetContainer returns the service container
func (m *ServiceModule) GetContainer() *ServiceContainer {
	return m.services
}

// ServiceConfig holds service configuration
type ServiceConfig struct {
	JWTSecret     string
	JWTExpiry     time.Duration
	EncryptionKey string
	RedisClient   interface{} // Redis client for TokenBlacklist
	Cache         cache.Cache // Cache for OIDC state storage
}

// DomainModule initializes domain service dependencies
type DomainModule struct {
	domain *DomainContainer
}

// NewDomainModule creates a new domain module
func NewDomainModule(repos *RepositoryContainer) *DomainModule {
	// Create domain service with repository dependencies
	domainService := domain.NewDomainService(
		repos.UserRepository,
		repos.WorkspaceRepository,
		repos.VMRepository,
		repos.AuditLogRepository,
	)

	// Create specialized domain services
	userDomainService := domain.NewUserDomainService(domainService)
	workspaceDomainService := domain.NewWorkspaceDomainService(domainService)
	vmDomainService := domain.NewVMDomainService(domainService)
	businessRuleService := domain.NewBusinessRuleService(domainService)

	return &DomainModule{
		domain: &DomainContainer{
			DomainService:          domainService,
			UserDomainService:      userDomainService,
			WorkspaceDomainService: workspaceDomainService,
			VMDomainService:        vmDomainService,
			BusinessRuleService:    businessRuleService,
		},
	}
}

// GetContainer returns the domain container
func (m *DomainModule) GetContainer() *DomainContainer {
	return m.domain
}

// InfrastructureModule initializes infrastructure dependencies
type InfrastructureModule struct {
	infrastructure *InfrastructureContainer
}

// NewInfrastructureModule creates a new infrastructure module
func NewInfrastructureModule(db *gorm.DB, redisClient *redis.Client) *InfrastructureModule {
	logger.Info("Initializing infrastructure module...")

	// Create transaction manager
	transactionManager := postgres.NewTransactionManager(db)

	// Note: Messaging bus and notification service are initialized in ServiceModule
	// as they require application-level dependencies

	logger.Info("Infrastructure module initialized")

	return &InfrastructureModule{
		infrastructure: &InfrastructureContainer{
			TransactionManager: transactionManager,
			// Infrastructure-level components can be added here in the future
		},
	}
}

// GetContainer returns the infrastructure container
func (m *InfrastructureModule) GetContainer() *InfrastructureContainer {
	return m.infrastructure
}

// WorkerModule initializes and manages background workers
type WorkerModule struct {
	workers *WorkerContainer
}

// WorkerContainer holds worker dependencies
type WorkerContainer struct {
	KubernetesSyncWorker *k8sworker.SyncWorker
	NetworkSyncWorker    *networkworker.SyncWorker
}

// NewWorkerModule creates a new worker module
func NewWorkerModule(
	serviceModule *ServiceModule,
	repositoryModule *RepositoryModule,
	cacheService cache.Cache,
	messagingBus interface{},
	logger *zap.Logger,
) *WorkerModule {
	// Get required services
	services := serviceModule.GetContainer()
	repos := repositoryModule.GetContainer()

	// Convert messaging bus
	var eventBus messaging.Bus
	if bus, ok := messagingBus.(messaging.Bus); ok {
		eventBus = bus
	} else {
		// Fallback to LocalBus if messaging is not available
		eventBus = messaging.NewLocalBus()
		logger.Warn("Messaging bus not available, using LocalBus for workers")
	}

	// Get Kubernetes and Network services
	var k8sService *kubernetesservice.Service
	if k8s, ok := services.KubernetesService.(*kubernetesservice.Service); ok {
		k8sService = k8s
	}

	var networkService *networkservice.Service
	if net, ok := services.NetworkService.(*networkservice.Service); ok {
		networkService = net
	}

	// Create Kubernetes sync worker
	var k8sWorker *k8sworker.SyncWorker
	if k8sService != nil {
		k8sWorker = k8sworker.NewSyncWorker(
			k8sService,
			services.CredentialService,
			repos.CredentialRepository,
			repos.WorkspaceRepository,
			cacheService,
			eventBus,
			logger,
			k8sworker.SyncWorkerConfig{
				SyncInterval:   5 * time.Minute,
				MaxConcurrency: 5,
			},
		)
		logger.Info("Kubernetes sync worker created")
	}

	// Create Network sync worker
	var networkWorker *networkworker.SyncWorker
	if networkService != nil {
		networkWorker = networkworker.NewSyncWorker(
			networkService,
			services.CredentialService,
			repos.CredentialRepository,
			repos.WorkspaceRepository,
			cacheService,
			eventBus,
			logger,
			networkworker.SyncWorkerConfig{
				SyncInterval:   5 * time.Minute,
				MaxConcurrency: 5,
			},
		)
		logger.Info("Network sync worker created")
	}

	return &WorkerModule{
		workers: &WorkerContainer{
			KubernetesSyncWorker: k8sWorker,
			NetworkSyncWorker:    networkWorker,
		},
	}
}

// GetContainer returns the worker container
func (m *WorkerModule) GetContainer() *WorkerContainer {
	return m.workers
}

// StartAll starts all workers
func (m *WorkerModule) StartAll(ctx context.Context) error {
	if m.workers.KubernetesSyncWorker != nil {
		if err := m.workers.KubernetesSyncWorker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Kubernetes sync worker: %w", err)
		}
	}

	if m.workers.NetworkSyncWorker != nil {
		if err := m.workers.NetworkSyncWorker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Network sync worker: %w", err)
		}
	}

	return nil
}

// StopAll stops all workers
func (m *WorkerModule) StopAll() {
	if m.workers.KubernetesSyncWorker != nil {
		m.workers.KubernetesSyncWorker.Stop()
	}

	if m.workers.NetworkSyncWorker != nil {
		m.workers.NetworkSyncWorker.Stop()
	}
}

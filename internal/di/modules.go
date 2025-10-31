package di

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/database/postgres"
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
			OIDCProviderRepository:           nil, // Will be set later after encryptor is available
		},
	}
}

// GetContainer returns the repository container
func (m *RepositoryModule) GetContainer() *RepositoryContainer {
	return m.repositories
}

// ServiceModule initializes service dependencies
type ServiceModule struct {
	services *ServiceContainer
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

	// Create security hasher
	logger.Info("Creating security hasher...")
	// hasher := security.NewHasher() - TODO: Use when implementing services
	logger.Info("Security hasher created")

	// Initialize TokenBlacklist with Redis client
	logger.Info("Initializing TokenBlacklist...")
	if redisClient == nil {
		logger.Warn("Redis client not available, token blacklisting disabled")
	}

	// blacklist := cache.NewTokenBlacklist(redisClient) - TODO: Use when implementing services
	logger.Info("TokenBlacklist created")

	// Create RBAC service (requires database for role/permission management)
	logger.Info("Creating RBAC service...")
	logger.Info("RBAC service created")

	// Create Auth service with all dependencies
	logger.Info("Creating Auth service...")
	logger.Info("Auth service created")

	// Create Audit Log service first (needed by other services)
	logger.Info("Creating Audit Log service...")
	logger.Info("Audit Log service created")

	// Create User service
	logger.Info("Creating User service...")
	logger.Info("User service created")

	// Create Credential service
	// Note: Requires encryption key from config
	logger.Info("Credential service created")

	// Create Workspace service
	// Note: Requires messaging bus
	logger.Info("Workspace service created")

	// Create VM service
	// Note: Requires cloud provider service and messaging bus
	logger.Info("VM service created")

	// Create OIDC service
	logger.Info("OIDC service created")

	// Create Logout service
	logger.Info("Logout service created")

	// Create System Monitoring service
	logger.Info("Creating System Monitoring service...")
	systemMonitoringService := service.NewSystemMonitoringService(
		logger.DefaultLogger.GetLogger(),
		nil,         // Config will be set later or use defaults
		redisClient, // Redis client for health checks
	)
	logger.Info("System Monitoring service created")

	// Create actual services
	logger.Info("Creating actual services...")

	// Create security components
	logger.Info("Creating security components...")
	hasher := security.NewBcryptHasher(12) // Use bcrypt with cost 12
	encryptor := security.NewAESEncryptor([]byte(config.EncryptionKey))
	blacklist := cache.NewTokenBlacklist(redisClient)
	logger.Info("Security components created")

	// Create RBACService first (needed by AuthService)
	logger.Info("Creating RBAC service...")
	rbacService := service.NewRBACService(db)
	logger.Info("RBAC service created")

	// Create AuthService
	logger.Info("Creating Auth service...")
	authService := service.NewAuthService(
		repos.UserRepository,
		repos.AuditLogRepository,
		rbacService,
		hasher,
		blacklist,
		config.JWTSecret,
		config.JWTExpiry,
	)
	logger.Info("Auth service created")

	// Create UserService
	logger.Info("Creating User service...")
	userService := service.NewUserService(repos.UserRepository, hasher, repos.AuditLogRepository)
	logger.Info("User service created")

	// Create CredentialService
	logger.Info("Creating Credential service...")
	credentialService := service.NewCredentialService(repos.CredentialRepository, repos.AuditLogRepository, encryptor)
	logger.Info("Credential service created")

	// Create Kubernetes service
	logger.Info("Creating Kubernetes service...")

	k8sService := service.NewKubernetesService(credentialService, logger.DefaultLogger.GetLogger())

	logger.Info("Kubernetes service created")

	// Create Network service (after credentialService is created)
	logger.Info("Creating Network service...")
	networkService := service.NewNetworkService(credentialService, logger.DefaultLogger.GetLogger())
	logger.Info("Network service created")

	// Create AuditLogService
	logger.Info("Creating Audit Log service...")
	auditLogService := service.NewAuditLogService(repos.AuditLogRepository)
	logger.Info("Audit Log service created")

	// Create CacheService for OIDC state storage
	logger.Info("Creating Cache service...")
	cacheService := service.NewCacheService(config.Cache)
	logger.Info("Cache service created")

	// Create OIDCProviderRepository (needs encryptor)
	logger.Info("Creating OIDC Provider repository...")
	oidcProviderRepo := postgres.NewOIDCProviderRepository(db, encryptor)
	repos.OIDCProviderRepository = oidcProviderRepo
	logger.Info("OIDC Provider repository created")

	// Create OIDCService
	logger.Info("Creating OIDC service...")
	oidcService := service.NewOIDCService(repos.UserRepository, repos.AuditLogRepository, authService, cacheService, oidcProviderRepo)
	logger.Info("OIDC service created")

	// Create WorkspaceService
	logger.Info("Creating Workspace service...")
	workspaceService := service.NewWorkspaceService(repos.WorkspaceRepository, repos.UserRepository, nil, repos.AuditLogRepository) // TODO: Add messaging bus
	logger.Info("Workspace service created")

	// Create NotificationService
	logger.Info("Creating Notification service...")
	notificationService := service.NewNotificationService(
		logger.DefaultLogger.GetLogger(),
		repos.NotificationRepository,
		repos.NotificationPreferencesRepository,
		repos.AuditLogRepository,
		repos.UserRepository,
		repos.WorkspaceRepository,
		nil, // TODO: Add EventService
	)
	logger.Info("Notification service created")

	// Create CloudProviderService first (needed by VMService)
	logger.Info("Creating Cloud Provider service...")
	cloudProviderService := service.NewCloudProviderService()
	logger.Info("Cloud Provider service created")

	// Create VMService
	logger.Info("Creating VM service...")
	vmService := service.NewVMService(
		repos.VMRepository,
		repos.WorkspaceRepository,
		cloudProviderService,
		nil, // TODO: Add messaging.Bus
		repos.AuditLogRepository,
	)
	logger.Info("VM service created")

	// Create ExportService
	logger.Info("Creating Export service...")
	exportService := service.NewExportService(
		logger.DefaultLogger.GetLogger(),
		repos.VMRepository,
		repos.WorkspaceRepository,
		repos.CredentialRepository,
		repos.AuditLogRepository,
	)
	logger.Info("Export service created")

	// Create CostAnalysisService
	logger.Info("Creating Cost Analysis service...")
	costAnalysisService := service.NewCostAnalysisService(
		repos.VMRepository,
		repos.CredentialRepository,
		repos.WorkspaceRepository,
		repos.AuditLogRepository,
	)
	logger.Info("Cost Analysis service created")

	// Create LogoutService
	logger.Info("Creating Logout service...")
	logoutService := service.NewLogoutService(
		blacklist,
		oidcService,
		repos.AuditLogRepository,
	)
	logger.Info("Logout service created")

	logger.Info("All services created successfully")

	// Debug: Check if KubernetesService is properly created
	if k8sService != nil {
		logger.Info("KubernetesService is not nil, assigning to container")
	} else {
		logger.Warn("KubernetesService is nil!")
	}

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
			CloudProviderService:    cloudProviderService,
			BusinessRuleService:     nil, // TODO: Implement BusinessRuleService
		},
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

	return &DomainModule{
		domain: &DomainContainer{
			DomainService:          domainService,
			UserDomainService:      userDomainService,
			WorkspaceDomainService: workspaceDomainService,
			VMDomainService:        vmDomainService,
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

	// Initialize messaging bus
	// messagingBus := messaging.NewBus() - TODO: Implement messaging bus

	// Initialize notification service
	// notificationService := notification.NewService() - TODO: Implement notification service

	logger.Info("Infrastructure module initialized")

	return &InfrastructureModule{
		infrastructure: &InfrastructureContainer{
			// MessagingBus:      messagingBus, - TODO: Implement messaging bus
			// NotificationService: notificationService, - TODO: Implement notification service
		},
	}
}

// GetContainer returns the infrastructure container
func (m *InfrastructureModule) GetContainer() *InfrastructureContainer {
	return m.infrastructure
}

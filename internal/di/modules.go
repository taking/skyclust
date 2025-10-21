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
func NewRepositoryModule(db interface{}) *RepositoryModule {
	// Cast db to *gorm.DB
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		logger.Errorf("Invalid database type, expected *gorm.DB")
		return &RepositoryModule{
			repositories: &RepositoryContainer{},
		}
	}

	// Create actual repositories using postgres implementations
	return &RepositoryModule{
		repositories: &RepositoryContainer{
			UserRepository:       postgres.NewUserRepository(gormDB),
			CredentialRepository: postgres.NewCredentialRepository(gormDB),
			AuditLogRepository:   postgres.NewAuditLogRepository(gormDB),
			WorkspaceRepository:  postgres.NewWorkspaceRepository(gormDB),
			VMRepository:         postgres.NewVMRepository(gormDB),
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

	// Create security dependencies
	logger.Info("Creating security hasher...")
	hasher := security.NewBcryptHasher(12) // bcrypt cost 12

	// Create cache dependencies with Redis client
	logger.Info("Initializing TokenBlacklist...")
	var redisClient *redis.Client
	if config.RedisClient != nil {
		if client, ok := config.RedisClient.(*redis.Client); ok {
			redisClient = client
			logger.Info("TokenBlacklist initialized with Redis client")
		}
	}

	if redisClient == nil {
		logger.Warn("Redis client not available, token blacklisting disabled")
	}

	blacklist := cache.NewTokenBlacklist(redisClient)
	logger.Info("TokenBlacklist created")

	// Create RBAC service (requires database for role/permission management)
	logger.Info("Creating RBAC service...")
	rbacService := service.NewRBACService(db)
	logger.Info("RBAC service created")

	// Create Auth service with all dependencies
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

	// Create Audit Log service first (needed by other services)
	logger.Info("Creating Audit Log service...")
	auditLogService := service.NewAuditLogService(repos.AuditLogRepository)
	logger.Info("Audit Log service created")

	// Create User service
	logger.Info("Creating User service...")
	userService := service.NewUserService(
		repos.UserRepository,
		hasher,
		repos.AuditLogRepository,
	)
	logger.Info("User service created")

	// Create Credential service
	// Note: Requires encryption key from config
	encryptor := security.NewAESEncryptor([]byte(config.EncryptionKey))
	credentialService := service.NewCredentialService(
		repos.CredentialRepository,
		repos.AuditLogRepository,
		encryptor,
	)

	// Create Workspace service
	// Note: Requires messaging bus
	workspaceService := service.NewWorkspaceService(
		repos.WorkspaceRepository,
		repos.UserRepository,
		nil, // TODO: Add messaging bus
		repos.AuditLogRepository,
	)

	// Create VM service
	// Note: Requires cloud provider service and messaging bus
	vmService := service.NewVMService(
		repos.VMRepository,
		repos.WorkspaceRepository,
		nil, // TODO: Add cloud provider service
		nil, // TODO: Add messaging bus
		repos.AuditLogRepository,
	)

	// Create OIDC service
	oidcService := service.NewOIDCService(
		repos.UserRepository,
		repos.AuditLogRepository,
		authService, // Use authService instead of rbacService
	)

	// Create Logout service
	logoutService := service.NewLogoutService(
		blacklist,
		oidcService, // Add oidcService
		repos.AuditLogRepository,
	)

	// Create Kubernetes service
	logger.Info("Creating Kubernetes service...")
	k8sService := service.NewKubernetesService(
		credentialService,
		logger.DefaultLogger.GetLogger(),
	)
	logger.Info("Kubernetes service created")

	return &ServiceModule{
		services: &ServiceContainer{
			AuthService:          authService,
			UserService:          userService,
			CredentialService:    credentialService,
			RBACService:          rbacService,
			AuditLogService:      auditLogService,
			KubernetesService:    k8sService,
			OIDCService:          oidcService,
			LogoutService:        logoutService,
			WorkspaceService:     workspaceService,
			VMService:            vmService,
			NotificationService:  nil, // TODO: Implement NotificationService
			ExportService:        nil, // TODO: Implement ExportService
			CostAnalysisService:  nil, // TODO: Implement CostAnalysisService
			CloudProviderService: nil, // TODO: Implement CloudProviderService
			BusinessRuleService:  nil, // TODO: Implement BusinessRuleService
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
func NewInfrastructureModule() *InfrastructureModule {
	return &InfrastructureModule{
		infrastructure: &InfrastructureContainer{
			Database:  nil, // Will be injected
			Cache:     nil, // Will be injected
			Messaging: nil, // Will be injected
			Logger:    logger.DefaultLogger,
		},
	}
}

// GetContainer returns the infrastructure container
func (m *InfrastructureModule) GetContainer() *InfrastructureContainer {
	return m.infrastructure
}

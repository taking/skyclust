package di

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/database"
	"skyclust/pkg/cache"
	"skyclust/pkg/config"
	"skyclust/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Container is an improved dependency injection container
type Container struct {
	mu sync.RWMutex

	// Modules
	repositoryModule     *RepositoryModule
	serviceModule        *ServiceModule
	domainModule         *DomainModule
	infrastructureModule *InfrastructureModule

	// Core dependencies
	db     *gorm.DB
	cache  cache.Cache
	logger *zap.Logger

	// Initialization state
	initialized bool
}

// NewContainerV2 creates a new improved DI container
func NewContainer() *Container {
	return &Container{
		initialized: false,
	}
}

// Initialize initializes the container with dependencies
func (c *Container) Initialize(ctx context.Context, cfg *config.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("container already initialized")
	}

	// Initialize infrastructure
	if err := c.initializeInfrastructure(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize modules
	logger.Info("Initializing repository module...")
	var redisClient *redis.Client
	if redisService, ok := c.cache.(*cache.RedisService); ok {
		redisClient = redisService.GetClient()
		logger.Info("Redis client available for TokenBlacklist")
	} else {
		logger.Warn("Redis client not available, TokenBlacklist will be disabled")
	}
	c.repositoryModule = NewRepositoryModule(c.db, redisClient)
	logger.Info("Repository module initialized")

	// Create service configuration with Redis client
	logger.Info("Creating service configuration...")

	serviceConfig := ServiceConfig{
		JWTSecret:     cfg.Security.JWTSecret,
		JWTExpiry:     cfg.Security.JWTExpiration,
		EncryptionKey: cfg.Security.EncryptionKey,
		RedisClient:   redisClient, // Pass Redis client for TokenBlacklist
		Cache:         c.cache,     // Pass cache for OIDC state storage
	}

	logger.Info("Initializing service module...")
	c.serviceModule = NewServiceModule(c.repositoryModule.GetContainer(), c.db, serviceConfig)
	logger.Info("Service module initialized")

	logger.Info("Initializing domain module...")
	c.domainModule = NewDomainModule(c.repositoryModule.GetContainer())
	logger.Info("Domain module initialized")

	logger.Info("Initializing infrastructure module...")
	c.infrastructureModule = NewInfrastructureModule(c.db, redisClient)
	logger.Info("Infrastructure module initialized")

	// Set infrastructure dependencies
	c.infrastructureModule.infrastructure.Database = c.db
	c.infrastructureModule.infrastructure.Cache = c.cache
	c.infrastructureModule.infrastructure.Logger = c.logger

	c.initialized = true
	return nil
}

// initializeInfrastructure initializes core infrastructure components
func (c *Container) initializeInfrastructure(ctx context.Context, cfg *config.Config) error {
	// Initialize database
	db, err := database.NewPostgresService(database.PostgresConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name, // Name -> Database 매핑
		SSLMode:         cfg.Database.SSLMode,
		MaxConns:        cfg.Database.MaxConns,
		MinConns:        cfg.Database.MinConns,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		SlowQueryLog:    true,
		SlowQueryTime:   1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	c.db = db.GetDB()

	// Initialize cache (Redis with fallback to Memory)
	redisService, err := cache.NewRedisService(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 10, // Default pool size
	})
	if err != nil {
		logger.Warnf("Failed to connect to Redis, falling back to memory cache: %v", err)
		c.cache = cache.NewMemoryCache()
	} else {
		c.cache = redisService
		logger.Info("Successfully initialized Redis cache")
	}

	// Initialize logger
	c.logger = logger.DefaultLogger.GetLogger()

	return nil
}

// GetUserRepository returns the user repository
func (c *Container) GetUserRepository() domain.UserRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().UserRepository
}

// GetWorkspaceRepository returns the workspace repository
func (c *Container) GetWorkspaceRepository() domain.WorkspaceRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().WorkspaceRepository
}

// GetVMRepository returns the VM repository
func (c *Container) GetVMRepository() domain.VMRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().VMRepository
}

// GetCredentialRepository returns the credential repository
func (c *Container) GetCredentialRepository() domain.CredentialRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().CredentialRepository
}

// GetAuditLogRepository returns the audit log repository
func (c *Container) GetAuditLogRepository() domain.AuditLogRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().AuditLogRepository
}

// GetOIDCProviderRepository returns the OIDC provider repository
func (c *Container) GetOIDCProviderRepository() domain.OIDCProviderRepository {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repositoryModule.GetContainer().OIDCProviderRepository
}

// GetUserService returns the user service
func (c *Container) GetUserService() domain.UserService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().UserService
}

// GetWorkspaceService returns the workspace service
func (c *Container) GetWorkspaceService() domain.WorkspaceService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().WorkspaceService
}

// GetVMService returns the VM service
func (c *Container) GetVMService() domain.VMService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().VMService
}

// GetDomainService returns the domain service
func (c *Container) GetDomainService() *domain.DomainService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.domainModule.GetContainer().DomainService
}

// GetUserDomainService returns the user domain service
func (c *Container) GetUserDomainService() *domain.UserDomainService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.domainModule.GetContainer().UserDomainService
}

// GetWorkspaceDomainService returns the workspace domain service
func (c *Container) GetWorkspaceDomainService() *domain.WorkspaceDomainService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.domainModule.GetContainer().WorkspaceDomainService
}

// GetVMDomainService returns the VM domain service
func (c *Container) GetVMDomainService() *domain.VMDomainService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.domainModule.GetContainer().VMDomainService
}

// GetBusinessRuleService returns the business rule service
func (c *Container) GetBusinessRuleService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().BusinessRuleService
}

// GetDatabase returns the database connection
func (c *Container) GetDatabase() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// GetCache returns the cache
func (c *Container) GetCache() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache
}

// GetRedisClient returns the Redis client if available
func (c *Container) GetRedisClient() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if cache is RedisService
	if redisService, ok := c.cache.(*cache.RedisService); ok {
		return redisService.GetClient()
	}

	// Redis not available, return nil
	return nil
}

// GetMessaging returns the messaging system
func (c *Container) GetMessaging() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.infrastructureModule.GetContainer().Messaging
}

// GetLogger returns the logger
func (c *Container) GetLogger() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// GetAuthService returns the auth service
func (c *Container) GetAuthService() domain.AuthService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().AuthService
}

// GetCredentialService returns the credential service
func (c *Container) GetCredentialService() domain.CredentialService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().CredentialService
}

// GetRBACService returns the RBAC service
func (c *Container) GetRBACService() domain.RBACService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().RBACService
}

// GetAuditLogService returns the audit log service
func (c *Container) GetAuditLogService() domain.AuditLogService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().AuditLogService
}

// GetOIDCService returns the OIDC service
func (c *Container) GetOIDCService() domain.OIDCService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().OIDCService
}

// GetLogoutService returns the logout service
func (c *Container) GetLogoutService() domain.LogoutService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().LogoutService
}

// GetNotificationService returns the notification service
func (c *Container) GetNotificationService() domain.NotificationService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().NotificationService
}

// GetSystemMonitoringService returns the system monitoring service
func (c *Container) GetSystemMonitoringService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().SystemMonitoringService
}

// GetKubernetesService returns the kubernetes service
func (c *Container) GetKubernetesService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().KubernetesService
}

// GetNetworkService returns the network service
func (c *Container) GetNetworkService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().NetworkService
}

// GetExportService returns the export service
func (c *Container) GetExportService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().ExportService
}

// GetCostAnalysisService returns the cost analysis service
func (c *Container) GetCostAnalysisService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().CostAnalysisService
}

// GetComputeService returns the compute service
func (c *Container) GetComputeService() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceModule.GetContainer().ComputeService
}

// Close closes all resources
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil
	}

	// Close database connection
	if c.db != nil {
		if sqlDB, err := c.db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// TODO: Close cache connection when cache interface has Close method
	// if c.cache != nil {
	// 	c.cache.Close()
	// }

	c.initialized = false
	return nil
}

// IsInitialized returns whether the container is initialized
func (c *Container) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

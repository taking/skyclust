package di

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"
	"skyclust/pkg/config"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/security"
)

// Container holds all dependencies
type Container struct {
	// Repositories
	UserRepo       domain.UserRepository
	CredentialRepo domain.CredentialRepository
	AuditLogRepo   domain.AuditLogRepository
	WorkspaceRepo  domain.WorkspaceRepository
	VMRepo         domain.VMRepository

	// Services
	UserService             domain.UserService
	CredentialService       domain.CredentialService
	AuditLogService         domain.AuditLogService
	AuthService             domain.AuthService
	OIDCService             domain.OIDCService
	LogoutService           *usecase.LogoutService
	PluginActivationService domain.PluginActivationService
	CacheService            domain.CacheService
	EventService            domain.EventService
	WorkspaceService        domain.WorkspaceService
	VMService               domain.VMService
	CostAnalysisService     *usecase.CostAnalysisService
	NotificationService     *usecase.NotificationService
	ExportService           *usecase.ExportService
	RBACService             domain.RBACService

	// HTTP Handlers
	AuthHandler         interface{}
	CredentialHandler   interface{}
	WorkspaceHandler    interface{}
	ProviderHandler     interface{}
	CostAnalysisHandler interface{}
	NotificationHandler interface{}
	ExportHandler       interface{}
	AdminUserHandler    interface{}
	SystemHandler       interface{}
	AuditHandler        interface{}

	// Infrastructure
	DB        *gorm.DB
	EventBus  messaging.Bus
	Hasher    security.PasswordHasher
	Encryptor security.Encryptor
	Logger    *zap.Logger
}

// ContainerBuilder builds the dependency injection container
type ContainerBuilder struct {
	config *config.Config
	ctx    context.Context
}

// NewContainerBuilder creates a new container builder
func NewContainerBuilder(ctx context.Context, cfg *config.Config) *ContainerBuilder {
	return &ContainerBuilder{
		config: cfg,
		ctx:    ctx,
	}
}

// Build creates and initializes the container
func (b *ContainerBuilder) Build() (*Container, error) {
	// Initialize infrastructure first
	infra, err := b.buildInfrastructure()
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repos, err := b.buildRepositories(infra)
	if err != nil {
		return nil, err
	}

	// Initialize services
	services, err := b.buildServices(repos, infra)
	if err != nil {
		return nil, err
	}

	// Initialize handlers
	handlers, err := b.buildHandlers(services, infra)
	if err != nil {
		return nil, err
	}

	return &Container{
		// Repositories
		UserRepo:       repos.UserRepo,
		CredentialRepo: repos.CredentialRepo,
		AuditLogRepo:   repos.AuditLogRepo,
		WorkspaceRepo:  repos.WorkspaceRepo,
		VMRepo:         repos.VMRepo,

		// Services
		UserService:             services.UserService,
		CredentialService:       services.CredentialService,
		AuditLogService:         services.AuditLogService,
		AuthService:             services.AuthService,
		OIDCService:             services.OIDCService,
		LogoutService:           services.LogoutService,
		PluginActivationService: services.PluginActivationService,
		CacheService:            services.CacheService,
		EventService:            services.EventService,
		WorkspaceService:        services.WorkspaceService,
		VMService:               services.VMService,
		CostAnalysisService:     services.CostAnalysisService,
		NotificationService:     services.NotificationService,
		ExportService:           services.ExportService,
		RBACService:             services.RBACService,

		// HTTP Handlers
		AuthHandler:         handlers.AuthHandler,
		CredentialHandler:   handlers.CredentialHandler,
		WorkspaceHandler:    handlers.WorkspaceHandler,
		ProviderHandler:     handlers.ProviderHandler,
		CostAnalysisHandler: handlers.CostAnalysisHandler,
		NotificationHandler: handlers.NotificationHandler,
		ExportHandler:       handlers.ExportHandler,
		AdminUserHandler:    handlers.AdminUserHandler,
		SystemHandler:       handlers.SystemHandler,
		AuditHandler:        handlers.AuditHandler,

		// Infrastructure
		DB:        infra.DB,
		EventBus:  infra.EventBus,
		Hasher:    infra.Hasher,
		Encryptor: infra.Encryptor,
		Logger:    infra.Logger,
	}, nil
}

// Close closes all resources
func (c *Container) Close() error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health checks the health of all services
func (c *Container) Health(ctx context.Context) error {
	// Check database connection
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check event bus
	if err := c.EventBus.Health(ctx); err != nil {
		return fmt.Errorf("event bus health check failed: %w", err)
	}

	return nil
}

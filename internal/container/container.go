package container

import (
	"context"
	"fmt"
	"skyclust/internal/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"skyclust/internal/infrastructure/database"
	events "skyclust/internal/infrastructure/messaging"
	"skyclust/internal/repository/postgres"
	"skyclust/internal/usecase"
	"skyclust/pkg/cache"
	"skyclust/pkg/config"
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

	// Infrastructure
	DB        *gorm.DB
	EventBus  events.Bus
	Hasher    security.PasswordHasher
	Encryptor security.Encryptor
}

// NewContainer creates a new dependency injection container
func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	// Initialize database
	dbConfig := database.PostgresConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
	}

	db, err := database.NewPostgresService(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize event bus
	var eventBus events.Bus
	if cfg.NATS.URL != "" {
		// Try to connect to NATS
		natsService, err := events.NewNATSService(events.NATSConfig{
			URL:     cfg.NATS.URL,
			Cluster: cfg.NATS.Cluster,
		})
		if err != nil {
			// Fallback to local event bus
			eventBus = events.NewLocalBus()
		} else {
			eventBus = natsService
		}
	} else {
		eventBus = events.NewLocalBus()
	}

	// Initialize Redis
	redisService, err := cache.NewRedisService(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize security services
	hasher := security.NewBcryptHasher(cfg.Security.BCryptCost)
	encryptor := security.NewAESEncryptor([]byte(cfg.Encryption.Key))

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.GetDB())
	credentialRepo := postgres.NewCredentialRepository(db.GetDB())
	auditLogRepo := postgres.NewAuditLogRepository(db.GetDB())
	workspaceRepo := postgres.NewWorkspaceRepository(db.GetDB())
	vmRepo := postgres.NewVMRepository(db.GetDB())

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize services
	userService := usecase.NewUserService(userRepo, hasher, auditLogRepo)
	credentialService := usecase.NewCredentialService(credentialRepo, auditLogRepo, encryptor)
	auditLogService := usecase.NewAuditLogService(auditLogRepo)

	// Initialize token blacklist
	tokenBlacklist := cache.NewTokenBlacklist(redisService.GetClient())

	authService := usecase.NewAuthService(userRepo, auditLogRepo, hasher, tokenBlacklist, cfg.Security.JWTSecret, cfg.Security.JWTExpiration)
	oidcService := usecase.NewOIDCService(userRepo, auditLogRepo, authService)
	pluginActivationService := usecase.NewPluginActivationService(credentialRepo, eventBus)
	cacheService := usecase.NewCacheService(redisService)
	eventService := usecase.NewEventService(eventBus)
	workspaceService := usecase.NewWorkspaceService(workspaceRepo, userRepo, eventBus, auditLogRepo)
	vmService := usecase.NewVMService(vmRepo, workspaceRepo, nil, eventBus, auditLogRepo) // Cloud provider will be injected later

	// Initialize logout service
	logoutService := usecase.NewLogoutService(tokenBlacklist, oidcService, auditLogRepo)

	costAnalysisService := usecase.NewCostAnalysisService(vmRepo, credentialRepo, workspaceRepo, auditLogRepo)
	notificationService := usecase.NewNotificationService(logger, auditLogRepo, userRepo, workspaceRepo, eventService)
	exportService := usecase.NewExportService(logger, vmRepo, workspaceRepo, credentialRepo, auditLogRepo)

	return &Container{
		UserRepo:       userRepo,
		CredentialRepo: credentialRepo,
		AuditLogRepo:   auditLogRepo,
		WorkspaceRepo:  workspaceRepo,
		VMRepo:         vmRepo,

		UserService:             userService,
		CredentialService:       credentialService,
		AuditLogService:         auditLogService,
		AuthService:             authService,
		OIDCService:             oidcService,
		LogoutService:           logoutService,
		PluginActivationService: pluginActivationService,
		CacheService:            cacheService,
		EventService:            eventService,
		WorkspaceService:        workspaceService,
		VMService:               vmService,
		CostAnalysisService:     costAnalysisService,
		NotificationService:     notificationService,
		ExportService:           exportService,

		DB:        db.GetDB(),
		EventBus:  eventBus,
		Hasher:    hasher,
		Encryptor: encryptor,
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

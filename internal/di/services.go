package di

import (
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/usecase"
	"skyclust/pkg/cache"
	pkglogger "skyclust/pkg/logger"
)

// Services holds all service dependencies
type Services struct {
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
}

// buildServices initializes all services
func (b *ContainerBuilder) buildServices(repos *Repositories, infra *Infrastructure) (*Services, error) {
	// Initialize services
	userService := usecase.NewUserService(repos.UserRepo, infra.Hasher, repos.AuditLogRepo)
	credentialService := usecase.NewCredentialService(repos.CredentialRepo, repos.AuditLogRepo, infra.Encryptor)
	auditLogService := usecase.NewAuditLogService(repos.AuditLogRepo)

	// Initialize Redis service
	redisService, err := cache.NewRedisService(cache.RedisConfig{
		Host:     b.config.Redis.Host,
		Port:     b.config.Redis.Port,
		Password: b.config.Redis.Password,
		DB:       b.config.Redis.DB,
		PoolSize: b.config.Redis.PoolSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis service: %w", err)
	}

	// Initialize token blacklist
	tokenBlacklist := cache.NewTokenBlacklist(redisService.GetClient())

	authService := usecase.NewAuthService(repos.UserRepo, repos.AuditLogRepo, infra.Hasher, tokenBlacklist, b.config.Security.JWTSecret, b.config.Security.JWTExpiration)
	oidcService := usecase.NewOIDCService(repos.UserRepo, repos.AuditLogRepo, authService)
	pluginActivationService := usecase.NewPluginActivationService(repos.CredentialRepo, infra.EventBus)
	cacheService := usecase.NewCacheService(redisService)
	eventService := usecase.NewEventService(infra.EventBus)
	workspaceService := usecase.NewWorkspaceService(repos.WorkspaceRepo, repos.UserRepo, infra.EventBus, repos.AuditLogRepo)
	vmService := usecase.NewVMService(repos.VMRepo, repos.WorkspaceRepo, nil, infra.EventBus, repos.AuditLogRepo) // Cloud provider will be injected later

	// Initialize logout service
	logoutService := usecase.NewLogoutService(tokenBlacklist, oidcService, repos.AuditLogRepo)

	// Initialize RBAC service
	rbacService := usecase.NewRBACService(infra.DB)

	costAnalysisService := usecase.NewCostAnalysisService(repos.VMRepo, repos.CredentialRepo, repos.WorkspaceRepo, repos.AuditLogRepo)

	// Create logger for services that need pkg/logger.Logger
	_, _ = pkglogger.NewLogger(&pkglogger.LoggerConfig{})
	notificationService := usecase.NewNotificationService(infra.Logger, repos.AuditLogRepo, repos.UserRepo, repos.WorkspaceRepo, eventService)
	exportService := usecase.NewExportService(infra.Logger, repos.VMRepo, repos.WorkspaceRepo, repos.CredentialRepo, repos.AuditLogRepo)

	return &Services{
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
		RBACService:             rbacService,
	}, nil
}

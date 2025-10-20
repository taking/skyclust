package di

import (
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/service"
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
	LogoutService           *service.LogoutService
	PluginActivationService domain.PluginActivationService
	CacheService            domain.CacheService
	EventService            domain.EventService
	WorkspaceService        domain.WorkspaceService
	VMService               domain.VMService
	CostAnalysisService     *service.CostAnalysisService
	NotificationService     *service.NotificationService
	ExportService           *service.ExportService
	RBACService             domain.RBACService
}

// buildServices initializes all services
func (b *ContainerBuilder) buildServices(repos *Repositories, infra *Infrastructure) (*Services, error) {
	// Initialize services
	userService := service.NewUserService(repos.UserRepo, infra.Hasher, repos.AuditLogRepo)
	credentialService := service.NewCredentialService(repos.CredentialRepo, repos.AuditLogRepo, infra.Encryptor)
	auditLogService := service.NewAuditLogService(repos.AuditLogRepo)

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

	// Initialize RBAC service first
	rbacService := service.NewRBACService(infra.DB)

	authService := service.NewAuthService(repos.UserRepo, repos.AuditLogRepo, rbacService, infra.Hasher, tokenBlacklist, b.config.Security.JWTSecret, b.config.Security.JWTExpiration)
	oidcService := service.NewOIDCService(repos.UserRepo, repos.AuditLogRepo, authService)
	pluginActivationService := service.NewPluginActivationService(repos.CredentialRepo, infra.EventBus)
	cacheService := service.NewCacheService(redisService)
	eventService := service.NewEventService(infra.EventBus)
	workspaceService := service.NewWorkspaceService(repos.WorkspaceRepo, repos.UserRepo, infra.EventBus, repos.AuditLogRepo)
	vmService := service.NewVMService(repos.VMRepo, repos.WorkspaceRepo, nil, infra.EventBus, repos.AuditLogRepo) // Cloud provider will be injected later

	// Initialize logout service
	logoutService := service.NewLogoutService(tokenBlacklist, oidcService, repos.AuditLogRepo)

	costAnalysisService := service.NewCostAnalysisService(repos.VMRepo, repos.CredentialRepo, repos.WorkspaceRepo, repos.AuditLogRepo)

	// Create logger for services that need pkg/logger.Logger
	_, _ = pkglogger.NewLogger(&pkglogger.LoggerConfig{})
	notificationService := service.NewNotificationService(infra.Logger, repos.AuditLogRepo, repos.UserRepo, repos.WorkspaceRepo, eventService)
	exportService := service.NewExportService(infra.Logger, repos.VMRepo, repos.WorkspaceRepo, repos.CredentialRepo, repos.AuditLogRepo)

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

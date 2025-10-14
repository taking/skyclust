package di

import (
	"skyclust/internal/api/admin"
	"skyclust/internal/api/audit"
	"skyclust/internal/api/auth"
	"skyclust/internal/api/cost_analysis"
	"skyclust/internal/api/credential"
	"skyclust/internal/api/export"
	"skyclust/internal/api/notification"
	"skyclust/internal/api/provider"
	"skyclust/internal/api/system"
	"skyclust/internal/api/workspace"
	pkglogger "skyclust/pkg/logger"
)

// Handlers holds all HTTP handler dependencies
type Handlers struct {
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
}

// buildHandlers initializes all HTTP handlers
func (b *ContainerBuilder) buildHandlers(services *Services, infra *Infrastructure) (*Handlers, error) {
	// Create logger for handlers that need pkg/logger.Logger
	pkgLog, _ := pkglogger.NewLogger(&pkglogger.LoggerConfig{})

	// Initialize HTTP handlers
	authHandler := auth.NewHandlerWithLogout(services.AuthService, services.UserService, services.LogoutService)
	credentialHandler := credential.NewHandler(services.CredentialService)
	workspaceHandler := workspace.NewHandler(services.WorkspaceService, services.UserService)

	// TODO: Initialize plugin manager properly
	// providerHandler := provider.NewHandler(pluginManager, auditLogRepo)
	providerHandler := &provider.Handler{} // Placeholder

	costAnalysisHandler := cost_analysis.NewHandler()
	notificationHandler := notification.NewHandler(services.NotificationService)
	exportHandler := export.NewHandler()
	adminUserHandler := admin.NewHandler(services.UserService, services.RBACService, pkgLog)
	systemHandler := system.NewHandler(pkgLog)
	auditHandler := audit.NewHandler(services.AuditLogService)

	return &Handlers{
		AuthHandler:         authHandler,
		CredentialHandler:   credentialHandler,
		WorkspaceHandler:    workspaceHandler,
		ProviderHandler:     providerHandler,
		CostAnalysisHandler: costAnalysisHandler,
		NotificationHandler: notificationHandler,
		ExportHandler:       exportHandler,
		AdminUserHandler:    adminUserHandler,
		SystemHandler:       systemHandler,
		AuditHandler:        auditHandler,
	}, nil
}

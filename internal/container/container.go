package container

import (
	"context"
	"fmt"

	"cmp/internal/domain"
	"cmp/internal/infrastructure/database"
	events "cmp/internal/infrastructure/messaging"
	"cmp/internal/repository/postgres"
	"cmp/internal/usecase"
	"cmp/pkg/shared/config"
	"cmp/pkg/shared/security"
	"gorm.io/gorm"
)

// Container holds all dependencies
type Container struct {
	// Repositories
	UserRepo      domain.UserRepository
	WorkspaceRepo domain.WorkspaceRepository
	VMRepo        domain.VMRepository

	// Services
	UserService      domain.UserService
	WorkspaceService domain.WorkspaceService
	VMService        domain.VMService

	// Infrastructure
	DB        *gorm.DB
	EventBus  events.Bus
	Hasher    security.PasswordHasher
	Encryptor security.Encryptor
}

// NewContainer creates a new dependency injection container
func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	// Initialize database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db := database.NewPostgresService(dbConfig)

	// Initialize event bus
	var eventBus events.Bus
	if cfg.NATS.URL != "" {
		// Try to connect to NATS
		natsService := events.NewNATSService(cfg.NATS.URL)
		if err := natsService.Connect(); err != nil {
			// Fallback to local event bus
			eventBus = events.NewLocalBus()
		} else {
			eventBus = events.NewNATSBus(natsService.GetConnection())
		}
	} else {
		eventBus = events.NewLocalBus()
	}

	// Initialize security services
	hasher := security.NewBcryptHasher()
	encryptor, _ := security.NewAESEncryptor(cfg.Encryption.Key)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.GetDB())
	workspaceRepo := postgres.NewWorkspaceRepository(db.GetDB())
	vmRepo := postgres.NewVMRepository(db.GetDB())

	// Initialize services
	userService := usecase.NewUserService(userRepo, hasher)
	workspaceService := usecase.NewWorkspaceService(workspaceRepo, userRepo)
	vmService := usecase.NewVMService(vmRepo, workspaceRepo, nil, eventBus) // Cloud provider will be injected later

	return &Container{
		UserRepo:         userRepo,
		WorkspaceRepo:    workspaceRepo,
		VMRepo:           vmRepo,
		UserService:      userService,
		WorkspaceService: workspaceService,
		VMService:        vmService,
		DB:               db.GetDB(),
		EventBus:         eventBus,
		Hasher:           hasher,
		Encryptor:        encryptor,
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

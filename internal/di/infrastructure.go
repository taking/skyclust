package di

import (
	"fmt"
	"skyclust/internal/infrastructure/database"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"
	"skyclust/pkg/security"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Infrastructure holds infrastructure dependencies
type Infrastructure struct {
	DB        *gorm.DB
	EventBus  messaging.Bus
	Hasher    security.PasswordHasher
	Encryptor security.Encryptor
	Logger    *zap.Logger
}

// buildInfrastructure initializes all infrastructure components
func (b *ContainerBuilder) buildInfrastructure() (*Infrastructure, error) {
	// Initialize database
	dbConfig := database.PostgresConfig{
		Host:     b.config.Database.Host,
		Port:     b.config.Database.Port,
		User:     b.config.Database.User,
		Password: b.config.Database.Password,
		Database: b.config.Database.Name,
		SSLMode:  b.config.Database.SSLMode,
		MaxConns: b.config.Database.MaxConns,
		MinConns: b.config.Database.MinConns,
	}

	db, err := database.NewPostgresService(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize event bus
	var eventBus messaging.Bus
	if b.config.NATS.URL != "" {
		// Try to connect to NATS
		natsService, err := messaging.NewNATSService(messaging.NATSConfig{
			URL:     b.config.NATS.URL,
			Cluster: b.config.NATS.Cluster,
		})
		if err != nil {
			// Fallback to local event bus
			eventBus = messaging.NewLocalBus()
		} else {
			eventBus = natsService
		}
	} else {
		eventBus = messaging.NewLocalBus()
	}

	// Initialize Redis
	_, err = cache.NewRedisService(cache.RedisConfig{
		Host:     b.config.Redis.Host,
		Port:     b.config.Redis.Port,
		Password: b.config.Redis.Password,
		DB:       b.config.Redis.DB,
		PoolSize: b.config.Redis.PoolSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize security services
	hasher := security.NewBcryptHasher(b.config.Security.BCryptCost)
	encryptor := security.NewAESEncryptor([]byte(b.config.Encryption.Key))

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return &Infrastructure{
		DB:        db.GetDB(),
		EventBus:  eventBus,
		Hasher:    hasher,
		Encryptor: encryptor,
		Logger:    logger,
	}, nil
}

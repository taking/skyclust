package database

import (
	"fmt"
	"skyclust/internal/domain"
	pkglogger "skyclust/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int
	MinConns        int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	SlowQueryLog    bool
	SlowQueryTime   time.Duration
}

// PostgresService implements the database service interface
type PostgresService struct {
	db        *gorm.DB
	config    PostgresConfig
	optimizer *DatabaseOptimizer
}

// NewPostgresService creates a new PostgreSQL service
func NewPostgresService(config PostgresConfig) (*PostgresService, error) {
	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	// Configure GORM logger - use silent mode for now
	logger := gormLogger.Default.LogMode(gormLogger.Silent)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statements for better performance
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool with optimized settings
	sqlDB.SetMaxOpenConns(config.MaxConns)
	sqlDB.SetMaxIdleConns(config.MinConns)

	// Set connection lifetime (default to 1 hour if not specified)
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Set idle connection timeout (default to 30 minutes if not specified)
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 30 * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	service := &PostgresService{
		db:        db,
		config:    config,
		optimizer: NewDatabaseOptimizer(db),
	}

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		// Log warning but don't fail startup - some databases might not support this
		pkglogger.DefaultLogger.GetLogger().Warn("Failed to enable uuid-ossp extension", zap.Error(err))
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Workspace{},
		&domain.Credential{},
		&domain.AuditLog{},
		&domain.WorkspaceUser{},
		&domain.UserRole{},
		&domain.RolePermission{},
		&domain.OIDCProvider{},
		&domain.Notification{},
		&domain.NotificationPreferences{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate schema: %w", err)
	}

	// Create performance indexes
	if err := service.optimizer.CreateIndexes(); err != nil {
		// Log warning but don't fail startup
		pkglogger.DefaultLogger.GetLogger().Warn("Failed to create some indexes", zap.Error(err))
	}

	// Optimize queries
	if err := service.optimizer.OptimizeQueries(); err != nil {
		// Log warning but don't fail startup
		pkglogger.DefaultLogger.GetLogger().Warn("Failed to optimize queries", zap.Error(err))
	}

	pkglogger.DefaultLogger.GetLogger().Info("Successfully connected to PostgreSQL database")
	return service, nil
}

// GetDB returns the GORM database instance
func (p *PostgresService) GetDB() *gorm.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresService) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks the database health
func (p *PostgresService) Health() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats returns database performance statistics
func (p *PostgresService) GetStats() (*DatabaseStats, error) {
	return p.optimizer.GetDatabaseStats()
}

// Optimize performs database optimization tasks
func (p *PostgresService) Optimize() error {
	return p.optimizer.OptimizeQueries()
}

// CleanupOldData removes old data to maintain performance
func (p *PostgresService) CleanupOldData() error {
	return p.optimizer.CleanupOldData()
}

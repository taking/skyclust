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

	// Migrate credentials table from user-based to workspace-based
	if err := migrateCredentialsToWorkspace(db); err != nil {
		pkglogger.DefaultLogger.GetLogger().Warn("Failed to migrate credentials to workspace-based", zap.Error(err))
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
		&domain.VM{},
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

// migrateCredentialsToWorkspace migrates credentials from user-based to workspace-based
func migrateCredentialsToWorkspace(db *gorm.DB) error {
	// Check if credentials table exists
	var hasTable bool
	if err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'credentials')").Scan(&hasTable).Error; err != nil {
		return err
	}

	if !hasTable {
		// Table doesn't exist, migration will be handled by AutoMigrate
		return nil
	}

	// Check if workspace_id column exists
	var hasWorkspaceID bool
	if err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'credentials' AND column_name = 'workspace_id')").Scan(&hasWorkspaceID).Error; err != nil {
		return err
	}

	if !hasWorkspaceID {
		// Add workspace_id column (nullable initially)
		if err := db.Exec("ALTER TABLE credentials ADD COLUMN IF NOT EXISTS workspace_id UUID").Error; err != nil {
			return fmt.Errorf("failed to add workspace_id column: %w", err)
		}

		// Add created_by column (nullable initially)
		if err := db.Exec("ALTER TABLE credentials ADD COLUMN IF NOT EXISTS created_by UUID").Error; err != nil {
			return fmt.Errorf("failed to add created_by column: %w", err)
		}

		// Migrate existing data: create a default workspace for each user and assign credentials
		// Create workspaces for users who have credentials but no workspace
		createWorkspacesSQL := `
			INSERT INTO workspaces (id, name, description, owner_id, created_at, updated_at)
			SELECT DISTINCT
				gen_random_uuid(),
				u.username || '''s Workspace',
				'Default workspace migrated from user credentials',
				u.id,
				NOW(),
				NOW()
			FROM users u
			WHERE EXISTS (
				SELECT 1 FROM credentials c WHERE c.user_id = u.id
			)
			AND NOT EXISTS (
				SELECT 1 FROM workspaces w WHERE w.owner_id = u.id
			)
		`
		if err := db.Exec(createWorkspacesSQL).Error; err != nil {
			pkglogger.DefaultLogger.GetLogger().Warn("Failed to create workspaces for credential migration", zap.Error(err))
		}

		// Update credentials with workspace_id and created_by
		updateCredentialsSQL := `
			UPDATE credentials c
			SET 
				workspace_id = w.id,
				created_by = c.user_id
			FROM workspaces w
			WHERE w.owner_id = c.user_id
			AND c.workspace_id IS NULL
		`
		if err := db.Exec(updateCredentialsSQL).Error; err != nil {
			pkglogger.DefaultLogger.GetLogger().Warn("Failed to update credentials with workspace_id", zap.Error(err))
		}

		// For credentials where user doesn't have a workspace, use the first available workspace
		fallbackUpdateSQL := `
			UPDATE credentials c
			SET 
				workspace_id = (SELECT id FROM workspaces LIMIT 1),
				created_by = c.user_id
			WHERE c.workspace_id IS NULL
			AND EXISTS (SELECT 1 FROM workspaces)
		`
		if err := db.Exec(fallbackUpdateSQL).Error; err != nil {
			pkglogger.DefaultLogger.GetLogger().Warn("Failed to set fallback workspace_id for credentials", zap.Error(err))
		}

		// Delete credentials that couldn't be migrated (no workspace available)
		if err := db.Exec("DELETE FROM credentials WHERE workspace_id IS NULL").Error; err != nil {
			pkglogger.DefaultLogger.GetLogger().Warn("Failed to delete unmigrated credentials", zap.Error(err))
		}
	}

	// Now make workspace_id NOT NULL if it's still nullable
	var isNullable string
	if err := db.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'credentials' AND column_name = 'workspace_id'").Scan(&isNullable).Error; err == nil {
		if isNullable == "YES" {
			// Set default workspace_id for any remaining null values
			if err := db.Exec(`
				UPDATE credentials 
				SET workspace_id = (SELECT id FROM workspaces LIMIT 1)
				WHERE workspace_id IS NULL
			`).Error; err != nil {
				pkglogger.DefaultLogger.GetLogger().Warn("Failed to set default workspace_id", zap.Error(err))
			}

			// Make workspace_id NOT NULL
			if err := db.Exec("ALTER TABLE credentials ALTER COLUMN workspace_id SET NOT NULL").Error; err != nil {
				pkglogger.DefaultLogger.GetLogger().Warn("Failed to set workspace_id NOT NULL", zap.Error(err))
			}
		}
	}

	// Check if user_id column still exists and remove it completely
	var hasUserID bool
	if err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'credentials' AND column_name = 'user_id')").Scan(&hasUserID).Error; err == nil {
		if hasUserID {
			// Check if there are any foreign key constraints referencing user_id
			var hasFK bool
			fkCheckSQL := `
				SELECT EXISTS (
					SELECT 1 
					FROM information_schema.table_constraints 
					WHERE table_name = 'credentials' 
					AND constraint_type = 'FOREIGN KEY'
					AND constraint_name LIKE '%user_id%'
				)
			`
			if err := db.Raw(fkCheckSQL).Scan(&hasFK).Error; err == nil && hasFK {
				// Find and drop foreign key constraint
				var constraintName string
				dropFKSQL := `
					SELECT constraint_name
					FROM information_schema.table_constraints
					WHERE table_name = 'credentials'
					AND constraint_type = 'FOREIGN KEY'
					AND constraint_name LIKE '%user_id%'
					LIMIT 1
				`
				if err := db.Raw(dropFKSQL).Scan(&constraintName).Error; err == nil && constraintName != "" {
					if err := db.Exec(fmt.Sprintf("ALTER TABLE credentials DROP CONSTRAINT IF EXISTS %s", constraintName)).Error; err != nil {
						pkglogger.DefaultLogger.GetLogger().Warn("Failed to drop foreign key constraint on user_id", zap.Error(err))
					}
				}
			}

			// Drop the user_id column
			if err := db.Exec("ALTER TABLE credentials DROP COLUMN IF EXISTS user_id").Error; err != nil {
				pkglogger.DefaultLogger.GetLogger().Warn("Failed to drop user_id column", zap.Error(err))
			} else {
				pkglogger.DefaultLogger.GetLogger().Info("Successfully dropped user_id column from credentials table")
			}

			// Also drop any indexes on user_id
			if err := db.Exec(`
				DROP INDEX IF EXISTS idx_credentials_user_id;
				DROP INDEX IF EXISTS idx_credentials_user_provider;
			`).Error; err != nil {
				pkglogger.DefaultLogger.GetLogger().Warn("Failed to drop user_id indexes (may not exist)", zap.Error(err))
			}
		}
	}

	return nil
}

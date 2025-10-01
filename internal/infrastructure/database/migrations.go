package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cmp/pkg/shared/logger"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db         *sql.DB
	migrations []Migration
	tableName  string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db:        db,
		tableName: "schema_migrations",
	}
}

// AddMigration adds a migration to the manager
func (m *MigrationManager) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// Migrate runs all pending migrations
func (m *MigrationManager) Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if !m.isMigrationApplied(applied, migration.Version) {
			logger.Info(fmt.Sprintf("Running migration version %d: %s", migration.Version, migration.Description))

			if err := migration.Up(m.db); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}

			if err := m.recordMigration(ctx, migration.Version, migration.Description); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
			}

			logger.Info(fmt.Sprintf("Migration completed version %d", migration.Version))
		}
	}

	return nil
}

// Rollback rolls back the last migration
func (m *MigrationManager) Rollback(ctx context.Context) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the last applied migration
	lastMigration := applied[len(applied)-1]

	// Find the migration definition
	var migration *Migration
	for _, m := range m.migrations {
		if m.Version == lastMigration.Version {
			migration = &m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %d not found", lastMigration.Version)
	}

	logger.Info(fmt.Sprintf("Rolling back migration version %d: %s", migration.Version, migration.Description))

	if err := migration.Down(m.db); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
	}

	if err := m.removeMigration(ctx, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
	}

	logger.Info(fmt.Sprintf("Migration rollback completed version %d", migration.Version))
	return nil
}

// GetStatus returns the status of all migrations
func (m *MigrationManager) GetStatus(ctx context.Context) ([]MigrationStatus, error) {
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var status []MigrationStatus
	for _, migration := range m.migrations {
		appliedTime := m.getMigrationAppliedTime(applied, migration.Version)
		status = append(status, MigrationStatus{
			Version:     migration.Version,
			Description: migration.Description,
			Applied:     appliedTime != nil,
			AppliedAt:   appliedTime,
		})
	}

	return status, nil
}

// createMigrationsTable creates the migrations table
func (m *MigrationManager) createMigrationsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations gets all applied migrations
func (m *MigrationManager) getAppliedMigrations(ctx context.Context) ([]AppliedMigration, error) {
	query := fmt.Sprintf(`
		SELECT version, description, applied_at
		FROM %s
		ORDER BY version
	`, m.tableName)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []AppliedMigration
	for rows.Next() {
		var migration AppliedMigration
		err := rows.Scan(&migration.Version, &migration.Description, &migration.AppliedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// isMigrationApplied checks if a migration is applied
func (m *MigrationManager) isMigrationApplied(applied []AppliedMigration, version int) bool {
	for _, migration := range applied {
		if migration.Version == version {
			return true
		}
	}
	return false
}

// recordMigration records a migration as applied
func (m *MigrationManager) recordMigration(ctx context.Context, version int, description string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, applied_at)
		VALUES ($1, $2, $3)
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, query, version, description, time.Now())
	return err
}

// removeMigration removes a migration record
func (m *MigrationManager) removeMigration(ctx context.Context, version int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = $1
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, query, version)
	return err
}

// getMigrationAppliedTime gets the applied time for a migration
func (m *MigrationManager) getMigrationAppliedTime(applied []AppliedMigration, version int) *time.Time {
	for _, migration := range applied {
		if migration.Version == version {
			return &migration.AppliedAt
		}
	}
	return nil
}

// AppliedMigration represents an applied migration
type AppliedMigration struct {
	Version     int
	Description string
	AppliedAt   time.Time
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     int
	Description string
	Applied     bool
	AppliedAt   *time.Time
}

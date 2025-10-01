package migration

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"cmp/pkg/shared/logger"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.Tx) error
	Down        func(*sql.Tx) error
	CreatedAt   time.Time
}

// Migrator manages database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
	tableName  string
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB, tableName string) *Migrator {
	if tableName == "" {
		tableName = "schema_migrations"
	}

	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
		tableName:  tableName,
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// AddMigrations adds multiple migrations
func (m *Migrator) AddMigrations(migrations ...Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			continue
		}

		logger.Infof("Running migration: %s - %s", migration.Version, migration.Description)

		if err := m.runMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}

		logger.Infof("Migration completed: %s", migration.Version)
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback(ctx context.Context) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the last applied migration
	var lastMigration *Migration
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			lastMigration = &migration
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("no migration found to rollback")
	}

	logger.Infof("Rolling back migration: %s - %s", lastMigration.Version, lastMigration.Description)

	if err := m.rollbackMigration(ctx, *lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Version, err)
	}

	logger.Infof("Migration rollback completed: %s", lastMigration.Version)
	return nil
}

// Status returns the status of migrations
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var status []MigrationStatus
	for _, migration := range m.migrations {
		appliedAt, isApplied := applied[migration.Version]
		status = append(status, MigrationStatus{
			Version:     migration.Version,
			Description: migration.Description,
			Applied:     isApplied,
			AppliedAt:   appliedAt,
		})
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     string
	Description string
	Applied     bool
	AppliedAt   time.Time
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations returns a map of applied migrations
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]time.Time, error) {
	query := fmt.Sprintf("SELECT version, applied_at FROM %s ORDER BY version", m.tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}

	return applied, nil
}

// runMigration runs a single migration
func (m *Migrator) runMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Warnf("Failed to rollback transaction: %v", err)
		}
	}()

	// Run the migration
	if err := migration.Up(tx); err != nil {
		return err
	}

	// Record the migration
	query := fmt.Sprintf("INSERT INTO %s (version, applied_at) VALUES (?, ?)", m.tableName)
	_, err = tx.ExecContext(ctx, query, migration.Version, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Warnf("Failed to rollback transaction: %v", err)
		}
	}()

	// Run the rollback
	if err := migration.Down(tx); err != nil {
		return err
	}

	// Remove the migration record
	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName)
	_, err = tx.ExecContext(ctx, query, migration.Version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Reset drops all tables and recreates them
func (m *Migrator) Reset(ctx context.Context) error {
	logger.Warn("Resetting database - this will drop all tables!")

	// Get all table names
	query := `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' AND tablename != 'schema_migrations'
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	// Drop all tables
	for _, table := range tables {
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := m.db.ExecContext(ctx, dropQuery); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Clear migrations table
	clearQuery := fmt.Sprintf("DELETE FROM %s", m.tableName)
	if _, err := m.db.ExecContext(ctx, clearQuery); err != nil {
		return fmt.Errorf("failed to clear migrations table: %w", err)
	}

	logger.Info("Database reset completed")
	return nil
}

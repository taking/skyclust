package migrations

import (
	"database/sql"
	"fmt"
)

// CreateWorkspacesTable creates the workspaces table
func CreateWorkspacesTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS workspaces (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			owner_id VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create workspaces table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id)",
		"CREATE INDEX IF NOT EXISTS idx_workspaces_created_at ON workspaces(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_workspaces_name ON workspaces(name)",
	}

	for _, indexQuery := range indexes {
		if _, err := db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropWorkspacesTable drops the workspaces table
func DropWorkspacesTable(db *sql.DB) error {
	query := `DROP TABLE IF EXISTS workspaces`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop workspaces table: %w", err)
	}
	return nil
}

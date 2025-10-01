package migrations

import (
	"database/sql"
	"fmt"
)

// CreateVMsTable creates the vms table
func CreateVMsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS vms (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			workspace_id VARCHAR(36) NOT NULL,
			provider VARCHAR(50) NOT NULL,
			instance_id VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			type VARCHAR(50) NOT NULL,
			region VARCHAR(50) NOT NULL,
			image_id VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create vms table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_vms_workspace_id ON vms(workspace_id)",
		"CREATE INDEX IF NOT EXISTS idx_vms_provider ON vms(provider)",
		"CREATE INDEX IF NOT EXISTS idx_vms_status ON vms(status)",
		"CREATE INDEX IF NOT EXISTS idx_vms_created_at ON vms(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_vms_instance_id ON vms(instance_id)",
	}

	for _, indexQuery := range indexes {
		if _, err := db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropVMsTable drops the vms table
func DropVMsTable(db *sql.DB) error {
	query := `DROP TABLE IF EXISTS vms`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop vms table: %w", err)
	}
	return nil
}

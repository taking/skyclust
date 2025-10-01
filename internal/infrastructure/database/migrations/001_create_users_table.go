package migrations

import (
	"database/sql"
	"fmt"
)

// CreateUsersTable creates the users table
func CreateUsersTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
	}

	for _, indexQuery := range indexes {
		if _, err := db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropUsersTable drops the users table
func DropUsersTable(db *sql.DB) error {
	query := `DROP TABLE IF EXISTS users`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop users table: %w", err)
	}
	return nil
}

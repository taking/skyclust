package database

import (
	"time"
)

// User represents a user in the database
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Workspace represents a workspace in the database
type Workspace struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	OwnerID   string                 `json:"owner_id" db:"owner_id"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// WorkspaceUser represents a user in a workspace
type WorkspaceUser struct {
	UserID      string    `json:"user_id" db:"user_id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// Credentials represents encrypted credentials for a cloud provider
type Credentials struct {
	ID          string            `json:"id" db:"id"`
	WorkspaceID string            `json:"workspace_id" db:"workspace_id"`
	Provider    string            `json:"provider" db:"provider"`
	Encrypted   []byte            `json:"-" db:"encrypted_data"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// Execution represents an OpenTofu execution
type Execution struct {
	ID          string     `json:"id" db:"id"`
	WorkspaceID string     `json:"workspace_id" db:"workspace_id"`
	Command     string     `json:"command" db:"command"`
	Status      string     `json:"status" db:"status"`
	Output      string     `json:"output" db:"output"`
	Error       string     `json:"error" db:"error"`
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

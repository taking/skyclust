package database

import (
	"time"
)

// User represents a user in the database with GORM tags
type GormUser struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"column:password_hash;not null" json:"-"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for User
func (GormUser) TableName() string {
	return "users"
}

// Workspace represents a workspace in the database with GORM tags
type GormWorkspace struct {
	ID        string                 `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string                 `gorm:"uniqueIndex;not null" json:"name"`
	OwnerID   string                 `gorm:"not null" json:"owner_id"`
	Settings  map[string]interface{} `gorm:"type:jsonb" json:"settings"`
	CreatedAt time.Time              `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Workspace
func (GormWorkspace) TableName() string {
	return "workspaces"
}

// WorkspaceUser represents a user in a workspace with GORM tags
type GormWorkspaceUser struct {
	UserID      string    `gorm:"primaryKey;type:uuid" json:"user_id"`
	WorkspaceID string    `gorm:"primaryKey;type:uuid" json:"workspace_id"`
	Role        string    `gorm:"not null" json:"role"`
	JoinedAt    time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// TableName specifies the table name for WorkspaceUser
func (GormWorkspaceUser) TableName() string {
	return "workspace_users"
}

// Credentials represents encrypted credentials for a cloud provider with GORM tags
type GormCredentials struct {
	ID          string            `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WorkspaceID string            `gorm:"not null;type:uuid" json:"workspace_id"`
	Provider    string            `gorm:"not null" json:"provider"`
	Encrypted   []byte            `gorm:"type:bytea;not null" json:"-"`
	Metadata    map[string]string `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Credentials
func (GormCredentials) TableName() string {
	return "credentials"
}

// Execution represents an OpenTofu execution with GORM tags
type GormExecution struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WorkspaceID string     `gorm:"not null;type:uuid" json:"workspace_id"`
	Command     string     `gorm:"not null" json:"command"`
	Status      string     `gorm:"not null" json:"status"`
	Output      string     `gorm:"type:text" json:"output"`
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   time.Time  `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// TableName specifies the table name for Execution
func (GormExecution) TableName() string {
	return "executions"
}

// Token represents JWT tokens for session management with GORM tags
type GormToken struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"column:token_hash;not null" json:"token"`
	UserID    string    `gorm:"not null;type:uuid" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for Token
func (GormToken) TableName() string {
	return "tokens"
}

// Conversion methods to maintain compatibility with existing interfaces

// ToUser converts GormUser to User
func (g *GormUser) ToUser() *User {
	return &User{
		ID:        g.ID,
		Email:     g.Email,
		Password:  g.Password,
		Name:      g.Name,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

// ToGormUser converts User to GormUser
func (u *User) ToGormUser() *GormUser {
	return &GormUser{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToWorkspace converts GormWorkspace to Workspace
func (g *GormWorkspace) ToWorkspace() *Workspace {
	return &Workspace{
		ID:        g.ID,
		Name:      g.Name,
		OwnerID:   g.OwnerID,
		Settings:  g.Settings,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

// ToGormWorkspace converts Workspace to GormWorkspace
func (w *Workspace) ToGormWorkspace() *GormWorkspace {
	return &GormWorkspace{
		ID:        w.ID,
		Name:      w.Name,
		OwnerID:   w.OwnerID,
		Settings:  w.Settings,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

// ToCredentials converts GormCredentials to Credentials
func (g *GormCredentials) ToCredentials() *Credentials {
	return &Credentials{
		ID:          g.ID,
		WorkspaceID: g.WorkspaceID,
		Provider:    g.Provider,
		Encrypted:   g.Encrypted,
		Metadata:    g.Metadata,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// ToGormCredentials converts Credentials to GormCredentials
func (c *Credentials) ToGormCredentials() *GormCredentials {
	return &GormCredentials{
		ID:          c.ID,
		WorkspaceID: c.WorkspaceID,
		Provider:    c.Provider,
		Encrypted:   c.Encrypted,
		Metadata:    c.Metadata,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToExecution converts GormExecution to Execution
func (g *GormExecution) ToExecution() *Execution {
	return &Execution{
		ID:          g.ID,
		WorkspaceID: g.WorkspaceID,
		Command:     g.Command,
		Status:      g.Status,
		Output:      g.Output,
		Error:       g.Error,
		StartedAt:   g.StartedAt,
		CompletedAt: g.CompletedAt,
	}
}

// ToGormExecution converts Execution to GormExecution
func (e *Execution) ToGormExecution() *GormExecution {
	return &GormExecution{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		Command:     e.Command,
		Status:      e.Status,
		Output:      e.Output,
		Error:       e.Error,
		StartedAt:   e.StartedAt,
		CompletedAt: e.CompletedAt,
	}
}

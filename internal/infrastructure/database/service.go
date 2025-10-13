package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database errors
var (
	ErrNotFound = errors.New("record not found")
)

// Service defines the database service interface
type Service interface {
	// GetDB returns the underlying GORM database instance
	GetDB() *gorm.DB
	// User management
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error

	// Token management
	StoreToken(ctx context.Context, userID, token string) error
	ValidateToken(ctx context.Context, token string) (string, error)
	DeleteToken(ctx context.Context, token string) error

	// Workspace management
	CreateWorkspace(ctx context.Context, workspace *Workspace) error
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	GetWorkspaceByID(ctx context.Context, workspaceID string) (*Workspace, error)
	ListWorkspacesByUser(ctx context.Context, userID string) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *Workspace) error
	DeleteWorkspace(ctx context.Context, workspaceID string) error

	// User-Workspace relationship
	AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error
	RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error
	GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error)
	GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error)

	// Credentials management
	CreateCredentials(ctx context.Context, cred *Credentials) error
	GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error)
	ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error)
	UpdateCredentials(ctx context.Context, cred *Credentials) error
	DeleteCredentials(ctx context.Context, workspaceID, credID string) error

	// Execution management
	CreateExecution(ctx context.Context, execution *Execution) error
	GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error)
	ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error)
	UpdateExecution(ctx context.Context, execution *Execution) error
	UpdateExecutionStatus(ctx context.Context, workspaceID, executionID, status string) error

	// State management
	GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error)
	SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error

	// Health check
	Ping(ctx context.Context) error
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewService creates a new database service with actual PostgreSQL connection
func NewService(config Config) (Service, error) {
	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&User{},
		&Workspace{},
		&WorkspaceUser{},
		&Credentials{},
		&Execution{},
		&AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &postgresService{db: db}, nil
}

// postgresService implements the database service using PostgreSQL
type postgresService struct {
	db *gorm.DB
}

func (p *postgresService) GetDB() *gorm.DB {
	return p.db
}

func (p *postgresService) CreateUser(ctx context.Context, user *User) error {
	return p.db.WithContext(ctx).Create(user).Error
}

func (p *postgresService) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	err := p.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (p *postgresService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := p.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (p *postgresService) UpdateUser(ctx context.Context, user *User) error {
	return p.db.WithContext(ctx).Save(user).Error
}

func (p *postgresService) DeleteUser(ctx context.Context, userID string) error {
	return p.db.WithContext(ctx).Where("id = ?", userID).Delete(&User{}).Error
}

func (p *postgresService) StoreToken(ctx context.Context, userID, token string) error {
	// Implement token storage logic
	return nil
}

func (p *postgresService) ValidateToken(ctx context.Context, token string) (string, error) {
	// Implement token validation logic
	return "", nil
}

func (p *postgresService) DeleteToken(ctx context.Context, token string) error {
	// Implement token deletion logic
	return nil
}

func (p *postgresService) CreateWorkspace(ctx context.Context, workspace *Workspace) error {
	return p.db.WithContext(ctx).Create(workspace).Error
}

func (p *postgresService) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	var workspace Workspace
	err := p.db.WithContext(ctx).Where("id = ?", workspaceID).First(&workspace).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &workspace, nil
}

func (p *postgresService) GetWorkspaceByID(ctx context.Context, workspaceID string) (*Workspace, error) {
	return p.GetWorkspace(ctx, workspaceID)
}

func (p *postgresService) ListWorkspacesByUser(ctx context.Context, userID string) ([]*Workspace, error) {
	var workspaces []*Workspace
	err := p.db.WithContext(ctx).Where("owner_id = ?", userID).Find(&workspaces).Error
	return workspaces, err
}

func (p *postgresService) UpdateWorkspace(ctx context.Context, workspace *Workspace) error {
	return p.db.WithContext(ctx).Save(workspace).Error
}

func (p *postgresService) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return p.db.WithContext(ctx).Where("id = ?", workspaceID).Delete(&Workspace{}).Error
}

func (p *postgresService) AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error {
	workspaceUser := &WorkspaceUser{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
	}
	return p.db.WithContext(ctx).Create(workspaceUser).Error
}

func (p *postgresService) RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error {
	return p.db.WithContext(ctx).Where("user_id = ? AND workspace_id = ?", userID, workspaceID).Delete(&WorkspaceUser{}).Error
}

func (p *postgresService) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error) {
	var users []*WorkspaceUser
	err := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&users).Error
	return users, err
}

func (p *postgresService) GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error) {
	var workspaces []*Workspace
	err := p.db.WithContext(ctx).
		Joins("JOIN workspace_users ON workspaces.id = workspace_users.workspace_id").
		Where("workspace_users.user_id = ?", userID).
		Find(&workspaces).Error
	return workspaces, err
}

func (p *postgresService) CreateCredentials(ctx context.Context, cred *Credentials) error {
	return p.db.WithContext(ctx).Create(cred).Error
}

func (p *postgresService) GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error) {
	var cred Credentials
	err := p.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", credID, workspaceID).First(&cred).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cred, nil
}

func (p *postgresService) ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error) {
	var creds []*Credentials
	err := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&creds).Error
	return creds, err
}

func (p *postgresService) UpdateCredentials(ctx context.Context, cred *Credentials) error {
	return p.db.WithContext(ctx).Save(cred).Error
}

func (p *postgresService) DeleteCredentials(ctx context.Context, workspaceID, credID string) error {
	return p.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", credID, workspaceID).Delete(&Credentials{}).Error
}

func (p *postgresService) CreateExecution(ctx context.Context, execution *Execution) error {
	return p.db.WithContext(ctx).Create(execution).Error
}

func (p *postgresService) GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error) {
	var execution Execution
	err := p.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", executionID, workspaceID).First(&execution).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &execution, nil
}

func (p *postgresService) ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error) {
	var executions []*Execution
	err := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&executions).Error
	return executions, err
}

func (p *postgresService) UpdateExecution(ctx context.Context, execution *Execution) error {
	return p.db.WithContext(ctx).Save(execution).Error
}

func (p *postgresService) UpdateExecutionStatus(ctx context.Context, workspaceID, executionID, status string) error {
	return p.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", executionID, workspaceID).
		Update("status", status).Error
}

func (p *postgresService) GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error) {
	// Implement state retrieval logic
	return make(map[string]interface{}), nil
}

func (p *postgresService) SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error {
	// Implement state saving logic
	return nil
}

func (p *postgresService) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Additional errors
var (
	ErrConflict = fmt.Errorf("conflict")
)

func init() {
	// This will be replaced with proper error handling
}

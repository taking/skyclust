package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type postgresService struct {
	db *gorm.DB
}

// GetDB returns the underlying GORM database instance
func (p *postgresService) GetDB() *gorm.DB {
	return p.db
}

func NewPostgresService(config Config) Service {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.User, config.Password, config.Database, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Skip AutoMigrate for now to avoid constraint conflicts
	// The database schema is already created by init-db.sql
	log.Println("Skipping AutoMigrate - using existing schema")

	log.Println("Connected to PostgreSQL with GORM")
	return &postgresService{db: db}
}

// User management methods
func (p *postgresService) CreateUser(ctx context.Context, user *User) error {
	gormUser := user.ToGormUser()
	result := p.db.WithContext(ctx).Create(gormUser)
	if result.Error != nil {
		return result.Error
	}
	// Update the original user with the generated ID
	user.ID = gormUser.ID
	user.CreatedAt = gormUser.CreatedAt
	user.UpdatedAt = gormUser.UpdatedAt
	return nil
}

func (p *postgresService) GetUser(ctx context.Context, userID string) (*User, error) {
	var gormUser GormUser
	result := p.db.WithContext(ctx).Where("id = ?", userID).First(&gormUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return gormUser.ToUser(), nil
}

func (p *postgresService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var gormUser GormUser
	result := p.db.WithContext(ctx).Where("email = ?", email).First(&gormUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User not found, return nil without error
		}
		return nil, result.Error
	}
	return gormUser.ToUser(), nil
}

func (p *postgresService) UpdateUser(ctx context.Context, user *User) error {
	gormUser := user.ToGormUser()
	result := p.db.WithContext(ctx).Save(gormUser)
	return result.Error
}

func (p *postgresService) DeleteUser(ctx context.Context, userID string) error {
	result := p.db.WithContext(ctx).Delete(&GormUser{}, "id = ?", userID)
	return result.Error
}

// Token management methods
func (p *postgresService) StoreToken(ctx context.Context, userID, token string) error {
	gormToken := &GormToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	result := p.db.WithContext(ctx).Create(gormToken)
	return result.Error
}

func (p *postgresService) ValidateToken(ctx context.Context, token string) (string, error) {
	var gormToken GormToken
	result := p.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&gormToken)
	if result.Error != nil {
		return "", result.Error
	}
	return gormToken.UserID, nil
}

func (p *postgresService) DeleteToken(ctx context.Context, token string) error {
	result := p.db.WithContext(ctx).Delete(&GormToken{}, "token = ?", token)
	return result.Error
}

// Workspace management methods
func (p *postgresService) CreateWorkspace(ctx context.Context, workspace *Workspace) error {
	gormWorkspace := workspace.ToGormWorkspace()
	result := p.db.WithContext(ctx).Create(gormWorkspace)
	if result.Error != nil {
		return result.Error
	}
	// Update the original workspace with the generated ID
	workspace.ID = gormWorkspace.ID
	workspace.CreatedAt = gormWorkspace.CreatedAt
	workspace.UpdatedAt = gormWorkspace.UpdatedAt
	return nil
}

func (p *postgresService) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	var gormWorkspace GormWorkspace
	result := p.db.WithContext(ctx).Where("id = ?", workspaceID).First(&gormWorkspace)
	if result.Error != nil {
		return nil, result.Error
	}
	return gormWorkspace.ToWorkspace(), nil
}

func (p *postgresService) GetWorkspaceByID(ctx context.Context, workspaceID string) (*Workspace, error) {
	// This is the same as GetWorkspace, just for interface compatibility
	return p.GetWorkspace(ctx, workspaceID)
}

func (p *postgresService) ListWorkspaces(ctx context.Context) ([]*Workspace, error) {
	var gormWorkspaces []GormWorkspace
	result := p.db.WithContext(ctx).Find(&gormWorkspaces)
	if result.Error != nil {
		return nil, result.Error
	}

	workspaces := make([]*Workspace, len(gormWorkspaces))
	for i, gw := range gormWorkspaces {
		workspaces[i] = gw.ToWorkspace()
	}
	return workspaces, nil
}

func (p *postgresService) ListWorkspacesByUser(ctx context.Context, userID string) ([]*Workspace, error) {
	// This is the same as GetUserWorkspaces, just for interface compatibility
	return p.GetUserWorkspaces(ctx, userID)
}

func (p *postgresService) UpdateWorkspace(ctx context.Context, workspace *Workspace) error {
	gormWorkspace := workspace.ToGormWorkspace()
	result := p.db.WithContext(ctx).Save(gormWorkspace)
	return result.Error
}

func (p *postgresService) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	result := p.db.WithContext(ctx).Delete(&GormWorkspace{}, "id = ?", workspaceID)
	return result.Error
}

// User-Workspace relationship methods
func (p *postgresService) AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error {
	gormWorkspaceUser := &GormWorkspaceUser{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
	}
	result := p.db.WithContext(ctx).Create(gormWorkspaceUser)
	return result.Error
}

func (p *postgresService) RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error {
	result := p.db.WithContext(ctx).Delete(&GormWorkspaceUser{}, "user_id = ? AND workspace_id = ?", userID, workspaceID)
	return result.Error
}

func (p *postgresService) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error) {
	var gormWorkspaceUsers []GormWorkspaceUser
	result := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&gormWorkspaceUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	workspaceUsers := make([]*WorkspaceUser, len(gormWorkspaceUsers))
	for i, gwu := range gormWorkspaceUsers {
		workspaceUsers[i] = &WorkspaceUser{
			UserID:      gwu.UserID,
			WorkspaceID: gwu.WorkspaceID,
			Role:        gwu.Role,
			JoinedAt:    gwu.JoinedAt,
		}
	}
	return workspaceUsers, nil
}

func (p *postgresService) GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error) {
	var gormWorkspaces []GormWorkspace
	result := p.db.WithContext(ctx).
		Joins("JOIN workspace_users ON workspaces.id = workspace_users.workspace_id").
		Where("workspace_users.user_id = ?", userID).
		Find(&gormWorkspaces)
	if result.Error != nil {
		return nil, result.Error
	}

	workspaces := make([]*Workspace, len(gormWorkspaces))
	for i, gw := range gormWorkspaces {
		workspaces[i] = gw.ToWorkspace()
	}
	return workspaces, nil
}

// Credentials management methods
func (p *postgresService) CreateCredentials(ctx context.Context, cred *Credentials) error {
	gormCred := cred.ToGormCredentials()
	result := p.db.WithContext(ctx).Create(gormCred)
	if result.Error != nil {
		return result.Error
	}
	// Update the original credentials with the generated ID
	cred.ID = gormCred.ID
	cred.CreatedAt = gormCred.CreatedAt
	cred.UpdatedAt = gormCred.UpdatedAt
	return nil
}

func (p *postgresService) GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error) {
	var gormCred GormCredentials
	result := p.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", credID, workspaceID).First(&gormCred)
	if result.Error != nil {
		return nil, result.Error
	}
	return gormCred.ToCredentials(), nil
}

func (p *postgresService) ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error) {
	var gormCreds []GormCredentials
	result := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&gormCreds)
	if result.Error != nil {
		return nil, result.Error
	}

	credentials := make([]*Credentials, len(gormCreds))
	for i, gc := range gormCreds {
		credentials[i] = gc.ToCredentials()
	}
	return credentials, nil
}

func (p *postgresService) UpdateCredentials(ctx context.Context, cred *Credentials) error {
	gormCred := cred.ToGormCredentials()
	result := p.db.WithContext(ctx).Save(gormCred)
	return result.Error
}

func (p *postgresService) DeleteCredentials(ctx context.Context, workspaceID, credID string) error {
	result := p.db.WithContext(ctx).Delete(&GormCredentials{}, "id = ? AND workspace_id = ?", credID, workspaceID)
	return result.Error
}

// Execution management methods
func (p *postgresService) CreateExecution(ctx context.Context, execution *Execution) error {
	gormExecution := execution.ToGormExecution()
	result := p.db.WithContext(ctx).Create(gormExecution)
	if result.Error != nil {
		return result.Error
	}
	// Update the original execution with the generated ID
	execution.ID = gormExecution.ID
	execution.StartedAt = gormExecution.StartedAt
	return nil
}

func (p *postgresService) GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error) {
	var gormExecution GormExecution
	result := p.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", executionID, workspaceID).First(&gormExecution)
	if result.Error != nil {
		return nil, result.Error
	}
	return gormExecution.ToExecution(), nil
}

func (p *postgresService) ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error) {
	var gormExecutions []GormExecution
	result := p.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&gormExecutions)
	if result.Error != nil {
		return nil, result.Error
	}

	executions := make([]*Execution, len(gormExecutions))
	for i, ge := range gormExecutions {
		executions[i] = ge.ToExecution()
	}
	return executions, nil
}

func (p *postgresService) UpdateExecution(ctx context.Context, execution *Execution) error {
	gormExecution := execution.ToGormExecution()
	result := p.db.WithContext(ctx).Save(gormExecution)
	return result.Error
}

func (p *postgresService) UpdateExecutionStatus(ctx context.Context, workspaceID, executionID, status string) error {
	result := p.db.WithContext(ctx).Model(&GormExecution{}).
		Where("id = ? AND workspace_id = ?", executionID, workspaceID).
		Update("status", status)
	return result.Error
}

// State management methods
func (p *postgresService) SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error {
	// This would be implemented if we add a state table
	return fmt.Errorf("not implemented")
}

func (p *postgresService) GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error) {
	// This would be implemented if we add a state table
	return nil, fmt.Errorf("not implemented")
}

// Ping checks database connectivity
func (p *postgresService) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

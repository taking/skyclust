package database

import (
	"context"
	"fmt"
)

// Service defines the database service interface
type Service interface {
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

// NewService creates a new database service
// This is a placeholder - in real implementation, you would connect to actual database
func NewService(config Config) Service {
	// For now, return a mock implementation
	return &mockService{}
}

// mockService is a temporary implementation for development
type mockService struct {
	users      map[string]*User
	workspaces map[string]*Workspace
	tokens     map[string]string // token -> userID
}

func (m *mockService) CreateUser(ctx context.Context, user *User) error {
	if m.users == nil {
		m.users = make(map[string]*User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockService) GetUser(ctx context.Context, userID string) (*User, error) {
	if m.users == nil {
		return nil, ErrNotFound
	}
	user, exists := m.users[userID]
	if !exists {
		return nil, ErrNotFound
	}
	return user, nil
}

func (m *mockService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.users == nil {
		return nil, ErrNotFound
	}
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockService) UpdateUser(ctx context.Context, user *User) error {
	if m.users == nil {
		m.users = make(map[string]*User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockService) DeleteUser(ctx context.Context, userID string) error {
	if m.users == nil {
		return nil
	}
	delete(m.users, userID)
	return nil
}

func (m *mockService) StoreToken(ctx context.Context, userID, token string) error {
	if m.tokens == nil {
		m.tokens = make(map[string]string)
	}
	m.tokens[token] = userID
	return nil
}

func (m *mockService) ValidateToken(ctx context.Context, token string) (string, error) {
	if m.tokens == nil {
		return "", ErrNotFound
	}
	userID, exists := m.tokens[token]
	if !exists {
		return "", ErrNotFound
	}
	return userID, nil
}

func (m *mockService) DeleteToken(ctx context.Context, token string) error {
	if m.tokens == nil {
		return nil
	}
	delete(m.tokens, token)
	return nil
}

func (m *mockService) CreateWorkspace(ctx context.Context, workspace *Workspace) error {
	if m.workspaces == nil {
		m.workspaces = make(map[string]*Workspace)
	}
	m.workspaces[workspace.ID] = workspace
	return nil
}

func (m *mockService) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	if m.workspaces == nil {
		return nil, ErrNotFound
	}
	ws, exists := m.workspaces[workspaceID]
	if !exists {
		return nil, ErrNotFound
	}
	return ws, nil
}

func (m *mockService) GetWorkspaceByID(ctx context.Context, workspaceID string) (*Workspace, error) {
	return m.GetWorkspace(ctx, workspaceID)
}

func (m *mockService) ListWorkspacesByUser(ctx context.Context, userID string) ([]*Workspace, error) {
	if m.workspaces == nil {
		return []*Workspace{}, nil
	}

	var userWorkspaces []*Workspace
	for _, ws := range m.workspaces {
		if ws.OwnerID == userID {
			userWorkspaces = append(userWorkspaces, ws)
		}
	}
	return userWorkspaces, nil
}

func (m *mockService) UpdateWorkspace(ctx context.Context, workspace *Workspace) error {
	if m.workspaces == nil {
		m.workspaces = make(map[string]*Workspace)
	}
	m.workspaces[workspace.ID] = workspace
	return nil
}

func (m *mockService) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	if m.workspaces == nil {
		return nil
	}
	delete(m.workspaces, workspaceID)
	return nil
}

func (m *mockService) AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error {
	// Mock implementation
	return nil
}

func (m *mockService) RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error {
	// Mock implementation
	return nil
}

func (m *mockService) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error) {
	// Mock implementation
	return []*WorkspaceUser{}, nil
}

func (m *mockService) GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error) {
	return m.ListWorkspacesByUser(ctx, userID)
}

func (m *mockService) Ping(ctx context.Context) error {
	return nil
}

// Credentials management methods
func (m *mockService) CreateCredentials(ctx context.Context, cred *Credentials) error {
	// Mock implementation
	return nil
}

func (m *mockService) GetCredentials(ctx context.Context, workspaceID, credID string) (*Credentials, error) {
	// Mock implementation
	return nil, ErrNotFound
}

func (m *mockService) ListCredentials(ctx context.Context, workspaceID string) ([]*Credentials, error) {
	// Mock implementation
	return []*Credentials{}, nil
}

func (m *mockService) UpdateCredentials(ctx context.Context, cred *Credentials) error {
	// Mock implementation
	return nil
}

func (m *mockService) DeleteCredentials(ctx context.Context, workspaceID, credID string) error {
	// Mock implementation
	return nil
}

// Execution management methods
func (m *mockService) CreateExecution(ctx context.Context, execution *Execution) error {
	// Mock implementation
	return nil
}

func (m *mockService) GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error) {
	// Mock implementation
	return nil, ErrNotFound
}

func (m *mockService) ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error) {
	// Mock implementation
	return []*Execution{}, nil
}

func (m *mockService) UpdateExecution(ctx context.Context, execution *Execution) error {
	// Mock implementation
	return nil
}

func (m *mockService) UpdateExecutionStatus(ctx context.Context, workspaceID, executionID, status string) error {
	// Mock implementation
	return nil
}

// State management methods
func (m *mockService) GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error) {
	// Mock implementation
	return map[string]interface{}{}, nil
}

func (m *mockService) SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error {
	// Mock implementation
	return nil
}

// Errors
var (
	ErrNotFound = fmt.Errorf("not found")
	ErrConflict = fmt.Errorf("conflict")
)

func init() {
	// This will be replaced with proper error handling
}

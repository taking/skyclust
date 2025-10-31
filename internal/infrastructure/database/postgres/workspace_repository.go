package postgres

import (
	"context"
	"fmt"
	"skyclust/internal/domain"

	"gorm.io/gorm"
	"skyclust/pkg/logger"
)

// WorkspaceRepository implements the domain.WorkspaceRepository interface
type WorkspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a new WorkspaceRepository
func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	result := r.db.WithContext(ctx).Create(workspace)
	if result.Error != nil {
		logger.Errorf("Failed to create workspace in database: %v - workspace: %+v", result.Error, workspace)
		return fmt.Errorf("failed to create workspace: %w", result.Error)
	}

	logger.Info(fmt.Sprintf("Workspace created in database: %s (%s)", workspace.ID, workspace.Name))
	return nil
}

// GetByID retrieves a workspace by ID
func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	var workspace domain.Workspace
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&workspace)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by ID: %w", result.Error)
	}

	return &workspace, nil
}

// GetByOwnerID retrieves workspaces by owner ID
func (r *WorkspaceRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspaces by owner ID: %w", result.Error)
	}

	return workspaces, nil
}

// Update updates a workspace
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	result := r.db.WithContext(ctx).Save(workspace)
	if result.Error != nil {
		logger.Errorf("Failed to update workspace in database: %v - workspace: %+v", result.Error, workspace)
		return fmt.Errorf("failed to update workspace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("workspace not found")
	}

	logger.Info(fmt.Sprintf("Workspace updated in database: %s", workspace.ID))
	return nil
}

// Delete deletes a workspace
func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Workspace{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete workspace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("workspace not found")
	}

	logger.Info(fmt.Sprintf("Workspace deleted from database: %s", id))
	return nil
}

// List retrieves a list of workspaces with pagination
func (r *WorkspaceRepository) List(ctx context.Context, limit, offset int) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", result.Error)
	}

	return workspaces, nil
}

// GetUserWorkspaces retrieves workspaces for a user
func (r *WorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return r.GetByOwnerID(ctx, userID)
}

// AddUserToWorkspace adds a user to a workspace
func (r *WorkspaceRepository) AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error {
	// This would require a workspace_users table
	// For now, we'll use the database service directly
	return fmt.Errorf("AddUserToWorkspace not implemented - requires workspace_users table")
}

// RemoveUserFromWorkspace removes a user from a workspace
func (r *WorkspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error {
	// This would require a workspace_users table
	// For now, we'll use the database service directly
	return fmt.Errorf("RemoveUserFromWorkspace not implemented - requires workspace_users table")
}

// GetWorkspaceMembers retrieves all members of a workspace
func (r *WorkspaceRepository) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	var users []*domain.User

	// This is a simplified implementation
	// In a real implementation, you would have a workspace_members table
	// For now, we'll return an empty list
	result := r.db.WithContext(ctx).
		Table("users").
		Where("id IN (SELECT user_id FROM workspace_members WHERE workspace_id = ?)", workspaceID).
		Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", result.Error)
	}

	return users, nil
}

// GetWorkspacesWithUsers retrieves workspaces with their users in a single query (optimized)
func (r *WorkspaceRepository) GetWorkspacesWithUsers(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace

	// Use JOIN to get workspaces with user information in a single query
	result := r.db.WithContext(ctx).
		Table("workspaces").
		Select("workspaces.*, users.id as owner_id, users.username as owner_username, users.email as owner_email").
		Joins("LEFT JOIN users ON workspaces.owner_id = users.id").
		Where("workspaces.owner_id = ? OR workspaces.id IN (SELECT workspace_id FROM workspace_users WHERE user_id = ?)", userID, userID).
		Order("workspaces.created_at DESC").
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspaces with users: %w", result.Error)
	}

	return workspaces, nil
}

// GetWorkspaceWithMembers retrieves a workspace with its members in a single query (optimized)
func (r *WorkspaceRepository) GetWorkspaceWithMembers(ctx context.Context, workspaceID string) (*domain.Workspace, []*domain.User, error) {
	var workspace domain.Workspace
	var users []*domain.User

	// Get workspace
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get workspace members in a single query
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN workspace_users ON users.id = workspace_users.user_id").
		Where("workspace_users.workspace_id = ?", workspaceID).
		Find(&users).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get workspace members: %w", err)
	}

	return &workspace, users, nil
}

// GetUserWorkspacesOptimized retrieves user workspaces with optimized query
func (r *WorkspaceRepository) GetUserWorkspacesOptimized(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace

	// First, get workspaces owned by the user
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", userID).
		Order("created_at DESC").
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", result.Error)
	}

	return workspaces, nil
}

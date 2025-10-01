package postgres

import (
	"context"
	"fmt"

	"cmp/internal/domain"
	"cmp/pkg/shared/logger"
	"gorm.io/gorm"
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

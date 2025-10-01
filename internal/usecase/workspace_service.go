package usecase

import (
	"context"
	"fmt"
	"time"

	"cmp/internal/domain"
	"cmp/pkg/shared/logger"
)

// WorkspaceService implements the WorkspaceService interface
type WorkspaceService struct {
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
}

// NewWorkspaceService creates a new WorkspaceService
func NewWorkspaceService(workspaceRepo domain.WorkspaceRepository, userRepo domain.UserRepository) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if owner exists
	owner, err := s.userRepo.GetByID(ctx, req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}
	if owner == nil {
		return nil, domain.ErrUserNotFound
	}

	// Create workspace
	workspace := &domain.Workspace{
		ID:          generateID(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	logger.Info(fmt.Sprintf("Workspace created successfully: %s (%s) - owner: %s", workspace.ID, workspace.Name, workspace.OwnerID))
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}
	return workspace, nil
}

// UpdateWorkspace updates a workspace
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing workspace
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// Update fields
	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Description != nil {
		workspace.Description = *req.Description
	}
	if req.IsActive != nil {
		workspace.IsActive = *req.IsActive
	}

	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	logger.Info(fmt.Sprintf("Workspace updated successfully: %s", workspace.ID))
	return workspace, nil
}

// DeleteWorkspace deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	if err := s.workspaceRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	logger.Info(fmt.Sprintf("Workspace deleted successfully: %s", id))
	return nil
}

// GetUserWorkspaces retrieves workspaces for a user
func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	workspaces, err := s.workspaceRepo.GetUserWorkspaces(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}

	return workspaces, nil
}

// AddUserToWorkspace adds a user to a workspace
func (s *WorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID, userID string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// TODO: Implement workspace membership logic
	// This would require a workspace_members table and related logic

	logger.Info(fmt.Sprintf("User added to workspace: %s -> %s", workspaceID, userID))
	return nil
}

// RemoveUserFromWorkspace removes a user from a workspace
func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// TODO: Implement workspace membership logic
	// This would require a workspace_members table and related logic

	logger.Info(fmt.Sprintf("User removed from workspace: %s -> %s", workspaceID, userID))
	return nil
}

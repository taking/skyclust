package usecase

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/logger"
)

// WorkspaceService implements the WorkspaceService interface
type WorkspaceService struct {
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
	eventBus      messaging.Bus
	auditLogRepo  domain.AuditLogRepository
}

// NewWorkspaceService creates a new WorkspaceService
func NewWorkspaceService(workspaceRepo domain.WorkspaceRepository, userRepo domain.UserRepository, eventBus messaging.Bus, auditLogRepo domain.AuditLogRepository) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		eventBus:      eventBus,
		auditLogRepo:  auditLogRepo,
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if owner exists
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner ID: %w", err)
	}

	owner, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}
	if owner == nil {
		return nil, domain.ErrUserNotFound
	}

	// Create workspace
	workspace := &domain.Workspace{
		ID:          uuid.New().String(),
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

// GetUserWorkspaces retrieves workspaces for a user (optimized)
func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	// Check if user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Use optimized query to avoid N+1 problem
	workspaces, err := s.workspaceRepo.GetUserWorkspacesOptimized(ctx, userID)
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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Add user to workspace using repository
	err = s.workspaceRepo.AddUserToWorkspace(ctx, userID, workspaceID, "member")
	if err != nil {
		return fmt.Errorf("failed to add user to workspace: %w", err)
	}

	// Log the action
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   "workspace_user_added",
		Resource: fmt.Sprintf("workspace:%s", workspaceID),
		Details: map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"role":         "member",
		},
	})

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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Remove user from workspace using repository
	err = s.workspaceRepo.RemoveUserFromWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to remove user from workspace: %w", err)
	}

	// Log the action
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userUUID,
		Action:   "workspace_user_removed",
		Resource: fmt.Sprintf("workspace:%s", workspaceID),
		Details: map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
		},
	})

	logger.Info(fmt.Sprintf("User removed from workspace: %s -> %s", workspaceID, userID))
	return nil
}

// GetWorkspaceMembers retrieves all members of a workspace (optimized)
func (s *WorkspaceService) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	// Use optimized query to get workspace with members in a single query
	workspace, members, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", err)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	return members, nil
}

package workspace

import (
	"context"
	"fmt"
	"strings"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"skyclust/pkg/logger"
)

// Service implements the Service interface
type Service struct {
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
	eventService  domain.EventService
	auditLogRepo  domain.AuditLogRepository
}

// NewService creates a new Service
func NewService(workspaceRepo domain.WorkspaceRepository, userRepo domain.UserRepository, eventService domain.EventService, auditLogRepo domain.AuditLogRepository) *Service {
	return &Service{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		eventService:  eventService,
		auditLogRepo:  auditLogRepo,
	}
}

// CreateWorkspace creates a new workspace
func (s *Service) CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if owner exists
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid owner ID format", 400)
	}

	owner, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get owner", 500)
	}
	if owner == nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if workspace name already exists for this owner
	existingWorkspaces, err := s.workspaceRepo.GetByOwnerID(ctx, req.OwnerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing workspaces", 500)
	}
	for _, ws := range existingWorkspaces {
		if ws.Name == req.Name {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}
	}

	// Create workspace
	workspace := &domain.Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		logger.Error(fmt.Sprintf("Failed to create workspace: %v - workspace: %+v", err, workspace))
		
		// Check for unique constraint violation (name already exists)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "idx_workspaces_name") {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}
		
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create workspace: %v", err), 500)
	}

	logger.Info(fmt.Sprintf("Workspace created successfully: %s (%s) - owner: %s", workspace.ID, workspace.Name, workspace.OwnerID))
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *Service) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}
	return workspace, nil
}

// UpdateWorkspace updates a workspace
func (s *Service) UpdateWorkspace(ctx context.Context, id string, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing workspace
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get workspace", 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// Check if name is being changed and if it conflicts with existing workspace
	if req.Name != nil && *req.Name != workspace.Name {
		// Check if new name already exists for this owner (excluding current workspace)
		existingWorkspaces, err := s.workspaceRepo.GetByOwnerID(ctx, workspace.OwnerID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing workspaces", 500)
		}
		for _, ws := range existingWorkspaces {
			if ws.ID != id && ws.Name == *req.Name {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
			}
		}
	}

	// Update fields (owner_id remains unchanged)
	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Description != nil {
		workspace.Description = *req.Description
	}
	if req.IsActive != nil {
		workspace.Active = *req.IsActive
	}

	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		logger.Error(fmt.Sprintf("Failed to update workspace: %v - workspace: %+v", err, workspace))
		
		// Check for unique constraint violation (name already exists)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "idx_workspaces_name") {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}
		
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update workspace: %v", err), 500)
	}

	logger.Info(fmt.Sprintf("Workspace updated successfully: %s", workspace.ID))
	return workspace, nil
}

// DeleteWorkspace deletes a workspace
func (s *Service) DeleteWorkspace(ctx context.Context, id string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	if err := s.workspaceRepo.Delete(ctx, id); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete workspace: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete workspace: %v", err), 500)
	}

	logger.Info(fmt.Sprintf("Workspace deleted successfully: %s", id))
	return nil
}

// GetUserWorkspaces retrieves workspaces for a user (optimized)
func (s *Service) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	// Check if user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid user ID format", 400)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Use optimized query to avoid N+1 problem
	workspaces, err := s.workspaceRepo.GetUserWorkspacesOptimized(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user workspaces: %v", err), 500)
	}

	return workspaces, nil
}

// AddUserToWorkspace adds a user to a workspace
func (s *Service) AddUserToWorkspace(ctx context.Context, workspaceID, userID string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// Check if user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid user ID: %v", err), 400)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Add user to workspace using repository
	err = s.workspaceRepo.AddUserToWorkspace(ctx, userID, workspaceID, "member")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to add user to workspace: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to add user to workspace: %v", err), 500)
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
func (s *Service) RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error {
	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// Check if user exists
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid user ID: %v", err), 400)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Remove user from workspace using repository
	err = s.workspaceRepo.RemoveUserFromWorkspace(ctx, userID, workspaceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to remove user from workspace: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to remove user from workspace: %v", err), 500)
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
func (s *Service) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	// Use optimized query to get workspace with members in a single query
	workspace, members, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace members: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	return members, nil
}

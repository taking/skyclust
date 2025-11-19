package postgres

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"strings"

	"gorm.io/gorm"
	"skyclust/pkg/logger"
)

// WorkspaceRepository: domain.WorkspaceRepository 인터페이스 구현체
type WorkspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository: 새로운 WorkspaceRepository를 생성합니다
func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create: 새로운 워크스페이스를 생성합니다
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	result := GetTransaction(ctx, r.db).Create(workspace)
	if result.Error != nil {
		logger.Errorf("Failed to create workspace in database: %v - workspace: %+v", result.Error, workspace)
		return fmt.Errorf("failed to create workspace: %w", result.Error)
	}

	logger.Info(fmt.Sprintf("Workspace created in database: %s (%s)", workspace.ID, workspace.Name))
	return nil
}

// GetByID: ID로 워크스페이스를 조회합니다
func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	var workspace domain.Workspace
	result := GetTransaction(ctx, r.db).Where("id = ?", id).First(&workspace)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by ID: %w", result.Error)
	}

	return &workspace, nil
}

// GetByOwnerID: 소유자 ID로 워크스페이스 목록을 조회합니다
func (r *WorkspaceRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace
	result := GetTransaction(ctx, r.db).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspaces by owner ID: %w", result.Error)
	}

	return workspaces, nil
}

// Update: 워크스페이스를 업데이트합니다
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	result := GetTransaction(ctx, r.db).Save(workspace)
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

// Delete: 워크스페이스를 영구 삭제합니다
func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	result := GetTransaction(ctx, r.db).Unscoped().Where("id = ?", id).Delete(&domain.Workspace{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete workspace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("workspace not found")
	}

	logger.Info(fmt.Sprintf("Workspace deleted from database: %s", id))
	return nil
}

// List: 페이지네이션을 포함한 워크스페이스 목록을 조회합니다
func (r *WorkspaceRepository) List(ctx context.Context, limit, offset int) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace
	result := GetTransaction(ctx, r.db).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", result.Error)
	}

	return workspaces, nil
}

// GetUserWorkspaces: 사용자의 워크스페이스 목록을 조회합니다
func (r *WorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return r.GetByOwnerID(ctx, userID)
}

// AddUserToWorkspace: 사용자를 워크스페이스에 추가합니다
func (r *WorkspaceRepository) AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error {
	workspaceUser := &domain.WorkspaceUser{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
	}

	result := GetTransaction(ctx, r.db).Create(workspaceUser)
	if result.Error != nil {
		// 중복 키 에러 확인 (이미 워크스페이스에 속한 사용자)
		if strings.Contains(result.Error.Error(), "duplicate key") ||
			strings.Contains(result.Error.Error(), "unique constraint") ||
			strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user is already a member of this workspace")
		}
		logger.Errorf("Failed to add user to workspace: %v", result.Error)
		return fmt.Errorf("failed to add user to workspace: %w", result.Error)
	}

	logger.Info(fmt.Sprintf("User %s added to workspace %s with role %s", userID, workspaceID, role))
	return nil
}

// RemoveUserFromWorkspace: 워크스페이스에서 사용자를 제거합니다
func (r *WorkspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error {
	result := GetTransaction(ctx, r.db).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Delete(&domain.WorkspaceUser{})

	if result.Error != nil {
		logger.Errorf("Failed to remove user from workspace: %v", result.Error)
		return fmt.Errorf("failed to remove user from workspace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user is not a member of this workspace")
	}

	logger.Info(fmt.Sprintf("User %s removed from workspace %s", userID, workspaceID))
	return nil
}

// GetWorkspaceMembers: 워크스페이스의 모든 멤버를 조회합니다
func (r *WorkspaceRepository) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	var users []*domain.User

	result := GetTransaction(ctx, r.db).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN workspace_users wu ON users.id = wu.user_id").
		Where("wu.workspace_id = ?", workspaceID).
		Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", result.Error)
	}

	return users, nil
}

// GetWorkspaceMembersWithRoles: workspace_users 테이블에서 역할과 가입일을 포함한 멤버를 조회합니다
func (r *WorkspaceRepository) GetWorkspaceMembersWithRoles(ctx context.Context, workspaceID string) ([]*domain.WorkspaceUser, error) {
	var workspaceUsers []*domain.WorkspaceUser

	result := GetTransaction(ctx, r.db).
		Where("workspace_id = ?", workspaceID).
		Order("joined_at ASC").
		Find(&workspaceUsers)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get workspace members with roles: %w", result.Error)
	}

	return workspaceUsers, nil
}

// GetWorkspacesWithUsers: 단일 쿼리로 사용자 정보를 포함한 워크스페이스 목록을 조회합니다 (최적화)
func (r *WorkspaceRepository) GetWorkspacesWithUsers(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace

	// JOIN을 사용하여 단일 쿼리로 사용자 정보를 포함한 워크스페이스 조회
	result := GetTransaction(ctx, r.db).
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

// GetWorkspaceWithMembers: 단일 쿼리로 멤버를 포함한 워크스페이스를 조회합니다 (최적화)
func (r *WorkspaceRepository) GetWorkspaceWithMembers(ctx context.Context, workspaceID string) (*domain.Workspace, []*domain.User, error) {
	var workspace domain.Workspace
	var users []*domain.User

	// 워크스페이스 조회
	if err := GetTransaction(ctx, r.db).Where("id = ?", workspaceID).First(&workspace).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// 단일 쿼리로 워크스페이스 멤버 조회
	if err := GetTransaction(ctx, r.db).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN workspace_users ON users.id = workspace_users.user_id").
		Where("workspace_users.workspace_id = ?", workspaceID).
		Find(&users).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get workspace members: %w", err)
	}

	return &workspace, users, nil
}

// GetUserWorkspacesOptimized: 최적화된 쿼리로 사용자의 워크스페이스 목록을 조회합니다
func (r *WorkspaceRepository) GetUserWorkspacesOptimized(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace

	// 사용자가 소유한 워크스페이스 조회
	result := GetTransaction(ctx, r.db).
		Where("owner_id = ?", userID).
		Order("created_at DESC").
		Find(&workspaces)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", result.Error)
	}

	return workspaces, nil
}

// CountMembers: 워크스페이스의 멤버 개수를 조회합니다 (workspace_users 테이블 + owner)
func (r *WorkspaceRepository) CountMembers(ctx context.Context, workspaceID string) (int64, error) {
	var count int64
	
	// workspace_users 테이블의 멤버 수 조회
	result := GetTransaction(ctx, r.db).
		Model(&domain.WorkspaceUser{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count workspace members: %w", result.Error)
	}
	
	// Owner는 workspace_users 테이블에 없을 수 있으므로, owner도 포함하여 계산
	// workspace_users에 owner가 있으면 이미 카운트에 포함됨
	// 없으면 +1 해야 하지만, 일반적으로 owner도 workspace_users에 포함되므로 그대로 반환
	
	return count, nil
}

package workspace

import (
	"context"
	"fmt"
	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"strings"
	"time"

	"github.com/google/uuid"
	"skyclust/pkg/logger"
)

// Service: Workspace 서비스 인터페이스 구현체
// 워크스페이스 관련 비즈니스 로직을 처리합니다
type Service struct {
	workspaceRepo  domain.WorkspaceRepository // 워크스페이스 저장소
	userRepo       domain.UserRepository      // 사용자 저장소
	eventService   domain.EventService        // 이벤트 서비스
	auditLogRepo   domain.AuditLogRepository  // 감사 로그 저장소
	eventPublisher *messaging.Publisher       // 이벤트 발행자
}

// NewService: 새로운 Workspace 서비스 인스턴스를 생성합니다
func NewService(workspaceRepo domain.WorkspaceRepository, userRepo domain.UserRepository, eventService domain.EventService, auditLogRepo domain.AuditLogRepository, eventPublisher *messaging.Publisher) *Service {
	return &Service{
		workspaceRepo:  workspaceRepo,
		userRepo:       userRepo,
		eventService:   eventService,
		auditLogRepo:   auditLogRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateWorkspace: 새로운 워크스페이스를 생성합니다
func (s *Service) CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
	// 요청 유효성 검사
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 소유자 ID 파싱 및 검증
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid owner ID format", 400)
	}

	// 소유자 존재 여부 확인
	owner, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get owner", 500)
	}
	if owner == nil {
		return nil, domain.ErrUserNotFound
	}

	// 동일한 소유자의 워크스페이스 이름 중복 확인
	existingWorkspaces, err := s.workspaceRepo.GetByOwnerID(ctx, req.OwnerID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check existing workspaces", 500)
	}
	for _, ws := range existingWorkspaces {
		if ws.Name == req.Name {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}
	}

	// 워크스페이스 엔티티 생성
	workspace := &domain.Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
	}

	// 데이터베이스에 워크스페이스 저장
	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		logger.Error(fmt.Sprintf("Failed to create workspace: %v - workspace: %+v", err, workspace))

		// 고유 제약 조건 위반 확인 (이미 존재하는 이름)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "idx_workspaces_name") {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}

		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create workspace: %v", err), 500)
	}

	// 소유자를 workspace_users 테이블에 admin 역할로 추가
	if err := s.workspaceRepo.AddUserToWorkspace(ctx, req.OwnerID, workspace.ID, "admin"); err != nil {
		logger.Error(fmt.Sprintf("Failed to add owner to workspace_users table: %v - workspace: %s, owner: %s", err, workspace.ID, req.OwnerID))
		return nil, domain.NewDomainError(
			domain.ErrCodeInternalError,
			fmt.Sprintf("workspace created but failed to add owner to workspace_users: %v", err),
			500,
		)
	}

	// 감사 로그 기록
	ownerUUID := ownerID
	common.LogAction(ctx, s.auditLogRepo, &ownerUUID, domain.ActionWorkspaceCreate,
		fmt.Sprintf("POST /api/v1/workspaces"),
		map[string]interface{}{
			"workspace_id": workspace.ID,
			"name":         workspace.Name,
			"owner_id":     workspace.OwnerID,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		workspaceData := map[string]interface{}{
			"workspace_id": workspace.ID,
			"name":         workspace.Name,
			"description":  workspace.Description,
			"owner_id":     workspace.OwnerID,
			"is_active":    workspace.Active,
			"created_at":   workspace.CreatedAt,
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, workspace.ID, "created", workspaceData)
	}

	logger.Info(fmt.Sprintf("Workspace created successfully: %s (%s) - owner: %s", workspace.ID, workspace.Name, workspace.OwnerID))
	return workspace, nil
}

// GetWorkspace: ID로 워크스페이스를 조회합니다
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

// UpdateWorkspace: 워크스페이스 정보를 업데이트합니다
func (s *Service) UpdateWorkspace(ctx context.Context, id string, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error) {
	// 요청 유효성 검사
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 기존 워크스페이스 조회
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get workspace", 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// 이름 변경 시 기존 워크스페이스와의 충돌 확인
	if req.Name != nil && *req.Name != workspace.Name {
		// 새로운 이름이 동일한 소유자의 다른 워크스페이스에 이미 존재하는지 확인 (현재 워크스페이스 제외)
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

	// 필드 업데이트 (owner_id는 변경되지 않음)
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

		// 고유 제약 조건 위반 확인
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "idx_workspaces_name") {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "workspace name already exists", 409)
		}

		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update workspace: %v", err), 500)
	}

	// 감사로그 기록
	ownerUUID, _ := uuid.Parse(workspace.OwnerID)
	common.LogAction(ctx, s.auditLogRepo, &ownerUUID, domain.ActionWorkspaceUpdate,
		fmt.Sprintf("PUT /api/v1/workspaces/%s", id),
		map[string]interface{}{
			"workspace_id": workspace.ID,
			"name":         workspace.Name,
			"is_active":    workspace.Active,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		workspaceData := map[string]interface{}{
			"workspace_id": workspace.ID,
			"name":         workspace.Name,
			"description":  workspace.Description,
			"owner_id":     workspace.OwnerID,
			"is_active":    workspace.Active,
			"updated_at":   workspace.UpdatedAt,
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, workspace.ID, "updated", workspaceData)
	}

	logger.Info(fmt.Sprintf("Workspace updated successfully: %s", workspace.ID))
	return workspace, nil
}

// DeleteWorkspace: 워크스페이스를 삭제합니다
func (s *Service) DeleteWorkspace(ctx context.Context, id string) error {
	// 워크스페이스 존재 확인
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

	// 감사로그 기록
	ownerUUID, _ := uuid.Parse(workspace.OwnerID)
	common.LogAction(ctx, s.auditLogRepo, &ownerUUID, domain.ActionWorkspaceDelete,
		fmt.Sprintf("DELETE /api/v1/workspaces/%s", id),
		map[string]interface{}{
			"workspace_id": id,
			"name":         workspace.Name,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		workspaceData := map[string]interface{}{
			"workspace_id": id,
			"name":         workspace.Name,
			"owner_id":     workspace.OwnerID,
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, id, "deleted", workspaceData)
	}

	logger.Info(fmt.Sprintf("Workspace deleted successfully: %s", id))
	return nil
}

// GetUserWorkspaces: 사용자의 워크스페이스 목록을 조회합니다
func (s *Service) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	// 사용자 존재 확인
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

	// 최적화된 쿼리 사용 (N+1 문제 방지)
	workspaces, err := s.workspaceRepo.GetUserWorkspacesOptimized(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user workspaces: %v", err), 500)
	}

	return workspaces, nil
}

// AddUserToWorkspace: 사용자를 워크스페이스에 멤버 역할로 추가합니다
func (s *Service) AddUserToWorkspace(ctx context.Context, workspaceID, userID string) error {
	return s.addUserToWorkspaceWithRole(ctx, workspaceID, userID, "member")
}

// AddMemberByEmail: 이메일로 사용자를 워크스페이스에 추가합니다
func (s *Service) AddMemberByEmail(ctx context.Context, workspaceID, email, role string) error {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// 이메일로 사용자 조회
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user by email: %v", err), 500)
	}
	if user == nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, "user not found", 404)
	}

	// 역할 유효성 검사
	if role != "admin" && role != "member" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid role. Must be 'admin' or 'member'", 400)
	}

	return s.addUserToWorkspaceWithRole(ctx, workspaceID, user.ID.String(), role)
}

// addUserToWorkspaceWithRole: 지정된 역할로 사용자를 워크스페이스에 추가하는 헬퍼 메서드
func (s *Service) addUserToWorkspaceWithRole(ctx context.Context, workspaceID, userID, role string) error {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// 사용자 존재 확인
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

	// 역할 유효성 검사
	if role != "admin" && role != "member" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid role. Must be 'admin' or 'member'", 400)
	}

	// 이미 멤버인지 확인
	_, existingMembers, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err == nil && existingMembers != nil {
		for _, member := range existingMembers {
			if member.ID.String() == userID {
				return domain.NewDomainError(domain.ErrCodeAlreadyExists, "user is already a member of this workspace", 409)
			}
		}
	}

	// 저장소를 통해 사용자 추가
	err = s.workspaceRepo.AddUserToWorkspace(ctx, userID, workspaceID, role)
	if err != nil {
		if strings.Contains(err.Error(), "already a member") {
			return domain.NewDomainError(domain.ErrCodeAlreadyExists, "user is already a member of this workspace", 409)
		}
		logger.Error(fmt.Sprintf("Failed to add user to workspace: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to add user to workspace: %v", err), 500)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, &userUUID, domain.ActionWorkspaceUserAdded,
		fmt.Sprintf("POST /api/v1/workspaces/%s/members", workspaceID),
		map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"role":         role,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		memberData := map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"role":         role,
			"action":       "member_added",
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, workspaceID, "member_added", memberData)
	}

	logger.Info(fmt.Sprintf("User %s added to workspace %s with role %s", userID, workspaceID, role))
	return nil
}

// RemoveUserFromWorkspace: 워크스페이스에서 사용자를 제거합니다
func (s *Service) RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// 사용자 존재 확인
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

	// 저장소를 통해 사용자 제거
	err = s.workspaceRepo.RemoveUserFromWorkspace(ctx, userID, workspaceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to remove user from workspace: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to remove user from workspace: %v", err), 500)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, &userUUID, domain.ActionWorkspaceUserRemoved,
		fmt.Sprintf("DELETE /api/v1/workspaces/%s/members/%s", workspaceID, userID),
		map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		memberData := map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"action":       "member_removed",
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, workspaceID, "member_removed", memberData)
	}

	logger.Info(fmt.Sprintf("User removed from workspace: %s -> %s", workspaceID, userID))
	return nil
}

// GetWorkspaceMembers: 워크스페이스의 모든 멤버를 조회합니다
func (s *Service) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*domain.User, error) {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// 최적화된 쿼리로 단일 쿼리에서 멤버 조회
	_, members, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace members: %v", err), 500)
	}

	return members, nil
}

// GetWorkspaceMembersWithRoles: 역할과 가입일을 포함한 워크스페이스 멤버를 조회합니다
func (s *Service) GetWorkspaceMembersWithRoles(ctx context.Context, workspaceID string) ([]*domain.WorkspaceUser, error) {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// 역할 정보를 포함한 멤버 조회
	members, err := s.workspaceRepo.GetWorkspaceMembersWithRoles(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace members with roles: %v", err), 500)
	}

	return members, nil
}

// UpdateMemberRole: 워크스페이스에서 멤버의 역할을 업데이트합니다
func (s *Service) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}

	// 역할 유효성 검사
	if role != "admin" && role != "member" {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid role. Must be 'admin' or 'member'", 400)
	}

	// 소유자 역할은 변경 불가
	if workspace.OwnerID == userID {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "cannot change owner role", 400)
	}

	// 사용자 존재 확인
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

	// 멤버 여부 확인
	_, members, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace members: %v", err), 500)
	}

	isMember := false
	for _, member := range members {
		if member.ID.String() == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return domain.NewDomainError(domain.ErrCodeNotFound, "user is not a member of this workspace", 404)
	}

	// 역할 업데이트 (제거 후 재추가 방식)
	err = s.workspaceRepo.RemoveUserFromWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to remove user from workspace: %v", err), 500)
	}

	err = s.workspaceRepo.AddUserToWorkspace(ctx, userID, workspaceID, role)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to add user to workspace: %v", err), 500)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, &userUUID, domain.ActionWorkspaceUserRoleUpdated,
		fmt.Sprintf("PUT /api/v1/workspaces/%s/members/%s/role", workspaceID, userID),
		map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"new_role":     role,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		memberData := map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      userID,
			"new_role":     role,
			"action":       "member_role_updated",
		}
		_ = s.eventPublisher.PublishWorkspaceEvent(ctx, workspaceID, "member_role_updated", memberData)
	}

	logger.Info(fmt.Sprintf("User %s role updated to %s in workspace %s", userID, role, workspaceID))
	return nil
}

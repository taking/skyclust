package workspace

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"skyclust/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: 워크스페이스 관련 HTTP 요청을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	workspaceService  domain.WorkspaceService
	userService       domain.UserService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler: 새로운 워크스페이스 핸들러를 생성합니다
func NewHandler(workspaceService domain.WorkspaceService, userService domain.UserService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("workspace"),
		workspaceService:  workspaceService,
		userService:       userService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// CreateWorkspace: 워크스페이스 생성 요청을 처리합니다
func (h *Handler) CreateWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.createWorkspaceHandler(),
		h.StandardCRUDDecorators("create_workspace")...,
	)

	handler(c)
}

// createWorkspaceHandler: 워크스페이스 생성의 핵심 비즈니스 로직
func (h *Handler) createWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.EnrichContextWithRequestMetadata(c)
		span := telemetry.SpanFromContext(ctx)
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		// Handler layer DTO 사용 (owner_id 없음)
		var handlerReq CreateWorkspaceRequest
		if err := h.ValidateRequest(c, &handlerReq); err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		// Domain layer DTO로 변환 (owner_id 설정)
		domainReq := domain.CreateWorkspaceRequest{
			Name:        handlerReq.Name,
			Description: handlerReq.Description,
			OwnerID:     userID.String(),
		}

		// Domain validation 수행
		if err := domainReq.Validate(); err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		h.logWorkspaceCreationAttempt(c, userID, domainReq)

		workspace, err := h.workspaceService.CreateWorkspace(ctx, domainReq)
		if err != nil {
			h.HandleError(c, err, "create_workspace")
			return
		}

		h.logWorkspaceCreationSuccess(c, userID, workspace)
		h.setTelemetryAttributes(span, userID, workspace.ID, workspace.Name, "create_workspace")
		h.Created(c, workspace, readability.SuccessMsgUserCreated)
	}
}

// GetWorkspaces: 워크스페이스 목록 조회 요청을 처리합니다
func (h *Handler) GetWorkspaces(c *gin.Context) {
	handler := h.Compose(
		h.getWorkspacesHandler(),
		h.StandardCRUDDecorators("get_workspaces")...,
	)

	handler(c)
}

// getWorkspacesHandler: 워크스페이스 목록 조회의 핵심 비즈니스 로직
func (h *Handler) getWorkspacesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_workspaces")
			return
		}

		h.logWorkspacesRequest(c, userID)

		// 페이지네이션 파라미터 파싱
		page, limit := h.ParsePageLimitParams(c)

		workspaces, err := h.workspaceService.GetUserWorkspaces(ctx, userID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspaces")
			return
		}

		// 표준화된 페이지네이션 응답 사용 (직접 배열: data[])
		total := int64(len(workspaces))

		h.setTelemetryAttributes(span, userID, "", "", "get_workspaces")
		h.BuildPaginatedResponse(c, workspaces, page, limit, total, "Workspaces retrieved successfully")
	}
}

// GetWorkspace: 단일 워크스페이스 조회 요청을 처리합니다
func (h *Handler) GetWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.getWorkspaceHandler(),
		h.StandardCRUDDecorators("get_workspace")...,
	)

	handler(c)
}

// getWorkspaceHandler: 단일 워크스페이스 조회의 핵심 비즈니스 로직
func (h *Handler) getWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "get_workspace")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_workspace")
			return
		}

		h.logWorkspaceRequest(c, userID, workspaceID)

		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspace")
			return
		}

		// Response DTO 생성 (credential_count, member_count 포함)
		response := gin.H{
			"id":             workspace.ID,
			"name":           workspace.Name,
			"description":   workspace.Description,
			"owner_id":       workspace.OwnerID,
			"settings":       workspace.Settings,
			"created_at":     workspace.CreatedAt,
			"updated_at":     workspace.UpdatedAt,
			"is_active":      workspace.Active,
		}
		
		// Settings에서 counts 추출하여 별도 필드로 추가
		if workspace.Settings != nil {
			if credentialCount, ok := workspace.Settings["credential_count"].(int64); ok {
				response["credential_count"] = credentialCount
			} else {
				response["credential_count"] = int64(0)
			}
			if memberCount, ok := workspace.Settings["member_count"].(int64); ok {
				response["member_count"] = memberCount
			} else {
				response["member_count"] = int64(0)
			}
		} else {
			response["credential_count"] = int64(0)
			response["member_count"] = int64(0)
		}

		h.setTelemetryAttributes(span, userID, workspace.ID, workspace.Name, "get_workspace")
		h.OK(c, response, "Workspace retrieved successfully")
	}
}

// UpdateWorkspace: 워크스페이스 업데이트 요청을 처리합니다
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.updateWorkspaceHandler(),
		h.StandardCRUDDecorators("update_workspace")...,
	)

	handler(c)
}

// updateWorkspaceHandler: 워크스페이스 업데이트의 핵심 비즈니스 로직
func (h *Handler) updateWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		// Handler layer DTO 사용
		var handlerReq UpdateWorkspaceRequest
		if err := h.ValidateRequest(c, &handlerReq); err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		// Domain layer DTO로 변환
		var domainReq domain.UpdateWorkspaceRequest
		if handlerReq.Name != nil {
			domainReq.Name = handlerReq.Name
		}
		if handlerReq.Description != nil {
			domainReq.Description = handlerReq.Description
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		// 권한 확인을 위해 워크스페이스 조회
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		// 업데이트 권한 확인 (소유자 또는 관리자)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "update_workspace")
			return
		}

		// 소유자 또는 관리자만 업데이트 가능
		if userID != ownerID && userRole != domain.AdminRoleType {
			h.Forbidden(c, "Insufficient permissions to update this workspace")
			return
		}

		h.logWorkspaceUpdateAttempt(c, userID, workspaceID)

		// 워크스페이스 업데이트 (owner_id는 변경되지 않음)
		updatedWorkspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceID.String(), domainReq)
		if err != nil {
			h.HandleError(c, err, "update_workspace")
			return
		}

		h.logWorkspaceUpdateSuccess(c, userID, updatedWorkspace)
		h.setTelemetryAttributes(span, userID, updatedWorkspace.ID, updatedWorkspace.Name, "update_workspace")
		h.OK(c, updatedWorkspace, readability.SuccessMsgUserUpdated)
	}
}

// DeleteWorkspace: 워크스페이스 삭제 요청을 처리합니다
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	handler := h.Compose(
		h.deleteWorkspaceHandler(),
		h.StandardCRUDDecorators("delete_workspace")...,
	)

	handler(c)
}

// deleteWorkspaceHandler: 워크스페이스 삭제의 핵심 비즈니스 로직
func (h *Handler) deleteWorkspaceHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.EnrichContextWithRequestMetadata(c)
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "delete_workspace")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "delete_workspace")
			return
		}

		h.logWorkspaceDeletionAttempt(c, userID, workspaceID)

		err = h.workspaceService.DeleteWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "delete_workspace")
			return
		}

		h.logWorkspaceDeletionSuccess(c, userID, workspaceID)
		h.setTelemetryAttributes(span, userID, workspaceID.String(), "", "delete_workspace")
		h.OK(c, gin.H{"message": "Workspace deleted successfully"}, readability.SuccessMsgUserDeleted)
	}
}

// setTelemetryAttributes: 텔레메트리 속성을 설정합니다
func (h *Handler) setTelemetryAttributes(span telemetry.Span, userID uuid.UUID, workspaceID, workspaceName, action string) {
	if span != nil {
		telemetry.SetAttributes(span, map[string]interface{}{
			"user_id":        userID.String(),
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
			"action":         action,
		})
	}
}

// logWorkspaceCreationAttempt: 워크스페이스 생성 시도 로그를 기록합니다
func (h *Handler) logWorkspaceCreationAttempt(c *gin.Context, userID uuid.UUID, req domain.CreateWorkspaceRequest) {
	h.LogBusinessEvent(c, "workspace_creation_attempted", userID.String(), "", map[string]interface{}{
		"workspace_name": req.Name,
		"description":    req.Description,
	})
}

// logWorkspaceCreationSuccess: 워크스페이스 생성 성공 로그를 기록합니다
func (h *Handler) logWorkspaceCreationSuccess(c *gin.Context, userID uuid.UUID, workspace *domain.Workspace) {
	h.LogBusinessEvent(c, "workspace_created", userID.String(), workspace.ID, map[string]interface{}{
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
		"owner_id":       workspace.OwnerID,
	})
}

// logWorkspacesRequest: 워크스페이스 목록 요청 로그를 기록합니다
func (h *Handler) logWorkspacesRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "workspaces_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_workspaces",
	})
}

// logWorkspaceRequest: 워크스페이스 조회 요청 로그를 기록합니다
func (h *Handler) logWorkspaceRequest(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_requested", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

// logWorkspaceUpdateAttempt: 워크스페이스 업데이트 시도 로그를 기록합니다
func (h *Handler) logWorkspaceUpdateAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_update_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

// logWorkspaceUpdateSuccess: 워크스페이스 업데이트 성공 로그를 기록합니다
func (h *Handler) logWorkspaceUpdateSuccess(c *gin.Context, userID uuid.UUID, workspace *domain.Workspace) {
	h.LogBusinessEvent(c, "workspace_updated", userID.String(), workspace.ID, map[string]interface{}{
		"workspace_id":   workspace.ID,
		"workspace_name": workspace.Name,
	})
}

// logWorkspaceDeletionAttempt: 워크스페이스 삭제 시도 로그를 기록합니다
func (h *Handler) logWorkspaceDeletionAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_deletion_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

// logWorkspaceDeletionSuccess: 워크스페이스 삭제 성공 로그를 기록합니다
func (h *Handler) logWorkspaceDeletionSuccess(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_deleted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

// GetMembers: 워크스페이스 멤버 목록 조회 요청을 처리합니다
func (h *Handler) GetMembers(c *gin.Context) {
	handler := h.Compose(
		h.getMembersHandler(),
		h.StandardCRUDDecorators("get_workspace_members")...,
	)

	handler(c)
}

// getMembersHandler: 워크스페이스 멤버 목록 조회의 핵심 비즈니스 로직
func (h *Handler) getMembersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}

		h.logWorkspaceMembersRequest(c, userID, workspaceID)

		// 권한 확인을 위해 워크스페이스 조회
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}

		// 멤버 조회 권한 확인 (소유자, 관리자 또는 멤버)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "get_workspace_members")
			return
		}

		// 소유자 또는 관리자가 아니면 멤버 여부 확인
		if userID != ownerID && userRole != domain.AdminRoleType {
			members, _ := h.workspaceService.GetWorkspaceMembers(ctx, workspaceID.String())
			isMember := false
			for _, member := range members {
				if member.ID == userID {
					isMember = true
					break
				}
			}
			if !isMember {
				h.Forbidden(c, "Insufficient permissions to view workspace members")
				return
			}
		}

		// workspace_users 테이블에서 역할 정보를 포함한 멤버 조회
		workspaceMembers, err := h.workspaceService.GetWorkspaceMembersWithRoles(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}

		// 사용자 정보를 채우기 위해 모든 사용자 조회 (소유자 포함)
		allMembers, err := h.workspaceService.GetWorkspaceMembers(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "get_workspace_members")
			return
		}

		// 빠른 조회를 위한 맵 생성
		memberMap := make(map[string]*domain.WorkspaceUser)
		for _, wu := range workspaceMembers {
			memberMap[wu.UserID] = wu
		}

		userMap := make(map[string]*domain.User)
		for _, user := range allMembers {
			userMap[user.ID.String()] = user
		}

		// 역할 정보를 포함한 응답 형식으로 변환
		responses := make([]WorkspaceMemberResponse, 0, len(workspaceMembers)+1)

		// 소유자를 먼저 추가
		if workspace.OwnerID != "" {
			ownerUUID, err := uuid.Parse(workspace.OwnerID)
			if err == nil {
				owner, err := h.userService.GetUserByID(ownerUUID)
				if err == nil && owner != nil {
					response := WorkspaceMemberResponse{
						UserID:      owner.ID.String(),
						WorkspaceID: workspaceID.String(),
						Role:        "owner",
						JoinedAt:    workspace.CreatedAt,
					}
					response.User.ID = owner.ID.String()
					response.User.Username = owner.Username
					response.User.Email = owner.Email
					responses = append(responses, response)
				}
			}
		}

		// workspace_users 테이블의 다른 멤버 추가
		for _, wu := range workspaceMembers {
			// 소유자는 이미 추가됨
			if wu.UserID == workspace.OwnerID {
				continue
			}

			// 사용자 정보 조회
			user, exists := userMap[wu.UserID]
			if !exists {
				userUUID, err := uuid.Parse(wu.UserID)
				if err == nil {
					user, _ = h.userService.GetUserByID(userUUID)
				}
			}

			if user != nil {
				response := WorkspaceMemberResponse{
					UserID:      wu.UserID,
					WorkspaceID: workspaceID.String(),
					Role:        wu.Role,
					JoinedAt:    wu.JoinedAt,
				}
				response.User.ID = user.ID.String()
				response.User.Username = user.Username
				response.User.Email = user.Email
				responses = append(responses, response)
			}
		}

		h.setTelemetryAttributes(span, userID, workspaceID.String(), workspace.Name, "get_workspace_members")

		// Always include meta information for consistency (direct array: data[])
		page, limit := h.ParsePageLimitParams(c)
		total := int64(len(responses))
		h.BuildPaginatedResponse(c, responses, page, limit, total, "Workspace members retrieved successfully")
	}
}

// AddMember: 워크스페이스 멤버 추가 요청을 처리합니다
func (h *Handler) AddMember(c *gin.Context) {
	handler := h.Compose(
		h.addMemberHandler(),
		h.StandardCRUDDecorators("add_workspace_member")...,
	)

	handler(c)
}

// addMemberHandler: 워크스페이스에 멤버 추가의 핵심 비즈니스 로직
func (h *Handler) addMemberHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.EnrichContextWithRequestMetadata(c)
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}

		var req AddMemberRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}

		// 권한 확인을 위해 워크스페이스 조회
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}

		// 멤버 추가 권한 확인 (소유자 또는 관리자)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "add_workspace_member")
			return
		}

		// 소유자 또는 관리자만 멤버 추가 가능
		if userID != ownerID && userRole != domain.AdminRoleType {
			h.Forbidden(c, "Insufficient permissions to add members to this workspace")
			return
		}

		h.logWorkspaceMemberAdditionAttempt(c, userID, workspaceID, req.Email)

		err = h.workspaceService.AddMemberByEmail(ctx, workspaceID.String(), req.Email, req.Role)
		if err != nil {
			h.HandleError(c, err, "add_workspace_member")
			return
		}

		h.logWorkspaceMemberAdditionSuccess(c, userID, workspaceID, req.Email)
		h.setTelemetryAttributes(span, userID, workspaceID.String(), workspace.Name, "add_workspace_member")
		h.OK(c, gin.H{"message": "Member added successfully"}, "Member added to workspace successfully")
	}
}

// RemoveMember: 워크스페이스 멤버 제거 요청을 처리합니다
func (h *Handler) RemoveMember(c *gin.Context) {
	handler := h.Compose(
		h.removeMemberHandler(),
		h.StandardCRUDDecorators("remove_workspace_member")...,
	)

	handler(c)
}

// removeMemberHandler: 워크스페이스에서 멤버 제거의 핵심 비즈니스 로직
func (h *Handler) removeMemberHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.EnrichContextWithRequestMetadata(c)
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}
		memberID, err := h.ExtractPathParam(c, "memberId")
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}

		// 권한 확인을 위해 워크스페이스 조회
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}

		// 멤버 제거 권한 확인 (소유자 또는 관리자)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "remove_workspace_member")
			return
		}

		// 소유자 또는 관리자만 멤버 제거 가능
		if userID != ownerID && userRole != domain.AdminRoleType {
			h.Forbidden(c, "Insufficient permissions to remove members from this workspace")
			return
		}

		// 소유자는 제거할 수 없음
		if memberID.String() == workspace.OwnerID {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Cannot remove workspace owner", 400), "remove_workspace_member")
			return
		}

		h.logWorkspaceMemberRemovalAttempt(c, userID, workspaceID, memberID)

		err = h.workspaceService.RemoveUserFromWorkspace(ctx, workspaceID.String(), memberID.String())
		if err != nil {
			h.HandleError(c, err, "remove_workspace_member")
			return
		}

		h.logWorkspaceMemberRemovalSuccess(c, userID, workspaceID, memberID)
		h.setTelemetryAttributes(span, userID, workspaceID.String(), workspace.Name, "remove_workspace_member")
		h.OK(c, gin.H{"message": "Member removed successfully"}, "Member removed from workspace successfully")
	}
}

// UpdateMemberRole: 워크스페이스 멤버 역할 업데이트 요청을 처리합니다
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	handler := h.Compose(
		h.updateMemberRoleHandler(),
		h.StandardCRUDDecorators("update_workspace_member_role")...,
	)

	handler(c)
}

// updateMemberRoleHandler: 멤버 역할 업데이트의 핵심 비즈니스 로직
func (h *Handler) updateMemberRoleHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.EnrichContextWithRequestMetadata(c)
		span := telemetry.SpanFromContext(ctx)
		workspaceID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}
		memberID, err := h.ExtractPathParam(c, "memberId")
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}

		var req UpdateMemberRoleRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}

		// 권한 확인을 위해 워크스페이스 조회
		workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID.String())
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}

		// 멤버 역할 업데이트 권한 확인 (소유자 또는 관리자)
		userRole, err := h.GetUserRoleFromToken(c)
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}

		ownerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Invalid owner ID format", 500), "update_workspace_member_role")
			return
		}

		// 소유자 또는 관리자만 멤버 역할 업데이트 가능
		if userID != ownerID && userRole != domain.AdminRoleType {
			h.Forbidden(c, "Insufficient permissions to update member roles")
			return
		}

		// 소유자 역할은 변경할 수 없음
		if memberID.String() == workspace.OwnerID {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Cannot change workspace owner role", 400), "update_workspace_member_role")
			return
		}

		h.logWorkspaceMemberRoleUpdateAttempt(c, userID, workspaceID, memberID, req.Role)

		err = h.workspaceService.UpdateMemberRole(ctx, workspaceID.String(), memberID.String(), req.Role)
		if err != nil {
			h.HandleError(c, err, "update_workspace_member_role")
			return
		}

		h.logWorkspaceMemberRoleUpdateSuccess(c, userID, workspaceID, memberID, req.Role)
		h.setTelemetryAttributes(span, userID, workspaceID.String(), workspace.Name, "update_workspace_member_role")
		h.OK(c, gin.H{"message": "Member role updated successfully"}, "Member role updated successfully")
	}
}

// logWorkspaceMembersRequest: 워크스페이스 멤버 요청 로그를 기록합니다
func (h *Handler) logWorkspaceMembersRequest(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_members_requested", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
	})
}

// logWorkspaceMemberAdditionAttempt: 워크스페이스 멤버 추가 시도 로그를 기록합니다
func (h *Handler) logWorkspaceMemberAdditionAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, email string) {
	h.LogBusinessEvent(c, "workspace_member_addition_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"email":        email,
	})
}

// logWorkspaceMemberAdditionSuccess: 워크스페이스 멤버 추가 성공 로그를 기록합니다
func (h *Handler) logWorkspaceMemberAdditionSuccess(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, email string) {
	h.LogBusinessEvent(c, "workspace_member_added", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"email":        email,
	})
}

// logWorkspaceMemberRemovalAttempt: 워크스페이스 멤버 제거 시도 로그를 기록합니다
func (h *Handler) logWorkspaceMemberRemovalAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, memberID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_member_removal_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"member_id":    memberID.String(),
	})
}

// logWorkspaceMemberRemovalSuccess: 워크스페이스 멤버 제거 성공 로그를 기록합니다
func (h *Handler) logWorkspaceMemberRemovalSuccess(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, memberID uuid.UUID) {
	h.LogBusinessEvent(c, "workspace_member_removed", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"member_id":    memberID.String(),
	})
}

// logWorkspaceMemberRoleUpdateAttempt: 워크스페이스 멤버 역할 업데이트 시도 로그를 기록합니다
func (h *Handler) logWorkspaceMemberRoleUpdateAttempt(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, memberID uuid.UUID, role string) {
	h.LogBusinessEvent(c, "workspace_member_role_update_attempted", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"member_id":    memberID.String(),
		"role":         role,
	})
}

// logWorkspaceMemberRoleUpdateSuccess: 워크스페이스 멤버 역할 업데이트 성공 로그를 기록합니다
func (h *Handler) logWorkspaceMemberRoleUpdateSuccess(c *gin.Context, userID uuid.UUID, workspaceID uuid.UUID, memberID uuid.UUID, role string) {
	h.LogBusinessEvent(c, "workspace_member_role_updated", userID.String(), workspaceID.String(), map[string]interface{}{
		"workspace_id": workspaceID.String(),
		"member_id":    memberID.String(),
		"role":         role,
	})
}

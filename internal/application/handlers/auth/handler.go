package auth

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: 인증 관련 HTTP 요청을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	authService       domain.AuthService
	userService       domain.UserService
	logoutService     domain.LogoutService
	rbacService       domain.RBACService
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandlerWithUserService: 초기화 확인을 위한 사용자 서비스를 포함한 인증 핸들러를 생성합니다
func NewHandlerWithUserService(authService domain.AuthService, userService domain.UserService, rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("auth"),
		authService:       authService,
		userService:       userService,
		rbacService:       rbacService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// NewHandler: 새로운 인증 핸들러를 생성합니다
func NewHandler(authService domain.AuthService, userService domain.UserService, rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("auth"),
		authService:       authService,
		userService:       userService,
		rbacService:       rbacService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// NewHandlerWithLogout: 로그아웃 서비스를 포함한 인증 핸들러를 생성합니다
func NewHandlerWithLogout(authService domain.AuthService, userService domain.UserService, logoutService domain.LogoutService, rbacService domain.RBACService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("auth"),
		authService:       authService,
		userService:       userService,
		logoutService:     logoutService,
		rbacService:       rbacService,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// Register: 사용자 등록 요청을 처리합니다 (데코레이터 패턴 사용)
func (h *Handler) Register(c *gin.Context) {
	handler := h.Compose(
		h.registerHandler(domain.CreateUserRequest{}),
		h.PublicDecorators("register")...,
	)

	handler(c)
}

// registerHandler: 사용자 등록의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) registerHandler(req domain.CreateUserRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.ExtractValidatedRequest(c, &req); err != nil {
			h.HandleError(c, err, "register")
			return
		}

		h.logUserRegistrationAttempt(c, req)

		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "register")
			return
		}

		// 시스템이 이미 초기화되었는지 확인 (보안 강화)
		// initialized가 true면 Register를 막아야 함
		if h.userService != nil {
			userCount, err := h.userService.GetUserCount()
			if err != nil {
				h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to check system initialization status", 500), "register")
				return
			}
			if userCount > 0 {
				// 이미 초기화된 경우 Register 거부
				h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "registration is only allowed during initial setup. system is already initialized", 403), "register")
				return
			}
		}

		// Register는 context를 받지 않지만, 감사로그를 위해 context를 enrich
		ctx := h.EnrichContextWithRequestMetadata(c)
		_ = ctx // Register는 context를 받지 않으므로 사용하지 않음 (향후 개선 시 사용)

		user, token, err := h.authService.Register(req)
		if err != nil {
			h.HandleError(c, err, "register")
			return
		}

		h.logUserRegistrationSuccess(c, user)
		h.Created(c, gin.H{
			"token": token,
			"user":  user,
		}, readability.SuccessMsgUserCreated)
	}
}

// Login: 사용자 로그인 요청을 처리합니다 (데코레이터 패턴 사용)
func (h *Handler) Login(c *gin.Context) {
	handler := h.Compose(
		h.loginHandler(struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}{}),
		h.PublicDecorators("login")...,
	)

	handler(c)
}

// loginHandler: 사용자 로그인의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) loginHandler(req struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := h.ExtractValidatedRequest(c, &req); err != nil {
			h.HandleError(c, err, "login")
			return
		}

		h.logUserLoginAttempt(c, req.Email)

		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "login")
			return
		}

		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		user, token, err := h.authService.LoginWithContext(req.Email, req.Password, clientIP, userAgent)
		if err != nil {
			h.HandleError(c, err, "login")
			return
		}

		h.logUserLoginSuccess(c, user)
		h.OK(c, gin.H{
			"token": token,
			"user":  user,
		}, readability.SuccessMsgLoginSuccess)
	}
}

// Logout: 사용자 로그아웃 요청을 처리합니다 (DELETE /auth/sessions/me)
// RESTful: 현재 세션 삭제
func (h *Handler) Logout(c *gin.Context) {
	handler := h.Compose(
		h.logoutHandler(),
		h.StandardCRUDDecorators("delete_session")...,
	)

	handler(c)
}

// GetSession: 현재 세션 정보를 조회합니다 (GET /auth/sessions/me)
// RESTful: 현재 세션 조회
func (h *Handler) GetSession(c *gin.Context) {
	handler := h.Compose(
		h.getSessionHandler(),
		h.StandardCRUDDecorators("get_session")...,
	)

	handler(c)
}

// logoutHandler: 사용자 로그아웃의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) logoutHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "delete_session")
			return
		}

		token, err := h.GetBearerTokenFromHeader(c)
		if err != nil {
			h.HandleError(c, err, "delete_session")
			return
		}

		h.logUserLogoutAttempt(c, userID)

		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "delete_session")
			return
		}

		// Logout은 context를 받지 않지만, 감사로그를 위해 context를 enrich
		ctx := h.EnrichContextWithRequestMetadata(c)
		_ = ctx // Logout은 context를 받지 않으므로 사용하지 않음 (향후 개선 시 사용)

		err = h.authService.Logout(userID, token)
		if err != nil {
			h.HandleError(c, err, "delete_session")
			return
		}

		h.logUserLogoutSuccess(c, userID)
		h.OK(c, gin.H{
			"message": "Session terminated successfully",
		}, readability.SuccessMsgLogoutSuccess)
	}
}

// getSessionHandler: 현재 세션 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getSessionHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_session")
			return
		}

		token, err := h.GetBearerTokenFromHeader(c)
		if err != nil {
			h.HandleError(c, err, "get_session")
			return
		}

		// Validate token
		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "get_session")
			return
		}

		user, err := h.authService.ValidateToken(token)
		if err != nil {
			h.HandleError(c, err, "get_session")
			return
		}

		// Get user roles
		var roles []string
		if h.rbacService != nil {
			userRoles, err := h.rbacService.GetUserRoles(userID)
			if err == nil {
				for _, role := range userRoles {
					roles = append(roles, string(role))
				}
			}
		}

		h.OK(c, gin.H{
			"user_id":  user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
			"roles":    roles,
			"active":   user.IsActive(),
		}, "Session retrieved successfully")
	}
}

// Me: 현재 사용자 정보를 반환합니다 (데코레이터 패턴 사용)
func (h *Handler) Me(c *gin.Context) {
	handler := h.Compose(
		h.meHandler(),
		h.StandardCRUDDecorators("me")...,
	)

	handler(c)
}

// meHandler: 현재 사용자 정보 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) meHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "me")
			return
		}

		h.logUserMeRequest(c, userID)

		user, err := h.userService.GetUserByID(userID)
		if err != nil {
			h.HandleError(c, err, "me")
			return
		}

		// Get role from token (primary role for UI control)
		// Note: Only role name is exposed, not detailed permissions for security
		role, err := h.GetUserRoleFromToken(c)
		if err != nil {
			// If role extraction fails, continue without role (non-critical)
			role = ""
		}

		// Build response with role information
		// Note: Using UserResponse DTO to expose only necessary information
		// Role name is exposed for UI control, but detailed permissions are not
		userResponse := &UserResponse{
			ID:           user.ID.String(),
			Username:     user.Username,
			Email:        user.Email,
			IsActive:     user.Active,
			Role:         string(role), // Primary role only (not detailed permissions)
			OIDCProvider: user.OIDCProvider,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}

		h.OK(c, userResponse, "User retrieved successfully")
	}
}

// GetUsers: 모든 사용자 목록을 조회합니다 (관리자만 가능)
func (h *Handler) GetUsers(c *gin.Context) {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_users")
		return
	}

	// Check if user has permission to list users (admin role only)
	if userRole != domain.AdminRoleType {
		h.Forbidden(c, "Insufficient permissions to list users")
		return
	}

	// Parse pagination parameters (page/limit based)
	page, limit := h.ParsePageLimitParams(c)

	// Get users from service
	users, total, err := h.userService.GetUsersWithFilters(domain.UserFilters{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		h.HandleError(c, err, "get_users")
		return
	}

	// Build user responses with role information
	// Note: Only primary role is exposed for UI control (not detailed permissions)
	userResponses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		// Get primary role (first role) for each user
		var primaryRole domain.Role
		if h.rbacService != nil {
			userRoles, err := h.rbacService.GetUserRoles(user.ID)
			if err == nil && len(userRoles) > 0 {
				primaryRole = userRoles[0] // Use first role as primary
			}
		}

		userResponse := &UserResponse{
			ID:           user.ID.String(),
			Username:     user.Username,
			Email:        user.Email,
			IsActive:     user.Active,
			Role:         string(primaryRole), // Primary role only
			OIDCProvider: user.OIDCProvider,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	// Use standardized pagination metadata
	paginationMeta := h.CalculatePaginationMeta(total, page, limit)

	h.OK(c, gin.H{
		"users":      userResponses,
		"pagination": paginationMeta,
	}, "Users retrieved successfully")
}

// GetUser: ID로 특정 사용자를 조회합니다
func (h *Handler) GetUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "get_user")
		return
	}

	// Get primary role (first role) for the user
	// Note: Only primary role is exposed for UI control (not detailed permissions)
	var primaryRole domain.Role
	if h.rbacService != nil {
		userRoles, err := h.rbacService.GetUserRoles(userID)
		if err == nil && len(userRoles) > 0 {
			primaryRole = userRoles[0] // Use first role as primary
		}
	}

	// Build response with role information
	userResponse := &UserResponse{
		ID:           user.ID.String(),
		Username:     user.Username,
		Email:        user.Email,
		IsActive:     user.Active,
		Role:         string(primaryRole), // Primary role only
		OIDCProvider: user.OIDCProvider,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	h.OK(c, userResponse, "User retrieved successfully")
}

// UpdateUser: 사용자 정보를 업데이트합니다
func (h *Handler) UpdateUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	currentUserRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	// Check if user can update this user (own profile or admin)
	if currentUserID != userID && currentUserRole != domain.AdminRoleType {
		h.Forbidden(c, "Insufficient permissions to update this user")
		return
	}

	var req domain.UpdateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	// Update user fields
	if req.Username != nil && *req.Username != "" {
		user.Username = *req.Username
	}
	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}
	if req.Password != nil && *req.Password != "" {
		// Hash new password
		hashedPassword, err := h.userService.HashPassword(*req.Password)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500), "update_user")
			return
		}
		user.PasswordHash = hashedPassword
	}
	if req.IsActive != nil {
		user.Active = *req.IsActive
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	h.OK(c, updatedUser, "User updated successfully")
}

// DeleteUser: 사용자를 삭제합니다
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := h.ParseUUID(c, "id")
	if err != nil {
		return
	}

	// Get current user info from token for authorization
	currentUserID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	currentUserRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	// Check if user can delete this user (admin only, cannot delete self)
	if currentUserRole != domain.AdminRoleType {
		h.Forbidden(c, "Only administrators can delete users")
		return
	}

	// Prevent self-deletion
	if currentUserID == userID {
		h.Forbidden(c, "Cannot delete your own account")
		return
	}

	// Check if user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	// Delete user
	err = h.userService.DeleteUserByID(userID)
	if err != nil {
		h.HandleError(c, err, "delete_user")
		return
	}

	h.OK(c, gin.H{"message": "User deleted successfully"}, readability.SuccessMsgUserDeleted)
}

// 로깅 헬퍼 메서드들

// logUserRegistrationAttempt: 사용자 등록 시도 로그를 기록합니다
func (h *Handler) logUserRegistrationAttempt(c *gin.Context, req domain.CreateUserRequest) {
	h.LogBusinessEvent(c, "user_registration_attempted", "", "", map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	})
}

// logUserRegistrationSuccess: 사용자 등록 성공 로그를 기록합니다
func (h *Handler) logUserRegistrationSuccess(c *gin.Context, user *domain.User) {
	h.LogBusinessEvent(c, "user_registered", user.ID.String(), "", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})
}

// logUserLoginAttempt: 사용자 로그인 시도 로그를 기록합니다
func (h *Handler) logUserLoginAttempt(c *gin.Context, email string) {
	h.LogBusinessEvent(c, "user_login_attempted", "", "", map[string]interface{}{
		"email": email,
	})
}

// logUserLoginSuccess: 사용자 로그인 성공 로그를 기록합니다
func (h *Handler) logUserLoginSuccess(c *gin.Context, user *domain.User) {
	h.LogBusinessEvent(c, "user_logged_in", user.ID.String(), "", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})
}

// logUserLogoutAttempt: 사용자 로그아웃 시도 로그를 기록합니다
func (h *Handler) logUserLogoutAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "user_logout_attempted", userID.String(), "", map[string]interface{}{
		"user_id": userID.String(),
	})
}

// logUserLogoutSuccess: 사용자 로그아웃 성공 로그를 기록합니다
func (h *Handler) logUserLogoutSuccess(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "user_logged_out", userID.String(), "", map[string]interface{}{
		"user_id": userID.String(),
	})
}

// logUserMeRequest: 사용자 정보 조회 요청 로그를 기록합니다
func (h *Handler) logUserMeRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "user_me_requested", userID.String(), "", map[string]interface{}{
		"user_id": userID.String(),
	})
}

package auth

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
		if err := h.ValidateRequest(c, &req); err != nil {
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

		user, accessToken, refreshToken, err := h.authService.Register(req)
		if err != nil {
			h.HandleError(c, err, "register")
			return
		}

		h.logUserRegistrationSuccess(c, user)
		h.Created(c, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    int(h.getJWTExpiry().Seconds()),
			"token_type":    "Bearer",
			"user":          user,
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
		if err := h.ValidateRequest(c, &req); err != nil {
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

		user, accessToken, refreshToken, err := h.authService.LoginWithContext(req.Email, req.Password, clientIP, userAgent)
		if err != nil {
			h.HandleError(c, err, "login")
			return
		}

		h.logUserLoginSuccess(c, user)
		h.OK(c, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    int(h.getJWTExpiry().Seconds()),
			"token_type":    "Bearer",
			"user":          user,
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

// RefreshToken: Refresh Token을 사용하여 새로운 Access Token과 Refresh Token을 발급합니다
func (h *Handler) RefreshToken(c *gin.Context) {
	handler := h.Compose(
		h.refreshTokenHandler(struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}{}),
		h.PublicDecorators("refresh_token")...,
	)

	handler(c)
}

// refreshTokenHandler: Refresh Token 핵심 비즈니스 로직을 처리합니다
func (h *Handler) refreshTokenHandler(req struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "refresh_token")
			return
		}

		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "refresh_token")
			return
		}

		accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
		if err != nil {
			h.HandleError(c, err, "refresh_token")
			return
		}

		h.OK(c, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    int(h.getJWTExpiry().Seconds()),
			"token_type":    "Bearer",
		}, "Token refreshed successfully")
	}
}

// RevokeToken: Refresh Token을 무효화합니다
func (h *Handler) RevokeToken(c *gin.Context) {
	handler := h.Compose(
		h.revokeTokenHandler(struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}{}),
		h.StandardCRUDDecorators("revoke_token")...,
	)

	handler(c)
}

// revokeTokenHandler: Refresh Token 무효화 핵심 비즈니스 로직을 처리합니다
func (h *Handler) revokeTokenHandler(req struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "revoke_token")
			return
		}

		if h.authService == nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "Authentication service is not available", 503), "revoke_token")
			return
		}

		err := h.authService.RevokeRefreshToken(req.RefreshToken)
		if err != nil {
			h.HandleError(c, err, "revoke_token")
			return
		}

		h.OK(c, gin.H{
			"message": "Refresh token revoked successfully",
		}, "Token revoked successfully")
	}
}

// getJWTExpiry: JWT 만료 시간을 가져옵니다 (helper method)
func (h *Handler) getJWTExpiry() time.Duration {
	// Try to get expiry from auth service if it implements the interface
	if expiryGetter, ok := h.authService.(interface{ GetJWTExpiry() time.Duration }); ok {
		return expiryGetter.GetJWTExpiry()
	}
	// Default to 15 minutes if not available
	return 15 * time.Minute
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
	// Batch fetch user roles to avoid N+1 query problem
	userIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	var userRolesMap map[uuid.UUID][]domain.Role
	if h.rbacService != nil {
		var err error
		userRolesMap, err = h.rbacService.GetUsersRoles(userIDs)
		if err != nil {
			h.LogWarn(c, "Failed to get user roles, continuing without roles", zap.Error(err))
			userRolesMap = make(map[uuid.UUID][]domain.Role)
		}
	} else {
		userRolesMap = make(map[uuid.UUID][]domain.Role)
	}

	userResponses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		// Get primary role (first role) for each user
		var primaryRole domain.Role
		if userRoles, exists := userRolesMap[user.ID]; exists && len(userRoles) > 0 {
			primaryRole = userRoles[0] // Use first role as primary
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

	// Use standardized paginated response (direct array: data[])
	h.BuildPaginatedResponse(c, userResponses, page, limit, total, "Users retrieved successfully")
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

	// Handler layer DTO 사용
	var handlerReq UpdateUserRequest
	if err := h.ValidateRequest(c, &handlerReq); err != nil {
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
	if handlerReq.Username != nil && *handlerReq.Username != "" {
		user.Username = *handlerReq.Username
	}
	if handlerReq.Email != nil && *handlerReq.Email != "" {
		user.Email = *handlerReq.Email
	}
	if handlerReq.Password != nil && *handlerReq.Password != "" {
		// Hash new password
		hashedPassword, err := h.userService.HashPassword(*handlerReq.Password)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500), "update_user")
			return
		}
		user.PasswordHash = hashedPassword
	}
	if handlerReq.IsActive != nil {
		user.Active = *handlerReq.IsActive
	}

	// Update user
	updatedUser, err := h.userService.UpdateUserDirect(user)
	if err != nil {
		h.HandleError(c, err, "update_user")
		return
	}

	h.OK(c, updatedUser, "User updated successfully")
}

// CreateUser: 관리자가 사용자를 생성합니다 (초기 설정 체크 없음)
func (h *Handler) CreateUser(c *gin.Context) {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_user")
		return
	}

	// Check if user has permission to create users (admin role only)
	if userRole != domain.AdminRoleType {
		h.Forbidden(c, "Only administrators can create users")
		return
	}

	// Parse request
	var req domain.CreateUserRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_user")
		return
	}

	// Enrich context with request metadata
	ctx := h.EnrichContextWithRequestMetadata(c)

	// Create user using userService (no initial setup check)
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		h.HandleError(c, err, "create_user")
		return
	}

	// Assign default role (user role for admin-created users)
	if h.rbacService != nil {
		if err := h.rbacService.AssignRole(user.ID, domain.UserRoleType); err != nil {
			h.LogWarn(c, "Failed to assign role to created user", zap.Error(err))
			// Continue even if role assignment fails
		}
	}

	// Get primary role for response
	var primaryRole domain.Role = domain.UserRoleType
	if h.rbacService != nil {
		userRoles, err := h.rbacService.GetUserRoles(user.ID)
		if err == nil && len(userRoles) > 0 {
			primaryRole = userRoles[0]
		}
	}

	// Build response
	userResponse := &UserResponse{
		ID:           user.ID.String(),
		Username:     user.Username,
		Email:        user.Email,
		IsActive:     user.Active,
		Role:         string(primaryRole),
		OIDCProvider: user.OIDCProvider,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	h.LogBusinessEvent(c, "user_created_by_admin", user.ID.String(), "", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	h.Created(c, userResponse, "User created successfully")
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

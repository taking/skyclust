package auth

import (
	"context"
	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/security"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey는 context.WithValue에서 사용할 커스텀 키 타입입니다
type contextKey string

const (
	contextKeyClientIP  contextKey = "client_ip"
	contextKeyUserAgent contextKey = "user_agent"
)

// Service: 인증 비즈니스 로직 구현체
type Service struct {
	userRepo          domain.UserRepository
	auditLogRepo      domain.AuditLogRepository
	rbacService       domain.RBACService
	hasher            security.PasswordHasher
	blacklist         *cache.TokenBlacklist
	refreshTokenStore *cache.RefreshTokenStore
	jwtSecret         string
	jwtExpiry         time.Duration
	refreshTokenExpiry time.Duration
	enableRotation    bool
}

// NewService: 새로운 인증 서비스를 생성합니다
func NewService(
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
	rbacService domain.RBACService,
	hasher security.PasswordHasher,
	blacklist *cache.TokenBlacklist,
	refreshTokenStore *cache.RefreshTokenStore,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	enableRotation bool,
) domain.AuthService {
	return &Service{
		userRepo:          userRepo,
		auditLogRepo:      auditLogRepo,
		rbacService:       rbacService,
		hasher:            hasher,
		blacklist:         blacklist,
		refreshTokenStore: refreshTokenStore,
		jwtSecret:         jwtSecret,
		jwtExpiry:         jwtExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		enableRotation:    enableRotation,
	}
}

// Register: 새로운 사용자 계정을 생성합니다
func (s *Service) Register(req domain.CreateUserRequest) (*domain.User, string, string, error) {
	// Check if email already exists (email is unique)
	if existing, _ := s.userRepo.GetByEmail(req.Email); existing != nil {
		return nil, "", "", domain.ErrUserAlreadyExists
	}

	// Check if this is the first user (make them admin) - MUST check BEFORE creating user
	userCount, err := s.userRepo.Count()
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to check user count", 500)
	}

	// Assign role based on user count
	var defaultRole domain.Role
	if userCount == 0 {
		// First user becomes admin
		defaultRole = domain.AdminRoleType
	} else {
		// Subsequent users get user role
		defaultRole = domain.UserRoleType
	}

	// Hash password
	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500)
	}

	// Create user
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Active:       true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to create user", 500)
	}

	if err := s.rbacService.AssignRole(user.ID, defaultRole); err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to assign role", 500)
	}

	// Get user roles for JWT token
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate JWT token with primary role (first role)
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	accessToken, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate access token", 500)
	}

	// Generate refresh token
	refreshToken, err := s.generateAndStoreRefreshToken(context.Background(), user.ID, "", "")
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate refresh token", 500)
	}

	// Log registration
	ctx := context.Background()
	common.LogAction(ctx, s.auditLogRepo, &user.ID, domain.ActionUserRegister,
		"POST /api/v1/auth/register",
		map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		},
	)

	return user, accessToken, refreshToken, nil
}

// Login: 사용자를 인증하고 JWT 토큰을 반환합니다
func (s *Service) Login(email, password string) (*domain.User, string, string, error) {
	// Get user by email (email is unique)
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	// Validate password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	// Get user roles for JWT token
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate JWT token with primary role (first role)
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	accessToken, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate access token", 500)
	}

	// Generate refresh token
	refreshToken, err := s.generateAndStoreRefreshToken(context.Background(), user.ID, "", "")
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate refresh token", 500)
	}

	// Note: Audit log is now created in LoginWithContext method

	return user, accessToken, refreshToken, nil
}

// LoginWithContext: 클라이언트 컨텍스트 정보를 포함하여 로그인을 수행합니다
func (s *Service) LoginWithContext(email, password, clientIP, userAgent string) (*domain.User, string, string, error) {
	// Get user by email (email is unique)
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	// Validate password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	// Get user roles for JWT token
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate JWT token with primary role (first role)
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	accessToken, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate access token", 500)
	}

	// Generate refresh token with client context
	ctx := context.Background()
	if clientIP == "" {
		clientIP = "127.0.0.1" // Default to localhost if no IP
	}
	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID, clientIP, userAgent)
	if err != nil {
		return nil, "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate refresh token", 500)
	}

	// Create audit log with client context using common helper
	ctx = context.WithValue(ctx, contextKeyClientIP, clientIP)
	ctx = context.WithValue(ctx, contextKeyUserAgent, userAgent)

	common.LogActionWithContext(ctx, s.auditLogRepo, &user.ID, domain.ActionUserLogin,
		"POST /api/v1/auth/login",
		map[string]interface{}{
			"email": user.Email,
		},
		clientIP,
		userAgent,
	)

	return user, accessToken, refreshToken, nil
}

// ValidateToken: JWT 토큰을 검증하고 사용자 정보를 반환합니다
func (s *Service) ValidateToken(tokenString string) (*domain.User, error) {
	// Check if token is blacklisted
	if s.blacklist != nil && s.blacklist.IsBlacklisted(context.Background(), tokenString) {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "token has been revoked", 401)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
	}

	if !token.Valid {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token claims", 401)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token claims", 401)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid user ID", 401)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "user not found", 401)
	}

	if !user.IsActive() {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	return user, nil
}

// Logout: 사용자를 로그아웃하고 토큰을 무효화합니다
func (s *Service) Logout(userID uuid.UUID, token string) error {
	ctx := context.Background()

	// Add access token to blacklist
	expiry := BlacklistTokenExpiry
	if err := s.blacklist.AddToBlacklist(ctx, token, expiry); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to invalidate access token", 500)
	}

	// Note: Refresh token deletion is handled separately via RevokeRefreshToken
	// This allows logout to work with just the access token

	// Log logout using common helper
	common.LogAction(ctx, s.auditLogRepo, &userID, domain.ActionUserLogout,
		"POST /api/v1/auth/logout",
		map[string]interface{}{
			"access_token_invalidated": true,
		},
	)

	return nil
}

// generateJWT generates a JWT token for a user
func (s *Service) generateJWT(userID uuid.UUID, username string, role domain.Role) (string, error) {
	now := time.Now()
	// Use default expiry if jwtExpiry is 0
	expiry := s.jwtExpiry
	if expiry == 0 {
		expiry = DefaultTokenExpiry
	}
	exp := now.Add(expiry)

	claims := jwt.MapClaims{
		"user_id":  userID.String(),
		"username": username,
		"role":     string(role),
		"exp":      exp.Unix(),
		"iat":      now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateAndStoreRefreshToken generates a refresh token and stores it in Redis
func (s *Service) generateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID, clientIP, userAgent string) (string, error) {
	if s.refreshTokenStore == nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "refresh token store not initialized", 500)
	}

	// Generate refresh token
	refreshToken, err := s.refreshTokenStore.GenerateRefreshToken()
	if err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate refresh token", 500)
	}

	// Set expiry
	expiry := s.refreshTokenExpiry
	if expiry == 0 {
		expiry = 7 * 24 * time.Hour // Default 7 days
	}

	// Create token data
	now := time.Now()
	tokenData := &cache.RefreshTokenData{
		UserID:    userID.String(),
		CreatedAt: now,
		ExpiresAt: now.Add(expiry),
		ClientIP:  clientIP,
		UserAgent: userAgent,
	}

	// Store in Redis
	if err := s.refreshTokenStore.StoreRefreshToken(ctx, refreshToken, tokenData, expiry); err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to store refresh token", 500)
	}

	return refreshToken, nil
}

// RefreshToken validates a refresh token and returns new access and refresh tokens
func (s *Service) RefreshToken(refreshToken string) (string, string, error) {
	ctx := context.Background()

	if s.refreshTokenStore == nil {
		return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "refresh token store not initialized", 500)
	}

	// Get refresh token data from Redis
	tokenData, err := s.refreshTokenStore.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid or expired refresh token", 401)
	}

	// Parse user ID
	userID, err := uuid.Parse(tokenData.UserID)
	if err != nil {
		return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "invalid user ID in refresh token", 500)
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return "", "", domain.NewDomainError(domain.ErrCodeUnauthorized, "user not found", 401)
	}

	// Check if user is active
	if !user.IsActive() {
		return "", "", domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	// Get user roles
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate new access token
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	newAccessToken, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate access token", 500)
	}

	// Generate new refresh token (token rotation)
	var newRefreshToken string
	if s.enableRotation {
		// Generate new refresh token
		newRefreshToken, err = s.generateAndStoreRefreshToken(ctx, user.ID, tokenData.ClientIP, tokenData.UserAgent)
		if err != nil {
			return "", "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate new refresh token", 500)
		}

		// Delete old refresh token
		if err := s.refreshTokenStore.DeleteRefreshToken(ctx, refreshToken); err != nil {
			// Log warning but continue (old token will expire anyway)
			// This is not critical as the old token will expire naturally
		}
	} else {
		// No rotation: reuse the same refresh token
		newRefreshToken = refreshToken
	}

	// Log token refresh
	common.LogAction(ctx, s.auditLogRepo, &user.ID, domain.ActionUserLogin, // Reuse login action for token refresh
		"POST /api/v1/auth/refresh",
		map[string]interface{}{
			"token_refreshed": true,
			"rotation_enabled": s.enableRotation,
		},
	)

	return newAccessToken, newRefreshToken, nil
}

// RevokeRefreshToken invalidates a refresh token
func (s *Service) RevokeRefreshToken(refreshToken string) error {
	ctx := context.Background()

	if s.refreshTokenStore == nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "refresh token store not initialized", 500)
	}

	// Get token data to get user ID for audit log
	tokenData, err := s.refreshTokenStore.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		// Token doesn't exist or expired, consider it already revoked
		return nil
	}

	// Delete refresh token
	if err := s.refreshTokenStore.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to revoke refresh token", 500)
	}

	// Log revocation
	userID, err := uuid.Parse(tokenData.UserID)
	if err == nil {
		common.LogAction(ctx, s.auditLogRepo, &userID, domain.ActionUserLogout,
			"POST /api/v1/auth/revoke",
			map[string]interface{}{
				"refresh_token_revoked": true,
			},
		)
	}

	return nil
}

// RevokeAllUserTokens invalidates all refresh tokens for a user
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	if s.refreshTokenStore == nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "refresh token store not initialized", 500)
	}

	// Delete all refresh tokens for the user
	if err := s.refreshTokenStore.DeleteUserRefreshTokens(ctx, userID.String()); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to revoke user refresh tokens", 500)
	}

	// Log revocation
	common.LogAction(ctx, s.auditLogRepo, &userID, domain.ActionUserLogout,
		"POST /api/v1/auth/revoke-all",
		map[string]interface{}{
			"all_refresh_tokens_revoked": true,
		},
	)

	return nil
}

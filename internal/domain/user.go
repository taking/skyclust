package domain

import (
	"time"

	"github.com/google/uuid"
)

// User: 시스템의 사용자를 나타내는 도메인 엔티티
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Username     string     `json:"username" gorm:"not null;size:50"` // 고유하지 않음 - 여러 사용자가 동일한 사용자명을 가질 수 있음
	Email        string     `json:"email" gorm:"uniqueIndex;not null;size:100"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;not null;size:255"`
	OIDCProvider string     `json:"oidc_provider,omitempty" gorm:"size:20"` // google, github, azure
	OIDCSubject  string     `json:"oidc_subject,omitempty" gorm:"size:100"` // OIDC subject ID
	Active       bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 관계
	// 참고: 자격증명은 이제 사용자 기반이 아닌 워크스페이스 기반입니다
	AuditLogs []AuditLog `json:"audit_logs,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	UserRoles []UserRole `json:"user_roles,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName: User의 테이블 이름을 반환합니다
func (User) TableName() string {
	return "users"
}

// IsActive: 사용자가 활성화되어 있는지 확인합니다
func (u *User) IsActive() bool {
	return u.Active
}

// Activate: 사용자를 활성화합니다
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// Deactivate: 사용자를 비활성화합니다
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

// UpdateProfile: 사용자 프로필 정보를 업데이트합니다
func (u *User) UpdateProfile(username, email string) error {
	if username == "" {
		return NewDomainError(ErrCodeValidationFailed, "username cannot be empty", 400)
	}
	if email == "" {
		return NewDomainError(ErrCodeValidationFailed, "email cannot be empty", 400)
	}

	u.Username = username
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// SetPasswordHash: 비밀번호 해시를 설정합니다
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
}

// CanAccessResource: 사용자가 리소스에 접근할 수 있는지 확인합니다
func (u *User) CanAccessResource(resourceUserID uuid.UUID, userRole Role) bool {
	return u.ID == resourceUserID || userRole == AdminRoleType
}

// IsAdmin: 사용자가 관리자 역할을 가지고 있는지 확인합니다
func (u *User) IsAdmin(userRoles []Role) bool {
	for _, role := range userRoles {
		if role == AdminRoleType {
			return true
		}
	}
	return false
}

// GetDisplayName: 사용자의 표시 이름을 반환합니다
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}

// IsOIDCUser: OIDC 사용자인지 확인합니다
func (u *User) IsOIDCUser() bool {
	return u.OIDCProvider != "" && u.OIDCSubject != ""
}

// SetOIDCInfo: OIDC 제공자 정보를 설정합니다
func (u *User) SetOIDCInfo(provider, subject string) {
	u.OIDCProvider = provider
	u.OIDCSubject = subject
	u.UpdatedAt = time.Now()
}

// UserFilters: 사용자 쿼리 필터
type UserFilters struct {
	Search string
	Role   string
	Status string
	Page   int
	Limit  int
}

// UserStats: 사용자 통계
type UserStats struct {
	TotalUsers    int64
	ActiveUsers   int64
	InactiveUsers int64
	NewUsersToday int64
}

// CreateUserRequest: 사용자 생성 요청
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate: CreateUserRequest의 유효성을 검사합니다
func (r *CreateUserRequest) Validate() error {
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return NewDomainError(ErrCodeValidationFailed, "username must be between 3 and 50 characters", 400)
	}
	if len(r.Email) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "email is required", 400)
	}
	if len(r.Password) < 8 {
		return NewDomainError(ErrCodeValidationFailed, "password must be at least 8 characters", 400)
	}
	return nil
}

// UpdateUserRequest: 사용자 업데이트 요청
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Validate: UpdateUserRequest의 유효성을 검사합니다
func (r *UpdateUserRequest) Validate() error {
	if r.Username != nil && (len(*r.Username) < 3 || len(*r.Username) > 50) {
		return NewDomainError(ErrCodeValidationFailed, "username must be between 3 and 50 characters", 400)
	}
	if r.Email != nil && len(*r.Email) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "email cannot be empty", 400)
	}
	if r.Password != nil && len(*r.Password) < 8 {
		return NewDomainError(ErrCodeValidationFailed, "password must be at least 8 characters", 400)
	}
	return nil
}

// OIDCLoginRequest: OIDC 로그인 요청
type OIDCLoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google github azure"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}

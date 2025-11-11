package workspace

import "skyclust/internal/domain"

// Validate: AddMemberRequest의 유효성을 검사합니다
func (r *AddMemberRequest) Validate() error {
	if len(r.Email) == 0 {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "email is required", 400)
	}
	// 간단한 이메일 형식 검증
	if !contains(r.Email, "@") {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "invalid email format", 400)
	}
	if len(r.Role) == 0 {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "role is required", 400)
	}
	if r.Role != "admin" && r.Role != "member" {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "role must be one of: admin, member", 400)
	}
	return nil
}

// Validate: UpdateMemberRoleRequest의 유효성을 검사합니다
func (r *UpdateMemberRoleRequest) Validate() error {
	if len(r.Role) == 0 {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "role is required", 400)
	}
	if r.Role != "admin" && r.Role != "member" {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "role must be one of: admin, member", 400)
	}
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

package domain

// Validate: CreateNotificationRequest의 유효성을 검사합니다
func (r *CreateNotificationRequest) Validate() error {
	if len(r.Title) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "title is required", 400)
	}
	if len(r.Title) > 200 {
		return NewDomainError(ErrCodeValidationFailed, "title must be less than 200 characters", 400)
	}
	if len(r.Message) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "message is required", 400)
	}
	if len(r.Message) > 1000 {
		return NewDomainError(ErrCodeValidationFailed, "message must be less than 1000 characters", 400)
	}
	if r.Type != "info" && r.Type != "warning" && r.Type != "error" && r.Type != "success" {
		return NewDomainError(ErrCodeValidationFailed, "type must be one of: info, warning, error, success", 400)
	}
	return nil
}

// Validate: UpdateNotificationPreferencesRequest의 유효성을 검사합니다
func (r *UpdateNotificationPreferencesRequest) Validate() error {
	// UpdateNotificationPreferencesRequest는 모든 필드가 선택적이므로 추가 검증 불필요
	// 단, 값이 제공된 경우 유효성 검사
	if r.EmailEnabled != nil {
		// Boolean 값이므로 추가 검증 불필요
	}
	if r.PushEnabled != nil {
		// Boolean 값이므로 추가 검증 불필요
	}
	return nil
}

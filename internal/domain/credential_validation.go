package domain

// Validate: CreateCredentialRequest의 유효성을 검사합니다
func (r *CreateCredentialRequest) Validate() error {
	if len(r.WorkspaceID) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "workspace_id is required", 400)
	}
	if len(r.Provider) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "provider is required", 400)
	}
	if r.Provider != "aws" && r.Provider != "gcp" && r.Provider != "openstack" && r.Provider != "azure" {
		return NewDomainError(ErrCodeValidationFailed, "provider must be one of: aws, gcp, openstack, azure", 400)
	}
	if len(r.Name) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "name is required", 400)
	}
	if len(r.Name) > 100 {
		return NewDomainError(ErrCodeValidationFailed, "name must be less than 100 characters", 400)
	}
	if r.Data == nil || len(r.Data) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "data is required", 400)
	}
	return nil
}

// Validate: UpdateCredentialRequest의 유효성을 검사합니다
func (r *UpdateCredentialRequest) Validate() error {
	if r.Name != nil {
		if len(*r.Name) == 0 {
			return NewDomainError(ErrCodeValidationFailed, "name cannot be empty", 400)
		}
		if len(*r.Name) > 100 {
			return NewDomainError(ErrCodeValidationFailed, "name must be less than 100 characters", 400)
		}
	}
	return nil
}

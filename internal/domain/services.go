package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DomainService provides core business logic that doesn't belong to a specific entity.
// This service contains cross-cutting business rules that involve multiple domain entities.
// Note: Currently, application services use repositories directly. This service is available
// for future use when complex business rules spanning multiple entities are needed.
type DomainService struct {
	userRepo      UserRepository
	workspaceRepo WorkspaceRepository
	vmRepo        VMRepository
	auditLogRepo  AuditLogRepository
}

// NewDomainService creates a new domain service
func NewDomainService(
	userRepo UserRepository,
	workspaceRepo WorkspaceRepository,
	vmRepo VMRepository,
	auditLogRepo AuditLogRepository,
) *DomainService {
	return &DomainService{
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		vmRepo:        vmRepo,
		auditLogRepo:  auditLogRepo,
	}
}

// UserDomainService provides user-specific business logic.
// This service encapsulates business rules specific to user operations.
// Note: Currently, application services use repositories directly. This service is available
// for future use when complex user-related business rules are needed.
type UserDomainService struct {
	*DomainService
}

// NewUserDomainService creates a new user domain service
func NewUserDomainService(domainService *DomainService) *UserDomainService {
	return &UserDomainService{DomainService: domainService}
}

// CreateUser creates a new user with business rules
func (s *UserDomainService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists (business rule)
	if existing, _ := s.userRepo.GetByEmail(req.Email); existing != nil {
		return nil, ErrUserAlreadyExists
	}

	if existing, _ := s.userRepo.GetByUsername(req.Username); existing != nil {
		return nil, NewDomainError(
			ErrCodeAlreadyExists,
			"username already exists",
			409,
		)
	}

	// Create user entity
	user := &User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return user, nil
}

// ValidateUserAccess validates if a user can access a resource
func (s *UserDomainService) ValidateUserAccess(ctx context.Context, userID uuid.UUID, resourceUserID uuid.UUID, userRole Role) bool {
	// Business rule: Users can access their own resources or admins can access any resource
	return userID == resourceUserID || userRole == AdminRoleType
}

// WorkspaceDomainService provides workspace-specific business logic.
// This service encapsulates business rules specific to workspace operations.
// Note: Currently, application services use repositories directly. This service is available
// for future use when complex workspace-related business rules are needed.
type WorkspaceDomainService struct {
	*DomainService
}

// NewWorkspaceDomainService creates a new workspace domain service
func NewWorkspaceDomainService(domainService *DomainService) *WorkspaceDomainService {
	return &WorkspaceDomainService{DomainService: domainService}
}

// CreateWorkspace creates a new workspace with business rules
func (s *WorkspaceDomainService) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest, ownerID uuid.UUID) (*Workspace, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if owner exists (business rule)
	owner, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return nil, NewDomainError(
			ErrCodeNotFound,
			"owner not found",
			404,
		)
	}
	if owner == nil {
		return nil, ErrUserNotFound
	}

	// Check if workspace name is unique for this owner (business rule)
	existingWorkspaces, err := s.workspaceRepo.GetByOwnerID(ctx, ownerID.String())
	if err != nil {
		return nil, NewDomainError(
			ErrCodeInternalError,
			"failed to check existing workspaces",
			500,
		)
	}

	for _, workspace := range existingWorkspaces {
		if workspace.Name == req.Name {
			return nil, NewDomainError(
				ErrCodeAlreadyExists,
				"workspace name already exists for this owner",
				409,
			)
		}
	}

	// Create workspace entity
	workspace := &Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID.String(),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return workspace, nil
}

// ValidateWorkspaceAccess validates if a user can access a workspace
func (s *WorkspaceDomainService) ValidateWorkspaceAccess(ctx context.Context, userID uuid.UUID, workspaceID string, userRole Role) (bool, error) {
	// Business rule: Workspace owners and admins can access any workspace
	if userRole == AdminRoleType {
		return true, nil
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	if workspace == nil {
		return false, ErrWorkspaceNotFound
	}

	// Check if user is the owner
	workspaceOwnerID, err := uuid.Parse(workspace.OwnerID)
	if err != nil {
		return false, NewDomainError(
			ErrCodeInternalError,
			"invalid workspace owner ID",
			500,
		)
	}

	return userID == workspaceOwnerID, nil
}

// VMDomainService provides VM-specific business logic.
// This service encapsulates business rules specific to VM operations.
// Note: Currently, application services use repositories directly. This service is available
// for future use when complex VM-related business rules are needed.
type VMDomainService struct {
	*DomainService
}

// NewVMDomainService creates a new VM domain service
func NewVMDomainService(domainService *DomainService) *VMDomainService {
	return &VMDomainService{DomainService: domainService}
}

// CreateVM creates a new VM with business rules
func (s *VMDomainService) CreateVM(ctx context.Context, req CreateVMRequest, userID uuid.UUID) (*VM, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if workspace exists and user has access (business rule)
	workspace, err := s.workspaceRepo.GetByID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, NewDomainError(
			ErrCodeInternalError,
			"failed to get workspace",
			500,
		)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Check if VM name is unique in workspace (business rule)
	existingVMs, err := s.vmRepo.GetByWorkspaceID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, NewDomainError(
			ErrCodeInternalError,
			"failed to check existing VMs",
			500,
		)
	}

	for _, vm := range existingVMs {
		if vm.Name == req.Name {
			return nil, ErrVMAlreadyExists
		}
	}

	// Create VM entity
	vm := &VM{
		ID:          uuid.New().String(),
		Name:        req.Name,
		WorkspaceID: req.WorkspaceID,
		Provider:    req.Provider,
		Type:        req.Type,
		Region:      req.Region,
		ImageID:     req.ImageID,
		Status:      VMStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	return vm, nil
}

// ValidateVMAccess validates if a user can access a VM
func (s *VMDomainService) ValidateVMAccess(ctx context.Context, userID uuid.UUID, vmID string, userRole Role) (bool, error) {
	// Business rule: Admins can access any VM
	if userRole == AdminRoleType {
		return true, nil
	}

	vm, err := s.vmRepo.GetByID(ctx, vmID)
	if err != nil {
		return false, err
	}
	if vm == nil {
		return false, ErrVMNotFound
	}

	// Check if user has access to the workspace
	workspace, err := s.workspaceRepo.GetByID(ctx, vm.WorkspaceID)
	if err != nil {
		return false, err
	}
	if workspace == nil {
		return false, ErrWorkspaceNotFound
	}

	workspaceOwnerID, err := uuid.Parse(workspace.OwnerID)
	if err != nil {
		return false, NewDomainError(
			ErrCodeInternalError,
			"invalid workspace owner ID",
			500,
		)
	}

	return userID == workspaceOwnerID, nil
}

// BusinessRuleService provides general business rules that apply across multiple resource types.
// This service validates access and ownership rules for various resources (users, workspaces, VMs).
// Note: Currently, application services handle access control directly. This service is available
// for future use when unified business rule validation is needed.
type BusinessRuleService struct {
	*DomainService
}

// NewBusinessRuleService creates a new business rule service
func NewBusinessRuleService(domainService *DomainService) *BusinessRuleService {
	return &BusinessRuleService{DomainService: domainService}
}

// ValidateResourceAccess validates access to any resource
func (s *BusinessRuleService) ValidateResourceAccess(ctx context.Context, userID uuid.UUID, resourceType string, resourceID string, userRole Role) (bool, error) {
	switch resourceType {
	case "user":
		resourceUserID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, NewDomainError(
				ErrCodeValidationFailed,
				"invalid resource ID",
				400,
			)
		}
		user, err := s.userRepo.GetByID(resourceUserID)
		if err != nil {
			return false, err
		}
		return user != nil && (userID == resourceUserID || userRole == AdminRoleType), nil
	case "workspace":
		workspaceService := NewWorkspaceDomainService(s.DomainService)
		return workspaceService.ValidateWorkspaceAccess(ctx, userID, resourceID, userRole)
	case "vm":
		vmService := NewVMDomainService(s.DomainService)
		return vmService.ValidateVMAccess(ctx, userID, resourceID, userRole)
	default:
		return false, NewDomainError(
			ErrCodeValidationFailed,
			"unknown resource type",
			400,
		)
	}
}

// ValidateResourceOwnership validates if a user owns a resource
func (s *BusinessRuleService) ValidateResourceOwnership(ctx context.Context, userID uuid.UUID, resourceType string, resourceID string) (bool, error) {
	switch resourceType {
	case "user":
		resourceUserID, err := uuid.Parse(resourceID)
		if err != nil {
			return false, NewDomainError(
				ErrCodeValidationFailed,
				"invalid resource ID",
				400,
			)
		}
		return userID == resourceUserID, nil
	case "workspace":
		workspace, err := s.workspaceRepo.GetByID(ctx, resourceID)
		if err != nil {
			return false, err
		}
		if workspace == nil {
			return false, ErrWorkspaceNotFound
		}
		workspaceOwnerID, err := uuid.Parse(workspace.OwnerID)
		if err != nil {
			return false, NewDomainError(
				ErrCodeInternalError,
				"invalid workspace owner ID",
				500,
			)
		}
		return userID == workspaceOwnerID, nil
	default:
		return false, NewDomainError(
			ErrCodeValidationFailed,
			"unknown resource type",
			400,
		)
	}
}

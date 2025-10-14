package di

import (
	"skyclust/internal/domain"
	"skyclust/internal/repository/postgres"
)

// Repositories holds all repository dependencies
type Repositories struct {
	UserRepo       domain.UserRepository
	CredentialRepo domain.CredentialRepository
	AuditLogRepo   domain.AuditLogRepository
	WorkspaceRepo  domain.WorkspaceRepository
	VMRepo         domain.VMRepository
}

// buildRepositories initializes all repositories
func (b *ContainerBuilder) buildRepositories(infra *Infrastructure) (*Repositories, error) {
	// Initialize repositories
	userRepo := postgres.NewUserRepository(infra.DB)
	credentialRepo := postgres.NewCredentialRepository(infra.DB)
	auditLogRepo := postgres.NewAuditLogRepository(infra.DB)
	workspaceRepo := postgres.NewWorkspaceRepository(infra.DB)
	vmRepo := postgres.NewVMRepository(infra.DB)

	return &Repositories{
		UserRepo:       userRepo,
		CredentialRepo: credentialRepo,
		AuditLogRepo:   auditLogRepo,
		WorkspaceRepo:  workspaceRepo,
		VMRepo:         vmRepo,
	}, nil
}

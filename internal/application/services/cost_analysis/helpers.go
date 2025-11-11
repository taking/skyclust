package cost_analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
)

// prefetchCredentialsByProvider prefetches and groups credentials by provider for efficient lookup
// Returns credentials grouped by provider, or error if workspace ID is invalid
func (s *Service) prefetchCredentialsByProvider(ctx context.Context, workspaceID string) (map[string][]*domain.Credential, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), 400)
	}

	allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		logger.Warnf("Failed to get credentials for workspace %s: %v, will use estimated costs", workspaceID, err)
		allCredentials = []*domain.Credential{}
	}

	credentialsByProvider := make(map[string][]*domain.Credential)
	for _, cred := range allCredentials {
		if cred.IsActive {
			credentialsByProvider[cred.Provider] = append(credentialsByProvider[cred.Provider], cred)
		}
	}

	return credentialsByProvider, nil
}

// calculateVMCostsWithHandling calculates VM costs with error handling and warning generation
func (s *Service) calculateVMCostsWithHandling(ctx context.Context, vm *domain.VM, startDate, endDate time.Time, credentialsByProvider map[string][]*domain.Credential) ([]CostData, CostWarning) {
	costs, err := s.calculateVMCostsOptimized(ctx, vm, startDate, endDate, credentialsByProvider)
	if err != nil {
		logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
		return nil, CostWarning{
			Code:         "VM_COST_CALCULATION_FAILED",
			Message:      fmt.Sprintf("Failed to calculate costs for VM %s: %v", vm.ID, err),
			Provider:     vm.Provider,
			ResourceType: ResourceTypeVM,
		}
	}
	return costs, CostWarning{}
}

// calculateClusterCostsWithHandling calculates Kubernetes cluster costs with error handling
func (s *Service) calculateClusterCostsWithHandling(ctx context.Context, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, CostWarning) {
	clusterCosts, clusterWarnings, err := s.calculateKubernetesCosts(ctx, workspaceID, startDate, endDate, includeNodeGroups)
	if err != nil {
		logger.Warnf("Failed to calculate Kubernetes costs: %v", err)
		return nil, nil, CostWarning{
			Code:         "KUBERNETES_COST_CALCULATION_FAILED",
			Message:      fmt.Sprintf("Failed to calculate Kubernetes cluster costs: %v", err),
			ResourceType: ResourceTypeCluster,
		}
	}
	return clusterCosts, clusterWarnings, CostWarning{}
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

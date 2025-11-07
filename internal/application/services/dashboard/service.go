package dashboard

import (
	"context"
	"fmt"
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

/**
 * Dashboard Service
 * 대시보드 요약 정보를 제공하는 서비스
 */

// Service 대시보드 서비스
type Service struct {
	workspaceRepo     domain.WorkspaceRepository
	vmRepo            domain.VMRepository
	credentialRepo    domain.CredentialRepository
	kubernetesService *kubernetesservice.Service
	networkService    *networkservice.Service
	logger            *zap.Logger
}

// NewService 새로운 대시보드 서비스 생성
func NewService(
	workspaceRepo domain.WorkspaceRepository,
	vmRepo domain.VMRepository,
	credentialRepo domain.CredentialRepository,
	kubernetesService *kubernetesservice.Service,
	networkService *networkservice.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		workspaceRepo:     workspaceRepo,
		vmRepo:            vmRepo,
		credentialRepo:    credentialRepo,
		kubernetesService: kubernetesService,
		networkService:    networkService,
		logger:            logger,
	}
}

// GetDashboardSummary 대시보드 요약 정보 조회
func (s *Service) GetDashboardSummary(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.DashboardSummary, error) {
	// 워크스페이스 존재 확인
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid workspace ID format", 400)
	}

	// VM 통계 조회
	vmStats, err := s.getVMStats(ctx, workspaceID, credentialID, region)
	if err != nil {
		s.logger.Warn("Failed to get VM stats", zap.Error(err))
		vmStats = &domain.VMStats{
			Total:      0,
			Running:    0,
			Stopped:    0,
			ByProvider: make(map[string]int),
		}
	}

	// Kubernetes 클러스터 통계 조회
	clusterStats, err := s.getClusterStats(ctx, workspaceID, credentialID, region)
	if err != nil {
		s.logger.Warn("Failed to get cluster stats", zap.Error(err))
		clusterStats = &domain.ClusterStats{
			Total:      0,
			Healthy:    0,
			ByProvider: make(map[string]int),
		}
	}

	// 네트워크 통계 조회
	networkStats, err := s.getNetworkStats(ctx, workspaceID, credentialID, region)
	if err != nil {
		s.logger.Warn("Failed to get network stats", zap.Error(err))
		networkStats = &domain.NetworkStats{
			VPCs:           0,
			Subnets:        0,
			SecurityGroups: 0,
		}
	}

	// 자격 증명 통계 조회
	credentialStats, err := s.getCredentialStats(ctx, workspaceUUID)
	if err != nil {
		s.logger.Warn("Failed to get credential stats", zap.Error(err))
		credentialStats = &domain.CredentialStats{
			Total:      0,
			ByProvider: make(map[string]int),
		}
	}

	// 멤버 통계 조회
	memberStats, err := s.getMemberStats(ctx, workspaceID)
	if err != nil {
		s.logger.Warn("Failed to get member stats", zap.Error(err))
		memberStats = &domain.MemberStats{
			Total: 0,
		}
	}

	return &domain.DashboardSummary{
		WorkspaceID: workspaceID,
		VMs:         *vmStats,
		Clusters:    *clusterStats,
		Networks:    *networkStats,
		Credentials: *credentialStats,
		Members:     *memberStats,
		LastUpdated: time.Now().Format(time.RFC3339),
	}, nil
}

// getVMStats VM 통계 조회
func (s *Service) getVMStats(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.VMStats, error) {
	vms, err := s.vmRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	stats := domain.VMStats{
		Total:      len(vms),
		Running:    0,
		Stopped:    0,
		ByProvider: make(map[string]int),
	}

	for _, vm := range vms {
		// region 필터링
		if region != nil && vm.Region != *region {
			continue
		}

		// 프로바이더별 카운트
		stats.ByProvider[vm.Provider]++

		// 상태별 카운트
		if vm.Status == domain.VMStatusRunning {
			stats.Running++
		} else if vm.Status == domain.VMStatusStopped {
			stats.Stopped++
		}
	}

	return &stats, nil
}

// getClusterStats Kubernetes 클러스터 통계 조회
func (s *Service) getClusterStats(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.ClusterStats, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid workspace ID format", 400)
	}

	// 워크스페이스의 모든 자격 증명 조회
	credentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, err
	}

	stats := domain.ClusterStats{
		Total:      0,
		Healthy:    0,
		ByProvider: make(map[string]int),
	}

	// 각 자격 증명에 대해 클러스터 조회
	for _, credential := range credentials {
		// credentialID 필터링
		if credentialID != nil && credential.ID.String() != *credentialID {
			continue
		}

		// Kubernetes 서비스를 통해 클러스터 조회
		// region이 지정된 경우 해당 리전만, 아니면 모든 리전 조회
		regionsToCheck := []string{}
		if region != nil {
			regionsToCheck = append(regionsToCheck, *region)
		} else {
			// 기본 리전 목록 (AWS: ap-northeast-3, GCP: asia-northeast3 등)
			switch credential.Provider {
			case "aws":
				regionsToCheck = []string{"ap-northeast-3", "ap-northeast-2", "us-east-1"}
			case "gcp":
				regionsToCheck = []string{"asia-northeast3", "asia-northeast1", "us-central1"}
			case "azure":
				regionsToCheck = []string{"koreacentral", "koreasouth", "eastus"}
			default:
				regionsToCheck = []string{""}
			}
		}

		for _, reg := range regionsToCheck {
			clusters, err := s.kubernetesService.ListEKSClusters(ctx, credential, reg)
			if err != nil {
				s.logger.Warn("Failed to list clusters for credential",
					zap.String("credential_id", credential.ID.String()),
					zap.String("provider", credential.Provider),
					zap.String("region", reg),
					zap.Error(err))
				continue
			}

			if clusters != nil && clusters.Clusters != nil {
				stats.Total += len(clusters.Clusters)
				stats.ByProvider[credential.Provider] += len(clusters.Clusters)

				// Healthy 클러스터 카운트 (status가 "ACTIVE" 또는 "RUNNING"인 경우)
				for _, cluster := range clusters.Clusters {
					if cluster.Status == "ACTIVE" || cluster.Status == "RUNNING" {
						stats.Healthy++
					}
				}
			}
		}
	}

	return &stats, nil
}

// getNetworkStats 네트워크 통계 조회
func (s *Service) getNetworkStats(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.NetworkStats, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid workspace ID format", 400)
	}

	// 워크스페이스의 모든 자격 증명 조회
	credentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, err
	}

	stats := domain.NetworkStats{
		VPCs:           0,
		Subnets:        0,
		SecurityGroups: 0,
	}

	// 각 자격 증명에 대해 네트워크 리소스 조회
	for _, credential := range credentials {
		// credentialID 필터링
		if credentialID != nil && credential.ID.String() != *credentialID {
			continue
		}

		// VPC 조회
		regionStr := ""
		if region != nil {
			regionStr = *region
		}
		vpcReq := networkservice.ListVPCsRequest{
			Region: regionStr,
		}
		vpcs, err := s.networkService.ListVPCs(ctx, credential, vpcReq)
		if err != nil {
			s.logger.Warn("Failed to list VPCs",
				zap.String("credential_id", credential.ID.String()),
				zap.String("provider", credential.Provider),
				zap.Error(err))
		} else if vpcs != nil {
			stats.VPCs += len(vpcs.VPCs)
		}

		// Subnet 조회 (VPC별로)
		if vpcs != nil {
			for _, vpc := range vpcs.VPCs {
				subnetReq := networkservice.ListSubnetsRequest{
					VPCID: vpc.ID,
					Region: regionStr,
				}
				subnets, err := s.networkService.ListSubnets(ctx, credential, subnetReq)
				if err != nil {
					s.logger.Warn("Failed to list subnets",
						zap.String("credential_id", credential.ID.String()),
						zap.String("vpc_id", vpc.ID),
						zap.Error(err))
				} else if subnets != nil {
					stats.Subnets += len(subnets.Subnets)
				}
			}
		}

		// Security Group 조회
		sgReq := networkservice.ListSecurityGroupsRequest{
			Region: regionStr,
		}
		securityGroups, err := s.networkService.ListSecurityGroups(ctx, credential, sgReq)
		if err != nil {
			s.logger.Warn("Failed to list security groups",
				zap.String("credential_id", credential.ID.String()),
				zap.String("provider", credential.Provider),
				zap.Error(err))
		} else if securityGroups != nil {
			stats.SecurityGroups += len(securityGroups.SecurityGroups)
		}
	}

	return &stats, nil
}

// getCredentialStats 자격 증명 통계 조회
func (s *Service) getCredentialStats(ctx context.Context, workspaceID uuid.UUID) (*domain.CredentialStats, error) {
	credentials, err := s.credentialRepo.GetByWorkspaceID(workspaceID)
	if err != nil {
		return nil, err
	}

	stats := domain.CredentialStats{
		Total:      len(credentials),
		ByProvider: make(map[string]int),
	}

	for _, credential := range credentials {
		stats.ByProvider[credential.Provider]++
	}

	return &stats, nil
}

// getMemberStats 멤버 통계 조회
func (s *Service) getMemberStats(ctx context.Context, workspaceID string) (*domain.MemberStats, error) {
	// Workspace repository를 통해 멤버 수 조회
	_, members, err := s.workspaceRepo.GetWorkspaceWithMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	stats := domain.MemberStats{
		Total: len(members),
	}

	return &stats, nil
}


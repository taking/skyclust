package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

/**
 * Dashboard Service
 * 대시보드 요약 정보를 제공하는 서비스
 */

// Service 대시보드 서비스
type Service struct {
	workspaceRepo         domain.WorkspaceRepository
	vmRepo                domain.VMRepository
	credentialRepo        domain.CredentialRepository
	kubernetesService     *kubernetesservice.Service
	networkService        *networkservice.Service
	eventService          domain.EventService
	logger                *zap.Logger
	natsConn              *nats.Conn
	debounceTimer         *time.Timer
	debounceMutex         sync.Mutex
	pendingRecalculations map[string]map[string]bool // workspace_id -> filter_key (credential_id:region)
}

// NewService 새로운 대시보드 서비스 생성
func NewService(
	workspaceRepo domain.WorkspaceRepository,
	vmRepo domain.VMRepository,
	credentialRepo domain.CredentialRepository,
	kubernetesService *kubernetesservice.Service,
	networkService *networkservice.Service,
	eventService domain.EventService,
	logger *zap.Logger,
) *Service {
	return &Service{
		workspaceRepo:         workspaceRepo,
		vmRepo:                vmRepo,
		credentialRepo:        credentialRepo,
		kubernetesService:     kubernetesService,
		networkService:        networkService,
		eventService:          eventService,
		logger:                logger,
		pendingRecalculations: make(map[string]map[string]bool),
	}
}

// SetNATSConnection NATS 연결을 설정합니다 (이벤트 구독을 위해)
func (s *Service) SetNATSConnection(conn *nats.Conn) {
	s.natsConn = conn
}

// GetDashboardSummary 대시보드 요약 정보 조회 (병렬 최적화 버전 사용)
func (s *Service) GetDashboardSummary(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.DashboardSummary, error) {
	return s.GetDashboardSummaryOptimized(ctx, workspaceID, credentialID, region)
}

// GetDashboardSummaryOptimized 대시보드 요약 정보를 병렬로 조회하여 성능 최적화
func (s *Service) GetDashboardSummaryOptimized(ctx context.Context, workspaceID string, credentialID *string, region *string) (*domain.DashboardSummary, error) {
	if err := s.validateWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid workspace ID format", 400)
	}

	stats := s.collectStatsInParallel(ctx, workspaceID, workspaceUUID, credentialID, region)
	summary := s.buildDashboardSummary(workspaceID, stats)

	// Dashboard summary 이벤트 발행 (SSE를 통해 실시간 업데이트)
	s.publishDashboardSummaryUpdate(ctx, summary, credentialID, region)

	return summary, nil
}

// publishDashboardSummaryUpdate 대시보드 요약 정보 업데이트 이벤트를 발행합니다
func (s *Service) publishDashboardSummaryUpdate(ctx context.Context, summary *domain.DashboardSummary, credentialID *string, region *string) {
	if s.eventService == nil {
		return
	}

	eventData := map[string]interface{}{
		"workspace_id": summary.WorkspaceID,
		"vms": map[string]interface{}{
			"total":       summary.VMs.Total,
			"running":     summary.VMs.Running,
			"stopped":     summary.VMs.Stopped,
			"by_provider": summary.VMs.ByProvider,
		},
		"clusters": map[string]interface{}{
			"total":       summary.Clusters.Total,
			"healthy":     summary.Clusters.Healthy,
			"by_provider": summary.Clusters.ByProvider,
		},
		"networks": map[string]interface{}{
			"vpcs":            summary.Networks.VPCs,
			"subnets":         summary.Networks.Subnets,
			"security_groups": summary.Networks.SecurityGroups,
		},
		"credentials": map[string]interface{}{
			"total":       summary.Credentials.Total,
			"by_provider": summary.Credentials.ByProvider,
		},
		"members": map[string]interface{}{
			"total": summary.Members.Total,
		},
		"last_updated": summary.LastUpdated,
	}

	if credentialID != nil {
		eventData["credential_id"] = *credentialID
	}
	if region != nil {
		eventData["region"] = *region
	}

	if err := s.eventService.Publish(ctx, "dashboard-summary-updated", eventData); err != nil {
		s.logger.Warn("Failed to publish dashboard summary update event",
			zap.String("workspace_id", summary.WorkspaceID),
			zap.Error(err))
	}
}

// validateWorkspace 워크스페이스 존재 여부를 확인합니다
func (s *Service) validateWorkspace(ctx context.Context, workspaceID string) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

// dashboardStats 모든 통계 데이터를 담는 구조체
type dashboardStats struct {
	vmStats            *domain.VMStats
	clusterStats       *domain.ClusterStats
	networkStats       *domain.NetworkStats
	credentialStats    *domain.CredentialStats
	memberStats        *domain.MemberStats
	vmStatsErr         error
	clusterStatsErr    error
	networkStatsErr    error
	credentialStatsErr error
	memberStatsErr     error
}

// collectStatsInParallel 모든 통계를 병렬로 수집합니다
func (s *Service) collectStatsInParallel(ctx context.Context, workspaceID string, workspaceUUID uuid.UUID, credentialID *string, region *string) *dashboardStats {
	var wg sync.WaitGroup
	var mu sync.Mutex
	stats := &dashboardStats{}

	// VM 통계 조회
	wg.Add(1)
	go s.collectVMStats(ctx, &wg, &mu, stats, workspaceID, credentialID, region)

	// Kubernetes 클러스터 통계 조회
	wg.Add(1)
	go s.collectClusterStats(ctx, &wg, &mu, stats, workspaceID, credentialID, region)

	// 네트워크 통계 조회
	wg.Add(1)
	go s.collectNetworkStats(ctx, &wg, &mu, stats, workspaceID, credentialID, region)

	// 자격 증명 통계 조회
	wg.Add(1)
	go s.collectCredentialStats(ctx, &wg, &mu, stats, workspaceUUID)

	// 멤버 통계 조회
	wg.Add(1)
	go s.collectMemberStats(ctx, &wg, &mu, stats, workspaceID)

	wg.Wait()
	return stats
}

// collectVMStats VM 통계를 수집합니다
func (s *Service) collectVMStats(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, stats *dashboardStats, workspaceID string, credentialID *string, region *string) {
	defer wg.Done()
	vmStats, err := s.getVMStats(ctx, workspaceID, credentialID, region)
	mu.Lock()
	stats.vmStats = vmStats
	stats.vmStatsErr = err
	mu.Unlock()
}

// collectClusterStats 클러스터 통계를 수집합니다
func (s *Service) collectClusterStats(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, stats *dashboardStats, workspaceID string, credentialID *string, region *string) {
	defer wg.Done()
	clusterStats, err := s.getClusterStats(ctx, workspaceID, credentialID, region)
	mu.Lock()
	stats.clusterStats = clusterStats
	stats.clusterStatsErr = err
	mu.Unlock()
}

// collectNetworkStats 네트워크 통계를 수집합니다
func (s *Service) collectNetworkStats(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, stats *dashboardStats, workspaceID string, credentialID *string, region *string) {
	defer wg.Done()
	networkStats, err := s.getNetworkStats(ctx, workspaceID, credentialID, region)
	mu.Lock()
	stats.networkStats = networkStats
	stats.networkStatsErr = err
	mu.Unlock()
}

// collectCredentialStats 자격 증명 통계를 수집합니다
func (s *Service) collectCredentialStats(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, stats *dashboardStats, workspaceUUID uuid.UUID) {
	defer wg.Done()
	credentialStats, err := s.getCredentialStats(ctx, workspaceUUID)
	mu.Lock()
	stats.credentialStats = credentialStats
	stats.credentialStatsErr = err
	mu.Unlock()
}

// collectMemberStats 멤버 통계를 수집합니다
func (s *Service) collectMemberStats(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, stats *dashboardStats, workspaceID string) {
	defer wg.Done()
	memberStats, err := s.getMemberStats(ctx, workspaceID)
	mu.Lock()
	stats.memberStats = memberStats
	stats.memberStatsErr = err
	mu.Unlock()
}

// buildDashboardSummary 통계 데이터로부터 대시보드 요약을 생성합니다
func (s *Service) buildDashboardSummary(workspaceID string, stats *dashboardStats) *domain.DashboardSummary {
	vmStats := s.handleVMStatsError(stats.vmStatsErr, stats.vmStats)
	clusterStats := s.handleClusterStatsError(stats.clusterStatsErr, stats.clusterStats)
	networkStats := s.handleNetworkStatsError(stats.networkStatsErr, stats.networkStats)
	credentialStats := s.handleCredentialStatsError(stats.credentialStatsErr, stats.credentialStats)
	memberStats := s.handleMemberStatsError(stats.memberStatsErr, stats.memberStats)

	return &domain.DashboardSummary{
		WorkspaceID: workspaceID,
		VMs:         *vmStats,
		Clusters:    *clusterStats,
		Networks:    *networkStats,
		Credentials: *credentialStats,
		Members:     *memberStats,
		LastUpdated: time.Now().Format(time.RFC3339),
	}
}

// handleVMStatsError VM 통계 에러를 처리하고 기본값을 반환합니다
func (s *Service) handleVMStatsError(err error, stats *domain.VMStats) *domain.VMStats {
	if err != nil {
		s.logger.Warn("Failed to get VM stats", zap.Error(err))
		return &domain.VMStats{
			Total:      0,
			Running:    0,
			Stopped:    0,
			ByProvider: make(map[string]int),
		}
	}
	return stats
}

// handleClusterStatsError 클러스터 통계 에러를 처리하고 기본값을 반환합니다
func (s *Service) handleClusterStatsError(err error, stats *domain.ClusterStats) *domain.ClusterStats {
	if err != nil {
		s.logger.Warn("Failed to get cluster stats", zap.Error(err))
		return &domain.ClusterStats{
			Total:      0,
			Healthy:    0,
			ByProvider: make(map[string]int),
		}
	}
	return stats
}

// handleNetworkStatsError 네트워크 통계 에러를 처리하고 기본값을 반환합니다
func (s *Service) handleNetworkStatsError(err error, stats *domain.NetworkStats) *domain.NetworkStats {
	if err != nil {
		s.logger.Warn("Failed to get network stats", zap.Error(err))
		return &domain.NetworkStats{
			VPCs:           0,
			Subnets:        0,
			SecurityGroups: 0,
		}
	}
	return stats
}

// handleCredentialStatsError 자격 증명 통계 에러를 처리하고 기본값을 반환합니다
func (s *Service) handleCredentialStatsError(err error, stats *domain.CredentialStats) *domain.CredentialStats {
	if err != nil {
		s.logger.Warn("Failed to get credential stats", zap.Error(err))
		return &domain.CredentialStats{
			Total:      0,
			ByProvider: make(map[string]int),
		}
	}
	return stats
}

// handleMemberStatsError 멤버 통계 에러를 처리하고 기본값을 반환합니다
func (s *Service) handleMemberStatsError(err error, stats *domain.MemberStats) *domain.MemberStats {
	if err != nil {
		s.logger.Warn("Failed to get member stats", zap.Error(err))
		return &domain.MemberStats{
			Total: 0,
		}
	}
	return stats
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
					VPCID:  vpc.ID,
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

// StartEventSubscriptions 리소스 변경 이벤트 구독을 시작합니다
func (s *Service) StartEventSubscriptions(ctx context.Context) error {
	if s.natsConn == nil {
		s.logger.Warn("NATS connection not available, skipping dashboard event subscriptions")
		return nil
	}

	// NATS 메시지 처리 헬퍼 함수 (압축 해제 포함)
	handleNATSMessage := func(m *nats.Msg) {
		data := m.Data

		// 압축 해제 시도
		if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
			// gzip 압축 해제
			decompressed, err := messaging.Decompress(data, messaging.CompressionGzip)
			if err != nil {
				s.logger.Warn("Failed to decompress gzip message, using raw data",
					zap.Error(err),
					zap.String("subject", m.Subject))
			} else {
				data = decompressed
			}
		} else if len(data) >= 4 && data[0] == 's' && data[1] == 'N' && data[2] == 'a' && data[3] == 'P' {
			// snappy 압축 해제
			decompressed, err := messaging.Decompress(data, messaging.CompressionSnappy)
			if err != nil {
				s.logger.Warn("Failed to decompress snappy message, using raw data",
					zap.Error(err),
					zap.String("subject", m.Subject))
			} else {
				data = decompressed
			}
		}

		// Event 구조체로 파싱 시도
		var event struct {
			Type        string                 `json:"type"`
			WorkspaceID string                 `json:"workspace_id,omitempty"`
			UserID      string                 `json:"user_id,omitempty"`
			Data        map[string]interface{} `json:"data"`
			Timestamp   int64                  `json:"timestamp"`
		}

		if err := json.Unmarshal(data, &event); err != nil {
			s.logger.Warn("Failed to parse event data",
				zap.Error(err),
				zap.String("subject", m.Subject))
			return
		}

		s.handleResourceChange(ctx, event.Type, event.Data, event.WorkspaceID)
	}

	// VM 이벤트 구독
	_, err := s.natsConn.Subscribe("cmp.events.vm.*.*.*.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VM created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.vm.*.*.*.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VM updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.vm.*.*.*.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VM deleted events", zap.Error(err))
	}

	// Kubernetes 클러스터 이벤트 구독
	_, err = s.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Kubernetes cluster created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Kubernetes cluster updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.kubernetes.*.*.*.clusters.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Kubernetes cluster deleted events", zap.Error(err))
	}

	// Network VPC 이벤트 구독
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VPC created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VPC updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.vpcs.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to VPC deleted events", zap.Error(err))
	}

	// Network Subnet 이벤트 구독
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Subnet created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Subnet updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.vpcs.*.subnets.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Subnet deleted events", zap.Error(err))
	}

	// Network Security Group 이벤트 구독
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Security Group created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Security Group updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.network.*.*.*.security-groups.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Security Group deleted events", zap.Error(err))
	}

	// Credential 이벤트 구독
	_, err = s.natsConn.Subscribe("cmp.events.credential.*.*.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Credential created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.credential.*.*.updated", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Credential updated events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.credential.*.*.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Credential deleted events", zap.Error(err))
	}

	// Workspace Member 이벤트 구독 (workspace_id는 이벤트에 포함됨)
	_, err = s.natsConn.Subscribe("cmp.events.workspace.*.members.created", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Workspace Member created events", zap.Error(err))
	}
	_, err = s.natsConn.Subscribe("cmp.events.workspace.*.members.deleted", handleNATSMessage)
	if err != nil {
		s.logger.Warn("Failed to subscribe to Workspace Member deleted events", zap.Error(err))
	}

	s.logger.Info("Dashboard event subscriptions started")
	return nil
}

// handleResourceChange 리소스 변경 이벤트를 처리합니다
func (s *Service) handleResourceChange(ctx context.Context, eventType string, eventData map[string]interface{}, eventWorkspaceID string) {
	var workspaceID string
	var credentialID *string
	var region *string

	// Workspace ID 추출
	if eventWorkspaceID != "" {
		workspaceID = eventWorkspaceID
	} else if wsID, ok := eventData["workspace_id"].(string); ok && wsID != "" {
		workspaceID = wsID
	} else if cid, ok := eventData["credential_id"].(string); ok && cid != "" {
		// Credential ID가 있으면 credential에서 workspace_id 추출
		credentialUUID, err := uuid.Parse(cid)
		if err != nil {
			s.logger.Warn("Failed to parse credential_id",
				zap.String("event_type", eventType),
				zap.String("credential_id", cid),
				zap.Error(err))
			return
		}

		credential, err := s.credentialRepo.GetByID(credentialUUID)
		if err != nil || credential == nil {
			s.logger.Warn("Failed to get credential for workspace_id extraction",
				zap.String("event_type", eventType),
				zap.String("credential_id", cid),
				zap.Error(err))
			return
		}

		workspaceID = credential.WorkspaceID.String()
		credentialID = &cid
	} else {
		s.logger.Debug("Event missing workspace_id and credential_id, skipping",
			zap.String("event_type", eventType),
			zap.Any("event_data", eventData))
		return
	}

	// Credential ID 추출 (이미 위에서 추출했으면 그대로 사용)
	if credentialID == nil {
		if cid, ok := eventData["credential_id"].(string); ok && cid != "" {
			credentialID = &cid
		}
	}

	// Region 추출
	if r, ok := eventData["region"].(string); ok && r != "" {
		region = &r
	}

	// Debounced 재계산 스케줄링
	s.scheduleSummaryRecalculation(workspaceID, credentialID, region)
}

// scheduleSummaryRecalculation 재계산을 스케줄링합니다 (Debouncing)
func (s *Service) scheduleSummaryRecalculation(workspaceID string, credentialID *string, region *string) {
	s.debounceMutex.Lock()
	defer s.debounceMutex.Unlock()

	// Pending recalculations에 추가
	if s.pendingRecalculations == nil {
		s.pendingRecalculations = make(map[string]map[string]bool)
	}
	if s.pendingRecalculations[workspaceID] == nil {
		s.pendingRecalculations[workspaceID] = make(map[string]bool)
	}

	// Filter key 생성: credential_id:region
	filterKey := fmt.Sprintf("%s:%s",
		getStringValue(credentialID),
		getStringValue(region))
	s.pendingRecalculations[workspaceID][filterKey] = true

	// 기존 타이머 취소
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
	}

	// 새 타이머 설정 (500ms 후 실행)
	s.debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
		s.processPendingRecalculations()
	})
}

// processPendingRecalculations 대기 중인 재계산을 처리합니다
func (s *Service) processPendingRecalculations() {
	s.debounceMutex.Lock()
	workspaces := s.pendingRecalculations
	s.pendingRecalculations = make(map[string]map[string]bool)
	s.debounceMutex.Unlock()

	// 각 workspace별로 재계산
	for workspaceID, filters := range workspaces {
		for filterKey := range filters {
			// filterKey 파싱: "credential_id:region"
			parts := strings.Split(filterKey, ":")
			var credentialID *string
			var region *string

			if len(parts) >= 1 && parts[0] != "" {
				credentialID = &parts[0]
			}
			if len(parts) >= 2 && parts[1] != "" {
				region = &parts[1]
			}

			// Dashboard Summary 재계산 및 이벤트 발행
			ctx := context.Background()
			_, err := s.GetDashboardSummary(ctx, workspaceID, credentialID, region)
			if err != nil {
				s.logger.Error("Failed to recalculate dashboard summary",
					zap.String("workspace_id", workspaceID),
					zap.String("credential_id", getStringValue(credentialID)),
					zap.String("region", getStringValue(region)),
					zap.Error(err))
				continue
			}

			// 이벤트는 GetDashboardSummary 내부에서 자동 발행됨
			s.logger.Debug("Dashboard summary recalculated",
				zap.String("workspace_id", workspaceID),
				zap.String("credential_id", getStringValue(credentialID)),
				zap.String("region", getStringValue(region)))
		}
	}
}

// getStringValue 포인터 문자열을 안전하게 반환합니다
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

package cost_analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// 기간 상수 (일 단위)
const (
	DaysPerWeek    = 7
	DaysPerMonth   = 30
	DaysPerQuarter = 90
	DaysPerYear    = 365
	MonthsPerYear  = 12
)

// 비율 상수
const (
	PercentageBase  = 100.0
	BudgetThreshold = 100.0
)

// Service: 비용 분석 서비스 구현체
type Service struct {
	vmRepo            domain.VMRepository
	credentialRepo    domain.CredentialRepository
	workspaceRepo     domain.WorkspaceRepository
	auditLogRepo      domain.AuditLogRepository
	credentialService domain.CredentialService
	kubernetesService *kubernetesservice.Service
	cache             cache.Cache
	logger            *zap.Logger
}

// NewService: 새로운 비용 분석 서비스를 생성합니다
func NewService(
	vmRepo domain.VMRepository,
	credentialRepo domain.CredentialRepository,
	workspaceRepo domain.WorkspaceRepository,
	auditLogRepo domain.AuditLogRepository,
	credentialService domain.CredentialService,
	kubernetesService *kubernetesservice.Service,
	cache cache.Cache,
) *Service {
	return &Service{
		vmRepo:            vmRepo,
		credentialRepo:    credentialRepo,
		workspaceRepo:     workspaceRepo,
		auditLogRepo:      auditLogRepo,
		credentialService: credentialService,
		kubernetesService: kubernetesService,
		cache:             cache,
		logger:            logger.DefaultLogger.GetLogger(),
	}
}

// GetCostSummary: 워크스페이스의 비용 요약을 조회합니다
// resourceTypes: 포함할 리소스 타입의 쉼표로 구분된 목록 (vm,cluster,node_group,node_pool,all)
// 비어있거나 "all"인 경우 모든 리소스 타입을 포함합니다
func (s *Service) GetCostSummary(ctx context.Context, workspaceID string, period string, resourceTypes string) (*CostSummary, error) {
	cacheKey := buildCostSummaryCacheKey(workspaceID, period, resourceTypes)

	// Try to get from cache first
	var cachedSummary CostSummary
	if found, _ := s.getFromCache(ctx, cacheKey, &cachedSummary); found {
		s.logger.Debug("Cost summary cache hit",
			zap.String("workspace_id", workspaceID),
			zap.String("period", period))
		return &cachedSummary, nil
	}

	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), HTTPStatusBadRequest)
	}

	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var allCosts []CostData
	var warnings []CostWarning

	var credentialsByProvider map[string][]*domain.Credential
	if includeVM {
		credentialsByProvider, err = s.prefetchCredentialsByProvider(ctx, workspaceID)
		if err != nil {
			return nil, err
		}
	}

	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), HTTPStatusInternalServerError)
		}

		for _, vm := range vms {
			costs, warning := s.calculateVMCostsWithHandling(ctx, vm, startDate, endDate, credentialsByProvider)
			if warning.Code != "" {
				warnings = append(warnings, warning)
				continue
			}
			allCosts = append(allCosts, costs...)
		}
	}

	if includeCluster {
		clusterCosts, clusterWarnings, warning := s.calculateClusterCostsWithHandling(ctx, workspaceID, startDate, endDate, includeNodeGroups)
		if warning.Code != "" {
			warnings = append(warnings, warning)
		} else {
			allCosts = append(allCosts, clusterCosts...)
			warnings = append(warnings, clusterWarnings...)
		}
	}

	summary := s.aggregateCosts(allCosts, startDate, endDate, period)
	summary.Warnings = warnings

	s.setCache(ctx, cacheKey, summary, CostAnalysisCacheTTL)

	return summary, nil
}

// parseResourceTypes: 리소스 타입 필터 문자열을 파싱합니다
// 반환값: (includeVM, includeCluster, includeNodeGroups)
func (s *Service) parseResourceTypes(resourceTypes string) (bool, bool, bool) {
	if resourceTypes == "" || resourceTypes == "all" {
		return true, true, true
	}

	types := strings.Split(strings.ToLower(resourceTypes), ",")
	includeVM := false
	includeCluster := false
	includeNodeGroups := false

	for _, t := range types {
		t = strings.TrimSpace(t)
		switch t {
		case ResourceTypeVM, "vms":
			includeVM = true
		case ResourceTypeCluster, "clusters", "kubernetes", "k8s":
			includeCluster = true
		case "node_group", "node_groups", "node_pool", "node_pools":
			includeNodeGroups = true
		case "all":
			includeVM = true
			includeCluster = true
			includeNodeGroups = true
		}
	}

	return includeVM, includeCluster, includeNodeGroups
}

// GetCostPredictions: 향후 기간에 대한 비용 예측을 생성합니다
func (s *Service) GetCostPredictions(ctx context.Context, workspaceID string, days int, resourceTypes string) ([]CostPrediction, []CostWarning, error) {
	cacheKey := buildCostPredictionsCacheKey(workspaceID, days, resourceTypes)

	var cachedPredictions []CostPrediction
	if found, _ := s.getFromCache(ctx, cacheKey, &cachedPredictions); found {
		s.logger.Debug("Cost predictions cache hit",
			zap.String("workspace_id", workspaceID),
			zap.Int("days", days))
		return cachedPredictions, []CostWarning{}, nil
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -CostPredictionHistoricalDays)

	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var historicalCosts []CostData
	var warnings []CostWarning

	var credentialsByProvider map[string][]*domain.Credential
	if includeVM {
		var err error
		credentialsByProvider, err = s.prefetchCredentialsByProvider(ctx, workspaceID)
		if err != nil {
			return nil, nil, err
		}
	}

	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), HTTPStatusInternalServerError)
		}

		for _, vm := range vms {
			costs, warning := s.calculateVMCostsWithHandling(ctx, vm, startDate, endDate, credentialsByProvider)
			if warning.Code != "" {
				warnings = append(warnings, warning)
				continue
			}
			historicalCosts = append(historicalCosts, costs...)
		}
	}

	if includeCluster {
		clusterCosts, clusterWarnings, warning := s.calculateClusterCostsWithHandling(ctx, workspaceID, startDate, endDate, includeNodeGroups)
		if warning.Code != "" {
			warnings = append(warnings, warning)
		} else {
			historicalCosts = append(historicalCosts, clusterCosts...)
			warnings = append(warnings, clusterWarnings...)
		}
	}

	predictions := s.generatePredictions(historicalCosts, days)

	s.setCache(ctx, cacheKey, predictions, CostAnalysisCacheTTL)

	return predictions, warnings, nil
}

// CheckBudgetAlerts: 워크스페이스가 예산 한도를 초과하는지 확인합니다
func (s *Service) CheckBudgetAlerts(ctx context.Context, workspaceID string, budgetLimit float64) ([]BudgetAlert, error) {
	summary, err := s.GetCostSummary(ctx, workspaceID, "1m", "all")
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get cost summary: %v", err), HTTPStatusInternalServerError)
	}

	var alerts []BudgetAlert
	percentage := (summary.TotalCost / budgetLimit) * PercentageBase

	if percentage >= BudgetThreshold {
		alerts = append(alerts, BudgetAlert{
			ID:          fmt.Sprintf("budget-exceeded-%s", workspaceID),
			WorkspaceID: workspaceID,
			BudgetLimit: budgetLimit,
			CurrentCost: summary.TotalCost,
			Percentage:  percentage,
			AlertLevel:  "critical",
			Message:     fmt.Sprintf("Budget exceeded by %.2f%%", percentage-100),
			CreatedAt:   time.Now(),
		})
	} else if percentage >= 80 {
		alerts = append(alerts, BudgetAlert{
			ID:          fmt.Sprintf("budget-warning-%s", workspaceID),
			WorkspaceID: workspaceID,
			BudgetLimit: budgetLimit,
			CurrentCost: summary.TotalCost,
			Percentage:  percentage,
			AlertLevel:  "warning",
			Message:     fmt.Sprintf("Budget usage at %.2f%%", percentage),
			CreatedAt:   time.Now(),
		})
	}

	return alerts, nil
}

// GetCostTrend: 워크스페이스의 비용 추세를 조회합니다
func (s *Service) GetCostTrend(ctx context.Context, workspaceID string, period string, resourceTypes string) (*CostTrend, error) {
	// Generate cache key
	cacheKey := fmt.Sprintf("cost_trend:%s:%s:%s", workspaceID, period, resourceTypes)

	// Try to get from cache first
	if s.cache != nil {
		var cachedTrend CostTrend
		if err := s.cache.Get(ctx, cacheKey, &cachedTrend); err == nil {
			s.logger.Debug("Cost trend cache hit",
				zap.String("workspace_id", workspaceID),
				zap.String("period", period))
			return &cachedTrend, nil
		}
	}

	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), HTTPStatusBadRequest)
	}

	// Parse resource types filter
	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var allCosts []CostData
	var warnings []CostWarning

	// Calculate VM costs if requested
	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), HTTPStatusInternalServerError)
		}

		// Optimize: Pre-fetch all credentials for the workspace to avoid N+1 queries
		workspaceUUID, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), HTTPStatusBadRequest)
		}

		allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
		if err != nil {
			s.logger.Warn("Failed to get credentials for workspace, will use estimated costs",
				zap.String("workspace_id", workspaceID),
				zap.Error(err))
			allCredentials = []*domain.Credential{}
		}

		// Group credentials by provider for efficient lookup
		credentialsByProvider := make(map[string][]*domain.Credential)
		for _, cred := range allCredentials {
			if cred.IsActive {
				credentialsByProvider[cred.Provider] = append(credentialsByProvider[cred.Provider], cred)
			}
		}

		for _, vm := range vms {
			costs, err := s.calculateVMCostsOptimized(ctx, vm, startDate, endDate, credentialsByProvider)
			if err != nil {
				s.logger.Warn("Failed to calculate VM costs",
					zap.String("vm_id", vm.ID),
					zap.Error(err))
				warnings = append(warnings, CostWarning{
					Code:         "VM_COST_CALCULATION_FAILED",
					Message:      fmt.Sprintf("Failed to calculate costs for VM %s: %v", vm.ID, err),
					Provider:     vm.Provider,
					ResourceType: ResourceTypeVM,
				})
				continue
			}
			allCosts = append(allCosts, costs...)
		}
	}

	// Calculate Kubernetes cluster costs if requested
	if includeCluster {
		clusterCosts, clusterWarnings, err := s.calculateKubernetesCosts(ctx, workspaceID, startDate, endDate, includeNodeGroups)
		if err != nil {
			s.logger.Warn("Failed to calculate Kubernetes costs", zap.Error(err))
			warnings = append(warnings, CostWarning{
				Code:         "KUBERNETES_COST_CALCULATION_FAILED",
				Message:      fmt.Sprintf("Failed to calculate Kubernetes cluster costs: %v", err),
				ResourceType: ResourceTypeCluster,
			})
		} else {
			allCosts = append(allCosts, clusterCosts...)
			warnings = append(warnings, clusterWarnings...)
		}
	}

	// Aggregate daily costs
	dailyCosts := s.aggregateDailyCosts(allCosts)

	// Calculate trend direction and percentage change
	trendDirection, changePercentage := s.calculateTrendMetrics(dailyCosts)

	trend := &CostTrend{
		DailyCosts:       dailyCosts,
		TrendDirection:   trendDirection,
		ChangePercentage: changePercentage,
		Warnings:         warnings,
	}

	s.setCache(ctx, cacheKey, trend, CostAnalysisCacheTTL)

	return trend, nil
}

// GetCostBreakdown: 워크스페이스의 비용 세부 내역을 조회합니다
func (s *Service) GetCostBreakdown(ctx context.Context, workspaceID string, period string, dimension string, resourceTypes string) (*CostBreakdown, []CostWarning, error) {
	cacheKey := buildCostBreakdownCacheKey(workspaceID, period, dimension, resourceTypes)

	var cachedBreakdown CostBreakdown
	if found, _ := s.getFromCache(ctx, cacheKey, &cachedBreakdown); found {
		s.logger.Debug("Cost breakdown cache hit",
			zap.String("workspace_id", workspaceID),
			zap.String("period", period),
			zap.String("dimension", dimension))
		return &cachedBreakdown, []CostWarning{}, nil
	}

	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), HTTPStatusBadRequest)
	}

	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var allCosts []CostData
	var warnings []CostWarning

	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), HTTPStatusInternalServerError)
		}

		// Optimize: Pre-fetch all credentials for the workspace to avoid N+1 queries
		workspaceUUID, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), HTTPStatusBadRequest)
		}

		allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
		if err != nil {
			s.logger.Warn("Failed to get credentials for workspace, will use estimated costs",
				zap.String("workspace_id", workspaceID),
				zap.Error(err))
			allCredentials = []*domain.Credential{}
		}

		// Group credentials by provider for efficient lookup
		credentialsByProvider := make(map[string][]*domain.Credential)
		for _, cred := range allCredentials {
			if cred.IsActive {
				credentialsByProvider[cred.Provider] = append(credentialsByProvider[cred.Provider], cred)
			}
		}

		for _, vm := range vms {
			costs, err := s.calculateVMCostsOptimized(ctx, vm, startDate, endDate, credentialsByProvider)
			if err != nil {
				s.logger.Warn("Failed to calculate VM costs",
					zap.String("vm_id", vm.ID),
					zap.Error(err))
				warnings = append(warnings, CostWarning{
					Code:         "VM_COST_CALCULATION_FAILED",
					Message:      fmt.Sprintf("Failed to calculate costs for VM %s: %v", vm.ID, err),
					Provider:     vm.Provider,
					ResourceType: ResourceTypeVM,
				})
				continue
			}
			allCosts = append(allCosts, costs...)
		}
	}

	if includeCluster {
		clusterCosts, clusterWarnings, warning := s.calculateClusterCostsWithHandling(ctx, workspaceID, startDate, endDate, includeNodeGroups)
		if warning.Code != "" {
			warnings = append(warnings, warning)
		} else {
			allCosts = append(allCosts, clusterCosts...)
			warnings = append(warnings, clusterWarnings...)
		}
	}

	// Aggregate costs by dimension
	breakdown := s.aggregateCostBreakdown(allCosts, dimension)

	s.setCache(ctx, cacheKey, breakdown, CostAnalysisCacheTTL)

	return breakdown, warnings, nil
}

// GetCostComparison: 기간 간 비용 비교를 조회합니다
func (s *Service) GetCostComparison(ctx context.Context, workspaceID string, currentPeriod string, comparePeriod string) (*CostComparison, error) {
	// Get current period summary
	currentSummary, err := s.GetCostSummary(ctx, workspaceID, currentPeriod, "all")
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get current period summary: %v", err), HTTPStatusInternalServerError)
	}

	// Calculate compare period dates
	now := time.Now()
	var compareStartDate, compareEndDate time.Time

	compareStartDate, compareEndDate, err = s.parseComparePeriod(comparePeriod, now)
	if err != nil {
		return nil, err
	}

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), HTTPStatusInternalServerError)
	}

	// Optimize: Pre-fetch all credentials for the workspace to avoid N+1 queries
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), HTTPStatusBadRequest)
	}

	allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		s.logger.Warn("Failed to get credentials for workspace, will use estimated costs",
			zap.String("workspace_id", workspaceID),
			zap.Error(err))
		allCredentials = []*domain.Credential{}
	}

	// Group credentials by provider for efficient lookup
	credentialsByProvider := make(map[string][]*domain.Credential)
	for _, cred := range allCredentials {
		if cred.IsActive {
			credentialsByProvider[cred.Provider] = append(credentialsByProvider[cred.Provider], cred)
		}
	}

	// Calculate compare period costs
	var compareCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCostsOptimized(ctx, vm, compareStartDate, compareEndDate, credentialsByProvider)
		if err != nil {
			s.logger.Warn("Failed to calculate VM costs",
				zap.String("vm_id", vm.ID),
				zap.Error(err))
			continue
		}
		compareCosts = append(compareCosts, costs...)
	}

	// Aggregate compare period costs
	compareTotal := 0.0
	for _, cost := range compareCosts {
		compareTotal += cost.Amount
	}

	// Calculate comparison metrics
	costChange := currentSummary.TotalCost - compareTotal
	var percentageChange float64
	if compareTotal > 0 {
		percentageChange = (costChange / compareTotal) * PercentageBase
	}

	return &CostComparison{
		CurrentPeriod: PeriodComparison{
			Period: currentPeriod,
			Cost:   currentSummary.TotalCost,
		},
		PreviousPeriod: PeriodComparison{
			Period: comparePeriod,
			Cost:   compareTotal,
		},
		Comparison: ComparisonData{
			CostChange:       costChange,
			PercentageChange: percentageChange,
		},
	}, nil
}

// Helper methods

// parseComparePeriod: 비교 기간 문자열을 파싱하여 시작일과 종료일을 반환합니다
func (s *Service) parseComparePeriod(period string, now time.Time) (time.Time, time.Time, error) {
	switch period {
	case "7d":
		compareEndDate := now.AddDate(0, 0, -DaysPerWeek)
		return compareEndDate.AddDate(0, 0, -DaysPerWeek), compareEndDate, nil
	case "30d", "1m":
		compareEndDate := now.AddDate(0, 0, -DaysPerMonth)
		return compareEndDate.AddDate(0, 0, -DaysPerMonth), compareEndDate, nil
	case "90d", "3m":
		compareEndDate := now.AddDate(0, 0, -DaysPerQuarter)
		return compareEndDate.AddDate(0, 0, -DaysPerQuarter), compareEndDate, nil
	case "1y":
		compareEndDate := now.AddDate(-1, 0, 0)
		return compareEndDate.AddDate(-1, 0, 0), compareEndDate, nil
	default:
		return time.Time{}, time.Time{}, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported compare period: %s", period), HTTPStatusBadRequest)
	}
}

// parsePeriod: 기간 문자열을 파싱하여 시작일과 종료일을 반환합니다
func (s *Service) parsePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now()

	switch period {
	case "7d":
		return now.AddDate(0, 0, -DaysPerWeek), now, nil
	case "30d", "1m":
		return now.AddDate(0, 0, -DaysPerMonth), now, nil
	case "90d", "3m":
		return now.AddDate(0, 0, -DaysPerQuarter), now, nil
	case "1y":
		return now.AddDate(-1, 0, 0), now, nil
	default:
		return time.Time{}, time.Time{}, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported period: %s", period), HTTPStatusBadRequest)
	}
}

// calculateVMCostsOptimized: 미리 가져온 자격증명을 사용하여 VM 비용을 계산합니다 (N+1 쿼리 방지를 위한 최적화)
func (s *Service) calculateVMCostsOptimized(ctx context.Context, vm *domain.VM, startDate, endDate time.Time, credentialsByProvider map[string][]*domain.Credential) ([]CostData, error) {
	credentials, exists := credentialsByProvider[vm.Provider]
	if !exists || len(credentials) == 0 {
		s.logger.Warn("No credentials found for provider, falling back to estimated costs",
			zap.String("provider", vm.Provider),
			zap.String("workspace_id", vm.WorkspaceID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	credential := s.selectActiveCredential(credentials)
	if credential == nil {
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	return s.getProviderCosts(ctx, credential, vm, startDate, endDate)
}

// selectActiveCredential 활성 자격증명을 선택합니다
func (s *Service) selectActiveCredential(credentials []*domain.Credential) *domain.Credential {
	for _, cred := range credentials {
		if cred.IsActive {
			return cred
		}
	}

	// 활성 자격증명이 없으면 첫 번째 자격증명 사용
	if len(credentials) > 0 {
		return credentials[0]
	}

	return nil
}

// getProviderCosts 프로바이더별 비용을 조회합니다
func (s *Service) getProviderCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	switch vm.Provider {
	case ProviderAWS:
		return s.getAWSCosts(ctx, credential, vm, startDate, endDate)
	case ProviderGCP:
		return s.getGCPCosts(ctx, credential, vm, startDate, endDate)
	default:
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}
}

// calculateEstimatedCosts: API를 사용할 수 없을 때 VM 사양을 기반으로 비용을 계산합니다
func (s *Service) calculateEstimatedCosts(vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	var costs []CostData
	current := startDate

	hourlyRate := s.getVMHourlyRate(vm)

	for current.Before(endDate) {
		nextDay := current.AddDate(0, 0, 1)
		if nextDay.After(endDate) {
			nextDay = endDate
		}

		hours := nextDay.Sub(current).Hours()
		dailyCost := hourlyRate * hours

		// Determine service name based on provider
		service := s.getVMServiceName(vm.Provider)

		costs = append(costs, CostData{
			Date:         current,
			Amount:       dailyCost,
			Currency:     CurrencyUSD,
			Service:      service,
			ResourceID:   vm.ID,
			ResourceType: ResourceTypeVM,
			Provider:     vm.Provider,
			Region:       vm.Region,
			WorkspaceID:  vm.WorkspaceID,
		})

		current = nextDay
	}

	return costs, nil
}

// getVMHourlyRate: VM의 시간당 요금을 반환합니다
func (s *Service) getVMHourlyRate(vm *domain.VM) float64 {
	// Mock pricing based on VM specifications
	// In reality, this would come from cloud provider pricing APIs

	baseRate := BaseHourlyRate

	cpuMultiplier := float64(vm.CPUs) * CPUMultiplierPerCore
	memoryMultiplier := float64(vm.Memory) * MemoryMultiplierPerGB
	storageMultiplier := float64(vm.Storage) * StorageMultiplierPerGB

	var providerMultiplier float64
	switch vm.Provider {
	case ProviderAWS:
		providerMultiplier = ProviderMultiplierAWS
	case ProviderGCP:
		providerMultiplier = ProviderMultiplierGCP
	case ProviderAzure:
		providerMultiplier = ProviderMultiplierAzure
	default:
		providerMultiplier = ProviderMultiplierAWS
	}

	return (baseRate + cpuMultiplier + memoryMultiplier + storageMultiplier) * providerMultiplier
}

// getVMServiceName: 프로바이더에 따라 VM의 서비스 이름을 반환합니다
func (s *Service) getVMServiceName(provider string) string {
	switch provider {
	case ProviderAWS:
		return ServiceNameAWSEC2
	case ProviderGCP:
		return ServiceNameGCPCompute
	case ProviderAzure:
		return ServiceNameAzureVirtualMachines
	default:
		return ServiceNameDefaultCompute
	}
}

// aggregateCosts: 비용 데이터를 집계하여 요약을 생성합니다

// generatePredictions: 과거 비용 데이터를 기반으로 예측을 생성합니다

// calculateVariance: 값들의 분산을 계산합니다

// aggregateDailyCosts: 비용을 일별로 집계합니다

// calculateTrendMetrics: 추세 방향과 변화율을 계산합니다

// aggregateCostBreakdown: 차원별로 비용을 집계합니다 (서비스, 프로바이더, 리전)

// getAWSCosts: AWS Cost Explorer API에서 실제 비용을 조회합니다

// getAWSCostExplorerClient: 자격증명으로부터 AWS Cost Explorer 클라이언트를 생성합니다

// queryAWSCostExplorer: 주어진 파라미터로 AWS Cost Explorer API를 조회합니다

// parseAWSCostExplorerResults: AWS Cost Explorer API 결과를 CostData 슬라이스로 파싱합니다

// getGCPCosts: GCP Cloud Billing API에서 실제 비용을 조회합니다

// parseFloat: 다양한 형식을 처리하여 문자열을 float64로 파싱합니다

// calculateKubernetesCosts: 워크스페이스의 Kubernetes 클러스터 비용을 계산합니다
// 반환값: 비용, 경고, 에러
func (s *Service) calculateKubernetesCosts(ctx context.Context, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), HTTPStatusBadRequest)
	}

	// Get all credentials for the workspace
	allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get credentials: %v", err), HTTPStatusInternalServerError)
	}

	var allCosts []CostData
	var warnings []CostWarning

	// Group credentials by provider
	credentialsByProvider := make(map[string][]*domain.Credential)
	for _, cred := range allCredentials {
		if cred.IsActive {
			credentialsByProvider[cred.Provider] = append(credentialsByProvider[cred.Provider], cred)
		}
	}

	// Calculate costs for each provider
	for provider, credentials := range credentialsByProvider {
		if len(credentials) == 0 {
			continue
		}

		// Use the first active credential for each provider
		credential := credentials[0]

		switch provider {
		case ProviderAWS:
			costs, providerWarnings, err := s.getAWSKubernetesCosts(ctx, credential, workspaceID, startDate, endDate, includeNodeGroups)
			if err != nil {
				s.logger.Warn("Failed to get AWS Kubernetes costs", zap.Error(err))
				warnings = append(warnings, s.formatKubernetesErrorWarning(err, ProviderAWS, credential))
				continue
			}
			allCosts = append(allCosts, costs...)
			warnings = append(warnings, providerWarnings...)
		case ProviderGCP:
			costs, providerWarnings, err := s.getGCPKubernetesCosts(ctx, credential, workspaceID, startDate, endDate, includeNodeGroups)
			if err != nil {
				s.logger.Warn("Failed to get GCP Kubernetes costs", zap.Error(err))
				warnings = append(warnings, s.formatKubernetesErrorWarning(err, ProviderGCP, credential))
				continue
			}
			allCosts = append(allCosts, costs...)
			warnings = append(warnings, providerWarnings...)
		default:
			// Skip other providers for now
			continue
		}
	}

	return allCosts, warnings, nil
}

// formatKubernetesErrorWarning: 에러를 사용자 친화적인 경고로 포맷합니다
func (s *Service) formatKubernetesErrorWarning(err error, provider string, credential *domain.Credential) CostWarning {
	errMsg := err.Error()

	code, message := s.parseKubernetesError(errMsg, provider)

	return CostWarning{
		Code:         code,
		Message:      message,
		Provider:     provider,
		ResourceType: ResourceTypeCluster,
	}
}

// parseKubernetesError: 에러 메시지를 파싱하여 적절한 에러 코드와 메시지를 반환합니다
func (s *Service) parseKubernetesError(errMsg, provider string) (string, string) {
	if s.isPermissionError(errMsg) {
		return s.getPermissionErrorMessage(provider)
	}

	if s.isAPIEnabledError(errMsg) {
		return s.getAPIEnabledErrorMessage(provider)
	}

	if s.isCredentialError(errMsg) {
		return "CREDENTIAL_ERROR", fmt.Sprintf("Invalid or expired %s credentials", provider)
	}

	return "KUBERNETES_COST_API_ERROR", fmt.Sprintf("Failed to retrieve %s Kubernetes cluster costs", provider)
}

// isPermissionError: 에러가 권한 관련인지 확인합니다
func (s *Service) isPermissionError(errMsg string) bool {
	return strings.Contains(errMsg, "AccessDeniedException") || strings.Contains(errMsg, "not authorized")
}

// isAPIEnabledError: 에러가 API 활성화 관련인지 확인합니다
func (s *Service) isAPIEnabledError(errMsg string) bool {
	return strings.Contains(errMsg, "SERVICE_DISABLED") || strings.Contains(errMsg, "not been used") || strings.Contains(errMsg, "disabled")
}

// isCredentialError: 에러가 자격증명 관련인지 확인합니다
func (s *Service) isCredentialError(errMsg string) bool {
	return strings.Contains(errMsg, "credentials") || strings.Contains(errMsg, "authentication")
}

// getPermissionErrorMessage: 프로바이더에 대한 권한 에러 코드와 메시지를 반환합니다
func (s *Service) getPermissionErrorMessage(provider string) (string, string) {
	code := "API_PERMISSION_DENIED"
	var message string

	switch provider {
	case ProviderAWS:
		message = "AWS IAM user does not have permission to access Cost Explorer API. Please grant 'ce:GetCostAndUsage' permission."
	case ProviderGCP:
		message = "GCP service account does not have permission to access Cloud Billing API."
	default:
		message = fmt.Sprintf("%s service account does not have required permissions", provider)
	}

	return code, message
}

// getAPIEnabledErrorMessage: 프로바이더에 대한 API 활성화 에러 코드와 메시지를 반환합니다
func (s *Service) getAPIEnabledErrorMessage(provider string) (string, string) {
	code := "API_NOT_ENABLED"
	var message string

	switch provider {
	case ProviderGCP:
		message = "GCP Cloud Billing API is not enabled. Please enable it in the GCP Console."
	case ProviderAWS:
		message = "AWS Cost Explorer API access is not configured for this account."
	default:
		message = fmt.Sprintf("%s cost API is not enabled", provider)
	}

	return code, message
}

// getAWSKubernetesCosts: AWS Cost Explorer API에서 EKS 비용을 조회합니다
// 반환값: 비용, 경고, 에러

// getGCPKubernetesCosts: GCP Cloud Billing API에서 GKE 비용을 조회합니다
// 반환값: 비용, 경고, 에러

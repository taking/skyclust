package cost_analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	serviceconstants "skyclust/internal/application/services"
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"sort"
	"strings"
	"time"

	billingv1 "cloud.google.com/go/billing/apiv1"
	billingpb "cloud.google.com/go/billing/apiv1/billingpb"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/option"
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
	cacheKey := fmt.Sprintf("cost_summary:%s:%s:%s", workspaceID, period, resourceTypes)

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
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), 400)
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
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
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
	cacheKey := fmt.Sprintf("cost_predictions:%s:%d:%s", workspaceID, days, resourceTypes)

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
			return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
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
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get cost summary: %v", err), 500)
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
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), 400)
	}

	// Parse resource types filter
	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var allCosts []CostData
	var warnings []CostWarning

	// Calculate VM costs if requested
	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
		}

		// Optimize: Pre-fetch all credentials for the workspace to avoid N+1 queries
		workspaceUUID, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), 400)
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
	cacheKey := fmt.Sprintf("cost_breakdown:%s:%s:%s:%s", workspaceID, period, dimension, resourceTypes)

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
		return nil, nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid period: %v", err), 400)
	}

	includeVM, includeCluster, includeNodeGroups := s.parseResourceTypes(resourceTypes)

	var allCosts []CostData
	var warnings []CostWarning

	if includeVM {
		vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
		}

		for _, vm := range vms {
			costs, err := s.calculateVMCosts(ctx, vm, startDate, endDate)
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
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get current period summary: %v", err), 500)
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
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
	}

	// Optimize: Pre-fetch all credentials for the workspace to avoid N+1 queries
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), 400)
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
		return time.Time{}, time.Time{}, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported compare period: %s", period), 400)
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
		return time.Time{}, time.Time{}, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported period: %s", period), 400)
	}
}

// calculateVMCostsOptimized: 미리 가져온 자격증명을 사용하여 VM 비용을 계산합니다 (N+1 쿼리 방지를 위한 최적화)
func (s *Service) calculateVMCostsOptimized(ctx context.Context, vm *domain.VM, startDate, endDate time.Time, credentialsByProvider map[string][]*domain.Credential) ([]CostData, error) {
	// Get credentials from pre-fetched map
	credentials, exists := credentialsByProvider[vm.Provider]
	if !exists || len(credentials) == 0 {
		s.logger.Warn("No credentials found for provider, falling back to estimated costs",
			zap.String("provider", vm.Provider),
			zap.String("workspace_id", vm.WorkspaceID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Use the first active credential
	var credential *domain.Credential
	for _, cred := range credentials {
		if cred.IsActive {
			credential = cred
			break
		}
	}

	if credential == nil && len(credentials) > 0 {
		credential = credentials[0]
	}

	// Get actual costs from cloud provider API
	switch vm.Provider {
	case ProviderAWS:
		return s.getAWSCosts(ctx, credential, vm, startDate, endDate)
	case ProviderGCP:
		return s.getGCPCosts(ctx, credential, vm, startDate, endDate)
	default:
		// For other providers, use estimated costs
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}
}

// calculateVMCosts: VM 비용을 계산합니다 (원본 메서드, 하위 호환성을 위해 유지)
func (s *Service) calculateVMCosts(ctx context.Context, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Get credentials for the provider and workspace
	workspaceUUID, err := uuid.Parse(vm.WorkspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), 400)
	}

	credentials, err := s.credentialRepo.GetByWorkspaceIDAndProvider(workspaceUUID, vm.Provider)
	if err != nil {
		s.logger.Warn("Failed to get credentials for provider, falling back to estimated costs",
			zap.String("provider", vm.Provider),
			zap.Error(err))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	if len(credentials) == 0 {
		s.logger.Warn("No credentials found for provider, falling back to estimated costs",
			zap.String("provider", vm.Provider),
			zap.String("workspace_id", vm.WorkspaceID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Use the first active credential
	var credential *domain.Credential
	for _, cred := range credentials {
		if cred.IsActive {
			credential = cred
			break
		}
	}

	if credential == nil && len(credentials) > 0 {
		credential = credentials[0]
	}

	// Get actual costs from cloud provider API
	switch vm.Provider {
	case ProviderAWS:
		return s.getAWSCosts(ctx, credential, vm, startDate, endDate)
	case ProviderGCP:
		return s.getGCPCosts(ctx, credential, vm, startDate, endDate)
	default:
		// For other providers, use estimated costs
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
func (s *Service) aggregateCosts(costs []CostData, startDate, endDate time.Time, period string) *CostSummary {
	summary := &CostSummary{
		Currency:   CurrencyUSD,
		Period:     period,
		StartDate:  startDate,
		EndDate:    endDate,
		ByProvider: make(map[string]float64),
	}

	for _, cost := range costs {
		summary.TotalCost += cost.Amount
		summary.ByProvider[cost.Provider] += cost.Amount
	}

	return summary
}

// generatePredictions: 과거 비용 데이터를 기반으로 예측을 생성합니다
func (s *Service) generatePredictions(historicalCosts []CostData, days int) []CostPrediction {
	// Simple linear regression for prediction
	// In a real implementation, you might use more sophisticated ML models

	// Calculate daily totals
	dailyTotals := make(map[string]float64)
	for _, cost := range historicalCosts {
		dateStr := cost.Date.Format(DateFormatISO)
		dailyTotals[dateStr] += cost.Amount
	}

	// Convert to sorted slice
	var dates []time.Time
	var values []float64
	for dateStr, total := range dailyTotals {
		date, _ := time.Parse(DateFormatISO, dateStr)
		dates = append(dates, date)
		values = append(values, total)
	}

	// Sort by date
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})
	sort.Slice(values, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	if len(values) < MinDataPointsForPrediction {
		// Not enough data for prediction
		return []CostPrediction{}
	}

	// Calculate linear regression
	n := len(values)
	var sumX, sumY, sumXY, sumXX float64

	for i := 0; i < n; i++ {
		x := float64(i)
		y := values[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope and intercept
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / float64(n)

	// Generate predictions
	var predictions []CostPrediction
	lastDate := dates[len(dates)-1]

	for i := 1; i <= days; i++ {
		predictedDate := lastDate.AddDate(0, 0, i)
		predictedValue := slope*float64(n+i-1) + intercept

		// Calculate confidence based on historical variance
		variance := s.calculateVariance(values)
		confidence := math.Max(MinCostPredictionConfidence, 1.0-(variance/CostVarianceNormalizationBase))

		// Calculate bounds (simple approach)
		bound := predictedValue * DefaultCostPredictionMargin

		predictions = append(predictions, CostPrediction{
			Date:       predictedDate,
			Predicted:  math.Max(0, predictedValue),
			Confidence: confidence,
			LowerBound: math.Max(0, predictedValue-bound),
			UpperBound: predictedValue + bound,
		})
	}

	return predictions
}

// calculateVariance: 값들의 분산을 계산합니다
func (s *Service) calculateVariance(values []float64) float64 {
	if len(values) < MinDataPointsForVariance {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values) - 1)

	return variance
}

// aggregateDailyCosts: 비용을 일별로 집계합니다
func (s *Service) aggregateDailyCosts(costs []CostData) []DailyCostData {
	dailyMap := make(map[string]float64)
	for _, cost := range costs {
		dateStr := cost.Date.Format(DateFormatISO)
		dailyMap[dateStr] += cost.Amount
	}

	var dailyCosts []DailyCostData
	for dateStr, amount := range dailyMap {
		date, _ := time.Parse(DateFormatISO, dateStr)
		dailyCosts = append(dailyCosts, DailyCostData{
			Date:   date,
			Amount: amount,
		})
	}

	// Sort by date
	sort.Slice(dailyCosts, func(i, j int) bool {
		return dailyCosts[i].Date.Before(dailyCosts[j].Date)
	})

	return dailyCosts
}

// calculateTrendMetrics: 추세 방향과 변화율을 계산합니다
func (s *Service) calculateTrendMetrics(dailyCosts []DailyCostData) (string, float64) {
	if len(dailyCosts) < MinDataPointsForPrediction {
		return TrendDirectionStable, 0.0
	}

	firstHalf := 0.0
	secondHalf := 0.0
	midPoint := len(dailyCosts) / 2

	for i := 0; i < midPoint; i++ {
		firstHalf += dailyCosts[i].Amount
	}

	for i := midPoint; i < len(dailyCosts); i++ {
		secondHalf += dailyCosts[i].Amount
	}

	if midPoint == 0 {
		return TrendDirectionStable, 0.0
	}

	firstAvg := firstHalf / float64(midPoint)
	secondAvg := secondHalf / float64(len(dailyCosts)-midPoint)

	if firstAvg == 0 {
		return TrendDirectionStable, 0.0
	}

	percentageChange := ((secondAvg - firstAvg) / firstAvg) * serviceconstants.PercentageBase

	trendDirection := TrendDirectionStable
	if percentageChange > TrendPercentageThreshold {
		trendDirection = TrendDirectionIncreasing
	} else if percentageChange < -TrendPercentageThreshold {
		trendDirection = TrendDirectionDecreasing
	}

	return trendDirection, percentageChange
}

// aggregateCostBreakdown: 차원별로 비용을 집계합니다 (서비스, 프로바이더, 리전)
func (s *Service) aggregateCostBreakdown(costs []CostData, dimension string) *CostBreakdown {
	totalCost := 0.0
	for _, cost := range costs {
		totalCost += cost.Amount
	}

	breakdown := make(CostBreakdown)

	switch dimension {
	case "service":
		serviceMap := make(map[string]float64)
		for _, cost := range costs {
			serviceMap[cost.Service] += cost.Amount
		}

		for service, amount := range serviceMap {
			percentage := (amount / totalCost) * 100
			breakdown[service] = CategoryBreakdown{
				Cost:       amount,
				Percentage: percentage,
			}
		}

	case "provider":
		providerMap := make(map[string]float64)
		providerServiceMap := make(map[string]map[string]float64)

		for _, cost := range costs {
			providerMap[cost.Provider] += cost.Amount
			if providerServiceMap[cost.Provider] == nil {
				providerServiceMap[cost.Provider] = make(map[string]float64)
			}
			providerServiceMap[cost.Provider][cost.Service] += cost.Amount
		}

		for provider, amount := range providerMap {
			percentage := (amount / totalCost) * 100
			breakdown[provider] = CategoryBreakdown{
				Cost:       amount,
				Percentage: percentage,
				Services:   providerServiceMap[provider],
			}
		}

	case "region":
		regionMap := make(map[string]float64)
		for _, cost := range costs {
			regionMap[cost.Region] += cost.Amount
		}

		for region, amount := range regionMap {
			percentage := (amount / totalCost) * 100
			breakdown[region] = CategoryBreakdown{
				Cost:       amount,
				Percentage: percentage,
			}
		}

	default:
		// Default to service breakdown
		serviceMap := make(map[string]float64)
		for _, cost := range costs {
			serviceMap[cost.Service] += cost.Amount
		}

		for service, amount := range serviceMap {
			percentage := (amount / totalCost) * 100
			breakdown[service] = CategoryBreakdown{
				Cost:       amount,
				Percentage: percentage,
			}
		}
	}

	return &breakdown
}

// getAWSCosts: AWS Cost Explorer API에서 실제 비용을 조회합니다
func (s *Service) getAWSCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	ceClient, _, err := s.getAWSCostExplorerClient(ctx, credential, vm.Region)
	if err != nil {
		return nil, err
	}

	// Prepare filters - filter by instance ID if available
	var filter *types.Expression
	if vm.InstanceID != "" {
		filter = &types.Expression{
			Dimensions: &types.DimensionValues{
				Key:    types.DimensionResourceId,
				Values: []string{vm.InstanceID},
			},
		}
	}

	result, err := s.queryAWSCostExplorer(ctx, ceClient, startDate, endDate, filter, nil)
	if err != nil {
		s.logger.Warn("Failed to get AWS costs from Cost Explorer API, falling back to estimated costs",
			zap.Error(err))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	costs := s.parseAWSCostExplorerResults(result, vm.InstanceID, vm.WorkspaceID, ResourceTypeVM)

	// If no costs found, fall back to estimated costs
	if len(costs) == 0 {
		s.logger.Warn("No AWS costs found from Cost Explorer API, falling back to estimated costs")
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	return costs, nil
}

// getAWSCostExplorerClient: 자격증명으로부터 AWS Cost Explorer 클라이언트를 생성합니다
func (s *Service) getAWSCostExplorerClient(ctx context.Context, credential *domain.Credential, defaultRegion string) (*costexplorer.Client, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	region := defaultRegion
	if r, ok := credData["region"].(string); ok && r != "" {
		region = r
	}
	if region == "" {
		region = AWSDefaultRegion
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	ceClient := costexplorer.NewFromConfig(cfg)
	return ceClient, region, nil
}

// queryAWSCostExplorer: 주어진 파라미터로 AWS Cost Explorer API를 조회합니다
func (s *Service) queryAWSCostExplorer(ctx context.Context, ceClient *costexplorer.Client, startDate, endDate time.Time, filter *types.Expression, groupBy []types.GroupDefinition) (*costexplorer.GetCostAndUsageOutput, error) {
	if groupBy == nil {
		groupBy = []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("REGION"),
			},
		}
	}

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate.Format(DateFormatISO)),
			End:   aws.String(endDate.Format(DateFormatISO)),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"BlendedCost", "UnblendedCost"},
		GroupBy:     groupBy,
	}

	if filter != nil {
		input.Filter = filter
	}

	return ceClient.GetCostAndUsage(ctx, input)
}

// parseAWSCostExplorerResults: AWS Cost Explorer API 결과를 CostData 슬라이스로 파싱합니다
func (s *Service) parseAWSCostExplorerResults(result *costexplorer.GetCostAndUsageOutput, resourceID, workspaceID, resourceType string) []CostData {
	var costs []CostData

	for _, resultByTime := range result.ResultsByTime {
		dateStr := aws.ToString(resultByTime.TimePeriod.Start)
		date, err := time.Parse(DateFormatISO, dateStr)
		if err != nil {
			s.logger.Warn("Failed to parse date",
				zap.String("date", dateStr),
				zap.Error(err))
			continue
		}

		for _, group := range resultByTime.Groups {
			var service, region string
			for i, key := range group.Keys {
				if i == GroupKeyIndexService {
					service = key
				} else if i == GroupKeyIndexRegion {
					region = key
				}
			}

			var amount float64
			var currency string
			if blendedCost, ok := group.Metrics["BlendedCost"]; ok {
				amountStr := aws.ToString(blendedCost.Amount)
				var parseErr error
				amount, parseErr = parseFloat(amountStr)
				if parseErr != nil {
					s.logger.Warn("Failed to parse cost amount",
						zap.String("amount", amountStr),
						zap.Error(parseErr))
					continue
				}
				currency = aws.ToString(blendedCost.Unit)
			}

			if amount > 0 {
				costs = append(costs, CostData{
					Date:         date,
					Amount:       amount,
					Currency:     currency,
					Service:      service,
					ResourceID:   resourceID,
					ResourceType: resourceType,
					Provider:     ProviderAWS,
					Region:       region,
					WorkspaceID:  workspaceID,
				})
			}
		}
	}

	return costs
}

// getGCPCosts: GCP Cloud Billing API에서 실제 비용을 조회합니다
func (s *Service) getGCPCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Get project ID
	projectID, ok := credData["project_id"].(string)
	if !ok || projectID == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "project_id not found in credential", 400)
	}

	// Create service account key from credential data
	serviceAccountKey := map[string]interface{}{
		"type":                        credData["type"],
		"project_id":                  credData["project_id"],
		"private_key_id":              credData["private_key_id"],
		"private_key":                 credData["private_key"],
		"client_email":                credData["client_email"],
		"client_id":                   credData["client_id"],
		"auth_uri":                    credData["auth_uri"],
		"token_uri":                   credData["token_uri"],
		"auth_provider_x509_cert_url": credData["auth_provider_x509_cert_url"],
		"client_x509_cert_url":        credData["client_x509_cert_url"],
	}

	// Convert to JSON
	keyBytes, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal service account key: %v", err), 500)
	}

	// Create billing client
	billingClient, err := billingv1.NewCloudBillingClient(ctx, option.WithCredentialsJSON(keyBytes))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create billing client: %v", err), 502)
	}
	defer billingClient.Close()

	// Get billing account for the project
	projectName := fmt.Sprintf("projects/%s", projectID)
	req := &billingpb.GetProjectBillingInfoRequest{
		Name: projectName,
	}

	projectBillingInfo, err := billingClient.GetProjectBillingInfo(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get GCP billing info, falling back to estimated costs",
			zap.Error(err))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Check if billing is enabled
	if !projectBillingInfo.BillingEnabled {
		s.logger.Warn("Billing not enabled for project, falling back to estimated costs",
			zap.String("project_id", projectID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Get billing account name
	billingAccountName := projectBillingInfo.BillingAccountName
	if billingAccountName == "" {
		s.logger.Warn("No billing account found for project, falling back to estimated costs",
			zap.String("project_id", projectID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Note: GCP Cloud Billing API doesn't provide direct cost queries like AWS Cost Explorer
	// For detailed cost data, we would need to use BigQuery Billing Export or Cloud Billing Budget API
	// Since that requires additional setup, we'll use estimated costs based on VM specifications
	// but mark it as coming from GCP pricing
	s.logger.Info("GCP billing account found, using estimated costs based on VM specifications",
		zap.String("billing_account", billingAccountName))

	return s.calculateEstimatedCosts(vm, startDate, endDate)
}

// parseFloat: 다양한 형식을 처리하여 문자열을 float64로 파싱합니다
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// calculateKubernetesCosts: 워크스페이스의 Kubernetes 클러스터 비용을 계산합니다
// 반환값: 비용, 경고, 에러
func (s *Service) calculateKubernetesCosts(ctx context.Context, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("invalid workspace ID: %v", err), 400)
	}

	// Get all credentials for the workspace
	allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get credentials: %v", err), 500)
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
func (s *Service) getAWSKubernetesCosts(ctx context.Context, credential *domain.Credential, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	ceClient, _, err := s.getAWSCostExplorerClient(ctx, credential, AWSDefaultRegion)
	if err != nil {
		return nil, nil, err
	}

	// Filter by EKS service
	filter := &types.Expression{
		Dimensions: &types.DimensionValues{
			Key: types.DimensionService,
			Values: []string{
				"Amazon Elastic Container Service for Kubernetes", // EKS service name
			},
		},
	}

	groupBy := []types.GroupDefinition{
		{
			Type: types.GroupDefinitionTypeDimension,
			Key:  aws.String("SERVICE"),
		},
		{
			Type: types.GroupDefinitionTypeDimension,
			Key:  aws.String("REGION"),
		},
	}

	result, err := s.queryAWSCostExplorer(ctx, ceClient, startDate, endDate, filter, groupBy)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get AWS Kubernetes costs: %v", err), 502)
	}

	costs := s.parseAWSCostExplorerResults(result, "", workspaceID, ResourceTypeCluster)

	// If node groups are requested, also get EC2 costs for EKS nodes
	var warnings []CostWarning
	if includeNodeGroups {
		// Node groups are EC2 instances, so they're tracked separately
		// We can get them by filtering EC2 costs for instances with EKS tags or by service
		// For now, we'll skip node group specific costs as they're part of EC2 costs
		logger.Info("Node group costs are included in EC2 service costs")
	}

	return costs, warnings, nil
}

// getGCPKubernetesCosts: GCP Cloud Billing API에서 GKE 비용을 조회합니다
// 반환값: 비용, 경고, 에러
func (s *Service) getGCPKubernetesCosts(ctx context.Context, credential *domain.Credential, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	projectID, ok := credData["project_id"].(string)
	if !ok || projectID == "" {
		return nil, nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "project_id not found in credential", 400)
	}

	// Create service account key from credential data
	serviceAccountKey := map[string]interface{}{
		"type":                        credData["type"],
		"project_id":                  credData["project_id"],
		"private_key_id":              credData["private_key_id"],
		"private_key":                 credData["private_key"],
		"client_email":                credData["client_email"],
		"client_id":                   credData["client_id"],
		"auth_uri":                    credData["auth_uri"],
		"token_uri":                   credData["token_uri"],
		"auth_provider_x509_cert_url": credData["auth_provider_x509_cert_url"],
		"client_x509_cert_url":        credData["client_x509_cert_url"],
	}

	keyBytes, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal service account key: %v", err), 500)
	}

	// Create billing client
	billingClient, err := billingv1.NewCloudBillingClient(ctx, option.WithCredentialsJSON(keyBytes))
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create billing client: %v", err), 502)
	}
	defer billingClient.Close()

	// Get billing account for the project
	projectName := fmt.Sprintf("projects/%s", projectID)
	projectInfo, err := billingClient.GetProjectBillingInfo(ctx, &billingpb.GetProjectBillingInfoRequest{
		Name: projectName,
	})
	if err != nil {
		// Check if it's a permission/API disabled error
		errMsg := err.Error()
		var warnings []CostWarning
		if strings.Contains(errMsg, "SERVICE_DISABLED") || strings.Contains(errMsg, "not been used") || strings.Contains(errMsg, "disabled") {
			warnings = append(warnings, CostWarning{
				Code:         "API_NOT_ENABLED",
				Message:      "GCP Cloud Billing API is not enabled. Please enable it in the GCP Console.",
				Provider:     ProviderGCP,
				ResourceType: ResourceTypeCluster,
			})
		} else if strings.Contains(errMsg, "PermissionDenied") || strings.Contains(errMsg, "permission") {
			warnings = append(warnings, CostWarning{
				Code:         "API_PERMISSION_DENIED",
				Message:      "GCP service account does not have permission to access Cloud Billing API.",
				Provider:     ProviderGCP,
				ResourceType: ResourceTypeCluster,
			})
		}
		return nil, warnings, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get project billing info: %v", err), 502)
	}

	if projectInfo.BillingAccountName == "" {
		return nil, nil, domain.NewDomainError(domain.ErrCodeNotFound, "no billing account associated with project", 404)
	}

	// For GKE, costs are tracked under Container Service
	// Since GCP Billing API is complex, we'll use a simplified approach
	// and estimate based on cluster count (similar to VM estimation)
	// In production, you would use Cloud Billing Export to BigQuery for detailed costs
	s.logger.Info("GKE costs retrieved",
		zap.String("project_id", projectID),
		zap.String("billing_account", projectInfo.BillingAccountName))

	// Return empty costs for now - GCP Billing API requires more complex setup
	// In production, use Cloud Billing Export API or BigQuery
	var costs []CostData
	var warnings []CostWarning
	warnings = append(warnings, CostWarning{
		Code:         "GKE_COST_NOT_IMPLEMENTED",
		Message:      "GKE cost calculation requires Cloud Billing Export setup. Currently returning empty costs.",
		Provider:     ProviderGCP,
		ResourceType: ResourceTypeCluster,
	})
	logger.Info("GKE cost calculation requires Cloud Billing Export setup - returning empty costs")
	return costs, warnings, nil
}

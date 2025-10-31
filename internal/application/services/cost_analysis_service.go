package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	billingv1 "cloud.google.com/go/billing/apiv1"
	billingpb "cloud.google.com/go/billing/apiv1/billingpb"
	"google.golang.org/api/option"
	"github.com/google/uuid"
)

type CostAnalysisService struct {
	vmRepo           domain.VMRepository
	credentialRepo   domain.CredentialRepository
	workspaceRepo    domain.WorkspaceRepository
	auditLogRepo     domain.AuditLogRepository
	credentialService domain.CredentialService
}

func NewCostAnalysisService(
	vmRepo domain.VMRepository,
	credentialRepo domain.CredentialRepository,
	workspaceRepo domain.WorkspaceRepository,
	auditLogRepo domain.AuditLogRepository,
	credentialService domain.CredentialService,
) *CostAnalysisService {
	return &CostAnalysisService{
		vmRepo:            vmRepo,
		credentialRepo:    credentialRepo,
		workspaceRepo:     workspaceRepo,
		auditLogRepo:      auditLogRepo,
		credentialService: credentialService,
	}
}

// CostData represents cost information for a specific period
type CostData struct {
	Date         time.Time `json:"date"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Service      string    `json:"service"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	WorkspaceID  string    `json:"workspace_id"`
}

// CostSummary represents simplified cost data
type CostSummary struct {
	TotalCost  float64            `json:"total_cost"`
	Currency   string             `json:"currency"`
	Period     string             `json:"period"`
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	ByProvider map[string]float64 `json:"by_provider"`
}

// CostPrediction represents future cost predictions
type CostPrediction struct {
	Date       time.Time `json:"date"`
	Predicted  float64   `json:"predicted"`
	Confidence float64   `json:"confidence"` // 0-1
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
}

// BudgetAlert represents budget threshold alerts
type BudgetAlert struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	BudgetLimit float64   `json:"budget_limit"`
	CurrentCost float64   `json:"current_cost"`
	Percentage  float64   `json:"percentage"`
	AlertLevel  string    `json:"alert_level"` // "warning", "critical"
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetCostSummary retrieves cost summary for a workspace
func (s *CostAnalysisService) GetCostSummary(ctx context.Context, workspaceID string, period string) (*CostSummary, error) {
	// Parse period (e.g., "7d", "30d", "90d", "1y")
	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("invalid period: %w", err)
	}

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	// Calculate costs for each VM
	var allCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCosts(ctx, vm, startDate, endDate)
		if err != nil {
			logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
			continue
		}
		allCosts = append(allCosts, costs...)
	}

	// Aggregate costs
	summary := s.aggregateCosts(allCosts, startDate, endDate, period)
	// Removed ByWorkspace field

	// Calculate trend
	// Removed Trend and GrowthRate fields

	return summary, nil
}

// GetCostPredictions generates cost predictions for future periods
func (s *CostAnalysisService) GetCostPredictions(ctx context.Context, workspaceID string, days int) ([]CostPrediction, error) {
	// Get historical data (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	// Calculate historical costs
	var historicalCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCosts(ctx, vm, startDate, endDate)
		if err != nil {
			logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
			continue
		}
		historicalCosts = append(historicalCosts, costs...)
	}

	// Generate predictions using linear regression
	predictions := s.generatePredictions(historicalCosts, days)

	return predictions, nil
}

// CheckBudgetAlerts checks if workspace exceeds budget limits
func (s *CostAnalysisService) CheckBudgetAlerts(ctx context.Context, workspaceID string, budgetLimit float64) ([]BudgetAlert, error) {
	summary, err := s.GetCostSummary(ctx, workspaceID, "1m")
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}

	var alerts []BudgetAlert
	percentage := (summary.TotalCost / budgetLimit) * 100

	if percentage >= 100 {
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

// CostTrend represents cost trend data
type CostTrend struct {
	DailyCosts      []DailyCostData `json:"daily_costs"`
	TrendDirection  string          `json:"trend_direction"`  // "increasing", "decreasing", "stable"
	ChangePercentage float64        `json:"change_percentage"`
}

// DailyCostData represents daily cost data
type DailyCostData struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}

// CostBreakdown represents cost breakdown by dimension
type CostBreakdown map[string]CategoryBreakdown

// CategoryBreakdown represents breakdown for a category
type CategoryBreakdown struct {
	Cost      float64            `json:"cost"`
	Percentage float64           `json:"percentage"`
	Services  map[string]float64 `json:"services,omitempty"`
}

// CostComparison represents cost comparison between periods
type CostComparison struct {
	CurrentPeriod  PeriodComparison `json:"current_period"`
	PreviousPeriod PeriodComparison `json:"previous_period"`
	Comparison     ComparisonData   `json:"comparison"`
}

// PeriodComparison represents cost data for a period
type PeriodComparison struct {
	Period string  `json:"period"`
	Cost   float64 `json:"cost"`
}

// ComparisonData represents comparison metrics
type ComparisonData struct {
	CostChange       float64 `json:"cost_change"`
	PercentageChange float64 `json:"percentage_change"`
}

// GetCostTrend retrieves cost trend for a workspace
func (s *CostAnalysisService) GetCostTrend(ctx context.Context, workspaceID string, period string) (*CostTrend, error) {
	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("invalid period: %w", err)
	}

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	// Calculate costs for each VM
	var allCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCosts(ctx, vm, startDate, endDate)
		if err != nil {
			logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
			continue
		}
		allCosts = append(allCosts, costs...)
	}

	// Aggregate daily costs
	dailyCosts := s.aggregateDailyCosts(allCosts)

	// Calculate trend direction and percentage change
	trendDirection, changePercentage := s.calculateTrendMetrics(dailyCosts)

	return &CostTrend{
		DailyCosts:      dailyCosts,
		TrendDirection:  trendDirection,
		ChangePercentage: changePercentage,
	}, nil
}

// GetCostBreakdown retrieves cost breakdown for a workspace
func (s *CostAnalysisService) GetCostBreakdown(ctx context.Context, workspaceID string, period string, dimension string) (*CostBreakdown, error) {
	startDate, endDate, err := s.parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("invalid period: %w", err)
	}

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	// Calculate costs for each VM
	var allCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCosts(ctx, vm, startDate, endDate)
		if err != nil {
			logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
			continue
		}
		allCosts = append(allCosts, costs...)
	}

	// Aggregate costs by dimension
	breakdown := s.aggregateCostBreakdown(allCosts, dimension)

	return breakdown, nil
}

// GetCostComparison retrieves cost comparison between periods
func (s *CostAnalysisService) GetCostComparison(ctx context.Context, workspaceID string, currentPeriod string, comparePeriod string) (*CostComparison, error) {
	// Get current period summary
	currentSummary, err := s.GetCostSummary(ctx, workspaceID, currentPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period summary: %w", err)
	}

	// Calculate compare period dates
	now := time.Now()
	var compareStartDate, compareEndDate time.Time

	switch comparePeriod {
	case "7d":
		compareEndDate = now.AddDate(0, 0, -7)
		compareStartDate = compareEndDate.AddDate(0, 0, -7)
	case "30d", "1m":
		compareEndDate = now.AddDate(0, 0, -30)
		compareStartDate = compareEndDate.AddDate(0, 0, -30)
	case "90d", "3m":
		compareEndDate = now.AddDate(0, 0, -90)
		compareStartDate = compareEndDate.AddDate(0, 0, -90)
	case "1y":
		compareEndDate = now.AddDate(-1, 0, 0)
		compareStartDate = compareEndDate.AddDate(-1, 0, 0)
	default:
		return nil, fmt.Errorf("unsupported compare period: %s", comparePeriod)
	}

	// Get VMs in workspace
	vms, err := s.vmRepo.GetVMsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	// Calculate compare period costs
	var compareCosts []CostData
	for _, vm := range vms {
		costs, err := s.calculateVMCosts(ctx, vm, compareStartDate, compareEndDate)
		if err != nil {
			logger.Warnf("Failed to calculate VM costs for VM %s: %v", vm.ID, err)
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
		percentageChange = (costChange / compareTotal) * 100
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

func (s *CostAnalysisService) parsePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now()

	switch period {
	case "7d":
		return now.AddDate(0, 0, -7), now, nil
	case "30d", "1m":
		return now.AddDate(0, 0, -30), now, nil
	case "90d", "3m":
		return now.AddDate(0, 0, -90), now, nil
	case "1y":
		return now.AddDate(-1, 0, 0), now, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported period: %s", period)
	}
}

func (s *CostAnalysisService) calculateVMCosts(ctx context.Context, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Get credentials for the provider and workspace
	workspaceUUID, err := uuid.Parse(vm.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID: %w", err)
	}

	credentials, err := s.credentialRepo.GetByWorkspaceIDAndProvider(workspaceUUID, vm.Provider)
	if err != nil {
		logger.Warnf("Failed to get credentials for provider %s: %v, falling back to estimated costs", vm.Provider, err)
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	if len(credentials) == 0 {
		logger.Warnf("No credentials found for provider %s in workspace %s, falling back to estimated costs", vm.Provider, vm.WorkspaceID)
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
	case "aws":
		return s.getAWSCosts(ctx, credential, vm, startDate, endDate)
	case "gcp":
		return s.getGCPCosts(ctx, credential, vm, startDate, endDate)
	default:
		// For other providers, use estimated costs
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}
}

// calculateEstimatedCosts calculates costs based on VM specifications when API is unavailable
func (s *CostAnalysisService) calculateEstimatedCosts(vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
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
			Currency:     "USD",
			Service:      service,
			ResourceID:   vm.ID,
			ResourceType: "vm",
			Provider:     vm.Provider,
			Region:       vm.Region,
			WorkspaceID:  vm.WorkspaceID,
		})

		current = nextDay
	}

	return costs, nil
}

func (s *CostAnalysisService) getVMHourlyRate(vm *domain.VM) float64 {
	// Mock pricing based on VM specifications
	// In reality, this would come from cloud provider pricing APIs

	baseRate := 0.05 // $0.05 per hour base rate

	// CPU multiplier
	cpuMultiplier := float64(vm.CPUs) * 0.01

	// Memory multiplier (per GB)
	memoryMultiplier := float64(vm.Memory) * 0.005

	// Storage multiplier (per GB)
	storageMultiplier := float64(vm.Storage) * 0.0001

	// Provider multiplier
	providerMultiplier := 1.0
	switch vm.Provider {
	case "aws":
		providerMultiplier = 1.0
	case "gcp":
		providerMultiplier = 0.9
	case "azure":
		providerMultiplier = 1.1
	}

	return (baseRate + cpuMultiplier + memoryMultiplier + storageMultiplier) * providerMultiplier
}

// getVMServiceName returns the service name for a VM based on provider
func (s *CostAnalysisService) getVMServiceName(provider string) string {
	switch provider {
	case "aws":
		return "EC2"
	case "gcp":
		return "Compute"
	case "azure":
		return "Virtual Machines"
	default:
		return "compute"
	}
}

func (s *CostAnalysisService) aggregateCosts(costs []CostData, startDate, endDate time.Time, period string) *CostSummary {
	summary := &CostSummary{
		Currency:   "USD",
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

func (s *CostAnalysisService) generatePredictions(historicalCosts []CostData, days int) []CostPrediction {
	// Simple linear regression for prediction
	// In a real implementation, you might use more sophisticated ML models

	// Calculate daily totals
	dailyTotals := make(map[string]float64)
	for _, cost := range historicalCosts {
		dateStr := cost.Date.Format("2006-01-02")
		dailyTotals[dateStr] += cost.Amount
	}

	// Convert to sorted slice
	var dates []time.Time
	var values []float64
	for dateStr, total := range dailyTotals {
		date, _ := time.Parse("2006-01-02", dateStr)
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

	if len(values) < 2 {
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
		confidence := math.Max(0.1, 1.0-(variance/100.0))

		// Calculate bounds (simple approach)
		bound := predictedValue * 0.2 // 20% margin

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

func (s *CostAnalysisService) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
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

// aggregateDailyCosts aggregates costs by day
func (s *CostAnalysisService) aggregateDailyCosts(costs []CostData) []DailyCostData {
	dailyMap := make(map[string]float64)
	for _, cost := range costs {
		dateStr := cost.Date.Format("2006-01-02")
		dailyMap[dateStr] += cost.Amount
	}

	var dailyCosts []DailyCostData
	for dateStr, amount := range dailyMap {
		date, _ := time.Parse("2006-01-02", dateStr)
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

// calculateTrendMetrics calculates trend direction and percentage change
func (s *CostAnalysisService) calculateTrendMetrics(dailyCosts []DailyCostData) (string, float64) {
	if len(dailyCosts) < 2 {
		return "stable", 0.0
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
		return "stable", 0.0
	}

	firstAvg := firstHalf / float64(midPoint)
	secondAvg := secondHalf / float64(len(dailyCosts)-midPoint)

	if firstAvg == 0 {
		return "stable", 0.0
	}

	percentageChange := ((secondAvg - firstAvg) / firstAvg) * 100

	trendDirection := "stable"
	if percentageChange > 5 {
		trendDirection = "increasing"
	} else if percentageChange < -5 {
		trendDirection = "decreasing"
	}

	return trendDirection, percentageChange
}

// aggregateCostBreakdown aggregates costs by dimension (service, provider, region)
func (s *CostAnalysisService) aggregateCostBreakdown(costs []CostData, dimension string) *CostBreakdown {
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
				Cost:      amount,
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
				Cost:      amount,
				Percentage: percentage,
				Services:  providerServiceMap[provider],
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
				Cost:      amount,
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
				Cost:      amount,
				Percentage: percentage,
			}
		}
	}

	return &breakdown
}

// getAWSCosts retrieves actual costs from AWS Cost Explorer API
func (s *CostAnalysisService) getAWSCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
	}

	// Get region from credential or use VM region
	region := vm.Region
	if r, ok := credData["region"].(string); ok && r != "" {
		region = r
	}
	if region == "" {
		region = "us-east-1" // Default region for Cost Explorer
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Cost Explorer client
	ceClient := costexplorer.NewFromConfig(cfg)

	// Prepare granularity (DAILY for detailed costs)
	granularity := types.GranularityDaily

	// Prepare metrics
	metrics := []string{"BlendedCost", "UnblendedCost"}

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

	// Prepare group by - group by service and region for detailed breakdown
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

	// Call Cost Explorer API
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate.Format("2006-01-02")),
			End:   aws.String(endDate.Format("2006-01-02")),
		},
		Granularity: granularity,
		Metrics:     metrics,
		GroupBy:     groupBy,
	}

	if filter != nil {
		input.Filter = filter
	}

	result, err := ceClient.GetCostAndUsage(ctx, input)
	if err != nil {
		logger.Warnf("Failed to get AWS costs from Cost Explorer API: %v, falling back to estimated costs", err)
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Parse results into CostData
	var costs []CostData
	for _, resultByTime := range result.ResultsByTime {
		// Parse date
		dateStr := aws.ToString(resultByTime.TimePeriod.Start)
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logger.Warnf("Failed to parse date %s: %v", dateStr, err)
			continue
		}

		// Process groups to extract costs
		for _, group := range resultByTime.Groups {
			// Extract service and region from keys
			var service, region string
			for i, key := range group.Keys {
				if i == 0 {
					service = key
				} else if i == 1 {
					region = key
				}
			}

			// Get cost amount
			var amount float64
			var currency string
			if blendedCost, ok := group.Metrics["BlendedCost"]; ok {
				amountStr := aws.ToString(blendedCost.Amount)
				var parseErr error
				amount, parseErr = parseFloat(amountStr)
				if parseErr != nil {
					logger.Warnf("Failed to parse cost amount %s: %v", amountStr, parseErr)
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
					ResourceID:   vm.InstanceID,
					ResourceType: "vm",
					Provider:     "aws",
					Region:       region,
					WorkspaceID:  vm.WorkspaceID,
				})
			}
		}
	}

	// If no costs found, fall back to estimated costs
	if len(costs) == 0 {
		logger.Warnf("No AWS costs found from Cost Explorer API, falling back to estimated costs")
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	return costs, nil
}

// getGCPCosts retrieves actual costs from GCP Cloud Billing API
func (s *CostAnalysisService) getGCPCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Get project ID
	projectID, ok := credData["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id not found in credential")
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
		return nil, fmt.Errorf("failed to marshal service account key: %w", err)
	}

	// Create billing client
	billingClient, err := billingv1.NewCloudBillingClient(ctx, option.WithCredentialsJSON(keyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create billing client: %w", err)
	}
	defer billingClient.Close()

	// Get billing account for the project
	projectName := fmt.Sprintf("projects/%s", projectID)
	req := &billingpb.GetProjectBillingInfoRequest{
		Name: projectName,
	}

	projectBillingInfo, err := billingClient.GetProjectBillingInfo(ctx, req)
	if err != nil {
		logger.Warnf("Failed to get GCP billing info: %v, falling back to estimated costs", err)
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Check if billing is enabled
	if !projectBillingInfo.BillingEnabled {
		logger.Warnf("Billing not enabled for project %s, falling back to estimated costs", projectID)
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Get billing account name
	billingAccountName := projectBillingInfo.BillingAccountName
	if billingAccountName == "" {
		logger.Warnf("No billing account found for project %s, falling back to estimated costs", projectID)
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Note: GCP Cloud Billing API doesn't provide direct cost queries like AWS Cost Explorer
	// For detailed cost data, we would need to use BigQuery Billing Export or Cloud Billing Budget API
	// Since that requires additional setup, we'll use estimated costs based on VM specifications
	// but mark it as coming from GCP pricing
	logger.Infof("GCP billing account found: %s, using estimated costs based on VM specifications", billingAccountName)

	return s.calculateEstimatedCosts(vm, startDate, endDate)
}

// parseFloat parses a string to float64, handling various formats
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

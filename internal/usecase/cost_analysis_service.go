package usecase

import (
	"skyclust/internal/domain"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

type CostAnalysisService struct {
	logger         *zap.Logger
	vmRepo         domain.VMRepository
	credentialRepo domain.CredentialRepository
	workspaceRepo  domain.WorkspaceRepository
	auditLogRepo   domain.AuditLogRepository
}

func NewCostAnalysisService(
	logger *zap.Logger,
	vmRepo domain.VMRepository,
	credentialRepo domain.CredentialRepository,
	workspaceRepo domain.WorkspaceRepository,
	auditLogRepo domain.AuditLogRepository,
) *CostAnalysisService {
	return &CostAnalysisService{
		logger:         logger,
		vmRepo:         vmRepo,
		credentialRepo: credentialRepo,
		workspaceRepo:  workspaceRepo,
		auditLogRepo:   auditLogRepo,
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

// CostSummary represents aggregated cost data
type CostSummary struct {
	TotalCost   float64            `json:"total_cost"`
	Currency    string             `json:"currency"`
	Period      string             `json:"period"`
	StartDate   time.Time          `json:"start_date"`
	EndDate     time.Time          `json:"end_date"`
	ByService   map[string]float64 `json:"by_service"`
	ByProvider  map[string]float64 `json:"by_provider"`
	ByRegion    map[string]float64 `json:"by_region"`
	ByWorkspace map[string]float64 `json:"by_workspace"`
	DailyCosts  []CostData         `json:"daily_costs"`
	Trend       string             `json:"trend"` // "increasing", "decreasing", "stable"
	GrowthRate  float64            `json:"growth_rate"`
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
			s.logger.Warn("Failed to calculate VM costs", zap.String("vm_id", vm.ID), zap.Error(err))
			continue
		}
		allCosts = append(allCosts, costs...)
	}

	// Aggregate costs
	summary := s.aggregateCosts(allCosts, startDate, endDate, period)
	summary.ByWorkspace[workspaceID] = summary.TotalCost

	// Calculate trend
	summary.Trend, summary.GrowthRate = s.calculateTrend(allCosts)

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
			s.logger.Warn("Failed to calculate VM costs", zap.String("vm_id", vm.ID), zap.Error(err))
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
	// This is a simplified calculation
	// In a real implementation, you would integrate with cloud provider billing APIs

	var costs []CostData
	current := startDate

	// Mock cost calculation based on VM specifications
	hourlyRate := s.getVMHourlyRate(vm)

	for current.Before(endDate) {
		nextDay := current.AddDate(0, 0, 1)
		if nextDay.After(endDate) {
			nextDay = endDate
		}

		hours := nextDay.Sub(current).Hours()
		dailyCost := hourlyRate * hours

		costs = append(costs, CostData{
			Date:         current,
			Amount:       dailyCost,
			Currency:     "USD",
			Service:      "compute",
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

func (s *CostAnalysisService) aggregateCosts(costs []CostData, startDate, endDate time.Time, period string) *CostSummary {
	summary := &CostSummary{
		Currency:    "USD",
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		ByService:   make(map[string]float64),
		ByProvider:  make(map[string]float64),
		ByRegion:    make(map[string]float64),
		ByWorkspace: make(map[string]float64),
		DailyCosts:  costs,
	}

	for _, cost := range costs {
		summary.TotalCost += cost.Amount
		summary.ByService[cost.Service] += cost.Amount
		summary.ByProvider[cost.Provider] += cost.Amount
		summary.ByRegion[cost.Region] += cost.Amount
		summary.ByWorkspace[cost.WorkspaceID] += cost.Amount
	}

	return summary
}

func (s *CostAnalysisService) calculateTrend(costs []CostData) (string, float64) {
	if len(costs) < 2 {
		return "stable", 0.0
	}

	// Sort by date
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].Date.Before(costs[j].Date)
	})

	// Calculate daily totals
	dailyTotals := make(map[string]float64)
	for _, cost := range costs {
		dateStr := cost.Date.Format("2006-01-02")
		dailyTotals[dateStr] += cost.Amount
	}

	// Convert to sorted slice
	var dailyValues []float64
	for _, total := range dailyTotals {
		dailyValues = append(dailyValues, total)
	}

	if len(dailyValues) < 2 {
		return "stable", 0.0
	}

	// Calculate growth rate
	first := dailyValues[0]
	last := dailyValues[len(dailyValues)-1]
	growthRate := ((last - first) / first) * 100

	var trend string
	if growthRate > 5 {
		trend = "increasing"
	} else if growthRate < -5 {
		trend = "decreasing"
	} else {
		trend = "stable"
	}

	return trend, growthRate
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

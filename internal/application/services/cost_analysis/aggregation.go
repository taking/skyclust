package cost_analysis

import (
	"sort"
	"time"

	serviceconstants "skyclust/internal/application/services"
)

// Cost Aggregation Functions
// All cost aggregation and breakdown operations are implemented in this file

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

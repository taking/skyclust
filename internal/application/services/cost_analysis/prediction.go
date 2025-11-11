package cost_analysis

import (
	"math"
	"sort"
	"time"
)

// Cost Prediction Functions
// All cost prediction and forecasting operations are implemented in this file

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

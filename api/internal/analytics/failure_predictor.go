package analytics

import (
	"math"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/internal/architecture"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultFailurePredictor implements the FailurePredictor interface
type DefaultFailurePredictor struct {
	logger *zap.Logger
}

// NewDefaultFailurePredictor creates a new default failure predictor
func NewDefaultFailurePredictor(logger *zap.Logger) *DefaultFailurePredictor {
	return &DefaultFailurePredictor{
		logger: logger,
	}
}

// PredictFailure predicts the likelihood of future payment failures for a customer
func (fp *DefaultFailurePredictor) PredictFailure(customerID string, history []*architecture.PaymentFailure) *Prediction {
	if len(history) == 0 {
		return nil
	}

	// Need at least 2 events for meaningful prediction
	if len(history) < 2 {
		return nil
	}

	// Calculate risk score
	riskScore := fp.PredictRiskScore(customerID, history)
	
	// Calculate failure probability
	failureProbability := fp.calculateFailureProbability(riskScore, history)
	
	// Predict next failure date
	nextFailureDate := fp.PredictNextFailureDate(customerID, history)
	
	// Calculate confidence
	confidence := fp.calculatePredictionConfidence(history)
	
	// Identify contributing factors
	factors := fp.identifyRiskFactors(customerID, history)
	
	// Set expiration (predictions valid for 30 days)
	expiresAt := time.Now().AddDate(0, 0, 30)
	
	prediction := &Prediction{
		ID:                 uuid.New().String(),
		CustomerID:         customerID,
		RiskScore:          riskScore,
		FailureProbability: failureProbability,
		NextFailureDate:    nextFailureDate,
		Confidence:         confidence,
		Factors:            factors,
		CreatedAt:          time.Now(),
		ExpiresAt:          expiresAt,
	}
	
	fp.logger.Info("Generated payment failure prediction",
		zap.String("customerID", customerID),
		zap.Float64("riskScore", riskScore),
		zap.Float64("failureProbability", failureProbability),
		zap.Float64("confidence", confidence))
	
	return prediction
}

// PredictRiskScore calculates a risk score for a customer based on their payment failure history
func (fp *DefaultFailurePredictor) PredictRiskScore(customerID string, history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}

	// Base risk score starts at 0
	riskScore := 0.0
	
	// Factor 1: Frequency of failures (30% weight)
	frequencyScore := fp.calculateFrequencyScore(history)
	riskScore += frequencyScore * 0.3
	
	// Factor 2: Amount of failures (25% weight)
	amountScore := fp.calculateAmountScore(history)
	riskScore += amountScore * 0.25
	
	// Factor 3: Recency of failures (20% weight)
	recencyScore := fp.calculateRecencyScore(history)
	riskScore += recencyScore * 0.2
	
	// Factor 4: Pattern consistency (15% weight)
	consistencyScore := fp.calculateConsistencyScore(history)
	riskScore += consistencyScore * 0.15
	
	// Factor 5: Business category risk (10% weight)
	categoryScore := fp.calculateBusinessCategoryScore(history)
	riskScore += categoryScore * 0.1
	
	// Normalize risk score to 0-100 range
	riskScore = math.Min(100.0, math.Max(0.0, riskScore))
	
	return riskScore
}

// PredictNextFailureDate predicts when the next payment failure is likely to occur
func (fp *DefaultFailurePredictor) PredictNextFailureDate(customerID string, history []*architecture.PaymentFailure) *time.Time {
	if len(history) < 2 {
		return nil
	}

	// Sort history by timestamp
	sortedHistory := fp.sortByTimestamp(history)
	
	// Calculate average time between failures
	timeGaps := fp.calculateTimeGaps(sortedHistory)
	if len(timeGaps) == 0 {
		return nil
	}
	
	// Calculate average gap
	averageGap := fp.calculateAverageGap(timeGaps)
	
	// Predict next failure based on last failure + average gap
	lastFailure := sortedHistory[len(sortedHistory)-1].OccurredAt
	nextFailure := lastFailure.Add(averageGap)
	
	// Adjust based on risk score (higher risk = sooner failure)
	riskScore := fp.PredictRiskScore(customerID, history)
	riskAdjustment := time.Duration(float64(averageGap) * (riskScore / 100.0) * 0.5)
	nextFailure = nextFailure.Add(-riskAdjustment)
	
	// Ensure prediction is in the future
	if nextFailure.Before(time.Now()) {
		nextFailure = time.Now().Add(24 * time.Hour) // Default to tomorrow
	}
	
	return &nextFailure
}

// Helper methods for risk calculation
func (fp *DefaultFailurePredictor) calculateFrequencyScore(history []*architecture.PaymentFailure) float64 {
	eventCount := len(history)
	
	// Exponential scoring: more events = exponentially higher risk
	if eventCount == 0 {
		return 0.0
	}
	
	// Base score increases with each event
	baseScore := float64(eventCount) * 10.0
	
	// Apply exponential growth for high-frequency customers
	if eventCount > 5 {
		baseScore *= math.Pow(1.5, float64(eventCount-5))
	}
	
	// Cap at 100
	return math.Min(100.0, baseScore)
}

func (fp *DefaultFailurePredictor) calculateAmountScore(history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Calculate total amount and average amount
	totalAmount := 0.0
	for _, event := range history {
		totalAmount += event.Amount
	}
	averageAmount := totalAmount / float64(len(history))
	
	// Score based on average amount (higher amounts = higher risk)
	// Normalize: $1000 = 50 points, $10000 = 100 points
	amountScore := (averageAmount / 1000.0) * 50.0
	
	// Cap at 100
	return math.Min(100.0, amountScore)
}

func (fp *DefaultFailurePredictor) calculateRecencyScore(history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Find most recent failure
	mostRecent := history[0].OccurredAt
	for _, event := range history {
		if event.OccurredAt.After(mostRecent) {
			mostRecent = event.OccurredAt
		}
	}
	
	// Calculate days since last failure
	daysSince := time.Since(mostRecent).Hours() / 24.0
	
	// Score: recent failures = higher risk
	// 0 days = 100 points, 30 days = 50 points, 90+ days = 0 points
	if daysSince <= 0 {
		return 100.0
	} else if daysSince >= 90 {
		return 0.0
	} else {
		// Linear interpolation
		return 100.0 - (daysSince / 90.0) * 100.0
	}
}

func (fp *DefaultFailurePredictor) calculateConsistencyScore(history []*architecture.PaymentFailure) float64 {
	if len(history) < 2 {
		return 0.0
	}
	
	// Calculate time gaps between failures
	timeGaps := fp.calculateTimeGaps(history)
	if len(timeGaps) == 0 {
		return 0.0
	}
	
	// Calculate standard deviation of gaps
	meanGap := fp.calculateAverageGap(timeGaps)
	variance := 0.0
	
	for _, gap := range timeGaps {
		diff := gap - meanGap
		variance += float64(diff * diff)
	}
	variance /= float64(len(timeGaps))
	
	// Lower variance = higher consistency = higher risk
	// Higher variance = lower consistency = lower risk
	// Normalize standard deviation to 0-100 scale
	stdDev := math.Sqrt(variance)
	normalizedStdDev := stdDev / float64(meanGap) // Coefficient of variation
	
	// Invert: lower CV = higher consistency = higher risk
	consistencyScore := 100.0 - (normalizedStdDev * 100.0)
	
	// Ensure score is in valid range
	return math.Max(0.0, math.Min(100.0, consistencyScore))
}

func (fp *DefaultFailurePredictor) calculateBusinessCategoryScore(history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Define risk levels for different business categories
	categoryRisk := map[string]float64{
		"retail":          30.0,
		"healthcare":      40.0,
		"technology":      25.0,
		"finance":         60.0,
		"construction":    70.0,
		"hospitality":     45.0,
		"manufacturing":   50.0,
		"education":       35.0,
		"real_estate":     65.0,
		"transportation":  55.0,
	}
	
	// Calculate weighted average based on event frequency
	categoryCounts := make(map[string]int)
	totalEvents := len(history)
	
	for _, event := range history {
		if event.BusinessCategory != "" {
			categoryCounts[event.BusinessCategory]++
		}
	}
	
	// Calculate weighted score
	totalScore := 0.0
	for category, count := range categoryCounts {
		risk := categoryRisk[category]
		if risk == 0 {
			risk = 50.0 // Default risk for unknown categories
		}
		weight := float64(count) / float64(totalEvents)
		totalScore += risk * weight
	}
	
	return totalScore
}

func (fp *DefaultFailurePredictor) calculateFailureProbability(riskScore float64, history []*architecture.PaymentFailure) float64 {
	// Convert risk score (0-100) to probability (0-1)
	// Use sigmoid function for smooth probability curve
	baseProbability := riskScore / 100.0
	
	// Apply sigmoid transformation for better probability distribution
	// This creates a more realistic probability curve
	sigmoidProbability := 1.0 / (1.0 + math.Exp(-5.0*(baseProbability-0.5)))
	
	// Adjust based on historical frequency
	frequencyAdjustment := float64(len(history)) * 0.05
	adjustedProbability := sigmoidProbability + frequencyAdjustment
	
	// Cap probability at 0.95 (never 100% certain)
	return math.Min(0.95, adjustedProbability)
}

func (fp *DefaultFailurePredictor) calculatePredictionConfidence(history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Base confidence starts at 0.5
	baseConfidence := 0.5
	
	// More data = higher confidence
	dataBonus := math.Min(0.3, float64(len(history))*0.05)
	
	// Recent data = higher confidence
	recencyBonus := fp.calculateRecencyConfidence(history)
	
	// Pattern consistency = higher confidence
	consistencyBonus := fp.calculateConsistencyConfidence(history)
	
	totalConfidence := baseConfidence + dataBonus + recencyBonus + consistencyBonus
	
	// Cap at 1.0
	return math.Min(1.0, totalConfidence)
}

func (fp *DefaultFailurePredictor) calculateRecencyConfidence(history []*architecture.PaymentFailure) float64 {
	if len(history) == 0 {
		return 0.0
	}
	
	// Find most recent failure
	mostRecent := history[0].OccurredAt
	for _, event := range history {
		if event.OccurredAt.After(mostRecent) {
			mostRecent = event.OccurredAt
		}
	}
	
	// Calculate days since last failure
	daysSince := time.Since(mostRecent).Hours() / 24.0
	
	// Recent data = higher confidence
	// 0 days = 0.2 bonus, 30 days = 0.1 bonus, 90+ days = 0.0 bonus
	if daysSince <= 0 {
		return 0.2
	} else if daysSince >= 90 {
		return 0.0
	} else {
		// Linear interpolation
		return 0.2 - (daysSince / 90.0) * 0.1
	}
}

func (fp *DefaultFailurePredictor) calculateConsistencyConfidence(history []*architecture.PaymentFailure) float64 {
	if len(history) < 2 {
		return 0.0
	}
	
	// Calculate time gaps between failures
	timeGaps := fp.calculateTimeGaps(history)
	if len(timeGaps) == 0 {
		return 0.0
	}
	
	// Calculate coefficient of variation
	meanGap := fp.calculateAverageGap(timeGaps)
	variance := 0.0
	
	for _, gap := range timeGaps {
		diff := gap - meanGap
		variance += float64(diff * diff)
	}
	variance /= float64(len(timeGaps))
	
	// Lower variance = higher consistency = higher confidence
	stdDev := math.Sqrt(variance)
	coefficientOfVariation := stdDev / float64(meanGap)
	
	// Convert to confidence bonus (0.0 to 0.1)
	confidenceBonus := math.Max(0.0, 0.1-0.1*coefficientOfVariation)
	
	return confidenceBonus
}

func (fp *DefaultFailurePredictor) identifyRiskFactors(customerID string, history []*architecture.PaymentFailure) []string {
	factors := make([]string, 0)
	
	// Factor 1: High frequency
	if len(history) >= 5 {
		factors = append(factors, "High failure frequency")
	}
	
	// Factor 2: High amounts
	totalAmount := 0.0
	for _, event := range history {
		totalAmount += event.Amount
	}
	if totalAmount >= 10000 {
		factors = append(factors, "High total failure amount")
	}
	
	// Factor 3: Recent failures
	if len(history) > 0 {
		mostRecent := history[0].OccurredAt
		for _, event := range history {
			if event.OccurredAt.After(mostRecent) {
				mostRecent = event.OccurredAt
			}
		}
		
		daysSince := time.Since(mostRecent).Hours() / 24.0
		if daysSince <= 7 {
			factors = append(factors, "Very recent failure")
		} else if daysSince <= 30 {
			factors = append(factors, "Recent failure")
		}
	}
	
	// Factor 4: Business category
	if len(history) > 0 {
		category := history[0].BusinessCategory
		highRiskCategories := []string{"construction", "real_estate", "finance"}
		for _, highRisk := range highRiskCategories {
			if category == highRisk {
				factors = append(factors, "High-risk business category")
				break
			}
		}
	}
	
	// Factor 5: Pattern consistency
	if len(history) >= 3 {
		timeGaps := fp.calculateTimeGaps(history)
		if len(timeGaps) > 0 {
			meanGap := fp.calculateAverageGap(timeGaps)
			variance := 0.0
			for _, gap := range timeGaps {
				diff := gap - meanGap
				variance += float64(diff * diff)
			}
			variance /= float64(len(timeGaps))
			
			// Low variance indicates consistent pattern
			if variance < float64(meanGap*meanGap)*0.1 {
				factors = append(factors, "Consistent failure pattern")
			}
		}
	}
	
	return factors
}

func (fp *DefaultFailurePredictor) sortByTimestamp(history []*architecture.PaymentFailure) []*architecture.PaymentFailure {
	sorted := make([]*architecture.PaymentFailure, len(history))
	copy(sorted, history)
	
	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].OccurredAt.After(sorted[j+1].OccurredAt) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return sorted
}

func (fp *DefaultFailurePredictor) calculateTimeGaps(history []*architecture.PaymentFailure) []time.Duration {
	if len(history) < 2 {
		return []time.Duration{}
	}
	
	gaps := make([]time.Duration, 0, len(history)-1)
	
	for i := 1; i < len(history); i++ {
		gap := history[i].OccurredAt.Sub(history[i-1].OccurredAt)
		gaps = append(gaps, gap)
	}
	
	return gaps
}

func (fp *DefaultFailurePredictor) calculateAverageGap(gaps []time.Duration) time.Duration {
	if len(gaps) == 0 {
		return 0
	}
	
	total := time.Duration(0)
	for _, gap := range gaps {
		total += gap
	}
	
	return total / time.Duration(len(gaps))
}

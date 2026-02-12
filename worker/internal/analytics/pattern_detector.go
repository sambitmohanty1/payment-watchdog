package analytics

import (
	"fmt"
	"sort"
	"time"

	"github.com/lexure-intelligence/payment-watchdog/internal/architecture"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultPatternDetector implements the PatternDetector interface
type DefaultPatternDetector struct {
	logger *zap.Logger
}

// NewDefaultPatternDetector creates a new default pattern detector
func NewDefaultPatternDetector(logger *zap.Logger) *DefaultPatternDetector {
	return &DefaultPatternDetector{
		logger: logger,
	}
}

// DetectPatterns detects various patterns in payment failure events
func (pd *DefaultPatternDetector) DetectPatterns(events []*architecture.PaymentFailure) []Pattern {
	patterns := make([]Pattern, 0)
	
	if len(events) == 0 {
		return patterns
	}

	// Detect recurring patterns
	recurringPatterns := pd.detectRecurringPatterns(events)
	patterns = append(patterns, recurringPatterns...)

	// Detect amount-based patterns
	amountPatterns := pd.detectAmountPatterns(events)
	patterns = append(patterns, amountPatterns...)

	// Detect time-based patterns
	timePatterns := pd.detectTimePatterns(events)
	patterns = append(patterns, timePatterns...)

	// Detect business category patterns
	businessPatterns := pd.DetectBusinessPatterns(events)
	patterns = append(patterns, businessPatterns...)

	return patterns
}

// DetectCustomerPatterns detects patterns specific to a customer
func (pd *DefaultPatternDetector) DetectCustomerPatterns(customerID string, events []*architecture.PaymentFailure) []CustomerPattern {
	customerEvents := pd.filterEventsByCustomer(events, customerID)
	if len(customerEvents) == 0 {
		return []CustomerPattern{}
	}

	patterns := make([]CustomerPattern, 0)
	
	// Detect customer-specific recurring patterns
	recurringPatterns := pd.detectRecurringPatterns(customerEvents)
	for _, pattern := range recurringPatterns {
		customerPattern := CustomerPattern{
			CustomerID:  customerID,
			Pattern:     pattern,
			Frequency:   float64(pd.calculateFrequency(customerEvents, pattern.Type)),
			TotalAmount: pd.calculateTotalAmount(customerEvents),
			RiskLevel:   pd.calculateCustomerRiskLevel(customerEvents),
		}
		patterns = append(patterns, customerPattern)
	}

	return patterns
}

// DetectTemporalPatterns detects time-based patterns in the data
func (pd *DefaultPatternDetector) DetectTemporalPatterns(events []*architecture.PaymentFailure, timeRange time.Duration) []TemporalPattern {
	temporalPatterns := make([]TemporalPattern, 0)
	
	if len(events) == 0 {
		return temporalPatterns
	}

	// Detect day-of-week patterns
	dayOfWeekPattern := pd.detectDayOfWeekPattern(events)
	if dayOfWeekPattern != nil {
		temporalPattern := TemporalPattern{
			Pattern:     *dayOfWeekPattern,
			TimeRange:   timeRange,
			Frequency:   float64(pd.calculateFrequency(events, PatternTypeDayOfWeek)),
			PeakTimes:   pd.findPeakTimes(events, "day_of_week"),
			Seasonality: "weekly",
		}
		temporalPatterns = append(temporalPatterns, temporalPattern)
	}

	// Detect time-of-day patterns
	timeOfDayPattern := pd.detectTimeOfDayPattern(events)
	if timeOfDayPattern != nil {
		temporalPattern := TemporalPattern{
			Pattern:     *timeOfDayPattern,
			TimeRange:   timeRange,
			Frequency:   float64(pd.calculateFrequency(events, PatternTypeTimeOfDay)),
			PeakTimes:   pd.findPeakTimes(events, "time_of_day"),
			Seasonality: "daily",
		}
		temporalPatterns = append(temporalPatterns, temporalPattern)
	}

	return temporalPatterns
}

// DetectBusinessPatterns detects business-related patterns in payment failures
func (pd *DefaultPatternDetector) DetectBusinessPatterns(events []*architecture.PaymentFailure) []Pattern {
	patterns := make([]Pattern, 0)
	
	if len(events) == 0 {
		return patterns
	}

	// Group events by business category
	categoryGroups := pd.groupEventsByBusinessCategory(events)
	
	for category, categoryEvents := range categoryGroups {
		if len(categoryEvents) >= 2 { // Minimum events for category pattern
			pattern := Pattern{
				ID:          uuid.New().String(),
				Type:        PatternTypeBusiness,
				Confidence:  pd.calculateBusinessCategoryConfidence(categoryEvents, events),
				Description: "Business category pattern detected",
				Evidence:    []string{category, "Category-specific failures"},
				CreatedAt:   time.Now(),
				Metadata: map[string]interface{}{
					"business_category": category,
					"event_count":       len(categoryEvents),
					"total_events":      len(events),
				},
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// detectRecurringPatterns detects recurring patterns in payment failures
func (pd *DefaultPatternDetector) detectRecurringPatterns(events []*architecture.PaymentFailure) []Pattern {
	patterns := make([]Pattern, 0)
	
	// Group events by customer and detect recurring patterns
	customerGroups := pd.groupEventsByCustomer(events)
	
	for customerID, customerEvents := range customerGroups {
		if len(customerEvents) >= 3 { // Minimum events for pattern detection
			pattern := Pattern{
				ID:          uuid.New().String(),
				Type:        PatternTypeRecurring,
				Confidence:  pd.calculateConfidence(customerEvents),
				Description: "Recurring payment failures detected for customer",
				Evidence:    []string{customerID, "Multiple failures over time"},
				CreatedAt:   time.Now(),
				Metadata: map[string]interface{}{
					"customer_id": customerID,
					"event_count": len(customerEvents),
					"time_span":   pd.calculateTimeSpan(customerEvents),
				},
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// detectAmountPatterns detects patterns based on payment amounts
func (pd *DefaultPatternDetector) detectAmountPatterns(events []*architecture.PaymentFailure) []Pattern {
	patterns := make([]Pattern, 0)
	
	if len(events) == 0 {
		return patterns
	}

	// Calculate amount statistics
	amounts := pd.extractAmounts(events)
	sort.Float64s(amounts)
	
	// Detect high-value patterns
	highValueThreshold := pd.calculateHighValueThreshold(amounts)
	highValueEvents := pd.filterHighValueEvents(events, highValueThreshold)
	
	if len(highValueEvents) > 0 {
		pattern := Pattern{
			ID:          uuid.New().String(),
			Type:        PatternTypeAmount,
			Confidence:  pd.calculateAmountConfidence(highValueEvents, events),
			Description: "High-value payment failure pattern detected",
			Evidence:    []string{"High failure amounts", "Significant financial impact"},
			CreatedAt:   time.Now(),
			Metadata: map[string]interface{}{
				"threshold":     highValueThreshold,
				"high_value_count": len(highValueEvents),
				"total_events":  len(events),
			},
		}
		patterns = append(patterns, pattern)
	}
	
	return patterns
}

// detectTimePatterns detects time-based patterns
func (pd *DefaultPatternDetector) detectTimePatterns(events []*architecture.PaymentFailure) []Pattern {
	patterns := make([]Pattern, 0)
	
	// Detect day-of-week pattern
	dayOfWeekPattern := pd.detectDayOfWeekPattern(events)
	if dayOfWeekPattern != nil {
		patterns = append(patterns, *dayOfWeekPattern)
	}

	// Detect time-of-day pattern
	timeOfDayPattern := pd.detectTimeOfDayPattern(events)
	if timeOfDayPattern != nil {
		patterns = append(patterns, *timeOfDayPattern)
	}
	
	return patterns
}

// Helper methods for pattern detection
func (pd *DefaultPatternDetector) groupEventsByCustomer(events []*architecture.PaymentFailure) map[string][]*architecture.PaymentFailure {
	groups := make(map[string][]*architecture.PaymentFailure)
	
	for _, event := range events {
		if event.CustomerID != "" {
			customerID := event.CustomerID
			groups[customerID] = append(groups[customerID], event)
		}
	}
	
	return groups
}

func (pd *DefaultPatternDetector) groupEventsByBusinessCategory(events []*architecture.PaymentFailure) map[string][]*architecture.PaymentFailure {
	groups := make(map[string][]*architecture.PaymentFailure)
	
	for _, event := range events {
		if event.BusinessCategory != "" {
			groups[event.BusinessCategory] = append(groups[event.BusinessCategory], event)
		}
	}
	
	return groups
}

func (pd *DefaultPatternDetector) filterEventsByCustomer(events []*architecture.PaymentFailure, customerID string) []*architecture.PaymentFailure {
	filtered := make([]*architecture.PaymentFailure, 0)
	
	for _, event := range events {
		if event.CustomerID == customerID {
			filtered = append(filtered, event)
		}
	}
	
	return filtered
}

func (pd *DefaultPatternDetector) extractAmounts(events []*architecture.PaymentFailure) []float64 {
	amounts := make([]float64, 0, len(events))
	
	for _, event := range events {
		amounts = append(amounts, event.Amount)
	}
	
	return amounts
}

func (pd *DefaultPatternDetector) calculateHighValueThreshold(amounts []float64) float64 {
	if len(amounts) == 0 {
		return 0
	}
	
	// Use 90th percentile as high-value threshold
	index := int(float64(len(amounts)) * 0.9)
	if index >= len(amounts) {
		index = len(amounts) - 1
	}
	
	return amounts[index]
}

func (pd *DefaultPatternDetector) filterHighValueEvents(events []*architecture.PaymentFailure, threshold float64) []*architecture.PaymentFailure {
	filtered := make([]*architecture.PaymentFailure, 0)
	
	for _, event := range events {
		if event.Amount >= threshold {
			filtered = append(filtered, event)
		}
	}
	
	return filtered
}

func (pd *DefaultPatternDetector) calculateConfidence(events []*architecture.PaymentFailure) float64 {
	if len(events) < 3 {
		return 0.3
	}
	
	// Simple confidence calculation based on event count and consistency
	baseConfidence := 0.5
	eventBonus := float64(len(events)) * 0.1
	consistencyBonus := pd.calculateConsistencyBonus(events)
	
	confidence := baseConfidence + eventBonus + consistencyBonus
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

func (pd *DefaultPatternDetector) calculateAmountConfidence(highValueEvents []*architecture.PaymentFailure, totalEvents []*architecture.PaymentFailure) float64 {
	if len(totalEvents) == 0 {
		return 0
	}
	
	// Confidence based on proportion of high-value events
	proportion := float64(len(highValueEvents)) / float64(len(totalEvents))
	return proportion * 0.8 + 0.2 // Base confidence of 0.2
}

func (pd *DefaultPatternDetector) calculateBusinessCategoryConfidence(categoryEvents []*architecture.PaymentFailure, totalEvents []*architecture.PaymentFailure) float64 {
	if len(totalEvents) == 0 {
		return 0
	}
	
	// Confidence based on proportion of category events
	proportion := float64(len(categoryEvents)) / float64(len(totalEvents))
	return proportion * 0.7 + 0.3 // Base confidence of 0.3
}

func (pd *DefaultPatternDetector) calculateConsistencyBonus(events []*architecture.PaymentFailure) float64 {
	if len(events) < 2 {
		return 0
	}
	
	// Calculate time consistency between events
	timeGaps := make([]time.Duration, 0)
	for i := 1; i < len(events); i++ {
		gap := events[i].OccurredAt.Sub(events[i-1].OccurredAt)
		timeGaps = append(timeGaps, gap)
	}
	
	if len(timeGaps) == 0 {
		return 0
	}
	
	// Calculate standard deviation of time gaps
	meanGap := pd.calculateMeanDuration(timeGaps)
	variance := 0.0
	for _, gap := range timeGaps {
		diff := gap - meanGap
		variance += float64(diff * diff)
	}
	variance /= float64(len(timeGaps))
	
	// Lower variance = higher consistency = higher bonus
	consistencyScore := 1.0 / (1.0 + variance/float64(time.Hour*24)) // Normalize by day
	return consistencyScore * 0.2 // Max bonus of 0.2
}

func (pd *DefaultPatternDetector) calculateMeanDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	total := time.Duration(0)
	for _, duration := range durations {
		total += duration
	}
	
	return total / time.Duration(len(durations))
}

func (pd *DefaultPatternDetector) calculateFrequency(events []*architecture.PaymentFailure, patternType PatternType) int {
	// Simple frequency calculation - count events
	return len(events)
}

func (pd *DefaultPatternDetector) calculateTotalAmount(events []*architecture.PaymentFailure) float64 {
	total := 0.0
	for _, event := range events {
		total += event.Amount
	}
	return total
}

func (pd *DefaultPatternDetector) calculateCustomerRiskLevel(events []*architecture.PaymentFailure) string {
	if len(events) == 0 {
		return "low"
	}
	
	// Simple risk level calculation based on event count and amounts
	eventCount := len(events)
	totalAmount := pd.calculateTotalAmount(events)
	
	if eventCount >= 5 || totalAmount >= 10000 {
		return "high"
	} else if eventCount >= 3 || totalAmount >= 5000 {
		return "medium"
	}
	
	return "low"
}

func (pd *DefaultPatternDetector) calculateTimeSpan(events []*architecture.PaymentFailure) time.Duration {
	if len(events) < 2 {
		return 0
	}
	
	// Find earliest and latest events
	earliest := events[0].OccurredAt
	latest := events[0].OccurredAt
	
	for _, event := range events {
		if event.OccurredAt.Before(earliest) {
			earliest = event.OccurredAt
		}
		if event.OccurredAt.After(latest) {
			latest = event.OccurredAt
		}
	}
	
	return latest.Sub(earliest)
}

func (pd *DefaultPatternDetector) detectDayOfWeekPattern(events []*architecture.PaymentFailure) *Pattern {
	if len(events) < 7 {
		return nil
	}
	
	// Count events by day of week
	dayCounts := make(map[time.Weekday]int)
	for _, event := range events {
		day := event.OccurredAt.Weekday()
		dayCounts[day]++
	}
	
	// Find the day with most events
	var maxDay time.Weekday
	maxCount := 0
	for day, count := range dayCounts {
		if count > maxCount {
			maxCount = count
			maxDay = day
		}
	}
	
	// Only create pattern if there's a significant difference
	if maxCount >= 2 && float64(maxCount)/float64(len(events)) >= 0.3 {
		return &Pattern{
			ID:          uuid.New().String(),
			Type:        PatternTypeDayOfWeek,
			Confidence:  pd.calculateConfidence(events),
			Description: "Day-of-week payment failure pattern detected",
			Evidence:    []string{maxDay.String(), "Peak failure day"},
			CreatedAt:   time.Now(),
			Metadata: map[string]interface{}{
				"peak_day":    maxDay.String(),
				"peak_count":  maxCount,
				"total_events": len(events),
			},
		}
	}
	
	return nil
}

func (pd *DefaultPatternDetector) detectTimeOfDayPattern(events []*architecture.PaymentFailure) *Pattern {
	if len(events) < 24 {
		return nil
	}
	
	// Count events by hour of day
	hourCounts := make(map[int]int)
	for _, event := range events {
		hour := event.OccurredAt.Hour()
		hourCounts[hour]++
	}
	
	// Find the hour with most events
	var maxHour int
	maxCount := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}
	
	// Only create pattern if there's a significant difference
	if maxCount >= 2 && float64(maxCount)/float64(len(events)) >= 0.15 {
		return &Pattern{
			ID:          uuid.New().String(),
			Type:        PatternTypeTimeOfDay,
			Confidence:  pd.calculateConfidence(events),
			Description: "Time-of-day payment failure pattern detected",
			Evidence:    []string{fmt.Sprintf("%02d:00", maxHour), "Peak failure hour"},
			CreatedAt:   time.Now(),
			Metadata: map[string]interface{}{
				"peak_hour":   maxHour,
				"peak_count":  maxCount,
				"total_events": len(events),
			},
		}
	}
	
	return nil
}

func (pd *DefaultPatternDetector) findPeakTimes(events []*architecture.PaymentFailure, patternType string) []time.Time {
	if len(events) == 0 {
		return []time.Time{}
	}
	
	// For simplicity, return the first few event times as peak times
	peakCount := 3
	if len(events) < peakCount {
		peakCount = len(events)
	}
	
	peakTimes := make([]time.Time, 0, peakCount)
	for i := 0; i < peakCount; i++ {
		peakTimes = append(peakTimes, events[i].OccurredAt)
	}
	
	return peakTimes
}

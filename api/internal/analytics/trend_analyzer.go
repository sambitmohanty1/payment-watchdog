package analytics

import (
	"fmt"
	"sort"
	"time"

	"github.com/payment-watchdog/internal/architecture"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultTrendAnalyzer implements the TrendAnalyzer interface
type DefaultTrendAnalyzer struct {
	logger *zap.Logger
}

// NewDefaultTrendAnalyzer creates a new default trend analyzer
func NewDefaultTrendAnalyzer(logger *zap.Logger) *DefaultTrendAnalyzer {
	return &DefaultTrendAnalyzer{
		logger: logger,
	}
}

// AnalyzeTrends analyzes trends in payment failure data over a specified time range
func (ta *DefaultTrendAnalyzer) AnalyzeTrends(data []*architecture.PaymentFailure, timeRange time.Duration) []Trend {
	trends := make([]Trend, 0)

	if len(data) == 0 {
		return trends
	}

	// Sort data by timestamp
	sortedData := ta.sortByTimestamp(data)

	// Analyze failure rate trends
	failureRateTrend := ta.analyzeFailureRateTrend(sortedData, timeRange)
	if failureRateTrend != nil {
		trends = append(trends, *failureRateTrend)
	}

	// Analyze amount trends
	amountTrend := ta.analyzeAmountTrend(sortedData, timeRange)
	if amountTrend != nil {
		trends = append(trends, *amountTrend)
	}

	// Analyze frequency trends
	frequencyTrend := ta.analyzeFrequencyTrend(sortedData, timeRange)
	if frequencyTrend != nil {
		trends = append(trends, *frequencyTrend)
	}

	// Analyze customer trends
	customerTrend := ta.analyzeCustomerTrend(sortedData, timeRange)
	if customerTrend != nil {
		trends = append(trends, *customerTrend)
	}

	return trends
}

// AnalyzeSeasonalPatterns analyzes seasonal patterns in the data
func (ta *DefaultTrendAnalyzer) AnalyzeSeasonalPatterns(data []*architecture.PaymentFailure) []SeasonalPattern {
	seasonalPatterns := make([]SeasonalPattern, 0)

	if len(data) == 0 {
		return seasonalPatterns
	}

	// Group data by month
	monthlyGroups := ta.groupByMonth(data)

	// Analyze monthly patterns
	for month, monthData := range monthlyGroups {
		if len(monthData) >= 2 { // Minimum data for pattern detection
			pattern := ta.createSeasonalPattern(month, monthData)
			seasonalPatterns = append(seasonalPatterns, pattern)
		}
	}

	// Group data by day of month
	dayGroups := ta.groupByDayOfMonth(data)

	// Analyze day-of-month patterns
	for day, dayData := range dayGroups {
		if len(dayData) >= 3 { // Minimum data for pattern detection
			pattern := ta.createDayOfMonthPattern(day, dayData)
			seasonalPatterns = append(seasonalPatterns, pattern)
		}
	}

	return seasonalPatterns
}

// AnalyzeBusinessCyclePatterns analyzes business cycle patterns
func (ta *DefaultTrendAnalyzer) AnalyzeBusinessCyclePatterns(data []*architecture.PaymentFailure) []BusinessCyclePattern {
	businessCyclePatterns := make([]BusinessCyclePattern, 0)

	if len(data) == 0 {
		return businessCyclePatterns
	}

	// Analyze weekly business cycles
	weeklyPattern := ta.analyzeWeeklyBusinessCycle(data)
	if weeklyPattern != nil {
		businessCyclePatterns = append(businessCyclePatterns, *weeklyPattern)
	}

	// Analyze monthly business cycles
	monthlyPattern := ta.analyzeMonthlyBusinessCycle(data)
	if monthlyPattern != nil {
		businessCyclePatterns = append(businessCyclePatterns, *monthlyPattern)
	}

	// Analyze quarterly business cycles
	quarterlyPattern := ta.analyzeQuarterlyBusinessCycle(data)
	if quarterlyPattern != nil {
		businessCyclePatterns = append(businessCyclePatterns, *quarterlyPattern)
	}

	return businessCyclePatterns
}

// Helper methods for trend analysis
func (ta *DefaultTrendAnalyzer) sortByTimestamp(data []*architecture.PaymentFailure) []*architecture.PaymentFailure {
	sorted := make([]*architecture.PaymentFailure, len(data))
	copy(sorted, data)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OccurredAt.Before(sorted[j].OccurredAt)
	})

	return sorted
}

func (ta *DefaultTrendAnalyzer) analyzeFailureRateTrend(data []*architecture.PaymentFailure, timeRange time.Duration) *Trend {
	if len(data) < 2 {
		return nil
	}

	// Calculate failure rates over time windows
	windowSize := timeRange / 4 // 4 windows for analysis
	rates := ta.calculateFailureRates(data, windowSize)

	if len(rates) < 2 {
		return nil
	}

	// Determine trend direction and magnitude
	direction, magnitude, confidence := ta.calculateTrendMetrics(rates)

	return &Trend{
		ID:          uuid.New().String(),
		Type:        TrendTypeFailureRate,
		Direction:   direction,
		Magnitude:   magnitude,
		Confidence:  confidence,
		TimeRange:   timeRange,
		Description: ta.generateTrendDescription(TrendTypeFailureRate, direction, magnitude),
		CreatedAt:   time.Now(),
	}
}

func (ta *DefaultTrendAnalyzer) analyzeAmountTrend(data []*architecture.PaymentFailure, timeRange time.Duration) *Trend {
	if len(data) < 2 {
		return nil
	}

	// Calculate average amounts over time windows
	windowSize := timeRange / 4
	amounts := ta.calculateAverageAmounts(data, windowSize)

	if len(amounts) < 2 {
		return nil
	}

	// Determine trend direction and magnitude
	direction, magnitude, confidence := ta.calculateTrendMetrics(amounts)

	return &Trend{
		ID:          uuid.New().String(),
		Type:        TrendTypeAmount,
		Direction:   direction,
		Magnitude:   magnitude,
		Confidence:  confidence,
		TimeRange:   timeRange,
		Description: ta.generateTrendDescription(TrendTypeAmount, direction, magnitude),
		CreatedAt:   time.Now(),
	}
}

func (ta *DefaultTrendAnalyzer) analyzeFrequencyTrend(data []*architecture.PaymentFailure, timeRange time.Duration) *Trend {
	if len(data) < 2 {
		return nil
	}

	// Calculate event frequencies over time windows
	windowSize := timeRange / 4
	frequencies := ta.calculateEventFrequencies(data, windowSize)

	if len(frequencies) < 2 {
		return nil
	}

	// Determine trend direction and magnitude
	direction, magnitude, confidence := ta.calculateTrendMetrics(frequencies)

	return &Trend{
		ID:          uuid.New().String(),
		Type:        TrendTypeFrequency,
		Direction:   direction,
		Magnitude:   magnitude,
		Confidence:  confidence,
		TimeRange:   timeRange,
		Description: ta.generateTrendDescription(TrendTypeFrequency, direction, magnitude),
		CreatedAt:   time.Now(),
	}
}

func (ta *DefaultTrendAnalyzer) analyzeCustomerTrend(data []*architecture.PaymentFailure, timeRange time.Duration) *Trend {
	if len(data) < 2 {
		return nil
	}

	// Calculate unique customer counts over time windows
	windowSize := timeRange / 4
	customerCounts := ta.calculateCustomerCounts(data, windowSize)

	if len(customerCounts) < 2 {
		return nil
	}

	// Determine trend direction and magnitude
	direction, magnitude, confidence := ta.calculateTrendMetrics(customerCounts)

	return &Trend{
		ID:          uuid.New().String(),
		Type:        TrendTypeCustomer,
		Direction:   direction,
		Magnitude:   magnitude,
		Confidence:  confidence,
		TimeRange:   timeRange,
		Description: ta.generateTrendDescription(TrendTypeCustomer, direction, magnitude),
		CreatedAt:   time.Now(),
	}
}

func (ta *DefaultTrendAnalyzer) calculateFailureRates(data []*architecture.PaymentFailure, windowSize time.Duration) []float64 {
	rates := make([]float64, 0)

	if len(data) == 0 {
		return rates
	}

	// Find time boundaries
	startTime := data[0].OccurredAt
	endTime := data[len(data)-1].OccurredAt

	// Create time windows
	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(windowSize) {
		windowEnd := currentTime.Add(windowSize)

		// Count events in this window
		eventCount := 0
		for _, event := range data {
			if event.OccurredAt.After(currentTime) && event.OccurredAt.Before(windowEnd) {
				eventCount++
			}
		}

		// Calculate rate (events per hour)
		rate := float64(eventCount) / (float64(windowSize) / float64(time.Hour))
		rates = append(rates, rate)
	}

	return rates
}

func (ta *DefaultTrendAnalyzer) calculateAverageAmounts(data []*architecture.PaymentFailure, windowSize time.Duration) []float64 {
	amounts := make([]float64, 0)

	if len(data) == 0 {
		return amounts
	}

	// Find time boundaries
	startTime := data[0].OccurredAt
	endTime := data[len(data)-1].OccurredAt

	// Create time windows
	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(windowSize) {
		windowEnd := currentTime.Add(windowSize)

		// Calculate average amount in this window
		totalAmount := 0.0
		eventCount := 0

		for _, event := range data {
			if event.OccurredAt.After(currentTime) && event.OccurredAt.Before(windowEnd) {
				totalAmount += event.Amount
				eventCount++
			}
		}

		if eventCount > 0 {
			averageAmount := totalAmount / float64(eventCount)
			amounts = append(amounts, averageAmount)
		} else {
			amounts = append(amounts, 0.0)
		}
	}

	return amounts
}

func (ta *DefaultTrendAnalyzer) calculateEventFrequencies(data []*architecture.PaymentFailure, windowSize time.Duration) []float64 {
	frequencies := make([]float64, 0)

	if len(data) == 0 {
		return frequencies
	}

	// Find time boundaries
	startTime := data[0].OccurredAt
	endTime := data[len(data)-1].OccurredAt

	// Create time windows
	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(windowSize) {
		windowEnd := currentTime.Add(windowSize)

		// Count events in this window
		eventCount := 0
		for _, event := range data {
			if event.OccurredAt.After(currentTime) && event.OccurredAt.Before(windowEnd) {
				eventCount++
			}
		}

		frequencies = append(frequencies, float64(eventCount))
	}

	return frequencies
}

func (ta *DefaultTrendAnalyzer) calculateCustomerCounts(data []*architecture.PaymentFailure, windowSize time.Duration) []float64 {
	customerCounts := make([]float64, 0)

	if len(data) == 0 {
		return customerCounts
	}

	// Find time boundaries
	startTime := data[0].OccurredAt
	endTime := data[len(data)-1].OccurredAt

	// Create time windows
	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(windowSize) {
		windowEnd := currentTime.Add(windowSize)

		// Count unique customers in this window
		customerSet := make(map[string]bool)
		for _, event := range data {
			if event.OccurredAt.After(currentTime) && event.OccurredAt.Before(windowEnd) {
				if event.CustomerID != "" {
					customerSet[event.CustomerID] = true
				}
			}
		}

		customerCounts = append(customerCounts, float64(len(customerSet)))
	}

	return customerCounts
}

func (ta *DefaultTrendAnalyzer) calculateTrendMetrics(values []float64) (TrendDirection, float64, float64) {
	if len(values) < 2 {
		return TrendDirectionStable, 0.0, 0.0
	}

	// Calculate linear regression slope
	slope := ta.calculateLinearRegressionSlope(values)

	// Determine direction
	var direction TrendDirection
	if slope > 0.1 {
		direction = TrendDirectionIncreasing
	} else if slope < -0.1 {
		direction = TrendDirectionDecreasing
	} else {
		direction = TrendDirectionStable
	}

	// Calculate magnitude (absolute slope)
	magnitude := abs(slope)

	// Calculate confidence based on data consistency
	confidence := ta.calculateTrendConfidence(values)

	return direction, magnitude, confidence
}

func (ta *DefaultTrendAnalyzer) calculateLinearRegressionSlope(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	n := float64(len(values))

	// Calculate means
	sumX := 0.0
	sumY := 0.0
	for i, value := range values {
		sumX += float64(i)
		sumY += value
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate slope
	numerator := 0.0
	denominator := 0.0
	for i, value := range values {
		x := float64(i)
		numerator += (x - meanX) * (value - meanY)
		denominator += (x - meanX) * (x - meanX)
	}

	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

func (ta *DefaultTrendAnalyzer) calculateTrendConfidence(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	// Calculate R-squared (coefficient of determination)
	slope := ta.calculateLinearRegressionSlope(values)

	// Calculate means
	sumY := 0.0
	for _, value := range values {
		sumY += value
	}
	meanY := sumY / float64(len(values))

	// Calculate total sum of squares
	totalSS := 0.0
	for _, value := range values {
		totalSS += (value - meanY) * (value - meanY)
	}

	// Calculate regression sum of squares
	regressionSS := 0.0
	for i := range values {
		predicted := meanY + slope*float64(i)
		regressionSS += (predicted - meanY) * (predicted - meanY)
	}

	if totalSS == 0 {
		return 0.0
	}

	rSquared := regressionSS / totalSS

	// Ensure confidence is in [0,1] range
	if rSquared < 0.0 {
		rSquared = 0.0
	} else if rSquared > 1.0 {
		rSquared = 1.0
	}

	// Ensure minimum confidence for meaningful trends
	if rSquared < 0.1 {
		rSquared = 0.1
	}

	return rSquared
}

func (ta *DefaultTrendAnalyzer) generateTrendDescription(trendType TrendType, direction TrendDirection, magnitude float64) string {
	baseDescription := ""

	switch trendType {
	case TrendTypeFailureRate:
		baseDescription = "Payment failure rate is"
	case TrendTypeAmount:
		baseDescription = "Payment failure amounts are"
	case TrendTypeFrequency:
		baseDescription = "Payment failure frequency is"
	case TrendTypeCustomer:
		baseDescription = "Number of customers with failures is"
	default:
		baseDescription = "Trend is"
	}

	switch direction {
	case TrendDirectionIncreasing:
		return baseDescription + " increasing significantly"
	case TrendDirectionDecreasing:
		return baseDescription + " decreasing significantly"
	case TrendDirectionStable:
		return baseDescription + " remaining stable"
	case TrendDirectionCyclical:
		return baseDescription + " showing cyclical patterns"
	default:
		return baseDescription + " showing mixed patterns"
	}
}

func (ta *DefaultTrendAnalyzer) groupByMonth(data []*architecture.PaymentFailure) map[int][]*architecture.PaymentFailure {
	monthlyGroups := make(map[int][]*architecture.PaymentFailure)

	for _, event := range data {
		month := int(event.OccurredAt.Month())
		monthlyGroups[month] = append(monthlyGroups[month], event)
	}

	return monthlyGroups
}

func (ta *DefaultTrendAnalyzer) groupByDayOfMonth(data []*architecture.PaymentFailure) map[int][]*architecture.PaymentFailure {
	dayGroups := make(map[int][]*architecture.PaymentFailure)

	for _, event := range data {
		day := event.OccurredAt.Day()
		dayGroups[day] = append(dayGroups[day], event)
	}

	return dayGroups
}

func (ta *DefaultTrendAnalyzer) createSeasonalPattern(month int, monthData []*architecture.PaymentFailure) SeasonalPattern {
	// Create a base pattern for the month
	pattern := Pattern{
		ID:          uuid.New().String(),
		Type:        PatternTypeSeasonal,
		Confidence:  ta.calculateSeasonalConfidence(monthData),
		Description: "Seasonal pattern detected for month",
		Evidence:    []string{time.Month(month).String(), "Monthly pattern"},
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"month":       month,
			"month_name":  time.Month(month).String(),
			"event_count": len(monthData),
		},
	}

	// Determine season
	season := ta.determineSeason(month)

	// Calculate strength based on event count
	strength := ta.calculateSeasonalStrength(monthData)

	// Find peak days in the month
	peakDays := ta.findPeakDaysInMonth(monthData)

	return SeasonalPattern{
		Pattern:    pattern,
		Season:     season,
		Year:       time.Now().Year(),
		Strength:   strength,
		PeakMonths: []int{month},
		PeakDays:   peakDays,
	}
}

func (ta *DefaultTrendAnalyzer) createDayOfMonthPattern(day int, dayData []*architecture.PaymentFailure) SeasonalPattern {
	// Create a base pattern for the day
	pattern := Pattern{
		ID:          uuid.New().String(),
		Type:        PatternTypeSeasonal,
		Confidence:  ta.calculateSeasonalConfidence(dayData),
		Description: "Day-of-month pattern detected",
		Evidence:    []string{fmt.Sprintf("Day %d", day), "Monthly day pattern"},
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"day":         day,
			"event_count": len(dayData),
		},
	}

	// Determine season based on current month
	currentMonth := time.Now().Month()
	season := ta.determineSeason(int(currentMonth))

	// Calculate strength based on event count
	strength := ta.calculateSeasonalStrength(dayData)

	return SeasonalPattern{
		Pattern:    pattern,
		Season:     season,
		Year:       time.Now().Year(),
		Strength:   strength,
		PeakMonths: []int{int(currentMonth)},
		PeakDays:   []int{day},
	}
}

func (ta *DefaultTrendAnalyzer) determineSeason(month int) string {
	switch month {
	case 12, 1, 2:
		return "Summer"
	case 3, 4, 5:
		return "Autumn"
	case 6, 7, 8:
		return "Winter"
	case 9, 10, 11:
		return "Spring"
	default:
		return "Unknown"
	}
}

func (ta *DefaultTrendAnalyzer) calculateSeasonalConfidence(monthData []*architecture.PaymentFailure) float64 {
	if len(monthData) < 2 {
		return 0.3
	}

	// Simple confidence calculation based on data volume
	baseConfidence := 0.5
	dataBonus := float64(len(monthData)) * 0.1

	confidence := baseConfidence + dataBonus
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

func (ta *DefaultTrendAnalyzer) calculateSeasonalStrength(monthData []*architecture.PaymentFailure) float64 {
	if len(monthData) == 0 {
		return 0.0
	}

	// Strength based on event count and consistency
	eventCount := len(monthData)
	baseStrength := float64(eventCount) * 0.1

	if baseStrength > 1.0 {
		baseStrength = 1.0
	}

	return baseStrength
}

func (ta *DefaultTrendAnalyzer) findPeakDaysInMonth(monthData []*architecture.PaymentFailure) []int {
	if len(monthData) == 0 {
		return []int{}
	}

	// Group by day and find peak days
	dayCounts := make(map[int]int)
	for _, event := range monthData {
		day := event.OccurredAt.Day()
		dayCounts[day]++
	}

	// Find top 3 days
	var peakDays []int
	maxCount := 0
	for _, count := range dayCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	// Add days with at least 50% of max count
	for day, count := range dayCounts {
		if count >= maxCount/2 {
			peakDays = append(peakDays, day)
		}
	}

	// Sort peak days
	sort.Ints(peakDays)

	// Limit to top 3
	if len(peakDays) > 3 {
		peakDays = peakDays[:3]
	}

	return peakDays
}

func (ta *DefaultTrendAnalyzer) analyzeWeeklyBusinessCycle(data []*architecture.PaymentFailure) *BusinessCyclePattern {
	if len(data) < 7 {
		return nil
	}

	// Group by day of week
	dayGroups := make(map[time.Weekday][]*architecture.PaymentFailure)
	for _, event := range data {
		day := event.OccurredAt.Weekday()
		dayGroups[day] = append(dayGroups[day], event)
	}

	// Check if there's a weekly pattern
	hasPattern := false
	for _, dayData := range dayGroups {
		if len(dayData) >= 2 {
			hasPattern = true
			break
		}
	}

	if !hasPattern {
		return nil
	}

	// Create pattern
	pattern := Pattern{
		ID:          uuid.New().String(),
		Type:        PatternTypeBusiness,
		Confidence:  ta.calculateBusinessCycleConfidence(dayGroups),
		Description: "Weekly business cycle pattern detected",
		Evidence:    []string{"Weekly pattern", "Day-of-week variations"},
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"cycle_type": "weekly",
			"days":       len(dayGroups),
		},
	}

	// Calculate strength and find peak periods
	strength := ta.calculateBusinessCycleStrength(dayGroups)
	peakPeriods := ta.findWeeklyPeakPeriods(dayGroups)

	return &BusinessCyclePattern{
		Pattern:     pattern,
		CycleType:   "weekly",
		Duration:    7 * 24 * time.Hour,
		Strength:    strength,
		PeakPeriods: peakPeriods,
	}
}

func (ta *DefaultTrendAnalyzer) analyzeMonthlyBusinessCycle(data []*architecture.PaymentFailure) *BusinessCyclePattern {
	if len(data) < 30 {
		return nil
	}

	// Group by week of month
	weekGroups := make(map[int][]*architecture.PaymentFailure)
	for _, event := range data {
		week := ta.getWeekOfMonth(event.OccurredAt)
		weekGroups[week] = append(weekGroups[week], event)
	}

	// Check if there's a monthly pattern
	hasPattern := false
	for _, weekData := range weekGroups {
		if len(weekData) >= 3 {
			hasPattern = true
			break
		}
	}

	if !hasPattern {
		return nil
	}

	// Create pattern
	pattern := Pattern{
		ID:          uuid.New().String(),
		Type:        PatternTypeBusiness,
		Confidence:  ta.calculateBusinessCycleConfidence(weekGroups),
		Description: "Monthly business cycle pattern detected",
		Evidence:    []string{"Monthly pattern", "Week-of-month variations"},
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"cycle_type": "monthly",
			"weeks":      len(weekGroups),
		},
	}

	// Calculate strength and find peak periods
	strength := ta.calculateBusinessCycleStrength(weekGroups)
	peakPeriods := ta.findMonthlyPeakPeriods(weekGroups)

	return &BusinessCyclePattern{
		Pattern:     pattern,
		CycleType:   "monthly",
		Duration:    30 * 24 * time.Hour,
		Strength:    strength,
		PeakPeriods: peakPeriods,
	}
}

func (ta *DefaultTrendAnalyzer) analyzeQuarterlyBusinessCycle(data []*architecture.PaymentFailure) *BusinessCyclePattern {
	if len(data) < 90 {
		return nil
	}

	// Group by month
	monthGroups := make(map[int][]*architecture.PaymentFailure)
	for _, event := range data {
		month := int(event.OccurredAt.Month())
		monthGroups[month] = append(monthGroups[month], event)
	}

	// Check if there's a quarterly pattern
	hasPattern := false
	for _, monthData := range monthGroups {
		if len(monthData) >= 5 {
			hasPattern = true
			break
		}
	}

	if !hasPattern {
		return nil
	}

	// Create pattern
	pattern := Pattern{
		ID:          uuid.New().String(),
		Type:        PatternTypeBusiness,
		Confidence:  ta.calculateBusinessCycleConfidence(monthGroups),
		Description: "Quarterly business cycle pattern detected",
		Evidence:    []string{"Quarterly pattern", "Month-of-quarter variations"},
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"cycle_type": "quarterly",
			"months":     len(monthGroups),
		},
	}

	// Calculate strength and find peak periods
	strength := ta.calculateBusinessCycleStrength(monthGroups)
	peakPeriods := ta.findQuarterlyPeakPeriods(monthGroups)

	return &BusinessCyclePattern{
		Pattern:     pattern,
		CycleType:   "quarterly",
		Duration:    90 * 24 * time.Hour,
		Strength:    strength,
		PeakPeriods: peakPeriods,
	}
}

func (ta *DefaultTrendAnalyzer) getWeekOfMonth(t time.Time) int {
	day := t.Day()
	return (day-1)/7 + 1
}

func (ta *DefaultTrendAnalyzer) calculateBusinessCycleConfidence(groups interface{}) float64 {
	// Generic confidence calculation for business cycles
	baseConfidence := 0.6
	groupBonus := 0.1

	confidence := baseConfidence + groupBonus
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

func (ta *DefaultTrendAnalyzer) calculateBusinessCycleStrength(groups interface{}) float64 {
	// Generic strength calculation for business cycles
	baseStrength := 0.5
	strengthBonus := 0.2

	strength := baseStrength + strengthBonus
	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

func (ta *DefaultTrendAnalyzer) findWeeklyPeakPeriods(dayGroups map[time.Weekday][]*architecture.PaymentFailure) []time.Time {
	var peakPeriods []time.Time

	// Find the day with most events
	var maxDay time.Weekday
	maxCount := 0
	for day, dayData := range dayGroups {
		if len(dayData) > maxCount {
			maxCount = len(dayData)
			maxDay = day
		}
	}

	// Create a sample peak time for the peak day
	peakTime := time.Now()
	for peakTime.Weekday() != maxDay {
		peakTime = peakTime.Add(24 * time.Hour)
	}
	peakTime = peakTime.Add(9 * time.Hour) // 9 AM

	peakPeriods = append(peakPeriods, peakTime)

	return peakPeriods
}

func (ta *DefaultTrendAnalyzer) findMonthlyPeakPeriods(weekGroups map[int][]*architecture.PaymentFailure) []time.Time {
	var peakPeriods []time.Time

	// Find the week with most events
	var maxWeek int
	maxCount := 0
	for week, weekData := range weekGroups {
		if len(weekData) > maxCount {
			maxCount = len(weekData)
			maxWeek = week
		}
	}

	// Create a sample peak time for the peak week
	peakTime := time.Now()
	peakTime = peakTime.AddDate(0, 0, (maxWeek-1)*7) // Adjust to peak week
	peakTime = peakTime.Add(9 * time.Hour)           // 9 AM

	peakPeriods = append(peakPeriods, peakTime)

	return peakPeriods
}

func (ta *DefaultTrendAnalyzer) findQuarterlyPeakPeriods(monthGroups map[int][]*architecture.PaymentFailure) []time.Time {
	var peakPeriods []time.Time

	// Find the month with most events
	var maxMonth int
	maxCount := 0
	for month, monthData := range monthGroups {
		if len(monthData) > maxCount {
			maxCount = len(monthData)
			maxMonth = month
		}
	}

	// Create a sample peak time for the peak month
	peakTime := time.Now()
	peakTime = time.Date(peakTime.Year(), time.Month(maxMonth), 15, 9, 0, 0, 0, peakTime.Location())

	peakPeriods = append(peakPeriods, peakTime)

	return peakPeriods
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

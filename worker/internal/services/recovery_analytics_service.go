package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/lexure-intelligence/payment-watchdog/internal/models"
)

// RecoveryAnalyticsService provides analytics for payment recovery operations
type RecoveryAnalyticsService struct {
	db     *gorm.DB
	logger *zap.Logger
	// Add any additional dependencies like cache, metrics client, etc.
}

// RecoveryMetrics represents the metrics for payment recovery
type RecoveryMetrics struct {
	RecoveryRate          float64   `json:"recovery_rate"`
	AverageRecoveryTime   float64   `json:"average_recovery_time"`
	RecoveryByMethod      []Metric  `json:"recovery_by_method"`
	RecoveryByFailureType []Metric  `json:"recovery_by_failure_type"`
	RecoveryAmounts       Amounts   `json:"recovery_amounts"`
	RecoveryTrends        []Trend   `json:"recovery_trends"`
	RecoveryScore         int       `json:"recovery_score"`
	LastUpdated           time.Time `json:"last_updated"`
}

// Metric represents a key-value metric pair
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Amounts represents monetary amounts for different recovery states
type Amounts struct {
	TotalFailed    float64 `json:"total_failed"`
	TotalRecovered float64 `json:"total_recovered"`
	TotalPending   float64 `json:"total_pending"`
}

// Trend represents a data point in a time series
type Trend struct {
	Date    time.Time `json:"date"`
	Value   float64   `json:"value"`
	Success bool      `json:"success"`
}

// RecoveryPattern represents patterns in recovery attempts
type RecoveryPattern struct {
	TimeOfDay    map[string]int `json:"time_of_day"`
	DayOfWeek    map[string]int `json:"day_of_week"`
	CommonErrors []string       `json:"common_errors"`
}

// RecoveryAnalytics represents the complete analytics data
type RecoveryAnalytics struct {
	Metrics  RecoveryMetrics `json:"metrics"`
	Patterns RecoveryPattern `json:"patterns"`
}

// DetailedRecoveryMetrics represents detailed metrics for payment recovery
type DetailedRecoveryMetrics struct {
	RecoveryRate          float64          `json:"recovery_rate"`            // Percentage of failed payments that were successfully recovered
	AverageRecoveryTime   int64            `json:"average_recovery_time"`    // Average time to recover a payment in seconds
	RecoveryByMethod      map[string]int64 `json:"recovery_by_method"`       // Count of recoveries by method (e.g., auto-retry, manual)
	RecoveryByFailureType map[string]int64 `json:"recovery_by_failure_type"` // Count of recoveries by failure type
	TotalRecoveredAmount  float64          `json:"total_recovered_amount"`   // Total amount recovered in the period
	TotalFailedAmount     float64          `json:"total_failed_amount"`      // Total amount that failed in the period
}

// RecoveryTrend represents the trend of recovery metrics over time
type RecoveryTrend struct {
	TimePeriod    string  `json:"time_period"` // e.g., "2023-01", "2023-02"
	RecoveryRate  float64 `json:"recovery_rate"`
	RecoveryCount int64   `json:"recovery_count"`
	FailedCount   int64   `json:"failed_count"`
}

// DetailedRecoveryPattern represents detected patterns in payment recovery
type DetailedRecoveryPattern struct {
	PatternType  string  `json:"pattern_type"`  // e.g., "time_of_day", "day_of_week"
	PatternValue string  `json:"pattern_value"` // e.g., "09:00-12:00", "Monday"
	RecoveryRate float64 `json:"recovery_rate"` // Recovery rate for this pattern
	SampleSize   int64   `json:"sample_size"`   // Number of samples in this pattern
}

// NewRecoveryAnalyticsService creates a new instance of RecoveryAnalyticsService
func NewRecoveryAnalyticsService(db *gorm.DB, logger *zap.Logger) *RecoveryAnalyticsService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RecoveryAnalyticsService{
		db:     db,
		logger: logger.Named("recovery_analytics"),
	}
}

// getRecoveryRateByHour calculates recovery success rates by hour of day
func (s *RecoveryAnalyticsService) getRecoveryRateByHour(ctx context.Context, companyID string, startTime, endTime time.Time) (map[int]float64, error) {
	var results []struct {
		Hour  int
		Rate  float64
		Count int64
	}

	query := `
		WITH recovery_attempts AS (
			SELECT 
				EXTRACT(HOUR FROM created_at) as hour,
				COUNT(*) as total,
				SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as recovered
			FROM payment_failure_events
			WHERE company_id = ? 
			  AND created_at BETWEEN ? AND ?
			GROUP BY EXTRACT(HOUR FROM created_at)
		)
		SELECT 
			hour::int as hour,
			CASE WHEN total > 0 THEN (recovered::float / total) * 100 ELSE 0 END as rate,
			total as count
		FROM recovery_attempts
		ORDER BY hour
	`

	if err := s.db.Raw(query, companyID, startTime, endTime).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get recovery rates by hour: %w", err)
	}

	rates := make(map[int]float64)
	for _, r := range results {
		rates[r.Hour] = r.Rate
	}

	return rates, nil
}

// calculateRecoveryScore calculates an overall recovery performance score (0-100)
func (s *RecoveryAnalyticsService) calculateRecoveryScore(metrics *RecoveryMetrics) (int, error) {
	if metrics == nil {
		return 0, fmt.Errorf("metrics cannot be nil")
	}

	// Simple scoring algorithm - can be enhanced based on business requirements
	score := 0

	// Recovery rate (0-50 points)
	score += int(metrics.RecoveryRate * 0.5)

	// Recovery time (0-30 points) - lower is better
	averageRecoveryTime := time.Duration(metrics.AverageRecoveryTime) * time.Second
	if averageRecoveryTime < time.Hour {
		score += 30
	} else if averageRecoveryTime < 24*time.Hour {
		score += 20
	} else if averageRecoveryTime < 7*24*time.Hour {
		score += 10
	}

	// Recovery amount (0-20 points)
	if metrics.RecoveryAmounts.TotalFailed > 0 {
		recoveryRatio := metrics.RecoveryAmounts.TotalRecovered / metrics.RecoveryAmounts.TotalFailed
		score += int(recoveryRatio * 20)
	}

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return score, nil
}

// GetRecoveryMetrics returns recovery metrics for a company with comprehensive analytics
func (s *RecoveryAnalyticsService) GetRecoveryMetrics(ctx context.Context, companyID string, startTime, endTime time.Time) (*RecoveryMetrics, error) {
	// Start a new span for this operation
	ctx, span := otel.Tracer("recovery-analytics").Start(ctx, "GetRecoveryMetrics")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("company_id", companyID),
		attribute.String("start_time", startTime.Format(time.RFC3339)),
		attribute.String("end_time", endTime.Format(time.RFC3339)),
	)

	// Initialize metrics with default values
	metrics := &RecoveryMetrics{
		RecoveryByMethod:      []Metric{},
		RecoveryByFailureType: []Metric{},
		LastUpdated:           time.Now().UTC(),
	}

	// Use a wait group to handle concurrent database queries
	var wg sync.WaitGroup
	errCh := make(chan error, 3) // Buffer for 3 potential errors
	resultCh := make(chan interface{}, 3)

	// 1. Get failed payments data
	wg.Add(1)
	go func() {
		defer wg.Done()
		var failedPayments struct {
			Count int64
			Sum   float64
		}

		if err := s.db.WithContext(ctx).
			Model(&models.PaymentFailureEvent{}).
			Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as sum").
			Where("company_id = ? AND created_at BETWEEN ? AND ?",
				companyID, startTime, endTime).
			Scan(&failedPayments).Error; err != nil {
			errCh <- fmt.Errorf("failed to get failed payments: %w", err)
			return
		}

		resultCh <- failedPayments
	}()

	// 2. Get recovered payments data
	recoveredCh := make(chan []struct {
		Method      string
		FailureType string
		Count       int64
		Sum         float64
		AvgTime     float64
	}, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var recoveredPayments []struct {
			Method      string
			FailureType string
			Count       int64
			Sum         float64
			AvgTime     float64
		}

		query := `
			SELECT 
				COALESCE(provider, 'unknown') as method,
				COALESCE(failure_reason, 'unknown') as failure_type,
				COUNT(*) as count,
				COALESCE(SUM(amount), 0) as sum,
				0 as avg_time
			FROM payment_failure_events
			WHERE company_id = ? 
				AND created_at BETWEEN ? AND ?
				AND status = 'resolved'
			GROUP BY provider, failure_reason
		`

		if err := s.db.Raw(query, companyID, startTime, endTime).Scan(&recoveredPayments).Error; err != nil {
			errCh <- fmt.Errorf("failed to get recovered payments: %w", err)
			return
		}

		recoveredCh <- recoveredPayments
	}()

	// 3. Get additional metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Get recovery success rate by hour of day
		hourlySuccess, err := s.getRecoveryRateByHour(ctx, companyID, startTime, endTime)
		if err != nil {
			errCh <- fmt.Errorf("failed to get hourly recovery rates: %w", err)
			return
		}
		resultCh <- hourlySuccess
	}()

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errCh)
		close(resultCh)
	}()

	// Process results
	var failedPayments struct {
		Count int64
		Sum   float64
	}
	received := 0

	for result := range resultCh {
		switch v := result.(type) {
		case struct {
			Count int64
			Sum   float64
		}:
			failedPayments = v
			metrics.RecoveryAmounts.TotalFailed = v.Sum
		case map[int]float64: // hourly success rates
			// Process hourly success rates if needed
			hourlyRates := make([]Metric, 0, len(v))
			for hour, rate := range v {
				hourlyRates = append(hourlyRates, Metric{
					Name:  fmt.Sprintf("%02d:00", hour),
					Value: rate,
				})
			}
			sort.Slice(hourlyRates, func(i, j int) bool {
				return hourlyRates[i].Name < hourlyRates[j].Name
			})
			// Store in appropriate field or use as needed
		}
		received++
		if received >= 2 { // We expect 2 results (failed payments and hourly rates)
			break
		}
	}

	// Check for errors
	for err := range errCh {
		s.logger.Error("error in GetRecoveryMetrics",
			zap.String("company_id", companyID),
			zap.Error(err))
		span.RecordError(err)
		// Continue processing other metrics even if some fail
	}

	// Process recovered payments
	recoveredPayments := <-recoveredCh
	var totalRecovered int64
	var totalRecoveryTime float64
	var totalRecoveredAmount float64

	for _, r := range recoveredPayments {
		metrics.RecoveryByMethod = append(metrics.RecoveryByMethod, Metric{
			Name:  r.Method,
			Value: float64(r.Count),
		})
		metrics.RecoveryByFailureType = append(metrics.RecoveryByFailureType, Metric{
			Name:  r.FailureType,
			Value: float64(r.Count),
		})
		totalRecovered += r.Count
		totalRecoveryTime += r.AvgTime * float64(r.Count)
		totalRecoveredAmount += r.Sum
	}

	// Calculate metrics
	if failedPayments.Count > 0 {
		metrics.RecoveryRate = float64(totalRecovered) / float64(failedPayments.Count) * 100
	}
	if totalRecovered > 0 {
		metrics.AverageRecoveryTime = totalRecoveryTime / float64(totalRecovered)
	}
	metrics.RecoveryAmounts.TotalRecovered = totalRecoveredAmount

	// Calculate recovery score
	score, err := s.calculateRecoveryScore(metrics)
	if err != nil {
		s.logger.Warn("failed to calculate recovery score",
			zap.String("company_id", companyID),
			zap.Error(err))
		span.RecordError(err)
	} else {
		metrics.RecoveryScore = score
	}

	// Add metrics to span attributes for observability
	span.SetAttributes(
		attribute.Float64("recovery_rate", metrics.RecoveryRate),
		attribute.Float64("average_recovery_time_seconds", metrics.AverageRecoveryTime),
		attribute.Float64("total_recovered_amount", metrics.RecoveryAmounts.TotalRecovered),
		attribute.Float64("total_failed_amount", metrics.RecoveryAmounts.TotalFailed),
		attribute.Int("recovery_score", metrics.RecoveryScore),
	)

	return metrics, nil
}

// GetRecoveryTrends returns recovery trends over time
func (s *RecoveryAnalyticsService) GetRecoveryTrends(ctx context.Context, companyID string, period string, limit int) ([]RecoveryTrend, error) {
	var trends []RecoveryTrend

	// This is a simplified query - adjust based on your schema and time period grouping
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC(?, created_at), 'YYYY-MM-DD') as time_period,
			COUNT(*) as total_failures,
			SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as recovered_count
		FROM payment_failure_events
		WHERE company_id = ?
		GROUP BY time_period
		ORDER BY time_period DESC
		LIMIT ?
	`

	timeUnit := "day" // Default to daily
	switch period {
	case "week":
		timeUnit = "week"
	case "month":
		timeUnit = "month"
	case "quarter":
		timeUnit = "quarter"
	case "year":
		timeUnit = "year"
	}

	rows, err := s.db.Raw(query, timeUnit, companyID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trend RecoveryTrend
		var total, recovered int64
		if err := rows.Scan(&trend.TimePeriod, &total, &recovered); err != nil {
			return nil, err
		}
		trend.FailedCount = total
		trend.RecoveryCount = recovered
		if total > 0 {
			trend.RecoveryRate = float64(recovered) / float64(total) * 100
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

// DetectRecoveryPatterns analyzes and detects patterns in recovery success
func (s *RecoveryAnalyticsService) DetectRecoveryPatterns(ctx context.Context, companyID string, startTime, endTime time.Time) ([]RecoveryPattern, error) {
	var patterns []RecoveryPattern

	// Example: Detect patterns by time of day
	timeOfDayPatterns, err := s.analyzeTimeOfDayPatterns(ctx, companyID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, timeOfDayPatterns...)

	// Example: Detect patterns by day of week
	dayOfWeekPatterns, err := s.analyzeDayOfWeekPatterns(ctx, companyID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, dayOfWeekPatterns...)

	return patterns, nil
}

// analyzeTimeOfDayPatterns detects recovery patterns by time of day
func (s *RecoveryAnalyticsService) analyzeTimeOfDayPatterns(ctx context.Context, companyID string, startTime, endTime time.Time) ([]RecoveryPattern, error) {
	var patterns []RecoveryPattern

	// This is a simplified query - adjust based on your schema
	query := `
		WITH time_slots AS (
			SELECT 
				CASE 
					WHEN EXTRACT(HOUR FROM created_at) BETWEEN 0 AND 5 THEN '00:00-06:00'
					WHEN EXTRACT(HOUR FROM created_at) BETWEEN 6 AND 11 THEN '06:00-12:00'
					WHEN EXTRACT(HOUR FROM created_at) BETWEEN 12 AND 17 THEN '12:00-18:00'
					ELSE '18:00-00:00'
			END as time_slot,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0) as recovered
		FROM payment_failure_events
		WHERE company_id = ? AND created_at BETWEEN ? AND ?
		GROUP BY time_slot
		)
		SELECT 
			time_slot as pattern_value,
			recovered as recovery_count,
			total as sample_size
		FROM time_slots
		WHERE total > 0
		ORDER BY recovery_count DESC
	`

	rows, err := s.db.Raw(query, companyID, startTime, endTime).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var patternValue string
		var recovered, total int64
		if err := rows.Scan(&patternValue, &recovered, &total); err != nil {
			return nil, err
		}
		p := RecoveryPattern{
			TimeOfDay: map[string]int{patternValue: int(recovered)},
		}
		if total > 0 {
			// Store recovery rate in CommonErrors for now (field mismatch)
			p.CommonErrors = append(p.CommonErrors, fmt.Sprintf("rate_%s_%.1f", patternValue, float64(recovered)/float64(total)*100))
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}

// analyzeDayOfWeekPatterns detects recovery patterns by day of week
func (s *RecoveryAnalyticsService) analyzeDayOfWeekPatterns(ctx context.Context, companyID string, startTime, endTime time.Time) ([]RecoveryPattern, error) {
	var patterns []RecoveryPattern

	// This is a simplified query - adjust based on your schema
	query := `
		WITH day_stats AS (
			SELECT 
				TO_CHAR(created_at, 'Day') as day_of_week,
				COUNT(*) as total,
				SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0) as recovered
			FROM payment_failure_events
			WHERE company_id = ? AND created_at BETWEEN ? AND ?
			GROUP BY day_of_week
		)
		SELECT 
			TRIM(day_of_week) as pattern_value,
			recovered as recovery_count,
			total as sample_size
		FROM day_stats
		WHERE total > 0
		ORDER BY 
			CASE TRIM(day_of_week)
				WHEN 'Monday' THEN 1
				WHEN 'Tuesday' THEN 2
				WHEN 'Wednesday' THEN 3
				WHEN 'Thursday' THEN 4
				WHEN 'Friday' THEN 5
				WHEN 'Saturday' THEN 6
				WHEN 'Sunday' THEN 7
			END
	`

	rows, err := s.db.Raw(query, companyID, startTime, endTime).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var patternValue string
		var recovered, total int64
		if err := rows.Scan(&patternValue, &recovered, &total); err != nil {
			return nil, err
		}
		p := RecoveryPattern{
			DayOfWeek: map[string]int{patternValue: int(recovered)},
		}
		if total > 0 {
			// Store recovery rate in CommonErrors for now (field mismatch)
			p.CommonErrors = append(p.CommonErrors, fmt.Sprintf("rate_%s_%.1f", patternValue, float64(recovered)/float64(total)*100))
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}

// GetRecoveryPerformanceScore calculates an overall recovery performance score (0-100)
func (s *RecoveryAnalyticsService) GetRecoveryPerformanceScore(ctx context.Context, companyID string) (int, error) {
	// Get metrics for the last 30 days
	now := time.Now()
	startTime := now.AddDate(0, -1, 0)

	metrics, err := s.GetRecoveryMetrics(ctx, companyID, startTime, now)
	if err != nil {
		return 0, err
	}

	// Calculate score based on recovery rate and other factors
	score := int(metrics.RecoveryRate) // Base score is the recovery rate

	// Adjust score based on recovery time (faster is better)
	averageRecoveryTimeHours := float64(metrics.AverageRecoveryTime) / 3600 // Convert seconds to hours
	if averageRecoveryTimeHours < 1 {
		score += 10 // Bonus for very fast recovery
	} else if averageRecoveryTimeHours < 24 {
		score += 5 // Small bonus for same-day recovery
	}

	// Cap the score at 100
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return score, nil
}

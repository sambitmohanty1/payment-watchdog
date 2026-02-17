package analytics

import (
	"sync"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/architecture"
	"go.uber.org/zap"
)

// AnalyticsEngine represents the core analytics engine for payment failure analysis
type AnalyticsEngine struct {
	patternDetector  PatternDetector
	trendAnalyzer    TrendAnalyzer
	failurePredictor FailurePredictor
	metrics          *AnalyticsMetrics
	logger           *zap.Logger
	mutex            sync.RWMutex
}

// PatternDetector interface for detecting patterns in payment failures
type PatternDetector interface {
	DetectPatterns(events []*architecture.PaymentFailure) []Pattern
	DetectCustomerPatterns(customerID string, events []*architecture.PaymentFailure) []CustomerPattern
	DetectTemporalPatterns(events []*architecture.PaymentFailure, timeRange time.Duration) []TemporalPattern
	DetectBusinessPatterns(events []*architecture.PaymentFailure) []Pattern
}

// TrendAnalyzer interface for analyzing trends in payment failures
type TrendAnalyzer interface {
	AnalyzeTrends(events []*architecture.PaymentFailure, timeRange time.Duration) []Trend
	AnalyzeSeasonalPatterns(events []*architecture.PaymentFailure) []SeasonalPattern
	AnalyzeBusinessCyclePatterns(events []*architecture.PaymentFailure) []BusinessCyclePattern
}

// FailurePredictor interface for predicting future payment failures
type FailurePredictor interface {
	PredictFailure(customerID string, history []*architecture.PaymentFailure) *Prediction
	PredictRiskScore(customerID string, history []*architecture.PaymentFailure) float64
	PredictNextFailureDate(customerID string, history []*architecture.PaymentFailure) *time.Time
}

// Pattern represents a detected pattern in payment failures
type Pattern struct {
	ID          string                 `json:"id"`
	Type        PatternType            `json:"type"`
	Confidence  float64                `json:"confidence"`
	Description string                 `json:"description"`
	Evidence    []string               `json:"evidence"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PatternType represents the type of pattern detected
type PatternType string

const (
	PatternTypeRecurring PatternType = "recurring"
	PatternTypeSeasonal  PatternType = "seasonal"
	PatternTypeBusiness  PatternType = "business"
	PatternTypeAmount    PatternType = "amount"
	PatternTypeTime      PatternType = "time"
	PatternTypeDayOfWeek PatternType = "day_of_week"
	PatternTypeTimeOfDay PatternType = "time_of_day"
)

// Trend represents a trend in payment failure data
type Trend struct {
	ID          string         `json:"id"`
	Type        TrendType      `json:"type"`
	Direction   TrendDirection `json:"direction"`
	Magnitude   float64        `json:"magnitude"`
	Confidence  float64        `json:"confidence"`
	TimeRange   time.Duration  `json:"time_range"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	Evidence    []string       `json:"evidence"`
}

// TrendType represents the type of trend
type TrendType string

const (
	TrendTypeFailureRate TrendType = "failure_rate"
	TrendTypeAmount      TrendType = "amount"
	TrendTypeFrequency   TrendType = "frequency"
	TrendTypeCustomer    TrendType = "customer"
)

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendDirectionIncreasing TrendDirection = "increasing"
	TrendDirectionDecreasing TrendDirection = "decreasing"
	TrendDirectionStable     TrendDirection = "stable"
	TrendDirectionCyclical   TrendDirection = "cyclical"
)

// SeasonalPattern represents a seasonal pattern in payment failures
type SeasonalPattern struct {
	Pattern    Pattern `json:"pattern"`
	Season     string  `json:"season"`
	Year       int     `json:"year"`
	Strength   float64 `json:"strength"`
	PeakMonths []int   `json:"peak_months"`
	PeakDays   []int   `json:"peak_days"`
}

// BusinessCyclePattern represents a business cycle pattern
type BusinessCyclePattern struct {
	Pattern     Pattern       `json:"pattern"`
	CycleType   string        `json:"cycle_type"`
	Duration    time.Duration `json:"duration"`
	Strength    float64       `json:"strength"`
	PeakPeriods []time.Time   `json:"peak_periods"`
	CycleLength time.Duration `json:"cycle_length"`
}

// CustomerPattern represents a pattern specific to a customer
type CustomerPattern struct {
	CustomerID  string  `json:"customer_id"`
	Pattern     Pattern `json:"pattern"`
	Frequency   float64 `json:"frequency"`
	TotalAmount float64 `json:"total_amount"`
	RiskLevel   string  `json:"risk_level"`
}

// TemporalPattern represents a time-based pattern
type TemporalPattern struct {
	Pattern     Pattern       `json:"pattern"`
	TimeRange   time.Duration `json:"time_range"`
	Frequency   float64       `json:"frequency"`
	PeakTimes   []time.Time   `json:"peak_times"`
	Seasonality string        `json:"seasonality"`
}

// Prediction represents a prediction of future payment failures
type Prediction struct {
	ID                 string     `json:"id"`
	CustomerID         string     `json:"customer_id"`
	RiskScore          float64    `json:"risk_score"`
	FailureProbability float64    `json:"failure_probability"`
	NextFailureDate    *time.Time `json:"next_failure_date"`
	Confidence         float64    `json:"confidence"`
	Factors            []string   `json:"factors"`
	CreatedAt          time.Time  `json:"created_at"`
	ExpiresAt          time.Time  `json:"expires_at"`
}

// AnalyticsMetrics tracks performance and usage metrics
type AnalyticsMetrics struct {
	PatternsDetected     int64         `json:"patterns_detected"`
	TrendsAnalyzed       int64         `json:"trends_analyzed"`
	PredictionsMade      int64         `json:"predictions_made"`
	ProcessingTime       time.Duration `json:"processing_time"`
	LastAnalysisTime     time.Time     `json:"last_analysis_time"`
	TotalEventsProcessed int64         `json:"total_events_processed"`
}

// AnalysisResult represents the result of a comprehensive analysis
type AnalysisResult struct {
	Patterns    []Pattern
	Trends      []Trend
	Predictions []*Prediction
	Metrics     *AnalyticsMetrics
	CreatedAt   time.Time
}

// NewAnalyticsEngine creates a new analytics engine instance
func NewAnalyticsEngine(
	patternDetector PatternDetector,
	trendAnalyzer TrendAnalyzer,
	failurePredictor FailurePredictor,
	logger *zap.Logger,
) *AnalyticsEngine {
	return &AnalyticsEngine{
		patternDetector:  patternDetector,
		trendAnalyzer:    trendAnalyzer,
		failurePredictor: failurePredictor,
		metrics:          &AnalyticsMetrics{},
		logger:           logger,
	}
}

// AnalyzePaymentFailures performs comprehensive analysis of payment failure events
func (ae *AnalyticsEngine) AnalyzePaymentFailures(events []*architecture.PaymentFailure) (*AnalysisResult, error) {
	ae.mutex.Lock()
	defer ae.mutex.Unlock()

	startTime := time.Now()
	ae.logger.Info("Starting payment failure analysis", zap.Int("event_count", len(events)))

	// Detect patterns
	patterns := ae.patternDetector.DetectPatterns(events)
	ae.metrics.PatternsDetected = int64(len(patterns))

	// Analyze trends
	trends := ae.trendAnalyzer.AnalyzeTrends(events, time.Hour*24) // Default 24-hour time range
	ae.metrics.TrendsAnalyzed = int64(len(trends))

	// Generate predictions for each unique customer
	customerIDs := ae.extractUniqueCustomers(events)
	predictions := make([]*Prediction, 0)

	for _, customerID := range customerIDs {
		customerEvents := ae.filterEventsByCustomer(events, customerID)
		if len(customerEvents) > 0 {
			prediction := ae.failurePredictor.PredictFailure(customerID, customerEvents)
			if prediction != nil {
				predictions = append(predictions, prediction)
			}
		}
	}
	ae.metrics.PredictionsMade = int64(len(predictions))

	// Calculate processing time
	processingTime := time.Since(startTime)
	ae.metrics.ProcessingTime = processingTime
	ae.metrics.LastAnalysisTime = time.Now()
	ae.metrics.TotalEventsProcessed = int64(len(events))

	ae.logger.Info("Analysis completed",
		zap.Int("patterns", len(patterns)),
		zap.Int("trends", len(trends)),
		zap.Int("predictions", len(predictions)),
		zap.Duration("processing_time", processingTime))

	return &AnalysisResult{
		Patterns:    patterns,
		Trends:      trends,
		Predictions: predictions,
		Metrics:     ae.metrics,
		CreatedAt:   time.Now(),
	}, nil
}

// extractUniqueCustomers extracts unique customer IDs from events
func (ae *AnalyticsEngine) extractUniqueCustomers(events []*architecture.PaymentFailure) []string {
	customerSet := make(map[string]bool)

	for _, event := range events {
		if event.CustomerID != "" {
			customerSet[event.CustomerID] = true
		}
	}

	customerIDs := make([]string, 0, len(customerSet))
	for customerID := range customerSet {
		customerIDs = append(customerIDs, customerID)
	}

	return customerIDs
}

// filterEventsByCustomer filters events for a specific customer
func (ae *AnalyticsEngine) filterEventsByCustomer(events []*architecture.PaymentFailure, customerID string) []*architecture.PaymentFailure {
	filtered := make([]*architecture.PaymentFailure, 0)

	for _, event := range events {
		if event.CustomerID == customerID {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetMetrics returns the current analytics metrics
func (ae *AnalyticsEngine) GetMetrics() *AnalyticsMetrics {
	ae.mutex.RLock()
	defer ae.mutex.RUnlock()
	return ae.metrics
}

// ResetMetrics resets all metrics to zero
func (ae *AnalyticsEngine) ResetMetrics() {
	ae.mutex.Lock()
	defer ae.mutex.Unlock()
	ae.metrics = &AnalyticsMetrics{}
}

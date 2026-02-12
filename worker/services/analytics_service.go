package services

import (
	"context"
	"fmt"
	"time"

	"github.com/lexure-intelligence/payment-watchdog/internal/analytics"
	"github.com/lexure-intelligence/payment-watchdog/internal/architecture"
	"github.com/lexure-intelligence/payment-watchdog/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AnalyticsService provides analytics capabilities for payment failures
type AnalyticsService struct {
	db               *gorm.DB
	analyticsEngine  *analytics.AnalyticsEngine
	patternDetector  analytics.PatternDetector
	trendAnalyzer    analytics.TrendAnalyzer
	failurePredictor analytics.FailurePredictor
	logger           *zap.Logger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB, logger *zap.Logger) *AnalyticsService {

	// Initialize analytics components
	patternDetector := analytics.NewDefaultPatternDetector(logger)
	trendAnalyzer := analytics.NewDefaultTrendAnalyzer(logger)
	failurePredictor := analytics.NewDefaultFailurePredictor(logger)

	// Create analytics engine
	analyticsEngine := analytics.NewAnalyticsEngine(
		patternDetector,
		trendAnalyzer,
		failurePredictor,
		logger,
	)

	return &AnalyticsService{
		db:               db,
		analyticsEngine:  analyticsEngine,
		patternDetector:  patternDetector,
		trendAnalyzer:    trendAnalyzer,
		failurePredictor: failurePredictor,
		logger:           logger,
	}
}

// AnalyzeCompanyPaymentFailures performs comprehensive analysis for a specific company
func (s *AnalyticsService) AnalyzeCompanyPaymentFailures(ctx context.Context, companyID string, timeRange time.Duration) (*analytics.AnalysisResult, error) {
	s.logger.Info("Starting company payment failure analysis",
		zap.String("company_id", companyID),
		zap.Duration("time_range", timeRange))

	// Fetch payment failures from database
	failures, err := s.getCompanyPaymentFailures(ctx, companyID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment failures: %w", err)
	}

	if len(failures) == 0 {
		s.logger.Info("No payment failures found for analysis",
			zap.String("company_id", companyID))
		return &analytics.AnalysisResult{
			Patterns:    []analytics.Pattern{},
			Trends:      []analytics.Trend{},
			Predictions: []*analytics.Prediction{},
			Metrics:     &analytics.AnalyticsMetrics{},
			CreatedAt:   time.Now(),
		}, nil
	}

	// Convert to architecture.PaymentFailure for analytics engine
	paymentFailures := s.convertToPaymentFailures(failures)

	// Perform analysis
	result, err := s.analyticsEngine.AnalyzePaymentFailures(paymentFailures)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze payment failures: %w", err)
	}

	s.logger.Info("Company payment failure analysis completed",
		zap.String("company_id", companyID),
		zap.Int("patterns", len(result.Patterns)),
		zap.Int("trends", len(result.Trends)),
		zap.Int("predictions", len(result.Predictions)))

	return result, nil
}

// AnalyzeCustomerPaymentFailures performs analysis for a specific customer
func (s *AnalyticsService) AnalyzeCustomerPaymentFailures(ctx context.Context, companyID, customerID string, timeRange time.Duration) (*analytics.AnalysisResult, error) {
	s.logger.Info("Starting customer payment failure analysis",
		zap.String("company_id", companyID),
		zap.String("customer_id", customerID),
		zap.Duration("time_range", timeRange))

	// Fetch customer-specific payment failures
	failures, err := s.getCustomerPaymentFailures(ctx, companyID, customerID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer payment failures: %w", err)
	}

	if len(failures) == 0 {
		s.logger.Info("No payment failures found for customer",
			zap.String("company_id", companyID),
			zap.String("customer_id", customerID))
		return &analytics.AnalysisResult{
			Patterns:    []analytics.Pattern{},
			Trends:      []analytics.Trend{},
			Predictions: []*analytics.Prediction{},
			Metrics:     &analytics.AnalyticsMetrics{},
			CreatedAt:   time.Now(),
		}, nil
	}

	// Convert to architecture.PaymentFailure for analytics engine
	paymentFailures := s.convertToPaymentFailures(failures)

	// Perform analysis
	result, err := s.analyticsEngine.AnalyzePaymentFailures(paymentFailures)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze customer payment failures: %w", err)
	}

	s.logger.Info("Customer payment failure analysis completed",
		zap.String("company_id", companyID),
		zap.String("customer_id", customerID),
		zap.Int("patterns", len(result.Patterns)),
		zap.Int("trends", len(result.Trends)),
		zap.Int("predictions", len(result.Predictions)))

	return result, nil
}

// GetCustomerRiskScore returns the current risk score for a customer
func (s *AnalyticsService) GetCustomerRiskScore(ctx context.Context, companyID, customerID string) (float64, error) {
	s.logger.Info("Getting customer risk score",
		zap.String("company_id", companyID),
		zap.String("customer_id", customerID))

	// Fetch customer payment failures (last 90 days for risk assessment)
	timeRange := 90 * 24 * time.Hour
	failures, err := s.getCustomerPaymentFailures(ctx, companyID, customerID, timeRange)
	if err != nil {
		return 0.0, fmt.Errorf("failed to fetch customer payment failures: %w", err)
	}

	if len(failures) == 0 {
		return 0.0, nil // No failures = no risk
	}

	// Convert to architecture.PaymentFailure for risk calculation
	paymentFailures := s.convertToPaymentFailures(failures)

	// Calculate risk score
	riskScore := s.failurePredictor.PredictRiskScore(customerID, paymentFailures)

	s.logger.Info("Customer risk score calculated",
		zap.String("company_id", companyID),
		zap.String("customer_id", customerID),
		zap.Float64("risk_score", riskScore))

	return riskScore, nil
}

// GetCompanyAnalyticsSummary returns a summary of analytics for a company
func (s *AnalyticsService) GetCompanyAnalyticsSummary(ctx context.Context, companyID string) (*CompanyAnalyticsSummary, error) {
	s.logger.Info("Getting company analytics summary",
		zap.String("company_id", companyID))

	// Analyze last 30 days
	timeRange := 30 * 24 * time.Hour
	_, err := s.AnalyzeCompanyPaymentFailures(ctx, companyID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze company payment failures: %w", err)
	}

	// Get metrics
	metrics := s.analyticsEngine.GetMetrics()

	summary := &CompanyAnalyticsSummary{
		CompanyID:            companyID,
		AnalysisPeriod:       timeRange,
		TotalPaymentFailures: metrics.TotalEventsProcessed,
		PatternsDetected:     metrics.PatternsDetected,
		TrendsAnalyzed:       metrics.TrendsAnalyzed,
		PredictionsMade:      metrics.PredictionsMade,
		LastAnalysisTime:     metrics.LastAnalysisTime,
		ProcessingTime:       metrics.ProcessingTime,
		HighRiskCustomers:    s.countHighRiskCustomers(ctx, companyID),
		TopFailureReasons:    s.getTopFailureReasons(ctx, companyID),
		CreatedAt:            time.Now(),
	}

	s.logger.Info("Company analytics summary generated",
		zap.String("company_id", companyID),
		zap.Int64("total_failures", summary.TotalPaymentFailures),
		zap.Int64("patterns", summary.PatternsDetected),
		zap.Int64("trends", summary.TrendsAnalyzed))

	return summary, nil
}

// CompanyAnalyticsSummary represents a summary of analytics for a company
type CompanyAnalyticsSummary struct {
	CompanyID            string          `json:"company_id"`
	AnalysisPeriod       time.Duration   `json:"analysis_period"`
	TotalPaymentFailures int64           `json:"total_payment_failures"`
	PatternsDetected     int64           `json:"patterns_detected"`
	TrendsAnalyzed       int64           `json:"trends_analyzed"`
	PredictionsMade      int64           `json:"predictions_made"`
	LastAnalysisTime     time.Time       `json:"last_analysis_time"`
	ProcessingTime       time.Duration   `json:"processing_time"`
	HighRiskCustomers    int64           `json:"high_risk_customers"`
	TopFailureReasons    []FailureReason `json:"top_failure_reasons"`
	CreatedAt            time.Time       `json:"created_at"`
}

// FailureReason represents a failure reason with count
type FailureReason struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// Helper methods

func (s *AnalyticsService) getCompanyPaymentFailures(ctx context.Context, companyID string, timeRange time.Duration) ([]models.PaymentFailureEvent, error) {
	s.logger.Info("🔍 ANALYTICS DEBUG: getCompanyPaymentFailures called",
		zap.String("company_id", companyID),
		zap.Duration("time_range", timeRange),
		zap.String("db_type", fmt.Sprintf("%T", s.db)),
		zap.Bool("db_is_nil", s.db == nil))

	var failures []models.PaymentFailureEvent

	startTime := time.Now().Add(-timeRange)

	s.logger.Info("🔍 ANALYTICS DEBUG: About to execute database query",
		zap.String("query", "SELECT * FROM payment_failure_events WHERE company_id = ? AND created_at >= ?"),
		zap.String("company_id", companyID),
		zap.Time("start_time", startTime))

	err := s.db.Where("company_id = ? AND created_at >= ?", companyID, startTime).
		Order("created_at DESC").
		Find(&failures).Error

	return failures, err
}

func (s *AnalyticsService) getCustomerPaymentFailures(ctx context.Context, companyID, customerID string, timeRange time.Duration) ([]models.PaymentFailureEvent, error) {
	var failures []models.PaymentFailureEvent

	startTime := time.Now().Add(-timeRange)

	err := s.db.Where("company_id = ? AND customer_id = ? AND created_at >= ?", companyID, customerID, startTime).
		Order("created_at DESC").
		Find(&failures).Error

	return failures, err
}

func (s *AnalyticsService) convertToPaymentFailures(events []models.PaymentFailureEvent) []*architecture.PaymentFailure {
	paymentFailures := make([]*architecture.PaymentFailure, len(events))

	for i, event := range events {
		paymentFailures[i] = &architecture.PaymentFailure{
			ID:                event.ID,
			CompanyID:         event.CompanyID,
			ProviderID:        event.ProviderID,
			ProviderEventID:   event.EventID,
			ProviderEventType: event.EventType,
			Amount:            event.Amount,
			Currency:          event.Currency,
			PaymentMethod:     "", // Not available in model
			CustomerID:        event.CustomerID,
			CustomerName:      event.CustomerName,
			CustomerEmail:     event.CustomerEmail,
			CustomerPhone:     "", // Not available in model
			FailureReason:     event.FailureReason,
			FailureCode:       event.FailureCode,
			FailureMessage:    event.FailureMessage,
			InvoiceID:         "",  // Not available in model
			InvoiceNumber:     "",  // Not available in model
			DueDate:           nil, // Not available in model
			BusinessCategory:  "",  // Not available in model
			Status:            architecture.PaymentFailureStatus(event.Status),
			Priority:          architecture.PaymentFailurePriority("medium"), // Default priority
			RiskScore:         0.0,                                           // Not available in model
			OccurredAt:        event.CreatedAt,
			DetectedAt:        event.CreatedAt,
			ProcessedAt:       event.ProcessedAt,
			ResolvedAt:        nil, // Not available in model
			CreatedAt:         event.CreatedAt,
			UpdatedAt:         event.UpdatedAt,
		}
	}

	return paymentFailures
}

func (s *AnalyticsService) countHighRiskCustomers(ctx context.Context, companyID string) int64 {
	// For Sprint 3, we don't have risk_score column yet, so return 0
	// This will be implemented in future sprints when risk scoring is added
	return 0
}

func (s *AnalyticsService) getTopFailureReasons(ctx context.Context, companyID string) []FailureReason {

	var reasons []FailureReason

	// Get top 5 failure reasons with counts

	err := s.db.Model(&models.PaymentFailureEvent{}).
		Select("failure_reason as reason, COUNT(*) as count").
		Where("company_id = ? AND failure_reason IS NOT NULL", companyID).
		Group("failure_reason").
		Order("count DESC").
		Limit(5).
		Scan(&reasons).Error

	if err != nil {
		// Log error but return empty slice
		return []FailureReason{}
	}

	return reasons
}

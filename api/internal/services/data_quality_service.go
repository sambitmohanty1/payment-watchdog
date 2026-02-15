package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataQualityMetrics holds data quality assessment results
type DataQualityMetrics struct {
	TotalRecords     int64
	ValidRecords     int64
	InvalidRecords   int64
	MissingData      int64
	DuplicateRecords int64
	OutlierRecords   int64
	
	// Quality scores (0-100)
	OverallQualityScore float64
	CompletenessScore   float64
	AccuracyScore       float64
	ConsistencyScore    float64
	
	// Issues and recommendations
	Issues          []string
	Recommendations []string
}

// DataQualityReport represents a data quality report
type DataQualityReport struct {
	ID                string    `json:"id"`
	CompanyID         string    `json:"company_id"`
	ReportType        string    `json:"report_type"`
	GeneratedAt       time.Time `json:"generated_at"`
	
	// Quality metrics
	TotalRecords     int64   `json:"total_records"`
	ValidRecords     int64   `json:"valid_records"`
	InvalidRecords   int64   `json:"invalid_records"`
	MissingData      int64   `json:"missing_data"`
	DuplicateRecords int64   `json:"duplicate_records"`
	OutlierRecords   int64   `json:"outlier_records"`
	
	// Quality scores
	OverallQualityScore float64 `json:"overall_quality_score"`
	CompletenessScore   float64 `json:"completeness_score"`
	AccuracyScore       float64 `json:"accuracy_score"`
	ConsistencyScore    float64 `json:"consistency_score"`
	
	// Issues and recommendations
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// DataQualityService handles data quality assessment and reporting
type DataQualityService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDataQualityService creates a new data quality service
func NewDataQualityService(db *gorm.DB, logger *zap.Logger) *DataQualityService {
	return &DataQualityService{
		db:     db,
		logger: logger,
	}
}

// GenerateQualityReport generates a data quality report for a company
func (s *DataQualityService) GenerateQualityReport(ctx context.Context, companyID, reportType string) (*DataQualityReport, error) {
	s.logger.Info("Generating data quality report", 
		zap.String("company_id", companyID),
		zap.String("report_type", reportType))
	
	// Get payment failure data for the company
	var paymentFailures []models.PaymentFailureEvent
	if err := s.db.Where("company_id = ?", companyID).Find(&paymentFailures).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment failure data: %w", err)
	}
	
	if len(paymentFailures) == 0 {
		s.logger.Warn("No payment failure data found for company", zap.String("company_id", companyID))
		return &DataQualityReport{
			CompanyID:         companyID,
			ReportType:        reportType,
			GeneratedAt:       time.Now(),
			OverallQualityScore: 0,
		}, nil
	}
	
	// Calculate quality metrics
	metrics := s.calculateQualityMetrics(paymentFailures)
	
	// Generate quality report
	report := &DataQualityReport{
		CompanyID:  companyID,
		ReportType: reportType,
		GeneratedAt: time.Now(),
		
		// Quality metrics
		TotalRecords:     metrics.TotalRecords,
		ValidRecords:     metrics.ValidRecords,
		InvalidRecords:   metrics.InvalidRecords,
		MissingData:      metrics.MissingData,
		DuplicateRecords: metrics.DuplicateRecords,
		OutlierRecords:   metrics.OutlierRecords,
		
		// Quality scores
		OverallQualityScore: metrics.OverallQualityScore,
		CompletenessScore:   metrics.CompletenessScore,
		AccuracyScore:       metrics.AccuracyScore,
		ConsistencyScore:    metrics.ConsistencyScore,
		
		// Issues and recommendations
		Issues:          metrics.Issues,
		Recommendations: metrics.Recommendations,
	}
	
	s.logger.Info("Data quality report generated successfully",
		zap.String("company_id", companyID),
		zap.Float64("overall_score", metrics.OverallQualityScore))
	
	return report, nil
}

// calculateQualityMetrics calculates quality metrics from payment failure data
func (s *DataQualityService) calculateQualityMetrics(failures []models.PaymentFailureEvent) *DataQualityMetrics {
	metrics := &DataQualityMetrics{
		TotalRecords: int64(len(failures)),
		Issues:       make([]string, 0),
		Recommendations: make([]string, 0),
	}
	
	if len(failures) == 0 {
		return metrics
	}
	
	// Check for missing data
	for _, failure := range failures {
		if failure.CustomerEmail == "" {
			metrics.MissingData++
		}
		if failure.FailureReason == "" {
			metrics.MissingData++
		}
		if failure.Amount <= 0 {
			metrics.InvalidRecords++
		}
	}
	
	// Check for duplicates (same event_id)
	eventIDs := make(map[string]bool)
	for _, failure := range failures {
		if eventIDs[failure.EventID] {
			metrics.DuplicateRecords++
		} else {
			eventIDs[failure.EventID] = true
		}
	}
	
	// Calculate valid records
	metrics.ValidRecords = metrics.TotalRecords - metrics.InvalidRecords - metrics.DuplicateRecords
	
	// Calculate quality scores
	metrics.CompletenessScore = float64(metrics.ValidRecords) / float64(metrics.TotalRecords) * 100
	metrics.AccuracyScore = float64(metrics.ValidRecords) / float64(metrics.TotalRecords) * 100
	metrics.ConsistencyScore = float64(metrics.ValidRecords) / float64(metrics.TotalRecords) * 100
	
	// Overall quality score (average of all scores)
	metrics.OverallQualityScore = (metrics.CompletenessScore + metrics.AccuracyScore + metrics.ConsistencyScore) / 3
	
	// Generate issues and recommendations
	if metrics.MissingData > 0 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("%d records have missing critical data", metrics.MissingData))
		metrics.Recommendations = append(metrics.Recommendations, "Ensure all webhook events include customer email and failure reason")
	}
	
	if metrics.DuplicateRecords > 0 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("%d duplicate events detected", metrics.DuplicateRecords))
		metrics.Recommendations = append(metrics.Recommendations, "Implement webhook idempotency to prevent duplicate processing")
	}
	
	if metrics.OverallQualityScore < 80 {
		metrics.Recommendations = append(metrics.Recommendations, "Review webhook data quality and implement validation")
	}
	
	return metrics
}

// GenerateDailyQualityReport generates daily quality reports for all companies
func (s *DataQualityService) GenerateDailyQualityReport(ctx context.Context) error {
	s.logger.Info("Starting daily quality report generation")
	
	// Get all companies with payment failure data
	var companies []string
	if err := s.db.Model(&models.PaymentFailureEvent{}).
		Distinct("company_id").
		Pluck("company_id", &companies).Error; err != nil {
		return fmt.Errorf("failed to get companies: %w", err)
	}
	
	// Generate reports for each company
	for _, companyID := range companies {
		if _, err := s.GenerateQualityReport(ctx, companyID, "daily"); err != nil {
			s.logger.Error("Failed to generate daily report for company",
				zap.String("company_id", companyID), zap.Error(err))
		}
	}
	
	s.logger.Info("Daily quality report generation completed", zap.Int("companies_processed", len(companies)))
	return nil
}

// GetQualityTrends gets quality trends for a company over time
func (s *DataQualityService) GetQualityTrends(ctx context.Context, companyID string, days int) ([]map[string]interface{}, error) {
	s.logger.Info("Getting quality trends", zap.String("company_id", companyID), zap.Int("days", days))
	
	// Get data for the last N days
	startDate := time.Now().AddDate(0, 0, -days)
	
	var failures []models.PaymentFailureEvent
	if err := s.db.Where("company_id = ? AND created_at >= ?", companyID, startDate).
		Order("created_at ASC").Find(&failures).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment failure data: %w", err)
	}
	
	// Group by day and calculate daily quality scores
	dailyScores := make(map[string]*DataQualityMetrics)
	
	for _, failure := range failures {
		date := failure.CreatedAt.Format("2006-01-02")
		if dailyScores[date] == nil {
			dailyScores[date] = &DataQualityMetrics{}
		}
		dailyScores[date].TotalRecords++
		
		// Simple quality check
		if failure.CustomerEmail != "" && failure.FailureReason != "" && failure.Amount > 0 {
			dailyScores[date].ValidRecords++
		} else {
			dailyScores[date].InvalidRecords++
		}
	}
	
	// Convert to response format
	var trends []map[string]interface{}
	for date, metrics := range dailyScores {
		if metrics.TotalRecords > 0 {
			qualityScore := float64(metrics.ValidRecords) / float64(metrics.TotalRecords) * 100
			trends = append(trends, map[string]interface{}{
				"date":           date,
				"total_records":  metrics.TotalRecords,
				"valid_records":  metrics.ValidRecords,
				"quality_score":  qualityScore,
			})
		}
	}
	
	return trends, nil
}

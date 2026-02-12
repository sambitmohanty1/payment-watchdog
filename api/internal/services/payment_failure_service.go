/*
 * IMPORTANT: This file should be created in github.com/payment-watchdog/internal/services/
 * NOT in the parent ComplyFlow/internal/ directory
 *
 * File Location: github.com/payment-watchdog/internal/services/payment_failure_service.go
 *
 * This service manages payment failure events and provides comprehensive querying capabilities
 * for the Payment Failure Intelligence Service.
 */

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/payment-watchdog/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentFailureService manages payment failure events
type PaymentFailureService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPaymentFailureService creates a new payment failure service
func NewPaymentFailureService(db *gorm.DB, logger *zap.Logger) *PaymentFailureService {
	return &PaymentFailureService{
		db:     db,
		logger: logger,
	}
}

// GetPaymentFailures returns payment failures with filtering and pagination
func (s *PaymentFailureService) GetPaymentFailures(ctx context.Context, companyID string, filters map[string]interface{}, page, limit int) ([]models.PaymentFailureEvent, int64, error) {
	s.logger.Info("GetPaymentFailures called",
		zap.String("company_id", companyID),
		zap.Int("page", page),
		zap.Int("limit", limit))

	var failures []models.PaymentFailureEvent
	var total int64

	// Build query
	query := s.db.Model(&models.PaymentFailureEvent{}).Where("company_id = ?", companyID)
	s.logger.Info("Database query built", zap.String("company_id", companyID))

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if provider, ok := filters["provider_id"].(string); ok && provider != "" {
		query = query.Where("provider_id = ?", provider)
	}
	if customerEmail, ok := filters["customer_email"].(string); ok && customerEmail != "" {
		query = query.Where("customer_email ILIKE ?", "%"+customerEmail+"%")
	}
	if failureReason, ok := filters["failure_reason"].(string); ok && failureReason != "" {
		query = query.Where("failure_reason = ?", failureReason)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Get total count
	s.logger.Info("Executing count query")
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Count query failed", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count payment failures: %w", err)
	}
	s.logger.Info("Count query successful", zap.Int64("total", total))

	// Apply pagination and ordering
	offset := (page - 1) * limit
	s.logger.Info("Executing find query", zap.Int("offset", offset), zap.Int("limit", limit))
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&failures).Error; err != nil {
		s.logger.Error("Find query failed", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to fetch payment failures: %w", err)
	}
	s.logger.Info("Find query successful", zap.Int("failures_count", len(failures)))

	s.logger.Info("GetPaymentFailures completed", zap.Int("failures_count", len(failures)), zap.Int64("total", total))
	return failures, total, nil
}

// GetPaymentFailure returns a specific payment failure
func (s *PaymentFailureService) GetPaymentFailure(ctx context.Context, failureID uuid.UUID, companyID string) (*models.PaymentFailureEvent, error) {
	var failure models.PaymentFailureEvent

	if err := s.db.Where("id = ? AND company_id = ?", failureID, companyID).First(&failure).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment failure not found")
		}
		return nil, fmt.Errorf("failed to fetch payment failure: %w", err)
	}

	return &failure, nil
}

// CreatePaymentFailure creates a new payment failure event
func (s *PaymentFailureService) CreatePaymentFailure(ctx context.Context, failure *models.PaymentFailureEvent) error {
	if err := s.db.Create(failure).Error; err != nil {
		return fmt.Errorf("failed to create payment failure: %w", err)
	}
	return nil
}

// UpdatePaymentFailure updates an existing payment failure event
func (s *PaymentFailureService) UpdatePaymentFailure(ctx context.Context, failure *models.PaymentFailureEvent) error {
	if err := s.db.Save(failure).Error; err != nil {
		return fmt.Errorf("failed to update payment failure: %w", err)
	}
	return nil
}

// DeletePaymentFailure deletes a payment failure event
func (s *PaymentFailureService) DeletePaymentFailure(ctx context.Context, failureID uuid.UUID, companyID string) error {
	if err := s.db.Where("id = ? AND company_id = ?", failureID, companyID).Delete(&models.PaymentFailureEvent{}).Error; err != nil {
		return fmt.Errorf("failed to delete payment failure: %w", err)
	}
	return nil
}

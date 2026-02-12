package services

import (
	"context"
	"fmt"
	"time"

	"github.com/payment-watchdog/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertService handles payment failure alerts and notifications
type AlertService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(db *gorm.DB, logger *zap.Logger) *AlertService {
	return &AlertService{
		db:     db,
		logger: logger,
	}
}

// ProcessNewPaymentFailures processes new payment failures and generates alerts
func (s *AlertService) ProcessNewPaymentFailures(ctx context.Context) error {
	var failures []models.PaymentFailureEvent
	
	// Find unprocessed payment failures
	if err := s.db.Where("status = ? AND alerted_at IS NULL", "received").
		Find(&failures).Error; err != nil {
		return fmt.Errorf("failed to fetch unprocessed failures: %w", err)
	}

	s.logger.Info("Processing payment failures for alerts", zap.Int("count", len(failures)))

	for _, failure := range failures {
		if err := s.generateAlert(ctx, &failure); err != nil {
			s.logger.Error("Failed to generate alert", 
				zap.String("failure_id", failure.ID.String()),
				zap.Error(err))
			continue
		}

		// Mark as alerted
		if err := s.db.Model(&failure).Updates(map[string]interface{}{
			"status":      "alerted",
			"alerted_at":  time.Now(),
			"updated_at":  time.Now(),
		}).Error; err != nil {
			s.logger.Error("Failed to update failure status", zap.Error(err))
		}
	}

	return nil
}

// generateAlert creates and sends an alert for a payment failure
func (s *AlertService) generateAlert(ctx context.Context, failure *models.PaymentFailureEvent) error {
	// Create alert record
	alert := &models.CustomerCommunication{
		PaymentFailureID: failure.ID,
		CompanyID:        failure.CompanyID,
		Channel:          "email", // Start with email for MVP
		TemplateID:       "payment_failure_alert",
		Subject:          fmt.Sprintf("Payment Failed - %s", failure.FailureReason),
		Content:          s.generateAlertContent(failure),
		Status:           "pending",
	}

	if err := s.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// Send the alert
	if err := s.sendAlert(alert); err != nil {
		// Update status to failed
		s.db.Model(alert).Update("status", "failed")
		return fmt.Errorf("failed to send alert: %w", err)
	}

	// Update status to sent
	if err := s.db.Model(alert).Updates(map[string]interface{}{
		"status":   "sent",
		"sent_at":  time.Now(),
		"updated_at": time.Now(),
	}).Error; err != nil {
		s.logger.Error("Failed to update alert status", zap.Error(err))
	}

	s.logger.Info("Alert sent successfully", 
		zap.String("alert_id", alert.ID.String()),
		zap.String("customer_email", failure.CustomerEmail))

	return nil
}

// generateAlertContent generates professional alert content
func (s *AlertService) generateAlertContent(failure *models.PaymentFailureEvent) string {
	// Generate professional alert content
	return fmt.Sprintf(`
Payment Failure Alert

We've detected a failed payment attempt for your account.

Details:
- Amount: %s %s
- Customer: %s
- Failure Reason: %s
- Time: %s

Please review your payment method and update if necessary.

This is an automated alert from your payment intelligence system.
	`, 
		failure.Currency, 
		fmt.Sprintf("%.2f", failure.Amount),
		failure.CustomerEmail,
		failure.FailureReason,
		failure.WebhookReceivedAt.Format("2006-01-02 15:04:05"))
}

// sendAlert sends the actual alert
func (s *AlertService) sendAlert(alert *models.CustomerCommunication) error {
	// TODO: Implement actual email/SMS sending
	// For MVP, just log the alert
	s.logger.Info("Alert generated", 
		zap.String("alert_id", alert.ID.String()),
		zap.String("channel", alert.Channel),
		zap.String("subject", alert.Subject))
	
	return nil
}

// GetAlertStats returns alert statistics
func (s *AlertService) GetAlertStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats = make(map[string]interface{})
	
	// Total alerts
	var totalAlerts int64
	if err := s.db.Model(&models.CustomerCommunication{}).
		Where("company_id = ?", companyID).Count(&totalAlerts).Error; err != nil {
		return nil, err
	}
	stats["total_alerts"] = totalAlerts
	
	// Alerts by status
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.CustomerCommunication{}).
		Select("status, count(*) as count").
		Where("company_id = ?", companyID).
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, err
	}
	stats["status_breakdown"] = statusStats
	
	// Alerts by channel
	var channelStats []struct {
		Channel string `json:"channel"`
		Count   int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.CustomerCommunication{}).
		Select("channel, count(*) as count").
		Where("company_id = ?", companyID).
		Group("channel").
		Find(&channelStats).Error; err != nil {
		return nil, err
	}
	stats["channel_breakdown"] = channelStats
	
	return stats, nil
}

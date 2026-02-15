package rules

import (
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
	"go.uber.org/zap"
)

// PaymentFailureRules contains all business rules for payment failure processing
type PaymentFailureRules struct {
	logger *zap.Logger
}

// NewPaymentFailureRules creates a new instance of payment failure rules
func NewPaymentFailureRules(logger *zap.Logger) *PaymentFailureRules {
	return &PaymentFailureRules{
		logger: logger,
	}
}

// GetDefaultRules returns the default set of business rules
func (pfr *PaymentFailureRules) GetDefaultRules() []*BasicRule {
	return []*BasicRule{
		pfr.createHighValueAlertRule(),
		pfr.createInsufficientFundsRetryRule(),
		pfr.createRiskScoringRule(),
	}
}

// High Value Alert Rule
func (pfr *PaymentFailureRules) createHighValueAlertRule() *BasicRule {
	return &BasicRule{
		Name:        "high_value_alert",
		Description: "Send immediate alert for high-value payment failures",
		Priority:    100,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.Amount >= 1000.0
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			pfr.logger.Info("High value payment failure detected",
				zap.String("event_id", event.ID.String()),
				zap.Float64("amount", event.Amount))

			return &BasicActionResult{
				RuleName:   "high_value_alert",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "High-value alert created",
				Data: map[string]interface{}{
					"alert_type": "high_value",
					"urgency":    "immediate",
				},
			}, nil
		},
	}
}

// Insufficient Funds Retry Rule
func (pfr *PaymentFailureRules) createInsufficientFundsRetryRule() *BasicRule {
	return &BasicRule{
		Name:        "insufficient_funds_retry",
		Description: "Schedule retry for insufficient funds after payday",
		Priority:    70,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "insufficient_funds"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			pfr.logger.Info("Insufficient funds detected, scheduling retry",
				zap.String("event_id", event.ID.String()))

			retryTime := time.Now().AddDate(0, 0, 7)

			return &BasicActionResult{
				RuleName:   "insufficient_funds_retry",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Retry scheduled for after payday",
				Data: map[string]interface{}{
					"retry_time": retryTime,
					"delay":      "7_days",
				},
			}, nil
		},
	}
}

// Risk Scoring Rule
func (pfr *PaymentFailureRules) createRiskScoringRule() *BasicRule {
	return &BasicRule{
		Name:        "risk_scoring",
		Description: "Calculate risk score for customer",
		Priority:    10,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return true
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			riskScore := calculateRiskScore(event)

			return &BasicActionResult{
				RuleName:   "risk_scoring",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Risk score calculated",
				Data: map[string]interface{}{
					"risk_score": riskScore,
				},
			}, nil
		},
	}
}

// Helper function for risk scoring
func calculateRiskScore(event *models.PaymentFailureEvent) int {
	baseScore := 0

	if event.Amount >= 1000 {
		baseScore += 30
	} else if event.Amount >= 500 {
		baseScore += 20
	} else if event.Amount >= 100 {
		baseScore += 10
	}

	switch event.FailureReason {
	case "insufficient_funds":
		baseScore += 20
	case "expired_card":
		baseScore += 40
	case "network_error":
		baseScore += 5
	}

	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

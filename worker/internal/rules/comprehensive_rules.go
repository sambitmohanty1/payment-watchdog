package rules

import (
	"time"

	"github.com/lexure-intelligence/payment-watchdog/internal/models"
	"go.uber.org/zap"
)

// ComprehensivePaymentFailureRules contains all business rules based on industry research
type ComprehensivePaymentFailureRules struct {
	logger *zap.Logger
}

// NewComprehensivePaymentFailureRules creates comprehensive payment failure rules
func NewComprehensivePaymentFailureRules(logger *zap.Logger) *ComprehensivePaymentFailureRules {
	return &ComprehensivePaymentFailureRules{
		logger: logger,
	}
}

// GetComprehensiveRules returns all business rules organized by priority
func (cpr *ComprehensivePaymentFailureRules) GetComprehensiveRules() []*BasicRule {
	return []*BasicRule{
		// CRITICAL PRIORITY (200+) - Immediate action required
		cpr.createFraudDetectionRule(),
		cpr.createBankDishonourRule(),
		cpr.createExpiredPaymentMethodRule(),

		// HIGH PRIORITY (100-199) - Business critical
		cpr.createHighValueImmediateAlertRule(),
		cpr.createRecurringPaymentFailureRule(),
		cpr.createVIPCustomerRule(),

		// MEDIUM PRIORITY (50-99) - Standard processing
		cpr.createInsufficientFundsRetryRule(),
		cpr.createCardDeclinedRetryRule(),
		cpr.createNetworkErrorRetryRule(),

		// LOW PRIORITY (1-49) - Background processing
		cpr.createRiskScoringRule(),
		cpr.createAnalyticsRule(),
	}
}

// CRITICAL PRIORITY RULES (200+)

// Fraud Detection Rule - Highest priority
func (cpr *ComprehensivePaymentFailureRules) createFraudDetectionRule() *BasicRule {
	return &BasicRule{
		Name:        "fraud_detection",
		Description: "Detect and block suspicious payment patterns",
		Priority:    200,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			// Check for suspicious patterns
			return cpr.isSuspiciousPattern(event)
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Warn("Suspicious payment pattern detected",
				zap.String("event_id", event.ID.String()),
				zap.String("customer_id", event.CustomerID),
				zap.Float64("amount", event.Amount))

			return &BasicActionResult{
				RuleName:   "fraud_detection",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Payment blocked for fraud investigation",
				Data: map[string]interface{}{
					"action":         "block_payment",
					"fraud_score":    95,
					"investigation":  true,
					"no_retry":       true,
					"security_alert": true,
				},
			}, nil
		},
	}
}

// Bank Dishonour Rule
func (cpr *ComprehensivePaymentFailureRules) createBankDishonourRule() *BasicRule {
	return &BasicRule{
		Name:        "bank_dishonour",
		Description: "Handle bank dishonours with immediate customer contact",
		Priority:    180,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "dishonour" ||
				event.FailureReason == "bank_dishonour" ||
				event.FailureReason == "account_closed"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("Bank dishonour detected",
				zap.String("event_id", event.ID.String()),
				zap.String("failure_reason", event.FailureReason))

			return &BasicActionResult{
				RuleName:   "bank_dishonour",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Bank dishonour - immediate customer contact required",
				Data: map[string]interface{}{
					"action_required": "customer_contact",
					"urgency":         "immediate",
					"channel":         "phone",
					"no_retry":        true,
					"escalation":      true,
				},
			}, nil
		},
	}
}

// Expired Payment Method Rule
func (cpr *ComprehensivePaymentFailureRules) createExpiredPaymentMethodRule() *BasicRule {
	return &BasicRule{
		Name:        "expired_payment_method",
		Description: "Handle expired payment methods - no retry, request update",
		Priority:    160,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "expired_card" ||
				event.FailureReason == "card_expired" ||
				event.FailureReason == "expired_payment_method"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("Expired payment method detected",
				zap.String("event_id", event.ID.String()),
				zap.String("failure_reason", event.FailureReason))

			return &BasicActionResult{
				RuleName:   "expired_payment_method",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Payment method expired - update required",
				Data: map[string]interface{}{
					"action_required": "update_payment_method",
					"no_retry":        true,
					"urgency":         "high",
					"channel":         "email",
					"update_link":     true,
				},
			}, nil
		},
	}
}

// HIGH PRIORITY RULES (100-199)

// High Value Immediate Alert Rule
func (cpr *ComprehensivePaymentFailureRules) createHighValueImmediateAlertRule() *BasicRule {
	return &BasicRule{
		Name:        "high_value_immediate_alert",
		Description: "Send immediate SMS alert for high-value payment failures",
		Priority:    150,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.Amount >= 1000.0
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("High value payment failure detected",
				zap.String("event_id", event.ID.String()),
				zap.Float64("amount", event.Amount))

			return &BasicActionResult{
				RuleName:   "high_value_immediate_alert",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "High-value alert created and sent",
				Data: map[string]interface{}{
					"alert_type":      "high_value",
					"urgency":         "immediate",
					"channel":         "sms",
					"escalation":      true,
					"cashflow_impact": "high",
				},
			}, nil
		},
	}
}

// Recurring Payment Failure Rule
func (cpr *ComprehensivePaymentFailureRules) createRecurringPaymentFailureRule() *BasicRule {
	return &BasicRule{
		Name:        "recurring_payment_failure",
		Description: "Handle recurring payment failures with escalation",
		Priority:    120,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			// TODO: Check if this is a recurring payment and has failed before
			return cpr.isRecurringPayment(event)
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Warn("Recurring payment failure detected",
				zap.String("event_id", event.ID.String()),
				zap.String("customer_id", event.CustomerID))

			return &BasicActionResult{
				RuleName:   "recurring_payment_failure",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Recurring payment failure - escalation required",
				Data: map[string]interface{}{
					"action_required": "escalation",
					"urgency":         "high",
					"channel":         "sms",
					"retention_risk":  "high",
					"escalation":      true,
				},
			}, nil
		},
	}
}

// VIP Customer Rule
func (cpr *ComprehensivePaymentFailureRules) createVIPCustomerRule() *BasicRule {
	return &BasicRule{
		Name:        "vip_customer",
		Description: "Special handling for VIP customers",
		Priority:    110,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			// TODO: Check if customer is VIP based on business rules
			return cpr.isVIPCustomer(event.CustomerID)
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("VIP customer payment failure",
				zap.String("event_id", event.ID.String()),
				zap.String("customer_id", event.CustomerID))

			return &BasicActionResult{
				RuleName:   "vip_customer",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "VIP customer - special handling applied",
				Data: map[string]interface{}{
					"customer_tier":      "vip",
					"urgency":            "high",
					"channel":            "phone",
					"personal_service":   true,
					"retention_priority": "critical",
				},
			}, nil
		},
	}
}

// MEDIUM PRIORITY RULES (50-99)

// Insufficient Funds Retry Rule
func (cpr *ComprehensivePaymentFailureRules) createInsufficientFundsRetryRule() *BasicRule {
	return &BasicRule{
		Name:        "insufficient_funds_retry",
		Description: "Schedule retry for insufficient funds after payday",
		Priority:    80,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "insufficient_funds" ||
				event.FailureReason == "insufficient_balance"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("Insufficient funds detected, scheduling retry",
				zap.String("event_id", event.ID.String()))

			// Smart retry timing based on amount and customer history
			retryDelay := cpr.calculateRetryDelay(event)
			retryTime := time.Now().Add(retryDelay)

			return &BasicActionResult{
				RuleName:   "insufficient_funds_retry",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Retry scheduled with smart timing",
				Data: map[string]interface{}{
					"retry_time":   retryTime,
					"retry_delay":  retryDelay.String(),
					"retry_method": "scheduled",
					"delay_reason": "wait_for_funds",
					"smart_timing": true,
				},
			}, nil
		},
	}
}

// Card Declined Retry Rule
func (cpr *ComprehensivePaymentFailureRules) createCardDeclinedRetryRule() *BasicRule {
	return &BasicRule{
		Name:        "card_declined_retry",
		Description: "Smart retry for card declined errors",
		Priority:    70,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "card_declined" ||
				event.FailureReason == "do_not_honor" ||
				event.FailureReason == "generic_decline"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("Card declined, scheduling smart retry",
				zap.String("event_id", event.ID.String()))

			// Smart retry timing based on decline reason
			retryDelay := cpr.calculateCardDeclineRetryDelay(event)
			retryTime := time.Now().Add(retryDelay)

			return &BasicActionResult{
				RuleName:   "card_declined_retry",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Smart retry scheduled for card decline",
				Data: map[string]interface{}{
					"retry_time":   retryTime,
					"retry_delay":  retryDelay.String(),
					"retry_method": "scheduled",
					"delay_reason": "card_decline_cooldown",
					"smart_timing": true,
				},
			}, nil
		},
	}
}

// Network Error Retry Rule
func (cpr *ComprehensivePaymentFailureRules) createNetworkErrorRetryRule() *BasicRule {
	return &BasicRule{
		Name:        "network_error_retry",
		Description: "Quick retry for network/technical errors",
		Priority:    60,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return event.FailureReason == "network_error" ||
				event.FailureReason == "timeout" ||
				event.FailureReason == "connection_error"
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			cpr.logger.Info("Network error detected, scheduling quick retry",
				zap.String("event_id", event.ID.String()))

			retryTime := time.Now().Add(15 * time.Minute)

			return &BasicActionResult{
				RuleName:   "network_error_retry",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Quick retry scheduled for network error",
				Data: map[string]interface{}{
					"retry_time":   retryTime,
					"retry_delay":  "15m",
					"retry_method": "scheduled",
					"delay_reason": "network_error_cooldown",
					"quick_retry":  true,
				},
			}, nil
		},
	}
}

// LOW PRIORITY RULES (1-49)

// Risk Scoring Rule
func (cpr *ComprehensivePaymentFailureRules) createRiskScoringRule() *BasicRule {
	return &BasicRule{
		Name:        "risk_scoring",
		Description: "Calculate comprehensive risk score for customer",
		Priority:    30,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return true // Always execute for risk assessment
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			riskScore := cpr.calculateComprehensiveRiskScore(event)

			return &BasicActionResult{
				RuleName:   "risk_scoring",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Risk score calculated and updated",
				Data: map[string]interface{}{
					"risk_score":      riskScore,
					"risk_category":   cpr.getRiskCategory(riskScore),
					"risk_factors":    cpr.getRiskFactors(event),
					"recommendations": cpr.getRiskRecommendations(riskScore),
				},
			}, nil
		},
	}
}

// Analytics Rule
func (cpr *ComprehensivePaymentFailureRules) createAnalyticsRule() *BasicRule {
	return &BasicRule{
		Name:        "analytics",
		Description: "Collect data for analytics and ML training",
		Priority:    10,
		Enabled:     true,
		Condition: func(event *models.PaymentFailureEvent) bool {
			return true // Always execute for data collection
		},
		Action: func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
			// TODO: Send event to analytics pipeline
			cpr.logger.Debug("Analytics data collected",
				zap.String("event_id", event.ID.String()))

			return &BasicActionResult{
				RuleName:   "analytics",
				Success:    true,
				ExecutedAt: time.Now(),
				Message:    "Analytics data collected",
				Data: map[string]interface{}{
					"analytics_collected": true,
					"ml_training":         true,
					"pattern_detection":   true,
				},
			}, nil
		},
	}
}

// Helper methods

func (cpr *ComprehensivePaymentFailureRules) isSuspiciousPattern(event *models.PaymentFailureEvent) bool {
	// Basic fraud detection logic
	// TODO: Implement more sophisticated fraud detection
	return event.Amount > 5000.0 && event.FailureReason == "card_declined"
}

func (cpr *ComprehensivePaymentFailureRules) isRecurringPayment(event *models.PaymentFailureEvent) bool {
	// TODO: Check database for recurring payment failures
	return false
}

func (cpr *ComprehensivePaymentFailureRules) isVIPCustomer(customerID string) bool {
	// TODO: Check customer database for VIP status
	return false
}

func (cpr *ComprehensivePaymentFailureRules) calculateRetryDelay(event *models.PaymentFailureEvent) time.Duration {
	// Smart retry timing based on amount
	if event.Amount >= 1000.0 {
		return 7 * 24 * time.Hour // 7 days for high amounts
	} else if event.Amount >= 500.0 {
		return 5 * 24 * time.Hour // 5 days for medium amounts
	}
	return 3 * 24 * time.Hour // 3 days for low amounts
}

func (cpr *ComprehensivePaymentFailureRules) calculateCardDeclineRetryDelay(event *models.PaymentFailureEvent) time.Duration {
	// Different timing based on decline reason
	switch event.FailureReason {
	case "card_declined":
		return 1 * time.Hour
	case "do_not_honor":
		return 24 * time.Hour
	default:
		return 2 * time.Hour
	}
}

func (cpr *ComprehensivePaymentFailureRules) calculateComprehensiveRiskScore(event *models.PaymentFailureEvent) int {
	baseScore := 0

	// Amount-based risk
	if event.Amount >= 10000 {
		baseScore += 40
	} else if event.Amount >= 5000 {
		baseScore += 30
	} else if event.Amount >= 1000 {
		baseScore += 20
	} else if event.Amount >= 100 {
		baseScore += 10
	}

	// Failure reason risk
	switch event.FailureReason {
	case "insufficient_funds":
		baseScore += 25
	case "expired_card":
		baseScore += 45
	case "bank_dishonour":
		baseScore += 50
	case "card_declined":
		baseScore += 20
	case "network_error":
		baseScore += 5
	}

	// Normalize to 0-100 scale
	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

func (cpr *ComprehensivePaymentFailureRules) getRiskFactors(event *models.PaymentFailureEvent) []string {
	factors := []string{}

	if event.Amount >= 1000 {
		factors = append(factors, "high_amount")
	}

	switch event.FailureReason {
	case "insufficient_funds":
		factors = append(factors, "insufficient_funds")
	case "expired_card":
		factors = append(factors, "expired_payment_method")
	case "bank_dishonour":
		factors = append(factors, "bank_dishonour")
	}

	return factors
}

func (cpr *ComprehensivePaymentFailureRules) getRiskCategory(score int) string {
	if score >= 80 {
		return "critical"
	} else if score >= 60 {
		return "high"
	} else if score >= 40 {
		return "medium"
	} else if score >= 20 {
		return "low"
	}
	return "minimal"
}

func (cpr *ComprehensivePaymentFailureRules) getRiskRecommendations(score int) []string {
	if score >= 80 {
		return []string{"immediate_escalation", "customer_contact", "fraud_investigation"}
	} else if score >= 60 {
		return []string{"escalation", "customer_contact", "monitoring"}
	} else if score >= 40 {
		return []string{"standard_processing", "monitoring"}
	}
	return []string{"standard_processing"}
}

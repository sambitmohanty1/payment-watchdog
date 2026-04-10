package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
)

// RulesService implements business logic for payment processing rules
type RulesService struct {
	logger *zap.Logger
}

// NewRulesService creates a new rules service
func NewRulesService(logger *zap.Logger) *RulesService {
	return &RulesService{
		logger: logger,
	}
}

// EvaluatePaymentFailure evaluates payment failure against business rules
func (s *RulesService) EvaluatePaymentFailure(ctx context.Context, paymentEvent *events.PaymentEvent) (*interfaces.RuleEvaluation, error) {
	s.logger.Info("Evaluating payment failure against rules",
		zap.String("payment_id", paymentEvent.ID),
		zap.Float64("amount", paymentEvent.Amount),
	)

	// Extract failure reason from metadata using safe casting (Phase 1 learning)
	failureReason := ""
	if val, ok := paymentEvent.Metadata["reason"]; ok {
		failureReason = fmt.Sprint(val)
	}

	// Classify the failure based on amount thresholds and failure reason
	ruleName := "standard_failure"
	action := "log_and_monitor"
	confidence := 0.80
	isMatch := true

	switch {
	case paymentEvent.Amount > 10000:
		ruleName = "high_value_failure"
		action = "escalate_immediately"
		confidence = 0.99
	case paymentEvent.Amount > 1000:
		ruleName = "medium_value_failure"
		action = "flag_for_review"
		confidence = 0.95
	case failureReason == "insufficient_funds":
		ruleName = "insufficient_funds_failure"
		action = "schedule_retry"
		confidence = 0.90
	case failureReason == "card_expired":
		ruleName = "expired_card_failure"
		action = "notify_customer"
		confidence = 0.98
	case failureReason == "fraud_suspected":
		ruleName = "fraud_detection"
		action = "block_and_alert"
		confidence = 0.85
	}

	evaluation := &interfaces.RuleEvaluation{
		RuleName:   ruleName,
		IsMatch:    isMatch,
		Action:     action,
		Confidence: confidence,
	}

	s.logger.Info("Rule evaluation completed",
		zap.String("payment_id", paymentEvent.ID),
		zap.String("rule", ruleName),
		zap.String("action", action),
		zap.Float64("confidence", confidence),
	)
	return evaluation, nil
}

// ApplyRecoveryStrategy applies recovery strategy based on rules evaluation
func (s *RulesService) ApplyRecoveryStrategy(ctx context.Context, paymentEvent *events.PaymentEvent) (*interfaces.RecoveryStrategy, error) {
	s.logger.Info("Applying recovery strategy",
		zap.String("payment_id", paymentEvent.ID),
	)

	// Evaluate the failure first to determine appropriate strategy
	evaluation, err := s.EvaluatePaymentFailure(ctx, paymentEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate payment failure for strategy: %w", err)
	}

	// Select recovery strategy based on the rule evaluation action
	// Since we don't have Stripe/Xero endpoints yet, we record what *would* happen
	strategyType := "exponential_backoff"
	nextDelay := 1 * time.Hour
	maxAttempts := 3

	switch evaluation.Action {
	case "escalate_immediately":
		strategyType = "immediate_escalation"
		nextDelay = 5 * time.Minute
		maxAttempts = 1
	case "flag_for_review":
		strategyType = "delayed_retry_with_review"
		nextDelay = 2 * time.Hour
		maxAttempts = 2
	case "schedule_retry":
		strategyType = "exponential_backoff"
		nextDelay = 30 * time.Minute
		maxAttempts = 5
	case "notify_customer":
		strategyType = "customer_notification_only"
		nextDelay = 24 * time.Hour
		maxAttempts = 0 // No automatic retry for expired cards
	case "block_and_alert":
		strategyType = "fraud_block"
		nextDelay = 0
		maxAttempts = 0 // No retry for suspected fraud
	}

	strategy := &interfaces.RecoveryStrategy{
		StrategyType:  strategyType,
		Amount:        paymentEvent.Amount,
		PaymentMethod: "original_method", // Preserved until ProviderRegistry confirms endpoint is live
		NextAttempt:   time.Now().Add(nextDelay),
		MaxAttempts:   maxAttempts,
	}

	s.logger.Info("Recovery strategy applied",
		zap.String("payment_id", paymentEvent.ID),
		zap.String("strategy", strategyType),
		zap.Int("max_attempts", maxAttempts),
	)
	return strategy, nil
}

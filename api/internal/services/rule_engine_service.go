package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/internal/models"
	"github.com/sambitmohanty1/payment-watchdog/internal/rules"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RuleEngineService integrates the rule engine with business operations
type RuleEngineService struct {
	db         *gorm.DB
	ruleEngine rules.RuleEngine
	logger     *zap.Logger
}

// NewRuleEngineService creates a new rule engine service
func NewRuleEngineService(db *gorm.DB, ruleEngine rules.RuleEngine, logger *zap.Logger) *RuleEngineService {
	return &RuleEngineService{
		db:         db,
		ruleEngine: ruleEngine,
		logger:     logger,
	}
}

// ProcessPaymentFailureEvent processes a payment failure event through the rule engine
func (s *RuleEngineService) ProcessPaymentFailureEvent(ctx context.Context, event *models.PaymentFailureEvent) error {
	s.logger.Info("Processing payment failure event through rule engine",
		zap.String("event_id", event.ID.String()),
		zap.String("provider", event.ProviderID),
		zap.String("failure_reason", event.FailureReason))

	// Execute all applicable rules
	results := s.ruleEngine.ExecuteRules(event)

	// Process rule execution results
	for _, result := range results {
		if !result.Success {
			s.logger.Error("Rule execution failed",
				zap.String("rule_name", result.RuleName),
				zap.Error(result.Error))
			continue
		}

		s.logger.Info("Rule executed successfully",
			zap.String("rule_name", result.RuleName),
			zap.String("message", result.Message))

		// Handle specific rule results
		if err := s.handleRuleResult(ctx, event, result); err != nil {
			s.logger.Error("Failed to handle rule result",
				zap.String("rule_name", result.RuleName),
				zap.Error(err))
		}
	}

	// Update event status based on rule execution
	if err := s.updateEventStatus(ctx, event, results); err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	s.logger.Info("Payment failure event processing completed",
		zap.String("event_id", event.ID.String()),
		zap.Int("rules_executed", len(results)))

	return nil
}

// handleRuleResult processes the result of a specific rule execution
func (s *RuleEngineService) handleRuleResult(ctx context.Context, event *models.PaymentFailureEvent, result *rules.ActionResult) error {
	switch result.RuleName {
	case "high_value_immediate_alert", "low_value_alert":
		return s.handleAlertRule(ctx, event, result)
	case "insufficient_funds_retry", "card_declined_retry", "network_error_retry":
		return s.handleRetryRule(ctx, event, result)
	case "bank_dishonour_rule", "expired_card_rule", "customer_communication":
		return s.handleCommunicationRule(ctx, event, result)
	case "risk_scoring":
		return s.handleRiskScoringRule(ctx, event, result)
	default:
		s.logger.Warn("Unknown rule result type", zap.String("rule_name", result.RuleName))
		return nil
	}
}

// handleAlertRule processes alert-related rule results
func (s *RuleEngineService) handleAlertRule(ctx context.Context, event *models.PaymentFailureEvent, result *rules.ActionResult) error {
	// Extract alert information from result data
	alertID, ok := result.Data["alert_id"].(string)
	if !ok {
		return fmt.Errorf("alert_id not found in rule result data")
	}

	// TODO: Save alert to database and trigger delivery
	s.logger.Info("Alert rule handled",
		zap.String("rule_name", result.RuleName),
		zap.String("alert_id", alertID))

	return nil
}

// handleRetryRule processes retry-related rule results
func (s *RuleEngineService) handleRetryRule(ctx context.Context, event *models.PaymentFailureEvent, result *rules.ActionResult) error {
	// Extract retry information from result data
	retryID, ok := result.Data["retry_id"].(string)
	if !ok {
		return fmt.Errorf("retry_id not found in rule result data")
	}

	// TODO: Save retry attempt to database and schedule execution
	s.logger.Info("Retry rule handled",
		zap.String("rule_name", result.RuleName),
		zap.String("retry_id", retryID))

	return nil
}

// handleCommunicationRule processes communication-related rule results
func (s *RuleEngineService) handleCommunicationRule(ctx context.Context, event *models.PaymentFailureEvent, result *rules.ActionResult) error {
	// Extract communication information from result data
	commID, ok := result.Data["communication_id"].(string)
	if !ok {
		return fmt.Errorf("communication_id not found in rule result data")
	}

	// TODO: Save communication to database and trigger delivery
	s.logger.Info("Communication rule handled",
		zap.String("rule_name", result.RuleName),
		zap.String("communication_id", commID))

	return nil
}

// handleRiskScoringRule processes risk scoring rule results
func (s *RuleEngineService) handleRiskScoringRule(ctx context.Context, event *models.PaymentFailureEvent, result *rules.ActionResult) error {
	// Extract risk score information from result data
	riskScore, ok := result.Data["risk_score"].(int)
	if !ok {
		return fmt.Errorf("risk_score not found in rule result data")
	}

	// TODO: Update customer risk profile in database
	s.logger.Info("Risk scoring rule handled",
		zap.String("rule_name", result.RuleName),
		zap.Int("risk_score", riskScore))

	return nil
}

// updateEventStatus updates the event status based on rule execution results
func (s *RuleEngineService) updateEventStatus(ctx context.Context, event *models.PaymentFailureEvent, results []*rules.ActionResult) error {
	// Determine new status based on rule results
	newStatus := "processed"

	for _, result := range results {
		if result.Success {
			// Check if any high-priority actions were taken
			if urgency, ok := result.Data["urgency"].(string); ok && urgency == "immediate" {
				newStatus = "alerted"
				break
			}

			// Check if retry was scheduled
			if _, ok := result.Data["retry_id"].(string); ok {
				newStatus = "retry_scheduled"
			}
		}
	}

	// Update event status
	event.Status = newStatus
	event.ProcessedAt = &[]time.Time{time.Now()}[0]

	// Save to database
	if err := s.db.Save(event).Error; err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	s.logger.Info("Event status updated",
		zap.String("event_id", event.ID.String()),
		zap.String("new_status", newStatus))

	return nil
}

// GetRuleEngineStats returns statistics about the rule engine
func (s *RuleEngineService) GetRuleEngineStats() map[string]interface{} {
	return s.ruleEngine.GetStats()
}

// EnableRule enables a specific rule by name
func (s *RuleEngineService) EnableRule(ruleName string) {
	s.ruleEngine.EnableRule(ruleName)
	s.logger.Info("Rule enabled", zap.String("rule_name", ruleName))
}

// DisableRule disables a specific rule by name
func (s *RuleEngineService) DisableRule(ruleName string) {
	s.ruleEngine.DisableRule(ruleName)
	s.logger.Info("Rule disabled", zap.String("rule_name", ruleName))
}

// GetRules returns all registered rules
func (s *RuleEngineService) GetRules() []*rules.Rule {
	return s.ruleEngine.GetRules()
}

// GetRuleByName returns a specific rule by name
func (s *RuleEngineService) GetRuleByName(name string) *rules.Rule {
	return s.ruleEngine.GetRuleByName(name)
}

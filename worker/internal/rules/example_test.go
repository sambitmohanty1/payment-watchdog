package rules

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sambitmohanty1/payment-watchdog/internal/models"
	"go.uber.org/zap"
)

func TestRuleEngine(t *testing.T) {
	// Create logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create rule engine factory
	factory := NewRuleEngineFactory(logger)

	// Create payment failure rule engine
	engine := factory.CreateBasicRuleEngine()

	// Test high-value payment failure
	t.Run("High Value Payment Failure", func(t *testing.T) {
		event := &models.PaymentFailureEvent{
			ID:                uuid.New(),
			CompanyID:         "company-123",
			ProviderID:        "stripe",
			EventID:           "evt_123",
			EventType:         "payment_intent.payment_failed",
			Amount:            2500.0,
			Currency:          "AUD",
			CustomerID:        "cus_123",
			CustomerEmail:     "test@example.com",
			CustomerName:      "Test Customer",
			FailureReason:     "card_declined",
			FailureCode:       "card_declined",
			FailureMessage:    "Your card was declined",
			Status:            "received",
			WebhookReceivedAt: time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		results := engine.ExecuteRules(event)

		if len(results) == 0 {
			t.Error("Expected rules to execute for high-value payment failure")
		}

		// Check if high-value alert rule was executed
		highValueAlertExecuted := false
		for _, result := range results {
			if result.RuleName == "high_value_alert" && result.Success {
				highValueAlertExecuted = true
				break
			}
		}

		if !highValueAlertExecuted {
			t.Error("Expected high-value alert rule to execute")
		}
	})

	// Test insufficient funds payment failure
	t.Run("Insufficient Funds Payment Failure", func(t *testing.T) {
		event := &models.PaymentFailureEvent{
			ID:                uuid.New(),
			CompanyID:         "company-123",
			ProviderID:        "stripe",
			EventID:           "evt_124",
			EventType:         "payment_intent.payment_failed",
			Amount:            500.0,
			Currency:          "AUD",
			CustomerID:        "cus_124",
			CustomerEmail:     "test2@example.com",
			CustomerName:      "Test Customer 2",
			FailureReason:     "insufficient_funds",
			FailureCode:       "insufficient_funds",
			FailureMessage:    "Insufficient funds",
			Status:            "received",
			WebhookReceivedAt: time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		results := engine.ExecuteRules(event)

		if len(results) == 0 {
			t.Error("Expected rules to execute for insufficient funds payment failure")
		}

		// Check if insufficient funds retry rule was executed
		insufficientFundsRetryExecuted := false
		for _, result := range results {
			if result.RuleName == "insufficient_funds_retry" && result.Success {
				insufficientFundsRetryExecuted = true
				break
			}
		}

		if !insufficientFundsRetryExecuted {
			t.Error("Expected insufficient funds retry rule to execute")
		}
	})

	// Test expired card payment failure
	t.Run("Expired Card Payment Failure", func(t *testing.T) {
		t.Skip("No expired card rule in default rules")
	})
}

func TestRuleEngineStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	factory := NewRuleEngineFactory(logger)
	engine := factory.CreateBasicRuleEngine()

	stats := engine.GetStats()

	expectedRules := 3 // Update to match actual number of default rules
	if stats["total_rules"] != expectedRules {
		t.Errorf("Expected %d total rules, got %v", expectedRules, stats["total_rules"])
	}

	if stats["enabled_rules"] != expectedRules {
		t.Errorf("Expected %d enabled rules, got %v", expectedRules, stats["enabled_rules"])
	}

	if stats["disabled_rules"] != 0 {
		t.Errorf("Expected 0 disabled rules, got %v", stats["disabled_rules"])
	}
}

func TestRuleEngineEnableDisable(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	factory := NewRuleEngineFactory(logger)
	engine := factory.CreateBasicRuleEngine()

	// Disable a rule
	engine.DisableRule("high_value_alert")

	// Check stats
	stats := engine.GetStats()
	if stats["enabled_rules"] != 2 { // 3 - 1 = 2
		t.Errorf("Expected 2 enabled rules after disabling one, got %v", stats["enabled_rules"])
	}

	if stats["disabled_rules"] != 1 {
		t.Errorf("Expected 1 disabled rule, got %v", stats["disabled_rules"])
	}

	// Re-enable the rule
	engine.EnableRule("high_value_alert")

	// Check stats again
	stats = engine.GetStats()
	if stats["enabled_rules"] != 3 {
		t.Errorf("Expected 3 enabled rules after re-enabling, got %v", stats["enabled_rules"])
	}
	if stats["disabled_rules"] != 0 {
		t.Errorf("Expected 0 disabled rules after re-enabling, got %v", stats["disabled_rules"])
	}
}

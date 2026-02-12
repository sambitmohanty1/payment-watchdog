package rules

import (
	"fmt"
	"sort"
	"time"

	"github.com/lexure-intelligence/payment-watchdog/internal/models"
	"go.uber.org/zap"
)

// BasicRule represents a business rule that can be executed
type BasicRule struct {
	Name        string
	Description string
	Condition   func(*models.PaymentFailureEvent) bool
	Action      func(*models.PaymentFailureEvent) (*BasicActionResult, error)
	Priority    int
	Enabled     bool
}

// BasicActionResult represents the result of executing a rule
type BasicActionResult struct {
	RuleName   string
	Success    bool
	Error      error
	ExecutedAt time.Time
	Message    string
	Data       map[string]interface{}
}

// BasicRuleEngine manages and executes business rules
type BasicRuleEngine struct {
	rules  []*BasicRule
	cache  map[string]interface{}
	logger *zap.Logger
}

// NewBasicRuleEngine creates a new basic rule engine instance
func NewBasicRuleEngine(logger *zap.Logger) *BasicRuleEngine {
	return &BasicRuleEngine{
		rules:  make([]*BasicRule, 0),
		cache:  make(map[string]interface{}),
		logger: logger,
	}
}

// AddRule adds a rule to the engine
func (e *BasicRuleEngine) AddRule(rule *BasicRule) {
	e.rules = append(e.rules, rule)
	// Sort rules by priority (highest first)
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority > e.rules[j].Priority
	})
	e.logger.Info("Rule added", zap.String("rule_name", rule.Name), zap.Int("priority", rule.Priority))
}

// ExecuteRules executes all applicable rules for a payment failure event
func (e *BasicRuleEngine) ExecuteRules(event *models.PaymentFailureEvent) []*BasicActionResult {
	var results []*BasicActionResult

	e.logger.Info("Executing rules",
		zap.String("event_id", event.ID.String()),
		zap.String("provider", event.ProviderID),
		zap.String("failure_reason", event.FailureReason))

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule condition is met
		if rule.Condition(event) {
			e.logger.Debug("Rule condition met",
				zap.String("rule_name", rule.Name),
				zap.String("event_id", event.ID.String()))

			// Execute rule action
			result, err := rule.Action(event)
			if err != nil {
				e.logger.Error("Rule execution failed",
					zap.String("rule_name", rule.Name),
					zap.Error(err))
				result = &BasicActionResult{
					RuleName:   rule.Name,
					Success:    false,
					Error:      err,
					ExecutedAt: time.Now(),
					Message:    fmt.Sprintf("Rule execution failed: %v", err),
				}
			}

			results = append(results, result)
		}
	}

	e.logger.Info("Rules execution completed",
		zap.String("event_id", event.ID.String()),
		zap.Int("rules_executed", len(results)))

	return results
}

// GetStats returns statistics about the rule engine
func (e *BasicRuleEngine) GetStats() map[string]interface{} {
	enabledCount := 0
	disabledCount := 0

	for _, rule := range e.rules {
		if rule.Enabled {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	return map[string]interface{}{
		"total_rules":    len(e.rules),
		"enabled_rules":  enabledCount,
		"disabled_rules": disabledCount,
		"cache_size":     len(e.cache),
	}
}

// EnableRule enables a specific rule by name
func (e *BasicRuleEngine) EnableRule(ruleName string) {
	for _, rule := range e.rules {
		if rule.Name == ruleName {
			rule.Enabled = true
			e.logger.Info("Rule enabled", zap.String("rule_name", ruleName))
			return
		}
	}
}

// DisableRule disables a specific rule by name
func (e *BasicRuleEngine) DisableRule(ruleName string) {
	for _, rule := range e.rules {
		if rule.Name == ruleName {
			rule.Enabled = false
			e.logger.Info("Rule disabled", zap.String("rule_name", ruleName))
			return
		}
	}
}

// GetRules returns all registered rules
func (e *BasicRuleEngine) GetRules() []*BasicRule {
	return e.rules
}

// GetRuleByName returns a specific rule by name
func (e *BasicRuleEngine) GetRuleByName(name string) *BasicRule {
	for _, rule := range e.rules {
		if rule.Name == name {
			return rule
		}
	}
	return nil
}

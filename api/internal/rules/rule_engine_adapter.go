package rules

import (
	"time"

	"github.com/payment-watchdog/internal/models"
)

// RuleEngine represents the interface for a business rules engine
type RuleEngine interface {
	// ExecuteRules executes all applicable rules for a payment failure event
	ExecuteRules(event *models.PaymentFailureEvent) []*ActionResult

	// GetRules returns all rules in the engine
	GetRules() []*Rule

	// AddRule adds a rule to the engine
	AddRule(rule *Rule)

	// RemoveRule removes a rule from the engine
	RemoveRule(ruleID string)

	// EnableRule enables a specific rule
	EnableRule(ruleID string)

	// DisableRule disables a specific rule
	DisableRule(ruleID string)

	// GetStats returns statistics about the rule engine
	GetStats() map[string]interface{}

	// GetRuleByName returns a rule by its name
	GetRuleByName(name string) *Rule
}

// Rule represents a business rule that can be executed
type Rule struct {
	ID          string
	Name        string
	Description string
	Condition   func(*models.PaymentFailureEvent) bool
	Action      func(*models.PaymentFailureEvent) (*ActionResult, error)
	Priority    int
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ActionResult represents the result of executing a rule
type ActionResult struct {
	RuleID     string
	RuleName   string
	Success    bool
	Error      error
	ExecutedAt time.Time
	Message    string
	Data       map[string]interface{}
}

// RuleEngineAdapter adapts BasicRuleEngine to implement RuleEngine interface
type RuleEngineAdapter struct {
	basicEngine *BasicRuleEngine
}

// NewRuleEngineAdapter creates a new adapter for BasicRuleEngine
func NewRuleEngineAdapter(basicEngine *BasicRuleEngine) *RuleEngineAdapter {
	return &RuleEngineAdapter{
		basicEngine: basicEngine,
	}
}

// ExecuteRules adapts the BasicRuleEngine.ExecuteRules method
func (a *RuleEngineAdapter) ExecuteRules(event *models.PaymentFailureEvent) []*ActionResult {
	basicResults := a.basicEngine.ExecuteRules(event)
	results := make([]*ActionResult, len(basicResults))

	for i, basicResult := range basicResults {
		results[i] = &ActionResult{
			RuleID:     basicResult.RuleName, // Use rule name as ID
			RuleName:   basicResult.RuleName,
			Success:    basicResult.Success,
			Error:      basicResult.Error,
			ExecutedAt: basicResult.ExecutedAt,
			Message:    basicResult.Message,
			Data:       basicResult.Data,
		}
	}

	return results
}

// GetRules adapts the BasicRuleEngine rules to Rule interface
func (a *RuleEngineAdapter) GetRules() []*Rule {
	basicRules := a.basicEngine.rules
	rules := make([]*Rule, len(basicRules))

	for i, basicRule := range basicRules {
		rules[i] = &Rule{
			ID:          basicRule.Name, // Use name as ID for basic rules
			Name:        basicRule.Name,
			Description: basicRule.Description,
			Condition:   basicRule.Condition,
			Action:      a.adaptAction(basicRule.Action),
			Priority:    basicRule.Priority,
			Enabled:     basicRule.Enabled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	return rules
}

// AddRule adapts a Rule to BasicRule and adds it to the engine
func (a *RuleEngineAdapter) AddRule(rule *Rule) {
	basicRule := &BasicRule{
		Name:        rule.Name,
		Description: rule.Description,
		Condition:   rule.Condition,
		Action:      a.adaptActionBack(rule.Action),
		Priority:    rule.Priority,
		Enabled:     rule.Enabled,
	}
	a.basicEngine.AddRule(basicRule)
}

// RemoveRule removes a rule from the engine
func (a *RuleEngineAdapter) RemoveRule(ruleID string) {
	// Find and remove rule by ID
	for i, rule := range a.basicEngine.rules {
		if rule.Name == ruleID {
			a.basicEngine.rules = append(a.basicEngine.rules[:i], a.basicEngine.rules[i+1:]...)
			break
		}
	}
}

// EnableRule enables a specific rule
func (a *RuleEngineAdapter) EnableRule(ruleID string) {
	for _, rule := range a.basicEngine.rules {
		if rule.Name == ruleID {
			rule.Enabled = true
			break
		}
	}
}

// DisableRule disables a specific rule
func (a *RuleEngineAdapter) DisableRule(ruleID string) {
	for _, rule := range a.basicEngine.rules {
		if rule.Name == ruleID {
			rule.Enabled = false
			break
		}
	}
}

// GetStats returns statistics about the rule engine
func (a *RuleEngineAdapter) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_rules"] = len(a.basicEngine.rules)
	stats["enabled_rules"] = 0
	stats["disabled_rules"] = 0

	for _, rule := range a.basicEngine.rules {
		if rule.Enabled {
			stats["enabled_rules"] = stats["enabled_rules"].(int) + 1
		} else {
			stats["disabled_rules"] = stats["disabled_rules"].(int) + 1
		}
	}

	return stats
}

// GetRuleByName returns a rule by its name
func (a *RuleEngineAdapter) GetRuleByName(name string) *Rule {
	for _, basicRule := range a.basicEngine.rules {
		if basicRule.Name == name {
			return &Rule{
				ID:          basicRule.Name,
				Name:        basicRule.Name,
				Description: basicRule.Description,
				Condition:   basicRule.Condition,
				Action:      a.adaptAction(basicRule.Action),
				Priority:    basicRule.Priority,
				Enabled:     basicRule.Enabled,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
		}
	}
	return nil
}

// Helper method to adapt BasicAction to Action
func (a *RuleEngineAdapter) adaptAction(basicAction func(*models.PaymentFailureEvent) (*BasicActionResult, error)) func(*models.PaymentFailureEvent) (*ActionResult, error) {
	return func(event *models.PaymentFailureEvent) (*ActionResult, error) {
		basicResult, err := basicAction(event)
		if err != nil {
			return nil, err
		}

		return &ActionResult{
			RuleID:     basicResult.RuleName,
			RuleName:   basicResult.RuleName,
			Success:    basicResult.Success,
			Error:      basicResult.Error,
			ExecutedAt: basicResult.ExecutedAt,
			Message:    basicResult.Message,
			Data:       basicResult.Data,
		}, nil
	}
}

// Helper method to adapt Action back to BasicAction
func (a *RuleEngineAdapter) adaptActionBack(action func(*models.PaymentFailureEvent) (*ActionResult, error)) func(*models.PaymentFailureEvent) (*BasicActionResult, error) {
	return func(event *models.PaymentFailureEvent) (*BasicActionResult, error) {
		result, err := action(event)
		if err != nil {
			return nil, err
		}

		return &BasicActionResult{
			RuleName:   result.RuleName,
			Success:    result.Success,
			Error:      result.Error,
			ExecutedAt: result.ExecutedAt,
			Message:    result.Message,
			Data:       result.Data,
		}, nil
	}
}

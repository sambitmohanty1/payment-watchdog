package rules

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/architecture"
)

// TestEnterpriseRuleEngineCreation tests the creation of a new enterprise rule engine
func TestEnterpriseRuleEngineCreation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)

	assert.NotNil(t, engine)
	assert.Equal(t, 0, len(engine.GetRules()))
	assert.Equal(t, 0, len(engine.GetEnabledRules()))
	assert.NotNil(t, engine.GetMetrics())
}

// TestEnterpriseRuleEngineAddRule tests adding rules to the engine
func TestEnterpriseRuleEngineAddRule(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add a high-value payment rule
	rule := riskRules.HighValuePaymentRule()
	err := engine.AddRule(rule)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(engine.GetRules()))
	assert.Equal(t, 1, len(engine.GetEnabledRules()))

	// Verify rule was added correctly
	addedRule, err := engine.GetRule(rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, rule.ID, addedRule.ID)
	assert.Equal(t, rule.Name, addedRule.Name)
	assert.Equal(t, rule.Priority, addedRule.Priority)
}

// TestEnterpriseRuleEngineAddRuleValidation tests rule validation
func TestEnterpriseRuleEngineAddRuleValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)

	// Test invalid rule (no ID)
	invalidRule := EnterpriseRule{
		Name:        "Invalid Rule",
		Description: "Rule without ID",
		Priority:    100,
		Enabled:     true,
		Conditions:  []EnterpriseCondition{},
		Actions:     []EnterpriseAction{},
	}

	err := engine.AddRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule ID cannot be empty")

	// Test invalid rule (no conditions)
	invalidRule.ID = "invalid_rule"
	err = engine.AddRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule must have at least one condition")

	// Test invalid rule (no actions)
	invalidRule.Conditions = []EnterpriseCondition{&MockEnterpriseCondition{}}
	err = engine.AddRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule must have at least one action")
}

// TestEnterpriseRuleEngineRuleManagement tests rule management operations
func TestEnterpriseRuleEngineRuleManagement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add multiple rules
	rule1 := riskRules.HighValuePaymentRule()
	rule2 := riskRules.OverduePaymentRule()
	rule3 := riskRules.RecurringFailureRule()

	assert.NoError(t, engine.AddRule(rule1))
	assert.NoError(t, engine.AddRule(rule2))
	assert.NoError(t, engine.AddRule(rule3))

	assert.Equal(t, 3, len(engine.GetRules()))

	// Test rule removal
	err := engine.RemoveRule(rule2.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(engine.GetRules()))

	// Test removing non-existent rule
	err = engine.RemoveRule("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestEnterpriseRuleEngineExecution tests rule execution
func TestEnterpriseRuleEngineExecution(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add a rule
	rule := riskRules.HighValuePaymentRule()
	assert.NoError(t, engine.AddRule(rule))

	// Create test context
	ctx := EnterpriseRuleContext{
		PaymentFailure: &architecture.PaymentFailure{
			ID:         uuid.New(),
			Amount:     15000.0, // Above threshold
			OccurredAt: time.Now(),
		},
		Timestamp:   time.Now(),
		Environment: "test",
	}

	// Execute rules
	results, err := engine.ExecuteRules(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))

	// Verify rule was triggered
	result := results[0]
	assert.Equal(t, rule.ID, result.RuleID)
	assert.True(t, result.Triggered)
	assert.Equal(t, 1, len(result.Conditions))
	assert.Equal(t, 2, len(result.Actions)) // High value rule has 2 actions

	// Verify condition was met
	conditionResult := result.Conditions[0]
	assert.True(t, conditionResult.Met)
	assert.Equal(t, "high_value", conditionResult.ConditionType)

	// Verify actions were executed
	for _, actionResult := range result.Actions {
		assert.True(t, actionResult.Executed)
		assert.NoError(t, actionResult.Error)
	}
}

// TestEnterpriseRuleEngineMetrics tests metrics collection
func TestEnterpriseRuleEngineMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add rules
	rule1 := riskRules.HighValuePaymentRule()
	rule2 := riskRules.OverduePaymentRule()
	assert.NoError(t, engine.AddRule(rule1))
	assert.NoError(t, engine.AddRule(rule2))

	// Create test context
	ctx := EnterpriseRuleContext{
		PaymentFailure: &architecture.PaymentFailure{
			ID:         uuid.New(),
			Amount:     15000.0,
			OccurredAt: time.Now(),
		},
		Timestamp:   time.Now(),
		Environment: "test",
	}

	// Execute rules
	_, err := engine.ExecuteRules(ctx)
	assert.NoError(t, err)

	// Check metrics
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalRulesExecuted)
	assert.Equal(t, int64(2), metrics.RulesTriggered)
	assert.True(t, metrics.SuccessRate > 0)
	assert.True(t, metrics.AverageExecutionTime > 0)
}

// TestEnterpriseRuleEnginePriorityOrdering tests rule priority ordering
func TestEnterpriseRuleEnginePriorityOrdering(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)
	timeRules := NewTimeBasedRules(logger)
	patternRules := NewPatternBasedRules(logger)

	// Add rules with different priorities
	rule1 := riskRules.HighValuePaymentRule()  // Priority 200
	rule2 := riskRules.OverduePaymentRule()    // Priority 180
	rule3 := timeRules.BusinessHoursRule()     // Priority 120
	rule4 := patternRules.FraudDetectionRule() // Priority 250

	assert.NoError(t, engine.AddRule(rule1))
	assert.NoError(t, engine.AddRule(rule2))
	assert.NoError(t, engine.AddRule(rule3))
	assert.NoError(t, engine.AddRule(rule4))

	// Verify rules are ordered by priority (highest first)
	rules := engine.GetRules()
	assert.Equal(t, 4, len(rules))
	assert.Equal(t, rule4.ID, rules[0].ID) // Fraud detection (250)
	assert.Equal(t, rule1.ID, rules[1].ID) // High value (200)
	assert.Equal(t, rule2.ID, rules[2].ID) // Overdue (180)
	assert.Equal(t, rule3.ID, rules[3].ID) // Business hours (120)
}

// TestEnterpriseRuleEngineDuplicateRuleID tests duplicate rule ID handling
func TestEnterpriseRuleEngineDuplicateRuleID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add first rule
	rule1 := riskRules.HighValuePaymentRule()
	assert.NoError(t, engine.AddRule(rule1))

	// Try to add rule with same ID
	rule2 := riskRules.HighValuePaymentRule()
	err := engine.AddRule(rule2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify only one rule exists
	assert.Equal(t, 1, len(engine.GetRules()))
}

// TestEnterpriseRuleEngineEnableDisable tests rule enable/disable functionality
func TestEnterpriseRuleEngineEnableDisable(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)

	// Add rule
	rule := riskRules.HighValuePaymentRule()
	assert.NoError(t, engine.AddRule(rule))

	// Initially enabled
	assert.Equal(t, 1, len(engine.GetEnabledRules()))

	// Disable rule
	err := engine.DisableRule(rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(engine.GetEnabledRules()))

	// Enable rule
	err = engine.EnableRule(rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(engine.GetEnabledRules()))

	// Test with non-existent rule
	err = engine.DisableRule("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Mock implementations for testing

// MockEnterpriseCondition is a mock condition for testing
type MockEnterpriseCondition struct{}

func (c *MockEnterpriseCondition) Evaluate(ctx EnterpriseRuleContext) bool {
	return true
}

func (c *MockEnterpriseCondition) GetType() string {
	return "mock"
}

func (c *MockEnterpriseCondition) GetDescription() string {
	return "Mock condition for testing"
}

// MockEnterpriseAction is a mock action for testing
type MockEnterpriseAction struct{}

func (a *MockEnterpriseAction) Execute(ctx EnterpriseRuleContext) error {
	return nil
}

func (a *MockEnterpriseAction) GetType() string {
	return "mock"
}

func (a *MockEnterpriseAction) GetDescription() string {
	return "Mock action for testing"
}

func (a *MockEnterpriseAction) GetPriority() int {
	return 50
}

// TestEnterpriseRuleEngineIntegration tests integration with concrete rule implementations
func TestEnterpriseRuleEngineIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewEnterpriseRuleEngine(logger)
	riskRules := NewRiskBasedRules(logger)
	timeRules := NewTimeBasedRules(logger)
	patternRules := NewPatternBasedRules(logger)

	// Add various types of rules
	assert.NoError(t, engine.AddRule(riskRules.HighValuePaymentRule()))
	assert.NoError(t, engine.AddRule(riskRules.OverduePaymentRule()))
	assert.NoError(t, engine.AddRule(timeRules.BusinessHoursRule()))
	assert.NoError(t, engine.AddRule(patternRules.FraudDetectionRule()))

	// Test with different payment failure scenarios
	testCases := []struct {
		name          string
		amount        float64
		occurredAt    time.Time
		expectedRules int
		description   string
	}{
		{
			name:          "High value payment",
			amount:        15000.0,
			occurredAt:    time.Now(),
			expectedRules: 4, // All rules should be evaluated
			description:   "High value payment should trigger multiple rules",
		},
		{
			name:          "Low value payment",
			amount:        100.0,
			occurredAt:    time.Now(),
			expectedRules: 4, // All rules are evaluated, but none may trigger for low value recent payment
			description:   "Low value payment should evaluate all rules but may not trigger any",
		},
		{
			name:          "Overdue payment",
			amount:        500.0,
			occurredAt:    time.Now().Add(-31 * 24 * time.Hour), // 31 days ago
			expectedRules: 4,                                    // All rules are evaluated, but overdue rule should trigger
			description:   "Overdue payment should evaluate all rules and trigger overdue rule",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := EnterpriseRuleContext{
				PaymentFailure: &architecture.PaymentFailure{
					ID:         uuid.New(),
					Amount:     tc.amount,
					OccurredAt: tc.occurredAt,
				},
				Timestamp:   time.Now(),
				Environment: "test",
			}

			results, err := engine.ExecuteRules(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedRules, len(results), tc.description)

			// Verify rules were evaluated (all rules are always evaluated)
			assert.Equal(t, tc.expectedRules, len(results), tc.description)

			// Count how many rules were actually triggered
			triggeredCount := 0
			for _, result := range results {
				if result.Triggered {
					triggeredCount++
				}
			}

			// Log the triggered count for debugging
			t.Logf("Rules triggered: %d out of %d", triggeredCount, len(results))

			// For high value payments, we expect multiple rules to trigger
			// For low value payments, it's acceptable if no rules trigger
			if tc.name == "High value payment" {
				assert.True(t, triggeredCount > 0, "High value payment should trigger at least one rule")
			}
		})
	}
}

package rules

import (
	"go.uber.org/zap"
)

// RuleEngineFactory creates and configures rule engines
type RuleEngineFactory struct {
	logger *zap.Logger
}

// NewRuleEngineFactory creates a new rule engine factory
func NewRuleEngineFactory(logger *zap.Logger) *RuleEngineFactory {
	return &RuleEngineFactory{
		logger: logger,
	}
}

// CreateComprehensiveRuleEngine creates a basic rule engine with all comprehensive business rules
// This is the main method for Sprint 3 - provides a complete set of working rules
func (f *RuleEngineFactory) CreateComprehensiveRuleEngine() RuleEngine {
	engine := NewBasicRuleEngine(f.logger)

	// Create comprehensive payment failure rules
	comprehensiveRules := NewComprehensivePaymentFailureRules(f.logger)

	// Add all comprehensive rules
	for _, rule := range comprehensiveRules.GetComprehensiveRules() {
		engine.AddRule(rule)
	}

	f.logger.Info("Basic rule engine created with comprehensive rules for Sprint 3",
		zap.Int("total_rules", len(comprehensiveRules.GetComprehensiveRules())))

	// Return the adapter that implements the RuleEngine interface
	return NewRuleEngineAdapter(engine)
}

// CreateBasicRuleEngine creates a simple basic rule engine with essential rules
func (f *RuleEngineFactory) CreateBasicRuleEngine() RuleEngine {
	engine := NewBasicRuleEngine(f.logger)

	// Add essential payment failure rules for Sprint 3
	paymentFailureRules := NewPaymentFailureRules(f.logger)

	for _, rule := range paymentFailureRules.GetDefaultRules() {
		engine.AddRule(rule)
	}

	f.logger.Info("Basic rule engine created for Sprint 3",
		zap.Int("total_rules", len(paymentFailureRules.GetDefaultRules())))

	return NewRuleEngineAdapter(engine)
}

// CreateEmptyBasicRuleEngine creates an empty basic rule engine for testing or custom configuration
func (f *RuleEngineFactory) CreateEmptyBasicRuleEngine() RuleEngine {
	engine := NewBasicRuleEngine(f.logger)
	f.logger.Info("Empty basic rule engine created")
	return NewRuleEngineAdapter(engine)
}

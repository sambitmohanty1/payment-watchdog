package rules

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/payment-watchdog/internal/architecture"
)

// EnterpriseRuleEngine represents the core business rules engine
type EnterpriseRuleEngine struct {
	rules    []EnterpriseRule
	executor EnterpriseRuleExecutor
	context  EnterpriseRuleContext
	metrics  *EnterpriseRuleMetrics
	logger   *zap.Logger
	mutex    sync.RWMutex
}

// EnterpriseRule represents a configurable business rule
type EnterpriseRule struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Priority    int                   `json:"priority"`
	Conditions  []EnterpriseCondition `json:"conditions"`
	Actions     []EnterpriseAction    `json:"actions"`
	Enabled     bool                  `json:"enabled"`
	Tags        []string              `json:"tags"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Metadata    map[string]string     `json:"metadata"`
}

// EnterpriseRuleContext provides context for rule evaluation
type EnterpriseRuleContext struct {
	PaymentFailure *architecture.PaymentFailure `json:"payment_failure"`
	Customer       *architecture.Customer       `json:"customer"`
	Company        *EnterpriseCompany           `json:"company"`
	Timestamp      time.Time                    `json:"timestamp"`
	Environment    string                       `json:"environment"`
	UserID         string                       `json:"user_id"`
	SessionID      string                       `json:"session_id"`
	Metadata       map[string]interface{}       `json:"metadata"`
}

// EnterpriseCompany represents company information for rule context
type EnterpriseCompany struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Industry    string            `json:"industry"`
	Size        string            `json:"size"`
	RiskProfile string            `json:"risk_profile"`
	Settings    map[string]string `json:"settings"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// EnterpriseCondition represents a rule condition that must be met
type EnterpriseCondition interface {
	Evaluate(ctx EnterpriseRuleContext) bool
	GetType() string
	GetDescription() string
}

// EnterpriseAction represents an action to be executed when a rule is triggered
type EnterpriseAction interface {
	Execute(ctx EnterpriseRuleContext) error
	GetType() string
	GetDescription() string
	GetPriority() int
}

// EnterpriseRuleExecutor handles the execution of rules
type EnterpriseRuleExecutor interface {
	ExecuteRule(rule EnterpriseRule, ctx EnterpriseRuleContext) (*EnterpriseRuleResult, error)
	ExecuteRules(rules []EnterpriseRule, ctx EnterpriseRuleContext) ([]*EnterpriseRuleResult, error)
}

// EnterpriseRuleResult represents the result of rule execution
type EnterpriseRuleResult struct {
	RuleID     string                       `json:"rule_id"`
	RuleName   string                       `json:"rule_name"`
	Triggered  bool                         `json:"triggered"`
	ExecutedAt time.Time                    `json:"executed_at"`
	Duration   time.Duration                `json:"duration"`
	Actions    []*EnterpriseActionResult    `json:"actions"`
	Conditions []*EnterpriseConditionResult `json:"conditions"`
	Error      error                        `json:"error,omitempty"`
	Metadata   map[string]interface{}       `json:"metadata"`
}

// EnterpriseActionResult represents the result of an action execution
type EnterpriseActionResult struct {
	ActionType string                 `json:"action_type"`
	ActionName string                 `json:"action_name"`
	Executed   bool                   `json:"executed"`
	ExecutedAt time.Time              `json:"executed_at"`
	Duration   time.Duration          `json:"duration"`
	Error      error                  `json:"error,omitempty"`
	Output     map[string]interface{} `json:"output"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// EnterpriseConditionResult represents the result of a condition evaluation
type EnterpriseConditionResult struct {
	ConditionType string                 `json:"condition_type"`
	ConditionName string                 `json:"condition_name"`
	Met           bool                   `json:"met"`
	EvaluatedAt   time.Time              `json:"evaluated_at"`
	Duration      time.Duration          `json:"duration"`
	Value         interface{}            `json:"value"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// EnterpriseRuleMetrics tracks rule engine performance and statistics
type EnterpriseRuleMetrics struct {
	TotalRulesExecuted   int64         `json:"total_rules_executed"`
	RulesTriggered       int64         `json:"rules_triggered"`
	ActionsExecuted      int64         `json:"actions_executed"`
	ConditionsEvaluated  int64         `json:"conditions_evaluated"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionTime    time.Time     `json:"last_execution_time"`
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	ErrorCount           int64         `json:"error_count"`
	SuccessRate          float64       `json:"success_rate"`
	mutex                sync.RWMutex
}

// NewEnterpriseRuleEngine creates a new enterprise rule engine instance
func NewEnterpriseRuleEngine(logger *zap.Logger) *EnterpriseRuleEngine {
	return &EnterpriseRuleEngine{
		rules:   make([]EnterpriseRule, 0),
		metrics: &EnterpriseRuleMetrics{},
		logger:  logger,
		mutex:   sync.RWMutex{},
	}
}

// AddRule adds a rule to the engine
func (e *EnterpriseRuleEngine) AddRule(rule EnterpriseRule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Validate rule
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	// Check for duplicate rule ID
	for _, existingRule := range e.rules {
		if existingRule.ID == rule.ID {
			return fmt.Errorf("rule with ID %s already exists", rule.ID)
		}
	}

	e.rules = append(e.rules, rule)

	// Sort rules by priority (highest first)
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority > e.rules[j].Priority
	})

	e.logger.Info("Enterprise rule added",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
		zap.Int("priority", rule.Priority))

	return nil
}

// RemoveRule removes a rule by ID
func (e *EnterpriseRuleEngine) RemoveRule(ruleID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			e.logger.Info("Enterprise rule removed",
				zap.String("rule_id", ruleID))
			return nil
		}
	}

	return fmt.Errorf("rule with ID %s not found", ruleID)
}

// GetRule retrieves a rule by ID
func (e *EnterpriseRuleEngine) GetRule(ruleID string) (*EnterpriseRule, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	for _, rule := range e.rules {
		if rule.ID == ruleID {
			return &rule, nil
		}
	}

	return nil, fmt.Errorf("rule with ID %s not found", ruleID)
}

// GetRules returns all registered rules
func (e *EnterpriseRuleEngine) GetRules() []EnterpriseRule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	rules := make([]EnterpriseRule, len(e.rules))
	copy(rules, e.rules)
	return rules
}

// GetEnabledRules returns only enabled rules
func (e *EnterpriseRuleEngine) GetEnabledRules() []EnterpriseRule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	var enabledRules []EnterpriseRule
	for _, rule := range e.rules {
		if rule.Enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	return enabledRules
}

// EnableRule enables a specific rule by ID
func (e *EnterpriseRuleEngine) EnableRule(ruleID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i := range e.rules {
		if e.rules[i].ID == ruleID {
			e.rules[i].Enabled = true
			e.rules[i].UpdatedAt = time.Now()
			e.logger.Info("Enterprise rule enabled",
				zap.String("rule_id", ruleID))
			return nil
		}
	}

	return fmt.Errorf("rule with ID %s not found", ruleID)
}

// DisableRule disables a specific rule by ID
func (e *EnterpriseRuleEngine) DisableRule(ruleID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i := range e.rules {
		if e.rules[i].ID == ruleID {
			e.rules[i].Enabled = false
			e.rules[i].UpdatedAt = time.Now()
			e.logger.Info("Enterprise rule disabled",
				zap.String("rule_id", ruleID))
			return nil
		}
	}

	return fmt.Errorf("rule with ID %s not found", ruleID)
}

// ExecuteRules executes all applicable rules for a given context
func (e *EnterpriseRuleEngine) ExecuteRules(ctx EnterpriseRuleContext) ([]*EnterpriseRuleResult, error) {
	e.mutex.RLock()
	enabledRules := make([]EnterpriseRule, 0)
	for _, rule := range e.rules {
		if rule.Enabled {
			enabledRules = append(enabledRules, rule)
		}
	}
	e.mutex.RUnlock()

	var results []*EnterpriseRuleResult
	startTime := time.Now()

	e.logger.Info("Executing enterprise rules",
		zap.Int("total_rules", len(enabledRules)),
		zap.String("context_id", fmt.Sprintf("%v", ctx.PaymentFailure.ID)))

	for _, rule := range enabledRules {
		result, err := e.executeRule(rule, ctx)
		if err != nil {
			e.logger.Error("Enterprise rule execution failed",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
			result = &EnterpriseRuleResult{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Triggered:  false,
				ExecutedAt: time.Now(),
				Error:      err,
			}
		}
		results = append(results, result)
	}

	executionTime := time.Since(startTime)
	e.recordExecutionMetrics(len(enabledRules), len(results), executionTime)

	e.logger.Info("Enterprise rules execution completed",
		zap.Int("rules_executed", len(results)),
		zap.Duration("total_time", executionTime))

	return results, nil
}

// executeRule executes a single rule
func (e *EnterpriseRuleEngine) executeRule(rule EnterpriseRule, ctx EnterpriseRuleContext) (*EnterpriseRuleResult, error) {
	startTime := time.Now()

	// Evaluate conditions
	var conditionResults []*EnterpriseConditionResult
	conditionsMet := true

	for _, condition := range rule.Conditions {
		conditionStart := time.Now()
		met := condition.Evaluate(ctx)
		conditionDuration := time.Since(conditionStart)

		conditionResult := &EnterpriseConditionResult{
			ConditionType: condition.GetType(),
			ConditionName: condition.GetDescription(),
			Met:           met,
			EvaluatedAt:   time.Now(),
			Duration:      conditionDuration,
		}
		conditionResults = append(conditionResults, conditionResult)

		if !met {
			conditionsMet = false
		}
	}

	// Execute actions if conditions are met
	var actionResults []*EnterpriseActionResult
	if conditionsMet {
		for _, action := range rule.Actions {
			actionStart := time.Now()
			err := action.Execute(ctx)
			actionDuration := time.Since(actionStart)

			actionResult := &EnterpriseActionResult{
				ActionType: action.GetType(),
				ActionName: action.GetDescription(),
				Executed:   err == nil,
				ExecutedAt: time.Now(),
				Duration:   actionDuration,
				Error:      err,
			}
			actionResults = append(actionResults, actionResult)
		}
	}

	executionTime := time.Since(startTime)

	return &EnterpriseRuleResult{
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		Triggered:  conditionsMet,
		ExecutedAt: time.Now(),
		Duration:   executionTime,
		Actions:    actionResults,
		Conditions: conditionResults,
	}, nil
}

// GetMetrics returns the current rule engine metrics
func (e *EnterpriseRuleEngine) GetMetrics() *EnterpriseRuleMetrics {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	metrics := &EnterpriseRuleMetrics{}
	*metrics = *e.metrics
	return metrics
}

// validateRule validates a rule before adding it
func (e *EnterpriseRuleEngine) validateRule(rule EnterpriseRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("rule must have at least one condition")
	}
	if len(rule.Actions) == 0 {
		return fmt.Errorf("rule must have at least one action")
	}
	return nil
}

// recordExecutionMetrics records execution metrics
func (e *EnterpriseRuleEngine) recordExecutionMetrics(totalRules, triggeredRules int, executionTime time.Duration) {
	e.metrics.mutex.Lock()
	defer e.metrics.mutex.Unlock()

	e.metrics.TotalRulesExecuted += int64(totalRules)
	e.metrics.RulesTriggered += int64(triggeredRules)
	e.metrics.TotalExecutionTime += executionTime
	e.metrics.LastExecutionTime = time.Now()

	if e.metrics.TotalRulesExecuted > 0 {
		e.metrics.AverageExecutionTime = e.metrics.TotalExecutionTime / time.Duration(e.metrics.TotalRulesExecuted)
		e.metrics.SuccessRate = float64(e.metrics.RulesTriggered) / float64(e.metrics.TotalRulesExecuted) * 100
	}
}

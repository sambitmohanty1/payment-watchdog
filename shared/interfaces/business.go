package interfaces

import (
	"context"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
)

// AnalyticsEngine defines interface for payment analytics and pattern detection
type AnalyticsEngine interface {
	AnalyzePaymentPatterns(ctx context.Context, companyID string, timeRange string) (*AnalyticsResult, error)
	GetFailurePredictions(ctx context.Context, companyID string) (*FailurePredictions, error)
	GetRecoveryRecommendations(ctx context.Context, paymentEvent *events.PaymentEvent) (*RecoveryRecommendations, error)
}

// RuleEngine defines interface for payment processing rules
type RuleEngine interface {
	EvaluatePaymentFailure(ctx context.Context, paymentEvent *events.PaymentEvent) (*RuleEvaluation, error)
	ApplyRecoveryStrategy(ctx context.Context, paymentEvent *events.PaymentEvent) (*RecoveryStrategy, error)
}

// MediatorService defines interface for payment mediator services (QuickBooks, Xero, etc.)
type MediatorService interface {
	SyncInvoices(ctx context.Context, companyID string) (*SyncResult, error)
	ProcessPaymentThroughMediator(ctx context.Context, paymentEvent *events.PaymentEvent) (*MediatorResult, error)
	GetAccountingStatus(ctx context.Context, companyID string) (*AccountingStatus, error)
}

// AnalyticsResult represents analytics analysis results
type AnalyticsResult struct {
	Patterns        []PaymentPattern `json:"patterns"`
	FailureRate     float64          `json:"failure_rate"`
	SuccessRate     float64          `json:"success_rate"`
	Recommendations []string         `json:"recommendations"`
}

// FailurePredictions represents ML-based failure predictions
type FailurePredictions struct {
	HighRiskPayments  []string    `json:"high_risk_payments"`
	OptimalRetryTimes []time.Time `json:"optimal_retry_times"`
	ConfidenceScore   float64     `json:"confidence_score"`
}

// RecoveryRecommendations represents recovery strategy recommendations
type RecoveryRecommendations struct {
	Strategy         string    `json:"strategy"`
	RetryAttempts    int       `json:"retry_attempts"`
	AmountToCharge   float64   `json:"amount_to_charge"`
	NextRetryTime    time.Time `json:"next_retry_time"`
	MaxRetryAttempts int       `json:"max_retry_attempts"`
}

// RuleEvaluation represents rule evaluation results
type RuleEvaluation struct {
	RuleName   string  `json:"rule_name"`
	IsMatch    bool    `json:"is_match"`
	Action     string  `json:"action"`
	Confidence float64 `json:"confidence"`
}

// RecoveryStrategy represents recovery strategy to apply
type RecoveryStrategy struct {
	StrategyType  string    `json:"strategy_type"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	NextAttempt   time.Time `json:"next_attempt"`
	MaxAttempts   int       `json:"max_attempts"`
}

// SyncResult represents mediator sync results
type SyncResult struct {
	SyncedCount  int      `json:"synced_count"`
	LastSyncTime string   `json:"last_sync_time"`
	Errors       []string `json:"errors"`
}

// MediatorResult represents mediator operation results
type MediatorResult struct {
	MediatorType  string  `json:"mediator_type"`
	Success       bool    `json:"success"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
}

// AccountingStatus represents accounting system status
type AccountingStatus struct {
	IsConnected bool    `json:"is_connected"`
	LastSync    string  `json:"last_sync"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
}

// PaymentPattern represents identified payment patterns
type PaymentPattern struct {
	PatternType   string  `json:"pattern_type"`
	Frequency     int     `json:"frequency"`
	AverageAmount float64 `json:"average_amount"`
	TimeOfDay     string  `json:"time_of_day"`
	DayOfWeek     string  `json:"day_of_week"`
}

package rules

import (
	"github.com/sambitmohanty1/payment-watchdog/worker/internal/models"
)

// PaymentFailureRule defines a rule for evaluating payment failures
type PaymentFailureRule interface {
	Name() string
	Evaluate(event *models.PaymentFailureEvent) (bool, error)
	Priority() int
}

// HighValueTransactionRule checks if the failure amount is significant
type HighValueTransactionRule struct {
	ThresholdCents int64
}

func NewHighValueTransactionRule(thresholdCents int64) *HighValueTransactionRule {
	return &HighValueTransactionRule{
		ThresholdCents: thresholdCents,
	}
}

func (r *HighValueTransactionRule) Name() string {
	return "HighValueTransaction"
}

func (r *HighValueTransactionRule) Evaluate(event *models.PaymentFailureEvent) (bool, error) {
	// LOGIC FIX: Compare cents to cents
	return event.AmountCents >= r.ThresholdCents, nil
}

func (r *HighValueTransactionRule) Priority() int {
	return 100
}

// RecurringFailureRule checks if this customer has failed recently
type RecurringFailureRule struct {
	MaxRetries int
}

func (r *RecurringFailureRule) Name() string {
	return "RecurringFailure"
}

func (r *RecurringFailureRule) Evaluate(event *models.PaymentFailureEvent) (bool, error) {
	return event.RetryCount >= r.MaxRetries, nil
}

func (r *RecurringFailureRule) Priority() int {
	return 50
}

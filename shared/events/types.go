package events

import (
	"time"

	"github.com/google/uuid"
)

// PaymentEvent represents a payment-related event in system
type PaymentEvent struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	CompanyID string                 `json:"company_id"`
	PaymentID string                 `json:"payment_id"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// EventTypes defines all event types in the system
const (
	PaymentFailureDetected  = "payment.failure.detected"
	PaymentFailureProcessed = "payment.failure.processed"
	PaymentRetryScheduled   = "payment.retry.scheduled"
	PaymentRecovered        = "payment.recovered"
)

// NewPaymentEvent creates a new payment event
func NewPaymentEvent(eventType, companyID, paymentID string, amount float64, currency string, metadata map[string]interface{}) *PaymentEvent {
	return &PaymentEvent{
		ID:        uuid.New().String(),
		EventType: eventType,
		CompanyID: companyID,
		PaymentID: paymentID,
		Amount:    amount,
		Currency:  currency,
		Status:    "created",
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}

// NewPaymentFailureEvent creates a payment failure event
func NewPaymentFailureEvent(companyID, paymentID string, amount float64, currency string, provider string, reason string) *PaymentEvent {
	return NewPaymentEvent(
		PaymentFailureDetected,
		companyID,
		paymentID,
		amount,
		currency,
		map[string]interface{}{
			"provider": provider,
			"reason":   reason,
		},
	)
}

// NewPaymentProcessedEvent creates a payment processed event
func NewPaymentProcessedEvent(companyID, paymentID string, amount float64, currency string, status string) *PaymentEvent {
	return NewPaymentEvent(
		PaymentFailureProcessed,
		companyID,
		paymentID,
		amount,
		currency,
		map[string]interface{}{
			"status": status,
		},
	)
}

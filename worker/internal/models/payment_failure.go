package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentFailureEvent represents a failed payment event from payment providers
type PaymentFailureEvent struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID  string    `json:"company_id" gorm:"not null;index"`
	ProviderID string    `json:"provider_id" gorm:"not null"`
	EventID    string    `json:"event_id" gorm:"not null;uniqueIndex"`
	EventType  string    `json:"event_type" gorm:"not null"`

	// Payment Details
	PaymentIntentID string     `json:"payment_intent_id"`
	TransactionID   string     `json:"transaction_id"`
	
	// FINANCIAL INTEGRITY FIX: int64 for cents (matches API)
	AmountCents     int64      `json:"amount_cents"` 
	Currency        string     `json:"currency" gorm:"default:'AUD'"`
	
	CustomerID      string     `json:"customer_id"`
	CustomerEmail   string     `json:"customer_email"`
	CustomerName    string     `json:"customer_name"`
	CustomerPhone   string     `json:"customer_phone"`
	Provider        string     `json:"provider"`
	RetryCount      int        `json:"retry_count" gorm:"default:0"`
	DueDate         *time.Time `json:"due_date,omitempty"`

	// Failure Information
	FailureReason  string `json:"failure_reason"`
	FailureCode    string `json:"failure_code"`
	FailureMessage string `json:"failure_message"`

	// Processing Status
	Status      string     `json:"status" gorm:"default:'received'"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	AlertedAt   *time.Time `json:"alerted_at,omitempty"`

	RawEventData   string    `json:"raw_event_data" gorm:"type:jsonb"`
	NormalizedData string    `json:"normalized_data" gorm:"type:jsonb"`
	
	WebhookReceivedAt time.Time `json:"webhook_received_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Helper to get human-readable amount for notifications
func (p *PaymentFailureEvent) AmountHuman() float64 {
	return float64(p.AmountCents) / 100.0
}

func (p *PaymentFailureEvent) TableName() string { return "payment_failure_events" }

func (p *PaymentFailureEvent) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

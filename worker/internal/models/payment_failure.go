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
	ProviderID string    `json:"provider_id" gorm:"not null"`          // stripe, paypal, etc.
	EventID    string    `json:"event_id" gorm:"not null;uniqueIndex"` // Provider's event ID for idempotency
	EventType  string    `json:"event_type" gorm:"not null"`           // payment_intent.payment_failed, etc.

	// Payment Details
	PaymentIntentID string `json:"payment_intent_id"`
	TransactionID   string `json:"transaction_id"`

	// FINANCIAL INTEGRITY FIX: Use int64 for cents. Never use float64 for money.
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency" gorm:"default:'AUD'"`

	// Computed field for backward compatibility with rules
	Amount float64 `json:"amount" gorm:"-"` // Computed from AmountCents

	CustomerID    string     `json:"customer_id"`
	CustomerEmail string     `json:"customer_email"`
	CustomerName  string     `json:"customer_name"`
	CustomerPhone string     `json:"customer_phone"`
	Provider      string     `json:"provider"`
	RetryCount    int        `json:"retry_count" gorm:"default:0"`
	DueDate       *time.Time `json:"due_date,omitempty"`

	// Failure Information
	FailureReason  string `json:"failure_reason"` // card_declined, etc.
	FailureCode    string `json:"failure_code"`
	FailureMessage string `json:"failure_message"`

	// Processing Status
	Status      string     `json:"status" gorm:"default:'received'"` // received, processed, retried
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	AlertedAt   *time.Time `json:"alerted_at,omitempty"`

	// Raw Data
	RawEventData   string `json:"raw_event_data" gorm:"type:jsonb"`
	NormalizedData string `json:"normalized_data" gorm:"type:jsonb"`

	// Metadata
	WebhookReceivedAt time.Time `json:"webhook_received_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DeadLetterEntry represents a failed event stored in Postgres (Reliability)
type DeadLetterEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventID   string    `gorm:"index" json:"event_id"`
	Payload   []byte    `gorm:"type:jsonb" json:"payload"`
	Error     string    `json:"error"`
	CompanyID string    `json:"company_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerCommunication tracks outreach attempts
type CustomerCommunication struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentFailureID uuid.UUID `json:"payment_failure_id" gorm:"not null;index"`
	CompanyID        string    `json:"company_id" gorm:"not null;index"`

	// Communication Details
	CommunicationType string `json:"communication_type"` // email, sms
	Channel           string `json:"channel"`            // email, sms
	Recipient         string `json:"recipient"`
	TemplateID        string `json:"template_id"`
	Subject           string `json:"subject"`
	Content           string `json:"content"`

	// Delivery Status
	Status            string                 `json:"status"` // sent, delivered, opened, clicked, failed
	ExternalID        string                 `json:"external_id"`
	ProviderMessageID string                 `json:"provider_message_id"`
	DeliveryResponse  map[string]interface{} `json:"delivery_response" gorm:"type:jsonb"`
	Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// Tracking
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	OpenedAt    *time.Time `json:"opened_at,omitempty"`
	ClickedAt   *time.Time `json:"clicked_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Table names
func (p *PaymentFailureEvent) TableName() string { return "payment_failure_events" }
func (d *DeadLetterEntry) TableName() string     { return "dead_letter_entries" }

// Hooks
func (p *PaymentFailureEvent) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	// Compute Amount from AmountCents for backward compatibility
	p.Amount = float64(p.AmountCents) / 100.0

	return nil
}

// GetAmount returns the amount in dollars (computed from cents)
func (p *PaymentFailureEvent) GetAmount() float64 {
	return float64(p.AmountCents) / 100.0
}

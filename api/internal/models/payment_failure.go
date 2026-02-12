//go:build !build
// +build !build

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
	PaymentIntentID string     `json:"payment_intent_id"`
	TransactionID   string     `json:"transaction_id"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency" gorm:"default:'AUD'"`
	CustomerID      string     `json:"customer_id"`
	CustomerEmail   string     `json:"customer_email"`
	CustomerName    string     `json:"customer_name"`
	CustomerPhone   string     `json:"customer_phone"`
	Provider        string     `json:"provider"`
	RetryCount      int        `json:"retry_count" gorm:"default:0"`
	DueDate         *time.Time `json:"due_date,omitempty"`

	// Failure Information
	FailureReason  string `json:"failure_reason"` // card_declined, insufficient_funds, etc.
	FailureCode    string `json:"failure_code"`
	FailureMessage string `json:"failure_message"`

	// Processing Status
	Status      string     `json:"status" gorm:"default:'received'"` // received, processed, alerted, retried, resolved
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

// RetryAttempt tracks payment retry attempts
type RetryAttempt struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentFailureID uuid.UUID `json:"payment_failure_id" gorm:"not null;index"`
	CompanyID        string    `json:"company_id" gorm:"not null;index"`

	// Retry Details
	AttemptNumber int     `json:"attempt_number"`
	RetryAmount   float64 `json:"retry_amount"`
	RetryMethod   string  `json:"retry_method"` // immediate, scheduled, manual
	Status        string  `json:"status"`       // pending, success, failed, cancelled

	// Provider Response
	ProviderRetryID  string                 `json:"provider_retry_id"`
	ProviderResponse map[string]interface{} `json:"provider_response" gorm:"type:jsonb"`

	// Timing
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	AttemptedAt *time.Time `json:"attempted_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

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
	Channel           string `json:"channel"` // email, sms
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

// Company represents a company using the service
type Company struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name   string    `json:"name" gorm:"not null"`
	Domain string    `json:"domain"`
	Status string    `json:"status" gorm:"default:'active'"`

	// Configuration
	StripeAccountID string                 `json:"stripe_account_id"`
	AlertSettings   map[string]interface{} `json:"alert_settings" gorm:"type:jsonb"`
	RetrySettings   map[string]interface{} `json:"retry_settings" gorm:"type:jsonb"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Table names
func (p *PaymentFailureEvent) TableName() string   { return "payment_failure_events" }
func (r *RetryAttempt) TableName() string          { return "retry_attempts" }
func (c *CustomerCommunication) TableName() string { return "customer_communications" }
func (c *Company) TableName() string               { return "companies" }

// Hooks
func (p *PaymentFailureEvent) BeforeCreate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (r *RetryAttempt) BeforeCreate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

func (c *CustomerCommunication) BeforeCreate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (p *PaymentFailureEvent) BeforeUpdate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	p.UpdatedAt = time.Now()
	return nil
}

func (r *RetryAttempt) BeforeUpdate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	r.UpdatedAt = time.Now()
	return nil
}

func (c *CustomerCommunication) BeforeUpdate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	c.UpdatedAt = time.Now()
	return nil
}

func (c *Company) BeforeUpdate(tx *gorm.DB) error {
	// Safety check: only execute if we have a valid database connection
	if tx == nil || tx.Statement == nil {
		return nil
	}

	c.UpdatedAt = time.Now()
	return nil
}

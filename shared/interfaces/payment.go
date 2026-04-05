package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sambitmohanty1/payment-watchdog/shared/events"
)

// DatabaseInterface defines interface for database operations
type DatabaseInterface interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetHealthStatus() *DatabaseHealthStatus
}

// DatabaseHealthStatus represents database health status
type DatabaseHealthStatus struct {
	IsConnected bool   `json:"is_connected"`
	LastCheck   string `json:"last_check"`
	ErrorCount  int    `json:"error_count"`
}

// PaymentProcessor defines interface for processing payment events
type PaymentProcessor interface {
	ProcessPaymentFailure(ctx context.Context, event *events.PaymentEvent) error
	ValidatePaymentData(ctx context.Context, payment map[string]interface{}) error
	GetProcessorStatus() *ProcessorStatus
}

// ProcessorStatus represents the status of a payment processor
type ProcessorStatus struct {
	IsHealthy   bool   `json:"is_healthy"`
	LastProcess string `json:"last_process"`
	ErrorCount  int    `json:"error_count"`
	Throughput  int    `json:"throughput"`
}

// EventBusInterface defines interface for event bus operations
type EventBusInterface interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	Close() error
	GetHealthStatus() *EventBusStatus
}

// EventHandler defines interface for event handling
type EventHandler interface {
	Handle(ctx context.Context, event interface{}) error
}

// Subscription defines interface for event subscription
type Subscription interface {
	ID() string
	Topic() string
	Unsubscribe() error
}

// EventBusStatus represents the status of the event bus
type EventBusStatus struct {
	IsConnected     bool `json:"is_connected"`
	SubscriberCount int  `json:"subscriber_count"`
	MessageCount    int  `json:"message_count"`
	ErrorCount      int  `json:"error_count"`
}

// Customer represents a customer in the system
type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaymentFailure represents a payment failure event
type PaymentFailure struct {
	ID                uuid.UUID              `json:"id"`
	ProviderID        string                 `json:"provider_id"`
	ProviderEventID   string                 `json:"provider_event_id"`
	ProviderEventType string                 `json:"provider_event_type"`
	CompanyID         string                 `json:"company_id"`
	CustomerID        string                 `json:"customer_id"`
	CustomerName      string                 `json:"customer_name"`
	CustomerEmail     string                 `json:"customer_email"`
	PaymentID         string                 `json:"payment_id"`
	TransactionID     string                 `json:"transaction_id"`
	InvoiceID         string                 `json:"invoice_id"`
	InvoiceNumber     string                 `json:"invoice_number"`
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	FailureReason     string                 `json:"failure_reason"`
	FailureCode       string                 `json:"failure_code"`
	FailureMessage    string                 `json:"failure_message"`
	Status            PaymentFailureStatus   `json:"status"`
	Priority          PaymentFailurePriority `json:"priority"`
	RiskScore         float64                `json:"risk_score"`
	BusinessCategory  string                 `json:"business_category"`
	IssueDate         *time.Time             `json:"issue_date,omitempty"`
	DueDate           *time.Time             `json:"due_date,omitempty"`
	OccurredAt        time.Time              `json:"occurred_at"`
	DetectedAt        time.Time              `json:"detected_at"`
	LineItems         []interface{}          `json:"line_items,omitempty"`
	ProviderMetadata  map[string]interface{} `json:"provider_metadata,omitempty"`
	RawData           map[string]interface{} `json:"raw_data,omitempty"`
	Metadata          map[string]interface{} `json:"metadata"`
	SyncSource        string                 `json:"sync_source"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// PaymentFailureStatus represents the status of a payment failure
type PaymentFailureStatus string

const (
	PaymentFailureStatusReceived   PaymentFailureStatus = "received"
	PaymentFailureStatusProcessing PaymentFailureStatus = "processing"
	PaymentFailureStatusResolved   PaymentFailureStatus = "resolved"
	PaymentFailureStatusFailed     PaymentFailureStatus = "failed"
)

// PaymentFailurePriority represents the priority of a payment failure
type PaymentFailurePriority string

const (
	PaymentFailurePriorityLow      PaymentFailurePriority = "low"
	PaymentFailurePriorityMedium   PaymentFailurePriority = "medium"
	PaymentFailurePriorityHigh     PaymentFailurePriority = "high"
	PaymentFailurePriorityCritical PaymentFailurePriority = "critical"
)

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	ProviderID        string    `json:"provider_id"`
	RequestsRemaining int       `json:"requests_remaining"`
	ResetTime         time.Time `json:"reset_time"`
	Limit             int       `json:"limit"`
	Window            string    `json:"window"`
}

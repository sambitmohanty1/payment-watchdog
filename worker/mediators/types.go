package mediators

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentFailure represents a unified payment failure across all providers
type PaymentFailure struct {
	ID                uuid.UUID              `json:"id"`
	CompanyID         string                 `json:"company_id"`
	
	// Provider identification
	ProviderID        string                 `json:"provider_id"`
	ProviderEventID   string                 `json:"provider_event_id"`
	ProviderEventType string                 `json:"provider_event_type"`
	
	// Payment details
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	PaymentMethod     string                 `json:"payment_method"`
	
	// Customer information
	CustomerID        string                 `json:"customer_id"`
	CustomerName      string                 `json:"customer_name"`
	CustomerEmail     string                 `json:"customer_email"`
	CustomerPhone     string                 `json:"customer_phone"`
	
	// Failure information
	FailureReason     string                 `json:"failure_reason"`
	FailureCode       string                 `json:"failure_code"`
	FailureMessage    string                 `json:"failure_message"`
	
	// Business context
	InvoiceID         string                 `json:"invoice_id"`
	InvoiceNumber     string                 `json:"invoice_number"`
	DueDate           *time.Time             `json:"due_date"`
	BusinessCategory  string                 `json:"business_category"`
	
	// Processing status
	Status            PaymentFailureStatus   `json:"status"`
	Priority          PaymentFailurePriority `json:"priority"`
	RiskScore         float64                `json:"risk_score"`
	
	// Timestamps
	OccurredAt        time.Time              `json:"occurred_at"`
	DetectedAt        time.Time              `json:"detected_at"`
	ProcessedAt       *time.Time             `json:"processed_at"`
	ResolvedAt        *time.Time             `json:"resolved_at"`
	
	// Integration details
	IntegrationID     *uuid.UUID             `json:"integration_id,omitempty"`
	OriginalInvoiceID string                 `json:"original_invoice_id,omitempty"`
	OriginalPaymentID string                 `json:"original_payment_id,omitempty"`
	SyncSource        string                 `json:"sync_source"`
	
	// Raw and normalized data
	RawData           json.RawMessage        `json:"raw_data"`
	NormalizedData    json.RawMessage        `json:"normalized_data"`
	
	// Provider-specific metadata
	ProviderMetadata  map[string]interface{} `json:"provider_metadata"`
	
	// Metadata
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	
	// Audit fields
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// PaymentFailureStatus represents the processing status
type PaymentFailureStatus string

const (
	PaymentFailureStatusReceived   PaymentFailureStatus = "received"
	PaymentFailureStatusProcessing PaymentFailureStatus = "processing"
	PaymentFailureStatusAnalyzed   PaymentFailureStatus = "analyzed"
	PaymentFailureStatusAlerted    PaymentFailureStatus = "alerted"
	PaymentFailureStatusRetrying   PaymentFailureStatus = "retrying"
	PaymentFailureStatusResolved   PaymentFailureStatus = "resolved"
	PaymentFailureStatusEscalated  PaymentFailureStatus = "escalated"
)

// PaymentFailurePriority represents the business priority
type PaymentFailurePriority string

const (
	PaymentFailurePriorityCritical PaymentFailurePriority = "critical"
	PaymentFailurePriorityHigh     PaymentFailurePriority = "high"
	PaymentFailurePriorityMedium   PaymentFailurePriority = "medium"
	PaymentFailurePriorityLow      PaymentFailurePriority = "low"
)

// Invoice represents a unified invoice across all providers
type Invoice struct {
	ID                uuid.UUID              `json:"id"`
	CompanyID         string                 `json:"company_id"`
	
	// Provider identification
	ProviderID        string                 `json:"provider_id"`
	ProviderInvoiceID string                 `json:"provider_invoice_id"`
	
	// Invoice details
	InvoiceNumber     string                 `json:"invoice_number"`
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	Status            string                 `json:"status"`
	
	// Customer information
	CustomerID        string                 `json:"customer_id"`
	CustomerName      string                 `json:"customer_name"`
	CustomerEmail     string                 `json:"customer_email"`
	
	// Dates
	IssueDate         time.Time              `json:"issue_date"`
	DueDate           time.Time              `json:"due_date"`
	PaidDate          *time.Time             `json:"paid_date"`
	
	// Line items
	LineItems         json.RawMessage        `json:"line_items"`
	
	// Provider-specific data
	ProviderMetadata  map[string]interface{} `json:"provider_metadata"`
	
	// Audit fields
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Customer represents a unified customer across all providers
type Customer struct {
	ID                uuid.UUID              `json:"id"`
	CompanyID         string                 `json:"company_id"`
	
	// Provider identification
	ProviderID        string                 `json:"provider_id"`
	ProviderCustomerID string                `json:"provider_customer_id"`
	
	// Customer details
	Name              string                 `json:"name"`
	Email             string                 `json:"email"`
	Phone             string                 `json:"phone"`
	
	// Business information
	BusinessName      string                 `json:"business_name"`
	BusinessCategory  string                 `json:"business_category"`
	
	// Address
	Address           json.RawMessage        `json:"address"`
	
	// Payment history
	TotalInvoiced     float64                `json:"total_invoiced"`
	TotalPaid         float64                `json:"total_paid"`
	OutstandingAmount float64                `json:"outstanding_amount"`
	
	// Risk assessment
	RiskScore         float64                `json:"risk_score"`
	RiskCategory      string                 `json:"risk_category"`
	
	// Provider-specific data
	ProviderMetadata  map[string]interface{} `json:"provider_metadata"`
	
	// Audit fields
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

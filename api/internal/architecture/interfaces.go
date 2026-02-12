package architecture

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentRetryConfig represents configuration for payment retry
type PaymentRetryConfig struct {
	NewAmount    *float64 `json:"new_amount,omitempty"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	RetryReason  string   `json:"retry_reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentProvider defines the core interface for all payment providers
type PaymentProvider interface {
	// Provider identification
	GetProviderID() string
	GetProviderName() string
	GetProviderType() ProviderType

	// Connection management
	Connect(ctx context.Context, config *ProviderConfig) error
	Disconnect(ctx context.Context) error
	IsConnected() bool

	// Data retrieval
	GetPaymentFailures(ctx context.Context, since time.Time) ([]interface{}, error)
	GetInvoices(ctx context.Context, since time.Time) ([]interface{}, error)
	GetCustomers(ctx context.Context, since time.Time) ([]interface{}, error)

	// Health and status
	GetHealthStatus() *HealthStatus
	GetRateLimitInfo() *RateLimitInfo

	// Sync management
	StartSync(ctx context.Context) error
	StopSync(ctx context.Context) error
	GetSyncStatus() *SyncStatus
}

// ProviderType represents the type of payment provider
type ProviderType string

const (
	ProviderTypeWebhook ProviderType = "webhook" // Stripe, PayPal
	ProviderTypeOAuth   ProviderType = "oauth"   // Xero, QuickBooks
	ProviderTypeAPI     ProviderType = "api"     // Bank APIs, CDR
	ProviderTypeManual  ProviderType = "manual"  // CSV uploads, manual entry
)

// EventBus defines the interface for asynchronous event communication
type EventBus interface {
	// Publish events
	Publish(ctx context.Context, topic string, event interface{}) error
	PublishAsync(ctx context.Context, topic string, event interface{}) error

	// Subscribe to events
	Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error)
	SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error)

	// Event management
	Unsubscribe(subscription Subscription) error
	Close() error
}

// EventHandler processes incoming events
type EventHandler func(ctx context.Context, event interface{}) error

// Subscription represents an event subscription
type Subscription interface {
	ID() string
	Topic() string
	Unsubscribe() error
}

// ProviderConfig contains configuration for all providers
type ProviderConfig struct {
	ProviderID   string       `json:"provider_id"`
	ProviderType ProviderType `json:"provider_type"`
	CompanyID    string       `json:"company_id"`

	// OAuth configuration
	OAuthConfig *OAuthConfig `json:"oauth_config,omitempty"`

	// API configuration
	APIConfig *APIConfig `json:"api_config,omitempty"`

	// Webhook configuration
	WebhookConfig *WebhookConfig `json:"webhook_config,omitempty"`

	// Sync configuration
	SyncConfig *SyncConfig `json:"sync_config"`

	// Rate limiting
	RateLimitConfig *RateLimitConfig `json:"rate_limit_config"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OAuth configuration
type OAuthConfig struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	RedirectURI  string    `json:"redirect_uri"`
	Scopes       []string  `json:"scopes"`
	AuthURL      string    `json:"auth_url"`
	TokenURL     string    `json:"token_url"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// API configuration
type APIConfig struct {
	BaseURL     string            `json:"base_url"`
	APIKey      string            `json:"api_key"`
	Headers     map[string]string `json:"headers"`
	Timeout     time.Duration     `json:"timeout"`
	RetryConfig *PaymentRetryConfig      `json:"retry_config"`
}

// Webhook configuration
type WebhookConfig struct {
	Endpoint   string            `json:"endpoint"`
	Secret     string            `json:"secret"`
	Events     []string          `json:"events"`
	Headers    map[string]string `json:"headers"`
	Validation *ValidationConfig `json:"validation"`
}

// Sync configuration
type SyncConfig struct {
	Frequency     time.Duration `json:"frequency"`      // How often to sync
	BatchSize     int           `json:"batch_size"`     // Records per batch
	Incremental   bool          `json:"incremental"`    // Use incremental sync
	RetryAttempts int           `json:"retry_attempts"` // Max retry attempts
	RetryDelay    time.Duration `json:"retry_delay"`    // Delay between retries
	Enabled       bool          `json:"enabled"`        // Whether sync is enabled
}

// Rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	RetryAfter        time.Duration `json:"retry_after"`
}

// Retry configuration
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Backoff     time.Duration `json:"backoff"`
	MaxBackoff  time.Duration `json:"max_backoff"`
}

// Validation configuration
type ValidationConfig struct {
	ValidateSignature bool          `json:"validate_signature"`
	ValidateTimestamp bool          `json:"validate_timestamp"`
	MaxAge            time.Duration `json:"max_age"`
}

// HealthStatus represents provider health information
type HealthStatus struct {
	ProviderID   string                 `json:"provider_id"`
	Status       string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	LastCheck    time.Time              `json:"last_check"`
	ResponseTime time.Duration          `json:"response_time"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// RateLimitInfo contains rate limiting information
type RateLimitInfo struct {
	ProviderID        string    `json:"provider_id"`
	RequestsRemaining int       `json:"requests_remaining"`
	ResetTime         time.Time `json:"reset_time"`
	Limit             int       `json:"limit"`
}

// SyncStatus represents synchronization status
type SyncStatus struct {
	ProviderID    string        `json:"provider_id"`
	Status        string        `json:"status"` // "active", "paused", "error"
	LastSyncAt    *time.Time    `json:"last_sync_at"`
	NextSyncAt    *time.Time    `json:"next_sync_at"`
	LastError     string        `json:"last_error,omitempty"`
	RecordsSynced int64         `json:"records_synced"`
	SyncDuration  time.Duration `json:"sync_duration"`
	Progress      float64       `json:"progress"` // 0.0 to 1.0
}

// PaymentFailure represents a unified payment failure across all providers
type PaymentFailure struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"not null;index"`

	// Provider identification
	ProviderID        string `json:"provider_id" gorm:"not null;index"`
	ProviderEventID   string `json:"provider_event_id" gorm:"not null;uniqueIndex"`
	ProviderEventType string `json:"provider_event_type" gorm:"not null"`

	// Payment details
	Amount        float64 `json:"amount" gorm:"not null"`
	Currency      string  `json:"currency" gorm:"default:'AUD'"`
	PaymentMethod string  `json:"payment_method"`

	// Customer information
	CustomerID    string `json:"customer_id" gorm:"index"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email" gorm:"index"`
	CustomerPhone string `json:"customer_phone"`

	// Failure information
	FailureReason  string `json:"failure_reason" gorm:"index"`
	FailureCode    string `json:"failure_code"`
	FailureMessage string `json:"failure_message"`

	// Business context
	InvoiceID        string     `json:"invoice_id"`
	InvoiceNumber    string     `json:"invoice_number"`
	DueDate          *time.Time `json:"due_date"`
	BusinessCategory string     `json:"business_category"`

	// Processing status
	Status    PaymentFailureStatus   `json:"status" gorm:"default:'received'"`
	Priority  PaymentFailurePriority `json:"priority" gorm:"default:'medium'"`
	RiskScore float64                `json:"risk_score" gorm:"default:0"`

	// Timestamps
	OccurredAt  time.Time  `json:"occurred_at" gorm:"not null"`
	DetectedAt  time.Time  `json:"detected_at" gorm:"not null"`
	ProcessedAt *time.Time `json:"processed_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`

	// Integration details
	IntegrationID     *uuid.UUID `json:"integration_id,omitempty"`
	OriginalInvoiceID string     `json:"original_invoice_id,omitempty"`
	OriginalPaymentID string     `json:"original_payment_id,omitempty"`
	SyncSource        string     `json:"sync_source" gorm:"default:'webhook'"`

	// Raw and normalized data
	RawData        json.RawMessage `json:"raw_data" gorm:"type:jsonb"`
	NormalizedData json.RawMessage `json:"normalized_data" gorm:"type:jsonb"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata" gorm:"type:jsonb"`

	// Metadata
	Tags     []string               `json:"tags" gorm:"type:jsonb"`
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:now()"`
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
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"not null;index"`

	// Provider identification
	ProviderID        string `json:"provider_id" gorm:"not null;index"`
	ProviderInvoiceID string `json:"provider_invoice_id" gorm:"not null;uniqueIndex"`

	// Invoice details
	InvoiceNumber string  `json:"invoice_number" gorm:"not null"`
	Amount        float64 `json:"amount" gorm:"not null"`
	Currency      string  `json:"currency" gorm:"default:'AUD'"`
	Status        string  `json:"status" gorm:"index"`

	// Customer information
	CustomerID    string `json:"customer_id" gorm:"index"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`

	// Dates
	IssueDate time.Time  `json:"issue_date"`
	DueDate   time.Time  `json:"due_date"`
	PaidDate  *time.Time `json:"paid_date"`

	// Line items
	LineItems json.RawMessage `json:"line_items" gorm:"type:jsonb"`

	// Provider-specific data
	ProviderMetadata map[string]interface{} `json:"provider_metadata" gorm:"type:jsonb"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:now()"`
}

// Customer represents a unified customer across all providers
type Customer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"not null;index"`

	// Provider identification
	ProviderID         string `json:"provider_id" gorm:"not null;index"`
	ProviderCustomerID string `json:"provider_customer_id" gorm:"not null;uniqueIndex"`

	// Customer details
	Name  string `json:"name" gorm:"not null"`
	Email string `json:"email" gorm:"index"`
	Phone string `json:"phone"`

	// Business information
	BusinessName     string `json:"business_name"`
	BusinessCategory string `json:"business_category"`

	// Address
	Address json.RawMessage `json:"address" gorm:"type:jsonb"`

	// Payment history
	TotalInvoiced     float64 `json:"total_invoiced" gorm:"default:0"`
	TotalPaid         float64 `json:"total_paid" gorm:"default:0"`
	OutstandingAmount float64 `json:"outstanding_amount" gorm:"default:0"`

	// Risk assessment
	RiskScore    float64 `json:"risk_score" gorm:"default:0"`
	RiskCategory string  `json:"risk_category"`

	// Provider-specific data
	ProviderMetadata map[string]interface{} `json:"provider_metadata" gorm:"type:jsonb"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:now()"`
}

// TableName specifies the table name for GORM
func (PaymentFailure) TableName() string {
	return "payment_failures"
}

func (Invoice) TableName() string {
	return "invoices"
}

func (Customer) TableName() string {
	return "customers"
}

// OAuthTokens represents OAuth 2.0 tokens
type OAuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

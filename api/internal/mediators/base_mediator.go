package mediators

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

// ProviderType represents the type of payment provider
type ProviderType string

const (
	ProviderTypeWebhook ProviderType = "webhook" // Stripe, PayPal
	ProviderTypeOAuth   ProviderType = "oauth"   // Xero, QuickBooks
	ProviderTypeAPI     ProviderType = "api"     // Bank APIs, CDR
	ProviderTypeManual  ProviderType = "manual"  // CSV uploads, manual entry
)

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

// BaseMediator provides common functionality for all payment provider mediators
type BaseMediator struct {
	config   *ProviderConfig
	logger   *zap.Logger
	eventBus EventBus

	// Connection state
	isConnected bool
	connectedAt *time.Time
	lastError   error

	// Sync state
	syncStatus   *SyncStatus
	syncMutex    sync.RWMutex
	syncTicker   *time.Ticker
	syncStopChan chan struct{}

	// Rate limiting
	rateLimiter *RateLimiter

	// Health monitoring
	healthStatus    *HealthStatus
	lastHealthCheck time.Time
}

// NewBaseMediator creates a new base mediator
func NewBaseMediator(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *BaseMediator {
	// Set default sync configuration if not provided
	if config.SyncConfig == nil {
		config.SyncConfig = &SyncConfig{
			Frequency:     15 * time.Minute,
			BatchSize:     100,
			Incremental:   true,
			RetryAttempts: 3,
			RetryDelay:    5 * time.Second,
			Enabled:       true,
		}
	}

	// Set default rate limiting if not provided
	if config.RateLimitConfig == nil {
		config.RateLimitConfig = &RateLimitConfig{
			RequestsPerMinute: 100,
			BurstSize:         10,
			RetryAfter:        1 * time.Minute,
		}
	}

	mediator := &BaseMediator{
		config:   config,
		logger:   logger,
		eventBus: eventBus,
		syncStatus: &SyncStatus{
			ProviderID:    config.ProviderID,
			Status:        "inactive",
			RecordsSynced: 0,
			Progress:      0.0,
		},
		rateLimiter: NewRateLimiter(config.RateLimitConfig),
		healthStatus: &HealthStatus{
			ProviderID: config.ProviderID,
			Status:     "unknown",
			LastCheck:  time.Now(),
		},
	}

	return mediator
}

// GetProviderID returns the provider identifier
func (b *BaseMediator) GetProviderID() string {
	return b.config.ProviderID
}

// GetProviderName returns the provider name
func (b *BaseMediator) GetProviderName() string {
	return b.config.ProviderID // Override in specific mediators
}

// GetProviderType returns the provider type
func (b *BaseMediator) GetProviderType() ProviderType {
	return b.config.ProviderType
}

// IsConnected returns the connection status
func (b *BaseMediator) IsConnected() bool {
	return b.isConnected
}

// GetHealthStatus returns the current health status
func (b *BaseMediator) GetHealthStatus() *HealthStatus {
	b.syncMutex.RLock()
	defer b.syncMutex.RUnlock()

	// Update health status if it's been more than 5 minutes
	if time.Since(b.lastHealthCheck) > 5*time.Minute {
		b.updateHealthStatus()
	}

	return b.healthStatus
}

// GetRateLimitInfo returns rate limiting information
func (b *BaseMediator) GetRateLimitInfo() map[string]interface{} {
	return b.rateLimiter.GetRateLimitInfo()
}

// GetSyncStatus returns the current sync status
func (b *BaseMediator) GetSyncStatus() *SyncStatus {
	b.syncMutex.RLock()
	defer b.syncMutex.RUnlock()
	return b.syncStatus
}

// StartSync starts the synchronization process
func (b *BaseMediator) StartSync(ctx context.Context) error {
	b.syncMutex.Lock()
	defer b.syncMutex.Unlock()

	if b.syncStatus.Status == "active" {
		return fmt.Errorf("sync already active for provider %s", b.config.ProviderID)
	}

	// Update sync status
	b.syncStatus.Status = "active"
	b.syncStatus.LastSyncAt = nil
	b.syncStatus.NextSyncAt = &time.Time{}
	*b.syncStatus.NextSyncAt = time.Now().Add(b.config.SyncConfig.Frequency)

	// Start sync ticker
	b.syncTicker = time.NewTicker(b.config.SyncConfig.Frequency)
	b.syncStopChan = make(chan struct{})

	// Start sync goroutine
	go b.runSyncLoop(ctx)

	b.logger.Info("Started sync for provider",
		zap.String("provider_id", b.config.ProviderID),
		zap.Duration("frequency", b.config.SyncConfig.Frequency))

	return nil
}

// StopSync stops the synchronization process
func (b *BaseMediator) StopSync(ctx context.Context) error {
	b.syncMutex.Lock()
	defer b.syncMutex.Unlock()

	if b.syncStatus.Status != "active" {
		return fmt.Errorf("sync not active for provider %s", b.config.ProviderID)
	}

	// Stop sync ticker
	if b.syncTicker != nil {
		b.syncTicker.Stop()
	}

	// Signal stop
	if b.syncStopChan != nil {
		close(b.syncStopChan)
	}

	// Update sync status
	b.syncStatus.Status = "paused"
	b.syncStatus.NextSyncAt = nil

	b.logger.Info("Stopped sync for provider",
		zap.String("provider_id", b.config.ProviderID))

	return nil
}

// runSyncLoop runs the main sync loop
func (b *BaseMediator) runSyncLoop(ctx context.Context) {
	// Initial sync
	if err := b.performSync(ctx); err != nil {
		b.logger.Error("Initial sync failed",
			zap.String("provider_id", b.config.ProviderID),
			zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Sync loop stopped due to context cancellation",
				zap.String("provider_id", b.config.ProviderID))
			return

		case <-b.syncStopChan:
			b.logger.Info("Sync loop stopped",
				zap.String("provider_id", b.config.ProviderID))
			return

		case <-b.syncTicker.C:
			if err := b.performSync(ctx); err != nil {
				b.logger.Error("Periodic sync failed",
					zap.String("provider_id", b.config.ProviderID),
					zap.Error(err))
			}
		}
	}
}

// performSync performs a single synchronization
func (b *BaseMediator) performSync(ctx context.Context) error {
	startTime := time.Now()

	b.syncMutex.Lock()
	b.syncStatus.Status = "syncing"
	b.syncStatus.LastSyncAt = &startTime
	b.syncMutex.Unlock()

	// This method should be overridden by specific mediators
	// Base implementation just logs and updates status
	b.logger.Debug("Performing sync",
		zap.String("provider_id", b.config.ProviderID))

	// Update sync status
	b.syncMutex.Lock()
	b.syncStatus.Status = "active"
	b.syncStatus.SyncDuration = time.Since(startTime)
	b.syncStatus.NextSyncAt = &time.Time{}
	*b.syncStatus.NextSyncAt = time.Now().Add(b.config.SyncConfig.Frequency)
	b.syncMutex.Unlock()

	return nil
}

// updateHealthStatus updates the health status
func (b *BaseMediator) updateHealthStatus() {
	startTime := time.Now()

	// Perform health check (override in specific mediators)
	status := "healthy"
	var err error

	// Update health status
	b.healthStatus.Status = status
	b.healthStatus.LastCheck = time.Now()
	b.healthStatus.ResponseTime = time.Since(startTime)
	if err != nil {
		b.healthStatus.Error = err.Error()
	} else {
		b.healthStatus.Error = ""
	}

	b.lastHealthCheck = time.Now()
}

// publishEvent publishes an event to the event bus
func (b *BaseMediator) publishEvent(ctx context.Context, topic string, event interface{}) error {
	if b.eventBus == nil {
		return fmt.Errorf("event bus not configured")
	}

	return b.eventBus.Publish(ctx, topic, event)
}

// publishEventAsync publishes an event asynchronously
func (b *BaseMediator) publishEventAsync(ctx context.Context, topic string, event interface{}) error {
	if b.eventBus == nil {
		return fmt.Errorf("event bus not configured")
	}

	return b.eventBus.PublishAsync(ctx, topic, event)
}

// waitForRateLimit waits for rate limiting if necessary
func (b *BaseMediator) waitForRateLimit() {
	b.rateLimiter.Wait()
}

// generateEventID generates a unique event ID
func (b *BaseMediator) generateEventID() string {
	return uuid.New().String()
}

// logProviderAction logs provider-specific actions
func (b *BaseMediator) logProviderAction(action string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("provider_id", b.config.ProviderID),
		zap.String("action", action),
	}, fields...)

	b.logger.Info("Provider action", allFields...)
}

// logProviderError logs provider-specific errors
func (b *BaseMediator) logProviderError(action string, err error, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("provider_id", b.config.ProviderID),
		zap.String("action", action),
		zap.Error(err),
	}, fields...)

	b.logger.Error("Provider error", allFields...)
}

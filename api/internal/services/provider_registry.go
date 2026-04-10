package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ProviderStatus represents the availability state of an external provider
type ProviderStatus struct {
	Available   bool      `json:"available"`
	LastChecked time.Time `json:"last_checked"`
	Endpoint    string    `json:"endpoint,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// ProviderRegistry tracks the availability of external payment providers (Stripe, Xero, PayTo, etc.)
// and allows callers to dynamically decide whether to make a live call or record intent.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]*ProviderStatus
	logger    *zap.Logger
	cacheTTL  time.Duration
}

// NewProviderRegistry creates a new provider registry.
// It auto-detects configured providers from environment variables on startup.
func NewProviderRegistry(logger *zap.Logger) *ProviderRegistry {
	r := &ProviderRegistry{
		providers: make(map[string]*ProviderStatus),
		logger:    logger,
		cacheTTL:  5 * time.Minute,
	}

	// Auto-detect provider availability from environment variables
	r.detectProviders()

	return r
}

// detectProviders checks environment variables to determine which providers are configured.
func (r *ProviderRegistry) detectProviders() {
	providerEnvMap := map[string]string{
		"stripe": "STRIPE_SECRET_KEY",
		"xero":   "XERO_CLIENT_ID",
		"payto":  "PAYTO_API_KEY",
		"becs":   "BECS_API_KEY",
	}

	for provider, envVar := range providerEnvMap {
		value := os.Getenv(envVar)
		available := value != ""

		r.providers[provider] = &ProviderStatus{
			Available:   available,
			LastChecked: time.Now(),
			Endpoint:    fmt.Sprintf("env:%s", envVar),
		}

		if available {
			r.logger.Info("External provider detected as available",
				zap.String("provider", provider),
				zap.String("env_var", envVar),
			)
		} else {
			r.logger.Info("External provider not configured — will record intent only",
				zap.String("provider", provider),
				zap.String("env_var", envVar),
			)
		}
	}
}

// IsAvailable checks if a specific provider endpoint is available.
// Returns true if the provider's API key/config is present and the provider was reachable
// at last check. If the cached status is stale (beyond cacheTTL), it re-checks.
func (r *ProviderRegistry) IsAvailable(provider string) bool {
	r.mu.RLock()
	status, exists := r.providers[strings.ToLower(provider)]
	r.mu.RUnlock()

	if !exists {
		return false
	}

	// If the cache is stale, refresh
	if time.Since(status.LastChecked) > r.cacheTTL {
		r.refreshProvider(provider)
		r.mu.RLock()
		status = r.providers[strings.ToLower(provider)]
		r.mu.RUnlock()
	}

	return status.Available
}

// refreshProvider re-checks a provider's environment variable configuration
func (r *ProviderRegistry) refreshProvider(provider string) {
	providerEnvMap := map[string]string{
		"stripe": "STRIPE_SECRET_KEY",
		"xero":   "XERO_CLIENT_ID",
		"payto":  "PAYTO_API_KEY",
		"becs":   "BECS_API_KEY",
	}

	envVar, ok := providerEnvMap[strings.ToLower(provider)]
	if !ok {
		return
	}

	value := os.Getenv(envVar)
	available := value != ""

	r.mu.Lock()
	r.providers[strings.ToLower(provider)] = &ProviderStatus{
		Available:   available,
		LastChecked: time.Now(),
		Endpoint:    fmt.Sprintf("env:%s", envVar),
	}
	r.mu.Unlock()
}

// RegisterProvider manually registers or updates a provider's availability.
// Useful for runtime health checks or when a provider comes online dynamically.
func (r *ProviderRegistry) RegisterProvider(provider string, available bool, endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[strings.ToLower(provider)] = &ProviderStatus{
		Available:   available,
		LastChecked: time.Now(),
		Endpoint:    endpoint,
	}

	r.logger.Info("Provider registration updated",
		zap.String("provider", provider),
		zap.Bool("available", available),
		zap.String("endpoint", endpoint),
	)
}

// MarkUnavailable marks a provider as unavailable (e.g., after a failed API call)
func (r *ProviderRegistry) MarkUnavailable(provider string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if status, exists := r.providers[strings.ToLower(provider)]; exists {
		status.Available = false
		status.LastChecked = time.Now()
		status.Error = err.Error()
	}

	r.logger.Warn("Provider marked as unavailable after failure",
		zap.String("provider", provider),
		zap.Error(err),
	)
}

// GetAllStatuses returns the current status of all registered providers
func (r *ProviderRegistry) GetAllStatuses() map[string]*ProviderStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ProviderStatus, len(r.providers))
	for k, v := range r.providers {
		copied := *v
		result[k] = &copied
	}
	return result
}

// ExecuteOrRecordIntent is the core dynamic dispatch method.
// If the provider is available, it calls executeFn. If not, it calls recordIntentFn.
// This eliminates the need for callers to manually check availability.
func (r *ProviderRegistry) ExecuteOrRecordIntent(
	ctx context.Context,
	provider string,
	executeFn func(ctx context.Context) error,
	recordIntentFn func(ctx context.Context) error,
) (executed bool, err error) {
	if r.IsAvailable(provider) {
		r.logger.Info("Provider available — executing live call",
			zap.String("provider", provider),
		)
		err := executeFn(ctx)
		if err != nil {
			// Mark provider unavailable on failure and fall back to intent recording
			r.MarkUnavailable(provider, err)
			r.logger.Warn("Live call failed — falling back to intent recording",
				zap.String("provider", provider),
				zap.Error(err),
			)
			return false, recordIntentFn(ctx)
		}
		return true, nil
	}

	r.logger.Info("Provider not available — recording intent",
		zap.String("provider", provider),
	)
	return false, recordIntentFn(ctx)
}

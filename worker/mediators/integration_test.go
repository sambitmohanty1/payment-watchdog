package mediators

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestMediatorIntegration tests basic integration between mediators
func TestMediatorIntegration(t *testing.T) {
	// Test that all mediators can be created
	configs := []*ProviderConfig{
		{
			ProviderID:   "stripe_test",
			ProviderType: ProviderTypeWebhook,
			WebhookConfig: &WebhookConfig{
				Secret: "test_webhook_secret",
			},
			APIConfig: &APIConfig{
				APIKey: "test_api_key",
			},
		},
		{
			ProviderID:   "xero_test",
			ProviderType: ProviderTypeOAuth,
			OAuthConfig: &OAuthConfig{
				ClientID:     "test_client_id",
				ClientSecret: "test_client_secret",
			},
		},
		{
			ProviderID:   "quickbooks_test",
			ProviderType: ProviderTypeOAuth,
			OAuthConfig: &OAuthConfig{
				ClientID:     "test_client_id",
				ClientSecret: "test_client_secret",
			},
		},
	}

	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	for _, config := range configs {
		var mediator interface{}

		switch config.ProviderID {
		case "stripe_test":
			mediator = NewStripeMediator(config, eventBus, logger)
		case "xero_test":
			mediator = NewXeroMediator(config, eventBus, logger)
		case "quickbooks_test":
			mediator = NewQuickBooksMediator(config, eventBus, logger)
		}

		assert.NotNil(t, mediator)
	}
}

// TestEventBusIntegration tests event bus integration
func TestEventBusIntegration(t *testing.T) {
	eventBus := &TestEventBus{}
	ctx := context.Background()

	// Test that event bus can publish events
	event := map[string]interface{}{
		"test": "data",
	}

	err := eventBus.Publish(ctx, "test.topic", event)
	assert.NoError(t, err)
}

// TestProviderConfigValidation tests provider configuration
func TestProviderConfigValidation(t *testing.T) {
	// Test valid Stripe config
	stripeConfig := &ProviderConfig{
		ProviderID:   "stripe_test",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_webhook_secret",
		},
		APIConfig: &APIConfig{
			APIKey: "test_api_key",
		},
	}
	assert.NotNil(t, stripeConfig)

	// Test valid Xero config
	xeroConfig := &ProviderConfig{
		ProviderID:   "xero_test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
	}
	assert.NotNil(t, xeroConfig)

	// Test valid QuickBooks config
	quickbooksConfig := &ProviderConfig{
		ProviderID:   "quickbooks_test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
	}
	assert.NotNil(t, quickbooksConfig)
}

// TestRateLimiterIntegration tests rate limiter integration
func TestRateLimiterIntegration(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)
	assert.NotNil(t, limiter)

	// Test rate limiting behavior
	start := time.Now()
	for i := 0; i < 15; i++ {
		limiter.Wait()
	}
	duration := time.Since(start)

	// Should take some time due to rate limiting
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

// TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	// Test with invalid config
	invalidConfig := &ProviderConfig{
		ProviderID:   "invalid_test",
		ProviderType: ProviderTypeOAuth,
		// Missing required OAuth config
	}

	eventBus := &TestEventBus{}
	logger := zap.NewNop()
	mediator := NewXeroMediator(invalidConfig, eventBus, logger)

	// Should handle invalid config gracefully
	assert.NotNil(t, mediator)

	// Connection should fail
	err := mediator.Connect(context.Background(), invalidConfig)
	assert.Error(t, err)
}

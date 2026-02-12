package mediators

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// BenchmarkStripeMediatorCreation benchmarks Stripe mediator creation
func BenchmarkStripeMediatorCreation(b *testing.B) {
	config := &ProviderConfig{
		ProviderID:   "stripe_test",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_webhook_secret",
		},
		APIConfig: &APIConfig{
			APIKey: "test_api_key",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStripeMediator(config, eventBus, logger)
	}
}

// BenchmarkXeroMediatorCreation benchmarks Xero mediator creation
func BenchmarkXeroMediatorCreation(b *testing.B) {
	config := &ProviderConfig{
		ProviderID:   "xero_test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewXeroMediator(config, eventBus, logger)
	}
}

// BenchmarkQuickBooksMediatorCreation benchmarks QuickBooks mediator creation
func BenchmarkQuickBooksMediatorCreation(b *testing.B) {
	config := &ProviderConfig{
		ProviderID:   "quickbooks_test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQuickBooksMediator(config, eventBus, logger)
	}
}

// BenchmarkRateLimiterWait benchmarks rate limiter wait operation
func BenchmarkRateLimiterWait(b *testing.B) {
	config := &RateLimitConfig{
		RequestsPerMinute: 1000,
		BurstSize:         100,
		RetryAfter:        1 * time.Millisecond,
	}

	limiter := NewRateLimiter(config)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Wait()
	}
}

// BenchmarkRateLimiterTryAcquire benchmarks rate limiter try acquire operation
func BenchmarkRateLimiterTryAcquire(b *testing.B) {
	config := &RateLimitConfig{
		RequestsPerMinute: 1000,
		BurstSize:         100,
		RetryAfter:        1 * time.Millisecond,
	}

	limiter := NewRateLimiter(config)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.TryAcquire()
	}
}

// BenchmarkEventBusPublish benchmarks event bus publish operation
func BenchmarkEventBusPublish(b *testing.B) {
	eventBus := &TestEventBus{}
	ctx := context.Background()
	event := map[string]interface{}{
		"test": "data",
		"id":   "123",
		"time": time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventBus.Publish(ctx, "test.topic", event)
	}
}

// BenchmarkProviderConfigValidation benchmarks provider config validation
func BenchmarkProviderConfigValidation(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, config := range configs {
			_ = config.ProviderID
			_ = config.ProviderType
		}
	}
}

// BenchmarkMediatorProviderInfo benchmarks provider info method calls
func BenchmarkMediatorProviderInfo(b *testing.B) {
	config := &ProviderConfig{
		ProviderID:   "stripe_test",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_webhook_secret",
		},
		APIConfig: &APIConfig{
			APIKey: "test_api_key",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()
	mediator := NewStripeMediator(config, eventBus, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mediator.GetProviderID()
		_ = mediator.GetProviderName()
		_ = mediator.GetProviderType()
	}
}

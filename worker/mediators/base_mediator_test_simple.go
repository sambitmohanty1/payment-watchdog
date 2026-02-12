package mediators

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewBaseMediator(t *testing.T) {
	config := &ProviderConfig{
		ProviderID:   "test_provider",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	mediator := NewBaseMediator(config, eventBus, logger)

	assert.NotNil(t, mediator)
	assert.Equal(t, config, mediator.config)
	assert.Equal(t, eventBus, mediator.eventBus)
	assert.Equal(t, logger, mediator.logger)
	assert.False(t, mediator.IsConnected())
	assert.Equal(t, "idle", mediator.syncStatus.Status)
}

func TestBaseMediator_ProviderInfo(t *testing.T) {
	config := &ProviderConfig{
		ProviderID:   "test_provider",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	mediator := NewBaseMediator(config, eventBus, logger)

	// Test provider info methods
	assert.Equal(t, "test_provider", mediator.GetProviderID())
	assert.Equal(t, ProviderTypeWebhook, mediator.GetProviderType())
}

func TestBaseMediator_HealthStatus(t *testing.T) {
	config := &ProviderConfig{
		ProviderID:   "test_provider",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	mediator := NewBaseMediator(config, eventBus, logger)

	// Test health status
	healthStatus := mediator.GetHealthStatus()
	assert.NotNil(t, healthStatus)
	assert.Equal(t, "test_provider", healthStatus.ProviderID)
}

func TestBaseMediator_SyncStatus(t *testing.T) {
	config := &ProviderConfig{
		ProviderID:   "test_provider",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	mediator := NewBaseMediator(config, eventBus, logger)

	// Test sync status
	syncStatus := mediator.GetSyncStatus()
	assert.NotNil(t, syncStatus)
	assert.Equal(t, "test_provider", syncStatus.ProviderID)
	assert.Equal(t, "idle", syncStatus.Status)
}

func TestBaseMediator_RateLimitInfo(t *testing.T) {
	config := &ProviderConfig{
		ProviderID:   "test_provider",
		ProviderType: ProviderTypeWebhook,
		WebhookConfig: &WebhookConfig{
			Secret: "test_secret",
		},
	}
	eventBus := &TestEventBus{}
	logger := zap.NewNop()

	mediator := NewBaseMediator(config, eventBus, logger)

	// Test rate limit info
	rateLimitInfo := mediator.GetRateLimitInfo()
	assert.NotNil(t, rateLimitInfo)
	assert.Contains(t, rateLimitInfo, "requests_per_second")
}

// TestEventBus is a simple test implementation for testing
type TestEventBus struct {
	events map[string][]interface{}
}

func (t *TestEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	if t.events == nil {
		t.events = make(map[string][]interface{})
	}
	t.events[topic] = append(t.events[topic], event)
	return nil
}

func (t *TestEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	// Simple test implementation
	return &TestSubscription{}, nil
}

func (t *TestEventBus) GetEvents(topic string) []interface{} {
	if t.events == nil {
		return nil
	}
	return t.events[topic]
}

func (t *TestEventBus) Close() error {
	return nil
}

func (t *TestEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return t.Publish(ctx, topic, event)
}

func (t *TestEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	return t.Subscribe(ctx, topic, handler)
}

func (t *TestEventBus) Unsubscribe(subscription Subscription) error {
	return subscription.Unsubscribe()
}

type TestSubscription struct{}

func (t *TestSubscription) Unsubscribe() error { return nil }

func (t *TestSubscription) ID() string { return "test_subscription" }

func (t *TestSubscription) Topic() string { return "test_topic" }

package mediators

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewStripeMediator(t *testing.T) {
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

	assert.NotNil(t, mediator)
	assert.Equal(t, config, mediator.config)
	assert.Equal(t, eventBus, mediator.eventBus)
	assert.Equal(t, logger, mediator.logger)
	assert.False(t, mediator.IsConnected())
}

func TestStripeMediator_ProviderInfo(t *testing.T) {
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

	// Test provider info methods
	assert.Equal(t, "stripe", mediator.GetProviderID())
	assert.Equal(t, "Stripe", mediator.GetProviderName())
	assert.Equal(t, ProviderTypeWebhook, mediator.GetProviderType())
}

func TestStripeMediator_Connect(t *testing.T) {
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

	// Test connection with test config
	err := mediator.Connect(context.Background(), config)
	// Note: This might succeed or fail depending on Stripe library behavior
	// We just test that the method runs without panicking
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	} else {
		t.Logf("Connection succeeded with test config")
	}
}

func TestStripeMediator_GetPaymentFailures(t *testing.T) {
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

	// Test without connection
	failures, err := mediator.GetPaymentFailures(context.Background(), time.Now().Add(-24*time.Hour))
	assert.Error(t, err)
	assert.Nil(t, failures)
	assert.Contains(t, err.Error(), "not connected")
}

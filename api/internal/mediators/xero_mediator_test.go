package mediators

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewXeroMediator(t *testing.T) {
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

	mediator := NewXeroMediator(config, eventBus, logger)

	assert.NotNil(t, mediator)
	assert.Equal(t, config, mediator.config)
	assert.Equal(t, eventBus, mediator.eventBus)
	assert.Equal(t, logger, mediator.logger)
	assert.False(t, mediator.IsConnected())
}

func TestXeroMediator_Connect(t *testing.T) {
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

	mediator := NewXeroMediator(config, eventBus, logger)

	// Test connection with invalid config (missing required fields)
	err := mediator.Connect(context.Background(), config)
	assert.Error(t, err) // Should fail due to missing OAuth tokens
}

func TestXeroMediator_GetPaymentFailures(t *testing.T) {
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

	mediator := NewXeroMediator(config, eventBus, logger)

	// Test without connection
	failures, err := mediator.GetPaymentFailures(context.Background(), time.Now().Add(-24*time.Hour))
	assert.Error(t, err)
	assert.Nil(t, failures)
	assert.Contains(t, err.Error(), "not connected")
}

func TestXeroMediator_ProviderInfo(t *testing.T) {
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

	mediator := NewXeroMediator(config, eventBus, logger)

	// Test provider info methods
	assert.Equal(t, "xero_test", mediator.GetProviderID())
	assert.Equal(t, "Xero", mediator.GetProviderName())
	assert.Equal(t, ProviderTypeOAuth, mediator.GetProviderType())
}

func TestXeroMediator_ConnectionState(t *testing.T) {
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

	mediator := NewXeroMediator(config, eventBus, logger)

	// Test initial connection state
	assert.False(t, mediator.IsConnected())

	// Test provider identification
	assert.Equal(t, "xero_test", mediator.GetProviderID())
	assert.Equal(t, "Xero", mediator.GetProviderName())
	assert.Equal(t, ProviderTypeOAuth, mediator.GetProviderType())
}

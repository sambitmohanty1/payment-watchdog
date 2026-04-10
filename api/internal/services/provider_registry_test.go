package services

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProviderRegistry(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Detection from Environment", func(t *testing.T) {
		// Set a test env var
		os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
		defer os.Unsetenv("STRIPE_SECRET_KEY")
		
		os.Unsetenv("XERO_CLIENT_ID")

		registry := NewProviderRegistry(logger)

		assert.True(t, registry.IsAvailable("stripe"))
		assert.False(t, registry.IsAvailable("xero"))
		assert.False(t, registry.IsAvailable("non_existent"))
	})

	t.Run("Manual Registration", func(t *testing.T) {
		registry := NewProviderRegistry(logger)
		registry.RegisterProvider("custom", true, "https://api.custom.com")

		assert.True(t, registry.IsAvailable("custom"))
		status := registry.GetAllStatuses()["custom"]
		assert.Equal(t, "https://api.custom.com", status.Endpoint)
	})

	t.Run("Mark Unavailable on Failure", func(t *testing.T) {
		registry := NewProviderRegistry(logger)
		registry.RegisterProvider("stripe", true, "env:STRIPE_SECRET_KEY")

		assert.True(t, registry.IsAvailable("stripe"))
		
		registry.MarkUnavailable("stripe", errors.New("connection timeout"))
		
		assert.False(t, registry.IsAvailable("stripe"))
		status := registry.GetAllStatuses()["stripe"]
		assert.Equal(t, "connection timeout", status.Error)
	})

	t.Run("ExecuteOrRecordIntent - Available", func(t *testing.T) {
		registry := NewProviderRegistry(logger)
		registry.RegisterProvider("stripe", true, "mock")

		executedCalled := false
		fallbackCalled := false

		executed, err := registry.ExecuteOrRecordIntent(
			context.Background(),
			"stripe",
			func(ctx context.Context) error {
				executedCalled = true
				return nil
			},
			func(ctx context.Context) error {
				fallbackCalled = true
				return nil
			},
		)

		assert.NoError(t, err)
		assert.True(t, executed)
		assert.True(t, executedCalled)
		assert.False(t, fallbackCalled)
	})

	t.Run("ExecuteOrRecordIntent - Unavailable", func(t *testing.T) {
		registry := NewProviderRegistry(logger)
		registry.RegisterProvider("stripe", false, "mock")

		executedCalled := false
		fallbackCalled := false

		executed, err := registry.ExecuteOrRecordIntent(
			context.Background(),
			"stripe",
			func(ctx context.Context) error {
				executedCalled = true
				return nil
			},
			func(ctx context.Context) error {
				fallbackCalled = true
				return nil
			},
		)

		assert.NoError(t, err)
		assert.False(t, executed)
		assert.False(t, executedCalled)
		assert.True(t, fallbackCalled)
	})

	t.Run("ExecuteOrRecordIntent - Live Failure Fallback", func(t *testing.T) {
		registry := NewProviderRegistry(logger)
		registry.RegisterProvider("stripe", true, "mock")

		executedCalled := false
		fallbackCalled := false

		executed, err := registry.ExecuteOrRecordIntent(
			context.Background(),
			"stripe",
			func(ctx context.Context) error {
				executedCalled = true
				return errors.New("api error")
			},
			func(ctx context.Context) error {
				fallbackCalled = true
				return nil
			},
		)

		// Should return success from fallback, but executed = false
		assert.NoError(t, err)
		assert.False(t, executed)
		assert.True(t, executedCalled)
		assert.True(t, fallbackCalled)
		assert.False(t, registry.IsAvailable("stripe"))
	})
}

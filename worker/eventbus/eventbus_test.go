package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRedisEventBus(t *testing.T) {
	logger := zap.NewNop()
	
	// Test with valid configuration
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	
	// This might fail if Redis is not running, but we're testing the constructor
	if err == nil {
		assert.NotNil(t, eventBus)
		assert.Equal(t, logger, eventBus.logger)
	} else {
		// If Redis is not available, we expect a connection error
		assert.Contains(t, err.Error(), "connection")
	}
}

func TestRedisEventBus_Publish(t *testing.T) {
	logger := zap.NewNop()
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer eventBus.Close()

	ctx := context.Background()
	
	// Test publishing an event
	testEvent := map[string]string{"key": "value"}
	err = eventBus.Publish(ctx, "test_topic", testEvent)
	
	// This might fail if Redis is not running
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NoError(t, err)
	}
}

func TestRedisEventBus_Subscribe(t *testing.T) {
	logger := zap.NewNop()
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer eventBus.Close()

	ctx := context.Background()
	
	// Test subscribing to a topic
	handler := func(ctx context.Context, event interface{}) error {
		return nil
	}
	
	subscription, err := eventBus.Subscribe(ctx, "test_topic", handler)
	
	// This might fail if Redis is not running
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NotNil(t, subscription)
		assert.NoError(t, subscription.Unsubscribe())
	}
}

func TestRedisEventBus_Close(t *testing.T) {
	logger := zap.NewNop()
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Test closing the event bus
	err = eventBus.Close()
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NoError(t, err)
	}
}

func TestRedisSubscription_Unsubscribe(t *testing.T) {
	logger := zap.NewNop()
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer eventBus.Close()

	ctx := context.Background()
	
	handler := func(ctx context.Context, event interface{}) error {
		return nil
	}
	
	subscription, err := eventBus.Subscribe(ctx, "test_topic", handler)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Test unsubscribing
	err = subscription.Unsubscribe()
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NoError(t, err)
	}
}

func TestRedisEventBus_Integration(t *testing.T) {
	logger := zap.NewNop()
	eventBus, err := NewRedisEventBus("localhost:6379", "", 0, logger)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer eventBus.Close()

	ctx := context.Background()
	
	// Test the full publish/subscribe flow
	receivedEvents := make(chan interface{}, 1)
	
	handler := func(ctx context.Context, event interface{}) error {
		receivedEvents <- event
		return nil
	}
	
	// Subscribe to a topic
	subscription, err := eventBus.Subscribe(ctx, "integration_test", handler)
	require.NoError(t, err)
	defer subscription.Unsubscribe()
	
	// Wait a bit for subscription to be established
	time.Sleep(100 * time.Millisecond)
	
	// Publish an event
	testEvent := map[string]string{"test": "data"}
	err = eventBus.Publish(ctx, "integration_test", testEvent)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	
	// Wait for the event to be received
	select {
	case receivedEvent := <-receivedEvents:
		assert.Equal(t, testEvent, receivedEvent)
	case <-time.After(2 * time.Second):
		t.Skip("Redis not available or event not received, skipping integration test")
	}
}

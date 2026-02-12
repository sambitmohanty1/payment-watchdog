package utils

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestEventBus provides a test implementation of EventBus
type TestEventBus struct {
	events map[string][]interface{}
	mutex  sync.RWMutex
}

// NewTestEventBus creates a new test event bus
func NewTestEventBus() *TestEventBus {
	return &TestEventBus{
		events: make(map[string][]interface{}),
	}
}

// Publish event for testing
func (t *TestEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.events[topic] = append(t.events[topic], event)
	return nil
}

// PublishAsync event for testing
func (t *TestEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return t.Publish(ctx, topic, event)
}

// Subscribe to events for testing
func (t *TestEventBus) Subscribe(ctx context.Context, topic string, handler func(context.Context, interface{}) error) (interface{}, error) {
	// For testing, we'll create a simple subscription
	subscription := &TestSubscription{
		id:      generateTestID(),
		topic:   topic,
		handler: handler,
		eventBus: t,
	}
	return subscription, nil
}

// SubscribeAsync to events for testing
func (t *TestEventBus) SubscribeAsync(ctx context.Context, topic string, handler func(context.Context, interface{}) error) (interface{}, error) {
	return t.Subscribe(ctx, topic, handler)
}

// Unsubscribe from events for testing
func (t *TestEventBus) Unsubscribe(subscription interface{}) error {
	// For testing, just return success
	return nil
}

// Close the test event bus
func (t *TestEventBus) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.events = make(map[string][]interface{})
	return nil
}

// GetEvents returns events for a topic
func (t *TestEventBus) GetEvents(topic string) []interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.events[topic]
}

// GetEventCount returns the count of events for a topic
func (t *TestEventBus) GetEventCount(topic string) int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return len(t.events[topic])
}

// ClearEvents clears all events for a topic
func (t *TestEventBus) ClearEvents(topic string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.events[topic] = make([]interface{}, 0)
}

// ClearAllEvents clears all events
func (t *TestEventBus) ClearAllEvents() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.events = make(map[string][]interface{})
}

// TestSubscription represents a test subscription
type TestSubscription struct {
	id       string
	topic    string
	handler  func(context.Context, interface{}) error
	eventBus *TestEventBus
}

// ID returns the subscription ID
func (t *TestSubscription) ID() string {
	return t.id
}

// Topic returns the subscription topic
func (t *TestSubscription) Topic() string {
	return t.topic
}

// Unsubscribe from the topic
func (t *TestSubscription) Unsubscribe() error {
	return t.eventBus.Unsubscribe(t)
}

// TestLogger provides a test logger implementation
type TestLogger struct {
	entries []LogEntry
	mutex   sync.RWMutex
}

// LogEntry represents a log entry
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
	Time    time.Time
}

// NewTestLogger creates a new test logger
func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries: make([]LogEntry, 0),
	}
}

// Info logs an info message
func (t *TestLogger) Info(msg string, fields ...zap.Field) {
	t.log("info", msg, fields...)
}

// Error logs an error message
func (t *TestLogger) Error(msg string, fields ...zap.Field) {
	t.log("error", msg, fields...)
}

// Debug logs a debug message
func (t *TestLogger) Debug(msg string, fields ...zap.Field) {
	t.log("debug", msg, fields...)
}

// Warn logs a warning message
func (t *TestLogger) Warn(msg string, fields ...zap.Field) {
	t.log("warn", msg, fields...)
}

// log logs a message with the given level
func (t *TestLogger) log(level, msg string, fields ...zap.Field) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	entry := LogEntry{
		Level:   level,
		Message: msg,
		Fields:  make(map[string]interface{}),
		Time:    time.Now(),
	}

	// Convert zap fields to map
	for _, field := range fields {
		entry.Fields[field.Key] = field.Interface
	}

	t.entries = append(t.entries, entry)
}

// GetEntries returns all log entries
func (t *TestLogger) GetEntries() []LogEntry {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.entries
}

// GetEntriesByLevel returns log entries by level
func (t *TestLogger) GetEntriesByLevel(level string) []LogEntry {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	var filtered []LogEntry
	for _, entry := range t.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// ClearEntries clears all log entries
func (t *TestLogger) ClearEntries() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.entries = make([]LogEntry, 0)
}

// TestConfig provides test configuration
type TestConfig struct {
	ProviderID   string
	CompanyID    string
	APIKey       string
	WebhookSecret string
	EventBus     *TestEventBus
	Logger       *TestLogger
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		ProviderID:    "test-provider",
		CompanyID:     "test-company",
		APIKey:        "test-api-key",
		WebhookSecret: "test-webhook-secret",
		EventBus:      NewTestEventBus(),
		Logger:        NewTestLogger(),
	}
}

// TestProviderConfig creates a provider config for testing
func (t *TestConfig) ProviderConfig() interface{} {
	// This would return the actual ProviderConfig type
	// For now, return a map for testing
	return map[string]interface{}{
		"provider_id": t.ProviderID,
		"company_id":  t.CompanyID,
		"api_config": map[string]interface{}{
			"api_key": t.APIKey,
		},
		"webhook_config": map[string]interface{}{
			"secret": t.WebhookSecret,
		},
	}
}

// TestContext provides a test context
func TestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	// Don't call cancel here as it would cancel the context immediately
	// The caller should handle cancellation
	return ctx
}

// TestContextWithTimeout provides a test context with custom timeout
func TestContextWithTimeout(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	// Don't call cancel here as it would cancel the context immediately
	// The caller should handle cancellation
	return ctx
}

// AssertNoError asserts that there is no error
func AssertNoError(t require.TestingT, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
}

// AssertError asserts that there is an error
func AssertError(t require.TestingT, err error, msgAndArgs ...interface{}) {
	require.Error(t, err, msgAndArgs...)
}

// AssertEqual asserts that two values are equal
func AssertEqual(t require.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	require.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t require.TestingT, object interface{}, msgAndArgs ...interface{}) {
	require.NotNil(t, object, msgAndArgs...)
}

// AssertTrue asserts that a value is true
func AssertTrue(t require.TestingT, value bool, msgAndArgs ...interface{}) {
	require.True(t, value, msgAndArgs...)
}

// AssertFalse asserts that a value is false
func AssertFalse(t require.TestingT, value bool, msgAndArgs ...interface{}) {
	require.False(t, value, msgAndArgs...)
}

// generateTestID generates a test ID
func generateTestID() string {
	return "test-" + time.Now().Format("20060102150405")
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// WaitForEventCount waits for a specific event count on a topic
func WaitForEventCount(eventBus *TestEventBus, topic string, expectedCount int, timeout time.Duration) bool {
	return WaitForCondition(func() bool {
		return eventBus.GetEventCount(topic) >= expectedCount
	}, timeout)
}

// WaitForLogEntry waits for a specific log entry
func WaitForLogEntry(logger *TestLogger, level, message string, timeout time.Duration) bool {
	return WaitForCondition(func() bool {
		entries := logger.GetEntriesByLevel(level)
		for _, entry := range entries {
			if entry.Message == message {
				return true
			}
		}
		return false
	}, timeout)
}

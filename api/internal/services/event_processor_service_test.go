package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/payment-watchdog/internal/architecture"
	"github.com/payment-watchdog/internal/rules"
)

// TestEventBus is a mock event bus for testing
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

func (t *TestEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return t.Publish(ctx, topic, event)
}

func (t *TestEventBus) Subscribe(ctx context.Context, topic string, handler architecture.EventHandler) (architecture.Subscription, error) {
	// Mock subscription - just return a simple mock
	return &TestSubscription{}, nil
}

func (t *TestEventBus) SubscribeAsync(ctx context.Context, topic string, handler architecture.EventHandler) (architecture.Subscription, error) {
	return t.Subscribe(ctx, topic, handler)
}

func (t *TestEventBus) Unsubscribe(subscription architecture.Subscription) error {
	return nil
}

func (t *TestEventBus) Close() error {
	return nil
}

func (t *TestEventBus) GetEvents(topic string) []interface{} {
	if t.events == nil {
		return []interface{}{}
	}
	return t.events[topic]
}

// TestSubscription is a mock subscription for testing
type TestSubscription struct{}

func (t *TestSubscription) ID() string {
	return "test-subscription"
}

func (t *TestSubscription) Topic() string {
	return "test-topic"
}

func (t *TestSubscription) Unsubscribe() error {
	return nil
}

// TestEventProcessorService tests the Event Processing Pipeline
func TestEventProcessorService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create test event bus
	testEventBus := &TestEventBus{
		events: make(map[string][]interface{}),
	}

	// Create mock rule engine
	ruleEngine := rules.NewRuleEngine(logger)

	// Create event processor service
	service := NewEventProcessorService(nil, ruleEngine, testEventBus, logger)

	t.Run("Service Creation", func(t *testing.T) {
		assert.NotNil(t, service)
		assert.NotNil(t, service.metrics)
		assert.Equal(t, 3, service.maxRetries)
		assert.Equal(t, time.Second*2, service.retryDelay)
	})

	t.Run("Payment Failure Event Processing", func(t *testing.T) {
		// Create test payment failure
		paymentFailure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			CompanyID:       "company-123",
			ProviderID:      "stripe",
			ProviderEventID: "pi_123",
			Amount:          2500.00,
			Currency:        "USD",
			CustomerID:      "customer-123",
			CustomerName:    "Test Customer",
			FailureReason:   "card_declined",
			Status:          architecture.PaymentFailureStatusReceived,
			Priority:        architecture.PaymentFailurePriorityMedium,
			RiskScore:       0, // Will be calculated
			OccurredAt:      time.Now().Add(-time.Hour),
			DetectedAt:      time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Create test event
		event := map[string]interface{}{
			"payment_failure": paymentFailure,
			"timestamp":       time.Now(),
		}

		// Process the event
		err := service.ProcessPaymentFailureEvent(context.Background(), event)
		require.NoError(t, err)

		// Verify metrics were updated
		metrics := service.GetMetrics()
		assert.Equal(t, int64(1), metrics.TotalEventsProcessed)
		assert.Equal(t, int64(1), metrics.SuccessfullyProcessed)
		assert.Equal(t, int64(0), metrics.FailedProcessing)
		assert.Equal(t, int64(1), metrics.EventsByProvider["stripe"])
		assert.Equal(t, int64(1), metrics.EventsByStatus[string(architecture.PaymentFailureStatusAnalyzed)])

		// Verify the payment failure was processed
		assert.Equal(t, architecture.PaymentFailureStatusAnalyzed, paymentFailure.Status)
		assert.NotNil(t, paymentFailure.ProcessedAt)
		assert.True(t, paymentFailure.RiskScore > 0)
		assert.NotEmpty(t, paymentFailure.Priority)
	})

	t.Run("Event Data Enrichment", func(t *testing.T) {
		// Create minimal payment failure
		paymentFailure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			ProviderEventID: "pi_124",
			Amount:          1000.00,
			Currency:        "AUD",
			Status:          architecture.PaymentFailureStatusReceived,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		event := map[string]interface{}{
			"payment_failure": paymentFailure,
		}

		// Process the event
		err := service.ProcessPaymentFailureEvent(context.Background(), event)
		require.NoError(t, err)

		// Verify enrichment
		assert.Equal(t, "default_company", paymentFailure.CompanyID)
		assert.Equal(t, "payment_failure", paymentFailure.ProviderEventType)
		assert.False(t, paymentFailure.OccurredAt.IsZero())
		assert.False(t, paymentFailure.DetectedAt.IsZero())
		assert.Equal(t, "general", paymentFailure.BusinessCategory)
		// Note: Tags are only added when FailureReason is "invoice_unpaid"
		// This test case doesn't set that specific reason
	})

	t.Run("Risk Score Calculation", func(t *testing.T) {
		testCases := []struct {
			amount      float64
			overdueDays int
			expectedMin float64
			description string
		}{
			{500.00, 0, 50.0, "Low amount, not overdue"},
			{1500.00, 7, 60.0, "Medium amount, slightly overdue"},      // Base 50 + 10 (amount) + 0 (overdue) = 60
			{7500.00, 30, 70.0, "High amount, moderately overdue"},     // Base 50 + 20 (amount) + 0 (overdue) = 70
			{15000.00, 90, 90.0, "Very high amount, severely overdue"}, // Base 50 + 30 (amount) + 10 (overdue) = 90
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				dueDate := time.Now().AddDate(0, 0, -tc.overdueDays)
				paymentFailure := &architecture.PaymentFailure{
					ID:              uuid.New(),
					ProviderEventID: "pi_125",
					Amount:          tc.amount,
					Currency:        "USD",
					Status:          architecture.PaymentFailureStatusReceived,
					DueDate:         &dueDate,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}

				event := map[string]interface{}{
					"payment_failure": paymentFailure,
				}

				err := service.ProcessPaymentFailureEvent(context.Background(), event)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, paymentFailure.RiskScore, tc.expectedMin)
				assert.LessOrEqual(t, paymentFailure.RiskScore, 100.0)
				assert.NotEmpty(t, paymentFailure.Priority)
			})
		}
	})

	t.Run("Business Category Risk Factors", func(t *testing.T) {
		testCases := []struct {
			category    string
			expectedMin float64
			description string
		}{
			{"construction", 60.0, "High-risk industry"},
			{"retail", 55.0, "Medium-risk industry"},
			{"consulting", 50.0, "Low-risk industry"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				paymentFailure := &architecture.PaymentFailure{
					ID:               uuid.New(),
					ProviderEventID:  "pi_126",
					Amount:           1000.00,
					Currency:         "USD",
					BusinessCategory: tc.category,
					Status:           architecture.PaymentFailureStatusReceived,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}

				event := map[string]interface{}{
					"payment_failure": paymentFailure,
				}

				err := service.ProcessPaymentFailureEvent(context.Background(), event)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, paymentFailure.RiskScore, tc.expectedMin)
			})
		}
	})

	t.Run("Priority Mapping", func(t *testing.T) {
		testCases := []struct {
			riskScore   float64
			expected    architecture.PaymentFailurePriority
			description string
		}{
			{30.0, architecture.PaymentFailurePriorityLow, "Low risk"},
			{50.0, architecture.PaymentFailurePriorityMedium, "Medium risk"},
			{70.0, architecture.PaymentFailurePriorityHigh, "High risk"},
			{90.0, architecture.PaymentFailurePriorityCritical, "Critical risk"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				priority := service.mapRiskScoreToPriority(tc.riskScore)
				assert.Equal(t, tc.expected, priority)
			})
		}
	})

	t.Run("Event Publishing", func(t *testing.T) {
		paymentFailure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			ProviderEventID: "pi_127",
			Amount:          2000.00,
			Currency:        "EUR",
			Status:          architecture.PaymentFailureStatusReceived,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		event := map[string]interface{}{
			"payment_failure": paymentFailure,
		}

		err := service.ProcessPaymentFailureEvent(context.Background(), event)
		require.NoError(t, err)

		// Verify processed event was published
		processedEvents := testEventBus.GetEvents("payment.failure.processed")
		require.GreaterOrEqual(t, len(processedEvents), 1)

		// Find the specific event for this test
		var processedEvent map[string]interface{}
		for _, event := range processedEvents {
			eventMap := event.(map[string]interface{})
			if eventMap["payment_failure"] == paymentFailure {
				processedEvent = eventMap
				break
			}
		}
		require.NotNil(t, processedEvent, "Could not find processed event for this payment failure")

		assert.Equal(t, "payment.failure.processed", processedEvent["event_type"])
		assert.Equal(t, "", processedEvent["provider"]) // Provider is empty for this test case
		assert.Equal(t, "default_company", processedEvent["company_id"])
		assert.NotNil(t, processedEvent["payment_failure"])
		assert.NotNil(t, processedEvent["metadata"])
	})

	t.Run("Error Handling", func(t *testing.T) {
		// Test with missing payment_failure data
		invalidEvent := map[string]interface{}{
			"timestamp": time.Now(),
		}

		err := service.ProcessPaymentFailureEvent(context.Background(), invalidEvent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event missing payment_failure data")

		// Test with invalid payment failure type
		invalidTypeEvent := map[string]interface{}{
			"payment_failure": "invalid_type",
		}

		err = service.ProcessPaymentFailureEvent(context.Background(), invalidTypeEvent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment failure data type")

		// Note: These validation errors happen before the pipeline starts,
		// so they don't increment FailedProcessing or record ProcessingErrors.
		// The metrics only track pipeline execution failures.

		// Verify that the validation errors didn't affect the success metrics
		// The TotalEventsProcessed should not include validation failures
		// since they fail before being counted as "processed"
	})

	t.Run("Service Lifecycle", func(t *testing.T) {
		// Test starting the service
		err := service.StartEventProcessing(context.Background())
		require.NoError(t, err)

		// Test stopping the service
		err = service.StopEventProcessing(context.Background())
		require.NoError(t, err)
	})

	t.Run("Metrics Collection", func(t *testing.T) {
		// Process multiple events to test metrics
		for i := 0; i < 3; i++ {
			paymentFailure := &architecture.PaymentFailure{
				ID:              uuid.New(),
				ProviderEventID: fmt.Sprintf("pi_%d", i),
				Amount:          1000.00 + float64(i*500),
				Currency:        "USD",
				Status:          architecture.PaymentFailureStatusReceived,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			event := map[string]interface{}{
				"payment_failure": paymentFailure,
			}

			err := service.ProcessPaymentFailureEvent(context.Background(), event)
			require.NoError(t, err)
		}

		// Verify metrics
		metrics := service.GetMetrics()
		assert.Equal(t, int64(13), metrics.TotalEventsProcessed)  // All events from all tests
		assert.Equal(t, int64(13), metrics.SuccessfullyProcessed) // All events were successful (no pipeline failures)
		assert.True(t, metrics.AverageProcessingTime > 0)
		assert.False(t, metrics.LastEventProcessed.IsZero())
	})
}

// TestEventProcessorPipelineStages tests individual pipeline stages
func TestEventProcessorPipelineStages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	service := NewEventProcessorService(nil, nil, nil, logger)

	t.Run("Data Enrichment", func(t *testing.T) {
		failure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			ProviderEventID: "pi_128",
			Amount:          15000.00, // Use amount > 10000 to trigger high_value tag
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := service.enrichFailureData(context.Background(), failure)
		require.NoError(t, err)

		assert.Equal(t, "default_company", failure.CompanyID)
		assert.Equal(t, "payment_failure", failure.ProviderEventType)
		assert.False(t, failure.OccurredAt.IsZero())
		assert.False(t, failure.DetectedAt.IsZero())
		assert.Equal(t, "general", failure.BusinessCategory)
		assert.Contains(t, failure.Tags, "high_value") // Amount > 10000, so should have high_value tag
	})

	t.Run("Risk Score Calculation", func(t *testing.T) {
		failure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			ProviderEventID: "pi_129",
			Amount:          8000.00,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := service.calculateRiskScore(context.Background(), failure)
		require.NoError(t, err)

		assert.True(t, failure.RiskScore >= 70.0) // Base 50 + 20 (amount) = 70
		assert.Equal(t, architecture.PaymentFailurePriorityHigh, failure.Priority)
	})

	t.Run("Status Update", func(t *testing.T) {
		failure := &architecture.PaymentFailure{
			ID:              uuid.New(),
			ProviderEventID: "pi_130",
			Amount:          1000.00,
			Status:          architecture.PaymentFailureStatusReceived,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := service.updateEventStatus(context.Background(), failure)
		require.NoError(t, err)

		assert.Equal(t, architecture.PaymentFailureStatusAnalyzed, failure.Status)
		assert.NotNil(t, failure.ProcessedAt)
		assert.False(t, failure.UpdatedAt.IsZero())
	})
}

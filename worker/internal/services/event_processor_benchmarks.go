package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/lexure-intelligence/payment-watchdog/internal/architecture"
)

// BenchmarkEventProcessorService provides comprehensive performance benchmarks
func BenchmarkEventProcessorService(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create test event bus
	testEventBus := &BenchmarkEventBus{
		events: make(map[string][]interface{}),
	}

	// Create event processor service
	service := NewEventProcessorService(nil, nil, testEventBus, logger)

	// Benchmark data preparation
	benchmarkData := prepareBenchmarkData()

	b.Run("SingleEventProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := benchmarkData[i%len(benchmarkData)]
			_ = service.ProcessPaymentFailureEvent(context.Background(), event)
		}
	})

	b.Run("ConcurrentEventProcessing", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				event := benchmarkData[i%len(benchmarkData)]
				_ = service.ProcessPaymentFailureEvent(context.Background(), event)
				i++
			}
		})
	})

	b.Run("HighVolumeProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Process multiple events in sequence
			for j := 0; j < 100; j++ {
				event := benchmarkData[(i+j)%len(benchmarkData)]
				_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			}
		}
	})
}

// BenchmarkEventProcessorLatency measures processing latency under various conditions
func BenchmarkEventProcessorLatency(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	service := NewEventProcessorService(nil, nil, nil, logger)

	b.Run("LowRiskEventLatency", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := createLowRiskEvent()
			start := time.Now()
			_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})

	b.Run("HighRiskEventLatency", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := createHighRiskEvent()
			start := time.Now()
			_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})

	b.Run("ComplexEventLatency", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := createComplexEvent()
			start := time.Now()
			_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})
}

// BenchmarkEventProcessorThroughput measures events processed per second
func BenchmarkEventProcessorThroughput(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	service := NewEventProcessorService(nil, nil, nil, logger)
	benchmarkData := prepareBenchmarkData()

	b.Run("Throughput100Events", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			for j := 0; j < 100; j++ {
				event := benchmarkData[j%len(benchmarkData)]
				_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			}
			duration := time.Since(start)
			throughput := float64(100) / duration.Seconds()
			b.ReportMetric(throughput, "events/sec")
		}
	})

	b.Run("Throughput1000Events", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			for j := 0; j < 1000; j++ {
				event := benchmarkData[j%len(benchmarkData)]
				_ = service.ProcessPaymentFailureEvent(context.Background(), event)
			}
			duration := time.Since(start)
			throughput := float64(1000) / duration.Seconds()
			b.ReportMetric(throughput, "events/sec")
		}
	})
}

// BenchmarkEventProcessorMemory measures memory allocation patterns
func BenchmarkEventProcessorMemory(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	service := NewEventProcessorService(nil, nil, nil, logger)
	benchmarkData := prepareBenchmarkData()

	b.Run("MemoryAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := benchmarkData[i%len(benchmarkData)]
			_ = service.ProcessPaymentFailureEvent(context.Background(), event)
		}
	})
}

// prepareBenchmarkData creates a variety of test events for benchmarking
func prepareBenchmarkData() []map[string]interface{} {
	var events []map[string]interface{}

	// Low risk events
	for i := 0; i < 10; i++ {
		events = append(events, createLowRiskEvent())
	}

	// Medium risk events
	for i := 0; i < 10; i++ {
		events = append(events, createMediumRiskEvent())
	}

	// High risk events
	for i := 0; i < 10; i++ {
		events = append(events, createHighRiskEvent())
	}

	// Complex events with various characteristics
	for i := 0; i < 10; i++ {
		events = append(events, createComplexEvent())
	}

	return events
}

// createLowRiskEvent creates a low-risk payment failure event
func createLowRiskEvent() map[string]interface{} {
	paymentFailure := &architecture.PaymentFailure{
		ID:              uuid.New(),
		CompanyID:       "company-benchmark",
		ProviderID:      "stripe",
		ProviderEventID: "pi_benchmark_low",
		Amount:          500.00,
		Currency:        "USD",
		CustomerID:      "customer-benchmark",
		CustomerName:    "Benchmark Customer",
		FailureReason:   "card_declined",
		Status:          architecture.PaymentFailureStatusReceived,
		Priority:        architecture.PaymentFailurePriorityLow,
		RiskScore:       0,
		OccurredAt:      time.Now().Add(-time.Hour),
		DetectedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return map[string]interface{}{
		"payment_failure": paymentFailure,
		"timestamp":       time.Now(),
	}
}

// createMediumRiskEvent creates a medium-risk payment failure event
func createMediumRiskEvent() map[string]interface{} {
	paymentFailure := &architecture.PaymentFailure{
		ID:              uuid.New(),
		CompanyID:       "company-benchmark",
		ProviderID:      "xero",
		ProviderEventID: "pi_benchmark_medium",
		Amount:          2500.00,
		Currency:        "AUD",
		CustomerID:      "customer-benchmark",
		CustomerName:    "Benchmark Customer",
		FailureReason:   "invoice_unpaid",
		Status:          architecture.PaymentFailureStatusReceived,
		Priority:        architecture.PaymentFailurePriorityMedium,
		RiskScore:       0,
		OccurredAt:      time.Now().Add(-time.Hour),
		DetectedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return map[string]interface{}{
		"payment_failure": paymentFailure,
		"timestamp":       time.Now(),
	}
}

// createHighRiskEvent creates a high-risk payment failure event
func createHighRiskEvent() map[string]interface{} {
	paymentFailure := &architecture.PaymentFailure{
		ID:              uuid.New(),
		CompanyID:       "company-benchmark",
		ProviderID:      "quickbooks",
		ProviderEventID: "pi_benchmark_high",
		Amount:          15000.00,
		Currency:        "USD",
		CustomerID:      "customer-benchmark",
		CustomerName:    "Benchmark Customer",
		FailureReason:   "invoice_unpaid",
		Status:          architecture.PaymentFailureStatusReceived,
		Priority:        architecture.PaymentFailurePriorityHigh,
		RiskScore:       0,
		OccurredAt:      time.Now().Add(-time.Hour),
		DetectedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return map[string]interface{}{
		"payment_failure": paymentFailure,
		"timestamp":       time.Now(),
	}
}

// createComplexEvent creates a complex payment failure event with various characteristics
func createComplexEvent() map[string]interface{} {
	dueDate := time.Now().AddDate(0, 0, -45) // 45 days overdue
	paymentFailure := &architecture.PaymentFailure{
		ID:              uuid.New(),
		CompanyID:       "company-benchmark",
		ProviderID:      "stripe",
		ProviderEventID: "pi_benchmark_complex",
		Amount:          8500.00,
		Currency:        "EUR",
		CustomerID:      "customer-benchmark",
		CustomerName:    "Benchmark Customer",
		FailureReason:   "invoice_unpaid",
		Status:          architecture.PaymentFailureStatusReceived,
		Priority:        architecture.PaymentFailurePriorityHigh,
		RiskScore:       0,
		DueDate:         &dueDate,
		BusinessCategory: "construction",
		Tags:            []string{"high_value", "overdue", "construction"},
		OccurredAt:      time.Now().Add(-time.Hour),
		DetectedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return map[string]interface{}{
		"payment_failure": paymentFailure,
		"timestamp":       time.Now(),
	}
}

// BenchmarkEventBus is a mock event bus for benchmarking
type BenchmarkEventBus struct {
	events map[string][]interface{}
}

func (b *BenchmarkEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	if b.events == nil {
		b.events = make(map[string][]interface{})
	}
	b.events[topic] = append(b.events[topic], event)
	return nil
}

func (b *BenchmarkEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return b.Publish(ctx, topic, event)
}

func (b *BenchmarkEventBus) Subscribe(ctx context.Context, topic string, handler architecture.EventHandler) (architecture.Subscription, error) {
	return &BenchmarkSubscription{}, nil
}

func (b *BenchmarkEventBus) SubscribeAsync(ctx context.Context, topic string, handler architecture.EventHandler) (architecture.Subscription, error) {
	return b.Subscribe(ctx, topic, handler)
}

func (b *BenchmarkEventBus) Unsubscribe(subscription architecture.Subscription) error {
	return nil
}

func (b *BenchmarkEventBus) Close() error {
	return nil
}

// BenchmarkSubscription is a mock subscription for benchmarking
type BenchmarkSubscription struct{}

func (b *BenchmarkSubscription) ID() string {
	return "benchmark-subscription"
}

func (b *BenchmarkSubscription) Topic() string {
	return "benchmark-topic"
}

func (b *BenchmarkSubscription) Unsubscribe() error {
	return nil
}

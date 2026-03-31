package interfaces

import (
	"context"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
)

// DatabaseInterface defines interface for database operations
type DatabaseInterface interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetHealthStatus() *DatabaseHealthStatus
}

// DatabaseHealthStatus represents database health status
type DatabaseHealthStatus struct {
	IsConnected bool   `json:"is_connected"`
	LastCheck   string `json:"last_check"`
	ErrorCount  int    `json:"error_count"`
}

// PaymentProcessor defines interface for processing payment events
type PaymentProcessor interface {
	ProcessPaymentFailure(ctx context.Context, event *events.PaymentEvent) error
	ValidatePaymentData(ctx context.Context, payment map[string]interface{}) error
	GetProcessorStatus() *ProcessorStatus
}

// ProcessorStatus represents the status of a payment processor
type ProcessorStatus struct {
	IsHealthy   bool   `json:"is_healthy"`
	LastProcess string `json:"last_process"`
	ErrorCount  int    `json:"error_count"`
	Throughput  int    `json:"throughput"`
}

// EventBusInterface defines interface for event bus operations
type EventBusInterface interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Subscribe(ctx context.Context, topic string, handler interface{}) error
	Close() error
	GetHealthStatus() *EventBusStatus
}

// EventHandler defines interface for event handling
type EventHandler interface {
	Handle(ctx context.Context, event interface{}) error
}

// Subscription defines interface for event subscription
type Subscription interface {
	ID() string
	Topic() string
	Unsubscribe() error
}

// EventBusStatus represents the status of the event bus
type EventBusStatus struct {
	IsConnected     bool `json:"is_connected"`
	SubscriberCount int  `json:"subscriber_count"`
	MessageCount    int  `json:"message_count"`
	ErrorCount      int  `json:"error_count"`
}

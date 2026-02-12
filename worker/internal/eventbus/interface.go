package eventbus

import "context"

// EventBus defines the interface for asynchronous event communication
type EventBus interface {
	// Publish events
	Publish(ctx context.Context, topic string, event interface{}) error
	PublishAsync(ctx context.Context, topic string, event interface{}) error
	
	// Subscribe to events
	Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error)
	SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error)
	
	// Event management
	Unsubscribe(subscription Subscription) error
	Close() error
}

// EventHandler processes incoming events
type EventHandler func(ctx context.Context, event interface{}) error

// Subscription represents an event subscription
type Subscription interface {
	ID() string
	Topic() string
	Unsubscribe() error
}

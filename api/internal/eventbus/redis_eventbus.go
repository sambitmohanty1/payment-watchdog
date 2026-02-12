package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RedisEventBus implements EventBus using Redis pub/sub
type RedisEventBus struct {
	client      *redis.Client
	logger      *zap.Logger
	subscribers map[string][]*RedisSubscription
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// RedisSubscription represents a Redis event subscription
type RedisSubscription struct {
	id       string
	topic    string
	handler  EventHandler
	eventBus *RedisEventBus
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRedisEventBus creates a new Redis-based event bus
func NewRedisEventBus(redisAddr, redisPassword string, db int, logger *zap.Logger) (*RedisEventBus, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	eventBus := &RedisEventBus{
		client:      client,
		logger:      logger,
		subscribers: make(map[string][]*RedisSubscription),
		ctx:         ctx,
		cancel:      cancel,
	}

	return eventBus, nil
}

// Publish publishes an event to a topic
func (r *RedisEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis
	result := r.client.Publish(ctx, topic, eventData)
	if result.Err() != nil {
		return fmt.Errorf("failed to publish event to Redis: %w", result.Err())
	}

	// Log successful publish
	r.logger.Debug("Event published",
		zap.String("topic", topic),
		zap.Int64("recipients", result.Val()),
		zap.String("event_type", fmt.Sprintf("%T", event)))

	return nil
}

// PublishAsync publishes an event asynchronously
func (r *RedisEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	go func() {
		if err := r.Publish(ctx, topic, event); err != nil {
			r.logger.Error("Async event publish failed",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}()

	return nil
}

// Subscribe subscribes to events on a topic
func (r *RedisEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	// Create subscription
	subCtx, cancel := context.WithCancel(ctx)
	subscription := &RedisSubscription{
		id:       uuid.New().String(),
		topic:    topic,
		handler:  handler,
		eventBus: r,
		ctx:      subCtx,
		cancel:   cancel,
	}

	// Add to subscribers
	r.mutex.Lock()
	r.subscribers[topic] = append(r.subscribers[topic], subscription)
	r.mutex.Unlock()

	// Start listening for events
	go r.listenForEvents(subscription)

	r.logger.Info("Subscription created",
		zap.String("subscription_id", subscription.id),
		zap.String("topic", topic))

	return subscription, nil
}

// SubscribeAsync subscribes to events asynchronously
func (r *RedisEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	// For Redis, async subscription is the same as sync
	return r.Subscribe(ctx, topic, handler)
}

// Unsubscribe removes a subscription
func (r *RedisEventBus) Unsubscribe(subscription Subscription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Find and remove subscription
	if subscribers, exists := r.subscribers[subscription.Topic()]; exists {
		for i, sub := range subscribers {
			if sub.ID() == subscription.ID() {
				// Cancel the subscription context
				sub.cancel()

				// Remove from slice
				r.subscribers[subscription.Topic()] = append(subscribers[:i], subscribers[i+1:]...)

				r.logger.Info("Subscription removed",
					zap.String("subscription_id", subscription.ID()),
					zap.String("topic", subscription.Topic()))

				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found: %s", subscription.ID())
}

// Close closes the event bus and cleans up resources
func (r *RedisEventBus) Close() error {
	r.cancel()

	// Close Redis client
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	r.logger.Info("Redis event bus closed")
	return nil
}

// listenForEvents listens for events on a specific subscription
func (r *RedisEventBus) listenForEvents(subscription *RedisSubscription) {
	// Create Redis pubsub
	pubsub := r.client.Subscribe(r.ctx, subscription.topic)
	defer pubsub.Close()

	// Channel for receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case <-subscription.ctx.Done():
			r.logger.Debug("Subscription context cancelled",
				zap.String("subscription_id", subscription.id),
				zap.String("topic", subscription.topic))
			return

		case msg := <-ch:
			if msg.Channel == subscription.topic {
				// Process the event
				if err := r.processEvent(subscription, msg.Payload); err != nil {
					r.logger.Error("Failed to process event",
						zap.String("subscription_id", subscription.id),
						zap.String("topic", subscription.topic),
						zap.Error(err))
				}
			}
		}
	}
}

// processEvent processes a single event
func (r *RedisEventBus) processEvent(subscription *RedisSubscription, payload string) error {
	// Try to deserialize as JSON first
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &eventData); err != nil {
		// If not JSON, treat as raw string
		eventData = map[string]interface{}{
			"data": payload,
		}
	}

	// Call the handler
	if err := subscription.handler(subscription.ctx, eventData); err != nil {
		return fmt.Errorf("handler failed: %w", err)
	}

	return nil
}

// ID returns the subscription ID
func (s *RedisSubscription) ID() string {
	return s.id
}

// Topic returns the subscription topic
func (s *RedisSubscription) Topic() string {
	return s.topic
}

// Unsubscribe removes this subscription
func (s *RedisSubscription) Unsubscribe() error {
	return s.eventBus.Unsubscribe(s)
}

// GetStats returns event bus statistics
func (r *RedisEventBus) GetStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]interface{})

	// Count subscribers per topic
	for topic, subscribers := range r.subscribers {
		stats[topic] = len(subscribers)
	}

	// Redis info
	if info, err := r.client.Info(r.ctx).Result(); err == nil {
		stats["redis_info"] = info
	}

	return stats
}

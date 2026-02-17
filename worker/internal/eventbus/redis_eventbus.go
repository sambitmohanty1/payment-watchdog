package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RedisEventBus implements EventBus using Redis Streams for durability
type RedisEventBus struct {
	client      *redis.Client
	logger      *zap.Logger
	subscribers map[string][]*RedisSubscription
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

type RedisSubscription struct {
	id       string
	topic    string
	handler  EventHandler
	eventBus *RedisEventBus
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewRedisEventBus(redisAddr, redisPassword string, db int, logger *zap.Logger) (*RedisEventBus, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisEventBus{
		client:      client,
		logger:      logger,
		subscribers: make(map[string][]*RedisSubscription),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Publish uses Redis Streams (XADD) for durability
func (r *RedisEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// XADD ensures persistence
	cmd := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"payload": eventData,
			"type":    fmt.Sprintf("%T", event),
		},
	})

	if cmd.Err() != nil {
		return fmt.Errorf("failed to publish to Redis Stream: %w", cmd.Err())
	}
	return nil
}

func (r *RedisEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	go func() {
		if err := r.Publish(ctx, topic, event); err != nil {
			r.logger.Error("Async publish failed", zap.String("topic", topic), zap.Error(err))
		}
	}()
	return nil
}

// Subscribe uses Consumer Groups to prevent data loss
func (r *RedisEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	subCtx, cancel := context.WithCancel(ctx)
	subscription := &RedisSubscription{
		id:       uuid.New().String(),
		topic:    topic,
		handler:  handler,
		eventBus: r,
		ctx:      subCtx,
		cancel:   cancel,
	}

	r.mutex.Lock()
	r.subscribers[topic] = append(r.subscribers[topic], subscription)
	r.mutex.Unlock()

	go r.listenForEvents(subscription)
	return subscription, nil
}

func (r *RedisEventBus) listenForEvents(subscription *RedisSubscription) {
	groupName := "payment-watchdog-workers"
	consumerName := "worker-" + subscription.id

	// Create Group if not exists
	r.client.XGroupCreateMkStream(r.ctx, subscription.topic, groupName, "0").Err()

	for {
		select {
		case <-subscription.ctx.Done():
			return
		default:
			// Read from Stream
			streams, err := r.client.XReadGroup(r.ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{subscription.topic, ">"},
				Count:    10,
				Block:    2 * time.Second,
			}).Result()

			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					payloadStr, ok := msg.Values["payload"].(string)
					if !ok {
						continue
					}

					var eventData map[string]interface{}
					if err := json.Unmarshal([]byte(payloadStr), &eventData); err == nil {
						if err := subscription.handler(subscription.ctx, eventData); err == nil {
							// Ack on success
							r.client.XAck(r.ctx, subscription.topic, groupName, msg.ID)
						} else {
							r.logger.Error("Handler failed", zap.Error(err))
						}
					}
				}
			}
		}
	}
}

func (r *RedisEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	return r.Subscribe(ctx, topic, handler)
}

func (r *RedisEventBus) Close() error {
	r.cancel()
	return r.client.Close()
}

func (s *RedisSubscription) ID() string         { return s.id }
func (s *RedisSubscription) Topic() string      { return s.topic }
func (s *RedisSubscription) Unsubscribe() error { return nil } // simplified

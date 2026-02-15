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

// Subscribe implements Consumer Group logic for reliable delivery
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

	// Start consuming in a goroutine
	go r.consumeStream(subscription)

	return subscription, nil
}

func (r *RedisEventBus) consumeStream(sub *RedisSubscription) {
	groupName := "payment-watchdog-workers"
	consumerName := "worker-" + sub.id

	// 1. Ensure Consumer Group exists (idempotent operation)
	// We ignore "BUSYGROUP" errors which mean it already exists
	_ = r.client.XGroupCreateMkStream(sub.ctx, sub.topic, groupName, "0").Err()

	r.logger.Info("Started stream consumer", 
		zap.String("topic", sub.topic), 
		zap.String("group", groupName))

	for {
		select {
		case <-sub.ctx.Done():
			return
		default:
			// 2. Read new messages (">") using blocking read
			streams, err := r.client.XReadGroup(sub.ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{sub.topic, ">"},
				Count:    10,
				Block:    2 * time.Second,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					r.logger.Error("Failed to read stream", zap.Error(err))
				}
				// Backoff slightly on error/empty to prevent tight loop burning CPU
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 3. Process messages
			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := r.handleMessage(sub, msg, groupName); err != nil {
						r.logger.Error("Failed to process message", 
							zap.String("msg_id", msg.ID), 
							zap.Error(err))
						// Note: We do NOT Ack here, so it remains in Pending Entries List (PEL)
						// to be picked up by a recovery process later.
					} else {
						// 4. Ack on success
						r.client.XAck(sub.ctx, sub.topic, groupName, msg.ID)
					}
				}
			}
		}
	}
}

func (r *RedisEventBus) handleMessage(sub *RedisSubscription, msg redis.XMessage, group string) error {
	payloadStr, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("invalid payload format")
	}

	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(payloadStr), &eventData); err != nil {
		// Fallback for simple strings
		eventData = map[string]interface{}{"data": payloadStr}
	}

	// Inject metadata if needed by handler
	eventData["_msg_id"] = msg.ID
	
	return sub.handler(sub.ctx, eventData)
}

func (r *RedisEventBus) Close() error {
	r.cancel()
	return r.client.Close()
}

// Interface compliance stub
func (r *RedisEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	// Worker primarily consumes, but if it needs to publish, use XADD
	data, _ := json.Marshal(event)
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{"payload": data},
	}).Err()
}

func (r *RedisEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return r.Publish(ctx, topic, event)
}

func (r *RedisEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	return r.Subscribe(ctx, topic, handler)
}

func (r *RedisEventBus) Unsubscribe(s Subscription) error { return nil }
func (s *RedisSubscription) ID() string                   { return s.id }
func (s *RedisSubscription) Topic() string                { return s.topic }
func (s *RedisSubscription) Unsubscribe() error           { return nil }

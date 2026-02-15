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
    client      *redis.Client
    logger      *zap.Logger
    subscribers map[string][]*RedisSubscription
    mutex       sync.RWMutex
    ctx         context.Context
    cancel      context.CancelFunc
}

type RedisSubscription struct {
    id       string
    topic    string
    handler  EventHandler
    eventBus *RedisEventBus
    ctx      context.Context
    cancel   context.CancelFunc
}

func NewRedisEventBus(redisAddr, redisPassword string, db int, logger *zap.Logger) (*RedisEventBus, error) {
    ctx, cancel := context.WithCancel(context.Background())
    client := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword,
        DB:       db,
    })

    if err := client.Ping(ctx).Err(); err != nil {
        cancel()
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &RedisEventBus{
        client:      client,
        logger:      logger,
        subscribers: make(map[string][]*RedisSubscription),
        ctx:         ctx,
        cancel:      cancel,
    }, nil
}

func (r *RedisEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
    subCtx, cancel := context.WithCancel(ctx)
    subscription := &RedisSubscription{
        id:       uuid.New().String(),
        topic:    topic,
        handler:  handler,
        eventBus: r,
        ctx:      subCtx,
        cancel:   cancel,
    }

    r.mutex.Lock()
    r.subscribers[topic] = append(r.subscribers[topic], subscription)
    r.mutex.Unlock()

    go r.consumeStream(subscription)

    return subscription, nil
}

// consumeStream handles the reliable consumption of messages using Redis Streams
func (r *RedisEventBus) consumeStream(sub *RedisSubscription) {
    groupName := "payment-watchdog-workers"
    consumerName := "worker-" + sub.id

    // 1. Create Consumer Group (idempotent)
    _ = r.client.XGroupCreateMkStream(sub.ctx, sub.topic, groupName, "0").Err()

    r.logger.Info("Started stream consumer", zap.String("topic", sub.topic))

    // 2. Recovery Phase: Process pending messages (PEL) from crashed workers
    pending, _ := r.client.XReadGroup(sub.ctx, &redis.XReadGroupArgs{
        Group:    groupName,
        Consumer: consumerName,
        Streams:  []string{sub.topic, "0"},
        Count:    100,
        Block:    0,
    }).Result()

    for _, stream := range pending {
        for _, msg := range stream.Messages {
            if err := r.handleMessage(sub, msg); err == nil {
                r.client.XAck(sub.ctx, sub.topic, groupName, msg.ID)
            }
        }
    }

    // 3. Normal Phase: Process new messages
    for {
        select {
        case <-sub.ctx.Done():
            return
        default:
            streams, err := r.client.XReadGroup(sub.ctx, &redis.XReadGroupArgs{
                Group:    groupName,
                Consumer: consumerName,
                Streams:  []string{sub.topic, ">"},
                Count:    1,
                Block:    2 * time.Second,
            }).Result()

            if err != nil {
                time.Sleep(100 * time.Millisecond)
                continue
            }

            for _, stream := range streams {
                for _, msg := range stream.Messages {
                    if err := r.handleMessage(sub, msg); err == nil {
                        r.client.XAck(sub.ctx, sub.topic, groupName, msg.ID)
                    }
                }
            }
        }
    }
}

func (r *RedisEventBus) handleMessage(sub *RedisSubscription, msg redis.XMessage) error {
    payloadStr, ok := msg.Values["payload"].(string)
    if !ok {
        return fmt.Errorf("invalid payload")
    }

    var eventData map[string]interface{}
    if err := json.Unmarshal([]byte(payloadStr), &eventData); err != nil {
        return err
    }
    return sub.handler(sub.ctx, eventData)
}

func (r *RedisEventBus) Close() error {
    r.cancel()
    return r.client.Close()
}

// Stubs for interface compliance
func (r *RedisEventBus) Publish(ctx context.Context, topic string, event interface{}) error { return nil }
func (r *RedisEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error { return nil }
func (r *RedisEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) { return r.Subscribe(ctx, topic, handler) }
func (r *RedisEventBus) Unsubscribe(s Subscription) error { return nil }
func (s *RedisSubscription) ID() string { return s.id }
func (s *RedisSubscription) Topic() string { return s.topic }
func (s *RedisSubscription) Unsubscribe() error { return nil }

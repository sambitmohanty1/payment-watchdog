package eventbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LocalEventBus implements EventBus for in-memory communication
type LocalEventBus struct {
	logger      *zap.Logger
	subscribers map[string][]*localSubscription
	mutex       sync.RWMutex
}

type localSubscription struct {
	id      string
	topic   string
	handler EventHandler
}

func (s *localSubscription) ID() string         { return s.id }
func (s *localSubscription) Topic() string      { return s.topic }
func (s *localSubscription) Unsubscribe() error { return nil }

// NewLocalEventBus creates a new in-memory event bus
func NewLocalEventBus(logger *zap.Logger) *LocalEventBus {
	return &LocalEventBus{
		logger:      logger,
		subscribers: make(map[string][]*localSubscription),
	}
}

func (l *LocalEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	l.mutex.RLock()
	subs, ok := l.subscribers[topic]
	l.mutex.RUnlock()

	if !ok {
		l.logger.Debug("No subscribers for topic", zap.String("topic", topic))
		return nil
	}

	for _, sub := range subs {
		if err := sub.handler(ctx, event); err != nil {
			l.logger.Error("Handler failed for topic", 
				zap.String("topic", topic), 
				zap.String("sub_id", sub.id), 
				zap.Error(err))
		}
	}

	return nil
}

func (l *LocalEventBus) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	go l.Publish(ctx, topic, event)
	return nil
}

func (l *LocalEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	sub := &localSubscription{
		id:      uuid.New().String(),
		topic:   topic,
		handler: handler,
	}

	l.mutex.Lock()
	l.subscribers[topic] = append(l.subscribers[topic], sub)
	l.mutex.Unlock()

	return sub, nil
}

func (l *LocalEventBus) SubscribeAsync(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	return l.Subscribe(ctx, topic, handler)
}

func (l *LocalEventBus) Unsubscribe(subscription Subscription) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	subs, ok := l.subscribers[subscription.Topic()]
	if !ok {
		return nil
	}

	for i, sub := range subs {
		if sub.id == subscription.ID() {
			l.subscribers[subscription.Topic()] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	return nil
}

func (l *LocalEventBus) Close() error {
	return nil
}

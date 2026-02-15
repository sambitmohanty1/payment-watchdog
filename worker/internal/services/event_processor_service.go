package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/worker/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/worker/internal/models"
)

type EventProcessorService struct {
	db       *gorm.DB
	eventBus eventbus.EventBus
	logger   *zap.Logger
}

func NewEventProcessorService(db *gorm.DB, eb eventbus.EventBus, logger *zap.Logger) *EventProcessorService {
	return &EventProcessorService{
		db:       db,
		eventBus: eb,
		logger:   logger,
	}
}

func (s *EventProcessorService) Start(ctx context.Context) error {
	s.logger.Info("Starting Event Processor Service")

	// Subscribe to the "payment_failures" stream
	_, err := s.eventBus.Subscribe(ctx, "payment_failures", s.handlePaymentFailure)
	if err != nil {
		return fmt.Errorf("failed to subscribe to payment_failures: %w", err)
	}

	<-ctx.Done()
	return nil
}

func (s *EventProcessorService) handlePaymentFailure(ctx context.Context, payload interface{}) error {
	s.logger.Info("Processing payment failure event")

	// 1. Unmarshal payload
	// The EventBus passes us a map or the raw JSON bytes depending on implementation.
	// Since we standardized the EventBus to send map[string]interface{} unmarshaled from JSON:
	
	// We need to marshal it back to unmarshal into the struct, or map manually. 
	// To be safe and clean, let's assume payload comes as map from the updated EventBus.
	
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var event models.PaymentFailureEvent
	if err := json.Unmarshal(bytes, &event); err != nil {
		s.logger.Error("Failed to parse event payload", zap.Error(err))
		return nil // Don't retry malformed data
	}

	// 2. Business Logic with Int64
	if event.AmountCents == 0 {
		s.logger.Warn("Received event with 0 amount", zap.String("event_id", event.EventID))
	}

	// 3. Example: Log High Value failures
	// $1000.00 = 100000 cents
	if event.AmountCents > 100000 {
		s.logger.Info("High value payment failure detected", 
			zap.Int64("amount_cents", event.AmountCents),
			zap.String("currency", event.Currency))
	}

	// 4. Save/Update state in Worker DB
	if err := s.db.Save(&event).Error; err != nil {
		s.logger.Error("Failed to save event to DB", zap.Error(err))
		return err // Return error to trigger Redis retry (NACK)
	}

	return nil
}

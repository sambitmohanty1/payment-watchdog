package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
	"github.com/sambitmohanty1/payment-watchdog/worker/config"
)

// PaymentProcessorService handles payment failure events
type PaymentProcessorService struct {
	db       interfaces.DatabaseInterface
	eventBus interfaces.EventBusInterface
	config   *config.WorkerConfig
	logger   interfaces.LoggerInterface
	status   *interfaces.ProcessorStatus
}

// NewPaymentProcessorService creates a new payment processor service
func NewPaymentProcessorService(db interfaces.DatabaseInterface, eventBus interfaces.EventBusInterface, config *config.WorkerConfig, logger interfaces.LoggerInterface) *PaymentProcessorService {
	return &PaymentProcessorService{
		db:       db,
		eventBus: eventBus,
		config:   config,
		logger:   logger,
		status: &interfaces.ProcessorStatus{
			IsHealthy:   true,
			LastProcess: "",
			ErrorCount:  0,
			Throughput:  0,
		},
	}
}

// StartEventListeners starts listening for payment events
func (s *PaymentProcessorService) StartEventListeners(ctx context.Context) error {
	s.logger.Info("Starting payment processor event listeners")

	// Subscribe to payment failure events
	if err := s.eventBus.Subscribe(ctx, events.PaymentFailureDetected, &FunctionHandler{handleFunc: s.handlePaymentFailure}); err != nil {
		return fmt.Errorf("failed to subscribe to payment failure events: %w", err)
	}

	s.logger.Info("Payment processor event listeners started successfully")
	return nil
}

// handlePaymentFailure processes payment failure events
func (s *PaymentProcessorService) handlePaymentFailure(ctx context.Context, event interface{}) error {
	s.logger.Info("Processing payment failure event", zap.String("event_id", event.(*events.PaymentEvent).ID))

	paymentEvent, ok := event.(*events.PaymentEvent)
	if !ok {
		return fmt.Errorf("invalid event type, expected PaymentEvent")
	}

	// Process the payment failure
	if err := s.processFailure(ctx, paymentEvent); err != nil {
		s.status.ErrorCount++
		s.logger.Error("Failed to process payment failure", zap.Error(err), zap.String("event_id", paymentEvent.ID))
		return err
	}

	// Publish processed event
	processedEvent := events.NewPaymentProcessedEvent(
		paymentEvent.CompanyID,
		paymentEvent.PaymentID,
		paymentEvent.Amount,
		paymentEvent.Currency,
		"processed",
	)

	if err := s.eventBus.Publish(ctx, events.PaymentFailureProcessed, processedEvent); err != nil {
		s.logger.Error("Failed to publish processed event", zap.Error(err), zap.String("event_id", processedEvent.ID))
		return err
	}

	s.status.LastProcess = paymentEvent.ID
	s.status.Throughput++
	s.logger.Info("Payment failure processed successfully", zap.String("event_id", paymentEvent.ID), zap.String("processed_event_id", processedEvent.ID))

	return nil
}

// processFailure processes the actual payment failure logic
func (s *PaymentProcessorService) processFailure(ctx context.Context, event *events.PaymentEvent) error {
	// TODO: Implement actual payment failure processing logic
	// This would include:
	// - Validate payment data
	// - Determine recovery strategy
	// - Execute recovery workflow
	// - Update database records

	s.logger.Info("Processing payment failure",
		zap.String("company_id", event.CompanyID),
		zap.String("payment_id", event.PaymentID),
		zap.Float64("amount", event.Amount),
		zap.String("currency", event.Currency),
	)

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetProcessorStatus returns the current processor status
func (s *PaymentProcessorService) GetProcessorStatus() *interfaces.ProcessorStatus {
	return s.status
}

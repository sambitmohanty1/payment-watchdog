package services

import (
	"context"
	"fmt"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
	"github.com/sambitmohanty1/payment-watchdog/worker/config"
	"go.uber.org/zap"
)

// EventProcessorService handles payment failure events
type EventProcessorService struct {
	db        interfaces.DatabaseInterface
	eventBus  interfaces.EventBusInterface
	config    *config.WorkerConfig
	analytics interfaces.AnalyticsEngine
	rules     interfaces.RuleEngine
	mediators interfaces.MediatorService
	logger    *zap.Logger
	status    *interfaces.ProcessorStatus
}

// NewEventProcessorService creates a new event processor service
func NewEventProcessorService(db interfaces.DatabaseInterface, eventBus interfaces.EventBusInterface, config *config.WorkerConfig, analytics interfaces.AnalyticsEngine, rules interfaces.RuleEngine, mediators interfaces.MediatorService, logger *zap.Logger) *EventProcessorService {
	return &EventProcessorService{
		db:        db,
		eventBus:  eventBus,
		config:    config,
		analytics: analytics,
		rules:     rules,
		mediators: mediators,
		logger:    logger,
		status: &interfaces.ProcessorStatus{
			IsHealthy:   true,
			LastProcess: "",
			ErrorCount:  0,
			Throughput:  0,
		},
	}
}

// StartEventListeners starts listening for payment events
func (s *EventProcessorService) StartEventListeners(ctx context.Context) error {
	s.logger.Info("Starting event processor event listeners")

	// Subscribe to payment failure events
	if err := s.eventBus.Subscribe(ctx, events.PaymentFailureDetected, s.handlePaymentFailure); err != nil {
		return fmt.Errorf("failed to subscribe to payment failure events: %w", err)
	}

	s.logger.Info("Event processor event listeners started successfully")
	return nil
}

// handlePaymentFailure processes payment failure events
func (s *EventProcessorService) handlePaymentFailure(ctx context.Context, event interface{}) error {
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
func (s *EventProcessorService) processFailure(ctx context.Context, event *events.PaymentEvent) error {
	s.logger.Info("Processing payment failure",
		zap.String("event_id", event.ID),
		zap.Float64("amount", event.Amount),
	)

	// Get business rule evaluation
	ruleEvaluation, err := s.rules.EvaluatePaymentFailure(ctx, event)
	if err != nil {
		s.status.ErrorCount++
		s.logger.Error("Failed to evaluate payment failure", zap.Error(err), zap.String("event_id", event.ID))
		return err
	}

	// Get recovery recommendations
	recommendations, err := s.analytics.GetRecoveryRecommendations(ctx, event)
	if err != nil {
		s.status.ErrorCount++
		s.logger.Error("Failed to get recovery recommendations", zap.Error(err), zap.String("event_id", event.ID))
		return err
	}

	// Apply recovery strategy
	strategy, err := s.rules.ApplyRecoveryStrategy(ctx, event)
	if err != nil {
		s.status.ErrorCount++
		s.logger.Error("Failed to apply recovery strategy", zap.Error(err), zap.String("event_id", event.ID))
		return err
	}

	// Process through mediator if recommended
	if strategy.StrategyType == "mediator" {
		mediatorResult, err := s.mediators.ProcessPaymentThroughMediator(ctx, event)
		if err != nil {
			s.status.ErrorCount++
			s.logger.Error("Failed to process through mediator", zap.Error(err), zap.String("event_id", event.ID))
			return err
		}

		s.logger.Info("Payment processed through mediator", zap.String("transaction_id", mediatorResult.TransactionID))
	} else {
		// Standard recovery processing
		s.logger.Info("Standard recovery processing applied", zap.String("event_id", event.ID))
	}

	// Update status
	s.status.LastProcess = event.ID
	s.status.Throughput++
	s.logger.Info("Payment failure processed successfully", zap.String("event_id", event.ID))

	return nil
}

// GetProcessorStatus returns the current processor status
func (s *EventProcessorService) GetProcessorStatus() *interfaces.ProcessorStatus {
	return s.status
}

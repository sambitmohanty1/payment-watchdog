package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/shared/events"
	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
	"github.com/sambitmohanty1/payment-watchdog/worker/config"
)

// PaymentProcessorService handles payment failure events
type PaymentProcessorService struct {
	db           interfaces.DatabaseInterface
	eventBus     interfaces.EventBusInterface
	rulesService *RulesService
	config       *config.WorkerConfig
	logger       interfaces.LoggerInterface
	status       *interfaces.ProcessorStatus
}

// NewPaymentProcessorService creates a new payment processor service
func NewPaymentProcessorService(db interfaces.DatabaseInterface, eventBus interfaces.EventBusInterface, rulesService *RulesService, config *config.WorkerConfig, logger interfaces.LoggerInterface) *PaymentProcessorService {
	return &PaymentProcessorService{
		db:           db,
		eventBus:     eventBus,
		rulesService: rulesService,
		config:       config,
		logger:       logger,
		status: &interfaces.ProcessorStatus{
			IsHealthy:   true,
			LastProcess: "",
			ErrorCount:  0,
			Throughput:  1, // Initial value
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
	s.logger.Info("Processing payment failure",
		zap.String("company_id", event.CompanyID),
		zap.String("payment_id", event.PaymentID),
		zap.Float64("amount", event.Amount),
		zap.String("currency", event.Currency),
	)

	// 1. Persistence Layer: Record to DB (Phase 1 Logic)
	type gormDBProvider interface {
		GetDB() *gorm.DB
	}

	if dbProvider, ok := s.db.(gormDBProvider); ok {
		db := dbProvider.GetDB()
		if db != nil {
			var providerID, providerEventType, customerID, customerEmail, customerName, failureReason, failureCode, failureMessage string

			if val, ok := event.Metadata["provider"]; ok { providerID = fmt.Sprint(val) }
			if val, ok := event.Metadata["provider_event_type"]; ok { providerEventType = fmt.Sprint(val) }
			if val, ok := event.Metadata["customer_id"]; ok { customerID = fmt.Sprint(val) }
			if val, ok := event.Metadata["customer_email"]; ok { customerEmail = fmt.Sprint(val) }
			if val, ok := event.Metadata["customer_name"]; ok { customerName = fmt.Sprint(val) }
			if val, ok := event.Metadata["reason"]; ok { failureReason = fmt.Sprint(val) }
			if val, ok := event.Metadata["failure_code"]; ok { failureCode = fmt.Sprint(val) }
			if val, ok := event.Metadata["failure_message"]; ok { failureMessage = fmt.Sprint(val) }

			err := db.Exec(`
				INSERT INTO payment_failure_events (
					id, company_id, provider_id, event_id, event_type, 
					amount_cents, currency, customer_id, customer_email, 
					customer_name, failure_reason, failure_code, failure_message, 
					status, created_at, updated_at
				) VALUES (
					gen_random_uuid(), ?, ?, ?, ?,
					?, ?, ?, ?,
					?, ?, ?, ?,
					'received', NOW(), NOW()
				) ON CONFLICT (event_id) DO NOTHING
			`, 
				event.CompanyID, providerID, event.ID, providerEventType, 
				int64(event.Amount * 100), event.Currency, customerID, customerEmail, 
				customerName, failureReason, failureCode, failureMessage,
			).Error

			if err != nil {
				s.logger.Error("Failed to insert payment failure event to DB", zap.Error(err))
			} else {
				s.logger.Info("Successfully recorded payment failure event to DB")
			}
		}
	}

	// 2. Intelligence Layer: Classify the failure (Phase 2 Logic)
	if s.rulesService != nil {
		evaluation, err := s.rulesService.EvaluatePaymentFailure(ctx, event)
		if err != nil {
			s.logger.Error("Failed to evaluate payment failure rules", zap.Error(err))
		} else {
			s.logger.Info("Failure classified", 
				zap.String("rule", evaluation.RuleName),
				zap.String("action", evaluation.Action),
				zap.Float64("confidence", evaluation.Confidence),
			)

			// Determine recovery strategy
			strategy, err := s.rulesService.ApplyRecoveryStrategy(ctx, event)
			if err != nil {
				s.logger.Error("Failed to apply recovery strategy", zap.Error(err))
			} else {
				s.logger.Info("Recovery strategy determined",
					zap.String("strategy", strategy.StrategyType),
					zap.Time("next_attempt", strategy.NextAttempt),
				)
			}
		}
	}

	return nil
}

// GetProcessorStatus returns the current processor status
func (s *PaymentProcessorService) GetProcessorStatus() *interfaces.ProcessorStatus {
	return s.status
}

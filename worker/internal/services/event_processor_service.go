package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/lexure-intelligence/payment-watchdog/internal/architecture"
	"github.com/lexure-intelligence/payment-watchdog/internal/rules"
)

// EventProcessorService processes payment failure events and applies business intelligence
type EventProcessorService struct {
	db         *gorm.DB
	ruleEngine *rules.RuleEngine
	eventBus   architecture.EventBus
	logger     *zap.Logger

	// Processing metrics
	metrics *EventProcessorMetrics

	// Configuration
	maxRetries int
	retryDelay time.Duration
}

// EventProcessorMetrics tracks processing performance and statistics
type EventProcessorMetrics struct {
	TotalEventsProcessed  int64
	SuccessfullyProcessed int64
	FailedProcessing      int64
	AverageProcessingTime time.Duration
	LastEventProcessed    time.Time
	EventsByProvider      map[string]int64
	EventsByStatus        map[string]int64
	ProcessingErrors      []ProcessingError
}

// ProcessingError represents an error during event processing
type ProcessingError struct {
	EventID      string    `json:"event_id"`
	Provider     string    `json:"provider"`
	ErrorType    string    `json:"error_type"`
	ErrorMessage string    `json:"error_message"`
	Retryable    bool      `json:"retryable"`
	RetryCount   int       `json:"retry_count"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewEventProcessorService creates a new event processor service
func NewEventProcessorService(db *gorm.DB, ruleEngine *rules.RuleEngine, eventBus architecture.EventBus, logger *zap.Logger) *EventProcessorService {
	metrics := &EventProcessorMetrics{
		EventsByProvider: make(map[string]int64),
		EventsByStatus:   make(map[string]int64),
		ProcessingErrors: make([]ProcessingError, 0),
	}

	return &EventProcessorService{
		db:         db,
		ruleEngine: ruleEngine,
		eventBus:   eventBus,
		logger:     logger,
		metrics:    metrics,
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}
}

// ProcessPaymentFailureEvent processes a payment failure event through the complete pipeline
func (e *EventProcessorService) ProcessPaymentFailureEvent(ctx context.Context, event map[string]interface{}) error {
	startTime := time.Now()

	// Extract payment failure from event
	paymentFailureData, ok := event["payment_failure"]
	if !ok {
		return fmt.Errorf("event missing payment_failure data")
	}

	// Convert to PaymentFailure struct
	paymentFailure, ok := paymentFailureData.(*architecture.PaymentFailure)
	if !ok {
		return fmt.Errorf("invalid payment failure data type")
	}

	e.logger.Info("Processing payment failure event",
		zap.String("event_id", paymentFailure.ID.String()),
		zap.String("provider", paymentFailure.ProviderID),
		zap.String("company_id", paymentFailure.CompanyID),
		zap.Float64("amount", paymentFailure.Amount))

	// Update metrics
	e.metrics.TotalEventsProcessed++
	e.metrics.EventsByProvider[paymentFailure.ProviderID]++
	e.metrics.LastEventProcessed = time.Now()

	// Process the event through the pipeline
	if err := e.processEventPipeline(ctx, paymentFailure); err != nil {
		e.logger.Error("Event processing pipeline failed",
			zap.String("event_id", paymentFailure.ID.String()),
			zap.Error(err))

		e.metrics.FailedProcessing++
		e.recordProcessingError(paymentFailure, err)
		return fmt.Errorf("event processing pipeline failed: %w", err)
	}

	// Update success metrics
	e.metrics.SuccessfullyProcessed++
	processingTime := time.Since(startTime)
	e.metrics.AverageProcessingTime = e.calculateAverageProcessingTime(processingTime)

	// Store processing event in database
	if err := e.storeProcessingEvent(ctx, paymentFailure, processingTime); err != nil {
		e.logger.Warn("Failed to store processing event, but processing succeeded",
			zap.String("event_id", paymentFailure.ID.String()),
			zap.Error(err))
		// Don't fail the entire operation if storage fails
	}

	e.logger.Info("Payment failure event processed successfully",
		zap.String("event_id", paymentFailure.ID.String()),
		zap.Duration("processing_time", processingTime))

	return nil
}

// processEventPipeline processes the event through all pipeline stages
func (e *EventProcessorService) processEventPipeline(ctx context.Context, failure *architecture.PaymentFailure) error {
	// Stage 1: Enrich with additional data
	if err := e.enrichFailureData(ctx, failure); err != nil {
		return fmt.Errorf("failed to enrich failure data: %w", err)
	}

	// Stage 2: Calculate risk score
	if err := e.calculateRiskScore(ctx, failure); err != nil {
		return fmt.Errorf("failed to calculate risk score: %w", err)
	}

	// Stage 3: Execute business rules
	if err := e.executeBusinessRules(ctx, failure); err != nil {
		return fmt.Errorf("failed to execute business rules: %w", err)
	}

	// Stage 4: Update status and persist
	if err := e.updateEventStatus(ctx, failure); err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	// Stage 5: Publish processed event
	if err := e.publishProcessedEvent(ctx, failure); err != nil {
		return fmt.Errorf("failed to publish processed event: %w", err)
	}

	return nil
}

// enrichFailureData enriches the payment failure with additional context
func (e *EventProcessorService) enrichFailureData(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Debug("Enriching failure data",
		zap.String("event_id", failure.ID.String()))

	// Set default values if not provided
	if failure.CompanyID == "" {
		failure.CompanyID = "default_company"
	}

	if failure.ProviderEventType == "" {
		failure.ProviderEventType = "payment_failure"
	}

	if failure.OccurredAt.IsZero() {
		failure.OccurredAt = time.Now()
	}

	if failure.DetectedAt.IsZero() {
		failure.DetectedAt = time.Now()
	}

	// Add business category if available
	if failure.BusinessCategory == "" {
		failure.BusinessCategory = "general"
	}

	// Add tags based on failure characteristics
	if failure.Amount > 10000 {
		failure.Tags = append(failure.Tags, "high_value")
	}

	if failure.FailureReason == "invoice_unpaid" {
		failure.Tags = append(failure.Tags, "invoice_based")
	}

	return nil
}

// calculateRiskScore calculates the risk score for the payment failure
func (e *EventProcessorService) calculateRiskScore(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Debug("Calculating risk score",
		zap.String("event_id", failure.ID.String()))

	// Base risk score starts at 50
	riskScore := 50.0

	// Factor 1: Amount-based risk
	if failure.Amount >= 10000 {
		riskScore += 30
	} else if failure.Amount >= 5000 {
		riskScore += 20
	} else if failure.Amount >= 1000 {
		riskScore += 10
	}

	// Factor 2: Overdue days (if available)
	if failure.DueDate != nil {
		overdueDays := int(time.Since(*failure.DueDate).Hours() / 24)
		if overdueDays > 90 {
			riskScore += 20
		} else if overdueDays > 60 {
			riskScore += 15
		} else if overdueDays > 30 {
			riskScore += 10
		} else if overdueDays > 7 {
			riskScore += 5
		}
	}

	// Factor 3: Business category risk
	switch failure.BusinessCategory {
	case "construction", "manufacturing", "healthcare":
		riskScore += 10 // High-risk industries
	case "retail", "hospitality":
		riskScore += 5 // Medium-risk industries
	}

	// Factor 4: Customer history (placeholder for future enhancement)
	// TODO: Query customer history from database
	// if customerHasPreviousFailures {
	//     riskScore += 15
	// }

	// Cap at 100
	if riskScore > 100 {
		riskScore = 100
	}

	failure.RiskScore = riskScore

	// Determine priority based on risk score
	failure.Priority = e.mapRiskScoreToPriority(riskScore)

	e.logger.Debug("Risk score calculated",
		zap.String("event_id", failure.ID.String()),
		zap.Float64("risk_score", riskScore),
		zap.String("priority", string(failure.Priority)))

	return nil
}

// mapRiskScoreToPriority maps risk score to priority level
func (e *EventProcessorService) mapRiskScoreToPriority(riskScore float64) architecture.PaymentFailurePriority {
	if riskScore >= 80 {
		return architecture.PaymentFailurePriorityCritical
	} else if riskScore >= 60 {
		return architecture.PaymentFailurePriorityHigh
	} else if riskScore >= 40 {
		return architecture.PaymentFailurePriorityMedium
	} else {
		return architecture.PaymentFailurePriorityLow
	}
}

// executeBusinessRules executes business rules on the payment failure
func (e *EventProcessorService) executeBusinessRules(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Debug("Executing business rules",
		zap.String("event_id", failure.ID.String()))

	// TODO: Convert architecture.PaymentFailure to models.PaymentFailureEvent
	// For now, we'll skip rule execution until the models are aligned
	// results := e.ruleEngine.ExecuteRules(ruleContext)

	// Placeholder for rule execution results
	var results []*rules.ActionResult

	// Process rule results
	for _, result := range results {
		if !result.Success {
			e.logger.Warn("Rule execution failed",
				zap.String("rule_name", result.RuleName),
				zap.Error(result.Error))
			continue
		}

		e.logger.Info("Rule executed successfully",
			zap.String("rule_name", result.RuleName),
			zap.String("message", result.Message))

		// Handle specific rule results
		if err := e.handleRuleResult(ctx, failure, result); err != nil {
			e.logger.Error("Failed to handle rule result",
				zap.String("rule_name", result.RuleName),
				zap.Error(err))
		}
	}

	return nil
}

// handleRuleResult processes the result of a specific rule execution
func (e *EventProcessorService) handleRuleResult(ctx context.Context, failure *architecture.PaymentFailure, result *rules.ActionResult) error {
	switch result.RuleName {
	case "high_risk_alert":
		return e.createHighRiskAlert(ctx, failure)
	case "retry_payment":
		return e.scheduleRetry(ctx, failure)
	case "customer_communication":
		return e.scheduleCustomerCommunication(ctx, failure)
	case "escalate_to_manager":
		return e.escalateToManager(ctx, failure)
	default:
		e.logger.Warn("Unknown rule result", zap.String("rule_name", result.RuleName))
		return nil
	}
}

// createHighRiskAlert creates a high-risk alert for critical payment failures
func (e *EventProcessorService) createHighRiskAlert(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Info("Creating high-risk alert",
		zap.String("event_id", failure.ID.String()),
		zap.Float64("amount", failure.Amount),
		zap.Float64("risk_score", failure.RiskScore))

	// TODO: Implement alert creation logic
	// This would typically create an alert record in the database
	// and potentially trigger notifications

	return nil
}

// scheduleRetry schedules a payment retry for the failure
func (e *EventProcessorService) scheduleRetry(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Info("Scheduling payment retry",
		zap.String("event_id", failure.ID.String()))

	// TODO: Implement retry scheduling logic
	// This would typically create a retry record and schedule it

	return nil
}

// scheduleCustomerCommunication schedules customer communication
func (e *EventProcessorService) scheduleCustomerCommunication(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Info("Scheduling customer communication",
		zap.String("event_id", failure.ID.String()))

	// TODO: Implement customer communication scheduling
	// This would typically create a communication record

	return nil
}

// escalateToManager escalates the failure to a manager
func (e *EventProcessorService) escalateToManager(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Info("Escalating to manager",
		zap.String("event_id", failure.ID.String()))

	// TODO: Implement escalation logic
	// This would typically create an escalation record

	return nil
}

// updateEventStatus updates the event status and persists to database
func (e *EventProcessorService) updateEventStatus(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Debug("Updating event status",
		zap.String("event_id", failure.ID.String()))

	// Update status to analyzed
	failure.Status = architecture.PaymentFailureStatusAnalyzed
	failure.ProcessedAt = &time.Time{}
	*failure.ProcessedAt = time.Now()
	failure.UpdatedAt = time.Now()

	// Save to database
	if e.db != nil {
		if err := e.db.Save(failure).Error; err != nil {
			return fmt.Errorf("failed to save failure status: %w", err)
		}
		e.logger.Debug("Payment failure status saved to database",
			zap.String("event_id", failure.ID.String()),
			zap.String("status", string(failure.Status)))
	} else {
		e.logger.Warn("Database not available, skipping persistence",
			zap.String("event_id", failure.ID.String()))
	}

	// Update metrics
	e.metrics.EventsByStatus[string(failure.Status)]++

	return nil
}

// publishProcessedEvent publishes the processed event to the event bus
func (e *EventProcessorService) publishProcessedEvent(ctx context.Context, failure *architecture.PaymentFailure) error {
	e.logger.Debug("Publishing processed event",
		zap.String("event_id", failure.ID.String()))

	processedEvent := map[string]interface{}{
		"event_id":        uuid.New().String(),
		"event_type":      "payment.failure.processed",
		"provider":        failure.ProviderID,
		"company_id":      failure.CompanyID,
		"payment_failure": failure,
		"timestamp":       time.Now(),
		"metadata": map[string]interface{}{
			"source":      "event_processor",
			"version":     "1.0",
			"environment": "production",
			"risk_score":  failure.RiskScore,
			"priority":    string(failure.Priority),
		},
	}

	return e.eventBus.Publish(ctx, "payment.failure.processed", processedEvent)
}

// recordProcessingError records a processing error for monitoring
func (e *EventProcessorService) recordProcessingError(failure *architecture.PaymentFailure, err error) {
	processingError := ProcessingError{
		EventID:      failure.ID.String(),
		Provider:     failure.ProviderID,
		ErrorType:    "processing_failure",
		ErrorMessage: err.Error(),
		Retryable:    true, // Most processing errors are retryable
		RetryCount:   0,
		Timestamp:    time.Now(),
	}

	e.metrics.ProcessingErrors = append(e.metrics.ProcessingErrors, processingError)
}

// calculateAverageProcessingTime calculates the running average processing time
func (e *EventProcessorService) calculateAverageProcessingTime(newTime time.Duration) time.Duration {
	if e.metrics.SuccessfullyProcessed == 0 {
		return newTime
	}

	totalTime := e.metrics.AverageProcessingTime * time.Duration(e.metrics.SuccessfullyProcessed-1)
	totalTime += newTime
	return totalTime / time.Duration(e.metrics.SuccessfullyProcessed)
}

// storeProcessingEvent stores the processing event in the database
func (e *EventProcessorService) storeProcessingEvent(ctx context.Context, failure *architecture.PaymentFailure, processingTime time.Duration) error {
	if e.db == nil {
		e.logger.Warn("Database not available, skipping event storage")
		return nil
	}

	// Create processing event record
	processingEvent := &ProcessingEventRecord{
		ID:               uuid.New(),
		PaymentFailureID: failure.ID,
		CompanyID:        failure.CompanyID,
		ProviderID:       failure.ProviderID,
		ProcessingTime:   processingTime,
		RiskScore:        failure.RiskScore,
		Priority:         string(failure.Priority),
		Status:           string(failure.Status),
		ProcessedAt:      time.Now(),
		CreatedAt:        time.Now(),
	}

	if err := e.db.Create(processingEvent).Error; err != nil {
		return fmt.Errorf("failed to store processing event: %w", err)
	}

	e.logger.Debug("Processing event stored in database",
		zap.String("event_id", processingEvent.ID.String()),
		zap.String("payment_failure_id", failure.ID.String()))

	return nil
}

// ProcessingEventRecord represents a processing event stored in the database
type ProcessingEventRecord struct {
	ID               uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentFailureID uuid.UUID     `json:"payment_failure_id" gorm:"type:uuid;not null;index"`
	CompanyID        string        `json:"company_id" gorm:"not null;index"`
	ProviderID       string        `json:"provider_id" gorm:"not null;index"`
	ProcessingTime   time.Duration `json:"processing_time" gorm:"not null"`
	RiskScore        float64       `json:"risk_score" gorm:"not null"`
	Priority         string        `json:"priority" gorm:"not null"`
	Status           string        `json:"status" gorm:"not null"`
	ProcessedAt      time.Time     `json:"processed_at" gorm:"not null"`
	CreatedAt        time.Time     `json:"created_at" gorm:"not null"`
}

// TableName returns the table name for ProcessingEventRecord
func (ProcessingEventRecord) TableName() string {
	return "processing_events"
}

// GetMetrics returns the current processing metrics
func (e *EventProcessorService) GetMetrics() *EventProcessorMetrics {
	return e.metrics
}

// StartEventProcessing starts the event processing service
func (e *EventProcessorService) StartEventProcessing(ctx context.Context) error {
	e.logger.Info("Starting event processing service")

	// Subscribe to payment failure events
	_, err := e.eventBus.Subscribe(ctx, "payment.failure.detected", e.handlePaymentFailureEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to payment failure events: %w", err)
	}

	e.logger.Info("Event processing service started successfully")
	return nil
}

// handlePaymentFailureEvent is the event handler that matches the EventHandler signature
func (e *EventProcessorService) handlePaymentFailureEvent(ctx context.Context, event interface{}) error {
	// Convert interface{} to map[string]interface{}
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event type: expected map[string]interface{}, got %T", event)
	}

	return e.ProcessPaymentFailureEvent(ctx, eventMap)
}

// StopEventProcessing stops the event processing service
func (e *EventProcessorService) StopEventProcessing(ctx context.Context) error {
	e.logger.Info("Stopping event processing service")

	// TODO: Implement graceful shutdown
	// This would typically unsubscribe from events and wait for in-flight processing to complete

	e.logger.Info("Event processing service stopped")
	return nil
}

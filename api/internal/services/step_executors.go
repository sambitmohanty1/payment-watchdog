package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
)

// PaymentRetryExecutor handles payment retry steps
type PaymentRetryExecutor struct {
	service *RecoveryOrchestrationService
	tracer  trace.Tracer
}

func (e *PaymentRetryExecutor) GetStepType() string {
	return "retry_payment"
}

func (e *PaymentRetryExecutor) Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error) {
	if e.tracer == nil {
		e.tracer = otel.Tracer("payment-retry-executor")
	}
	
	ctx, span := e.tracer.Start(ctx, "execute_payment_retry")
	defer span.End()

	// Parse step configuration
	var config struct {
		Provider     string `json:"provider"`
		MaxRetries   int    `json:"max_retries"`
		RetryDelay   int    `json:"retry_delay_minutes"`
		UpdateAmount bool   `json:"update_amount"`
		NewAmount    *float64 `json:"new_amount,omitempty"`
	}

	if err := json.Unmarshal(step.Config, &config); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse retry config: %w", err)
	}

	// Get payment failure from execution context
	paymentFailure, ok := execution.Context["payment_failure"].(*models.PaymentFailureEvent)
	if !ok {
		err := fmt.Errorf("payment failure not found in execution context")
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("provider", config.Provider),
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.Float64("original_amount", paymentFailure.Amount),
	)

	// Submit retry job to retry service
	retryData := map[string]interface{}{
		"payment_failure_id": paymentFailure.ID.String(),
		"provider":          config.Provider,
		"original_amount":   paymentFailure.Amount,
		"retry_reason":      "workflow_retry",
		"workflow_execution_id": execution.ID.String(),
		"step_id":          step.ID.String(),
	}

	if config.UpdateAmount && config.NewAmount != nil {
		retryData["new_amount"] = *config.NewAmount
		span.SetAttributes(attribute.Float64("new_amount", *config.NewAmount))
	}

	// Submit to retry service
	retryJob, err := e.service.retryService.SubmitJob(ctx, "payment_retry", execution.CompanyID.String(), retryData)
	if err != nil {
		span.RecordError(err)
		return &StepResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to submit retry job: %v", err),
			ShouldRetry:  true,
		}, nil
	}

	// Create recovery action record
	recoveryAction := &models.RecoveryAction{
		CompanyID:           execution.CompanyID,
		PaymentFailureID:    paymentFailure.ID,
		WorkflowExecutionID: &execution.ID,
		ActionType:          "payment_retry",
		ActionData:          step.Config,
		Status:              "pending",
		Provider:            config.Provider,
		ExternalID:          retryJob.ID,
		ScheduledAt:         &time.Time{},
	}
	*recoveryAction.ScheduledAt = time.Now()

	if err := e.service.db.WithContext(ctx).Create(recoveryAction).Error; err != nil {
		// Log error but don't fail the step
		span.RecordError(err)
	}

	return &StepResult{
		Success:    true,
		ExternalID: retryJob.ID,
		Data: map[string]interface{}{
			"retry_job_id":     retryJob.ID,
			"provider":         config.Provider,
			"scheduled_at":     time.Now(),
			"recovery_action_id": recoveryAction.ID.String(),
		},
	}, nil
}

// EmailExecutor handles email communication steps
type EmailExecutor struct {
	service *RecoveryOrchestrationService
	tracer  trace.Tracer
}

func (e *EmailExecutor) GetStepType() string {
	return "send_email"
}

func (e *EmailExecutor) Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error) {
	if e.tracer == nil {
		e.tracer = otel.Tracer("email-executor")
	}
	
	ctx, span := e.tracer.Start(ctx, "execute_send_email")
	defer span.End()

	// Parse step configuration
	var config struct {
		TemplateID   string            `json:"template_id"`
		TemplateName string            `json:"template_name"`
		ToEmail      string            `json:"to_email"`
		Subject      string            `json:"subject"`
		Variables    map[string]string `json:"variables"`
	}

	if err := json.Unmarshal(step.Config, &config); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse email config: %w", err)
	}

	// Get payment failure from execution context
	paymentFailure, ok := execution.Context["payment_failure"].(*models.PaymentFailureEvent)
	if !ok {
		err := fmt.Errorf("payment failure not found in execution context")
		span.RecordError(err)
		return nil, err
	}

	// Determine recipient email
	recipientEmail := config.ToEmail
	if recipientEmail == "" {
		recipientEmail = paymentFailure.CustomerEmail
	}

	span.SetAttributes(
		attribute.String("template_id", config.TemplateID),
		attribute.String("template_name", config.TemplateName),
		attribute.String("recipient_email", recipientEmail),
	)

	// Prepare template variables
	templateVars := make(map[string]interface{})
	for k, v := range config.Variables {
		templateVars[k] = v
	}

	// Add default variables from payment failure
	templateVars["customer_name"] = paymentFailure.CustomerName
	templateVars["customer_email"] = paymentFailure.CustomerEmail
	templateVars["amount"] = paymentFailure.Amount
	templateVars["currency"] = paymentFailure.Currency
	templateVars["failure_reason"] = paymentFailure.FailureReason
	templateVars["transaction_id"] = paymentFailure.TransactionID

	// Send email through communication service
	emailResult, err := e.service.communicationService.SendEmail(ctx, &CommunicationRequest{
		CompanyID:    execution.CompanyID,
		TemplateID:   config.TemplateID,
		TemplateName: config.TemplateName,
		Recipient:    recipientEmail,
		Subject:      config.Subject,
		Variables:    templateVars,
		Context: map[string]interface{}{
			"payment_failure_id":    paymentFailure.ID.String(),
			"workflow_execution_id": execution.ID.String(),
			"step_id":              step.ID.String(),
		},
	})

	if err != nil {
		span.RecordError(err)
		return &StepResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to send email: %v", err),
			ShouldRetry:  true,
		}, nil
	}

	// Create recovery action record
	recoveryAction := &models.RecoveryAction{
		CompanyID:           execution.CompanyID,
		PaymentFailureID:    paymentFailure.ID,
		WorkflowExecutionID: &execution.ID,
		ActionType:          "email_sent",
		ActionData:          step.Config,
		Status:              "completed",
		Provider:            "email",
		ExternalID:          emailResult.MessageID,
		ExecutedAt:          &time.Time{},
		CompletedAt:         &time.Time{},
	}
	now := time.Now()
	*recoveryAction.ExecutedAt = now
	*recoveryAction.CompletedAt = now

	if err := e.service.db.WithContext(ctx).Create(recoveryAction).Error; err != nil {
		span.RecordError(err)
	}

	return &StepResult{
		Success:    true,
		ExternalID: emailResult.MessageID,
		Data: map[string]interface{}{
			"message_id":         emailResult.MessageID,
			"recipient":          recipientEmail,
			"template_used":      emailResult.TemplateUsed,
			"sent_at":           now,
			"recovery_action_id": recoveryAction.ID.String(),
		},
	}, nil
}

// SMSExecutor handles SMS communication steps
type SMSExecutor struct {
	service *RecoveryOrchestrationService
	tracer  trace.Tracer
}

func (e *SMSExecutor) GetStepType() string {
	return "send_sms"
}

func (e *SMSExecutor) Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error) {
	if e.tracer == nil {
		e.tracer = otel.Tracer("sms-executor")
	}
	
	ctx, span := e.tracer.Start(ctx, "execute_send_sms")
	defer span.End()

	// Parse step configuration
	var config struct {
		TemplateID   string            `json:"template_id"`
		TemplateName string            `json:"template_name"`
		ToPhone      string            `json:"to_phone"`
		Message      string            `json:"message"`
		Variables    map[string]string `json:"variables"`
	}

	if err := json.Unmarshal(step.Config, &config); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse SMS config: %w", err)
	}

	// Get payment failure from execution context
	paymentFailure, ok := execution.Context["payment_failure"].(*models.PaymentFailureEvent)
	if !ok {
		err := fmt.Errorf("payment failure not found in execution context")
		span.RecordError(err)
		return nil, err
	}

	// Determine recipient phone
	recipientPhone := config.ToPhone
	if recipientPhone == "" {
		recipientPhone = paymentFailure.CustomerPhone
	}

	if recipientPhone == "" {
		return &StepResult{
			Success:      false,
			ErrorMessage: "No phone number available for SMS",
			ShouldRetry:  false,
		}, nil
	}

	span.SetAttributes(
		attribute.String("template_id", config.TemplateID),
		attribute.String("template_name", config.TemplateName),
		attribute.String("recipient_phone", recipientPhone),
	)

	// Prepare template variables
	templateVars := make(map[string]interface{})
	for k, v := range config.Variables {
		templateVars[k] = v
	}

	// Add default variables from payment failure
	templateVars["customer_name"] = paymentFailure.CustomerName
	templateVars["amount"] = paymentFailure.Amount
	templateVars["currency"] = paymentFailure.Currency

	// Send SMS through communication service
	smsResult, err := e.service.communicationService.SendSMS(ctx, &CommunicationRequest{
		CompanyID:    execution.CompanyID,
		TemplateID:   config.TemplateID,
		TemplateName: config.TemplateName,
		Recipient:    recipientPhone,
		Message:      config.Message,
		Variables:    templateVars,
		Context: map[string]interface{}{
			"payment_failure_id":    paymentFailure.ID.String(),
			"workflow_execution_id": execution.ID.String(),
			"step_id":              step.ID.String(),
		},
	})

	if err != nil {
		span.RecordError(err)
		return &StepResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to send SMS: %v", err),
			ShouldRetry:  true,
		}, nil
	}

	// Create recovery action record
	recoveryAction := &models.RecoveryAction{
		CompanyID:           execution.CompanyID,
		PaymentFailureID:    paymentFailure.ID,
		WorkflowExecutionID: &execution.ID,
		ActionType:          "sms_sent",
		ActionData:          step.Config,
		Status:              "completed",
		Provider:            "sms",
		ExternalID:          smsResult.MessageID,
		ExecutedAt:          &time.Time{},
		CompletedAt:         &time.Time{},
	}
	now := time.Now()
	*recoveryAction.ExecutedAt = now
	*recoveryAction.CompletedAt = now

	if err := e.service.db.WithContext(ctx).Create(recoveryAction).Error; err != nil {
		span.RecordError(err)
	}

	return &StepResult{
		Success:    true,
		ExternalID: smsResult.MessageID,
		Data: map[string]interface{}{
			"message_id":         smsResult.MessageID,
			"recipient":          recipientPhone,
			"template_used":      smsResult.TemplateUsed,
			"sent_at":           now,
			"recovery_action_id": recoveryAction.ID.String(),
		},
	}, nil
}

// WaitExecutor handles wait/delay steps
type WaitExecutor struct {
	service *RecoveryOrchestrationService
	tracer  trace.Tracer
}

func (e *WaitExecutor) GetStepType() string {
	return "wait"
}

func (e *WaitExecutor) Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error) {
	if e.tracer == nil {
		e.tracer = otel.Tracer("wait-executor")
	}
	
	ctx, span := e.tracer.Start(ctx, "execute_wait")
	defer span.End()

	// Parse step configuration
	var config struct {
		WaitMinutes int    `json:"wait_minutes"`
		WaitHours   int    `json:"wait_hours"`
		WaitDays    int    `json:"wait_days"`
		Reason      string `json:"reason"`
	}

	if err := json.Unmarshal(step.Config, &config); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse wait config: %w", err)
	}

	// Calculate total wait duration
	totalMinutes := config.WaitMinutes + (config.WaitHours * 60) + (config.WaitDays * 24 * 60)
	if totalMinutes <= 0 {
		totalMinutes = 1 // Minimum 1 minute wait
	}

	waitDuration := time.Duration(totalMinutes) * time.Minute

	span.SetAttributes(
		attribute.Int("wait_minutes", totalMinutes),
		attribute.String("reason", config.Reason),
	)

	// Perform the wait
	select {
	case <-time.After(waitDuration):
		// Wait completed successfully
		return &StepResult{
			Success: true,
			Data: map[string]interface{}{
				"wait_duration_minutes": totalMinutes,
				"reason":               config.Reason,
				"completed_at":         time.Now(),
			},
		}, nil
	case <-ctx.Done():
		// Context cancelled
		return &StepResult{
			Success:      false,
			ErrorMessage: "Wait cancelled due to context cancellation",
			ShouldRetry:  false,
		}, nil
	}
}

// ConditionalExecutor handles conditional logic steps
type ConditionalExecutor struct {
	service *RecoveryOrchestrationService
	tracer  trace.Tracer
}

func (e *ConditionalExecutor) GetStepType() string {
	return "conditional"
}

func (e *ConditionalExecutor) Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error) {
	if e.tracer == nil {
		e.tracer = otel.Tracer("conditional-executor")
	}
	
	ctx, span := e.tracer.Start(ctx, "execute_conditional")
	defer span.End()

	// Parse step configuration
	var config struct {
		Conditions []struct {
			Field    string      `json:"field"`
			Operator string      `json:"operator"`
			Value    interface{} `json:"value"`
		} `json:"conditions"`
		Logic      string `json:"logic"` // "AND" or "OR"
		OnTrue     string `json:"on_true"`   // Action if condition is true
		OnFalse    string `json:"on_false"`  // Action if condition is false
		SkipSteps  int    `json:"skip_steps"` // Number of steps to skip
	}

	if err := json.Unmarshal(step.Config, &config); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse conditional config: %w", err)
	}

	// Get payment failure from execution context
	paymentFailure, ok := execution.Context["payment_failure"].(*models.PaymentFailureEvent)
	if !ok {
		err := fmt.Errorf("payment failure not found in execution context")
		span.RecordError(err)
		return nil, err
	}

	// Evaluate conditions
	conditionResults := make([]bool, len(config.Conditions))
	for i, condition := range config.Conditions {
		conditionResults[i] = e.evaluateCondition(paymentFailure, condition.Field, condition.Operator, condition.Value)
	}

	// Apply logic
	var finalResult bool
	if config.Logic == "OR" {
		finalResult = false
		for _, result := range conditionResults {
			if result {
				finalResult = true
				break
			}
		}
	} else { // Default to AND
		finalResult = true
		for _, result := range conditionResults {
			if !result {
				finalResult = false
				break
			}
		}
	}

	span.SetAttributes(
		attribute.Bool("condition_result", finalResult),
		attribute.String("logic", config.Logic),
		attribute.Int("conditions_count", len(config.Conditions)),
	)

	// Determine action based on result
	var action string
	if finalResult {
		action = config.OnTrue
	} else {
		action = config.OnFalse
	}

	resultData := map[string]interface{}{
		"condition_result":    finalResult,
		"action_taken":       action,
		"conditions_evaluated": len(config.Conditions),
		"logic_used":         config.Logic,
	}

	// Handle specific actions
	switch action {
	case "continue":
		// Continue to next step normally
		return &StepResult{
			Success: true,
			Data:    resultData,
		}, nil
	case "skip":
		// Skip specified number of steps
		resultData["steps_to_skip"] = config.SkipSteps
		return &StepResult{
			Success: true,
			Data:    resultData,
		}, nil
	case "stop":
		// Stop workflow execution
		return &StepResult{
			Success:      false,
			ErrorMessage: "Workflow stopped due to conditional logic",
			Data:         resultData,
		}, nil
	default:
		// Unknown action, continue normally
		return &StepResult{
			Success: true,
			Data:    resultData,
		}, nil
	}
}

func (e *ConditionalExecutor) evaluateCondition(paymentFailure *models.PaymentFailureEvent, field, operator string, value interface{}) bool {
	var fieldValue interface{}

	// Extract field value from payment failure
	switch field {
	case "amount":
		fieldValue = paymentFailure.Amount
	case "currency":
		fieldValue = paymentFailure.Currency
	case "failure_reason":
		fieldValue = paymentFailure.FailureReason
	case "provider":
		fieldValue = paymentFailure.Provider
	case "customer_email":
		fieldValue = paymentFailure.CustomerEmail
	case "retry_count":
		fieldValue = paymentFailure.RetryCount
	case "days_overdue":
		if paymentFailure.DueDate != nil {
			fieldValue = int(time.Since(*paymentFailure.DueDate).Hours() / 24)
		}
	default:
		return false
	}

	// Evaluate condition based on operator
	switch operator {
	case "eq", "equals":
		return fieldValue == value
	case "ne", "not_equals":
		return fieldValue != value
	case "gt", "greater_than":
		return compareNumbers(fieldValue, value) > 0
	case "gte", "greater_than_or_equal":
		return compareNumbers(fieldValue, value) >= 0
	case "lt", "less_than":
		return compareNumbers(fieldValue, value) < 0
	case "lte", "less_than_or_equal":
		return compareNumbers(fieldValue, value) <= 0
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := value.(string); ok {
				return contains(str, substr)
			}
		}
	case "in":
		if values, ok := value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return true
				}
			}
		}
	}

	return false
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/lexure-intelligence/payment-watchdog/internal/models"
)

// RecoveryOrchestrationService manages automated recovery workflows
type RecoveryOrchestrationService struct {
	db                    *gorm.DB
	retryService          *RetryService
	communicationService  *CommunicationService
	analyticsService      *AnalyticsService
	stepExecutors         map[string]StepExecutor
	tracer                trace.Tracer
	logger                *zap.Logger
	mu                    sync.RWMutex
	activeExecutions      map[uuid.UUID]*WorkflowExecution
	executionWorkers      int
	workerPool            chan struct{}
}

// WorkflowExecution represents an active workflow execution
type WorkflowExecution struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	PaymentFailureID  uuid.UUID
	CompanyID         uuid.UUID
	Status            string
	CurrentStepIndex  int
	Context           map[string]interface{}
	StartedAt         time.Time
	CancelFunc        context.CancelFunc
	mu                sync.RWMutex
}

// StepExecutor interface for different types of workflow steps
type StepExecutor interface {
	Execute(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) (*StepResult, error)
	GetStepType() string
}

// StepResult represents the result of a step execution
type StepResult struct {
	Success      bool                   `json:"success"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	NextDelay    time.Duration          `json:"next_delay,omitempty"`
	ShouldRetry  bool                   `json:"should_retry,omitempty"`
	ExternalID   string                 `json:"external_id,omitempty"`
}

// TriggerCondition represents conditions for workflow triggering
type TriggerCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// WorkflowTriggerConditions represents the complete trigger conditions
type WorkflowTriggerConditions struct {
	Conditions []TriggerCondition `json:"conditions"`
	Logic      string             `json:"logic"` // "AND" or "OR"
}

// NewRecoveryOrchestrationService creates a new recovery orchestration service
func NewRecoveryOrchestrationService(
	db *gorm.DB,
	retryService *RetryService,
	communicationService *CommunicationService,
	analyticsService *AnalyticsService,
	logger *zap.Logger,
) *RecoveryOrchestrationService {
	service := &RecoveryOrchestrationService{
		db:                   db,
		retryService:         retryService,
		communicationService: communicationService,
		analyticsService:     analyticsService,
		stepExecutors:        make(map[string]StepExecutor),
		tracer:               otel.Tracer("recovery-orchestration"),
		activeExecutions:     make(map[uuid.UUID]*WorkflowExecution),
		executionWorkers:     10, // Configurable
		workerPool:           make(chan struct{}, 10),
	}

	// Register default step executors
	service.RegisterStepExecutor(&PaymentRetryExecutor{service: service})
	service.RegisterStepExecutor(&EmailExecutor{service: service})
	service.RegisterStepExecutor(&SMSExecutor{service: service})
	service.RegisterStepExecutor(&WaitExecutor{service: service})

	return service
}

// RegisterStepExecutor registers a step executor for a specific step type
func (r *RecoveryOrchestrationService) RegisterStepExecutor(executor StepExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stepExecutors[executor.GetStepType()] = executor
}

// TriggerWorkflowsForFailure triggers appropriate workflows for a payment failure
func (r *RecoveryOrchestrationService) TriggerWorkflowsForFailure(ctx context.Context, paymentFailure *models.PaymentFailureEvent) error {
	ctx, span := r.tracer.Start(ctx, "trigger_workflows_for_failure")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.String("company_id", paymentFailure.CompanyID),
	)

	// Get active workflows for the company
	var workflows []models.RecoveryWorkflow
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", paymentFailure.CompanyID, true).
		Order("priority DESC").
		Preload("Steps", "is_active = ?", true).
		Find(&workflows).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get workflows: %w", err)
	}

	// Find matching workflows based on trigger conditions
	var matchingWorkflows []models.RecoveryWorkflow
	for _, workflow := range workflows {
		if r.evaluateTriggerConditions(paymentFailure, workflow.TriggerConditions) {
			matchingWorkflows = append(matchingWorkflows, workflow)
		}
	}

	span.SetAttributes(attribute.Int("matching_workflows", len(matchingWorkflows)))

	// Execute matching workflows (highest priority first)
	for _, workflow := range matchingWorkflows {
		if err := r.StartWorkflowExecution(ctx, &workflow, paymentFailure); err != nil {
			log.Printf("Failed to start workflow %s: %v", workflow.ID, err)
			continue
		}
	}

	return nil
}

// StartWorkflowExecution starts a new workflow execution
func (r *RecoveryOrchestrationService) StartWorkflowExecution(ctx context.Context, workflow *models.RecoveryWorkflow, paymentFailure *models.PaymentFailureEvent) error {
	ctx, span := r.tracer.Start(ctx, "start_workflow_execution")
	defer span.End()

	// Create workflow execution record
	execution := &models.RecoveryWorkflowExecution{
		ID:               uuid.New(),
		WorkflowID:       workflow.ID,
		PaymentFailureID: paymentFailure.ID,
		CompanyID:        workflow.CompanyID,
		Status:           "pending",
		TotalSteps:       len(workflow.Steps),
		StartedAt:        time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(execution).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create workflow execution: %w", err)
	}

	// Create active execution context
	execCtx, cancel := context.WithCancel(ctx)
	activeExecution := &WorkflowExecution{
		ID:               execution.ID,
		WorkflowID:       workflow.ID,
		PaymentFailureID: paymentFailure.ID,
		CompanyID:        workflow.CompanyID,
		Status:           "running",
		CurrentStepIndex: 0,
		Context: map[string]interface{}{
			"payment_failure": paymentFailure,
			"workflow":        workflow,
		},
		StartedAt:  time.Now(),
		CancelFunc: cancel,
	}

	// Store active execution
	r.mu.Lock()
	r.activeExecutions[execution.ID] = activeExecution
	r.mu.Unlock()

	// Start execution in goroutine
	go r.executeWorkflow(execCtx, activeExecution, workflow)

	span.SetAttributes(
		attribute.String("execution_id", execution.ID.String()),
		attribute.String("workflow_id", workflow.ID.String()),
	)

	return nil
}

// executeWorkflow executes a workflow asynchronously
func (r *RecoveryOrchestrationService) executeWorkflow(ctx context.Context, execution *WorkflowExecution, workflow *models.RecoveryWorkflow) {
	// Acquire a worker from the pool
	r.workerPool <- struct{}{}
	defer func() { <-r.workerPool }()

	// Set up logging and tracing
	logger := r.logger.With(
		zap.String("workflow_id", execution.WorkflowID.String()),
		zap.String("execution_id", execution.ID.String()),
		zap.String("company_id", execution.CompanyID.String()),
	)

	// Update execution status to running
	if err := r.updateExecutionStatus(ctx, execution.ID, "running"); err != nil {
		logger.Error("Failed to update execution status to running", zap.Error(err))
		return
	}

	// Ensure the execution is cleaned up when done
	defer func() {
		r.mu.Lock()
		delete(r.activeExecutions, execution.ID)
		r.mu.Unlock()

		// Update execution status to completed
		status := "completed"
		if execution.Status == "failed" {
			status = "failed"
		}

		if err := r.updateExecutionStatus(ctx, execution.ID, status); err != nil {
			logger.Error("Failed to update execution status", 
				zap.String("status", status), 
				zap.Error(err))
		}

		// Update completed at timestamp
		if err := r.updateExecutionCompletedAt(ctx, execution.ID, time.Now()); err != nil {
			logger.Error("Failed to update execution completed at", zap.Error(err))
		}

		// Log workflow completion
		duration := time.Since(execution.StartedAt)
		if execution.Status == "completed" {
			logger.Info("Workflow execution completed successfully", 
				zap.Duration("duration", duration),
				zap.Int("steps_completed", execution.CurrentStepIndex))
		} else if execution.Status == "failed" {
			logger.Error("Workflow execution failed", 
				zap.Duration("duration", duration),
				zap.Int("steps_completed", execution.CurrentStepIndex))
		}
	}()

	// Execute each step in sequence
	for i := execution.CurrentStepIndex; i < len(workflow.Steps); i++ {
		step := workflow.Steps[i]
		
		// Update current step
		execution.CurrentStepIndex = i
		if err := r.updateCurrentStep(ctx, execution.ID, &step.ID); err != nil {
			logger.Error("Failed to update current step", 
				zap.String("step_id", step.ID.String()),
				zap.Error(err))
			execution.Status = "failed"
			return
		}

		// Execute the step
		if err := r.executeStep(ctx, execution, &step); err != nil {
			logger.Error("Step execution failed", 
				zap.String("step_id", step.ID.String()),
				zap.String("step_type", step.StepType),
				zap.Error(err))

			// Check if the step is marked as critical
			if step.IsCritical {
				execution.Status = "failed"
				return
			}
			// For non-critical steps, continue to the next step
			continue
		}

		// Update successful step count
		if err := r.incrementExecutionCounter(ctx, execution.ID, "steps_completed"); err != nil {
			logger.Warn("Failed to increment steps_completed counter", zap.Error(err))
		}
	}

	// If we've reached here, all steps completed successfully
	execution.Status = "completed"
	ctx, span := r.tracer.Start(ctx, "execute_workflow")
	defer span.End()

	// Acquire worker slot
	r.workerPool <- struct{}{}
	defer func() { <-r.workerPool }()

	defer func() {
		// Clean up active execution
		r.mu.Lock()
		delete(r.activeExecutions, execution.ID)
		r.mu.Unlock()
	}()

	// Update execution status
	if err := r.updateExecutionStatus(ctx, execution.ID, "running"); err != nil {
		log.Printf("Failed to update execution status: %v", err)
		return
	}

	// Execute workflow steps
	for i, step := range workflow.Steps {
		if !step.IsActive {
			continue
		}

		execution.mu.Lock()
		execution.CurrentStepIndex = i
		execution.mu.Unlock()

		// Update current step in database
		if err := r.updateCurrentStep(ctx, execution.ID, &step.ID); err != nil {
			log.Printf("Failed to update current step: %v", err)
		}

		// Execute step with delay if specified
		if step.DelayMinutes > 0 {
			select {
			case <-time.After(time.Duration(step.DelayMinutes) * time.Minute):
			case <-ctx.Done():
				r.updateExecutionStatus(ctx, execution.ID, "cancelled")
				return
			}
		}

		// Execute the step
		if err := r.executeStep(ctx, execution, &step); err != nil {
			log.Printf("Step execution failed: %v", err)
			r.updateExecutionStatus(ctx, execution.ID, "failed")
			return
		}

		// Check if execution was cancelled
		select {
		case <-ctx.Done():
			r.updateExecutionStatus(ctx, execution.ID, "cancelled")
			return
		default:
		}
	}

	// Mark execution as completed
	r.updateExecutionStatus(ctx, execution.ID, "completed")
	r.updateExecutionCompletedAt(ctx, execution.ID, time.Now())

	span.SetAttributes(
		attribute.String("execution_id", execution.ID.String()),
		attribute.String("final_status", "completed"),
	)
}

// executeStep executes a single workflow step
func (r *RecoveryOrchestrationService) executeStep(ctx context.Context, execution *WorkflowExecution, step *models.RecoveryWorkflowStep) error {
	ctx, span := r.tracer.Start(ctx, "execute_step")
	defer span.End()

	span.SetAttributes(
		attribute.String("step_id", step.ID.String()),
		attribute.String("step_type", step.StepType),
		attribute.String("step_name", step.StepName),
	)

	// Create step execution record
	stepExecution := &models.RecoveryStepExecution{
		ID:                  uuid.New(),
		WorkflowExecutionID: execution.ID,
		StepID:              step.ID,
		Status:              "running",
		StartedAt:           time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(stepExecution).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create step execution: %w", err)
	}

	// Get step executor
	r.mu.RLock()
	executor, exists := r.stepExecutors[step.StepType]
	r.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no executor found for step type: %s", step.StepType)
		r.updateStepExecutionStatus(ctx, stepExecution.ID, "failed", err.Error())
		span.RecordError(err)
		return err
	}

	// Execute the step
	startTime := time.Now()
	result, err := executor.Execute(ctx, execution, step)
	duration := time.Since(startTime)

	// Update step execution with results
	updateData := map[string]interface{}{
		"completed_at": time.Now(),
		"duration_ms":  duration.Milliseconds(),
	}

	if err != nil {
		updateData["status"] = "failed"
		updateData["error_message"] = err.Error()
		span.RecordError(err)
	} else {
		updateData["status"] = "completed"
		if result != nil {
			if resultJSON, jsonErr := json.Marshal(result); jsonErr == nil {
				updateData["result"] = resultJSON
			}
			if result.ExternalID != "" {
				updateData["external_id"] = result.ExternalID
			}
		}
	}

	if updateErr := r.db.WithContext(ctx).Model(stepExecution).Updates(updateData).Error; updateErr != nil {
		log.Printf("Failed to update step execution: %v", updateErr)
	}

	// Update execution counters
	if err != nil {
		r.incrementExecutionCounter(ctx, execution.ID, "failed_steps")
	} else {
		r.incrementExecutionCounter(ctx, execution.ID, "successful_steps")
	}
	r.incrementExecutionCounter(ctx, execution.ID, "completed_steps")

	return err
}

// evaluateTriggerConditions evaluates if a payment failure matches workflow trigger conditions
func (r *RecoveryOrchestrationService) evaluateTriggerConditions(paymentFailure *models.PaymentFailureEvent, conditionsJSON []byte) bool {
	if len(conditionsJSON) == 0 {
		return true // No conditions means always trigger
	}

	var conditions WorkflowTriggerConditions
	if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
		log.Printf("Failed to unmarshal trigger conditions: %v", err)
		return false
	}

	results := make([]bool, len(conditions.Conditions))
	for i, condition := range conditions.Conditions {
		results[i] = r.evaluateCondition(paymentFailure, condition)
	}

	// Apply logic (AND/OR)
	if conditions.Logic == "OR" {
		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	} else { // Default to AND
		for _, result := range results {
			if !result {
				return false
			}
		}
		return true
	}
}

// evaluateCondition evaluates a single trigger condition
func (r *RecoveryOrchestrationService) evaluateCondition(paymentFailure *models.PaymentFailureEvent, condition TriggerCondition) bool {
	var fieldValue interface{}

	// Extract field value from payment failure
	switch condition.Field {
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
	case "days_overdue":
		if paymentFailure.DueDate != nil {
			fieldValue = int(time.Since(*paymentFailure.DueDate).Hours() / 24)
		}
	default:
		return false
	}

	// Evaluate condition based on operator
	switch condition.Operator {
	case "eq", "equals":
		return fieldValue == condition.Value
	case "ne", "not_equals":
		return fieldValue != condition.Value
	case "gt", "greater_than":
		return compareNumbers(fieldValue, condition.Value) > 0
	case "gte", "greater_than_or_equal":
		return compareNumbers(fieldValue, condition.Value) >= 0
	case "lt", "less_than":
		return compareNumbers(fieldValue, condition.Value) < 0
	case "lte", "less_than_or_equal":
		return compareNumbers(fieldValue, condition.Value) <= 0
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := condition.Value.(string); ok {
				return contains(str, substr)
			}
		}
	case "in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return true
				}
			}
		}
	}

	return false
}

// Helper functions
func compareNumbers(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 || 
		(len(substr) > 0 && findSubstring(str, substr)))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Database helper methods
func (r *RecoveryOrchestrationService) updateExecutionStatus(ctx context.Context, executionID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("id = ?", executionID).
		Update("status", status).Error
}

func (r *RecoveryOrchestrationService) updateCurrentStep(ctx context.Context, executionID uuid.UUID, stepID *uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("id = ?", executionID).
		Update("current_step_id", stepID).Error
}

func (r *RecoveryOrchestrationService) updateExecutionCompletedAt(ctx context.Context, executionID uuid.UUID, completedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("id = ?", executionID).
		Update("completed_at", completedAt).Error
}

func (r *RecoveryOrchestrationService) incrementExecutionCounter(ctx context.Context, executionID uuid.UUID, field string) error {
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("id = ?", executionID).
		Update(field, gorm.Expr(field+" + ?", 1)).Error
}

func (r *RecoveryOrchestrationService) updateStepExecutionStatus(ctx context.Context, stepExecutionID uuid.UUID, status, errorMessage string) error {
	updates := map[string]interface{}{
		"status":       status,
		"completed_at": time.Now(),
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	return r.db.WithContext(ctx).Model(&models.RecoveryStepExecution{}).
		Where("id = ?", stepExecutionID).
		Updates(updates).Error
}

// Public API methods

// CreateWorkflow creates a new recovery workflow
func (r *RecoveryOrchestrationService) CreateWorkflow(ctx context.Context, workflow *models.RecoveryWorkflow) error {
	ctx, span := r.tracer.Start(ctx, "create_workflow")
	defer span.End()

	workflow.ID = uuid.New()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	// Set step workflow IDs
	for i := range workflow.Steps {
		workflow.Steps[i].ID = uuid.New()
		workflow.Steps[i].WorkflowID = workflow.ID
		workflow.Steps[i].CreatedAt = time.Now()
		workflow.Steps[i].UpdatedAt = time.Now()
	}

	if err := r.db.WithContext(ctx).Create(workflow).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	return nil
}

// GetWorkflows retrieves workflows for a company
func (r *RecoveryOrchestrationService) GetWorkflows(ctx context.Context, companyID uuid.UUID, filters map[string]interface{}, page, limit int) ([]models.RecoveryWorkflow, int64, error) {
	var workflows []models.RecoveryWorkflow
	var total int64

	query := r.db.WithContext(ctx).Model(&models.RecoveryWorkflow{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Steps", "is_active = ?", true).
		Order("priority DESC, created_at DESC").
		Find(&workflows).Error; err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// GetWorkflow retrieves a specific workflow
func (r *RecoveryOrchestrationService) GetWorkflow(ctx context.Context, workflowID, companyID uuid.UUID) (*models.RecoveryWorkflow, error) {
	var workflow models.RecoveryWorkflow
	if err := r.db.WithContext(ctx).
		Where("id = ? AND company_id = ?", workflowID, companyID).
		Preload("Steps", "is_active = ?", true).
		First(&workflow).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (r *RecoveryOrchestrationService) UpdateWorkflow(ctx context.Context, workflowID, companyID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflow{}).
		Where("id = ? AND company_id = ?", workflowID, companyID).
		Updates(updates).Error
}

// DeleteWorkflow soft deletes a workflow
func (r *RecoveryOrchestrationService) DeleteWorkflow(ctx context.Context, workflowID, companyID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflow{}).
		Where("id = ? AND company_id = ?", workflowID, companyID).
		Update("is_active", false).Error
}

// TriggerWorkflowManually manually triggers a workflow
func (r *RecoveryOrchestrationService) TriggerWorkflowManually(ctx context.Context, workflowID, paymentFailureID, companyID uuid.UUID) error {
	// Get workflow
	workflow, err := r.GetWorkflow(ctx, workflowID, companyID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Get payment failure
	var paymentFailure models.PaymentFailureEvent
	if err := r.db.WithContext(ctx).First(&paymentFailure, "id = ? AND company_id = ?", paymentFailureID, companyID).Error; err != nil {
		return fmt.Errorf("payment failure not found: %w", err)
	}

	// Start workflow execution
	return r.StartWorkflowExecution(ctx, workflow, &paymentFailure)
}

// GetRecoveryMetrics retrieves recovery performance metrics
func (r *RecoveryOrchestrationService) GetRecoveryMetrics(ctx context.Context, companyID uuid.UUID, timeRange time.Duration) (*models.RecoveryMetrics, error) {
	startTime := time.Now().Add(-timeRange)
	
	var metrics models.RecoveryMetrics
	
	// Get workflow execution metrics
	var totalExecutions, successfulExecutions, failedExecutions int64
	r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("company_id = ? AND created_at >= ?", companyID, startTime).
		Count(&totalExecutions)
	
	r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("company_id = ? AND created_at >= ? AND status = ?", companyID, startTime, "completed").
		Count(&successfulExecutions)
	
	r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("company_id = ? AND created_at >= ? AND status = ?", companyID, startTime, "failed").
		Count(&failedExecutions)

	// Get recovery action metrics
	var totalActions, successfulActions int64
	r.db.WithContext(ctx).Model(&models.RecoveryAction{}).
		Where("company_id = ? AND created_at >= ?", companyID, startTime).
		Count(&totalActions)
	
	r.db.WithContext(ctx).Model(&models.RecoveryAction{}).
		Where("company_id = ? AND created_at >= ? AND status = ?", companyID, startTime, "completed").
		Count(&successfulActions)

	// Calculate metrics
	metrics.CompanyID = companyID
	metrics.PeriodStart = startTime
	metrics.PeriodEnd = time.Now()
	metrics.PeriodType = "custom"
	metrics.TotalWorkflowExecutions = int(totalExecutions)
	metrics.SuccessfulWorkflowExecutions = int(successfulExecutions)
	metrics.FailedWorkflowExecutions = int(failedExecutions)
	metrics.TotalRecoveryActions = int(totalActions)
	metrics.SuccessfulRecoveryActions = int(successfulActions)
	
	if totalActions > 0 {
		metrics.RecoverySuccessRate = float64(successfulActions) / float64(totalActions) * 100
	}

	return &metrics, nil
}

func (r *RecoveryOrchestrationService) GetWorkflowExecutions(ctx context.Context, companyID uuid.UUID, filters map[string]interface{}, page, limit int) ([]models.RecoveryWorkflowExecution, int64, error) {
	var executions []models.RecoveryWorkflowExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if workflowID, ok := filters["workflow_id"].(string); ok && workflowID != "" {
		if id, err := uuid.Parse(workflowID); err == nil {
			query = query.Where("workflow_id = ?", id)
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Workflow").
		Preload("PaymentFailure").
		Preload("StepExecutions").
		Order("created_at DESC").
		Find(&executions).Error; err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

func (r *RecoveryOrchestrationService) PauseWorkflowExecution(ctx context.Context, executionID uuid.UUID) error {
	// Cancel active execution
	r.mu.Lock()
	if activeExecution, exists := r.activeExecutions[executionID]; exists {
		activeExecution.CancelFunc()
		delete(r.activeExecutions, executionID)
	}
	r.mu.Unlock()

	// Update database status
	return r.db.WithContext(ctx).Model(&models.RecoveryWorkflowExecution{}).
		Where("id = ?", executionID).
		Updates(map[string]interface{}{
			"status":    "paused",
			"paused_at": time.Now(),
		}).Error
}

func (r *RecoveryOrchestrationService) ResumeWorkflowExecution(ctx context.Context, executionID uuid.UUID) error {
	// Get execution from database
	var execution models.RecoveryWorkflowExecution
	if err := r.db.WithContext(ctx).
		Preload("Workflow.Steps", "is_active = ?", true).
		Preload("PaymentFailure").
		First(&execution, "id = ?", executionID).Error; err != nil {
		return fmt.Errorf("failed to get workflow execution: %w", err)
	}

	if execution.Status != "paused" {
		return fmt.Errorf("execution is not paused")
	}

	// Update status to running
	if err := r.db.WithContext(ctx).Model(&execution).
		Updates(map[string]interface{}{
			"status":    "running",
			"paused_at": nil,
		}).Error; err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	// Restart execution
	return r.StartWorkflowExecution(ctx, &execution.Workflow, &execution.PaymentFailure)
}

func (r *RecoveryOrchestrationService) CancelWorkflowExecution(ctx context.Context, executionID uuid.UUID) error {
	// Cancel active execution
	r.mu.Lock()
	if activeExecution, exists := r.activeExecutions[executionID]; exists {
		activeExecution.CancelFunc()
		delete(r.activeExecutions, executionID)
	}
	r.mu.Unlock()

	// Update database status
	return r.updateExecutionStatus(ctx, executionID, "cancelled")
}

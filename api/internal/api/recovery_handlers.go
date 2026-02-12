package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/payment-watchdog/internal/models"
	"github.com/payment-watchdog/internal/services"
)

// RecoveryHandlers contains handlers for recovery workflow endpoints
type RecoveryHandlers struct {
	recoveryService      *services.RecoveryOrchestrationService
	communicationService *services.CommunicationService
	tracer               trace.Tracer
}

// NewRecoveryHandlers creates new recovery handlers
func NewRecoveryHandlers(
	recoveryService *services.RecoveryOrchestrationService,
	communicationService *services.CommunicationService,
) *RecoveryHandlers {
	return &RecoveryHandlers{
		recoveryService:      recoveryService,
		communicationService: communicationService,
		tracer:               otel.Tracer("recovery-handlers"),
	}
}

// CreateWorkflow creates a new recovery workflow
func (h *RecoveryHandlers) CreateWorkflow(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_workflow")
	defer span.End()

	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID required"})
		return
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req struct {
		Name              string                 `json:"name" binding:"required"`
		Description       string                 `json:"description"`
		Priority          int                    `json:"priority"`
		TriggerConditions map[string]interface{} `json:"trigger_conditions"`
		Steps             []struct {
			StepOrder    int                    `json:"step_order" binding:"required"`
			StepType     string                 `json:"step_type" binding:"required"`
			StepName     string                 `json:"step_name" binding:"required"`
			Description  string                 `json:"description"`
			Config       map[string]interface{} `json:"config"`
			Conditions   map[string]interface{} `json:"conditions"`
			DelayMinutes int                    `json:"delay_minutes"`
			IsParallel   bool                   `json:"is_parallel"`
		} `json:"steps" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create workflow
	workflow := &models.RecoveryWorkflow{
		CompanyID:   companyUUID,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		IsActive:    true,
		CreatedBy:   c.GetString("user_id"),
	}

	// Convert trigger conditions to JSON
	if req.TriggerConditions != nil {
		conditionsJSON, err := json.Marshal(req.TriggerConditions)
		if err != nil {
			span.RecordError(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger conditions"})
			return
		}
		workflow.TriggerConditions = conditionsJSON
	}

	// Create workflow steps
	for _, stepReq := range req.Steps {
		step := models.RecoveryWorkflowStep{
			StepOrder:    stepReq.StepOrder,
			StepType:     stepReq.StepType,
			StepName:     stepReq.StepName,
			Description:  stepReq.Description,
			DelayMinutes: stepReq.DelayMinutes,
			IsParallel:   stepReq.IsParallel,
			IsActive:     true,
		}

		// Convert config to JSON
		if stepReq.Config != nil {
			configJSON, err := json.Marshal(stepReq.Config)
			if err != nil {
				span.RecordError(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step config"})
				return
			}
			step.Config = configJSON
		}

		// Convert conditions to JSON
		if stepReq.Conditions != nil {
			conditionsJSON, err := json.Marshal(stepReq.Conditions)
			if err != nil {
				span.RecordError(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step conditions"})
				return
			}
			step.Conditions = conditionsJSON
		}

		workflow.Steps = append(workflow.Steps, step)
	}

	// Save to database
	if err := h.recoveryService.CreateWorkflow(ctx, workflow); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	span.SetAttributes(
		attribute.String("workflow_id", workflow.ID.String()),
		attribute.String("workflow_name", workflow.Name),
		attribute.Int("steps_count", len(workflow.Steps)),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Workflow created successfully",
		"workflow": workflow,
	})
}

// GetWorkflows retrieves recovery workflows for a company
func (h *RecoveryHandlers) GetWorkflows(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_workflows")
	defer span.End()

	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID required"})
		return
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filters := make(map[string]interface{})
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}

	workflows, total, err := h.recoveryService.GetWorkflows(ctx, companyUUID, filters, page, limit)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workflows"})
		return
	}

	span.SetAttributes(
		attribute.Int("workflows_count", len(workflows)),
		attribute.Int64("total_workflows", total),
		attribute.Int("page", page),
		attribute.Int("limit", limit),
	)

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetWorkflow retrieves a specific recovery workflow
func (h *RecoveryHandlers) GetWorkflow(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_workflow")
	defer span.End()

	workflowID := c.Param("id")
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	workflow, err := h.recoveryService.GetWorkflow(ctx, workflowUUID, companyUUID)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	span.SetAttributes(
		attribute.String("workflow_id", workflow.ID.String()),
		attribute.String("workflow_name", workflow.Name),
	)

	c.JSON(http.StatusOK, gin.H{"workflow": workflow})
}

// UpdateWorkflow updates an existing recovery workflow
func (h *RecoveryHandlers) UpdateWorkflow(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "update_workflow")
	defer span.End()

	workflowID := c.Param("id")
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req struct {
		Name              *string                 `json:"name"`
		Description       *string                 `json:"description"`
		Priority          *int                    `json:"priority"`
		IsActive          *bool                   `json:"is_active"`
		TriggerConditions *map[string]interface{} `json:"trigger_conditions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TriggerConditions != nil {
		conditionsJSON, err := json.Marshal(*req.TriggerConditions)
		if err != nil {
			span.RecordError(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger conditions"})
			return
		}
		updates["trigger_conditions"] = conditionsJSON
	}

	if err := h.recoveryService.UpdateWorkflow(ctx, workflowUUID, companyUUID, updates); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
		return
	}

	span.SetAttributes(attribute.String("workflow_id", workflowUUID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Workflow updated successfully"})
}

// DeleteWorkflow soft deletes a recovery workflow
func (h *RecoveryHandlers) DeleteWorkflow(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "delete_workflow")
	defer span.End()

	workflowID := c.Param("id")
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	if err := h.recoveryService.DeleteWorkflow(ctx, workflowUUID, companyUUID); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
		return
	}

	span.SetAttributes(attribute.String("workflow_id", workflowUUID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted successfully"})
}

// GetWorkflowExecutions retrieves workflow executions
func (h *RecoveryHandlers) GetWorkflowExecutions(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_workflow_executions")
	defer span.End()

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if workflowID := c.Query("workflow_id"); workflowID != "" {
		filters["workflow_id"] = workflowID
	}

	executions, total, err := h.recoveryService.GetWorkflowExecutions(ctx, companyUUID, filters, page, limit)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve executions"})
		return
	}

	span.SetAttributes(
		attribute.Int("executions_count", len(executions)),
		attribute.Int64("total_executions", total),
	)

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// PauseWorkflowExecution pauses a running workflow execution
func (h *RecoveryHandlers) PauseWorkflowExecution(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "pause_workflow_execution")
	defer span.End()

	executionID := c.Param("id")
	executionUUID, err := uuid.Parse(executionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	if err := h.recoveryService.PauseWorkflowExecution(ctx, executionUUID); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause execution"})
		return
	}

	span.SetAttributes(attribute.String("execution_id", executionUUID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Workflow execution paused successfully"})
}

// ResumeWorkflowExecution resumes a paused workflow execution
func (h *RecoveryHandlers) ResumeWorkflowExecution(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "resume_workflow_execution")
	defer span.End()

	executionID := c.Param("id")
	executionUUID, err := uuid.Parse(executionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	if err := h.recoveryService.ResumeWorkflowExecution(ctx, executionUUID); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume execution"})
		return
	}

	span.SetAttributes(attribute.String("execution_id", executionUUID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Workflow execution resumed successfully"})
}

// CancelWorkflowExecution cancels a workflow execution
func (h *RecoveryHandlers) CancelWorkflowExecution(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "cancel_workflow_execution")
	defer span.End()

	executionID := c.Param("id")
	executionUUID, err := uuid.Parse(executionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	if err := h.recoveryService.CancelWorkflowExecution(ctx, executionUUID); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel execution"})
		return
	}

	span.SetAttributes(attribute.String("execution_id", executionUUID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Workflow execution cancelled successfully"})
}

// TriggerWorkflow manually triggers a workflow for a payment failure
func (h *RecoveryHandlers) TriggerWorkflow(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "trigger_workflow")
	defer span.End()

	var req struct {
		WorkflowID       string `json:"workflow_id" binding:"required"`
		PaymentFailureID string `json:"payment_failure_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflowUUID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	paymentFailureUUID, err := uuid.Parse(req.PaymentFailureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment failure ID"})
		return
	}

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	if err := h.recoveryService.TriggerWorkflowManually(ctx, workflowUUID, paymentFailureUUID, companyUUID); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger workflow"})
		return
	}

	span.SetAttributes(
		attribute.String("workflow_id", workflowUUID.String()),
		attribute.String("payment_failure_id", paymentFailureUUID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Workflow triggered successfully"})
}

// Communication Template Endpoints

// CreateTemplate creates a new communication template
func (h *RecoveryHandlers) CreateTemplate(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_template")
	defer span.End()

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req struct {
		Name         string                 `json:"name" binding:"required"`
		Description  string                 `json:"description"`
		TemplateType string                 `json:"template_type" binding:"required"`
		Subject      string                 `json:"subject"`
		Content      string                 `json:"content" binding:"required"`
		Variables    map[string]interface{} `json:"variables"`
		Conditions   map[string]interface{} `json:"conditions"`
		IsDefault    bool                   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.CommunicationTemplate{
		CompanyID:    companyUUID,
		Name:         req.Name,
		Description:  req.Description,
		TemplateType: req.TemplateType,
		Subject:      req.Subject,
		Content:      req.Content,
		IsActive:     true,
		IsDefault:    req.IsDefault,
		CreatedBy:    c.GetString("user_id"),
	}

	// Convert variables to JSON
	if req.Variables != nil {
		variablesJSON, err := json.Marshal(req.Variables)
		if err != nil {
			span.RecordError(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variables"})
			return
		}
		template.Variables = variablesJSON
	}

	// Convert conditions to JSON
	if req.Conditions != nil {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			span.RecordError(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conditions"})
			return
		}
		template.Conditions = conditionsJSON
	}

	if err := h.communicationService.CreateTemplate(ctx, template); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	span.SetAttributes(
		attribute.String("template_id", template.ID.String()),
		attribute.String("template_name", template.Name),
		attribute.String("template_type", template.TemplateType),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Template created successfully",
		"template": template,
	})
}

// GetTemplates retrieves communication templates
func (h *RecoveryHandlers) GetTemplates(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_templates")
	defer span.End()

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	templateType := c.Query("type")

	templates, total, err := h.communicationService.GetTemplates(ctx, companyUUID, templateType, page, limit)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}

	span.SetAttributes(
		attribute.Int("templates_count", len(templates)),
		attribute.Int64("total_templates", total),
	)

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetRecoveryMetrics retrieves recovery performance metrics
func (h *RecoveryHandlers) GetRecoveryMetrics(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "get_recovery_metrics")
	defer span.End()

	companyID := c.GetString("company_id")
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Parse time range
	timeRange := c.DefaultQuery("time_range", "30d")
	var duration time.Duration
	switch timeRange {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	case "90d":
		duration = 90 * 24 * time.Hour
	default:
		duration = 30 * 24 * time.Hour
	}

	metrics, err := h.recoveryService.GetRecoveryMetrics(ctx, companyUUID, duration)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	span.SetAttributes(
		attribute.String("time_range", timeRange),
		attribute.String("company_id", companyUUID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// RecoveryWorkflow represents an automated recovery workflow for payment failures
type RecoveryWorkflow struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID   uuid.UUID      `json:"company_id" gorm:"type:uuid;not null;index"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	Priority    int            `json:"priority" gorm:"default:1;index"` // Higher number = higher priority
	
	// Trigger conditions
	TriggerConditions datatypes.JSON `json:"trigger_conditions" gorm:"type:jsonb"` // JSON conditions for when to trigger
	
	// Workflow steps
	Steps []RecoveryWorkflowStep `json:"steps" gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE"`
	
	// Execution tracking
	Executions []RecoveryWorkflowExecution `json:"executions,omitempty" gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy string    `json:"created_by" gorm:"size:255"`
	
	// Relations
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// RecoveryWorkflowStep represents a single step in a recovery workflow
type RecoveryWorkflowStep struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID uuid.UUID `json:"workflow_id" gorm:"type:uuid;not null;index"`
	StepOrder  int       `json:"step_order" gorm:"not null;index"`
	
	// Step configuration
	StepType    string          `json:"step_type" gorm:"size:50;not null"` // retry_payment, send_email, send_sms, wait, conditional
	StepName    string          `json:"step_name" gorm:"size:255;not null"`
	Description string          `json:"description" gorm:"type:text"`
	Config      datatypes.JSON  `json:"config" gorm:"type:jsonb"` // Step-specific configuration
	
	// Conditional logic
	Conditions datatypes.JSON `json:"conditions,omitempty" gorm:"type:jsonb"` // Conditions for step execution
	
	// Timing
	DelayMinutes int  `json:"delay_minutes" gorm:"default:0"` // Delay before executing this step
	IsParallel   bool `json:"is_parallel" gorm:"default:false"` // Can run in parallel with next step
	
	// Status
	IsActive   bool `json:"is_active" gorm:"default:true"`
	IsCritical bool `json:"is_critical" gorm:"default:false"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Workflow   RecoveryWorkflow            `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	Executions []RecoveryStepExecution     `json:"executions,omitempty" gorm:"foreignKey:StepID;constraint:OnDelete:CASCADE"`
}

// RecoveryWorkflowExecution tracks the execution of a workflow for a specific payment failure
type RecoveryWorkflowExecution struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID        uuid.UUID `json:"workflow_id" gorm:"type:uuid;not null;index"`
	PaymentFailureID  uuid.UUID `json:"payment_failure_id" gorm:"type:uuid;not null;index"`
	CompanyID         uuid.UUID `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Execution status
	Status        string    `json:"status" gorm:"size:50;not null;index"` // pending, running, completed, failed, paused, cancelled
	CurrentStepID *uuid.UUID `json:"current_step_id,omitempty" gorm:"type:uuid;index"`
	
	// Timing
	StartedAt   time.Time  `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	PausedAt    *time.Time `json:"paused_at,omitempty"`
	
	// Results
	TotalSteps      int             `json:"total_steps" gorm:"default:0"`
	CompletedSteps  int             `json:"completed_steps" gorm:"default:0"`
	FailedSteps     int             `json:"failed_steps" gorm:"default:0"`
	SuccessfulSteps int             `json:"successful_steps" gorm:"default:0"`
	ExecutionLog    datatypes.JSON  `json:"execution_log,omitempty" gorm:"type:jsonb"`
	
	// Error handling
	LastError    string         `json:"last_error,omitempty" gorm:"type:text"`
	RetryCount   int            `json:"retry_count" gorm:"default:0"`
	NextRetryAt  *time.Time     `json:"next_retry_at,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Workflow       RecoveryWorkflow        `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	PaymentFailure PaymentFailureEvent     `json:"payment_failure,omitempty" gorm:"foreignKey:PaymentFailureID"`
	Company        Company                 `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	CurrentStep    *RecoveryWorkflowStep   `json:"current_step,omitempty" gorm:"foreignKey:CurrentStepID"`
	StepExecutions []RecoveryStepExecution `json:"step_executions,omitempty" gorm:"foreignKey:WorkflowExecutionID;constraint:OnDelete:CASCADE"`
}

// RecoveryStepExecution tracks the execution of individual workflow steps
type RecoveryStepExecution struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowExecutionID uuid.UUID `json:"workflow_execution_id" gorm:"type:uuid;not null;index"`
	StepID              uuid.UUID `json:"step_id" gorm:"type:uuid;not null;index"`
	
	// Execution details
	Status      string         `json:"status" gorm:"size:50;not null;index"` // pending, running, completed, failed, skipped
	StartedAt   time.Time      `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Duration    int64          `json:"duration_ms,omitempty"` // Duration in milliseconds
	
	// Results
	Result       datatypes.JSON `json:"result,omitempty" gorm:"type:jsonb"`
	ErrorMessage string         `json:"error_message,omitempty" gorm:"type:text"`
	RetryCount   int            `json:"retry_count" gorm:"default:0"`
	
	// Action tracking
	ActionType   string         `json:"action_type" gorm:"size:100"` // payment_retry, email_sent, sms_sent, etc.
	ActionData   datatypes.JSON `json:"action_data,omitempty" gorm:"type:jsonb"`
	ExternalID   string         `json:"external_id,omitempty" gorm:"size:255;index"` // External system reference
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	WorkflowExecution RecoveryWorkflowExecution `json:"workflow_execution,omitempty" gorm:"foreignKey:WorkflowExecutionID"`
	Step              RecoveryWorkflowStep      `json:"step,omitempty" gorm:"foreignKey:StepID"`
}

// CommunicationTemplate represents templates for customer communications
type CommunicationTemplate struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID   uuid.UUID `json:"company_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	
	// Template configuration
	TemplateType string `json:"template_type" gorm:"size:50;not null;index"` // email, sms, in_app
	Subject      string `json:"subject,omitempty" gorm:"size:500"` // For email templates
	Content      string `json:"content" gorm:"type:text;not null"`
	
	// Template variables and personalization
	Variables    datatypes.JSON `json:"variables,omitempty" gorm:"type:jsonb"` // Available template variables
	Conditions   datatypes.JSON `json:"conditions,omitempty" gorm:"type:jsonb"` // Conditions for template selection
	
	// Metadata
	IsActive    bool      `json:"is_active" gorm:"default:true;index"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy   string    `json:"created_by" gorm:"size:255"`
	
	// Usage tracking
	UsageCount  int       `json:"usage_count" gorm:"default:0"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	
	// Relations
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// RecoveryAction represents individual recovery actions taken
type RecoveryAction struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID         uuid.UUID `json:"company_id" gorm:"type:uuid;not null;index"`
	PaymentFailureID  uuid.UUID `json:"payment_failure_id" gorm:"type:uuid;not null;index"`
	WorkflowExecutionID *uuid.UUID `json:"workflow_execution_id,omitempty" gorm:"type:uuid;index"`
	StepExecutionID   *uuid.UUID `json:"step_execution_id,omitempty" gorm:"type:uuid;index"`
	
	// Action details
	ActionType   string         `json:"action_type" gorm:"size:100;not null;index"` // payment_retry, email_sent, sms_sent, etc.
	ActionData   datatypes.JSON `json:"action_data" gorm:"type:jsonb"`
	Status       string         `json:"status" gorm:"size:50;not null;index"` // pending, completed, failed
	
	// Provider information
	Provider     string `json:"provider" gorm:"size:100;index"` // stripe, xero, quickbooks, etc.
	ExternalID   string `json:"external_id,omitempty" gorm:"size:255;index"`
	
	// Results
	Result       datatypes.JSON `json:"result,omitempty" gorm:"type:jsonb"`
	ErrorMessage string         `json:"error_message,omitempty" gorm:"type:text"`
	
	// Timing
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Company           Company                    `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	PaymentFailure    PaymentFailureEvent        `json:"payment_failure,omitempty" gorm:"foreignKey:PaymentFailureID"`
	WorkflowExecution *RecoveryWorkflowExecution `json:"workflow_execution,omitempty" gorm:"foreignKey:WorkflowExecutionID"`
	StepExecution     *RecoveryStepExecution     `json:"step_execution,omitempty" gorm:"foreignKey:StepExecutionID"`
}

// RecoveryMetrics represents aggregated recovery performance metrics
type RecoveryMetrics struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Time period
	PeriodStart time.Time `json:"period_start" gorm:"not null;index"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null;index"`
	PeriodType  string    `json:"period_type" gorm:"size:20;not null"` // daily, weekly, monthly
	
	// Workflow metrics
	TotalWorkflowExecutions    int     `json:"total_workflow_executions" gorm:"default:0"`
	SuccessfulWorkflowExecutions int   `json:"successful_workflow_executions" gorm:"default:0"`
	FailedWorkflowExecutions   int     `json:"failed_workflow_executions" gorm:"default:0"`
	AverageExecutionTime       float64 `json:"average_execution_time_minutes" gorm:"default:0"`
	
	// Recovery metrics
	TotalRecoveryActions       int     `json:"total_recovery_actions" gorm:"default:0"`
	SuccessfulRecoveryActions  int     `json:"successful_recovery_actions" gorm:"default:0"`
	RecoverySuccessRate        float64 `json:"recovery_success_rate" gorm:"default:0"`
	TotalAmountRecovered       float64 `json:"total_amount_recovered" gorm:"default:0"`
	
	// Communication metrics
	EmailsSent                 int     `json:"emails_sent" gorm:"default:0"`
	SMSSent                    int     `json:"sms_sent" gorm:"default:0"`
	CommunicationResponseRate  float64 `json:"communication_response_rate" gorm:"default:0"`
	
	// Performance metrics
	AverageTimeToRecovery      float64 `json:"average_time_to_recovery_hours" gorm:"default:0"`
	FirstAttemptSuccessRate    float64 `json:"first_attempt_success_rate" gorm:"default:0"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relations
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// TableName methods for custom table names
func (RecoveryWorkflow) TableName() string {
	return "recovery_workflows"
}

func (RecoveryWorkflowStep) TableName() string {
	return "recovery_workflow_steps"
}

func (RecoveryWorkflowExecution) TableName() string {
	return "recovery_workflow_executions"
}

func (RecoveryStepExecution) TableName() string {
	return "recovery_step_executions"
}

func (CommunicationTemplate) TableName() string {
	return "communication_templates"
}

func (RecoveryAction) TableName() string {
	return "recovery_actions"
}

func (RecoveryMetrics) TableName() string {
	return "recovery_metrics"
}

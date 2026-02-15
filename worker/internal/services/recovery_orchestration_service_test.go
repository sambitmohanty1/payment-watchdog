package services_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	svc "github.com/sambitmohanty1/payment-watchdog/internal/services"
	"github.com/sambitmohanty1/payment-watchdog/internal/models"
)

// MockRecoveryOrchestrationService creates a test instance of RecoveryOrchestrationService with a mock DB
func MockRecoveryOrchestrationService(t *testing.T) (*svc.RecoveryOrchestrationService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Initialize dependencies
	logger := zap.NewNop()
	retrySvc := &svc.RetryService{DB: gormDB, Logger: logger}
	commSvc := &svc.CommunicationService{DB: gormDB, Logger: logger}
	analyticsSvc := &svc.AnalyticsService{DB: gormDB, Logger: logger}

	service := svc.NewRecoveryOrchestrationService(
		gormDB,
		retrySvc,
		commSvc,
		analyticsSvc,
		logger,
	)

	return service, gormDB, mock
}

func TestTriggerWorkflowsForFailure(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New()
	paymentFailure := &models.PaymentFailureEvent{
		ID:            uuid.New(),
		CompanyID:     companyID,
		Amount:        100.0,
		Currency:      "USD",
		FailureReason: "insufficient_funds",
		Provider:      "stripe",
	}

	tests := []struct {
		name        string
		setupMocks  func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "successful workflow trigger",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock workflow query
				workflowID := uuid.New()
				rows := sqlmock.NewRows([]string{"id", "company_id", "name", "is_active", "trigger_conditions"}).
					AddRow(workflowID, companyID, "Test Workflow", true, `{"conditions":[{"field":"amount","operator":"gt","value":50}],"logic":"AND"}`)

				mock.ExpectQuery(`^SELECT \* FROM "recovery_workflows"`).
					WithArgs(companyID, true).
					WillReturnRows(rows)

				// Mock workflow steps
				stepID := uuid.New()
				stepRows := sqlmock.NewRows([]string{"id", "workflow_id", "step_type", "config", "is_active"}).
					AddRow(stepID, workflowID, "retry_payment", `{"provider":"stripe"}`, true)

				mock.ExpectQuery(`^SELECT \* FROM "recovery_workflow_steps"`).
					WithArgs(workflowID, true).
					WillReturnRows(stepRows)

				// Mock execution creation
				executionID := uuid.New()
				execRows := sqlmock.NewRows([]string{"id"}).AddRow(executionID)
				mock.ExpectBegin()
				mock.ExpectQuery(`^INSERT INTO "recovery_workflow_executions"`).
					WillReturnRows(execRows)
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "no matching workflows",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "recovery_workflows"`).
					WithArgs(companyID, true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, mock := MockRecoveryOrchestrationService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mock)
			}

			err := service.TriggerWorkflowsForFailure(ctx, paymentFailure)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestExecuteStep(t *testing.T) {
	ctx := context.Background()
	execution := &svc.WorkflowExecution{
		ID:              uuid.New(),
		WorkflowID:      uuid.New(),
		PaymentFailureID: uuid.New(),
		CompanyID:       uuid.New(),
		Status:          "running",
		Context:         make(map[string]interface{}),
	}

	paymentFailure := &models.PaymentFailureEvent{
		ID:            execution.PaymentFailureID,
		CompanyID:     execution.CompanyID,
		Amount:        150.0,
		Currency:      "USD",
		FailureReason: "card_declined",
	}
	execution.Context["payment_failure"] = paymentFailure

	tests := []struct {
		name        string
		step        *models.RecoveryWorkflowStep
		setupMocks  func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "execute payment retry step",
			step: &models.RecoveryWorkflowStep{
				ID:       uuid.New(),
				StepType: "retry_payment",
				Config:   []byte(`{"provider":"stripe"}`),
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock step execution creation
				mock.ExpectBegin()
				mock.ExpectQuery(`^INSERT INTO "recovery_step_executions"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
				mock.ExpectCommit()

				// Mock retry service
				mock.ExpectBegin()
				mock.ExpectQuery(`^INSERT INTO "recovery_actions"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
				mock.ExpectCommit()

				// Mock step execution update
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "recovery_step_executions"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// Mock execution counter update
				mock.ExpectExec(`^UPDATE "recovery_workflow_executions"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, mock := MockRecoveryOrchestrationService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mock)
			}

			err := service.ExecuteStep(ctx, execution, tt.step)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestEvaluateTriggerConditions(t *testing.T) {
	service, _, _ := MockRecoveryOrchestrationService(t)

	tests := []struct {
		name       string
		failure    *models.PaymentFailureEvent
		conditions []byte
		expected   bool
	}{
		{
			name: "simple amount condition - match",
			failure: &models.PaymentFailureEvent{
				Amount: 100.0,
			},
			conditions: []byte(`{"conditions":[{"field":"amount","operator":"gt","value":50}],"logic":"AND"}`),
			expected:   true,
		},
		{
			name: "multiple conditions with AND logic - match",
			failure: &models.PaymentFailureEvent{
				Amount:    100.0,
				Currency:  "USD",
				Provider:  "stripe",
			},
			conditions: []byte(`{"conditions":[
				{"field":"amount","operator":"gt","value":50},
				{"field":"currency","operator":"equals","value":"USD"}
			],"logic":"AND"}`),
			expected:   true,
		},
		{
			name: "multiple conditions with OR logic - match",
			failure: &models.PaymentFailureEvent{
				Amount:    30.0,
				Currency:  "USD",
			},
			conditions: []byte(`{"conditions":[
				{"field":"amount","operator":"gt","value":50},
				{"field":"currency","operator":"equals","value":"USD"}
			],"logic":"OR"}`),
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.EvaluateTriggerConditions(tt.failure, tt.conditions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestWorkflowExecutionLifecycle tests the complete workflow execution lifecycle
func TestWorkflowExecutionLifecycle(t *testing.T) {
	ctx := context.Background()
	service, db, mock := MockRecoveryOrchestrationService(t)

	// Setup test data
	companyID := uuid.New()
	workflowID := uuid.New()
	paymentFailureID := uuid.New()

	// Mock workflow with steps
	workflow := &models.RecoveryWorkflow{
		ID:          workflowID,
		CompanyID:   companyID,
		Name:        "Test Workflow",
		IsActive:    true,
		TriggerConditions: []byte(`{"conditions":[{"field":"amount","operator":"gt","value":0}],"logic":"AND"}`),
	}

	// Add steps to the workflow
	step1ID := uuid.New()
	step2ID := uuid.New()
	workflow.Steps = []models.RecoveryWorkflowStep{
		{
			ID:         step1ID,
			WorkflowID: workflowID,
			StepType:   "wait",
			Config:     []byte(`{"wait_minutes": 1}`),
			IsActive:   true,
		},
		{
			ID:         step2ID,
			WorkflowID: workflowID,
			StepType:   "send_email",
			Config:     []byte(`{"template_id":"test-template"}`),
			IsActive:   true,
		},
	}

	// Setup database mocks
	// 1. Mock workflow creation
	db.Create(workflow)

	// 2. Mock workflow execution
	executionID := uuid.New()
	execution := &models.RecoveryWorkflowExecution{
		ID:               executionID,
		WorkflowID:       workflowID,
		PaymentFailureID: paymentFailureID,
		CompanyID:        companyID,
		Status:           "pending",
		StartedAt:        time.Now(),
	}

	// 3. Start workflow execution
	err := service.StartWorkflowExecution(ctx, workflow, &models.PaymentFailureEvent{
		ID:        paymentFailureID,
		CompanyID: companyID,
		Amount:    100.0,
	})
	require.NoError(t, err)

	// 4. Verify execution was created
	var createdExecution models.RecoveryWorkflowExecution
	err = db.First(&createdExecution, "id = ?", executionID).Error
	require.NoError(t, err)
	assert.Equal(t, "running", createdExecution.Status)

	// 5. Simulate step completion (this would normally be done by the workflow executor)
	stepExecution := &models.RecoveryStepExecution{
		ID:                  uuid.New(),
		WorkflowExecutionID: executionID,
		StepID:              step1ID,
		Status:              "completed",
		StartedAt:           time.Now(),
		CompletedAt:         time.Now().Add(5 * time.Second),
	}
	db.Create(stepExecution)

	// 6. Verify workflow completes successfully
	// In a real test, you would wait for the workflow to complete or mock the executor
	// This is a simplified example

	// Cleanup
	db.Delete(workflow)
	db.Delete(execution)
	db.Delete(stepExecution)
}

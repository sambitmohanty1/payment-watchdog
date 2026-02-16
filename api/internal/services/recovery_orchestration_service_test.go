package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
)

// MockRecoveryOrchestrationService creates a test service with mock DB
func MockRecoveryOrchestrationService(t *testing.T) (*RecoveryOrchestrationService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	logger := zap.NewNop()

	// Create mock services
	retryService := &mockRetryService{}
	communicationService := &mockCommunicationService{}
	analyticsService := &mockAnalyticsService{}

	service := NewRecoveryOrchestrationService(
		gormDB,
		retryService,
		communicationService,
		analyticsService,
		logger,
	)

	return service, gormDB, mock
}

// Mock services for testing
type mockRetryService struct{}

func (m *mockRetryService) RetryPayment(ctx context.Context, executionID uuid.UUID, config *models.RetryConfig) error {
	return nil
}

type mockCommunicationService struct{}

func (m *mockCommunicationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	return nil
}

type mockAnalyticsService struct{}

func (m *mockAnalyticsService) GetRecoveryMetrics(ctx context.Context, companyID uuid.UUID, startTime, endTime time.Time) (*models.RecoveryMetrics, error) {
	return &models.RecoveryMetrics{
		TotalRecoveredAmount:  1000.0,
		TotalFailedAmount:     5000.0,
		RecoveryRate:          50.0,
		AverageRecoveryTime:   3600.0,
		TotalRecovered:        10,
		TotalFailed:           20,
		RecoveryByMethod:      []models.Metric{},
		RecoveryByFailureType: []models.Metric{},
		RecoveryTrends:        []models.Trend{},
		RecoveryScore:         75,
		LastUpdated:           time.Now(),
	}, nil
}

func TestRecoveryOrchestrationService_Creation(t *testing.T) {
	service, _, _ := MockRecoveryOrchestrationService(t)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.logger)
}

func TestRecoveryOrchestrationService_TriggerWorkflowsForFailure(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New()

	paymentFailure := &models.PaymentFailureEvent{
		ID:            uuid.New(),
		CompanyID:     companyID.String(),
		AmountCents:   10000, // $100.00 in cents
		Currency:      "USD",
		FailureReason: "insufficient_funds",
		Provider:      "stripe",
	}

	tests := []struct {
		name        string
		setupMocks  func(mock sqlmock.Sqlmock)
		expectError bool
		expectCount int
	}{
		{
			name: "successful workflow trigger",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock workflow query
				rows := sqlmock.NewRows([]string{"id", "name", "company_id", "trigger_conditions", "is_active"}).
					AddRow(uuid.New(), "Test Workflow", companyID.String(), `{"failure_reason": "insufficient_funds"}`, true)
				mock.ExpectQuery(`SELECT.*FROM recovery_workflows`).
					WithArgs(companyID.String()).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "no workflows found",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock empty workflow query
				rows := sqlmock.NewRows([]string{"id", "name", "company_id", "trigger_conditions", "is_active"})
				mock.ExpectQuery(`SELECT.*FROM recovery_workflows`).
					WithArgs(companyID.String()).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, gormDB, mock := MockRecoveryOrchestrationService(t)

			// Setup mocks
			tt.setupMocks(mock)

			// Execute test
			workflows, err := service.TriggerWorkflowsForFailure(ctx, paymentFailure)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, workflows)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCount, len(workflows))
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())

			// Clean up
			sqlDB, _ := gormDB.DB()
			sqlDB.Close()
		})
	}
}

func TestRecoveryOrchestrationService_ExecuteWorkflow(t *testing.T) {
	ctx := context.Background()
	workflowID := uuid.New()
	companyID := uuid.New()

	execution := &models.RecoveryWorkflowExecution{
		ID:               uuid.New(),
		WorkflowID:       workflowID,
		PaymentFailureID: uuid.New(),
		CompanyID:        companyID,
		Status:           "running",
		ExecutionLog:     make(map[string]interface{}),
	}

	paymentFailure := &models.PaymentFailureEvent{
		ID:            execution.PaymentFailureID,
		CompanyID:     execution.CompanyID,
		AmountCents:   15000, // $150.00 in cents
		Currency:      "USD",
		FailureReason: "card_declined",
	}

	tests := []struct {
		name        string
		step        *models.RecoveryWorkflowStep
		setupMocks  func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "successful step execution",
			step: &models.RecoveryWorkflowStep{
				ID:          uuid.New(),
				WorkflowID:  workflowID,
				StepOrder:   1,
				StepType:    "retry_payment",
				StepName:    "Retry Payment",
				Description: "Retry failed payment with new amount",
				Config: map[string]interface{}{
					"retry_amount": 20000, // $200.00 in cents
				},
				Conditions: map[string]interface{}{
					"max_retries": 3,
				},
				DelayMinutes: 0,
				IsParallel:   false,
				IsActive:     true,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock execution insertion
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO recovery_workflow_executions`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "step execution with delay",
			step: &models.RecoveryWorkflowStep{
				ID:           uuid.New(),
				WorkflowID:   workflowID,
				StepOrder:    2,
				StepType:     "wait",
				StepName:     "Wait before retry",
				Description:  "Wait 5 minutes before retrying",
				DelayMinutes: 5,
				IsParallel:   false,
				IsActive:     true,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock execution insertion
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO recovery_workflow_executions`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, gormDB, mock := MockRecoveryOrchestrationService(t)

			// Setup mocks
			tt.setupMocks(mock)

			// Add payment failure to execution context
			execution.ExecutionLog["payment_failure"] = paymentFailure

			// Execute test
			err := service.executeStep(ctx, execution, tt.step)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())

			// Clean up
			sqlDB, _ := gormDB.DB()
			sqlDB.Close()
		})
	}
}

func TestRecoveryOrchestrationService_GetWorkflowExecution(t *testing.T) {
	ctx := context.Background()
	workflowID := uuid.New()
	companyID := uuid.New()
	executionID := uuid.New()

	tests := []struct {
		name        string
		setupMocks  func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "successful retrieval",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock execution query
				rows := sqlmock.NewRows([]string{
					"id", "workflow_id", "payment_failure_id", "company_id", "status",
					"current_step_id", "started_at", "completed_at", "paused_at",
					"total_steps", "completed_steps", "failed_steps", "successful_steps",
					"execution_log", "last_error", "retry_count", "next_retry_at",
					"created_at", "updated_at",
				}).
					AddRow(executionID, workflowID, paymentFailureID, companyID, "running",
						uuid.New(), time.Now(), nil, nil, 1, 0, 0, 0,
						make(map[string]interface{}), "", 0, 0, nil, nil, time.Now())
				mock.ExpectQuery(`SELECT.*FROM recovery_workflow_executions`).
					WithArgs(executionID).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "execution not found",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock empty query
				rows := sqlmock.NewRows([]string{
					"id", "workflow_id", "payment_failure_id", "company_id", "status",
					"current_step_id", "started_at", "completed_at", "paused_at",
					"total_steps", "completed_steps", "failed_steps", "successful_steps",
					"execution_log", "last_error", "retry_count", "next_retry_at",
					"created_at", "updated_at",
				})
				mock.ExpectQuery(`SELECT.*FROM recovery_workflow_executions`).
					WithArgs(executionID).
					WillReturnRows(rows)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, gormDB, mock := MockRecoveryOrchestrationService(t)

			// Setup mocks
			tt.setupMocks(mock)

			// Execute test
			execution, err := service.GetWorkflowExecution(ctx, executionID)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, execution)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, execution)
				assert.Equal(t, executionID, execution.ID)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())

			// Clean up
			sqlDB, _ := gormDB.DB()
			sqlDB.Close()
		})
	}
}

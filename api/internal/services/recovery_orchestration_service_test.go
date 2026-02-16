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
	service := NewRecoveryOrchestrationService(gormDB, logger)

	return service, gormDB, mock
}

func TestRecoveryOrchestrationService_Creation(t *testing.T) {
	service, _, _ := MockRecoveryOrchestrationService(t)
	
	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestRecoveryOrchestrationService_TriggerWorkflows(t *testing.T) {
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
				assert.NotNil(t, workflows)
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
		Context:          make(map[string]interface{}),
	}

	paymentFailure := &models.PaymentFailureEvent{
		ID:            execution.PaymentFailureID,
		CompanyID:     execution.CompanyID.String(),
		AmountCents:   15000, // $150.00 in cents
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
			name: "successful step execution",
			step: &models.RecoveryWorkflowStep{
				ID:          uuid.New(),
				WorkflowID:  workflowID,
				Name:        "Send Email",
				Type:        "email",
				Config:      map[string]interface{}{"template": "payment_failed"},
				Order:       1,
				IsEnabled:   true,
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

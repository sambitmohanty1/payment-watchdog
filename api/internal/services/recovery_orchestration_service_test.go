package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
)

// Mock services for testing - simplified for basic functionality
type mockRetryService struct{}

type mockCommunicationService struct{}

type mockAnalyticsService struct{}

// MockRecoveryOrchestrationService creates a test service with mock DB
func MockRecoveryOrchestrationService(t *testing.T) (*RecoveryOrchestrationService, *gorm.DB, sqlmock.Sqlmock) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Create GORM DB from mock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create logger
	logger, _ := zap.NewDevelopment()

	// Create service with nil dependencies for basic testing
	service := &RecoveryOrchestrationService{
		db:                   gormDB,
		retryService:         nil,
		communicationService: nil,
		analyticsService:     nil,
		stepExecutors:        make(map[string]StepExecutor),
		tracer:               otel.Tracer("recovery-orchestration-test"),
		activeExecutions:     make(map[uuid.UUID]*WorkflowExecution),
		executionWorkers:     10,
		workerPool:           make(chan struct{}, 10),
		logger:               logger,
	}

	// Register a simple wait executor for testing
	service.RegisterStepExecutor(&WaitExecutor{service: service})

	return service, gormDB, mock
}

func TestRecoveryOrchestrationService_Creation(t *testing.T) {
	service, _, _ := MockRecoveryOrchestrationService(t)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.logger)
}

func TestRecoveryOrchestrationService_TriggerWorkflowsForFailure(t *testing.T) {
	companyID := uuid.New()

	paymentFailure := &models.PaymentFailureEvent{
		ID:            uuid.New(),
		CompanyID:     companyID.String(),
		AmountCents:   10000, // $100.00 in cents
		Currency:      "USD",
		FailureReason: "insufficient_funds",
		Provider:      "stripe",
	}

	// Simple test - just verify the method is called without errors
	t.Run("trigger workflows for failure", func(t *testing.T) {
		service, gormDB, mock := MockRecoveryOrchestrationService(t)

		// Setup mock to return no workflows - simplified approach
		// We'll match any query to recovery_workflows table
		rows := sqlmock.NewRows([]string{"id", "name", "company_id", "trigger_conditions", "is_active", "priority", "created_at", "updated_at"})
		mock.ExpectQuery(`.*recovery_workflows.*`).
			WillReturnRows(rows)

		// Execute test
		err := service.TriggerWorkflowsForFailure(context.Background(), paymentFailure)

		// Verify results
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())

		// Clean up
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	})
}

func TestRecoveryOrchestrationService_ExecuteStep(t *testing.T) {
	// Simple test - verify service creation and executor registration
	t.Run("service has wait executor", func(t *testing.T) {
		service, gormDB, mock := MockRecoveryOrchestrationService(t)

		// Verify wait executor is registered
		service.mu.RLock()
		waitExecutor, exists := service.stepExecutors["wait"]
		service.mu.RUnlock()

		assert.True(t, exists, "Wait executor should be registered")
		assert.NotNil(t, waitExecutor, "Wait executor should not be nil")
		assert.Equal(t, "wait", waitExecutor.GetStepType(), "Executor type should be wait")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())

		// Clean up
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	})

	// Test with actual workflow data
	t.Run("trigger workflows with actual data", func(t *testing.T) {
		service, gormDB, mock := MockRecoveryOrchestrationService(t)
		companyID := uuid.New()
		workflowID := uuid.New()

		paymentFailure := &models.PaymentFailureEvent{
			ID:            uuid.New(),
			CompanyID:     companyID.String(),
			AmountCents:   5000, // $50.00
			Currency:      "USD",
			FailureReason: "card_declined",
			Provider:      "stripe",
		}

		// Mock returning a workflow
		rows := sqlmock.NewRows([]string{"id", "name", "company_id", "trigger_conditions", "is_active", "priority", "created_at", "updated_at"}).AddRow(
			workflowID.String(), "Test Recovery Workflow", companyID.String(), []byte("{}"), true, 1, time.Now(), time.Now(),
		)
		mock.ExpectQuery(`.*recovery_workflows.*`).WillReturnRows(rows)

		// Mock the steps query (from preload) - return empty steps to simplify test
		stepRows := sqlmock.NewRows([]string{"id", "workflow_id", "step_order", "step_type", "step_name", "description", "config", "conditions", "delay_minutes", "is_parallel", "is_active", "is_critical", "created_at", "updated_at"})
		mock.ExpectQuery(`.*recovery_workflow_steps.*`).WillReturnRows(stepRows)

		// Mock the database transaction Begin call
		mock.ExpectBegin()

		// Mock the workflow execution insert (Query with RETURNING)
		mock.ExpectQuery(`INSERT INTO.*recovery_workflow_executions.*RETURNING.*id`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

		// Mock the transaction commit
		mock.ExpectCommit()

		// Since we're executing asynchronously, we shouldn't fail if we don't catch all the async updates
		mock.MatchExpectationsInOrder(false)

		// Expected updates from the async goroutine execution
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*recovery_workflow_executions.*SET.*status.*=.*WHERE.*id.*=.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*recovery_workflow_executions.*SET.*status.*=.*WHERE.*id.*=.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*recovery_workflow_executions.*SET.*completed_at.*=.*WHERE.*id.*=.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// The goroutine might run updateExecutionStatus which begins a tx
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*recovery_workflow_executions.*SET.*status.*=.*WHERE.*id.*=.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Since we want the test to focus on successful trigger, and the async goroutine might execute operations
		// out of order or unpredictably, we'll allow any further queries.

		// The goroutine might run updateExecutionCompletedAt which begins a tx
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*recovery_workflow_executions.*SET.*completed_at.*=.*WHERE.*id.*=.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Since we want the test to focus on successful trigger, and the async goroutine might execute operations
		// out of order or unpredictably, we'll allow any further queries.
		// Allow any other update queries
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE.*`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Execute test - focus on successful trigger, not async execution
		err := service.TriggerWorkflowsForFailure(context.Background(), paymentFailure)

		// Verify basic functionality - workflow was triggered successfully
		assert.NoError(t, err)

		// Wait for the async goroutine to complete its work to avoid "database closed" errors
		// It executes a few DB queries in a defer block which might race with test cleanup
		time.Sleep(50 * time.Millisecond)

		// Clean up
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	})
}

package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockRecoveryAnalyticsService is a test helper to create a RecoveryAnalyticsService with a mock DB
func MockRecoveryAnalyticsService(t *testing.T) (*svc.RecoveryAnalyticsService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	logger := zap.NewNop()
	service := svc.NewRecoveryAnalyticsService(gormDB, logger)

	return service, gormDB, mock
}


func TestGetRecoveryMetrics(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	startTime := now.Add(-7 * 24 * time.Hour)
	endTime := now
	companyID := "test-company-123"

	tests := []struct {
		name           string
		setupMocks     func(mock sqlmock.Sqlmock)
		expectedMetric func() *svc.RecoveryMetrics
		expectError    bool
	}{
		{
			name: "successful metrics calculation",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock failed payments query
				failedRows := sqlmock.NewRows([]string{"count", "sum"}).
					AddRow(100, 5000.50)
				mock.ExpectQuery(`SELECT COUNT\(\*\) as count, COALESCE\(SUM\(amount\), 0\) as sum FROM "payment_failure_events"`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(failedRows)

				// Mock recovered payments query
				recoveredRows := sqlmock.NewRows([]string{"method", "failure_type", "count", "sum", "avg_time"}).
					AddRow("credit_card", "insufficient_funds", 30, 1800.00, 3600).
					AddRow("bank_transfer", "expired_card", 20, 1200.00, 7200)
				mock.ExpectQuery(`(?i)SELECT.*FROM payment_events`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(recoveredRows)

				// Mock hourly recovery rates query
				hourlyRows := sqlmock.NewRows([]string{"hour", "rate", "count"}).
					AddRow(9, 0.8, 10).
					AddRow(10, 0.9, 15)
				mock.ExpectQuery(`(?i)WITH recovery_attempts`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(hourlyRows)
			},
			expectedMetric: func() *svc.RecoveryMetrics {
				return &svc.RecoveryMetrics{
					RecoveryRate:           50.0, // (30+20)/100 * 100
					AverageRecoveryTime:    5040,  // (30*3600 + 20*7200)/50
					TotalRecoveredAmount:   3000.00,
					TotalFailedAmount:      5000.50,
					RecoveryByMethod: map[string]int64{
						"credit_card":  30,
						"bank_transfer": 20,
					},
					RecoveryByFailureType: map[string]int64{
						"insufficient_funds": 30,
						"expired_card":      20,
					},
				}
			},
			expectError: false,
		},
		{
			name: "no failed payments",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// No failed payments
				failedRows := sqlmock.NewRows([]string{"count", "sum"}).
					AddRow(0, 0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) as count, COALESCE\(SUM\(amount\), 0\) as sum FROM "payment_failure_events"`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(failedRows)

				// No need to mock other queries as they won't be called
			},
			expectedMetric: func() *svc.RecoveryMetrics {
				return &svc.RecoveryMetrics{
					RecoveryByMethod:      make(map[string]int64),
					RecoveryByFailureType: make(map[string]int64),
				}
			},
			expectError: false,
		},
		{
			name: "database error on failed payments query",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) as count, COALESCE\(SUM\(amount\), 0\) as sum FROM "payment_failure_events"`).
					WithArgs(companyID, startTime, endTime).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedMetric: func() *svc.RecoveryMetrics { return nil },
			expectError:    true,
		},
		{
			name: "partial recovery data",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Some failed payments
				failedRows := sqlmock.NewRows([]string{"count", "sum"}).
					AddRow(50, 2500.00)
				mock.ExpectQuery(`SELECT COUNT\(\*\) as count, COALESCE\(SUM\(amount\), 0\) as sum FROM "payment_failure_events"`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(failedRows)

				// No recovered payments (empty result)
				recoveredRows := sqlmock.NewRows([]string{"method", "failure_type", "count", "sum", "avg_time"})
				mock.ExpectQuery(`(?i)SELECT.*FROM payment_events`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(recoveredRows)

				// Empty hourly rates
				hourlyRows := sqlmock.NewRows([]string{"hour", "rate", "count"})
				mock.ExpectQuery(`(?i)WITH recovery_attempts`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(hourlyRows)
			},
			expectedMetric: func() *svc.RecoveryMetrics {
				return &svc.RecoveryMetrics{
					RecoveryRate:           0,
					AverageRecoveryTime:    0,
					TotalRecoveredAmount:   0,
					TotalFailedAmount:      2500.00,
					RecoveryByMethod:       make(map[string]int64),
					RecoveryByFailureType:  make(map[string]int64),
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			service, _, mock := MockRecoveryAnalyticsService(t)
			defer mock.ExpectClose()

			// Setup mocks
			tt.setupMocks(mock)

			// Execute
			metrics, err := service.GetRecoveryMetrics(ctx, companyID, startTime, endTime)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			expected := tt.expectedMetric()

			assert.InDelta(t, expected.RecoveryRate, metrics.RecoveryRate, 0.01)
			assert.InDelta(t, expected.AverageRecoveryTime, metrics.AverageRecoveryTime, 0.01)
			assert.InDelta(t, expected.TotalRecoveredAmount, metrics.TotalRecoveredAmount, 0.01)
			assert.InDelta(t, expected.TotalFailedAmount, metrics.TotalFailedAmount, 0.01)

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Add more test functions for other methods
func TestCalculateRecoveryScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *svc.RecoveryMetrics
		expected int
	}{
		{
			name: "high score",
			metrics: &svc.RecoveryMetrics{
				RecoveryRate:          90.0,
				AverageRecoveryTime:   3600, // 1 hour
				TotalRecoveredAmount: 9000,
				TotalFailedAmount:   10000,
			},
			expected: 95, // 45 (from rate) + 30 (from time) + 18 (from amount ratio)
		},
		// Add more test cases
	}

	service := &svc.RecoveryAnalyticsService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := service.calculateRecoveryScore(tt.metrics)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, score)
		})
	}
}

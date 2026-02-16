package services

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
func MockRecoveryAnalyticsService(t *testing.T) (*RecoveryAnalyticsService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	logger := zap.NewNop()
	service := NewRecoveryAnalyticsService(gormDB, logger)

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
		expectedMetric func() *RecoveryMetrics
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
				recoveredRows := sqlmock.NewRows([]string{"count", "sum"}).
					AddRow(50, 3000.00)
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
			expectedMetric: func() *RecoveryMetrics {
				return &RecoveryMetrics{
					RecoveryRate:        50.0, // (30+20)/100 * 100
					AverageRecoveryTime: 5040, // (30*3600 + 20*7200)/50
					RecoveryByMethod: []Metric{
						{Name: "credit_card", Value: 30},
						{Name: "bank_transfer", Value: 20},
					},
					RecoveryByFailureType: []Metric{
						{Name: "insufficient_funds", Value: 30},
						{Name: "expired_card", Value: 20},
					},
					RecoveryAmounts: Amounts{
						TotalFailed:    5000.50,
						TotalRecovered: 3000.00,
						TotalPending:   0,
					},
					RecoveryTrends: []Trend{
						{Date: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), Value: 0.8, Success: true},
						{Date: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), Value: 0.9, Success: true},
					},
				}
			},
			expectError: false,
		},
		{
			name: "no payment data",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// No failed payments
				failedRows := sqlmock.NewRows([]string{"count", "sum"}).
					AddRow(0, 0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) as count, COALESCE\(SUM\(amount\), 0\) as sum FROM "payment_failure_events"`).
					WithArgs(companyID, startTime, endTime).
					WillReturnRows(failedRows)

				// No need to mock other queries as they won't be called
			},
			expectedMetric: func() *RecoveryMetrics {
				return &RecoveryMetrics{
					RecoveryByMethod:      []Metric{},
					RecoveryByFailureType: []Metric{},
					RecoveryAmounts: Amounts{
						TotalFailed:    0,
						TotalRecovered: 0,
						TotalPending:   0,
					},
					RecoveryTrends: []Trend{},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, gormDB, mock := MockRecoveryAnalyticsService(t)

			// Setup mocks
			tt.setupMocks(mock)

			// Execute test
			metrics, err := service.GetRecoveryMetrics(ctx, companyID, startTime, endTime)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metrics)

				expected := tt.expectedMetric()
				assert.Equal(t, expected.RecoveryRate, metrics.RecoveryRate)
				assert.Equal(t, expected.AverageRecoveryTime, metrics.AverageRecoveryTime)
				assert.Equal(t, expected.RecoveryAmounts.TotalFailed, metrics.RecoveryAmounts.TotalFailed)
				assert.Equal(t, expected.RecoveryAmounts.TotalRecovered, metrics.RecoveryAmounts.TotalRecovered)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())

			// Clean up
			sqlDB, _ := gormDB.DB()
			sqlDB.Close()
		})
	}
}

// Add more test functions for other methods
func TestCalculateRecoveryScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *RecoveryMetrics
		expected int
	}{
		{
			name: "high score",
			metrics: &RecoveryMetrics{
				RecoveryRate:        90.0,
				AverageRecoveryTime: 3600, // 1 hour
				RecoveryAmounts: Amounts{
					TotalRecovered: 9000,
					TotalFailed:    10000,
				},
			},
			expected: 95, // 45 (from rate) + 30 (from time) + 18 (from amount ratio)
		},
		{
			name: "low score",
			metrics: &RecoveryMetrics{
				RecoveryRate:        20.0,
				AverageRecoveryTime: 7200, // 2 hours
				RecoveryAmounts: Amounts{
					TotalRecovered: 1000,
					TotalFailed:    10000,
				},
			},
			expected: 25, // 10 (from rate) + 5 (from time) + 8 (from amount ratio)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, _ := MockRecoveryAnalyticsService(t)
			score, _ := service.calculateRecoveryScore(tt.metrics)
			assert.Equal(t, tt.expected, score)
		})
	}
}

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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

// TestPaymentFailureEvent is a SQLite-compatible version for testing
type TestPaymentFailureEvent struct {
	ID                string     `json:"id" gorm:"primary_key"`
	CompanyID         string     `json:"company_id" gorm:"not null;index"`
	ProviderID        string     `json:"provider_id" gorm:"not null"`
	EventID           string     `json:"event_id" gorm:"not null;uniqueIndex"`
	EventType         string     `json:"event_type" gorm:"not null"`
	PaymentIntentID   string     `json:"payment_intent_id"`
	TransactionID     string     `json:"transaction_id"`
	Amount            float64    `json:"amount"`
	Currency          string     `json:"currency" gorm:"default:'AUD'"`
	CustomerID        string     `json:"customer_id"`
	CustomerEmail     string     `json:"customer_email"`
	CustomerName      string     `json:"customer_name"`
	CustomerPhone     string     `json:"customer_phone"`
	Provider          string     `json:"provider"`
	RetryCount        int        `json:"retry_count" gorm:"default:0"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	FailureReason     string     `json:"failure_reason"`
	FailureCode       string     `json:"failure_code"`
	FailureMessage    string     `json:"failure_message"`
	Status            string     `json:"status" gorm:"default:'received'"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
	AlertedAt         *time.Time `json:"alerted_at,omitempty"`
	RawEventData      string     `json:"raw_event_data"`
	NormalizedData    string     `json:"normalized_data"`
	WebhookReceivedAt time.Time  `json:"webhook_received_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (TestPaymentFailureEvent) TableName() string {
	return "payment_failure_events"
}

// setupTestDB creates a SQLite database for testing (ephemeral approach)
func setupTestDB(t *testing.T) *gorm.DB {
	// Use SQLite in-memory database for ephemeral testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to create SQLite in-memory database")

	// Auto migrate the schema with test-compatible model
	err = db.AutoMigrate(&TestPaymentFailureEvent{})
	require.NoError(t, err)

	return db
}

func TestGetRecoveryMetrics(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	startTime := now.Add(-7 * 24 * time.Hour)
	endTime := now
	companyID := "test-company-123"

	t.Run("successful metrics calculation", func(t *testing.T) {
		// Setup SQLite in-memory database (secure, no external dependencies)
		db := setupTestDB(t)

		// Create service with real database
		logger, _ := zap.NewDevelopment()
		service := NewRecoveryAnalyticsService(db, logger)

		// Insert test data with proper EventID (SQLite-compatible)
		testEvents := []TestPaymentFailureEvent{
			// Failed payments
			{
				ID:            fmt.Sprintf("id_failed_1_%d", time.Now().UnixNano()),
				CompanyID:     companyID,
				ProviderID:    "stripe",
				EventID:       fmt.Sprintf("event_failed_1_%d", time.Now().UnixNano()),
				EventType:     "payment_intent.payment_failed",
				Amount:        60.0,
				FailureReason: "card_declined",
				Status:        "failed",
				CreatedAt:     startTime.Add(1 * time.Hour),
			},
			{
				ID:            fmt.Sprintf("id_failed_2_%d", time.Now().UnixNano()),
				CompanyID:     companyID,
				ProviderID:    "paypal",
				EventID:       fmt.Sprintf("event_failed_2_%d", time.Now().UnixNano()),
				EventType:     "payment_intent.payment_failed",
				Amount:        60.0,
				FailureReason: "insufficient_funds",
				Status:        "failed",
				CreatedAt:     startTime.Add(3 * time.Hour),
			},
			// Resolved payments
			{
				ID:            fmt.Sprintf("id_resolved_1_%d", time.Now().UnixNano()),
				CompanyID:     companyID,
				ProviderID:    "stripe",
				EventID:       fmt.Sprintf("event_resolved_1_%d", time.Now().UnixNano()),
				EventType:     "payment_intent.payment_failed",
				Amount:        60.0,
				FailureReason: "card_declined",
				Status:        "resolved",
				CreatedAt:     startTime.Add(2 * time.Hour),
			},
			{
				ID:            fmt.Sprintf("id_resolved_2_%d", time.Now().UnixNano()),
				CompanyID:     companyID,
				ProviderID:    "paypal",
				EventID:       fmt.Sprintf("event_resolved_2_%d", time.Now().UnixNano()),
				EventType:     "payment_intent.payment_failed",
				Amount:        60.0,
				FailureReason: "insufficient_funds",
				Status:        "resolved",
				CreatedAt:     startTime.Add(4 * time.Hour),
			},
		}

		// Add more failed payments to reach a reasonable total
		for i := 0; i < 96; i++ {
			testEvents = append(testEvents, TestPaymentFailureEvent{
				ID:            fmt.Sprintf("id_failed_bulk_%d_%d", i, time.Now().UnixNano()),
				CompanyID:     companyID,
				ProviderID:    "stripe",
				EventID:       fmt.Sprintf("event_failed_bulk_%d_%d", i, time.Now().UnixNano()),
				EventType:     "payment_intent.payment_failed",
				Amount:        52.08, // ~5000 total / 100
				FailureReason: "card_declined",
				Status:        "failed",
				CreatedAt:     startTime.Add(time.Duration(i) * time.Hour),
			})
		}

		// Insert all test data
		for _, event := range testEvents {
			err := db.Create(&event).Error
			require.NoError(t, err)
		}

		// Debug: Check if data was actually inserted
		var count int64
		err := db.Raw("SELECT COUNT(*) FROM payment_failure_events WHERE company_id = ?", companyID).Scan(&count).Error
		require.NoError(t, err)
		t.Logf("Debug - Total records inserted: %d", count)

		// Debug: Check failed vs resolved counts
		var failedCount, resolvedCount int64
		db.Raw("SELECT COUNT(*) FROM payment_failure_events WHERE company_id = ? AND status = 'failed'", companyID).Scan(&failedCount)
		db.Raw("SELECT COUNT(*) FROM payment_failure_events WHERE company_id = ? AND status = 'resolved'", companyID).Scan(&resolvedCount)
		t.Logf("Debug - Failed: %d, Resolved: %d", failedCount, resolvedCount)

		// Execute the service method
		metrics, err := service.GetRecoveryMetrics(ctx, companyID, startTime, endTime)

		// Debug: Print actual values
		t.Logf("Debug - Metrics returned:")
		t.Logf("  RecoveryRate: %f", metrics.RecoveryRate)
		t.Logf("  TotalFailed: %f", metrics.RecoveryAmounts.TotalFailed)
		t.Logf("  TotalRecovered: %f", metrics.RecoveryAmounts.TotalRecovered)
		t.Logf("  RecoveryByMethod length: %d", len(metrics.RecoveryByMethod))
		t.Logf("  RecoveryByFailureType length: %d", len(metrics.RecoveryByFailureType))

		// Assertions
		require.NoError(t, err)
		assert.True(t, metrics.RecoveryRate > 0, "Recovery rate should be positive")
		assert.True(t, metrics.RecoveryAmounts.TotalFailed > 5000, "Total failed should be around 5000+")
		assert.Equal(t, 120.0, metrics.RecoveryAmounts.TotalRecovered, "Total recovered should be 120.0")
		assert.True(t, len(metrics.RecoveryByMethod) > 0, "Should have recovery by method data")
		assert.True(t, len(metrics.RecoveryByFailureType) > 0, "Should have recovery by failure type data")
	})

	t.Run("no failed payments", func(t *testing.T) {
		// Setup SQLite in-memory database
		db := setupTestDB(t)

		// Create service with real database
		logger, _ := zap.NewDevelopment()
		service := NewRecoveryAnalyticsService(db, logger)

		// Execute with no data
		metrics, err := service.GetRecoveryMetrics(ctx, companyID, startTime, endTime)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, 0.0, metrics.RecoveryRate, "Recovery rate should be 0")
		assert.Equal(t, 0.0, metrics.RecoveryAmounts.TotalFailed, "Total failed should be 0")
		assert.Equal(t, 0.0, metrics.RecoveryAmounts.TotalRecovered, "Total recovered should be 0")
		assert.Equal(t, 0.0, metrics.RecoveryAmounts.TotalPending, "Total pending should be 0")
	})
}

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
					TotalFailed:    10000,
					TotalRecovered: 9000,
					TotalPending:   0,
				},
			},
			expected: 95, // 45 (from rate) + 30 (from time) + 18 (from amount ratio)
		},
		// Add more test cases
	}

	service := &RecoveryAnalyticsService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := service.calculateRecoveryScore(tt.metrics)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, score)
		})
	}
}

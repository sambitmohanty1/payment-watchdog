package services_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	svc "github.com/sambitmohanty1/payment-watchdog/api/internal/services"
)

// MockRecoveryOrchestrationService creates a test instance of RecoveryOrchestrationService with a mock DB
func MockRecoveryOrchestrationService(t *testing.T) (*svc.RecoveryOrchestrationService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"}) // Mock Redis client

	service := svc.NewRecoveryOrchestrationService(
		gormDB,
		nil, // RetryService
		nil, // CommunicationService
		nil, // AnalyticsService
		redisClient,
		logger,
	)

	return service, gormDB, mock
}

// ... Rest of the test file, assuming there are test functions ...

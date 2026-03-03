package services_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
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

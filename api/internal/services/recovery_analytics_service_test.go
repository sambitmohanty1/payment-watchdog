package services

import (
	"testing"

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
	t.Run("service creation test", func(t *testing.T) {
		service, gormDB, mock := MockRecoveryAnalyticsService(t)

		// Simple test - just verify service creation
		assert.NotNil(t, service)
		assert.NotNil(t, gormDB)
		assert.NotNil(t, mock)

		// Clean up
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	})
}

//go:build test

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	sqlite "gorm.io/driver/sqlite"
)

// setupTestDB creates a new in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// You can add any test-specific migrations here if needed

	return db
}

// setupTestLogger creates a test logger
func setupTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "Failed to create test logger")
	return logger
}

// loadTestEnv loads test environment variables
func loadTestEnv() error {
	testEnvPath := filepath.Join("config", "test.env")
	if _, err := os.Stat(testEnvPath); err == nil {
		return godotenv.Load(testEnvPath)
	}
	return nil
}

// TestMain is the entry point for tests
func TestMain(m *testing.M) {
	// Load test environment variables
	if err := loadTestEnv(); err != nil {
		panic(err)
	}

	// Run tests
	os.Exit(m.Run())
}

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// DistributedLockTestSuite provides comprehensive tests for distributed locking
type DistributedLockTestSuite struct {
	suite.Suite
	redisClient *redis.Client
	lockService *DistributedLockService
	logger      *zap.Logger
	ctx         context.Context
}

// SetupSuite runs once before all tests
func (suite *DistributedLockTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = zaptest.NewLogger(suite.T())

	// Connect to Redis (using test database)
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Adjust for your test environment
		Password: "",
		DB:       1, // Use separate DB for tests
	})

	// Test Redis connection
	err := suite.redisClient.Ping(suite.ctx).Err()
	require.NoError(suite.T(), err, "Failed to connect to Redis")

	suite.lockService = NewDistributedLockService(suite.redisClient, suite.logger)
}

// TearDownSuite runs once after all tests
func (suite *DistributedLockTestSuite) TearDownSuite() {
	// Clean up test data
	suite.redisClient.FlushDB(suite.ctx)
	suite.redisClient.Close()
}

// SetupTest runs before each test
func (suite *DistributedLockTestSuite) SetupTest() {
	// Clean up any remaining locks
	suite.redisClient.FlushDB(suite.ctx)
}

// TestAcquireLock tests basic lock acquisition
func (suite *DistributedLockTestSuite) TestAcquireLock() {
	resourceKey := "test_resource"
	ttl := 30 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), lock)
	assert.Equal(suite.T(), suite.lockService.prefix+resourceKey, lock.key)
	assert.NotEmpty(suite.T(), lock.value)
	assert.True(suite.T(), lock.acquiredAt.Before(time.Now().Add(time.Second)))
	assert.Equal(suite.T(), ttl, lock.ttl)

	// Verify lock exists in Redis
	exists := suite.redisClient.Exists(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), int64(1), exists)

	// Verify lock value matches
	value := suite.redisClient.Get(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), lock.value, value)

	// Verify TTL
	redisTTL := suite.redisClient.TTL(suite.ctx, lock.key).Val()
	assert.True(suite.T(), redisTTL > 20*time.Second) // Should be close to original TTL
}

// TestTryAcquireLock tests lock acquisition without retries
func (suite *DistributedLockTestSuite) TestTryAcquireLock() {
	resourceKey := "test_resource_try"
	ttl := 30 * time.Second

	// First acquisition should succeed
	lock1, err := suite.lockService.TryAcquireLock(suite.ctx, resourceKey, ttl)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), lock1)

	// Second acquisition should fail
	lock2, err := suite.lockService.TryAcquireLock(suite.ctx, resourceKey, ttl)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), lock2)
	assert.Contains(suite.T(), err.Error(), "lock is already held")
}

// TestReleaseLock tests lock release
func (suite *DistributedLockTestSuite) TestReleaseLock() {
	resourceKey := "test_resource_release"
	ttl := 30 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	// Verify lock exists
	exists := suite.redisClient.Exists(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), int64(1), exists)

	// Release lock
	err = lock.Release(suite.ctx)
	assert.NoError(suite.T(), err)

	// Verify lock is released
	exists = suite.redisClient.Exists(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), int64(0), exists)
}

// TestReleaseLockByWrongOwner tests release by non-owner
func (suite *DistributedLockTestSuite) TestReleaseLockByWrongOwner() {
	resourceKey := "test_resource_wrong_owner"
	ttl := 30 * time.Second

	lock1, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	// Create a fake lock with same key but different value
	fakeLock := &Lock{
		key:        lock1.key,
		value:      "fake_value",
		acquiredAt: time.Now(),
		ttl:        ttl,
		service:    suite.lockService,
	}

	// Attempt to release with fake lock should fail
	err = fakeLock.Release(suite.ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "lock was not held by this instance")

	// Original lock should still exist
	exists := suite.redisClient.Exists(suite.ctx, lock1.key).Val()
	assert.Equal(suite.T(), int64(1), exists)
}

// TestExtendLock tests lock TTL extension
func (suite *DistributedLockTestSuite) TestExtendLock() {
	resourceKey := "test_resource_extend"
	ttl := 10 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	// Check initial TTL
	redisTTL := suite.redisClient.TTL(suite.ctx, lock.key).Val()
	assert.True(suite.T(), redisTTL > 8*time.Second) // Should be close to original TTL

	// Extend lock
	additionalTTL := 20 * time.Second
	err = lock.Extend(suite.ctx, additionalTTL)
	assert.NoError(suite.T(), err)

	// Check extended TTL
	redisTTL = suite.redisClient.TTL(suite.ctx, lock.key).Val()
	assert.True(suite.T(), redisTTL > 18*time.Second) // Should be close to extended TTL
}

// TestIsHeld tests lock ownership check
func (suite *DistributedLockTestSuite) TestIsHeld() {
	resourceKey := "test_resource_isheld"
	ttl := 30 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	// Check if lock is held
	held, err := lock.IsHeld(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), held)

	// Release lock
	err = lock.Release(suite.ctx)
	assert.NoError(suite.T(), err)

	// Check if lock is still held
	held, err = lock.IsHeld(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), held)
}

// TestGetRemainingTTL tests TTL retrieval
func (suite *DistributedLockTestSuite) TestGetRemainingTTL() {
	resourceKey := "test_resource_ttl"
	ttl := 30 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	// Get remaining TTL
	remainingTTL, err := lock.GetRemainingTTL(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), remainingTTL > 25*time.Second) // Should be close to original TTL
	assert.True(suite.T(), remainingTTL <= ttl)
}

// TestAcquirePaymentFailureLock tests payment failure specific lock
func (suite *DistributedLockTestSuite) TestAcquirePaymentFailureLock() {
	paymentFailureID := uuid.New()

	lock, err := suite.lockService.AcquirePaymentFailureLock(suite.ctx, paymentFailureID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), lock)

	expectedKey := suite.lockService.prefix + "payment_failure:" + paymentFailureID.String()
	assert.Equal(suite.T(), expectedKey, lock.key)
}

// TestAcquireWorkflowExecutionLock tests workflow execution specific lock
func (suite *DistributedLockTestSuite) TestAcquireWorkflowExecutionLock() {
	executionID := uuid.New()

	lock, err := suite.lockService.AcquireWorkflowExecutionLock(suite.ctx, executionID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), lock)

	expectedKey := suite.lockService.prefix + "workflow_execution:" + executionID.String()
	assert.Equal(suite.T(), expectedKey, lock.key)
}

// TestConcurrentLockAcquisition tests concurrent lock acquisition
func (suite *DistributedLockTestSuite) TestConcurrentLockAcquisition() {
	resourceKey := "test_resource_concurrent"
	ttl := 30 * time.Second

	// Channel to signal completion
	done := make(chan bool, 2)

	// First goroutine acquires lock
	go func() {
		lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), lock)

		// Hold lock for 2 seconds
		time.Sleep(2 * time.Second)

		// Release lock
		err = lock.Release(suite.ctx)
		assert.NoError(suite.T(), err)

		done <- true
	}()

	// Second goroutine tries to acquire same lock
	go func() {
		// Wait a bit to ensure first goroutine gets the lock
		time.Sleep(500 * time.Millisecond)

		start := time.Now()
		lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
		elapsed := time.Since(start)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), lock)

		// Should have waited for first lock to be released (at least 1.5 seconds)
		assert.True(suite.T(), elapsed > 1*time.Second)

		// Release lock
		err = lock.Release(suite.ctx)
		assert.NoError(suite.T(), err)

		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestWithLock tests the WithLock helper function
func (suite *DistributedLockTestSuite) TestWithLock() {
	resourceKey := "test_resource_with_lock"
	executed := false

	err := suite.lockService.WithLock(suite.ctx, resourceKey, func(ctx context.Context) error {
		executed = true
		return nil
	}, 30*time.Second)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), executed)

	// Lock should be released
	locks, err := suite.lockService.GetActiveLocks(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), locks)
}

// TestWithPaymentFailureLock tests the WithPaymentFailureLock helper function
func (suite *DistributedLockTestSuite) TestWithPaymentFailureLock() {
	paymentFailureID := uuid.New()
	executed := false

	err := suite.lockService.WithPaymentFailureLock(suite.ctx, paymentFailureID, func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), executed)
}

// TestGetActiveLocks tests active locks retrieval
func (suite *DistributedLockTestSuite) TestGetActiveLocks() {
	// Create multiple locks
	locks := make([]*Lock, 3)
	for i := 0; i < 3; i++ {
		lock, err := suite.lockService.AcquireLock(suite.ctx, "test_resource_"+string(rune(i)), 30*time.Second)
		require.NoError(suite.T(), err)
		locks[i] = lock
	}

	// Get active locks
	activeLocks, err := suite.lockService.GetActiveLocks(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activeLocks, 3)

	// Clean up
	for _, lock := range locks {
		lock.Release(suite.ctx)
	}
}

// TestForceReleaseLock tests force release functionality
func (suite *DistributedLockTestSuite) TestForceReleaseLock() {
	resourceKey := "test_resource_force"

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, 30*time.Second)
	require.NoError(suite.T(), err)

	// Verify lock exists
	exists := suite.redisClient.Exists(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), int64(1), exists)

	// Force release
	err = suite.lockService.ForceReleaseLock(suite.ctx, resourceKey)
	assert.NoError(suite.T(), err)

	// Verify lock is released
	exists = suite.redisClient.Exists(suite.ctx, lock.key).Val()
	assert.Equal(suite.T(), int64(0), exists)
}

// TestCleanupExpiredLocks tests cleanup functionality
func (suite *DistributedLockTestSuite) TestCleanupExpiredLocks() {
	// Create a lock with very short TTL
	_, err := suite.lockService.AcquireLock(suite.ctx, "test_resource_cleanup", 1*time.Second)
	require.NoError(suite.T(), err)

	// Wait for lock to expire
	time.Sleep(2 * time.Second)

	// Run cleanup
	cleaned, err := suite.lockService.CleanupExpiredLocks(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), cleaned >= 1)
}

// TestLockInfo tests lock information retrieval
func (suite *DistributedLockTestSuite) TestLockInfo() {
	resourceKey := "test_resource_info"
	ttl := 30 * time.Second

	lock, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, ttl)
	require.NoError(suite.T(), err)

	info := lock.GetLockInfo()
	assert.Equal(suite.T(), lock.key, info["key"])
	assert.Equal(suite.T(), lock.value, info["value"])
	assert.Equal(suite.T(), lock.acquiredAt, info["acquired_at"])
	assert.Equal(suite.T(), lock.ttl, info["ttl"])
	assert.True(suite.T(), info["age"].(time.Duration) >= 0)
}

// TestContextCancellation tests context cancellation during lock acquisition
func (suite *DistributedLockTestSuite) TestContextCancellation() {
	resourceKey := "test_resource_cancel"

	// First acquire a lock
	lock1, err := suite.lockService.AcquireLock(suite.ctx, resourceKey, 30*time.Second)
	require.NoError(suite.T(), err)

	// Create a cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Try to acquire lock with short timeout
	start := time.Now()
	lock2, err := suite.lockService.AcquireLock(ctx, resourceKey, 30*time.Second)
	elapsed := time.Since(start)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), lock2)
	assert.True(suite.T(), elapsed < 1*time.Second) // Should timeout quickly
	assert.Contains(suite.T(), err.Error(), "context deadline exceeded")

	// Release first lock
	lock1.Release(suite.ctx)
}

// Test Distributed Lock Service Integration with Recovery Orchestration
func TestDistributedLockServiceIntegration(t *testing.T) {
	// This test requires a running Redis instance
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite.Run(t, new(DistributedLockTestSuite))
}

// BenchmarkLockAcquisition benchmarks lock acquisition performance
func BenchmarkLockAcquisition(b *testing.B) {
	logger := zaptest.NewLogger(b)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2, // Use separate DB for benchmarks
	})
	defer redisClient.Close()

	lockService := NewDistributedLockService(redisClient, logger)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			resourceKey := fmt.Sprintf("bench_resource_%d", i%10) // Use 10 different keys
			lock, err := lockService.TryAcquireLock(ctx, resourceKey, 30*time.Second)
			if err == nil {
				lock.Release(ctx)
			}
			i++
		}
	})
}

package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DistributedLockService provides Redis-based distributed locking
type DistributedLockService struct {
	redisClient *redis.Client
	logger      *zap.Logger
	prefix      string
	defaultTTL  time.Duration
	retryDelay  time.Duration
	maxRetries  int
}

// Lock represents a distributed lock
type Lock struct {
	key        string
	value      string
	acquiredAt time.Time
	ttl        time.Duration
	service    *DistributedLockService
}

// NewDistributedLockService creates a new distributed lock service
func NewDistributedLockService(redisClient *redis.Client, logger *zap.Logger) *DistributedLockService {
	// AC 3.2: Configure distributed_lock_service to use local Redis in sovereign mode
	if os.Getenv("SOVEREIGN_MODE") == "true" {
		localRedisHost := os.Getenv("REDIS_HOST")
		if localRedisHost != "" {
			redisPort := os.Getenv("REDIS_PORT")
			if redisPort == "" {
				redisPort = "6379"
			}
			redisPassword := os.Getenv("REDIS_PASSWORD")
			logger.Info("Sovereign Mode active: configuring DistributedLockService to use local Redis", zap.String("host", localRedisHost))

			// Replace the injected redis client with a local one
			redisClient = redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", localRedisHost, redisPort),
				Password: redisPassword,
				DB:       0,
			})
		}
	}

	return &DistributedLockService{
		redisClient: redisClient,
		logger:      logger,
		prefix:      "payment_watchdog:lock:",
		defaultTTL:  30 * time.Minute, // Default lock TTL
		retryDelay:  100 * time.Millisecond,
		maxRetries:  50, // Maximum retry attempts
	}
}

// AcquireLock acquires a distributed lock for a given resource
func (d *DistributedLockService) AcquireLock(ctx context.Context, resourceKey string, ttl time.Duration) (*Lock, error) {
	if ttl <= 0 {
		ttl = d.defaultTTL
	}

	lockKey := d.prefix + resourceKey
	lockValue := d.generateLockValue()

	d.logger.Debug("Attempting to acquire lock",
		zap.String("resource_key", resourceKey),
		zap.String("lock_key", lockKey),
		zap.Duration("ttl", ttl))

	// Use SET NX EX command for atomic lock acquisition
	result, err := d.redisClient.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		d.logger.Error("Failed to attempt lock acquisition",
			zap.String("resource_key", resourceKey),
			zap.Error(err))
		return nil, fmt.Errorf("redis SETNX failed: %w", err)
	}

	if !result {
		// Lock is already held, try to acquire with retries
		return d.acquireWithRetry(ctx, lockKey, lockValue, ttl, resourceKey)
	}

	d.logger.Info("Lock acquired successfully",
		zap.String("resource_key", resourceKey),
		zap.String("lock_value", lockValue))

	return &Lock{
		key:        lockKey,
		value:      lockValue,
		acquiredAt: time.Now(),
		ttl:        ttl,
		service:    d,
	}, nil
}

// acquireWithRetry attempts to acquire a lock with exponential backoff
func (d *DistributedLockService) acquireWithRetry(ctx context.Context, lockKey, lockValue string, ttl time.Duration, resourceKey string) (*Lock, error) {
	var lastErr error
	retryDelay := d.retryDelay

	for attempt := 1; attempt <= d.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}

		result, err := d.redisClient.SetNX(ctx, lockKey, lockValue, ttl).Result()
		if err != nil {
			lastErr = err
			d.logger.Warn("Retry attempt failed",
				zap.String("resource_key", resourceKey),
				zap.Int("attempt", attempt),
				zap.Error(err))
			continue
		}

		if result {
			d.logger.Info("Lock acquired after retries",
				zap.String("resource_key", resourceKey),
				zap.Int("attempt", attempt))

			return &Lock{
				key:        lockKey,
				value:      lockValue,
				acquiredAt: time.Now(),
				ttl:        ttl,
				service:    d,
			}, nil
		}

		// Exponential backoff with jitter
		retryDelay = time.Duration(float64(retryDelay) * 1.5)
		if retryDelay > 5*time.Second {
			retryDelay = 5 * time.Second
		}
		// Add jitter to prevent thundering herd
		jitterDuration, _ := rand.Int(rand.Reader, big.NewInt(int64(retryDelay)/2))
		jitter := time.Duration(jitterDuration.Int64())
		retryDelay += jitter

		d.logger.Debug("Lock still held, retrying",
			zap.String("resource_key", resourceKey),
			zap.Int("attempt", attempt),
			zap.Duration("next_retry_delay", retryDelay))
	}

	return nil, fmt.Errorf("failed to acquire lock after %d attempts: %w", d.maxRetries, lastErr)
}

// TryAcquireLock attempts to acquire a lock without retries
func (d *DistributedLockService) TryAcquireLock(ctx context.Context, resourceKey string, ttl time.Duration) (*Lock, error) {
	if ttl <= 0 {
		ttl = d.defaultTTL
	}

	lockKey := d.prefix + resourceKey
	lockValue := d.generateLockValue()

	result, err := d.redisClient.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis SETNX failed: %w", err)
	}

	if !result {
		return nil, errors.New("lock is already held")
	}

	return &Lock{
		key:        lockKey,
		value:      lockValue,
		acquiredAt: time.Now(),
		ttl:        ttl,
		service:    d,
	}, nil
}

// Release releases the distributed lock
func (l *Lock) Release(ctx context.Context) error {
	l.service.logger.Debug("Releasing lock",
		zap.String("key", l.key),
		zap.String("value", l.value))

	// Use Lua script for atomic release (check and delete)
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	result, err := l.service.redisClient.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		l.service.logger.Error("Failed to release lock",
			zap.String("key", l.key),
			zap.Error(err))
		return fmt.Errorf("failed to execute release script: %w", err)
	}

	if result.(int64) == 0 {
		l.service.logger.Warn("Lock was not held by this instance",
			zap.String("key", l.key),
			zap.String("expected_value", l.value))
		return errors.New("lock was not held by this instance")
	}

	l.service.logger.Info("Lock released successfully", zap.String("key", l.key))
	return nil
}

// Extend extends the lock TTL
func (l *Lock) Extend(ctx context.Context, additionalTTL time.Duration) error {
	l.service.logger.Debug("Extending lock",
		zap.String("key", l.key),
		zap.Duration("additional_ttl", additionalTTL))

	// Use Lua script for atomic extension (check and expire)
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("EXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end
	`

	result, err := l.service.redisClient.Eval(ctx, script, []string{l.key}, l.value, int(additionalTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to execute extend script: %w", err)
	}

	if result.(int64) == 0 {
		return errors.New("lock was not held by this instance")
	}

	l.ttl = additionalTTL
	l.service.logger.Debug("Lock extended successfully",
		zap.String("key", l.key),
		zap.Duration("new_ttl", l.ttl))

	return nil
}

// IsHeld checks if the lock is still held
func (l *Lock) IsHeld(ctx context.Context) (bool, error) {
	value, err := l.service.redisClient.Get(ctx, l.key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("failed to check lock: %w", err)
	}
	return value == l.value, nil
}

// GetRemainingTTL returns the remaining time-to-live for the lock
func (l *Lock) GetRemainingTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := l.service.redisClient.TTL(ctx, l.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// GetLockInfo returns information about the lock
func (l *Lock) GetLockInfo() map[string]interface{} {
	return map[string]interface{}{
		"key":         l.key,
		"value":       l.value,
		"acquired_at": l.acquiredAt,
		"ttl":         l.ttl,
		"age":         time.Since(l.acquiredAt),
	}
}

// generateLockValue creates a unique lock value
func (d *DistributedLockService) generateLockValue() string {
	// Use UUID + timestamp for uniqueness
	uuid := uuid.New().String()
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s:%d", uuid, timestamp)
}

// AcquirePaymentFailureLock acquires a lock specifically for a payment failure
func (d *DistributedLockService) AcquirePaymentFailureLock(ctx context.Context, paymentFailureID uuid.UUID) (*Lock, error) {
	resourceKey := fmt.Sprintf("payment_failure:%s", paymentFailureID.String())
	return d.AcquireLock(ctx, resourceKey, d.defaultTTL)
}

// TryAcquirePaymentFailureLock attempts to acquire a payment failure lock without retries
func (d *DistributedLockService) TryAcquirePaymentFailureLock(ctx context.Context, paymentFailureID uuid.UUID) (*Lock, error) {
	resourceKey := fmt.Sprintf("payment_failure:%s", paymentFailureID.String())
	return d.TryAcquireLock(ctx, resourceKey, d.defaultTTL)
}

// AcquireWorkflowExecutionLock acquires a lock specifically for a workflow execution
func (d *DistributedLockService) AcquireWorkflowExecutionLock(ctx context.Context, executionID uuid.UUID) (*Lock, error) {
	resourceKey := fmt.Sprintf("workflow_execution:%s", executionID.String())
	return d.AcquireLock(ctx, resourceKey, d.defaultTTL)
}

// TryAcquireWorkflowExecutionLock attempts to acquire a workflow execution lock without retries
func (d *DistributedLockService) TryAcquireWorkflowExecutionLock(ctx context.Context, executionID uuid.UUID) (*Lock, error) {
	resourceKey := fmt.Sprintf("workflow_execution:%s", executionID.String())
	return d.TryAcquireLock(ctx, resourceKey, d.defaultTTL)
}

// GetActiveLocks returns information about currently active locks (for monitoring)
func (d *DistributedLockService) GetActiveLocks(ctx context.Context) ([]map[string]interface{}, error) {
	pattern := d.prefix + "*"
	keys, err := d.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan lock keys: %w", err)
	}

	var locks []map[string]interface{}
	for _, key := range keys {
		value, err := d.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Lock expired
			}
			d.logger.Warn("Failed to get lock value", zap.String("key", key), zap.Error(err))
			continue
		}

		ttl, err := d.redisClient.TTL(ctx, key).Result()
		if err != nil {
			d.logger.Warn("Failed to get lock TTL", zap.String("key", key), zap.Error(err))
			ttl = -1
		}

		locks = append(locks, map[string]interface{}{
			"key":   key,
			"value": value,
			"ttl":   ttl,
		})
	}

	return locks, nil
}

// ForceReleaseLock forcefully releases a lock (for admin/recovery purposes)
func (d *DistributedLockService) ForceReleaseLock(ctx context.Context, resourceKey string) error {
	lockKey := d.prefix + resourceKey

	d.logger.Warn("Forcefully releasing lock", zap.String("key", lockKey))

	err := d.redisClient.Del(ctx, lockKey).Err()
	if err != nil {
		return fmt.Errorf("failed to force release lock: %w", err)
	}

	return nil
}

// CleanupExpiredLocks removes expired locks (maintenance task)
func (d *DistributedLockService) CleanupExpiredLocks(ctx context.Context) (int, error) {
	pattern := d.prefix + "*"
	keys, err := d.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to scan lock keys: %w", err)
	}

	cleaned := 0
	for _, key := range keys {
		ttl, err := d.redisClient.TTL(ctx, key).Result()
		if err != nil {
			d.logger.Warn("Failed to check lock TTL during cleanup", zap.String("key", key), zap.Error(err))
			continue
		}

		// If TTL is -1 (no expiration) or -2 (key doesn't exist), skip or clean up
		if ttl == -2 {
			// Key already expired
			cleaned++
			continue
		}

		if ttl == -1 {
			// Key has no expiration, this might be a problem
			d.logger.Warn("Lock has no expiration", zap.String("key", key))
			// Optionally set a reasonable TTL
			d.redisClient.Expire(ctx, key, d.defaultTTL)
		}
	}

	return cleaned, nil
}

// WithLock executes a function while holding a distributed lock
func (d *DistributedLockService) WithLock(ctx context.Context, resourceKey string, fn func(context.Context) error, ttl time.Duration) error {
	lock, err := d.AcquireLock(ctx, resourceKey, ttl)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	defer func() {
		if releaseErr := lock.Release(ctx); releaseErr != nil {
			d.logger.Error("Failed to release lock in defer",
				zap.String("resource_key", resourceKey),
				zap.Error(releaseErr))
		}
	}()

	return fn(ctx)
}

// WithPaymentFailureLock executes a function while holding a payment failure lock
func (d *DistributedLockService) WithPaymentFailureLock(ctx context.Context, paymentFailureID uuid.UUID, fn func(context.Context) error) error {
	resourceKey := fmt.Sprintf("payment_failure:%s", paymentFailureID.String())
	return d.WithLock(ctx, resourceKey, fn, d.defaultTTL)
}

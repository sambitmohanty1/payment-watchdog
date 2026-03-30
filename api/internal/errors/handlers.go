package errors

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// ErrorType categorizes different types of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeConnection    ErrorType = "connection"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeRuntime       ErrorType = "runtime"
	ErrorTypeBusiness      ErrorType = "business"
)

// ErrorContext provides context for error handling
type ErrorContext struct {
	Type      ErrorType
	Operation string
	Resource  string
	Retryable bool
	Timeout   time.Duration
	UserID    string
	RequestID string
}

// ErrorHandler defines the interface for error handlers
type ErrorHandler interface {
	Handle(ctx context.Context, err error, context ErrorContext) error
	ShouldRetry(err error) bool
	GetRetryDelay(attempt int) time.Duration
}

// RedisConnectionError handles Redis connection failures
type RedisConnectionError struct {
	logger *zap.Logger
}

func (h *RedisConnectionError) Handle(ctx context.Context, err error, context ErrorContext) error {
	h.logger.Error("Redis connection failed",
		zap.String("error_type", string(context.Type)),
		zap.String("operation", context.Operation),
		zap.String("resource", context.Resource),
		zap.Error(err),
		zap.Bool("retryable", context.Retryable),
		zap.Duration("timeout", context.Timeout),
	)

	if context.Retryable {
		h.logger.Info("Scheduling retry for Redis connection",
			zap.Time("retry_at", time.Now()),
			zap.Duration("delay", h.GetRetryDelay(1)),
		)
	}

	return err
}

func (h *RedisConnectionError) ShouldRetry(err error) bool {
	// Retry on connection refused, timeouts, and temporary failures
	if isConnectionError(err) {
		return true
	}
	return false
}

func (h *RedisConnectionError) GetRetryDelay(attempt int) time.Duration {
	// Exponential backoff: 1s, 2s, 4s, 8s, 16s, max 30s
	delay := time.Duration(attempt) * time.Second
	if delay > 30*time.Second {
		return 30 * time.Second
	}
	return delay
}

// ValidationError handles input validation failures
type ValidationError struct {
	logger *zap.Logger
}

func (h *ValidationError) Handle(ctx context.Context, err error, context ErrorContext) error {
	h.logger.Error("Input validation failed",
		zap.String("error_type", string(context.Type)),
		zap.String("operation", context.Operation),
		zap.String("resource", context.Resource),
		zap.Error(err),
		zap.Bool("retryable", false), // Validation errors should not retry
	)

	return err
}

func (h *ValidationError) ShouldRetry(err error) bool {
	return false
}

func (h *ValidationError) GetRetryDelay(attempt int) time.Duration {
	return 0 // No retry for validation errors
}

// ConfigurationError handles configuration failures
type ConfigurationError struct {
	logger *zap.Logger
}

func (h *ConfigurationError) Handle(ctx context.Context, err error, context ErrorContext) error {
	h.logger.Error("Configuration error",
		zap.String("error_type", string(context.Type)),
		zap.String("operation", context.Operation),
		zap.String("resource", context.Resource),
		zap.Error(err),
		zap.Bool("retryable", false), // Configuration errors should not retry
	)

	return err
}

func (h *ConfigurationError) ShouldRetry(err error) bool {
	return false
}

func (h *ConfigurationError) GetRetryDelay(attempt int) time.Duration {
	return 0 // No retry for configuration errors
}

// isConnectionError checks if error is connection-related
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"timeout",
		"network is unreachable",
		"no such host",
		"connection reset",
		"temporary failure",
	}

	for _, connErr := range connectionErrors {
		if contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr)))
}

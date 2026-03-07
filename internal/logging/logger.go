package logging

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds logging configuration
type Config struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
	Output string `mapstructure:"output"` // stdout, stderr
}

// DefaultConfig returns default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewLogger creates a new zap logger with the given configuration
func NewLogger(cfg Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	zapConfig.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))

	// Configure output
	if cfg.Output == "stderr" {
		zapConfig.OutputPaths = []string{"stderr"}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
	}

	// Disable stack trace for production
	if cfg.Format == "json" {
		zapConfig.DisableStacktrace = true
	}

	return zapConfig.Build()
}

// NewDevelopmentLogger creates a development-friendly logger
func NewDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	return config.Build()
}

// NewProductionLogger creates a production-ready logger
func NewProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.DisableStacktrace = true
	return config.Build()
}

// NewNopLogger creates a no-op logger for testing
func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}

// WithServiceContext adds service context to logger
func WithServiceContext(logger *zap.Logger, serviceName, version string) *zap.Logger {
	return logger.With(
		zap.String("service", serviceName),
		zap.String("version", version),
	)
}

// WithRequestContext adds request context to logger
func WithRequestContext(logger *zap.Logger, requestID, userID string) *zap.Logger {
	return logger.With(
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
	)
}

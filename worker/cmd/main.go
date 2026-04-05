package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
	"github.com/sambitmohanty1/payment-watchdog/worker/config"
	"github.com/sambitmohanty1/payment-watchdog/worker/internal/database"
	"github.com/sambitmohanty1/payment-watchdog/worker/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/worker/internal/services"
)

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// ZapLoggerAdapter adapts zap.Logger to LoggerInterface
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *ZapLoggerAdapter) Debug(msg string, fields ...interface{}) {
	z.logger.Debug(msg, convertFields(fields)...)
}

func (z *ZapLoggerAdapter) Info(msg string, fields ...interface{}) {
	z.logger.Info(msg, convertFields(fields)...)
}

func (z *ZapLoggerAdapter) Warn(msg string, fields ...interface{}) {
	z.logger.Warn(msg, convertFields(fields)...)
}

func (z *ZapLoggerAdapter) Error(msg string, fields ...interface{}) {
	z.logger.Error(msg, convertFields(fields)...)
}

func (z *ZapLoggerAdapter) Fatal(msg string, fields ...interface{}) {
	z.logger.Fatal(msg, convertFields(fields)...)
}

func (z *ZapLoggerAdapter) Sync() error {
	return z.logger.Sync()
}

// convertFields handles field conversion safely
func convertFields(fields []interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(fmt.Sprintf("field%d", i), field)
	}
	return zapFields
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Add service context
	logger = logger.With(zap.String("service", "payment-watchdog-worker"), zap.String("version", "1.0.0"))

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
		fx.Provide(
			// Provide WorkerConfig from environment variables
			func() *config.WorkerConfig {
				// Get database configuration from environment
				dbHost := getEnvOrDefault("DATABASE_HOST", "lexure-mvp-postgres")
				dbPort := getEnvIntOrDefault("DATABASE_PORT", 5403)
				dbName := getEnvOrDefault("DATABASE_NAME", "lexure_intelligence_mvp")
				dbUser := getEnvOrDefault("DATABASE_USER", "postgres")
				dbPassword := getEnvOrDefault("DATABASE_PASSWORD", "password")
				dbSSLMode := getEnvOrDefault("DATABASE_SSL_MODE", "disable")

				// Get Redis configuration from environment
				redisHost := getEnvOrDefault("REDIS_HOST", "lexure-redis-sovereign-au.sovereign-au.svc.cluster.local")
				redisPort := getEnvIntOrDefault("REDIS_PORT", 6379)
				redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")

				// Get logging configuration from environment
				logLevel := getEnvOrDefault("LOG_LEVEL", "info")

				// Get sovereign mode from environment
				sovereignMode := getEnvBoolOrDefault("SOVEREIGN_MODE", false)

				return &config.WorkerConfig{
					Database: &interfaces.DatabaseConfig{
						Host:     dbHost,
						Port:     dbPort,
						Name:     dbName,
						User:     dbUser,
						Password: dbPassword,
						SSLMode:  dbSSLMode,
					},
					Redis: &interfaces.RedisConfig{
						Host:     redisHost,
						Port:     redisPort,
						Password: redisPassword,
						DB:       0,
					},
					Logging: &interfaces.LoggingConfig{
						Level: logLevel,
					},
					Sovereign: sovereignMode,
				}
			},
			database.NewPostgresDatabase,
		),
		fx.Provide(
			// Provide interfaces as concrete types
			func(db *database.PostgresDatabase) interfaces.DatabaseInterface {
				return db
			},
			func(eb *eventbus.RedisEventBus) interfaces.EventBusInterface {
				return eb
			},
			func(logger *zap.Logger) interfaces.LoggerInterface {
				return &ZapLoggerAdapter{logger: logger}
			},
			// Provide DatabaseConfig from WorkerConfig
			func(cfg *config.WorkerConfig) *interfaces.DatabaseConfig {
				return cfg.Database
			},
			// Provide zap.Logger directly for database
			func() *zap.Logger {
				return logger
			},
			// Provide Redis configuration for EventBus
			func(logger *zap.Logger) (*eventbus.RedisEventBus, error) {
				// Read Redis configuration from environment variables
				redisAddr := os.Getenv("REDIS_ADDR")
				if redisAddr == "" {
					redisHost := os.Getenv("REDIS_HOST")
					redisPort := os.Getenv("REDIS_PORT")

					// Default values if environment variables are not set
					if redisHost == "" {
						redisHost = "lexure-redis-sovereign-au.sovereign-au.svc.cluster.local"
					}
					if redisPort == "" {
						redisPort = "6379"
					}

					redisAddr = redisHost + ":" + redisPort
				}

				redisPassword := os.Getenv("REDIS_PASSWORD")

				logger.Info("Connecting to Redis", zap.String("addr", redisAddr))
				return eventbus.NewRedisEventBus(redisAddr, redisPassword, 0, logger)
			},
		),
		fx.Provide(
			services.NewPaymentProcessorService,
		),
		fx.Invoke(func(processor *services.PaymentProcessorService) {
			processor.StartEventListeners(context.Background())
		}),
		fx.StopTimeout(30*time.Second),
	)

	if err := app.Start(context.Background()); err != nil {
		log.Fatal("Failed to start worker", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
}

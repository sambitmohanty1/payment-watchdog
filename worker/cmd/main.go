package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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
			config.Load,
			database.NewPostgresDatabase,
			eventbus.NewRedisEventBus,
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

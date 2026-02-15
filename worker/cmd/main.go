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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/internal/rules"
	"github.com/sambitmohanty1/payment-watchdog/internal/services"
)

func main() {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: zap.NewNop()}
		}),
		fx.Provide(
			config.Load,
			initLogger,
			initDatabase,
			rules.NewRuleEngineFactory,
			func(ref rules.RuleEngineFactory) rules.RuleEngine {
				return ref.CreateComprehensiveRuleEngine()
			},
			eventbus.NewRedisEventBus,
			services.NewEventProcessorService,
		),
		fx.Invoke(startWorker),
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
	if err := app.Stop(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Worker shutdown complete")
}

func initLogger() *zap.Logger {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	var logLevel zap.AtomicLevel
	switch level {
	case "debug":
		logLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		logLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		logLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config := zap.NewProductionConfig()
	config.Level = logLevel
	logger, _ := config.Build()
	return logger
}

func initDatabase(logger *zap.Logger) (*gorm.DB, error) {
	// Similar to original initDatabase
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DATABASE_PASSWORD")
	if password == "" {
		password = "password"
	}
	dbname := os.Getenv("DATABASE_NAME")
	if dbname == "" {
		dbname = "payment_watchdog"
	}
	port := os.Getenv("DATABASE_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

func startWorker(lc fx.Lifecycle, processor *services.EventProcessorService, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting Payment Watchdog Worker...")
			return processor.StartEventProcessing(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping Payment Watchdog Worker...")
			return processor.StopEventProcessing(ctx)
		},
	})
}

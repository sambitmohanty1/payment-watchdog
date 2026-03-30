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

	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/rules"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
)

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
			func() *zap.Logger { return logger },
			config.Load,
			initDatabase,
			func(logger *zap.Logger, db *gorm.DB) *rules.RuleEngineFactory {
				return rules.NewRuleEngineFactory(logger, db)
			},
			func(ref rules.RuleEngineFactory) rules.RuleEngine {
				return ref.CreateComprehensiveRuleEngine()
			},
			func(cfg *config.Config, logger *zap.Logger) (eventbus.EventBus, error) {
				redisConfig, err := validation.ValidateRedisConnection(logger, cfg)
				if err != nil {
					return nil, fmt.Errorf("failed to validate Redis connection: %w", err)
				}

				return eventbus.NewRedisEventBus(
					redisConfig.Address,  // Validated Redis address
					redisConfig.Password, // Validated Redis password
					redisConfig.DB,       // Validated Redis DB
					logger,
				)
			},
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
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "lexure-postgres-sovereign-au"
	}
	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DATABASE_PASSWORD")
	if password == "" {
		password = "postgres"
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
	// AC 1.3: Validate Sovereign Data Infrastructure Hardening
	cfg := config.Get()
	if cfg.SovereignMode && !cfg.IsSovereignCompliant() {
		logger.Fatal("Sovereign Compliance Check Failed. System configured with non-AU endpoints.",
			zap.String("db_host", cfg.Database.Host))
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting Payment Watchdog Worker...")

			// Start event processing with keep-alive and health server
			if err := processor.Start(ctx); err != nil {
				return fmt.Errorf("failed to start event processing: %w", err)
			}

			logger.Info("Worker started successfully with keep-alive and health monitoring")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping Payment Watchdog Worker...")
			return processor.Stop(ctx)
		},
	})
}

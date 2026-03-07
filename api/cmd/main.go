package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/sambitmohanty1/payment-watchdog/api/internal/api"
    "github.com/sambitmohanty1/payment-watchdog/api/internal/config"
    "github.com/sambitmohanty1/payment-watchdog/api/internal/database"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Fatal("Failed to load config", zap.Error(err))
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    if err := database.Migrate(db); err != nil {
        logger.Fatal("Failed to run migrations", zap.Error(err))
    }

    // 1. OBSERVABILITY: Metrics Server
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        logger.Info("Starting metrics server on :9090")
        http.ListenAndServe(":9090", nil)
    }()

    r := gin.Default()
    api.SetupRoutes(r, db, logger)

    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("listen: %s\n", zap.Error(err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}

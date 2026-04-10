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
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Failed to create logger")
		os.Exit(1)
	}
	defer logger.Sync()

	c, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		c.Database.Host, c.Database.User, c.Database.Password, c.Database.Name, c.Database.Port, c.Database.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// 1. OBSERVABILITY: Metrics Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("Starting metrics server on :9090")
		http.ListenAndServe(":9090", nil)
	}()

	// Initialize Server with full routing
	server := api.NewServer(logger, db, c)
	server.SetupRoutes()

	srv := &http.Server{
		Addr:    ":" + c.Server.Port,
		Handler: server.GetEngine(),
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

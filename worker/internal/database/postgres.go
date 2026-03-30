package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
)

// PostgresDatabase implements DatabaseInterface for PostgreSQL
type PostgresDatabase struct {
	db     *gorm.DB
	logger *zap.Logger
	config *interfaces.DatabaseConfig
}

// NewPostgresDatabase creates a new PostgreSQL database connection
func NewPostgresDatabase(config *interfaces.DatabaseConfig, logger *zap.Logger) *PostgresDatabase {
	return &PostgresDatabase{
		config: config,
		logger: logger,
	}
}

// Connect establishes database connection
func (p *PostgresDatabase) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.Name,
		p.config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		p.logger.Error("Failed to connect to database", zap.Error(err))
		return err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		p.logger.Error("Failed to get database instance", zap.Error(err))
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		p.logger.Error("Database connection test failed", zap.Error(err))
		return err
	}

	p.db = db
	p.logger.Info("Database connected successfully", zap.String("host", p.config.Host), zap.Int("port", p.config.Port))
	return nil
}

// Disconnect closes database connection
func (p *PostgresDatabase) Disconnect(ctx context.Context) error {
	if p.db != nil {
		if sqlDB, err := p.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				p.logger.Error("Failed to close database", zap.Error(err))
				return err
			}
		}
		p.db = nil
		p.logger.Info("Database disconnected")
	}
	return nil
}

// IsConnected returns connection status
func (p *PostgresDatabase) IsConnected() bool {
	if p.db == nil {
		return false
	}
	
	sqlDB, err := p.db.DB()
	if err != nil {
		return false
	}
	
	if err := sqlDB.Ping(); err != nil {
		return false
	}
	
	return true
}

// GetHealthStatus returns database health status
func (p *PostgresDatabase) GetHealthStatus() *interfaces.DatabaseHealthStatus {
	if p.IsConnected() {
		return &interfaces.DatabaseHealthStatus{
			IsConnected: true,
			LastCheck:   time.Now().Format(time.RFC3339),
			ErrorCount:   0,
		}
	}
	
	return &interfaces.DatabaseHealthStatus{
		IsConnected: false,
		LastCheck:   time.Now().Format(time.RFC3339),
		ErrorCount:   1,
	}
}

// GetDB returns underlying GORM instance
func (p *PostgresDatabase) GetDB() *gorm.DB {
	return p.db
}

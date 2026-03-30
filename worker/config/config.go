package config

import (
	"github.com/spf13/viper"
	"github.com/sambitmohanty1/payment-watchdog/shared/interfaces"
)

// WorkerConfig represents worker-specific configuration
type WorkerConfig struct {
	Database  *interfaces.DatabaseConfig  `mapstructure:"database"`
	Redis     *interfaces.RedisConfig     `mapstructure:"redis"`
	Logging   *interfaces.LoggingConfig   `mapstructure:"logging"`
	Sovereign bool                        `mapstructure:"sovereign_mode"`
}

// Load loads worker configuration from file
func Load() (*WorkerConfig, error) {
	viper.SetConfigName("worker")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	config := &WorkerConfig{}
	return config, viper.ReadInConfig()
}

// GetDatabaseConfig returns database configuration
func (c *WorkerConfig) GetDatabaseConfig() *interfaces.DatabaseConfig {
	return c.Database
}

// GetRedisConfig returns Redis configuration
func (c *WorkerConfig) GetRedisConfig() *interfaces.RedisConfig {
	return c.Redis
}

// GetLoggingConfig returns logging configuration
func (c *WorkerConfig) GetLoggingConfig() *interfaces.LoggingConfig {
	return c.Logging
}

// GetSovereignMode returns sovereign mode setting
func (c *WorkerConfig) GetSovereignMode() bool {
	return c.Sovereign
}

// IsProduction returns if running in production
func (c *WorkerConfig) IsProduction() bool {
	return viper.GetString("environment") == "production"
}

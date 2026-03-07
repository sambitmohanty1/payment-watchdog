package logging

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadConfig loads logging configuration from viper
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	// Set defaults
	viper.SetDefault("log.level", cfg.Level)
	viper.SetDefault("log.format", cfg.Format)
	viper.SetDefault("log.output", cfg.Output)

	// Bind environment variables
	if err := viper.BindEnv("log.level", "LOG_LEVEL"); err != nil {
		return cfg, fmt.Errorf("failed to bind LOG_LEVEL: %w", err)
	}
	if err := viper.BindEnv("log.format", "LOG_FORMAT"); err != nil {
		return cfg, fmt.Errorf("failed to bind LOG_FORMAT: %w", err)
	}
	if err := viper.BindEnv("log.output", "LOG_OUTPUT"); err != nil {
		return cfg, fmt.Errorf("failed to bind LOG_OUTPUT: %w", err)
	}

	// Read config
	cfg.Level = viper.GetString("log.level")
	cfg.Format = viper.GetString("log.format")
	cfg.Output = viper.GetString("log.output")

	return cfg, nil
}

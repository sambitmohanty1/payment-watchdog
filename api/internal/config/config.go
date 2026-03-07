package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for the service
type Config struct {
	Server        ServerConfig   `mapstructure:"server"`
	Database      DatabaseConfig `mapstructure:"database"`
	Stripe        StripeConfig   `mapstructure:"stripe"`
	Xero          XeroConfig     `mapstructure:"xero"`
	Email         EmailConfig    `mapstructure:"email"`
	Log           LogConfig      `mapstructure:"log"`
	SovereignMode bool           `mapstructure:"sovereign_mode"`
}

// isLocal checks if the host is a local loopback or standard internal network
func isLocal(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || strings.Contains(host, "lexure-mvp-postgres") || strings.Contains(host, "svc.cluster.local")
}

// IsSovereignCompliant checks if the infrastructure dependencies comply with AU residency laws.
func (c *Config) IsSovereignCompliant() bool {
	if !c.SovereignMode {
		return true
	}
	// GCP: .australia-southeast1, .australia-southeast2
	// AWS: .ap-southeast-2
	// OCI: .ap-sydney-1, .ap-melbourne-1
	// Azure: .australiaeast, .australiasoutheast
	host := c.Database.Host
	isAUCloud := strings.Contains(host, ".au") || strings.Contains(host, "ap-southeast-2") || strings.Contains(host, "ap-sydney-1") || strings.Contains(host, "ap-melbourne-1") || strings.Contains(host, "australia-southeast1") || strings.Contains(host, "australia-southeast2") || strings.Contains(host, "australiaeast") || strings.Contains(host, "australiasoutheast")
	return isAUCloud || isLocal(host)
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     string `mapstructure:"port"`
	Host     string `mapstructure:"host"`
	HTTPS    bool   `mapstructure:"https"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// StripeConfig holds Stripe configuration
type StripeConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	PublishableKey string `mapstructure:"publishable_key"`
}

// XeroConfig holds Xero configuration
type XeroConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider  string `mapstructure:"provider"` // smtp, sendgrid, etc.
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// Load loads configuration from file and environment variables
func Load() error {
	// Create logger for configuration loading
	logger, err := logging.NewDevelopmentLogger()
	if err != nil {
		return fmt.Errorf("failed to create config logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting configuration loading",
		zap.String("component", "config-loader"),
		zap.Time("started_at", time.Now()))

	// Set defaults to match Kubernetes service configuration
	viper.SetDefault("server.port", "8085")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.https", false)
	viper.SetDefault("server.cert_file", "./certs/server.crt")
	viper.SetDefault("server.key_file", "./certs/server.key")
	viper.SetDefault("database.host", "lexure-mvp-postgres")
	viper.SetDefault("database.port", 5403)
	viper.SetDefault("database.name", "lexure_intelligence_mvp")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("sovereign_mode", false)

	logger.Debug("Configuration defaults set",
		zap.String("component", "config-loader"))

	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config") // Kubernetes ConfigMap mount path
	viper.AddConfigPath(".")

	logger.Debug("Configuration paths set",
		zap.String("component", "config-loader"),
		zap.Strings("paths", []string{"./config", "/app/config", "."}))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error("Failed to read config file",
				zap.String("component", "config-loader"),
				zap.Error(err))
			return fmt.Errorf("failed to read config file: %w", err)
		}
		logger.Info("No config file found, using defaults and environment",
			zap.String("component", "config-loader"))
	} else {
		logger.Info("Config file read successfully",
			zap.String("component", "config-loader"),
			zap.String("config_file", viper.ConfigFileUsed()))
	}

	// Enable automatic environment variable loading
	viper.AutomaticEnv()
	logger.Debug("Automatic environment loading enabled",
		zap.String("component", "config-loader"))

	// Bind specific environment variables with proper error handling
	envVars := []struct {
		key, env string
	}{
		{"server.port", "SERVER_PORT"},
		{"server.host", "SERVER_HOST"},
		{"server.https", "SERVER_HTTPS"},
		{"server.cert_file", "SERVER_CERT_FILE"},
		{"server.key_file", "SERVER_KEY_FILE"},
		{"database.host", "DATABASE_HOST"},
		{"database.port", "DATABASE_PORT"},
		{"database.name", "DATABASE_NAME"},
		{"database.user", "DATABASE_USER"},
		{"database.password", "DATABASE_PASSWORD"},
		{"stripe.secret_key", "STRIPE_SECRET_KEY"},
		{"stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET"},
		{"email.host", "EMAIL_HOST"},
		{"email.port", "EMAIL_PORT"},
		{"email.username", "EMAIL_USERNAME"},
		{"email.password", "EMAIL_PASSWORD"},
		{"log.level", "LOG_LEVEL"},
		{"sovereign_mode", "SOVEREIGN_MODE"},
	}

	for _, envVar := range envVars {
		if err := viper.BindEnv(envVar.key, envVar.env); err != nil {
			logger.Error("Failed to bind environment variable",
				zap.String("component", "config-loader"),
				zap.String("key", envVar.key),
				zap.String("env", envVar.env),
				zap.Error(err))
			return fmt.Errorf("failed to bind %s: %w", envVar.env, err)
		}
	}

	logger.Debug("All environment variables bound",
		zap.String("component", "config-loader"),
		zap.Int("bound_vars", len(envVars)))

	logger.Info("Configuration loading completed successfully",
		zap.String("component", "config-loader"),
		zap.String("server_port", viper.GetString("server.port")),
		zap.String("database_host", viper.GetString("database.host")),
		zap.Bool("sovereign_mode", viper.GetBool("sovereign_mode")))

	return nil
}

// Get returns the current configuration
func Get() *Config {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal config: %v", err))
	}
	return &config
}

package config

import (
	"fmt"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Lock     LockConfig     `mapstructure:"lock"`
	Stripe   StripeConfig   `mapstructure:"stripe"`
	Xero     XeroConfig     `mapstructure:"xero"`
	Email    EmailConfig    `mapstructure:"email"`
	Log      LogConfig      `mapstructure:"log"`
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

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LockConfig holds distributed lock configuration
type LockConfig struct {
	DefaultTTL string `mapstructure:"default_ttl"` // Duration string (e.g., "30m")
	RetryDelay string `mapstructure:"retry_delay"` // Duration string (e.g., "100ms")
	MaxRetries int    `mapstructure:"max_retries"`
	Prefix     string `mapstructure:"prefix"`
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
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("lock.default_ttl", "30m")
	viper.SetDefault("lock.retry_delay", "100ms")
	viper.SetDefault("lock.max_retries", 50)
	viper.SetDefault("lock.prefix", "payment_watchdog:lock:")
	viper.SetDefault("log.level", "info")

	logger.Debug("Configuration defaults set",
		zap.String("component", "config-loader"),
		zap.String("server_port", viper.GetString("server.port")),
		zap.String("database_host", viper.GetString("database.host")))

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
		{"redis.host", "REDIS_HOST"},
		{"redis.port", "REDIS_PORT"},
		{"redis.password", "REDIS_PASSWORD"},
		{"redis.db", "REDIS_DB"},
		{"lock.default_ttl", "LOCK_DEFAULT_TTL"},
		{"lock.retry_delay", "LOCK_RETRY_DELAY"},
		{"lock.max_retries", "LOCK_MAX_RETRIES"},
		{"lock.prefix", "LOCK_PREFIX"},
		{"stripe.secret_key", "STRIPE_SECRET_KEY"},
		{"stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET"},
		{"email.host", "EMAIL_HOST"},
		{"email.port", "EMAIL_PORT"},
		{"email.username", "EMAIL_USERNAME"},
		{"email.password", "EMAIL_PASSWORD"},
		{"log.level", "LOG_LEVEL"},
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
		zap.String("database_name", viper.GetString("database.name")))

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

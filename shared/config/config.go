package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// ServiceConfig represents the unified configuration structure
// This replaces the fragmented configuration approach
type ServiceConfig struct {
	Service       ServiceInfo         `mapstructure:"service"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Sovereign     SovereignConfig     `mapstructure:"sovereign"`
	Workflow      WorkflowConfig      `mapstructure:"workflow"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
}

// ServiceInfo contains service-specific configuration
type ServiceInfo struct {
	Name        string `mapstructure:"name"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
	Host        string `mapstructure:"host"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Addr           string `mapstructure:"addr"`
	Password       string `mapstructure:"password"`
	DB             int    `mapstructure:"db"`
	PoolSize       int    `mapstructure:"pool_size"`
	ConnectTimeout string `mapstructure:"connect_timeout"`
	MaxRetries     int    `mapstructure:"max_retries"`
}

// ObservabilityConfig contains observability configuration
type ObservabilityConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

type TracingConfig struct {
	Enabled    bool    `mapstructure:"enabled"`
	SampleRate float64 `mapstructure:"sample_rate"`
	Exporter   string  `mapstructure:"exporter"`
	Endpoint   string  `mapstructure:"endpoint"`
}

// SovereignConfig contains sovereign compliance configuration
type SovereignConfig struct {
	Mode              bool     `mapstructure:"mode"`
	ValidateEndpoints bool     `mapstructure:"validate_endpoints"`
	AllowedRegions    []string `mapstructure:"allowed_regions"`
	DataResidency     string   `mapstructure:"data_residency"`
}

// WorkflowConfig contains workflow configuration
type WorkflowConfig struct {
	MaxConcurrentWorkflows int    `mapstructure:"max_concurrent_workflows"`
	MaxRetryAttempts       int    `mapstructure:"max_retry_attempts"`
	RetryDelay             string `mapstructure:"retry_delay"`
	Timeout                string `mapstructure:"timeout"`
}

// NotificationsConfig contains notification configuration
type NotificationsConfig struct {
	Email EmailConfig `mapstructure:"email"`
	SMS   SMSConfig   `mapstructure:"sms"`
}

type EmailConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	From          string `mapstructure:"from"`
	SMTPHost      string `mapstructure:"smtp_host"`
	SMTPPort      int    `mapstructure:"smtp_port"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	RetryAttempts int    `mapstructure:"retry_attempts"`
}

type SMSConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	From          string `mapstructure:"from"`
	Provider      string `mapstructure:"provider"`
	APIKey        string `mapstructure:"api_key"`
	RetryAttempts int    `mapstructure:"retry_attempts"`
}

// ConfigLoader handles loading configuration from multiple sources
type ConfigLoader struct {
	serviceName string
	environment string
	validators  []Validator
}

// Validator interface for configuration validation
type Validator interface {
	Validate(config *ServiceConfig) error
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(serviceName, environment string) *ConfigLoader {
	return &ConfigLoader{
		serviceName: serviceName,
		environment: environment,
		validators: []Validator{
			&ServiceValidator{},
			&DatabaseValidator{},
			&RedisValidator{},
			&SovereignValidator{},
		},
	}
}

// Load loads configuration from multiple sources with precedence
func (cl *ConfigLoader) Load() (*ServiceConfig, error) {
	config := &ServiceConfig{}

	// 1. Set up Viper with multiple sources
	v := viper.New()

	// Set configuration file name based on service
	v.SetConfigName(cl.serviceName)
	v.SetConfigType("yaml")

	// Add configuration search paths
	v.AddConfigPath("/etc/payment-watchdog")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Enable environment variable support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 2. Try to read configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, rely on environment variables
			fmt.Printf("Warning: Configuration file not found, using environment variables only\n")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 3. Unmarshal configuration
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 4. Override with environment variables (highest precedence)
	cl.overrideFromEnvironment(config)

	// 5. Set defaults
	cl.setDefaults(config)

	// 6. Validate configuration
	if err := cl.validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// overrideFromEnvironment overrides configuration with environment variables
func (cl *ConfigLoader) overrideFromEnvironment(config *ServiceConfig) {
	// Service configuration
	if port := os.Getenv("PORT"); port != "" {
		config.Service.Port = parseInt(port)
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Service.Environment = env
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Service.LogLevel = logLevel
	}

	// Database configuration
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		config.Database.Port = parseInt(port)
	}
	if name := os.Getenv("DATABASE_NAME"); name != "" {
		config.Database.Name = name
	}
	if user := os.Getenv("DATABASE_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DATABASE_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if sslMode := os.Getenv("DATABASE_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Redis configuration
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Redis.Addr = addr
	} else {
		// Build from host and port if addr not provided
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if host != "" && port != "" {
			config.Redis.Addr = fmt.Sprintf("%s:%s", host, port)
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}

	// Sovereign configuration
	if sovereignMode := os.Getenv("SOVEREIGN_MODE"); sovereignMode != "" {
		config.Sovereign.Mode = parseBool(sovereignMode)
	}
}

// setDefaults sets reasonable defaults for configuration
func (cl *ConfigLoader) setDefaults(config *ServiceConfig) {
	if config.Service.Name == "" {
		config.Service.Name = cl.serviceName
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8080
	}
	if config.Service.Environment == "" {
		config.Service.Environment = cl.environment
	}
	if config.Service.LogLevel == "" {
		config.Service.LogLevel = "info"
	}

	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "require"
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 100
	}

	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 100
	}
	if config.Redis.ConnectTimeout == "" {
		config.Redis.ConnectTimeout = "10s"
	}
	if config.Redis.MaxRetries == 0 {
		config.Redis.MaxRetries = 3
	}

	if config.Workflow.MaxConcurrentWorkflows == 0 {
		config.Workflow.MaxConcurrentWorkflows = 100
	}
	if config.Workflow.MaxRetryAttempts == 0 {
		config.Workflow.MaxRetryAttempts = 3
	}
	if config.Workflow.RetryDelay == "" {
		config.Workflow.RetryDelay = "5m"
	}
}

// validate validates the configuration using all registered validators
func (cl *ConfigLoader) validate(config *ServiceConfig) error {
	for _, validator := range cl.validators {
		if err := validator.Validate(config); err != nil {
			return err
		}
	}
	return nil
}

// Helper functions
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

func parseBool(s string) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return false
}

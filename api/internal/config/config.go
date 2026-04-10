package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for the service
type Config struct {
	Server        ServerConfig   `mapstructure:"server"`
	Database      DatabaseConfig `mapstructure:"database"`
	Redis         RedisConfig    `mapstructure:"redis"`
	Stripe        StripeConfig   `mapstructure:"stripe"`
	Xero          XeroConfig     `mapstructure:"xero"`
	Email         EmailConfig    `mapstructure:"email"`
	Log           LogConfig      `mapstructure:"log"`
	SovereignMode bool           `mapstructure:"sovereign_mode"`
}

// isLocal checks if the host is a local loopback or standard internal network
func isLocal(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || strings.Contains(host, "postgres-sovereign-au") || strings.Contains(host, "svc.cluster.local")
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

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
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

// NewLogger creates a new logger for configuration loading
func NewLogger() (*zap.Logger, error) {
	return logging.NewDevelopmentLogger()
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	setDefaults()
	setupPaths()

	if err := readConfigFile(); err != nil {
		return nil, err
	}

	if err := bindEnvironment(); err != nil {
		return nil, err
	}

	return Get(), nil
}

func setDefaults() {
	// Set defaults to match Kubernetes service configuration
	viper.SetDefault("server.port", "8085")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.https", false)
	viper.SetDefault("server.cert_file", "./certs/server.crt")
	viper.SetDefault("server.key_file", "./certs/server.key")
	viper.SetDefault("database.host", "postgres-sovereign-au")
	viper.SetDefault("database.port", 5403)
	viper.SetDefault("database.name", "payment_watchdog")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("redis.host", "redis-sovereign-au")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("sovereign_mode", false)
}

func setupPaths() {
	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config") // Kubernetes ConfigMap mount path
	viper.AddConfigPath(".")
}

func readConfigFile() error {
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	return nil
}

func bindEnvironment() error {
	// Enable automatic environment variable loading
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Explicitly bind the critical connection details
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.name", "DATABASE_NAME")
	viper.BindEnv("database.port", "DATABASE_PORT")


	// Check for potential environment variable conflicts
	if err := checkEnvironmentConflicts(); err != nil {
		// Don't fail startup - just log the warning
		// In production, this would be logged to monitoring
	}

	return nil
}

func checkEnvironmentConflicts() error {
	// Check for potential environment variable conflicts
	conflicts := []struct {
		standard, legacy string
		description      string
	}{
		{"DATABASE_HOST", "DB_HOST", "Database hostname - both DATABASE_HOST and DB_HOST set"},
		{"DATABASE_USER", "DB_USER", "Database user - both DATABASE_USER and DB_USER set"},
		{"DATABASE_PASSWORD", "DB_PASSWORD", "Database password - both DATABASE_PASSWORD and DB_PASSWORD set"},
		{"DATABASE_NAME", "DB_NAME", "Database name - both DATABASE_NAME and DB_NAME set"},
		{"DATABASE_PORT", "DB_PORT", "Database port - both DATABASE_PORT and DB_PORT set"},
	}

	var conflictStrings []string
	for _, conflict := range conflicts {
		standard := os.Getenv(conflict.standard)
		legacy := os.Getenv(conflict.legacy)

		if standard != "" && legacy != "" {
			// Both are set - log warning about potential confusion
			conflictStrings = append(conflictStrings,
				fmt.Sprintf("CONFLICT: %s (using %s as priority)", conflict.description, conflict.standard))
		}
	}

	if len(conflictStrings) > 0 {
		// Return warning message instead of error to allow startup
		return fmt.Errorf("environment variable conflicts detected: %s", strings.Join(conflictStrings, "; "))
	}

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

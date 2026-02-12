package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
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
	fmt.Println("🔍 CONFIG DEBUG: Starting configuration loading...")

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

	fmt.Println("🔍 CONFIG DEBUG: Defaults set")

	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config") // Kubernetes ConfigMap mount path
	viper.AddConfigPath(".")

	fmt.Println("🔍 CONFIG DEBUG: Config paths set")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("🔍 CONFIG DEBUG: Failed to read config file: %v\n", err)
			return fmt.Errorf("failed to read config file: %w", err)
		}
		fmt.Println("🔍 CONFIG DEBUG: No config file found, using defaults and environment")
	} else {
		fmt.Println("🔍 CONFIG DEBUG: Config file read successfully")
	}

	// Enable automatic environment variable loading
	viper.AutomaticEnv()
	fmt.Println("🔍 CONFIG DEBUG: Automatic environment loading enabled")

	// Bind specific environment variables with proper error handling
	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind SERVER_PORT: %v\n", err)
		return fmt.Errorf("failed to bind SERVER_PORT: %w", err)
	}
	if err := viper.BindEnv("server.host", "SERVER_HOST"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind SERVER_HOST: %v\n", err)
		return fmt.Errorf("failed to bind SERVER_HOST: %w", err)
	}
	if err := viper.BindEnv("server.https", "SERVER_HTTPS"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind SERVER_HTTPS: %v\n", err)
		return fmt.Errorf("failed to bind SERVER_HTTPS: %w", err)
	}
	if err := viper.BindEnv("server.cert_file", "SERVER_CERT_FILE"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind SERVER_CERT_FILE: %v\n", err)
		return fmt.Errorf("failed to bind SERVER_CERT_FILE: %w", err)
	}
	if err := viper.BindEnv("server.key_file", "SERVER_KEY_FILE"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind SERVER_KEY_FILE: %v\n", err)
		return fmt.Errorf("failed to bind SERVER_KEY_FILE: %w", err)
	}
	if err := viper.BindEnv("database.host", "DATABASE_HOST"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind DATABASE_HOST: %v\n", err)
		return fmt.Errorf("failed to bind DATABASE_HOST: %w", err)
	}
	if err := viper.BindEnv("database.port", "DATABASE_PORT"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind DATABASE_PORT: %v\n", err)
		return fmt.Errorf("failed to bind DATABASE_PORT: %v", err)
	}
	if err := viper.BindEnv("database.name", "DATABASE_NAME"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind DATABASE_NAME: %v\n", err)
		return fmt.Errorf("failed to bind DATABASE_NAME: %w", err)
	}
	if err := viper.BindEnv("database.user", "DATABASE_USER"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind DATABASE_USER: %v\n", err)
		return fmt.Errorf("failed to bind DATABASE_USER: %w", err)
	}
	if err := viper.BindEnv("database.password", "DATABASE_PASSWORD"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind DATABASE_PASSWORD: %v\n", err)
		return fmt.Errorf("failed to bind DATABASE_PASSWORD: %w", err)
	}
	if err := viper.BindEnv("stripe.secret_key", "STRIPE_SECRET_KEY"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind STRIPE_SECRET_KEY: %v\n", err)
		return fmt.Errorf("failed to bind STRIPE_SECRET_KEY: %w", err)
	}
	if err := viper.BindEnv("stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind STRIPE_WEBHOOK_SECRET: %v\n", err)
		return fmt.Errorf("failed to bind STRIPE_WEBHOOK_SECRET: %w", err)
	}
	if err := viper.BindEnv("email.host", "EMAIL_HOST"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind EMAIL_HOST: %v\n", err)
		return fmt.Errorf("failed to bind EMAIL_HOST: %v", err)
	}
	if err := viper.BindEnv("email.port", "EMAIL_PORT"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind EMAIL_PORT: %v\n", err)
		return fmt.Errorf("failed to bind EMAIL_PORT: %v", err)
	}
	if err := viper.BindEnv("email.username", "EMAIL_USERNAME"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind EMAIL_USERNAME: %v\n", err)
		return fmt.Errorf("failed to bind EMAIL_USERNAME: %v", err)
	}
	if err := viper.BindEnv("email.password", "EMAIL_PASSWORD"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind EMAIL_PASSWORD: %v\n", err)
		return fmt.Errorf("failed to bind EMAIL_PASSWORD: %v", err)
	}
	if err := viper.BindEnv("log.level", "LOG_LEVEL"); err != nil {
		fmt.Printf("🔍 CONFIG DEBUG: Failed to bind LOG_LEVEL: %v\n", err)
		return fmt.Errorf("failed to bind LOG_LEVEL: %v", err)
	}

	fmt.Println("🔍 CONFIG DEBUG: All environment variables bound")

	// Log final configuration values
	fmt.Printf("🔍 CONFIG DEBUG: Final configuration values:\n")
	fmt.Printf("  server.port: %s\n", viper.GetString("server.port"))
	fmt.Printf("  server.host: %s\n", viper.GetString("server.host"))
	fmt.Printf("  server.https: %t\n", viper.GetBool("server.https"))
	fmt.Printf("  server.cert_file: %s\n", viper.GetString("server.cert_file"))
	fmt.Printf("  server.key_file: %s\n", viper.GetString("server.key_file"))
	fmt.Printf("  database.host: %s\n", viper.GetString("database.host"))
	fmt.Printf("  database.port: %d\n", viper.GetInt("database.port"))
	fmt.Printf("  database.name: %s\n", viper.GetString("database.name"))
	fmt.Printf("  database.user: %s\n", viper.GetString("database.user"))
	fmt.Printf("  database.ssl_mode: %s\n", viper.GetString("database.ssl_mode"))

	fmt.Println("🔍 CONFIG DEBUG: Configuration loading completed successfully")
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

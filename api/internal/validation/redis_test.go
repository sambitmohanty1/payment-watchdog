package validation

import (
	"testing"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/logging"
	"go.uber.org/zap"
)

func TestValidateRedisConnection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "lexure-redis-sovereign-au",
					Port:     6379,
					Password: "",
					DB:       0,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid host - localhost",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "localhost",
					Port:     6379,
					Password: "",
					DB:       0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateRedisConnection(logger, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRedisConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	config := logging.SanitizeConfig{
		RedactPasswords:  true,
		RedactHosts:     true,
		RedactEmails:    true,
		MaxStringLength: 50,
	}

	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "password field",
			key:      "REDIS_PASSWORD",
			value:    "secret123",
			expected: "[REDACTED]",
		},
		{
			name:     "host field",
			key:      "REDIS_HOST",
			value:    "localhost",
			expected: "[LOCALHOST]",
		},
		{
			name:     "regular field",
			key:      "REDIS_PORT",
			value:    "6379",
			expected: "6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logging.SanitizeForLogging(logger, tt.key, tt.value, config)
			if result != tt.expected {
				t.Errorf("SanitizeForLogging() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

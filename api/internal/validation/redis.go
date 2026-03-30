package validation

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"go.uber.org/zap"
)

// RedisConnectionConfig holds validated Redis connection parameters
type RedisConnectionConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Address  string
}

// ValidateRedisConnection validates Redis connection parameters and returns a safe config
func ValidateRedisConnection(logger *zap.Logger, cfg *config.Config) (*RedisConnectionConfig, error) {
	redisConfig := &RedisConnectionConfig{}
	
	// Get Redis configuration with fallback to defaults
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = cfg.Redis.Host
		if redisHost == "" {
			redisHost = "lexure-redis-sovereign-au"
		}
	}
	
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = fmt.Sprintf("%d", cfg.Redis.Port)
		if redisPort == "0" {
			redisPort = "6379"
		}
	}
	
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = cfg.Redis.Password
	}
	
	redisDB := cfg.Redis.DB
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err != nil {
			logger.Warn("Invalid REDIS_DB, using config default", 
				zap.String("provided", dbStr),
				zap.Int("default", cfg.Redis.DB),
				zap.Error(err))
		} else {
			redisDB = db
		}
	}
	
	// Validate host
	if err := validateHost(redisHost); err != nil {
		return nil, fmt.Errorf("invalid Redis host: %w", err)
	}
	
	// Validate port
	if err := validatePort(redisPort); err != nil {
		return nil, fmt.Errorf("invalid Redis port: %w", err)
	}
	
	// Validate DB
	if redisDB < 0 || redisDB > 15 {
		return nil, fmt.Errorf("invalid Redis DB: %d (must be 0-15)", redisDB)
	}
	
	// Construct address
	address := fmt.Sprintf("%s:%s", redisHost, redisPort)
	
	redisConfig.Host = redisHost
	redisConfig.Port = redisPort
	redisConfig.Password = redisPassword
	redisConfig.DB = redisDB
	redisConfig.Address = address
	
	logger.Info("Redis connection validated",
		zap.String("host", redisHost),
		zap.String("port", redisPort),
		zap.Int("db", redisDB),
		zap.String("address", address))
	
	return redisConfig, nil
}

// validateHost validates Redis hostname
func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	
	// Check for localhost variations that shouldn't be used in production
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return fmt.Errorf("localhost not allowed in production")
	}
	
	// Basic hostname validation
	if len(host) > 253 {
		return fmt.Errorf("host too long (max 253 characters)")
	}
	
	// Check for valid characters
	for _, char := range host {
		if !isValidHostChar(char) {
			return fmt.Errorf("invalid character in host: %c", char)
		}
	}
	
	return nil
}

// validatePort validates Redis port
func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be numeric: %w", err)
	}
	
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be 1-65535: %d", portNum)
	}
	
	// Warn about privileged ports
	if portNum < 1024 {
		return fmt.Errorf("privileged port not recommended: %d", portNum)
	}
	
	return nil
}

// isValidHostChar checks if character is valid in hostname
func isValidHostChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		   (char >= 'A' && char <= 'Z') ||
		   (char >= '0' && char <= '9') ||
		   char == '-' || char == '.' || char == '_'
}

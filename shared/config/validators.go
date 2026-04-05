package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ServiceValidator validates service configuration
type ServiceValidator struct{}

func (v *ServiceValidator) Validate(config *ServiceConfig) error {
	if config.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	
	if config.Service.Port <= 0 || config.Service.Port > 65535 {
		return fmt.Errorf("service port must be between 1 and 65535")
	}
	
	if config.Service.Environment == "" {
		return fmt.Errorf("service environment is required")
	}
	
	validEnvironments := []string{"development", "staging", "production", "sovereign-au"}
	envValid := false
	for _, env := range validEnvironments {
		if config.Service.Environment == env {
			envValid = true
			break
		}
	}
	if !envValid {
		return fmt.Errorf("invalid environment: %s. Must be one of: %v", config.Service.Environment, validEnvironments)
	}
	
	return nil
}

// DatabaseValidator validates database configuration
type DatabaseValidator struct{}

func (v *DatabaseValidator) Validate(config *ServiceConfig) error {
	db := config.Database
	
	if db.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if db.Port <= 0 || db.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	
	if db.Name == "" {
		return fmt.Errorf("database name is required")
	}
	
	if db.User == "" {
		return fmt.Errorf("database user is required")
	}
	
	if db.Password == "" {
		return fmt.Errorf("database password is required")
	}
	
	// Validate SSL mode
	validSSLModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
	sslValid := false
	for _, mode := range validSSLModes {
		if db.SSLMode == mode {
			sslValid = true
			break
		}
	}
	if !sslValid {
		return fmt.Errorf("invalid database SSL mode: %s. Must be one of: %v", db.SSLMode, validSSLModes)
	}
	
	return nil
}

// RedisValidator validates Redis configuration
type RedisValidator struct{}

func (v *RedisValidator) Validate(config *ServiceConfig) error {
	redis := config.Redis
	
	if redis.Addr == "" {
		return fmt.Errorf("Redis address is required")
	}
	
	// Validate Redis address format
	if !strings.Contains(redis.Addr, ":") {
		return fmt.Errorf("Redis address must be in format host:port")
	}
	
	parts := strings.Split(redis.Addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("Redis address must be in format host:port")
	}
	
	port := parts[1]
	if len(port) == 0 {
		return fmt.Errorf("Redis port is required")
	}
	
	// Validate Redis DB
	if redis.DB < 0 || redis.DB > 15 {
		return fmt.Errorf("Redis DB must be between 0 and 15")
	}
	
	return nil
}

// SovereignValidator validates sovereign compliance configuration
type SovereignValidator struct{}

func (v *SovereignValidator) Validate(config *ServiceConfig) error {
	sovereign := config.Sovereign
	
	// If sovereign mode is enabled, perform additional validations
	if sovereign.Mode {
		if sovereign.DataResidency == "" {
			return fmt.Errorf("data residency is required when sovereign mode is enabled")
		}
		
		// Validate data residency is Australia
		if !strings.EqualFold(sovereign.DataResidency, "australia") && 
		   !strings.EqualFold(sovereign.DataResidency, "au") {
			return fmt.Errorf("data residency must be 'australia' or 'au' when sovereign mode is enabled")
		}
		
		// Validate database host for sovereign compliance
		if err := v.validateEndpointSovereignty(config.Database.Host, "database"); err != nil {
			return err
		}
		
		// Validate Redis host for sovereign compliance
		if redisHost := extractHostFromAddr(config.Redis.Addr); redisHost != "" {
			if err := v.validateEndpointSovereignty(redisHost, "redis"); err != nil {
				return err
			}
		}
		
		// Validate allowed regions
		if len(sovereign.AllowedRegions) == 0 {
			// Default to Australian regions
			sovereign.AllowedRegions = []string{"ap-southeast-2", "ap-southeast-3"}
		} else {
			for _, region := range sovereign.AllowedRegions {
				if !strings.HasPrefix(region, "ap-") {
					return fmt.Errorf("allowed region '%s' is not in Asia Pacific", region)
				}
			}
		}
	}
	
	return nil
}

// validateEndpointSovereignty validates that an endpoint complies with sovereign requirements
func (v *SovereignValidator) validateEndpointSovereignty(endpoint, serviceType string) error {
	// Skip validation for localhost and cluster-local services
	if strings.Contains(endpoint, "localhost") || 
	   strings.Contains(endpoint, "127.0.0.1") ||
	   strings.Contains(endpoint, ".svc.cluster.local") {
		return nil
	}
	
	// Extract hostname from endpoint
	host := extractHostFromAddr(endpoint)
	if host == "" {
		return fmt.Errorf("invalid %s endpoint format: %s", serviceType, endpoint)
	}
	
	// Check for non-Australian cloud providers
	nonAustralianPatterns := []string{
		`.*\.us-.*\.amazonaws\.com$`,
		`.*\.eu-.*\.amazonaws\.com$`,
		`.*\.us-.*\.googleapis\.com$`,
		`.*\.eu-.*\.googleapis\.com$`,
		`.*\.us-.*\.azure\.com$`,
		`.*\.eu-.*\.azure\.com$`,
	}
	
	for _, pattern := range nonAustralianPatterns {
		matched, err := regexp.MatchString(pattern, host)
		if err != nil {
			return fmt.Errorf("error validating sovereign compliance for %s: %w", serviceType, err)
		}
		if matched {
			return fmt.Errorf("%s endpoint '%s' violates sovereign compliance - non-Australian region detected", serviceType, endpoint)
		}
	}
	
	return nil
}

// extractHostFromAddr extracts hostname from address
func extractHostFromAddr(addr string) string {
	// Remove port if present
	if strings.Contains(addr, ":") {
		parts := strings.Split(addr, ":")
		if len(parts) >= 2 {
			addr = parts[0]
		}
	}
	
	// Parse URL if it looks like one
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		if u, err := url.Parse(addr); err == nil {
			return u.Hostname()
		}
	}
	
	return addr
}

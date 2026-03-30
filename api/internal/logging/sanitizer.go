package logging

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// SanitizeConfig sanitizes configuration values for logging
type SanitizeConfig struct {
	RedactPasswords bool
	RedactHosts     bool
	RedactEmails    bool
	MaxStringLength int
}

// SanitizeForLogging returns a sanitized version of sensitive config values
func SanitizeForLogging(logger *zap.Logger, key, value string, config SanitizeConfig) string {
	sanitized := value

	if config.RedactPasswords && isPasswordField(key) {
		sanitized = "[REDACTED]"
	} else if config.RedactHosts && isHostField(key) {
		sanitized = sanitizeHost(value)
	} else if config.RedactEmails && isEmailField(key) {
		sanitized = sanitizeEmail(value)
	} else if config.MaxStringLength > 0 && len(sanitized) > config.MaxStringLength {
		sanitized = fmt.Sprintf("%s...", sanitized[:config.MaxStringLength])
	}

	return sanitized
}

// isPasswordField checks if a key contains password-related terms
func isPasswordField(key string) bool {
	lowerKey := strings.ToLower(key)
	passwordTerms := []string{"password", "secret", "token", "key", "auth"}

	for _, term := range passwordTerms {
		if strings.Contains(lowerKey, term) {
			return true
		}
	}
	return false
}

// isHostField checks if a key contains host-related terms
func isHostField(key string) bool {
	lowerKey := strings.ToLower(key)
	hostTerms := []string{"host", "url", "endpoint", "address", "server"}

	for _, term := range hostTerms {
		if strings.Contains(lowerKey, term) {
			return true
		}
	}
	return false
}

// isEmailField checks if a key contains email-related terms
func isEmailField(key string) bool {
	lowerKey := strings.ToLower(key)
	emailTerms := []string{"email", "mail", "from", "smtp"}

	for _, term := range emailTerms {
		if strings.Contains(lowerKey, term) {
			return true
		}
	}
	return false
}

// sanitizeHost removes sensitive information from host values
func sanitizeHost(host string) string {
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return "[LOCALHOST]"
	}

	// Remove potential credentials from host strings
	if idx := strings.Index(host, "@"); idx >= 0 && idx < len(host)-1 {
		// Check if there's a password-like pattern after @
		remaining := host[idx+1:]
		if containsPasswordPattern(remaining) {
			return "[HOST_WITH_CREDENTIALS]"
		}
	}

	return host
}

// SanitizeEmail removes sensitive information from email values
func SanitizeEmail(value string) string {
	if strings.Contains(value, "@") {
		// Check if there's a password-like pattern after @
		remaining := value[strings.Index(value, "@")+1:]
		if containsPasswordPattern(remaining) {
			return "[EMAIL_WITH_CREDENTIALS]"
		}
		// Truncate long emails
		if len(value) > 50 {
			return fmt.Sprintf("%s...", value[:47])
		}
	}

	return value
}

// containsPasswordPattern checks for common password patterns
func containsPasswordPattern(s string) bool {
	lower := strings.ToLower(s)
	passwordPatterns := []string{"pass", "pwd", "secret", "key"}

	for _, pattern := range passwordPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

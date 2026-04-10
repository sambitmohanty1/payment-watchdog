package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds production-grade HTTP security headers to
// every response. These headers are recommended by OWASP and required by
// most security compliance frameworks (PCI-DSS, ISO 27001, ASD ISM).
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enforce HTTPS for 1 year, including subdomains
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Prevent MIME-type sniffing attacks
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking by forbidding iframe embedding
		c.Header("X-Frame-Options", "DENY")

		// Enable browser XSS filter (legacy browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information sent with requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features (camera, microphone, etc.)
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Content Security Policy — tightened for an API-only server
		// Adjust if the API serves any HTML pages directly
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Remove the server fingerprint header added by gin/net-http
		c.Header("Server", "")

		c.Next()
	}
}

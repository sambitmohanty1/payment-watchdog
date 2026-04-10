package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware restricts cross-origin requests to the allowed origins
// defined in config. For allowed origins it sets the standard CORS headers,
// handles preflight OPTIONS requests, and forwards all other requests.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	// Build a quick lookup set for O(1) checks
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// If origin is allowed (or empty — same-origin request) set CORS headers
		if origin == "" {
			c.Next()
			return
		}

		if _, allowed := originSet[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Max-Age", "86400") // 24 h preflight cache
			c.Header("Vary", "Origin")
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			if _, allowed := originSet[origin]; allowed {
				c.AbortWithStatus(http.StatusNoContent)
			} else {
				c.AbortWithStatus(http.StatusForbidden)
			}
			return
		}

		c.Next()
	}
}

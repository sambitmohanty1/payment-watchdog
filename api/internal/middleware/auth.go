package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// InitFirebase initialises the Firebase Admin SDK.
// serviceAccountPath is the path to the downloaded service-account JSON file.
// If the path is empty the SDK falls back to Application Default Credentials
// (works automatically inside GCP/OCI environments with Workload Identity).
func InitFirebase(ctx context.Context, projectID, serviceAccountPath string, logger *zap.Logger) (*firebase.App, error) {
	var opts []option.ClientOption

	if serviceAccountPath != "" {
		opts = append(opts, option.WithCredentialsFile(serviceAccountPath))
		logger.Info("Firebase: using service account file", zap.String("path", serviceAccountPath))
	} else {
		logger.Info("Firebase: using Application Default Credentials (ADC)")
	}

	cfg := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// AuthMiddleware is a Gin middleware that validates Firebase ID tokens supplied
// as "Authorization: Bearer <id_token>" headers.
//
// Public routes (health, Xero OAuth, Stripe webhooks) must be registered
// BEFORE this middleware is applied to the route group.
func AuthMiddleware(app *firebase.App, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Auth: missing or malformed Authorization header",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.FullPath()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required (Bearer token)",
			})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		client, err := app.Auth(c.Request.Context())
		if err != nil {
			logger.Error("Auth: could not create Firebase Auth client", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication service unavailable",
			})
			return
		}

		token, err := client.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			logger.Warn("Auth: invalid Firebase token",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Inject claims into the Gin context for downstream handlers
		c.Set("user_id", token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			c.Set("user_email", email)
		}

		// EXTRACT TENANT ID from custom claims (crucial for SaaS scaling)
		if tenantID, ok := token.Claims["tenant_id"].(string); ok {
			c.Set("tenant_id", tenantID)
			logger.Debug("Auth: tenant_id identified", zap.String("tenant_id", tenantID))
		} else {
			// During migration, some users might lack a tenant_id
			logger.Warn("Auth: user has no tenant_id claim", zap.String("user_id", token.UID))
		}

		c.Next()
	}
}

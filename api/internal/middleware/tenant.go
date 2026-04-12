package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantIsolationMiddleware ensures every database operation is scoped to the tenant's schema.
// This implements the "Schema-per-tenant" model for Payment Watchdog SaaS.
func TenantIsolationMiddleware(redisClient *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDRaw, exists := c.Get("tenant_id")
		
		// For Public/Global endpoints (like health), we don't enforce tenant isolation
		if !exists {
			c.Next()
			return
		}

		tenantID := fmt.Sprintf("%v", tenantIDRaw)
		schemaName := fmt.Sprintf("tenant_%s", tenantID)

		// 1. Check if schema is already provisioned (via Redis Cache)
		cacheKey := fmt.Sprintf("provisioned:%s", schemaName)
		if redisClient != nil {
			val, err := redisClient.Get(context.Background(), cacheKey).Result()
			if err == nil && val == "true" {
				// Already provisioned, proceed to isolation
				setupIsolation(c, schemaName, logger)
				return
			}
		}

		// 2. Not in cache, we need to ensure it's provisioned
		db, exists := c.Get("db")
		if !exists {
			logger.Error("SaaS: Database connection missing from context")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server configuration error"})
			return
		}

		gormDB := db.(*gorm.DB)

		// Automated Provisioning logic
		logger.Info("SaaS: Provisioning schema for new tenant", zap.String("tenant_id", tenantID))
		if err := database.ProvisionTenantSchema(gormDB, tenantID); err != nil {
			logger.Error("SaaS: Failed to provision tenant schema", 
				zap.String("tenant_id", tenantID), 
				zap.Error(err))
				
			// Requirement: Block all requests if migration fails
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to set up your account environment. Please contact support.",
			})
			return
		}

		// 3. Cache success in Redis for 24 hours
		if redisClient != nil {
			redisClient.Set(context.Background(), cacheKey, "true", 24*time.Hour)
		}

		// 4. Proceed to isolation
		setupIsolation(c, schemaName, logger)
	}
}

// setupIsolation helper to set the search_path
func setupIsolation(c *gin.Context, schemaName string, logger *zap.Logger) {
	db, _ := c.Get("db")
	gormDB := db.(*gorm.DB)

	// SET search_path TO tenant_<id>, public
	if err := gormDB.Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName)).Error; err != nil {
		logger.Error("SaaS: Failed to set tenant search_path", 
			zap.String("schema", schemaName), 
			zap.Error(err))
			
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Failed to isolate tenant environment",
		})
		return
	}

	c.Next()
}

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantIsolationMiddleware ensures every database operation is scoped to the tenant's schema.
// This implements the "Schema-per-tenant" model for Payment Watchdog SaaS.
func TenantIsolationMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		
		// For Public/Global endpoints (like health), we don't enforce tenant isolation
		if !exists {
			c.Next()
			return
		}

		// Ensure the GORM database instance in this request uses the isolated schema
		// WARNING: This assumes your DB user has permissions to run SET search_path
		db, exists := c.Get("db")
		if !exists {
			logger.Error("SaaS: Database connection missing from context")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server configuration error",
			})
			return
		}

		gormDB := db.(*gorm.DB)
		
		// Standardized schema name: tenant_<id>
		schemaName := fmt.Sprintf("tenant_%s", tenantID)
		
		// SET search_path TO tenant_<id>, public
		// The 'public' schema acts as a fallback for global tables
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
}

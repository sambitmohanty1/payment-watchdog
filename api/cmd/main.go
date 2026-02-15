package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/api"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/database"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/mediators"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/rules"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Payment Watchdog - Payment Failure Intelligence Service")

	// Initialize database
	db, err := initDatabase(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Run database migrations FIRST
	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	logger.Info("Starting rule engine initialization...")
	// Initialize rule engine
	ruleEngineFactory := rules.NewRuleEngineFactory(logger)
	logger.Info("Rule engine factory created successfully")
	ruleEngine := ruleEngineFactory.CreateComprehensiveRuleEngine()
	logger.Info("Comprehensive rule engine created successfully")

	logger.Info("Rule engine initialized",
		zap.Int("total_rules", len(ruleEngine.GetRules())))

	// Initialize services AFTER migrations
	paymentFailureService := services.NewPaymentFailureService(db, logger)

	// Initialize analytics service for Sprint 3
	analyticsService := services.NewAnalyticsService(db, logger)

	// Initialize enhanced webhook service with Sprint 2 features
	webhookSecret := viper.GetString("stripe.webhook_secret")
	webhookService := services.NewWebhookService(db, ruleEngine, webhookSecret)

	alertService := services.NewAlertService(db, logger)

	// Initialize enhanced retry service with exponential backoff and dead letter queue
	retryService := services.NewRetryService(db, 3, 2*time.Second, 30*time.Second, 5)

	dataQualityService := services.NewDataQualityService(db, logger)

	// Initialize communication service
	communicationService := services.NewCommunicationService(db, nil, nil)

	// Initialize recovery orchestration service
	recoveryOrchestrationService := services.NewRecoveryOrchestrationService(db, retryService, communicationService, analyticsService, logger)

	// Initialize monitoring service for Sprint 2 observability
	monitoringService := services.NewMonitoringService(webhookService.GetMetrics())

	// Initialize vault client for secure secret management
	var vaultClient *services.VaultClient
	if vaultURL := viper.GetString("vault.url"); vaultURL != "" {
		if vaultToken := viper.GetString("vault.token"); vaultToken != "" {
			var err error
			vaultClient, err = services.NewVaultClient(vaultURL, vaultToken)
			if err != nil {
				logger.Warn("Failed to initialize Vault client, using config-based secrets", zap.Error(err))
			} else {
				logger.Info("Vault client initialized successfully")
			}
		}
	} else {
		logger.Info("Using config-based secrets (Vault not configured)")
	}

	// Initialize mediators
	var xeroMediator *mediators.XeroMediator
	if xeroClientID := viper.GetString("xero.client_id"); xeroClientID != "" {
		// Create Xero mediator configuration
		xeroConfig := &mediators.ProviderConfig{
			ProviderID:   "xero",
			ProviderType: mediators.ProviderTypeOAuth,
			CompanyID:    "default", // This should be dynamic based on the company
			OAuthConfig: &mediators.OAuthConfig{
				ClientID:     xeroClientID,
				ClientSecret: viper.GetString("xero.client_secret"),
				RedirectURI:  viper.GetString("xero.redirect_uri"),
				Scopes: []string{
					"offline_access",
					"accounting.transactions",
					"accounting.contacts",
					"accounting.settings",
				},
				AuthURL:  "https://login.xero.com/identity/connect/authorize",
				TokenURL: "https://identity.xero.com/connect/token",
			},
		}

		// Create event bus (simplified for now)
		var eventBus mediators.EventBus = nil // TODO: Initialize proper event bus

		xeroMediator = mediators.NewXeroMediator(xeroConfig, eventBus, logger)
		logger.Info("Xero mediator initialized successfully")
	} else {
		logger.Info("Xero configuration not found, skipping Xero mediator initialization")
	}

	// Initialize API handlers with services
	apiHandlers := api.NewHandlers(paymentFailureService, webhookService, alertService, retryService, dataQualityService, analyticsService, recoveryOrchestrationService, communicationService, logger)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, company_id")
		c.Header("Access-Control-Expose-Headers", "Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "payment-watchdog",
			"version":   "v1.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// Enhanced health check with monitoring service
	router.GET("/health/detailed", monitoringService.HandleHealthCheck)

	// Metrics endpoint for Sprint 2 observability
	router.GET("/metrics", monitoringService.HandleMetrics)

	// Start monitoring service health monitoring
	monitoringService.StartHealthMonitoring()

	// Load secrets from Vault if available, otherwise use config
	if vaultClient != nil {
		logger.Info("Loading secrets from Vault...")
		if secrets, err := vaultClient.LoadSecretsFromVault("payment-watchdog"); err == nil {
			// Override config with Vault secrets
			for key, value := range secrets {
				viper.Set(key, value)
			}
			logger.Info("Secrets loaded from Vault successfully")
		} else {
			logger.Warn("Failed to load secrets from Vault, using config", zap.Error(err))
		}
	}

	// API endpoints
	apiV1 := router.Group("/api/v1")
	{
		// Webhook endpoints
		webhookGroup := apiV1.Group("/webhooks")
		{
			webhookGroup.POST("/stripe", webhookService.HandleStripeWebhook)
			// Test endpoint for development
			webhookGroup.POST("/test", webhookService.HandleTestWebhook)
		}

		// Payment failure endpoints
		failuresGroup := apiV1.Group("/failures")
		{
			failuresGroup.GET("", apiHandlers.GetPaymentFailures)
			failuresGroup.GET("/:id", apiHandlers.GetPaymentFailure)
			failuresGroup.POST("/:id/retry", apiHandlers.RetryPayment)
		}

		// Alert endpoints
		alertsGroup := apiV1.Group("/alerts")
		{
			alertsGroup.GET("", apiHandlers.GetAlerts)
			alertsGroup.GET("/:id", apiHandlers.GetAlert)
		}

		// Dashboard endpoints
		dashboardGroup := apiV1.Group("/dashboard")
		{
			dashboardGroup.GET("/stats", apiHandlers.GetDashboardStats)
			dashboardGroup.GET("/export", apiHandlers.ExportData)
			dashboardGroup.GET("/quality", apiHandlers.GetDataQualityReport)
			dashboardGroup.GET("/quality/trends", apiHandlers.GetDataQualityTrends)
		}

		// Retry service endpoints for Sprint 2
		retryGroup := apiV1.Group("/retry")
		{
			retryGroup.GET("/jobs/:id", func(c *gin.Context) {
				jobID := c.Param("id")
				job, err := retryService.GetJobStatus(jobID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, job)
			})
			retryGroup.GET("/jobs", func(c *gin.Context) {
				companyID := c.Query("company_id")
				if companyID == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
					return
				}
				jobs, err := retryService.GetCompanyJobs(companyID, 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, jobs)
			})
			retryGroup.POST("/jobs/:id/retry", func(c *gin.Context) {
				jobID := c.Param("id")
				if err := retryService.RetryFailedJob(jobID); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Job queued for retry"})
			})
			retryGroup.GET("/stats", func(c *gin.Context) {
				stats := retryService.GetStats()
				c.JSON(http.StatusOK, stats)
			})
		}

		// Analytics endpoints for Sprint 3 - NEW VERSION
		analyticsGroup := apiV1.Group("/analytics-v2")
		{
			// Working test endpoint
			analyticsGroup.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message":    "Analytics endpoint is working",
					"version":    "v2",
					"timestamp":  time.Now().UTC(),
					"company_id": c.Query("company_id"),
				})
			})

			// Fixed company summary endpoint
			analyticsGroup.GET("/company/summary", func(c *gin.Context) {
				companyID := c.Query("company_id")
				if companyID == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"company_id": companyID,
					"message":    "Analytics working - now implementing full functionality",
					"timestamp":  time.Now().UTC(),
				})
			})
		}

		// Xero integration endpoints (using mediator pattern)
		if xeroMediator != nil {
			xeroHandlers := api.NewXeroHandlers(xeroMediator, logger)
			api.RegisterXeroRoutes(apiV1, xeroHandlers)
		}
	}

	// Start server
	port := viper.GetString("server.port")
	if port == "" {
		port = "8085"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		httpsEnabled := viper.GetBool("server.https")
		if httpsEnabled {
			certFile := viper.GetString("server.cert_file")
			keyFile := viper.GetString("server.key_file")
			logger.Info("Starting HTTPS server",
				zap.String("port", port),
				zap.String("cert_file", certFile),
				zap.String("key_file", keyFile))

			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start HTTPS server", zap.Error(err))
			}
		} else {
			logger.Info("Starting HTTP server", zap.String("port", port))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start HTTP server", zap.Error(err))
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger() (*zap.Logger, error) {
	level := viper.GetString("log.level")
	var logLevel zap.AtomicLevel

	switch level {
	case "debug":
		logLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		logLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		logLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config := zap.NewProductionConfig()
	config.Level = logLevel
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	return config.Build()
}

func initDatabase(logger *zap.Logger) (*gorm.DB, error) {
	host := viper.GetString("database.host")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbname := viper.GetString("database.name")
	port := viper.GetInt("database.port")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

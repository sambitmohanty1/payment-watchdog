package api

import (
	"context"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/middleware"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/mediators"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/rules"
)

// Server represents the API server
type Server struct {
	engine      *gin.Engine
	logger      *zap.Logger
	db          *gorm.DB
	redis       *redis.Client
	config      *config.Config
	firebaseApp *firebase.App
}

// NewServer creates a new API server instance
func NewServer(logger *zap.Logger, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Initialise Firebase (non-fatal if not configured yet)
	var fbApp *firebase.App
	if cfg.Firebase.ProjectID != "" {
		app, err := middleware.InitFirebase(
			context.Background(),
			cfg.Firebase.ProjectID,
			cfg.Firebase.ServiceAccountPath,
			logger,
		)
		if err != nil {
			logger.Warn("Firebase init failed — auth middleware disabled", zap.Error(err))
		} else {
			fbApp = app
			logger.Info("Firebase Auth initialised", zap.String("project", cfg.Firebase.ProjectID))
		}
	} else {
		logger.Warn("Firebase not configured (FIREBASE_PROJECT_ID not set) — auth middleware disabled")
	}

	// Global middleware — applied to every request
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(middleware.SecurityHeadersMiddleware())
	engine.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	// SaaS: Inject DB into every request context for isolation middleware
	engine.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	return &Server{
		engine:      engine,
		logger:      logger,
		db:          db,
		redis:       redisClient,
		config:      cfg,
		firebaseApp: fbApp,
	}
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() {
	// Create services with proper constructor arguments
	paymentFailureService := services.NewPaymentFailureService(s.db, s.logger)

	// Create dependencies
	ruleEngineFactory := rules.NewRuleEngineFactory(s.logger)
	ruleEngine := ruleEngineFactory.CreateComprehensiveRuleEngine()

	// Instantiate previously nil services with real implementations
	// Note: We safely bypass Redis and External SMS providers to prioritize DB persistence
	webhookService := services.NewWebhookService(s.db, nil, ruleEngine, s.config.Stripe.WebhookSecret)
	alertService := services.NewAlertService(s.db, s.logger)
	retryService := services.NewRetryService(s.db, 3, 5*time.Second, 1*time.Minute, 5)
	dataQualityService := services.NewDataQualityService(s.db, s.logger)
	analyticsService := services.NewAnalyticsService(s.db, s.logger)

	// Create provider registry for dynamic endpoint availability checks
	// This auto-detects Stripe/Xero/PayTo from environment variables
	providerRegistry := services.NewProviderRegistry(s.logger)
	retryService.SetProviderRegistry(providerRegistry)

	communicationService := services.NewCommunicationService(s.db, nil, nil)
	recoveryService := services.NewRecoveryOrchestrationService(s.db, retryService, communicationService, analyticsService, s.logger)

	// 1. Initialize Event Bus for internal communication
	bus := eventbus.NewLocalEventBus(s.logger)

	// 2. Instantiate Event Processor Service to handle recovery & reconciliation matches
	eventProcessor := services.NewEventProcessorService(s.db, ruleEngine, bus, s.logger)
	if err := eventProcessor.StartEventProcessing(context.Background()); err != nil {
		s.logger.Error("Failed to start event processor", zap.Error(err))
	}

	// 3. Create Xero Mediator with appropriate config
	xeroProviderConfig := &mediators.ProviderConfig{
		ProviderID:   "xero-primary",
		ProviderType: mediators.ProviderTypeOAuth,
		CompanyID:    "global", // In production, this would be dynamic
		OAuthConfig: &mediators.OAuthConfig{
			ClientID:     "XERO_CLIENT_ID", // Placeholder, will be injected via ENV or DB
			ClientSecret: "XERO_CLIENT_SECRET",
			TokenURL:     "https://identity.xero.com/connect/token",
			AuthURL:      "https://login.xero.com/identity/connect/authorize",
			Scopes:       []string{"offline_access", "accounting.transactions", "accounting.contacts"},
		},
		SyncConfig: &mediators.SyncConfig{
			Frequency: 30 * time.Minute,
			Enabled:   true,
		},
	}
	xeroMediator := mediators.NewXeroMediator(xeroProviderConfig, bus, s.logger)

	// 4. Create Xero handlers with the real mediator
	xeroHandlers := NewXeroHandlers(xeroMediator, s.logger)

	// Create handlers with available services
	handlers := NewHandlers(
		s.db,
		s.redis,
		paymentFailureService,
		webhookService,
		alertService,
		retryService,
		dataQualityService,
		analyticsService,
		recoveryService,
		communicationService,
		xeroHandlers,
		s.firebaseApp,
		s.logger,
	)

	// API routes
	api := s.engine.Group("/api")
	{
		// === PUBLIC routes — no auth required ===
		api.GET("/health", handlers.HealthCheck)
		api.POST("/webhooks/stripe", handlers.HandleStripeWebhook)

		// === PROTECTED routes — require valid Firebase ID token ===
		protected := api.Group("")
		if s.firebaseApp != nil {
			protected.Use(middleware.AuthMiddleware(s.firebaseApp, s.logger))
			// SaaS: Multi-tenant database isolation with automated provisioning
			protected.Use(middleware.TenantIsolationMiddleware(s.redis, s.logger))
		}
		{
			// Payment failures endpoints
			protected.GET("/payment-failures", handlers.GetPaymentFailures)
			protected.GET("/payment-failures/:id", handlers.GetPaymentFailure)
			protected.POST("/payment-failures/:id/retry", handlers.RetryPayment)

			// Alerts endpoints
			protected.GET("/alerts", handlers.GetAlerts)
			protected.GET("/alerts/:id", handlers.GetAlert)

			// Dashboard endpoints
			protected.GET("/dashboard/stats", handlers.GetDashboardStats)

			// Data quality endpoints
			protected.GET("/data-quality/report", handlers.GetDataQualityReport)
			protected.GET("/data-quality/trends", handlers.GetDataQualityTrends)

			// Analytics endpoints
			protected.GET("/analytics/company/summary", handlers.GetCompanyAnalyticsSummary)
			protected.POST("/analytics/company/payment-failures/analyze", handlers.AnalyzeCompanyPaymentFailures)
			protected.POST("/analytics/customer/payment-failures/analyze", handlers.AnalyzeCustomerPaymentFailures)
			protected.GET("/analytics/customer/risk-score", handlers.GetCustomerRiskScore)

			// Export endpoints
			protected.GET("/export", handlers.ExportData)

			// Onboarding endpoints (for first-time Google/Email users)
			protected.POST("/onboarding/provision", handlers.ProvisionTenant)
		}

		// Recovery endpoints (moved inside /api for proxy compatibility)
		recovery := api.Group("/recovery")
		{
			recovery.GET("/workflows", handlers.recoveryHandlers.GetWorkflows)
			recovery.POST("/workflows", handlers.recoveryHandlers.CreateWorkflow)
			recovery.GET("/workflows/:id", handlers.recoveryHandlers.GetWorkflow)
			recovery.PUT("/workflows/:id", handlers.recoveryHandlers.UpdateWorkflow)
			recovery.DELETE("/workflows/:id", handlers.recoveryHandlers.DeleteWorkflow)
			recovery.POST("/workflows/:id/execute", handlers.recoveryHandlers.ExecuteWorkflow)
			recovery.GET("/workflows/:id/status", handlers.recoveryHandlers.GetWorkflowStatus)
			recovery.GET("/workflows/:id/logs", handlers.recoveryHandlers.GetWorkflowLogs)
		}

		// Xero endpoints (moved inside /api for proxy compatibility)
		xero := api.Group("/xero")
		{
			xero.GET("/oauth/url", handlers.xeroHandlers.GetOAuthURL)
			xero.GET("/oauth/callback", handlers.xeroHandlers.OAuthCallback)
			xero.GET("/contacts", handlers.xeroHandlers.GetContacts)
			xero.POST("/contacts", handlers.xeroHandlers.CreateContact)
			xero.GET("/invoices", handlers.xeroHandlers.GetInvoices)
			xero.POST("/invoices", handlers.xeroHandlers.CreateInvoice)
		}

		// Stripe Webhook (Hybrid Support)
		// Supports both generic /webhooks/stripe and tenant-specific /webhooks/stripe/:tenant_id
		api.POST("/webhooks/stripe", handlers.HandleStripeWebhook)
		api.POST("/webhooks/stripe/:tenant_id", handlers.HandleStripeWebhook)
	}
}

// GetEngine returns the Gin engine
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

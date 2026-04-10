package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/rules"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/eventbus"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/mediators"
	"context"
)

// Server represents the API server
type Server struct {
	engine *gin.Engine
	logger *zap.Logger
	db     *gorm.DB
	config *config.Config
}

// NewServer creates a new API server instance
func NewServer(logger *zap.Logger, db *gorm.DB, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Add middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	return &Server{
		engine: engine,
		logger: logger,
		db:     db,
		config: cfg,
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
		paymentFailureService,
		webhookService,
		alertService,
		retryService,
		dataQualityService,
		analyticsService,
		recoveryService,
		communicationService,
		xeroHandlers,
		s.logger,
	)

	// API routes
	api := s.engine.Group("/api")
	{
		// Health check endpoint
		api.GET("/health", handlers.HealthCheck)

		// Payment failures endpoints
		api.GET("/payment-failures", handlers.GetPaymentFailures)
		api.GET("/payment-failures/:id", handlers.GetPaymentFailure)
		api.POST("/payment-failures/:id/retry", handlers.RetryPayment)

		// Alerts endpoints
		api.GET("/alerts", handlers.GetAlerts)
		api.GET("/alerts/:id", handlers.GetAlert)

		// Dashboard endpoints
		api.GET("/dashboard/stats", handlers.GetDashboardStats)

		// Data quality endpoints
		api.GET("/data-quality/report", handlers.GetDataQualityReport)
		api.GET("/data-quality/trends", handlers.GetDataQualityTrends)

		// Analytics endpoints
		api.GET("/analytics/company/summary", handlers.GetCompanyAnalyticsSummary)
		api.POST("/analytics/company/payment-failures/analyze", handlers.AnalyzeCompanyPaymentFailures)
		api.POST("/analytics/customer/payment-failures/analyze", handlers.AnalyzeCustomerPaymentFailures)
		api.GET("/analytics/customer/risk-score", handlers.GetCustomerRiskScore)

		// Export endpoints
		api.GET("/export", handlers.ExportData)

		// Webhooks
		api.POST("/webhooks/stripe", handlers.HandleStripeWebhook)
	}

	// Recovery endpoints
	recovery := s.engine.Group("/recovery")
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

	// Xero endpoints
	xero := s.engine.Group("/xero")
	{
		xero.GET("/oauth/url", handlers.xeroHandlers.GetOAuthURL)
		xero.GET("/oauth/callback", handlers.xeroHandlers.OAuthCallback)
		xero.GET("/contacts", handlers.xeroHandlers.GetContacts)
		xero.POST("/contacts", handlers.xeroHandlers.CreateContact)
		xero.GET("/invoices", handlers.xeroHandlers.GetInvoices)
		xero.POST("/invoices", handlers.xeroHandlers.CreateInvoice)
	}
}

// GetEngine returns the Gin engine
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

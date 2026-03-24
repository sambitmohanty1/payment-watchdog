package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/config"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// TODO: Implement proper service constructors - using nil for now to get compilation working
	var webhookService *services.WebhookService = nil
	alertService := services.NewAlertService(s.db, s.logger)
	var retryService *services.RetryService = nil
	dataQualityService := services.NewDataQualityService(s.db, s.logger)
	analyticsService := services.NewAnalyticsService(s.db, s.logger)
	var recoveryService *services.RecoveryOrchestrationService = nil
	var communicationService *services.CommunicationService = nil

	// Create handlers with available services
	handlers := NewHandlers(
		paymentFailureService,
		webhookService,
		alertService,
		retryService,
		dataQualityService,
		analyticsService,
		recoveryService,
		communicationService,
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
		xero.POST("/oauth/callback", handlers.xeroHandlers.OAuthCallback)
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

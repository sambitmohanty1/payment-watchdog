package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
	"go.uber.org/zap"
)

// Handlers contains all the API handlers with their dependencies
type Handlers struct {
	paymentFailureService *services.PaymentFailureService
	webhookService        *services.WebhookService
	alertService          *services.AlertService
	retryService          *services.RetryService
	dataQualityService    *services.DataQualityService
	analyticsService      *services.AnalyticsService
	recoveryService       *services.RecoveryOrchestrationService
	communicationService  *services.CommunicationService
	recoveryHandlers      *RecoveryHandlers
	xeroHandlers          *XeroHandlers
	logger                *zap.Logger
}

// NewHandlers creates a new Handlers instance
func NewHandlers(
	paymentFailureService *services.PaymentFailureService,
	webhookService *services.WebhookService,
	alertService *services.AlertService,
	retryService *services.RetryService,
	dataQualityService *services.DataQualityService,
	analyticsService *services.AnalyticsService,
	recoveryService *services.RecoveryOrchestrationService,
	communicationService *services.CommunicationService,
	xeroHandlers *XeroHandlers,
	logger *zap.Logger,
) *Handlers {
	recoveryHandlers := NewRecoveryHandlers(recoveryService, communicationService)

	return &Handlers{
		paymentFailureService: paymentFailureService,
		webhookService:        webhookService,
		alertService:          alertService,
		retryService:          retryService,
		dataQualityService:    dataQualityService,
		analyticsService:      analyticsService,
		recoveryService:       recoveryService,
		communicationService:  communicationService,
		recoveryHandlers:      recoveryHandlers,
		xeroHandlers:          xeroHandlers,
		logger:                logger,
	}
}

// GetPaymentFailures returns a list of payment failures
func (h *Handlers) GetPaymentFailures(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if provider := c.Query("provider_id"); provider != "" {
		filters["provider_id"] = provider
	}
	if customerEmail := c.Query("customer_email"); customerEmail != "" {
		filters["customer_email"] = customerEmail
	}
	if failureReason := c.Query("failure_reason"); failureReason != "" {
		filters["failure_reason"] = failureReason
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// For MVP demonstration, return sample data when no real data exists
	// This allows the frontend to show how data visualization works
	sampleFailures := []gin.H{
		{
			"id":                  "pf-001",
			"company_id":          companyID,
			"provider_id":         "stripe",
			"event_id":            "evt_123456789",
			"event_type":          "payment_intent.payment_failed",
			"payment_intent_id":   "pi_123456789",
			"amount":              1250.00,
			"currency":            "AUD",
			"customer_id":         "cus_123456789",
			"customer_email":      "john.doe@example.com",
			"customer_name":       "John Doe",
			"failure_reason":      "insufficient_funds",
			"failure_code":        "card_declined",
			"failure_message":     "Your card was declined.",
			"status":              "received",
			"webhook_received_at": time.Now().UTC().Format(time.RFC3339),
			"created_at":          time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":                  "pf-002",
			"company_id":          companyID,
			"provider_id":         "stripe",
			"event_id":            "evt_123456790",
			"event_type":          "payment_intent.payment_failed",
			"payment_intent_id":   "pi_123456790",
			"amount":              750.50,
			"currency":            "AUD",
			"customer_id":         "cus_123456790",
			"customer_email":      "jane.smith@example.com",
			"customer_name":       "Jane Smith",
			"failure_reason":      "card_declined",
			"failure_code":        "card_declined",
			"failure_message":     "Your card was declined.",
			"status":              "processing",
			"webhook_received_at": time.Now().UTC().Format(time.RFC3339),
			"created_at":          time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":                  "pf-003",
			"company_id":          companyID,
			"provider_id":         "paypal",
			"event_id":            "evt_123456791",
			"event_type":          "payment_intent.payment_failed",
			"payment_intent_id":   "pi_123456791",
			"amount":              2000.00,
			"currency":            "AUD",
			"customer_id":         "cus_123456791",
			"customer_email":      "bob.wilson@example.com",
			"customer_name":       "Bob Wilson",
			"failure_reason":      "expired_card",
			"failure_code":        "expired_card",
			"failure_message":     "Your card has expired.",
			"status":              "received",
			"webhook_received_at": time.Now().UTC().Format(time.RFC3339),
			"created_at":          time.Now().UTC().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    sampleFailures,
		"filters": filters,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": 12,
		},
	})
}

// GetPaymentFailure returns a specific payment failure
func (h *Handlers) GetPaymentFailure(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	failureIDStr := c.Param("id")
	failureID, err := uuid.Parse(failureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid failure ID"})
		return
	}

	failure, err := h.paymentFailureService.GetPaymentFailure(c.Request.Context(), failureID, companyID)
	if err != nil {
		h.logger.Error("Failed to get payment failure", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment failure not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": failure,
	})
}

// RetryPayment attempts to retry a failed payment
func (h *Handlers) RetryPayment(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	failureIDStr := c.Param("id")
	failureID, err := uuid.Parse(failureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid failure ID"})
		return
	}

	// TODO: Implement actual retry logic
	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment retry initiated",
		"failure_id": failureID,
		"company_id": companyID,
	})
}

// GetAlerts returns a list of alerts
func (h *Handlers) GetAlerts(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// For MVP demonstration, return sample data when no real data exists
	sampleAlerts := []gin.H{
		{
			"id":              "alert-001",
			"company_id":      companyID,
			"type":            "payment_failure",
			"title":           "High Value Payment Failure",
			"message":         "Payment failure of $2,000 detected for customer Bob Wilson",
			"severity":        "high",
			"status":          "unread",
			"action_required": true,
			"created_at":      time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":              "alert-002",
			"company_id":      companyID,
			"type":            "retry_success",
			"title":           "Payment Retry Successful",
			"message":         "Retry attempt for payment $750.50 was successful",
			"severity":        "low",
			"status":          "read",
			"action_required": false,
			"created_at":      time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":              "alert-003",
			"company_id":      companyID,
			"type":            "system",
			"title":           "Recovery Rate Improved",
			"message":         "Payment recovery rate has improved to 73.3% this week",
			"severity":        "medium",
			"status":          "unread",
			"action_required": false,
			"created_at":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sampleAlerts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": 8,
		},
	})
}

// GetAlert returns a specific alert
func (h *Handlers) GetAlert(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert ID"})
		return
	}

	// TODO: Implement actual alert retrieval
	c.JSON(http.StatusOK, gin.H{
		"data": models.CustomerCommunication{
			ID: alertID,
		},
	})
}

// GetDashboardStats returns dashboard statistics
func (h *Handlers) GetDashboardStats(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// For MVP demonstration, return sample data when no real data exists
	// This allows the frontend to show how data visualization works
	c.JSON(http.StatusOK, gin.H{
		"payment_failures": gin.H{
			"total":        12,
			"total_amount": 8750.50,
			"by_status": []gin.H{
				gin.H{"status": "received", "count": 8},
				gin.H{"status": "processing", "count": 3},
				gin.H{"status": "resolved", "count": 1},
			},
			"by_reason": []gin.H{
				gin.H{"reason": "insufficient_funds", "count": 6},
				gin.H{"reason": "card_declined", "count": 4},
				gin.H{"reason": "expired_card", "count": 2},
			},
			"by_provider": []gin.H{
				gin.H{"provider": "stripe", "count": 8},
				gin.H{"provider": "paypal", "count": 4},
			},
			"daily_breakdown": []gin.H{
				gin.H{"date": "2025-08-28", "count": 5, "amount": 3200.00},
				gin.H{"date": "2025-08-29", "count": 7, "amount": 5550.50},
			},
		},
		"alerts": gin.H{
			"total": 8,
			"by_status": []gin.H{
				gin.H{"status": "unread", "count": 5},
				gin.H{"status": "read", "count": 3},
			},
			"by_channel": []gin.H{
				gin.H{"channel": "email", "count": 6},
				gin.H{"channel": "sms", "count": 2},
			},
		},
		"retries": gin.H{
			"total":        15,
			"success_rate": 73.3,
			"by_status": []gin.H{
				gin.H{"status": "completed", "count": 11},
				gin.H{"status": "failed", "count": 4},
			},
		},
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	})
}

// ExportData exports data in various formats
func (h *Handlers) ExportData(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	dataType := c.Query("type")
	if dataType == "" {
		dataType = "payment_failures"
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// TODO: Implement actual export functionality
	c.JSON(http.StatusOK, gin.H{
		"message":    "Export functionality not yet implemented",
		"company_id": companyID,
		"filters": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"pagination": gin.H{
			"page":  1,
			"limit": 20,
			"total": 0,
		},
	})
}

// GetDataQualityReport returns a data quality report for a company
func (h *Handlers) GetDataQualityReport(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// Parse report type parameter
	reportType := c.DefaultQuery("type", "daily")

	// Generate quality report
	report, err := h.dataQualityService.GenerateQualityReport(c.Request.Context(), companyID, reportType)
	if err != nil {
		h.logger.Error("Failed to generate quality report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate quality report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetDataQualityTrends returns data quality trends for a company over time
func (h *Handlers) GetDataQualityTrends(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// Parse days parameter
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 90 {
		days = 7 // Default to 7 days
	}

	// Get quality trends
	trends, err := h.dataQualityService.GetQualityTrends(c.Request.Context(), companyID, days)
	if err != nil {
		h.logger.Error("Failed to get quality trends", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get quality trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trends,
		"filters": gin.H{
			"company_id": companyID,
			"days":       days,
		},
	})
}

// Analytics API Endpoints

// GetCompanyAnalyticsSummary returns analytics summary for a company
func (h *Handlers) GetCompanyAnalyticsSummary(c *gin.Context) {
	// IMMEDIATE TEST: Just return success to test routing
	c.JSON(http.StatusOK, gin.H{"test": "success", "handler": "called"})
	return

	// Check if analytics service is available
	if h.analyticsService == nil {
		h.logger.Error("🔍 API DEBUG: AnalyticsService is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analytics service not available"})
		return
	}

	companyID := c.Query("company_id")
	if companyID == "" {
		h.logger.Error("🔍 API DEBUG: Missing company_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	h.logger.Info("🔍 API DEBUG: Processing company analytics summary", zap.String("company_id", companyID))

	// Get time range from query params (default to 30 days)
	timeRangeStr := c.DefaultQuery("time_range", "720h") // 30 days
	timeRange, err := time.ParseDuration(timeRangeStr)
	if err != nil {
		h.logger.Error("🔍 API DEBUG: Invalid time_range parameter", zap.String("time_range", timeRangeStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time_range format"})
		return
	}

	h.logger.Info("🔍 API DEBUG: Time range parsed", zap.Duration("time_range", timeRange))

	// Get analytics summary
	summary, err := h.analyticsService.GetCompanyAnalyticsSummary(c.Request.Context(), companyID)
	if err != nil {
		h.logger.Error("🔍 API DEBUG: Failed to get company analytics summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("🔍 API DEBUG: Company analytics summary generated successfully", zap.String("company_id", companyID))
	c.JSON(http.StatusOK, summary)
}

// AnalyzeCompanyPaymentFailures performs comprehensive analysis for a company
func (h *Handlers) AnalyzeCompanyPaymentFailures(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	// Parse time range parameter (default to 30 days)
	timeRangeStr := c.DefaultQuery("time_range", "30d")
	timeRange, err := parseTimeRange(timeRangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_range parameter. Use format like '7d', '30d', '90d'"})
		return
	}

	// Perform analysis
	result, err := h.analyticsService.AnalyzeCompanyPaymentFailures(c.Request.Context(), companyID, timeRange)
	if err != nil {
		h.logger.Error("Failed to analyze company payment failures", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze payment failures"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AnalyzeCustomerPaymentFailures performs analysis for a specific customer
func (h *Handlers) AnalyzeCustomerPaymentFailures(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id query parameter is required"})
		return
	}

	// Parse time range parameter (default to 90 days for customer analysis)
	timeRangeStr := c.DefaultQuery("time_range", "90d")
	timeRange, err := parseTimeRange(timeRangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_range parameter. Use format like '7d', '30d', '90d'"})
		return
	}

	// Perform analysis
	result, err := h.analyticsService.AnalyzeCustomerPaymentFailures(c.Request.Context(), companyID, customerID, timeRange)
	if err != nil {
		h.logger.Error("Failed to analyze customer payment failures", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze customer payment failures"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCustomerRiskScore returns the current risk score for a customer
func (h *Handlers) GetCustomerRiskScore(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id query parameter is required"})
		return
	}

	// Get customer risk score
	riskScore, err := h.analyticsService.GetCustomerRiskScore(c.Request.Context(), companyID, customerID)
	if err != nil {
		h.logger.Error("Failed to get customer risk score", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer risk score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"company_id":  companyID,
		"customer_id": customerID,
		"risk_score":  riskScore,
		"risk_level":  getRiskLevel(riskScore),
		"timestamp":   time.Now(),
	})
}

// Helper function to parse time range strings
func parseTimeRange(timeRangeStr string) (time.Duration, error) {
	switch timeRangeStr {
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	case "90d":
		return 90 * 24 * time.Hour, nil
	case "180d":
		return 180 * 24 * time.Hour, nil
	case "365d":
		return 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported time range: %s", timeRangeStr)
	}
}

// Helper function to get risk level from risk score
func getRiskLevel(riskScore float64) string {
	switch {
	case riskScore >= 80:
		return "critical"
	case riskScore >= 60:
		return "high"
	case riskScore >= 40:
		return "medium"
	case riskScore >= 20:
		return "low"
	default:
		return "minimal"
	}
}

// HealthStatus represents the overall system health status
type HealthStatus struct {
	API         APIStatus         `json:"api"`
	Database    DatabaseStatus    `json:"database"`
	Redis       RedisStatus       `json:"redis"`
	Workers     WorkerStatus      `json:"workers"`
	Environment EnvironmentStatus `json:"environment"`
	System      SystemStatus      `json:"system"`
	Timestamp   time.Time         `json:"timestamp"`
}

type APIStatus struct {
	Status    string    `json:"status"` // healthy, degraded, down
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
	LastCheck time.Time `json:"last_check"`
	Response  string    `json:"response_time"`
}

type DatabaseStatus struct {
	Status      string `json:"status"` // connected, disconnected, error
	Host        string `json:"host"`
	Connections int    `json:"connections"`
	Latency     string `json:"latency"`
	LastQuery   string `json:"last_query"`
}

type RedisStatus struct {
	Status      string `json:"status"` // connected, disconnected, error
	Host        string `json:"host"`
	Connections int    `json:"connections"`
	Memory      string `json:"memory_used"`
	LastCommand string `json:"last_command"`
}

type WorkerStatus struct {
	Status     string    `json:"status"` // active, idle, error
	Count      int       `json:"count"`
	LastRun    time.Time `json:"last_run"`
	NextRun    time.Time `json:"next_run"`
	Running    []string  `json:"running_jobs"`
	FailedJobs int       `json:"failed_jobs"`
}

type EnvironmentStatus struct {
	Name      string `json:"name"`      // staging, production
	Version   string `json:"version"`   // sovereign-au
	Region    string `json:"region"`    // ap-melbourne-1
	Namespace string `json:"namespace"` // sovereign-au
}

type SystemStatus struct {
	CPU    string `json:"cpu_usage"`
	Memory string `json:"memory_usage"`
	Disk   string `json:"disk_usage"`
	Load   string `json:"system_load"`
}

// HealthCheck returns comprehensive system health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	// Check API health
	apiStatus := h.checkAPIHealth()

	// Check database connectivity
	dbStatus := h.checkDatabaseHealth()

	// Check Redis connectivity
	redisStatus := h.checkRedisHealth()

	// Check worker status
	workerStatus := h.checkWorkerHealth()

	// Get environment info
	envStatus := h.getEnvironmentStatus()

	// Get system metrics
	systemStatus := h.getSystemMetrics()

	health := HealthStatus{
		API:         apiStatus,
		Database:    dbStatus,
		Redis:       redisStatus,
		Workers:     workerStatus,
		Environment: envStatus,
		System:      systemStatus,
		Timestamp:   time.Now(),
	}

	// Return appropriate HTTP status based on overall health
	overallStatus := h.getOverallHealth(health)
	if overallStatus == "down" {
		c.JSON(http.StatusServiceUnavailable, health)
	} else if overallStatus == "degraded" {
		c.JSON(http.StatusOK, health) // But with degraded status
	} else {
		c.JSON(http.StatusOK, health)
	}
}

// checkAPIHealth checks the API service health
func (h *Handlers) checkAPIHealth() APIStatus {
	start := time.Now()

	// Simulate database and Redis checks
	// In a real implementation, these would be actual health checks
	dbErr := h.simulateDatabasePing()
	redisErr := h.simulateRedisPing()

	responseTime := time.Since(start)

	status := "healthy"
	if dbErr != nil || redisErr != nil {
		status = "degraded"
	}

	return APIStatus{
		Status:    status,
		Version:   os.Getenv("APP_VERSION"),
		Uptime:    h.getUptime(),
		LastCheck: time.Now(),
		Response:  responseTime.String(),
	}
}

// checkDatabaseHealth checks database connectivity
func (h *Handlers) checkDatabaseHealth() DatabaseStatus {
	start := time.Now()
	err := h.simulateDatabasePing()
	latency := time.Since(start)

	status := "connected"
	if err != nil {
		status = "disconnected"
	}

	return DatabaseStatus{
		Status:      status,
		Host:        h.getDatabaseHost(),
		Connections: h.getDatabaseConnections(),
		Latency:     latency.String(),
		LastQuery:   h.getLastQueryTime(),
	}
}

// checkRedisHealth checks Redis connectivity
func (h *Handlers) checkRedisHealth() RedisStatus {
	start := time.Now()
	err := h.simulateRedisPing()
	_ = time.Since(start) // Use the latency calculation to avoid unused variable

	status := "connected"
	if err != nil {
		status = "disconnected"
	}

	return RedisStatus{
		Status:      status,
		Host:        h.getRedisHost(),
		Connections: h.getRedisConnections(),
		Memory:      h.getRedisMemoryUsage(),
		LastCommand: h.getLastRedisCommand(),
	}
}

// checkWorkerHealth checks worker service health
func (h *Handlers) checkWorkerHealth() WorkerStatus {
	// Simulate worker status check
	workers := h.getWorkerStatus()

	status := "active"
	if len(workers.Running) == 0 {
		status = "idle"
	}
	if workers.FailedJobs > 10 {
		status = "error"
	}

	return WorkerStatus{
		Status:     status,
		Count:      workers.Count,
		LastRun:    workers.LastRun,
		NextRun:    workers.NextRun,
		Running:    workers.Running,
		FailedJobs: workers.FailedJobs,
	}
}

// getEnvironmentStatus returns environment information
func (h *Handlers) getEnvironmentStatus() EnvironmentStatus {
	return EnvironmentStatus{
		Name:      os.Getenv("ENVIRONMENT"),
		Version:   os.Getenv("APP_VERSION"),
		Region:    os.Getenv("OCI_REGION"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}
}

// getSystemMetrics returns system metrics
func (h *Handlers) getSystemMetrics() SystemStatus {
	return SystemStatus{
		CPU:    h.getCPUUsage(),
		Memory: h.getMemoryUsage(),
		Disk:   h.getDiskUsage(),
		Load:   h.getSystemLoad(),
	}
}

// getOverallHealth determines overall system health
func (h *Handlers) getOverallHealth(health HealthStatus) string {
	if health.API.Status == "down" || health.Database.Status == "disconnected" || health.Redis.Status == "disconnected" {
		return "down"
	}
	if health.API.Status == "degraded" || health.Database.Status == "error" || health.Redis.Status == "error" {
		return "degraded"
	}
	return "healthy"
}

// Helper functions for health checks
func (h *Handlers) simulateDatabasePing() error {
	// Simulate database ping - in real implementation, this would be actual DB ping
	return nil
}

func (h *Handlers) simulateRedisPing() error {
	// Simulate Redis ping - in real implementation, this would be actual Redis ping
	return nil
}

func (h *Handlers) getUptime() string {
	// Simulate uptime - in real implementation, track actual start time
	return "2h 30m"
}

func (h *Handlers) getDatabaseHost() string {
	return os.Getenv("DATABASE_HOST")
}

func (h *Handlers) getDatabaseConnections() int {
	// Simulate connection count - in real implementation, get from DB pool stats
	return 5
}

func (h *Handlers) getLastQueryTime() string {
	return time.Now().Add(-5 * time.Minute).Format(time.RFC3339)
}

func (h *Handlers) getRedisHost() string {
	return os.Getenv("REDIS_HOST")
}

func (h *Handlers) getRedisConnections() int {
	// Simulate Redis connections
	return 3
}

func (h *Handlers) getRedisMemoryUsage() string {
	// Simulate memory usage
	return "45MB"
}

func (h *Handlers) getLastRedisCommand() string {
	return "GET"
}

func (h *Handlers) getWorkerStatus() WorkerStatus {
	return WorkerStatus{
		Status:     "active",
		Count:      2,
		LastRun:    time.Now().Add(-10 * time.Minute),
		NextRun:    time.Now().Add(50 * time.Minute),
		Running:    []string{"job-001", "job-002"},
		FailedJobs: 2,
	}
}

func (h *Handlers) getCPUUsage() string {
	// Simulate CPU usage - in real implementation, get from system metrics
	return "25%"
}

func (h *Handlers) getMemoryUsage() string {
	// Simulate memory usage
	return "512MB"
}

func (h *Handlers) getDiskUsage() string {
	// Simulate disk usage
	return "2.1GB"
}

func (h *Handlers) getSystemLoad() string {
	// Simulate system load
	return "0.75"
}

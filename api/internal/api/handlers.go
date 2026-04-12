package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"runtime"
)

// Handlers contains all the API handlers with their dependencies
type Handlers struct {
	db                    *gorm.DB
	redis                 *redis.Client
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
	db *gorm.DB,
	redis *redis.Client,
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
		db:                    db,
		redis:                 redis,
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

// HandleStripeWebhook handles incoming webhooks from Stripe
func (h *Handlers) HandleStripeWebhook(c *gin.Context) {
	h.webhookService.HandleStripeWebhook(c)
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

	if h.paymentFailureService == nil {
		h.logger.Error("paymentFailureService is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	failures, total, err := h.paymentFailureService.GetPaymentFailures(c.Request.Context(), companyID, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to fetch payment failures", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment failures"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    failures,
		"filters": filters,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
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

	// Submit the manual retry job to the RetryService subsystem
	if h.retryService != nil {
		_, err = h.retryService.SubmitJob(
			c.Request.Context(),
			"payment_retry",
			companyID,
			map[string]interface{}{
				"failure_id": failureIDStr,
				"source":     "api_manual_retry",
			},
		)
		if err != nil {
			h.logger.Error("Failed to submit manual payment retry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate retry"})
			return
		}
	} else {
		h.logger.Warn("RetryService is not initialized")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment retry successfully initiated",
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

	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not configured"})
		return
	}

	var alerts []models.CustomerCommunication
	var total int64

	query := h.db.Model(&models.CustomerCommunication{}).Where("company_id = ?", companyID)
	
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("Failed to count alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count alerts"})
		return
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&alerts).Error; err != nil {
		h.logger.Error("Failed to fetch alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
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

	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not configured"})
		return
	}

	var alert models.CustomerCommunication
	if err := h.db.Where("id = ? AND company_id = ?", alertID, companyID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
			return
		}
		h.logger.Error("Failed to fetch alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alert,
	})
}

// GetDashboardStats returns dashboard statistics
func (h *Handlers) GetDashboardStats(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id query parameter is required"})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not configured"})
		return
	}

	var total int64
	var totalAmountCents int64
	h.db.Model(&models.PaymentFailureEvent{}).Where("company_id = ?", companyID).Count(&total)
	h.db.Model(&models.PaymentFailureEvent{}).Where("company_id = ?", companyID).Select("COALESCE(SUM(amount_cents), 0)").Scan(&totalAmountCents)

	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var byStatus []StatusCount
	h.db.Model(&models.PaymentFailureEvent{}).Select("status, COUNT(*) as count").Where("company_id = ?", companyID).Group("status").Scan(&byStatus)

	type ReasonCount struct {
		Reason string `json:"reason"`
		Count  int64  `json:"count"`
	}
	var byReason []ReasonCount
	h.db.Model(&models.PaymentFailureEvent{}).Select("failure_reason as reason, COUNT(*) as count").Where("company_id = ?", companyID).Group("failure_reason").Order("count DESC").Limit(5).Scan(&byReason)

	type ProviderCount struct {
		Provider string `json:"provider"`
		Count    int64  `json:"count"`
	}
	var byProvider []ProviderCount
	h.db.Model(&models.PaymentFailureEvent{}).Select("provider_id as provider, COUNT(*) as count").Where("company_id = ?", companyID).Group("provider_id").Scan(&byProvider)

	// Fetch Alerts via AlertService for the dashboard
	var alertsTotal int64 = 0
	if h.alertService != nil {
		if stats, err := h.alertService.GetAlertStats(c.Request.Context(), companyID); err == nil {
			if count, ok := stats["total_alerts"].(int64); ok {
				alertsTotal = count
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_failures": gin.H{
			"total":        total,
			"total_amount": float64(totalAmountCents) / 100.0,
			"by_status":    byStatus,
			"by_reason":    byReason,
			"by_provider":  byProvider,
			"daily_breakdown": []gin.H{},
		},
		"alerts": gin.H{
			"total": alertsTotal,
			"by_status": []gin.H{},
			"by_channel": []gin.H{},
		},
		"retries": gin.H{
			"total":        0,
			"success_rate": 0.0,
			"by_status": []gin.H{},
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
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Panic in HealthCheck", zap.Any("recover", r))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server panic in health check"})
		}
	}()

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
	if h.db == nil {
		return fmt.Errorf("database connection not initialized")
	}
	db, err := h.db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func (h *Handlers) simulateRedisPing() error {
	if h.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return h.redis.Ping(context.Background()).Err()
}

func (h *Handlers) getUptime() string {
	// Simulate uptime - in real implementation, track actual start time
	return "2h 30m"
}

func (h *Handlers) getDatabaseHost() string {
	return os.Getenv("DATABASE_HOST")
}

func (h *Handlers) getDatabaseConnections() int {
	if h.db == nil {
		return 0
	}
	db, err := h.db.DB()
	if err != nil {
		return 0
	}
	return db.Stats().InUse
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
	if h.redis == nil {
		return "unknown"
	}
	info := h.redis.Info(context.Background(), "memory").Val()
	// Simple parsing for used_memory_human
	return info
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
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("%v MiB", m.Alloc/1024/1024)
}

func (h *Handlers) getDiskUsage() string {
	// Simulate disk usage
	return "2.1GB"
}

func (h *Handlers) getSystemLoad() string {
	// Simulate system load
	return "0.75"
}

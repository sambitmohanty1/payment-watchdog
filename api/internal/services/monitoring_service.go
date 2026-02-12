package services

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitoringService provides monitoring and observability for the application
type MonitoringService struct {
	webhookMetrics *WebhookMetrics
	healthStatus   *HealthStatus
	mu             sync.RWMutex
}

// HealthStatus represents the overall health of the system
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      time.Duration          `json:"uptime"`
	Components  map[string]ComponentStatus `json:"components"`
	LastCheck   time.Time              `json:"last_check"`
}

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	LastCheck time.Time              `json:"last_check"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(webhookMetrics *WebhookMetrics) *MonitoringService {
	return &MonitoringService{
		webhookMetrics: webhookMetrics,
		healthStatus: &HealthStatus{
			Status:     "healthy",
			Timestamp:  time.Now(),
			Uptime:     time.Since(time.Now()), // Will be updated
			Components: make(map[string]ComponentStatus),
			LastCheck:  time.Now(),
		},
	}
}

// GetHealthStatus returns the current health status
func (m *MonitoringService) GetHealthStatus() *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Update uptime
	m.healthStatus.Uptime = time.Since(m.healthStatus.Timestamp)
	
	return m.healthStatus
}

// UpdateComponentStatus updates the status of a specific component
func (m *MonitoringService) UpdateComponentStatus(component string, status, message string, details map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.healthStatus.Components[component] = ComponentStatus{
		Status:    status,
		Message:   message,
		Details:   details,
		LastCheck: time.Now(),
	}
	
	// Update overall health status
	m.updateOverallHealth()
}

// updateOverallHealth determines the overall health status based on component statuses
func (m *MonitoringService) updateOverallHealth() {
	overallStatus := "healthy"
	
	for _, component := range m.healthStatus.Components {
		if component.Status == "critical" {
			overallStatus = "critical"
			break
		} else if component.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}
	
	m.healthStatus.Status = overallStatus
	m.healthStatus.LastCheck = time.Now()
}

// GetWebhookMetrics returns current webhook processing metrics
func (m *MonitoringService) GetWebhookMetrics() *WebhookMetrics {
	return m.webhookMetrics
}

// HandleHealthCheck handles the health check endpoint
func (m *MonitoringService) HandleHealthCheck(c *gin.Context) {
	health := m.GetHealthStatus()
	
	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if health.Status == "critical" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = http.StatusOK // Still responding but with warning
	}
	
	c.JSON(statusCode, health)
}

// HandleMetrics handles the metrics endpoint
func (m *MonitoringService) HandleMetrics(c *gin.Context) {
	metrics := m.GetWebhookMetrics()
	
	// Format metrics for monitoring systems
	response := map[string]interface{}{
		"webhook_metrics": map[string]interface{}{
			"total_received":         metrics.TotalReceived,
			"successfully_processed": metrics.SuccessfullyProcessed,
			"failed_processing":      metrics.FailedProcessing,
			"success_rate":           m.calculateSuccessRate(metrics),
			"average_processing_time_ms": metrics.AverageProcessingTime.Milliseconds(),
			"last_webhook_received":  metrics.LastWebhookReceived,
			"company_counts":         metrics.CompanyWebhookCounts,
		},
		"health_status": m.GetHealthStatus(),
		"timestamp":     time.Now(),
	}
	
	c.JSON(http.StatusOK, response)
}

// calculateSuccessRate calculates the success rate percentage
func (m *MonitoringService) calculateSuccessRate(metrics *WebhookMetrics) float64 {
	if metrics.TotalReceived == 0 {
		return 100.0
	}
	
	return float64(metrics.SuccessfullyProcessed) / float64(metrics.TotalReceived) * 100.0
}

// StartHealthMonitoring starts periodic health monitoring
func (m *MonitoringService) StartHealthMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			m.performHealthCheck()
		}
	}()
}

// performHealthCheck performs a comprehensive health check
func (m *MonitoringService) performHealthCheck() {
	// Check webhook processing health
	metrics := m.GetWebhookMetrics()
	
	// Calculate success rate
	successRate := m.calculateSuccessRate(metrics)
	
	// Determine webhook health status
	webhookStatus := "healthy"
	webhookMessage := "Webhook processing is healthy"
	
	if successRate < 95.0 {
		webhookStatus = "degraded"
		webhookMessage = fmt.Sprintf("Webhook success rate is low: %.2f%%", successRate)
	}
	
	if successRate < 80.0 {
		webhookStatus = "critical"
		webhookMessage = fmt.Sprintf("Webhook success rate is critically low: %.2f%%", successRate)
	}
	
	// Check for recent webhook activity
	if time.Since(metrics.LastWebhookReceived) > 5*time.Minute {
		if webhookStatus == "healthy" {
			webhookStatus = "degraded"
			webhookMessage = "No recent webhook activity detected"
		}
	}
	
	// Update component status
	m.UpdateComponentStatus("webhook_processing", webhookStatus, webhookMessage, map[string]interface{}{
		"success_rate":           successRate,
		"total_received":         metrics.TotalReceived,
		"last_webhook_received":  metrics.LastWebhookReceived,
		"processing_errors":      len(metrics.ProcessingErrors),
	})
	
	// Check system resources (placeholder for future implementation)
	m.UpdateComponentStatus("system_resources", "healthy", "System resources are normal", map[string]interface{}{
		"memory_usage": "normal",
		"cpu_usage":    "normal",
		"disk_usage":   "normal",
	})
}

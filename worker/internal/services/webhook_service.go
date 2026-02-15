package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/sambitmohanty1/payment-watchdog/internal/rules"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// WebhookProcessor handles webhook processing with retry logic and rate limiting
type WebhookProcessor struct {
	MaxRetries      int
	RetryDelay      time.Duration
	DeadLetterQueue chan WebhookEvent
	RateLimiter     *rate.Limiter
}

// WebhookEvent represents a webhook event for processing
type WebhookEvent struct {
	CompanyID string
	Event     *stripe.Event
	RawBody   []byte
	Headers   http.Header
	Timestamp time.Time
}

// WebhookError represents webhook processing errors
type WebhookError struct {
	Type       string    `json:"type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Retryable  bool      `json:"retryable"`
	CompanyID  string    `json:"company_id"`
	EventID    string    `json:"event_id"`
	Timestamp  time.Time `json:"timestamp"`
	RetryCount int       `json:"retry_count"`
}

// WebhookMetrics tracks webhook processing metrics
type WebhookMetrics struct {
	TotalReceived         int64
	SuccessfullyProcessed int64
	FailedProcessing      int64
	AverageProcessingTime time.Duration
	LastWebhookReceived   time.Time
	CompanyWebhookCounts  map[string]int64
	ProcessingErrors      []WebhookError
}

// WebhookService handles incoming webhook events
type WebhookService struct {
	db            *gorm.DB
	ruleEngine    rules.RuleEngine
	processor     *WebhookProcessor
	metrics       *WebhookMetrics
	webhookSecret string
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB, ruleEngine rules.RuleEngine, webhookSecret string) *WebhookService {
	processor := &WebhookProcessor{
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		DeadLetterQueue: make(chan WebhookEvent, 100),
		RateLimiter:     rate.NewLimiter(rate.Limit(100), 200), // 100 req/sec, burst of 200
	}

	metrics := &WebhookMetrics{
		CompanyWebhookCounts: make(map[string]int64),
		ProcessingErrors:     make([]WebhookError, 0),
	}

	return &WebhookService{
		db:            db,
		ruleEngine:    ruleEngine,
		processor:     processor,
		metrics:       metrics,
		webhookSecret: webhookSecret,
	}
}

// HandleStripeWebhook processes incoming Stripe webhooks
func (s *WebhookService) HandleStripeWebhook(c *gin.Context) {
	fmt.Printf("=== Webhook Processing Started ===\n")

	// Try to get company_id from query parameter first (for testing)
	companyID := c.Query("company_id")
	fmt.Printf("Company ID from query: %s\n", companyID)

	// Read request body first to extract company_id from webhook data if not provided
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("ERROR: Failed to read request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	fmt.Printf("Request body length: %d bytes\n", len(body))
	fmt.Printf("Request headers: %+v\n", c.Request.Header)

	// Verify webhook signature
	fmt.Printf("Verifying webhook signature...\n")
	event, err := s.verifyWebhookSignature(c.Request.Header, body)
	if err != nil {
		fmt.Printf("ERROR: Webhook signature verification failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"})
		return
	}
	fmt.Printf("Webhook signature verified successfully. Event ID: %s, Type: %s\n", event.ID, event.Type)

	// If company_id not provided in query, try to extract from webhook data
	if companyID == "" {
		fmt.Printf("No company_id in query, extracting from webhook data...\n")
		// For now, use a default company_id for testing
		// In production, you'd want to extract this from the webhook data or use a mapping
		companyID = "default_company"
	}
	fmt.Printf("Final company ID: %s\n", companyID)

	// Rate limiting check
	fmt.Printf("Checking rate limiting...\n")
	if !s.processor.RateLimiter.Allow() {
		fmt.Printf("ERROR: Rate limit exceeded\n")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}
	fmt.Printf("Rate limit check passed\n")

	// Create webhook event
	webhookEvent := WebhookEvent{
		CompanyID: companyID,
		Event:     event,
		RawBody:   body,
		Headers:   c.Request.Header,
		Timestamp: time.Now(),
	}
	fmt.Printf("Created webhook event for company: %s\n", companyID)

	// Process webhook with retry logic
	fmt.Printf("Processing webhook with retry logic...\n")
	err = s.processor.ProcessWithRetry(webhookEvent)
	if err != nil {
		fmt.Printf("ERROR: Webhook processing failed after retries: %v\n", err)
		s.logWebhookError(WebhookError{
			Type:      "processing",
			Severity:  "critical",
			Message:   fmt.Sprintf("webhook processing failed after retries: %v", err),
			Retryable: false,
			CompanyID: companyID,
			EventID:   event.ID,
			Timestamp: time.Now(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook processing failed"})
		return
	}
	fmt.Printf("Webhook processing completed successfully\n")

	// Update metrics
	fmt.Printf("Updating metrics...\n")
	s.updateMetrics(companyID, true, time.Since(webhookEvent.Timestamp))
	fmt.Printf("Metrics updated successfully\n")

	fmt.Printf("=== Webhook Processing Completed Successfully ===\n")
	c.JSON(http.StatusOK, gin.H{"status": "webhook processed successfully", "company_id": companyID})
}

// ProcessWithRetry processes webhook events with exponential backoff retry
func (wp *WebhookProcessor) ProcessWithRetry(event WebhookEvent) error {
	for attempt := 0; attempt < wp.MaxRetries; attempt++ {
		if err := wp.processEvent(event); err == nil {
			return nil
		}

		if attempt < wp.MaxRetries-1 {
			delay := wp.RetryDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}

	// Send to dead letter queue after max retries
	select {
	case wp.DeadLetterQueue <- event:
		// Successfully queued
	default:
		// Queue is full, log error
		fmt.Printf("Dead letter queue full, dropping event: %s\n", event.Event.ID)
	}

	return errors.New("max retries exceeded")
}

// processEvent processes a single webhook event
func (wp *WebhookProcessor) processEvent(event WebhookEvent) error {
	// Process the webhook event based on its type
	switch event.Event.Type {
	case "payment_intent.payment_failed":
		// Extract payment intent data
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Event.Data.Raw, &paymentIntent); err != nil {
			return fmt.Errorf("failed to unmarshal payment intent: %v", err)
		}

		// Log the payment failure event
		fmt.Printf("Processing payment failure: PaymentIntent %s failed for company %s\n",
			paymentIntent.ID, event.CompanyID)

		// TODO: Add actual business logic here:
		// 1. Store payment failure event in database
		// 2. Trigger alerts/notifications
		// 3. Apply business rules
		// 4. Update company metrics

		return nil

	default:
		// Log unknown event types
		fmt.Printf("Received unknown webhook event type: %s for company %s\n",
			event.Event.Type, event.CompanyID)
		return nil
	}
}

// verifyWebhookSignature verifies the Stripe webhook signature
func (s *WebhookService) verifyWebhookSignature(headers http.Header, body []byte) (*stripe.Event, error) {
	fmt.Printf("=== Signature Verification Started ===\n")

	// Get the signature from headers
	signature := headers.Get("Stripe-Signature")
	if signature == "" {
		fmt.Printf("ERROR: Missing Stripe signature header\n")
		return nil, errors.New("missing Stripe signature")
	}
	fmt.Printf("Stripe signature header: %s\n", signature)
	fmt.Printf("Webhook secret configured: %s\n", s.webhookSecret)
	fmt.Printf("Request body (first 100 chars): %s\n", string(body[:min(len(body), 100)]))

	// Verify the signature
	fmt.Printf("Calling webhook.ConstructEvent...\n")
	event, err := webhook.ConstructEvent(body, signature, s.webhookSecret)
	if err != nil {
		fmt.Printf("ERROR: webhook.ConstructEvent failed: %v\n", err)
		return nil, fmt.Errorf("signature verification failed: %v", err)
	}
	fmt.Printf("Signature verification successful. Event created: %d\n", event.Created)

	// Check if the event is too old (replay protection)
	eventTime := time.Unix(event.Created, 0)
	timeSinceEvent := time.Since(eventTime)
	fmt.Printf("Event timestamp: %s, Time since event: %s\n", eventTime.Format(time.RFC3339), timeSinceEvent)

	if timeSinceEvent > 5*time.Minute {
		fmt.Printf("ERROR: Event too old (replay protection). Event age: %s\n", timeSinceEvent)
		return nil, errors.New("webhook event too old (replay protection)")
	}
	fmt.Printf("Event age check passed\n")

	fmt.Printf("=== Signature Verification Completed Successfully ===\n")
	return &event, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// logWebhookError logs webhook processing errors
func (s *WebhookService) logWebhookError(err WebhookError) {
	s.metrics.ProcessingErrors = append(s.metrics.ProcessingErrors, err)

	// Log error with structured logging
	fmt.Printf("Webhook Error: Type=%s, Severity=%s, Company=%s, Message=%s\n",
		err.Type, err.Severity, err.CompanyID, err.Message)
}

// updateMetrics updates webhook processing metrics
func (s *WebhookService) updateMetrics(companyID string, success bool, processingTime time.Duration) {
	s.metrics.TotalReceived++
	s.metrics.LastWebhookReceived = time.Now()

	if success {
		s.metrics.SuccessfullyProcessed++
	} else {
		s.metrics.FailedProcessing++
	}

	// Update company-specific counts
	s.metrics.CompanyWebhookCounts[companyID]++

	// Update average processing time
	if s.metrics.AverageProcessingTime == 0 {
		s.metrics.AverageProcessingTime = processingTime
	} else {
		s.metrics.AverageProcessingTime = (s.metrics.AverageProcessingTime + processingTime) / 2
	}
}

// GetMetrics returns current webhook metrics
func (s *WebhookService) GetMetrics() *WebhookMetrics {
	return s.metrics
}

// HandleTestWebhook creates a test webhook event for development
func (s *WebhookService) HandleTestWebhook(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	startTime := time.Now()
	fmt.Printf("Test webhook started for company: %s\n", companyID)

	// Create a test payment failure event with proper Stripe types
	testEvent := &stripe.Event{
		ID:      "evt_test_" + uuid.New().String()[:8],
		Type:    "payment_intent.payment_failed",
		Created: time.Now().Unix(),
		Data: &stripe.EventData{
			Raw: []byte(`{
				"id": "pi_test_` + uuid.New().String()[:8] + `",
				"amount": 5000,
				"currency": "usd",
				"customer": "cus_test_` + uuid.New().String()[:8] + `",
				"last_payment_error": {
					"code": "card_declined",
					"message": "Your card was declined"
				}
			}`),
		},
	}

	fmt.Printf("Created test event: %s\n", testEvent.ID)

	// Process the test event
	webhookEvent := WebhookEvent{
		CompanyID: companyID,
		Event:     testEvent,
		Timestamp: time.Now(),
	}

	// Process with retry logic
	fmt.Printf("Processing webhook event with retry logic...\n")
	err := s.processor.ProcessWithRetry(webhookEvent)
	if err != nil {
		fmt.Printf("Webhook processing failed: %v\n", err)
		// Update metrics for failed processing
		s.updateMetrics(companyID, false, time.Since(startTime))
		fmt.Printf("Updated metrics for failed processing\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test webhook processing failed"})
		return
	}

	fmt.Printf("Webhook processing successful, updating metrics...\n")
	// Update metrics for successful processing
	s.updateMetrics(companyID, true, time.Since(startTime))
	fmt.Printf("Metrics updated successfully. Total received: %d\n", s.metrics.TotalReceived)

	c.JSON(http.StatusOK, gin.H{
		"status":     "test webhook processed successfully",
		"event_id":   testEvent.ID,
		"message":    "Test payment failure event created and processed",
		"company_id": companyID,
	})
}

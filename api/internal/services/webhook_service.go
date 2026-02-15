package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"gorm.io/gorm"

	"payment-watchdog/api/internal/models"
	"payment-watchdog/api/internal/rules"
)

type WebhookService struct {
	db            *gorm.DB
	redisClient   *redis.Client
	ruleEngine    rules.RuleEngine
	webhookSecret string
}

func NewWebhookService(db *gorm.DB, rc *redis.Client, ruleEngine rules.RuleEngine, webhookSecret string) *WebhookService {
	// Ensure DLQ table exists
	_ = db.AutoMigrate(&models.DeadLetterEntry{})
	return &WebhookService{
		db:            db,
		redisClient:   rc,
		ruleEngine:    ruleEngine,
		webhookSecret: webhookSecret,
	}
}

func (s *WebhookService) HandleStripeWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read_failed"})
		return
	}

	// 1. Security: Verify Signature
	event, err := webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), s.webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_signature"})
		return
	}

	// 2. Reliability: Distributed Rate Limiting (Redis)
	// Limit to 100 req/sec globally across all pods
	limitKey := "global_webhook_rate_limit"
	count, _ := s.redisClient.Incr(ctx, limitKey).Result()
	if count == 1 {
		s.redisClient.Expire(ctx, limitKey, 1*time.Second)
	}
	if count > 100 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited"})
		return
	}

	// 3. Reliability: Idempotency Check (Redis)
	// If we've seen this Event ID before, verify status
	dedupKey := fmt.Sprintf("processed_event:%s", event.ID)
	wasSet, _ := s.redisClient.SetNX(ctx, dedupKey, "processing", 24*time.Hour).Result()
	if !wasSet {
		c.JSON(http.StatusOK, gin.H{"status": "duplicate_ignored"})
		return
	}

	// 4. Processing logic with DLQ fallback
	if err := s.processEvent(ctx, &event, body); err != nil {
		s.logToDLQ(event.ID, body, err)
		// Return 200 to Stripe so they don't retry indefinitely; we handle it via DLQ
		c.JSON(http.StatusOK, gin.H{"status": "queued_for_review", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "id": event.ID})
}

func (s *WebhookService) processEvent(ctx context.Context, event *stripe.Event, rawBody []byte) error {
	if event.Type == "payment_intent.payment_failed" {
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return err
		}

		// Use db transaction for atomicity
		return s.db.Transaction(func(tx *gorm.DB) error {
			failure := models.PaymentFailureEvent{
				EventID:        event.ID,
				ProviderID:     "stripe",
				EventType:      event.Type,
				PaymentIntentID: pi.ID,
				AmountCents:    pi.Amount, // Correct Int64 handling
				Currency:       string(pi.Currency),
				Status:         "received",
				RawEventData:   string(rawBody),
				WebhookReceivedAt: time.Now(),
			}
			return tx.Create(&failure).Error
		})
	}
	return nil
}

func (s *WebhookService) logToDLQ(eventID string, payload []byte, err error) {
	entry := models.DeadLetterEntry{
		EventID:   eventID,
		Payload:   payload,
		Error:     err.Error(),
		CreatedAt: time.Now(),
	}
	// Use background context to ensure write happens even if request cancels
	s.db.WithContext(context.Background()).Create(&entry)
}

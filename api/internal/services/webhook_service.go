package services

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/stripe/stripe-go/v74"
    "github.com/stripe/stripe-go/v74/webhook"
    "gorm.io/gorm"

    "github.com/sambitmohanty1/payment-watchdog/api/internal/models"
    "github.com/sambitmohanty1/payment-watchdog/api/internal/rules"
)

type WebhookService struct {
    db            *gorm.DB
    redisClient   *redis.Client
    ruleEngine    rules.RuleEngine
    webhookSecret string
}

func NewWebhookService(db *gorm.DB, rc *redis.Client, ruleEngine rules.RuleEngine, webhookSecret string) *WebhookService {
    // Ensure DLQ table exists
    _ = db.AutoMigrate(&models.DeadLetterEntry{})
    return &WebhookService{
        db:            db,
        redisClient:   rc,
        ruleEngine:    ruleEngine,
        webhookSecret: webhookSecret,
    }
}

func (s *WebhookService) HandleStripeWebhook(c *gin.Context) {
    // Limit body size to prevent memory exhaustion attacks
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1048576) // 1MB limit

    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "request body too large or unreadable"})
        return
    }

    // 1. SECURITY: Real Signature Verification
    // In dev/test, if secret is empty, you might skip this, but for prod it's mandatory.
    if s.webhookSecret != "" {
        event, err := webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), s.webhookSecret)
        if err != nil {
            fmt.Printf("Webhook signature verification failed: %v\n", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_signature"})
            return
        }
        
        // 3. Reliability: Idempotency Check (Redis)
        // Check if we have processed this specific Stripe Event ID before
        ctx := c.Request.Context()
        dedupKey := fmt.Sprintf("processed_event:%s", event.ID)
        wasSet, _ := s.redisClient.SetNX(ctx, dedupKey, "processing", 24*time.Hour).Result()
        if !wasSet {
            c.JSON(http.StatusOK, gin.H{"status": "duplicate_ignored"})
            return
        }

        // 4. Reliability: Distributed Rate Limiting (Redis)
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

        // 5. Process with DLQ Fallback
        if err := s.processEvent(ctx, &event, body); err != nil {
            s.logToDLQ(event.ID, body, err)
            c.JSON(http.StatusOK, gin.H{"status": "queued_for_review", "error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"status": "success", "id": event.ID})
    } else {
        // Fallback for local testing without signature
        c.JSON(http.StatusOK, gin.H{"status": "dev_mode_processed"})
    }
}

func (s *WebhookService) processEvent(ctx context.Context, event *stripe.Event, rawBody []byte) error {
    if event.Type == "payment_intent.payment_failed" {
        var pi stripe.PaymentIntent
        if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
            return err
        }

        return s.db.Transaction(func(tx *gorm.DB) error {
            failure := models.PaymentFailureEvent{
                EventID:           event.ID,
                ProviderID:        "stripe",
                EventType:         event.Type,
                PaymentIntentID:   pi.ID,
                AmountCents:       pi.Amount, // FINANCIAL INTEGRITY: Uses int64
                Currency:          string(pi.Currency),
                Status:            "received",
                RawEventData:      string(rawBody),
                WebhookReceivedAt: time.Now(),
            }
            return tx.Create(&failure).Error
        })
    }
    return nil
}

func (s *WebhookService) logToDLQ(eventID string, payload []byte, err error) {
    entry := models.DeadLetterEntry{
        EventID:   eventID,
        Payload:   payload,
        Error:     err.Error(),
        CreatedAt: time.Now(),
    }
    s.db.WithContext(context.Background()).Create(&entry)
}

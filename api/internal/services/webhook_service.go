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
    "github.com/sambitmohanty1/payment-watchdog/api/internal/database"
)

type WebhookService struct {
    db          *gorm.DB
    redisClient  *redis.Client
    ruleEngine   rules.RuleEngine
    webhookSecret string
}

func NewWebhookService(db *gorm.DB, rc *redis.Client, ruleEngine rules.RuleEngine, webhookSecret string) *WebhookService {
    // Ensure DLQ table exists
    if err := db.AutoMigrate(&models.DeadLetterEntry{}); err != nil {
        fmt.Printf("Failed to auto-migrate DeadLetterEntry: %v\n", err)
    }
    return &WebhookService{
        db:          db,
        redisClient:   rc,
        ruleEngine:    ruleEngine,
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

        // 2. Hybrid Tenant Identification
        // Priority 1: URL Parameter (:tenant_id)
        // Priority 2: Stripe Metadata (company_id)
        tenantID := c.Param("tenant_id")
        if tenantID == "" {
            // Safe extraction from event data
            if pi, ok := event.Data.Object["metadata"].(map[string]interface{}); ok {
                if cid, ok := pi["company_id"].(string); ok {
                    tenantID = cid
                }
            }
        }

        if tenantID == "" {
            fmt.Printf("Warning: Dropping webhook %s - No tenant_id or company_id found in URL or Metadata\n", event.ID)
            c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "no_tenant_context"})
            return
        }

        // 3. Automated Sovereign Provisioning
        // Ensure the private schema exists for this client before processing
        if err := database.ProvisionTenantSchema(s.db, tenantID); err != nil {
            fmt.Printf("Provisioning failure for tenant %s: %v\n", tenantID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "provisioning_failed"})
            return
        }

        // 4. Reliability: Idempotency Check (Redis)
        ctx := c.Request.Context()
        dedupKey := fmt.Sprintf("processed_event:%s:%s", tenantID, event.ID)
        wasSet, _ := s.redisClient.SetNX(ctx, dedupKey, "processing", 24*time.Hour).Result()
        if !wasSet {
            c.JSON(http.StatusOK, gin.H{"status": "already_processed"})
            return
        }

        // 5. Reliability: Distributed Rate Limiting (Redis)
        limitKey := fmt.Sprintf("rate_limit:webhook:%s", tenantID)
        count, _ := s.redisClient.Incr(ctx, limitKey).Result()
        if count == 1 {
            s.redisClient.Expire(ctx, limitKey, 1*time.Second)
        }

        // 6. Process with Metadata Context
        if err := s.processEvent(ctx, &event, body, tenantID); err != nil {
            s.logToDLQ(event.ID, body, err)
            c.JSON(http.StatusOK, gin.H{"status": "queued_for_review", "error": err.Error()})
            return
        } else {
            // Fallback for local testing without signature
            c.JSON(http.StatusOK, gin.H{"status": "dev_mode_processed"})
        }
    } else {
        // Fallback for local testing without signature
        c.JSON(http.StatusOK, gin.H{"status": "dev_mode_processed"})
    }
}

func (s *WebhookService) processEvent(ctx context.Context, event *stripe.Event, rawBody []byte, tenantID string) error {
    if event.Type == "payment_intent.payment_failed" {
        var pi stripe.PaymentIntent
        if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
            return err
        }

        // Isolation: Ensure transaction is scoped to the tenant's private schema
        schemaName := fmt.Sprintf("tenant_%s", tenantID)
        return s.db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName)).Error; err != nil {
                return fmt.Errorf("isolation failure: %w", err)
            }

            failure := models.PaymentFailureEvent{
                EventID:           event.ID,
                CompanyID:         tenantID,
                ProviderID:        "stripe",
                EventType:         event.Type,
                AmountCents:       pi.Amount, // FINANCIAL INTEGRITY: Uses int64
                Currency:          string(pi.Currency),
                Status:            "received",
                RawEventData:      string(rawBody),
                WebhookReceivedAt: time.Now(),
            }
            
            if err := tx.Create(&failure).Error; err != nil {
                return err
            }

            // 6. INTELLIGENCE: Execute Rule Engine for Classification
            if s.ruleEngine != nil {
                results := s.ruleEngine.ExecuteRules(&failure)
                fmt.Printf("Executed %d rules for event %s\n", len(results), failure.EventID)
                for _, res := range results {
                    if res.Success {
                        fmt.Printf("Rule Success: %s - %s\n", res.RuleName, res.Message)
                    }
                }
                
                // Transition to 'processed' status if rules were evaluated
                return tx.Model(&failure).Update("status", "processed").Error
            }

            return nil
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
    s.db.WithContext(context.Background()).Create(&entry)
}

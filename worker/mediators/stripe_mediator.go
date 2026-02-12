package mediators

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.uber.org/zap"
)

// StripeMediator implements PaymentProvider for Stripe
type StripeMediator struct {
	*BaseMediator
	webhookSecret string
	eventBus      EventBus
}

// Stripe-specific configuration
type StripeConfig struct {
	WebhookSecret  string `json:"webhook_secret"`
	PublishableKey string `json:"publishable_key"`
	SecretKey      string `json:"secret_key"`
}

// StripePaymentIntent represents a Stripe payment intent
type StripePaymentIntent struct {
	ID               string                `json:"id"`
	Amount           int64                 `json:"amount"`
	Currency         string                `json:"currency"`
	Customer         *stripe.Customer      `json:"customer"`
	LastPaymentError *stripe.Error         `json:"last_payment_error"`
	PaymentMethod    *stripe.PaymentMethod `json:"payment_method"`
	Status           string                `json:"status"`
	Created          int64                 `json:"created"`
	RawData          json.RawMessage       `json:"-"`
}

// StripeInvoice represents a Stripe invoice
type StripeInvoice struct {
	ID         string           `json:"id"`
	AmountDue  int64            `json:"amount_due"`
	AmountPaid int64            `json:"amount_paid"`
	Currency   string           `json:"currency"`
	Customer   *stripe.Customer `json:"customer"`
	Status     string           `json:"status"`
	DueDate    *int64           `json:"due_date"`
	Created    int64            `json:"created"`
	RawData    json.RawMessage  `json:"-"`
}

// StripeCustomer represents a Stripe customer
type StripeCustomer struct {
	ID      string          `json:"id"`
	Email   string          `json:"email"`
	Name    string          `json:"name"`
	Created int64           `json:"created"`
	RawData json.RawMessage `json:"-"`
}

// NewStripeMediator creates a new Stripe mediator
func NewStripeMediator(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *StripeMediator {
	base := NewBaseMediator(config, eventBus, logger)

	mediator := &StripeMediator{
		BaseMediator: base,
		eventBus:     eventBus,
	}

	return mediator
}

// GetProviderID returns the provider ID
func (s *StripeMediator) GetProviderID() string {
	return "stripe"
}

// GetProviderName returns the provider name
func (s *StripeMediator) GetProviderName() string {
	return "Stripe"
}

// GetProviderType returns the provider type
func (s *StripeMediator) GetProviderType() ProviderType {
	return ProviderTypeWebhook
}

// Connect establishes connection to Stripe
func (s *StripeMediator) Connect(ctx context.Context, config *ProviderConfig) error {
	if config.APIConfig == nil || config.APIConfig.APIKey == "" {
		return fmt.Errorf("API configuration required for Stripe")
	}

	// Get webhook secret from webhook config
	if config.WebhookConfig != nil {
		s.webhookSecret = config.WebhookConfig.Secret
	}

	// Set the global Stripe key
	stripe.Key = config.APIConfig.APIKey

	// Validate connection
	if err := s.validateConnection(ctx); err != nil {
		return fmt.Errorf("failed to validate Stripe connection: %w", err)
	}

	s.isConnected = true
	s.connectedAt = &time.Time{}
	*s.connectedAt = time.Now()

	s.logger.Info("Stripe mediator connected successfully",
		zap.String("company_id", config.CompanyID))

	return nil
}

// Disconnect disconnects from Stripe
func (s *StripeMediator) Disconnect(ctx context.Context) error {
	s.isConnected = false
	s.connectedAt = nil

	s.logger.Info("Stripe mediator disconnected",
		zap.String("provider_id", s.config.ProviderID))

	return nil
}

// ProcessWebhook processes incoming Stripe webhooks
func (s *StripeMediator) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.IsConnected() {
		return fmt.Errorf("Stripe mediator not connected")
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	// Process based on event type
	switch event.Type {
	case "payment_intent.payment_failed":
		return s.processPaymentFailure(ctx, &event)
	case "invoice.payment_failed":
		return s.processInvoiceFailure(ctx, &event)
	case "charge.failed":
		return s.processChargeFailure(ctx, &event)
	default:
		s.logger.Debug("Unhandled Stripe event type", zap.String("event_type", event.Type))
		return nil
	}
}

// GetPaymentFailures retrieves payment failures from Stripe
func (s *StripeMediator) GetPaymentFailures(ctx context.Context, since time.Time) ([]*PaymentFailure, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("Stripe mediator not connected")
	}

	// For Stripe, we primarily get payment failures via webhooks
	// This method could be used to query historical data if needed
	// For now, return empty slice as failures are processed via webhooks
	return []*PaymentFailure{}, nil
}

// GetInvoices retrieves invoices from Stripe
func (s *StripeMediator) GetInvoices(ctx context.Context, since time.Time) ([]*Invoice, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("Stripe mediator not connected")
	}

	// This would be implemented to get invoices from Stripe API
	// For now, return empty slice
	return []*Invoice{}, nil
}

// GetCustomers retrieves customers from Stripe
func (s *StripeMediator) GetCustomers(ctx context.Context) ([]*Customer, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("Stripe mediator not connected")
	}

	// This would be implemented to get customers from Stripe API
	// For now, return empty slice
	return []*Customer{}, nil
}

// mapRiskScoreToPriority maps a risk score to a priority level
func (s *StripeMediator) mapRiskScoreToPriority(riskScore float64) PaymentFailurePriority {
	switch {
	case riskScore >= 80:
		return PaymentFailurePriorityCritical
	case riskScore >= 60:
		return PaymentFailurePriorityHigh
	case riskScore >= 40:
		return PaymentFailurePriorityMedium
	default:
		return PaymentFailurePriorityLow
	}
}

// StartSync starts the synchronization process
func (s *StripeMediator) StartSync(ctx context.Context) error {
	if !s.IsConnected() {
		return fmt.Errorf("Stripe mediator not connected")
	}

	return s.BaseMediator.StartSync(ctx)
}

// performSync performs the actual synchronization
func (s *StripeMediator) performSync(ctx context.Context) error {
	startTime := time.Now()

	// Update sync status
	s.syncMutex.Lock()
	s.syncStatus.Status = "syncing"
	s.syncStatus.LastSyncAt = &startTime
	s.syncMutex.Unlock()

	// For Stripe, sync is primarily webhook-driven
	// This method could be used for historical data sync if needed

	// Update sync status
	s.syncMutex.Lock()
	s.syncStatus.Status = "completed"
	s.syncStatus.LastSyncAt = &startTime
	s.syncStatus.NextSyncAt = &time.Time{}
	*s.syncStatus.NextSyncAt = time.Now().Add(s.config.SyncConfig.Frequency)
	s.syncMutex.Unlock()

	s.logger.Info("Stripe sync completed successfully",
		zap.String("provider_id", s.config.ProviderID),
		zap.Duration("sync_duration", time.Since(startTime)))

	return nil
}

// validateConnection validates the Stripe connection
func (s *StripeMediator) validateConnection(ctx context.Context) error {
	// For Stripe, we can validate by making a simple API call
	// or checking if the API key is valid
	// For now, assume connection is valid if API key is provided
	return nil
}

// processPaymentFailure processes a payment failure event
func (s *StripeMediator) processPaymentFailure(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	// Map to unified model
	failure := s.mapStripePaymentIntentToFailure(&paymentIntent)

	// Publish event
	if err := s.publishEvent(ctx, "payment.failure.detected", failure); err != nil {
		return fmt.Errorf("failed to publish payment failure event: %w", err)
	}

	s.logger.Info("Payment failure event processed and published",
		zap.String("event_id", event.ID),
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.String("company_id", s.config.CompanyID))

	return nil
}

// processInvoiceFailure processes an invoice failure event
func (s *StripeMediator) processInvoiceFailure(ctx context.Context, event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Map to unified model
	failure := s.mapStripeInvoiceToFailure(&invoice)

	// Publish event
	if err := s.publishEvent(ctx, "payment.failure.detected", failure); err != nil {
		return fmt.Errorf("failed to publish invoice failure event: %w", err)
	}

	s.logger.Info("Invoice failure event processed and published",
		zap.String("event_id", event.ID),
		zap.String("invoice_id", invoice.ID),
		zap.String("company_id", s.config.CompanyID))

	return nil
}

// processChargeFailure processes a charge failure event
func (s *StripeMediator) processChargeFailure(ctx context.Context, event *stripe.Event) error {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		return fmt.Errorf("failed to unmarshal charge: %w", err)
	}

	// Map to unified model
	failure := s.mapStripeChargeToFailure(&charge)

	// Publish event
	if err := s.publishEvent(ctx, "payment.failure.detected", failure); err != nil {
		return fmt.Errorf("failed to publish charge failure event: %w", err)
	}

	s.logger.Info("Charge failure event processed and published",
		zap.String("event_id", event.ID),
		zap.String("charge_id", charge.ID),
		zap.String("company_id", s.config.CompanyID))

	return nil
}

// mapStripePaymentIntentToFailure maps a Stripe payment intent to unified PaymentFailure model
func (s *StripeMediator) mapStripePaymentIntentToFailure(paymentIntent *stripe.PaymentIntent) *PaymentFailure {
	// Generate unique event ID
	eventID := s.generateEventID()

	// Calculate risk score based on amount and failure reason
	riskScore := s.calculateRiskScore(float64(paymentIntent.Amount)/100, paymentIntent.LastPaymentError)

	// Map to unified PaymentFailure model
	failure := &PaymentFailure{
		ProviderID:        "stripe",
		ProviderEventID:   eventID,
		ProviderEventType: "payment_intent.payment_failed",
		Amount:            float64(paymentIntent.Amount) / 100, // Convert from cents
		Currency:          string(paymentIntent.Currency),
		CustomerID:        paymentIntent.Customer.ID,
		CustomerEmail:     s.getCustomerEmail(paymentIntent.Customer),
		CustomerName:      s.getCustomerName(paymentIntent.Customer),
		FailureReason:     s.mapFailureReason(paymentIntent.LastPaymentError),
		FailureCode:       string(paymentIntent.LastPaymentError.Code),
		FailureMessage:    paymentIntent.LastPaymentError.Msg,
		Status:            "received",
		Priority:          s.mapRiskScoreToPriority(riskScore),
		RiskScore:         riskScore,
		OccurredAt:        time.Unix(paymentIntent.Created, 0),
		DetectedAt:        time.Now(),
		SyncSource:        "webhook",
		RawData:           s.serializeStripeEvent(paymentIntent),
		ProviderMetadata: map[string]interface{}{
			"stripe_payment_intent_id": paymentIntent.ID,
			"stripe_customer_id":       paymentIntent.Customer.ID,
			"stripe_payment_method":    string(paymentIntent.PaymentMethod.Type),
		},
	}

	return failure
}

// mapStripeInvoiceToFailure maps a Stripe invoice to unified PaymentFailure model
func (s *StripeMediator) mapStripeInvoiceToFailure(invoice *stripe.Invoice) *PaymentFailure {
	// Generate unique event ID
	eventID := s.generateEventID()

	// Calculate risk score based on amount
	riskScore := s.calculateRiskScore(float64(invoice.AmountDue)/100, nil)

	// Map to unified PaymentFailure model
	failure := &PaymentFailure{
		ProviderID:        "stripe",
		ProviderEventID:   eventID,
		ProviderEventType: "invoice.payment_failed",
		Amount:            float64(invoice.AmountDue) / 100, // Convert from cents
		Currency:          string(invoice.Currency),
		CustomerID:        invoice.Customer.ID,
		CustomerEmail:     s.getCustomerEmail(invoice.Customer),
		CustomerName:      s.getCustomerName(invoice.Customer),
		FailureReason:     "invoice_payment_failed",
		FailureCode:       "INVOICE_PAYMENT_FAILED",
		FailureMessage:    "Invoice payment failed",
		Status:            "received",
		Priority:          s.mapRiskScoreToPriority(riskScore),
		RiskScore:         riskScore,
		OccurredAt:        time.Unix(invoice.Created, 0),
		DetectedAt:        time.Now(),
		SyncSource:        "webhook",
		RawData:           s.serializeStripeEvent(invoice),
		ProviderMetadata: map[string]interface{}{
			"stripe_invoice_id":  invoice.ID,
			"stripe_customer_id": invoice.Customer.ID,
		},
	}

	return failure
}

// mapStripeChargeToFailure maps a Stripe charge to unified PaymentFailure model
func (s *StripeMediator) mapStripeChargeToFailure(charge *stripe.Charge) *PaymentFailure {
	// Generate unique event ID
	eventID := s.generateEventID()

	// Calculate risk score based on amount
	riskScore := s.calculateRiskScore(float64(charge.Amount)/100, nil)

	// Map to unified PaymentFailure model
	failure := &PaymentFailure{
		ProviderID:        "stripe",
		ProviderEventID:   eventID,
		ProviderEventType: "charge.failed",
		Amount:            float64(charge.Amount) / 100, // Convert from cents
		Currency:          string(charge.Currency),
		CustomerID:        charge.Customer.ID,
		CustomerEmail:     s.getCustomerEmail(charge.Customer),
		CustomerName:      s.getCustomerName(charge.Customer),
		FailureReason:     "charge_failed",
		FailureCode:       "CHARGE_FAILED",
		FailureMessage:    "Charge failed",
		Status:            "received",
		Priority:          s.mapRiskScoreToPriority(riskScore),
		RiskScore:         riskScore,
		OccurredAt:        time.Unix(charge.Created, 0),
		DetectedAt:        time.Now(),
		SyncSource:        "webhook",
		RawData:           s.serializeStripeEvent(charge),
		ProviderMetadata: map[string]interface{}{
			"stripe_charge_id":   charge.ID,
			"stripe_customer_id": charge.Customer.ID,
		},
	}

	return failure
}

// getCustomerEmail extracts customer email from Stripe customer
func (s *StripeMediator) getCustomerEmail(customer *stripe.Customer) string {
	if customer == nil {
		return ""
	}
	return customer.Email
}

// getCustomerName extracts customer name from Stripe customer
func (s *StripeMediator) getCustomerName(customer *stripe.Customer) string {
	if customer == nil {
		return ""
	}
	return customer.Name
}

// mapFailureReason maps Stripe failure reason to unified reason
func (s *StripeMediator) mapFailureReason(lastError *stripe.Error) string {
	if lastError == nil {
		return "unknown"
	}

	switch lastError.Code {
	case "card_declined":
		return "card_declined"
	case "insufficient_funds":
		return "insufficient_funds"
	case "expired_card":
		return "expired_card"
	case "incorrect_cvc":
		return "incorrect_cvc"
	case "processing_error":
		return "processing_error"
	default:
		return "other"
	}
}

// calculateRiskScore calculates risk score based on amount and failure details
func (s *StripeMediator) calculateRiskScore(amount float64, lastError *stripe.Error) float64 {
	// Base risk score starts at 50
	riskScore := 50.0

	// Increase risk based on amount (higher amounts = higher risk)
	if amount > 10000 {
		riskScore += 30
	} else if amount > 5000 {
		riskScore += 20
	} else if amount > 1000 {
		riskScore += 10
	}

	// Increase risk based on failure reason
	if lastError != nil {
		switch lastError.Code {
		case "insufficient_funds":
			riskScore += 20 // High risk - customer may be in financial trouble
		case "expired_card":
			riskScore += 10 // Medium risk - customer may not be active
		case "card_declined":
			riskScore += 15 // Medium-high risk
		case "processing_error":
			riskScore += 5 // Low risk - technical issue
		}
	}

	// Cap at 100
	if riskScore > 100 {
		riskScore = 100
	}

	return riskScore
}

// serializeStripeEvent serializes Stripe event data
func (s *StripeMediator) serializeStripeEvent(event interface{}) json.RawMessage {
	data, err := json.Marshal(event)
	if err != nil {
		return nil
	}
	return data
}

// publishEvent publishes an event to the event bus
func (s *StripeMediator) publishEvent(ctx context.Context, topic string, event interface{}) error {
	if s.eventBus == nil {
		return fmt.Errorf("event bus not configured")
	}

	return s.eventBus.Publish(ctx, topic, event)
}

// generateEventID generates a unique event ID
func (s *StripeMediator) generateEventID() string {
	return uuid.New().String()
}

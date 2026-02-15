package mediators

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/models"
)

// RecoveryActionInterface defines recovery actions for payment providers
type RecoveryActionInterface interface {
	RetryPayment(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentRetryConfig) (*RecoveryActionResult, error)
	UpdatePaymentMethod(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *UpdatePaymentMethodConfig) (*RecoveryActionResult, error)
	SendPaymentLink(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentLinkConfig) (*RecoveryActionResult, error)
	GetRecoveryCapabilities() []string
}

// PaymentRetryConfig represents configuration for payment retry
type PaymentRetryConfig struct {
	NewAmount    *float64 `json:"new_amount,omitempty"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	RetryReason  string   `json:"retry_reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePaymentMethodConfig represents configuration for updating payment method
type UpdatePaymentMethodConfig struct {
	NewPaymentMethodID string `json:"new_payment_method_id"`
	CustomerID         string `json:"customer_id"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentLinkConfig represents configuration for sending payment links
type PaymentLinkConfig struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RecoveryActionResult represents the result of a recovery action
type RecoveryActionResult struct {
	Success      bool                   `json:"success"`
	ActionID     string                 `json:"action_id"`
	ExternalID   string                 `json:"external_id,omitempty"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	NextRetryAt  *time.Time             `json:"next_retry_at,omitempty"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// StripeRecoveryActions implements recovery actions for Stripe
type StripeRecoveryActions struct {
	mediator *StripeMediator
	tracer   trace.Tracer
}

// NewStripeRecoveryActions creates a new Stripe recovery actions handler
func NewStripeRecoveryActions(mediator *StripeMediator) *StripeRecoveryActions {
	return &StripeRecoveryActions{
		mediator: mediator,
		tracer:   otel.Tracer("stripe-recovery-actions"),
	}
}

// GetRecoveryCapabilities returns the recovery capabilities for Stripe
func (s *StripeRecoveryActions) GetRecoveryCapabilities() []string {
	return []string{
		"retry_payment",
		"update_payment_method",
		"send_payment_link",
		"create_invoice",
		"send_invoice_reminder",
	}
}

// RetryPayment retries a failed payment in Stripe
func (s *StripeRecoveryActions) RetryPayment(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentRetryConfig) (*RecoveryActionResult, error) {
	ctx, span := s.tracer.Start(ctx, "stripe_retry_payment")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.String("transaction_id", paymentFailure.TransactionID),
		attribute.Float64("original_amount", paymentFailure.Amount),
	)

	// For now, simulate Stripe API call
	// In production, this would use the actual Stripe SDK
	
	// Determine retry amount
	retryAmount := paymentFailure.Amount
	if config.NewAmount != nil {
		retryAmount = *config.NewAmount
		span.SetAttributes(attribute.Float64("retry_amount", retryAmount))
	}

	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// Simulate success/failure based on amount (for demo purposes)
	success := retryAmount < 1000 // Smaller amounts more likely to succeed

	if success {
		return &RecoveryActionResult{
			Success:    true,
			ActionID:   fmt.Sprintf("stripe_retry_%d", time.Now().Unix()),
			ExternalID: fmt.Sprintf("pi_%d", time.Now().Unix()),
			Status:     "succeeded",
			Message:    "Payment retry successful",
			Data: map[string]interface{}{
				"amount":         retryAmount,
				"currency":       paymentFailure.Currency,
				"payment_method": config.PaymentMethod,
				"retry_reason":   config.RetryReason,
			},
		}, nil
	} else {
		return &RecoveryActionResult{
			Success:      false,
			ActionID:     fmt.Sprintf("stripe_retry_%d", time.Now().Unix()),
			Status:       "failed",
			ErrorCode:    "card_declined",
			ErrorMessage: "Your card was declined",
			NextRetryAt:  timePtr(time.Now().Add(24 * time.Hour)),
		}, nil
	}
}

// UpdatePaymentMethod updates the payment method for a customer in Stripe
func (s *StripeRecoveryActions) UpdatePaymentMethod(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *UpdatePaymentMethodConfig) (*RecoveryActionResult, error) {
	ctx, span := s.tracer.Start(ctx, "stripe_update_payment_method")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.String("customer_id", config.CustomerID),
		attribute.String("new_payment_method_id", config.NewPaymentMethodID),
	)

	// Simulate API call
	time.Sleep(150 * time.Millisecond)

	return &RecoveryActionResult{
		Success:    true,
		ActionID:   fmt.Sprintf("stripe_update_pm_%d", time.Now().Unix()),
		ExternalID: config.NewPaymentMethodID,
		Status:     "updated",
		Message:    "Payment method updated successfully",
		Data: map[string]interface{}{
			"customer_id":           config.CustomerID,
			"new_payment_method_id": config.NewPaymentMethodID,
			"updated_at":           time.Now(),
		},
	}, nil
}

// SendPaymentLink creates and sends a payment link via Stripe
func (s *StripeRecoveryActions) SendPaymentLink(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentLinkConfig) (*RecoveryActionResult, error) {
	ctx, span := s.tracer.Start(ctx, "stripe_send_payment_link")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.Float64("amount", config.Amount),
		attribute.String("currency", config.Currency),
	)

	// Simulate API call
	time.Sleep(200 * time.Millisecond)

	paymentLinkID := fmt.Sprintf("plink_%d", time.Now().Unix())
	paymentLinkURL := fmt.Sprintf("https://checkout.stripe.com/pay/%s", paymentLinkID)

	return &RecoveryActionResult{
		Success:    true,
		ActionID:   fmt.Sprintf("stripe_payment_link_%d", time.Now().Unix()),
		ExternalID: paymentLinkID,
		Status:     "created",
		Message:    "Payment link created successfully",
		Data: map[string]interface{}{
			"payment_link_id":  paymentLinkID,
			"payment_link_url": paymentLinkURL,
			"amount":          config.Amount,
			"currency":        config.Currency,
			"description":     config.Description,
			"expires_at":      config.ExpiresAt,
		},
	}, nil
}

// XeroRecoveryActions implements recovery actions for Xero
type XeroRecoveryActions struct {
	mediator *XeroMediator
	tracer   trace.Tracer
}

// NewXeroRecoveryActions creates a new Xero recovery actions handler
func NewXeroRecoveryActions(mediator *XeroMediator) *XeroRecoveryActions {
	return &XeroRecoveryActions{
		mediator: mediator,
		tracer:   otel.Tracer("xero-recovery-actions"),
	}
}

// GetRecoveryCapabilities returns the recovery capabilities for Xero
func (x *XeroRecoveryActions) GetRecoveryCapabilities() []string {
	return []string{
		"send_invoice_reminder",
		"update_invoice_due_date",
		"create_payment_link",
		"mark_invoice_sent",
	}
}

// RetryPayment for Xero focuses on invoice-based recovery
func (x *XeroRecoveryActions) RetryPayment(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentRetryConfig) (*RecoveryActionResult, error) {
	ctx, span := x.tracer.Start(ctx, "xero_retry_payment")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
		attribute.String("transaction_id", paymentFailure.TransactionID),
	)

	// For Xero, "retry payment" typically means sending invoice reminders
	// or updating invoice terms
	
	// Simulate API call
	time.Sleep(120 * time.Millisecond)

	return &RecoveryActionResult{
		Success:    true,
		ActionID:   fmt.Sprintf("xero_invoice_reminder_%d", time.Now().Unix()),
		ExternalID: paymentFailure.TransactionID,
		Status:     "sent",
		Message:    "Invoice reminder sent successfully",
		Data: map[string]interface{}{
			"invoice_id":     paymentFailure.TransactionID,
			"reminder_type":  "overdue",
			"sent_at":       time.Now(),
			"retry_reason":   config.RetryReason,
		},
	}, nil
}

// UpdatePaymentMethod for Xero updates invoice payment terms
func (x *XeroRecoveryActions) UpdatePaymentMethod(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *UpdatePaymentMethodConfig) (*RecoveryActionResult, error) {
	ctx, span := x.tracer.Start(ctx, "xero_update_payment_terms")
	defer span.End()

	// For Xero, this might involve updating invoice payment terms or due dates
	time.Sleep(100 * time.Millisecond)

	return &RecoveryActionResult{
		Success:    true,
		ActionID:   fmt.Sprintf("xero_update_terms_%d", time.Now().Unix()),
		ExternalID: paymentFailure.TransactionID,
		Status:     "updated",
		Message:    "Invoice payment terms updated",
		Data: map[string]interface{}{
			"invoice_id":    paymentFailure.TransactionID,
			"updated_terms": "Net 30",
			"updated_at":    time.Now(),
		},
	}, nil
}

// SendPaymentLink creates a payment link for Xero invoice
func (x *XeroRecoveryActions) SendPaymentLink(ctx context.Context, paymentFailure *models.PaymentFailureEvent, config *PaymentLinkConfig) (*RecoveryActionResult, error) {
	ctx, span := x.tracer.Start(ctx, "xero_send_payment_link")
	defer span.End()

	// Simulate creating a Xero payment link
	time.Sleep(180 * time.Millisecond)

	paymentLinkID := fmt.Sprintf("xero_plink_%d", time.Now().Unix())
	paymentLinkURL := fmt.Sprintf("https://invoice.xero.com/pay/%s", paymentLinkID)

	return &RecoveryActionResult{
		Success:    true,
		ActionID:   fmt.Sprintf("xero_payment_link_%d", time.Now().Unix()),
		ExternalID: paymentLinkID,
		Status:     "created",
		Message:    "Xero payment link created successfully",
		Data: map[string]interface{}{
			"payment_link_id":  paymentLinkID,
			"payment_link_url": paymentLinkURL,
			"invoice_id":       paymentFailure.TransactionID,
			"amount":          config.Amount,
			"currency":        config.Currency,
		},
	}, nil
}

// RecoveryActionMediator coordinates recovery actions across providers
type RecoveryActionMediator struct {
	stripeActions *StripeRecoveryActions
	xeroActions   *XeroRecoveryActions
	tracer        trace.Tracer
}

// NewRecoveryActionMediator creates a new recovery action mediator
func NewRecoveryActionMediator(stripeMediator *StripeMediator, xeroMediator *XeroMediator) *RecoveryActionMediator {
	return &RecoveryActionMediator{
		stripeActions: NewStripeRecoveryActions(stripeMediator),
		xeroActions:   NewXeroRecoveryActions(xeroMediator),
		tracer:        otel.Tracer("recovery-action-mediator"),
	}
}

// ExecuteRecoveryAction executes a recovery action based on the provider
func (r *RecoveryActionMediator) ExecuteRecoveryAction(ctx context.Context, provider, actionType string, paymentFailure *models.PaymentFailureEvent, config map[string]interface{}) (*RecoveryActionResult, error) {
	ctx, span := r.tracer.Start(ctx, "execute_recovery_action")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider", provider),
		attribute.String("action_type", actionType),
		attribute.String("payment_failure_id", paymentFailure.ID.String()),
	)

	var recoveryInterface RecoveryActionInterface

	// Select appropriate recovery action handler
	switch provider {
	case "stripe":
		recoveryInterface = r.stripeActions
	case "xero":
		recoveryInterface = r.xeroActions
	default:
		err := fmt.Errorf("unsupported provider for recovery actions: %s", provider)
		span.RecordError(err)
		return nil, err
	}

	// Execute the specific action
	switch actionType {
	case "retry_payment":
		retryConfig := &PaymentRetryConfig{}
		if err := mapToStruct(config, retryConfig); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse retry config: %w", err)
		}
		return recoveryInterface.RetryPayment(ctx, paymentFailure, retryConfig)

	case "update_payment_method":
		updateConfig := &UpdatePaymentMethodConfig{}
		if err := mapToStruct(config, updateConfig); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse update payment method config: %w", err)
		}
		return recoveryInterface.UpdatePaymentMethod(ctx, paymentFailure, updateConfig)

	case "send_payment_link":
		linkConfig := &PaymentLinkConfig{}
		if err := mapToStruct(config, linkConfig); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse payment link config: %w", err)
		}
		return recoveryInterface.SendPaymentLink(ctx, paymentFailure, linkConfig)

	default:
		err := fmt.Errorf("unsupported action type: %s", actionType)
		span.RecordError(err)
		return nil, err
	}
}

// GetProviderCapabilities returns recovery capabilities for a provider
func (r *RecoveryActionMediator) GetProviderCapabilities(provider string) ([]string, error) {
	switch provider {
	case "stripe":
		return r.stripeActions.GetRecoveryCapabilities(), nil
	case "xero":
		return r.xeroActions.GetRecoveryCapabilities(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func mapToStruct(input map[string]interface{}, output interface{}) error {
	// Simple implementation - in production, use a proper mapping library
	// like mapstructure or similar
	
	// For now, we'll just return nil to indicate successful mapping
	// The actual implementation would convert the map to the struct
	return nil
}

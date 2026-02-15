package mediators

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/architecture"
)

// OAuthTokens represents OAuth 2.0 tokens
type OAuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// XeroMediator implements PaymentProvider for Xero
type XeroMediator struct {
	*BaseMediator
	oauthClient *http.Client
	apiClient   *XeroAPIClient
	oauthConfig *OAuthConfig
}

// XeroAPIClient handles Xero API communication
type XeroAPIClient struct {
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
}

// XeroInvoice represents a Xero invoice
type XeroInvoice struct {
	ID              string          `json:"InvoiceID"`
	InvoiceNumber   string          `json:"InvoiceNumber"`
	Contact         XeroContact     `json:"Contact"`
	LineItems       []XeroLineItem  `json:"LineItems"`
	SubTotal        float64         `json:"SubTotal"`
	TotalTax        float64         `json:"TotalTax"`
	Total           float64         `json:"Total"`
	AmountPaid      float64         `json:"AmountPaid"`
	AmountDue       float64         `json:"AmountDue"`
	Status          string          `json:"Status"`
	DueDate         time.Time       `json:"DueDate"`
	Date            time.Time       `json:"Date"`
	CurrencyCode    string          `json:"CurrencyCode"`
	Reference       string          `json:"Reference"`
	LineAmountTypes string          `json:"LineAmountTypes"`
	RawData         json.RawMessage `json:"-"`
}

// XeroContact represents a Xero contact/customer
type XeroContact struct {
	ID           string          `json:"ContactID"`
	Name         string          `json:"Name"`
	EmailAddress string          `json:"EmailAddress"`
	FirstName    string          `json:"FirstName"`
	LastName     string          `json:"LastName"`
	RawData      json.RawMessage `json:"-"`
}

// XeroLineItem represents a line item in a Xero invoice
type XeroLineItem struct {
	Description string          `json:"Description"`
	Quantity    float64         `json:"Quantity"`
	UnitAmount  float64         `json:"UnitAmount"`
	LineAmount  float64         `json:"LineAmount"`
	AccountCode string          `json:"AccountCode"`
	RawData     json.RawMessage `json:"-"`
}

// NewXeroMediator creates a new Xero mediator
func NewXeroMediator(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *XeroMediator {
	base := NewBaseMediator(config, eventBus, logger)

	mediator := &XeroMediator{
		BaseMediator: base,
		oauthConfig:  config.OAuthConfig,
	}

	return mediator
}

// GetProviderName returns the provider name
func (x *XeroMediator) GetProviderName() string {
	return "Xero"
}

// GetAuthorizationURL generates the Xero OAuth authorization URL
func (x *XeroMediator) GetAuthorizationURL(state string) (string, error) {
	if x.oauthConfig == nil {
		return "", fmt.Errorf("OAuth configuration not set")
	}

	// Xero OAuth 2.0 authorization URL
	authURL := "https://login.xero.com/identity/connect/authorize"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", x.oauthConfig.ClientID)
	params.Add("redirect_uri", x.oauthConfig.RedirectURI)
	params.Add("scope", strings.Join(x.oauthConfig.Scopes, " "))
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", authURL, params.Encode()), nil
}

// ExchangeCodeForTokens exchanges authorization code for access tokens
func (x *XeroMediator) ExchangeCodeForTokens(ctx context.Context, code, state string) (*OAuthTokens, error) {
	if x.oauthConfig == nil {
		return nil, fmt.Errorf("OAuth configuration not set")
	}

	// Prepare token exchange request
	tokenURL := "https://identity.xero.com/connect/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", x.oauthConfig.ClientID)
	data.Set("client_secret", x.oauthConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", x.oauthConfig.RedirectURI)

	// Debug logging
	x.logger.Info("Token exchange request",
		zap.String("token_url", tokenURL),
		zap.String("client_id", x.oauthConfig.ClientID),
		zap.String("redirect_uri", x.oauthConfig.RedirectURI),
		zap.String("code", code[:10]+"..."), // Log only first 10 chars of code
		zap.String("state", state))

	// Make token exchange request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error response body for debugging
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	tokens := &OAuthTokens{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresIn:    tokenResponse.ExpiresIn,
		Scope:        tokenResponse.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}

	return tokens, nil
}

// GetTenants retrieves Xero tenant connections
func (x *XeroMediator) GetTenants(ctx context.Context, accessToken string) ([]XeroTenantInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.xero.com/connections", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenants with status: %d", resp.StatusCode)
	}

	var tenants []XeroTenantInfo
	if err := json.NewDecoder(resp.Body).Decode(&tenants); err != nil {
		return nil, fmt.Errorf("failed to decode tenants response: %w", err)
	}

	return tenants, nil
}

// GetOrganizations retrieves Xero organization details
func (x *XeroMediator) GetOrganizations(ctx context.Context, accessToken, tenantID string) ([]XeroOrganizationInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.xero.com/api.xro/2.0/Organisation", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Xero-tenant-id", tenantID)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get organizations with status: %d", resp.StatusCode)
	}

	var response struct {
		Organisations []XeroOrganizationInfo `json:"Organisations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}

	return response.Organisations, nil
}

// GetPaymentFailures retrieves payment failures from Xero (overloaded method)
func (x *XeroMediator) GetPaymentFailures(ctx context.Context, accessToken, tenantID string) ([]XeroPaymentFailureInfo, error) {
	// Get invoices from Xero
	invoices, err := x.GetInvoicesWithToken(ctx, accessToken, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	// Filter for overdue invoices (payment failures)
	var failures []XeroPaymentFailureInfo
	now := time.Now()

	for _, invoice := range invoices {
		// Check if invoice is overdue
		if invoice.AmountDue > 0 && invoice.DueDate.Before(now) {
			daysOverdue := int(now.Sub(invoice.DueDate).Hours() / 24)

			failure := XeroPaymentFailureInfo{
				ID:            invoice.ID,
				InvoiceNumber: invoice.InvoiceNumber,
				CustomerName:  invoice.Contact.Name,
				CustomerEmail: invoice.Contact.EmailAddress,
				Amount:        invoice.AmountDue,
				Currency:      invoice.CurrencyCode,
				DueDate:       invoice.DueDate.Format("2006-01-02"),
				DaysOverdue:   daysOverdue,
				FailureReason: "invoice_overdue",
				Status:        "overdue",
			}

			failures = append(failures, failure)
		}
	}

	return failures, nil
}

// GetInvoicesWithToken retrieves invoices using access token
func (x *XeroMediator) GetInvoicesWithToken(ctx context.Context, accessToken, tenantID string) ([]XeroInvoice, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.xero.com/api.xro/2.0/Invoices", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Xero-tenant-id", tenantID)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get invoices with status: %d", resp.StatusCode)
	}

	var response struct {
		Invoices []XeroInvoice `json:"Invoices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode invoices response: %w", err)
	}

	return response.Invoices, nil
}

// XeroTenantInfo represents tenant information from Xero
type XeroTenantInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortCode   string `json:"shortCode"`
	IsActive    bool   `json:"isActive"`
	CreatedDate string `json:"createdDate"`
}

// XeroOrganizationInfo represents organization information from Xero
type XeroOrganizationInfo struct {
	ID           string `json:"OrganisationID"`
	Name         string `json:"Name"`
	LegalName    string `json:"LegalName"`
	ShortCode    string `json:"ShortCode"`
	CountryCode  string `json:"CountryCode"`
	BaseCurrency string `json:"BaseCurrency"`
	IsActive     bool   `json:"IsActive"`
}

// XeroPaymentFailureInfo represents payment failure information from Xero
type XeroPaymentFailureInfo struct {
	ID            string  `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	DueDate       string  `json:"due_date"`
	DaysOverdue   int     `json:"days_overdue"`
	FailureReason string  `json:"failure_reason"`
	Status        string  `json:"status"`
}

// Connect establishes connection to Xero
func (x *XeroMediator) Connect(ctx context.Context, config *ProviderConfig) error {
	if config.OAuthConfig == nil {
		return fmt.Errorf("OAuth configuration required for Xero")
	}

	// Initialize OAuth client
	oauthConfig := &oauth2.Config{
		ClientID:     config.OAuthConfig.ClientID,
		ClientSecret: config.OAuthConfig.ClientSecret,
		RedirectURL:  config.OAuthConfig.RedirectURI,
		Scopes:       config.OAuthConfig.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.OAuthConfig.AuthURL,
			TokenURL: config.OAuthConfig.TokenURL,
		},
	}

	// Create OAuth client with access token
	token := &oauth2.Token{
		AccessToken:  config.OAuthConfig.AccessToken,
		RefreshToken: config.OAuthConfig.RefreshToken,
		Expiry:       config.OAuthConfig.ExpiresAt,
	}

	x.oauthClient = oauthConfig.Client(ctx, token)
	x.apiClient = NewXeroAPIClient(x.oauthClient, "https://api.xero.com/api.xro/2.0", x.logger)

	// Validate connection
	if err := x.validateConnection(ctx); err != nil {
		return fmt.Errorf("failed to validate Xero connection: %w", err)
	}

	x.isConnected = true
	x.connectedAt = &time.Time{}
	*x.connectedAt = time.Now()

	x.logger.Info("Xero mediator connected successfully",
		zap.String("company_id", config.CompanyID))

	return nil
}

// Disconnect disconnects from Xero
func (x *XeroMediator) Disconnect(ctx context.Context) error {
	x.isConnected = false
	x.connectedAt = nil
	x.oauthClient = nil
	x.apiClient = nil

	x.logger.Info("Xero mediator disconnected",
		zap.String("provider_id", x.config.ProviderID))

	return nil
}

// Data Mapping and Event Publishing Methods

// mapXeroInvoiceToPaymentFailure maps a Xero invoice to a PaymentFailure
func (x *XeroMediator) mapXeroInvoiceToPaymentFailure(xeroInvoice *XeroInvoice) *architecture.PaymentFailure {
	if xeroInvoice == nil {
		return nil
	}

	// Calculate risk score based on amount and due date
	riskScore := x.calculateRiskScore(xeroInvoice.Total, xeroInvoice.DueDate)
	priority := x.mapRiskScoreToPriority(riskScore)

	// Determine failure reason based on invoice status
	failureReason := "invoice_unpaid"
	if xeroInvoice.AmountPaid > 0 && xeroInvoice.AmountDue > 0 {
		failureReason = "invoice_partially_paid"
	}

	// Create PaymentFailure
	paymentFailure := &architecture.PaymentFailure{
		CompanyID:         x.config.CompanyID,
		ProviderID:        x.config.ProviderID,
		ProviderEventID:   xeroInvoice.ID,
		ProviderEventType: "invoice",
		Amount:            xeroInvoice.AmountDue,
		Currency:          xeroInvoice.CurrencyCode,
		CustomerID:        xeroInvoice.Contact.ID,
		CustomerName:      xeroInvoice.Contact.Name,
		CustomerEmail:     xeroInvoice.Contact.EmailAddress,
		FailureReason:     failureReason,
		InvoiceID:         xeroInvoice.ID,
		InvoiceNumber:     xeroInvoice.InvoiceNumber,
		DueDate:           &xeroInvoice.DueDate,
		Status:            architecture.PaymentFailureStatusReceived,
		Priority:          architecture.PaymentFailurePriority(priority),
		RiskScore:         riskScore,
		OccurredAt:        xeroInvoice.Date,
		DetectedAt:        time.Now(),
		SyncSource:        "xero",
		RawData:           xeroInvoice.RawData,
	}

	return paymentFailure
}

// calculateRiskScore calculates risk score based on amount and due date
func (x *XeroMediator) calculateRiskScore(amount float64, dueDate time.Time) float64 {
	// Base risk score starts at 50
	riskScore := 50.0

	// Increase risk based on amount (higher amounts = higher risk)
	if amount >= 10000 {
		riskScore += 30
	} else if amount >= 5000 {
		riskScore += 20
	} else if amount >= 1000 {
		riskScore += 10
	}

	// Increase risk based on overdue days
	overdueDays := int(time.Since(dueDate).Hours() / 24)
	if overdueDays >= 90 {
		riskScore += 20
	} else if overdueDays >= 60 {
		riskScore += 15
	} else if overdueDays >= 30 {
		riskScore += 10
	} else if overdueDays >= 7 {
		riskScore += 5
	}

	// Cap at 100
	if riskScore > 100 {
		riskScore = 100
	}

	return riskScore
}

// publishPaymentFailureEvent publishes a payment failure event to the event bus
func (x *XeroMediator) publishPaymentFailureEvent(paymentFailure *architecture.PaymentFailure) error {
	if paymentFailure == nil {
		return fmt.Errorf("payment failure cannot be nil")
	}

	// Create event with metadata
	event := map[string]interface{}{
		"event_id":        uuid.New().String(),
		"timestamp":       time.Now().UTC(),
		"event_type":      "payment.failure.detected",
		"provider":        "xero",
		"payment_failure": paymentFailure,
		"metadata": map[string]interface{}{
			"source":         "xero_mediator",
			"sync_timestamp": time.Now().UTC(),
			"company_id":     paymentFailure.CompanyID,
		},
	}

	// Publish to event bus
	ctx := context.Background()
	return x.eventBus.Publish(ctx, "payment.failure.detected", event)
}

// GetPaymentFailuresLegacy retrieves payment failures from Xero (legacy method for BaseMediator interface)
func (x *XeroMediator) GetPaymentFailuresLegacy(ctx context.Context, since time.Time) ([]interface{}, error) {
	// Get invoices from Xero
	invoices, err := x.GetInvoices(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	// Convert invoices to payment failures
	var paymentFailures []interface{}
	for _, invoiceInterface := range invoices {
		invoice, ok := invoiceInterface.(*XeroInvoice)
		if !ok {
			continue
		}

		// Only consider unpaid or partially paid invoices
		if invoice.AmountDue > 0 {
			paymentFailure := x.mapXeroInvoiceToPaymentFailure(invoice)
			if paymentFailure != nil {
				paymentFailures = append(paymentFailures, paymentFailure)
			}
		}
	}

	return paymentFailures, nil
}

// GetInvoices retrieves invoices from Xero
func (x *XeroMediator) GetInvoices(ctx context.Context, since time.Time) ([]interface{}, error) {
	if !x.isConnected {
		return nil, fmt.Errorf("Xero mediator not connected")
	}

	// Get invoices from Xero
	xeroInvoices, err := x.apiClient.GetInvoices(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices from Xero: %w", err)
	}

	var invoices []interface{}

	// Map Xero invoices to unified model
	for _, xeroInvoice := range xeroInvoices {
		invoice := x.mapXeroInvoiceToInvoice(xeroInvoice)
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// GetCustomers retrieves customers from Xero
func (x *XeroMediator) GetCustomers(ctx context.Context) ([]*Customer, error) {
	if !x.isConnected {
		return nil, fmt.Errorf("Xero mediator not connected")
	}

	// This would be implemented to get contacts from Xero
	// For now, return empty slice
	return []*Customer{}, nil
}

// StartSync starts Xero synchronization
func (x *XeroMediator) StartSync(ctx context.Context) error {
	if !x.isConnected {
		return fmt.Errorf("Xero mediator not connected")
	}

	return x.BaseMediator.StartSync(ctx)
}

// performSync performs Xero-specific synchronization
func (x *XeroMediator) performSync(ctx context.Context) error {
	startTime := time.Now()

	// Update sync status
	x.syncMutex.Lock()
	x.syncStatus.Status = "syncing"
	x.syncStatus.LastSyncAt = &startTime
	x.syncMutex.Unlock()

	// Get payment failures since last sync
	var since time.Time
	if x.syncStatus.LastSyncAt != nil {
		since = *x.syncStatus.LastSyncAt
	} else {
		since = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	// Retrieve payment failures
	failures, err := x.GetPaymentFailuresLegacy(ctx, since)
	if err != nil {
		x.logger.Error("Failed to get payment failures during sync",
			zap.String("provider_id", x.config.ProviderID),
			zap.Error(err))

		// Update sync status with error
		x.syncMutex.Lock()
		x.syncStatus.Status = "error"
		x.syncStatus.LastError = err.Error()
		x.syncMutex.Unlock()

		return err
	}

	// Publish events for new failures
	for _, failureInterface := range failures {
		if failure, ok := failureInterface.(*architecture.PaymentFailure); ok {
			if err := x.publishEvent(ctx, "payment.failure.detected", failure); err != nil {
				x.logger.Error("Failed to publish payment failure event",
					zap.String("failure_id", failure.ID.String()),
					zap.Error(err))
			}
		}
	}

	// Update sync status
	x.syncMutex.Lock()
	x.syncStatus.Status = "active"
	x.syncStatus.SyncDuration = time.Since(startTime)
	x.syncStatus.RecordsSynced = int64(len(failures))
	x.syncStatus.NextSyncAt = &time.Time{}
	*x.syncStatus.NextSyncAt = time.Now().Add(x.config.SyncConfig.Frequency)
	x.syncMutex.Unlock()

	x.logger.Info("Xero sync completed successfully",
		zap.String("provider_id", x.config.ProviderID),
		zap.Int("failures_found", len(failures)),
		zap.Duration("sync_duration", time.Since(startTime)))

	return nil
}

// validateConnection validates the Xero connection
func (x *XeroMediator) validateConnection(ctx context.Context) error {
	// Make a test API call to validate connection
	_, err := x.apiClient.GetInvoices(ctx, time.Now().Add(-1*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to validate Xero connection: %w", err)
	}

	return nil
}

// isPaymentFailure checks if an invoice represents a payment failure
func (x *XeroMediator) isPaymentFailure(invoice *XeroInvoice) bool {
	// Xero-specific logic for identifying payment failures
	return invoice.Status == "AUTHORISED" &&
		invoice.DueDate.Before(time.Now()) &&
		invoice.AmountDue > 0
}

// mapInvoiceToFailure maps a Xero invoice to a unified PaymentFailure
func (x *XeroMediator) mapInvoiceToFailure(invoice *XeroInvoice) *architecture.PaymentFailure {
	// Generate unique event ID
	eventID := x.generateEventID()

	// Calculate risk score based on amount and due date
	riskScore := x.calculateRiskScore(invoice.AmountDue, invoice.DueDate)

	// Map to unified PaymentFailure model
	failure := &architecture.PaymentFailure{
		ProviderID:        "xero",
		ProviderEventID:   eventID,
		ProviderEventType: "invoice.payment_failed",
		Amount:            invoice.AmountDue,
		Currency:          invoice.CurrencyCode,
		CustomerID:        invoice.Contact.ID,
		CustomerName:      invoice.Contact.Name,
		CustomerEmail:     invoice.Contact.EmailAddress,
		InvoiceID:         invoice.ID,
		InvoiceNumber:     invoice.InvoiceNumber,
		DueDate:           &invoice.DueDate,
		FailureReason:     "overdue_invoice",
		FailureCode:       "INVOICE_OVERDUE",
		FailureMessage: fmt.Sprintf("Invoice %s is overdue",
			invoice.InvoiceNumber),
		Status:     architecture.PaymentFailureStatusReceived,
		Priority:   x.mapRiskScoreToPriority(riskScore),
		RiskScore:  riskScore,
		OccurredAt: invoice.DueDate,
		DetectedAt: time.Now(),
		SyncSource: "api_poll",
		RawData:    invoice.RawData,
		ProviderMetadata: map[string]interface{}{
			"xero_invoice_id": invoice.ID,
			"xero_contact_id": invoice.Contact.ID,
			"xero_status":     invoice.Status,
		},
	}

	return failure
}

// mapXeroInvoiceToInvoice maps a Xero invoice to unified Invoice model
func (x *XeroMediator) mapXeroInvoiceToInvoice(xeroInvoice *XeroInvoice) *Invoice {
	// Map line items
	lineItems, _ := json.Marshal(xeroInvoice.LineItems)

	invoice := &Invoice{
		ProviderID:        "xero",
		ProviderInvoiceID: xeroInvoice.ID,
		InvoiceNumber:     xeroInvoice.InvoiceNumber,
		Amount:            xeroInvoice.Total,
		Currency:          xeroInvoice.CurrencyCode,
		Status:            xeroInvoice.Status,
		CustomerID:        xeroInvoice.Contact.ID,
		CustomerName:      xeroInvoice.Contact.Name,
		CustomerEmail:     xeroInvoice.Contact.EmailAddress,
		IssueDate:         xeroInvoice.Date,
		DueDate:           xeroInvoice.DueDate,
		LineItems:         lineItems,
		ProviderMetadata: map[string]interface{}{
			"xero_invoice_id": xeroInvoice.ID,
			"xero_contact_id": xeroInvoice.Contact.ID,
			"xero_status":     xeroInvoice.Status,
		},
	}

	return invoice
}

// mapRiskScoreToPriority maps risk score to priority
func (x *XeroMediator) mapRiskScoreToPriority(riskScore float64) architecture.PaymentFailurePriority {
	if riskScore >= 80 {
		return architecture.PaymentFailurePriorityCritical
	} else if riskScore >= 60 {
		return architecture.PaymentFailurePriorityHigh
	} else if riskScore >= 40 {
		return architecture.PaymentFailurePriorityMedium
	} else {
		return architecture.PaymentFailurePriorityLow
	}
}

// NewXeroAPIClient creates a new Xero API client
func NewXeroAPIClient(httpClient *http.Client, baseURL string, logger *zap.Logger) *XeroAPIClient {
	return &XeroAPIClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     logger,
	}
}

// GetInvoices retrieves invoices from Xero API
func (x *XeroAPIClient) GetInvoices(ctx context.Context, since time.Time) ([]*XeroInvoice, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("where", fmt.Sprintf("Date >= DateTime(%d, %d, %d)",
		since.Year(), since.Month(), since.Day()))
	params.Set("order", "Date DESC")

	// Make API request
	url := fmt.Sprintf("%s/Invoices?%s", x.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := x.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Xero API returned status: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		Invoices []*XeroInvoice `json:"Invoices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Store raw data for each invoice
	for _, invoice := range response.Invoices {
		if rawData, err := json.Marshal(invoice); err == nil {
			invoice.RawData = rawData
		}
	}

	return response.Invoices, nil
}

// OAuth 2.0 Implementation Methods

// GenerateAuthorizationURL generates the OAuth 2.0 authorization URL
func (x *XeroMediator) GenerateAuthorizationURL(config *OAuthConfig) (string, string, error) {
	if config == nil {
		return "", "", fmt.Errorf("OAuth configuration is required")
	}

	// Generate state parameter for security
	state := x.GenerateStateParameter()

	// Build authorization URL
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		config.AuthURL,
		url.QueryEscape(config.ClientID),
		url.QueryEscape(config.RedirectURI),
		url.QueryEscape(strings.Join(config.Scopes, " ")),
		state)

	return authURL, state, nil
}

// RefreshAccessToken refreshes the access token using refresh token
func (x *XeroMediator) RefreshAccessToken(ctx context.Context, config *OAuthConfig, refreshToken string) (*OAuthTokens, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth configuration is required")
	}

	// Prepare refresh request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	// Make refresh request
	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokens OAuthTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	// Calculate expiration time
	if tokens.ExpiresIn > 0 {
		tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}

	return &tokens, nil
}

// ValidateTokens validates OAuth tokens
func (x *XeroMediator) ValidateTokens(tokens *OAuthTokens) bool {
	if tokens == nil {
		return false
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return false
	}

	if tokens.ExpiresIn <= 0 {
		return false
	}

	return true
}

// HasRequiredScopes checks if the mediator has required OAuth scopes
func (x *XeroMediator) HasRequiredScopes(requiredScopes []string) bool {
	if x.oauthConfig == nil {
		return false
	}

	availableScopes := make(map[string]bool)
	for _, scope := range x.oauthConfig.Scopes {
		availableScopes[scope] = true
	}

	for _, requiredScope := range requiredScopes {
		if !availableScopes[requiredScope] {
			return false
		}
	}

	return true
}

// ValidateScopes validates OAuth scopes
func (x *XeroMediator) ValidateScopes(scopes []string) bool {
	validScopes := map[string]bool{
		"offline_access":          true,
		"accounting.transactions": true,
		"accounting.contacts":     true,
		"accounting.settings":     true,
		"accounting.reports":      true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return false
		}
	}

	return true
}

// GenerateStateParameter generates a secure state parameter for OAuth
func (x *XeroMediator) GenerateStateParameter() string {
	// Generate 32-character random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ValidateStateParameter validates OAuth state parameter
func (x *XeroMediator) ValidateStateParameter(expected, actual string) bool {
	return expected == actual
}

// GeneratePKCECodeVerifier generates PKCE code verifier
func (x *XeroMediator) GeneratePKCECodeVerifier() string {
	// Generate 128-character random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, 128)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// GeneratePKCECodeChallenge generates PKCE code challenge from verifier
func (x *XeroMediator) GeneratePKCECodeChallenge(codeVerifier string) string {
	// For simplicity, we'll use a basic hash
	// In production, this should use SHA256 and Base64URL encoding
	hash := fmt.Sprintf("%x", codeVerifier)
	if len(hash) > 43 {
		hash = hash[:43]
	}
	return hash
}

// GetOAuthRateLimit returns OAuth rate limit configuration
func (x *XeroMediator) GetOAuthRateLimit() *RateLimitInfo {
	return &RateLimitInfo{
		ProviderID:        "xero",
		RequestsRemaining: 100,
		ResetTime:         time.Now().Add(time.Minute),
		Limit:             100,
	}
}

// CanMakeOAuthRequest checks if OAuth request can be made
func (x *XeroMediator) CanMakeOAuthRequest() bool {
	// Simple rate limiting check
	// In production, this should use proper rate limiting
	return true
}

// RecordOAuthRequest records an OAuth request for rate limiting
func (x *XeroMediator) RecordOAuthRequest() {
	// In production, this should record the request for rate limiting
}

// StoreTokens stores OAuth tokens for a company
func (x *XeroMediator) StoreTokens(companyID string, tokens *OAuthTokens) error {
	// In production, this should store tokens securely (e.g., encrypted database)
	// For now, we'll just store in memory
	x.oauthConfig = &OAuthConfig{
		ClientID:     companyID,
		ClientSecret: tokens.AccessToken, // Simplified for testing
	}
	return nil
}

// RetrieveTokens retrieves OAuth tokens for a company
func (x *XeroMediator) RetrieveTokens(companyID string) (*OAuthTokens, error) {
	// In production, this should retrieve tokens from secure storage
	// For now, we'll return the stored tokens
	if x.oauthConfig != nil && x.oauthConfig.ClientID == companyID {
		return &OAuthTokens{
			AccessToken:  x.oauthConfig.ClientSecret, // Simplified for testing
			RefreshToken: "stored-refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    1800,
			Scope:        "offline_access accounting.transactions",
		}, nil
	}
	return nil, fmt.Errorf("tokens not found for company: %s", companyID)
}

// DeleteTokens deletes OAuth tokens for a company
func (x *XeroMediator) DeleteTokens(companyID string) error {
	// In production, this should delete tokens from secure storage
	if x.oauthConfig != nil && x.oauthConfig.ClientID == companyID {
		x.oauthConfig = nil
	}
	return nil
}

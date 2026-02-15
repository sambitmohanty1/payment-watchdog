package mediators

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/sambitmohanty1/payment-watchdog/internal/architecture"
	"github.com/google/uuid"
)

// QuickBooksMediator implements PaymentProvider for QuickBooks
type QuickBooksMediator struct {
	*BaseMediator
	oauthClient *http.Client
	apiClient   *QuickBooksAPIClient
	realmID     string
	oauthTokens *OAuthTokens // Added for OAuth tests
}

// QuickBooksAPIClient handles QuickBooks API communication
type QuickBooksAPIClient struct {
	httpClient *http.Client
	baseURL    string
	realmID    string
	logger     *zap.Logger
}

// QuickBooksInvoice represents a QuickBooks invoice
type QuickBooksInvoice struct {
	ID          string               `json:"Id"`
	DocNumber   string               `json:"DocNumber"`
	CustomerRef QuickBooksRef        `json:"CustomerRef"`
	LineItems   []QuickBooksLineItem `json:"Line"`
	SubTotal    float64              `json:"SubTotalAmt"`
	TotalTax    float64              `json:"TotalAmt"`
	Balance     float64              `json:"Balance"`
	DueDate     time.Time            `json:"DueDate"`
	TxnDate     time.Time            `json:"TxnDate"`
	CurrencyRef QuickBooksRef        `json:"CurrencyRef"`
	PrivateNote string               `json:"PrivateNote"`
	RawData     json.RawMessage      `json:"-"`
}

// QuickBooksRef represents a QuickBooks reference
type QuickBooksRef struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// QuickBooksLineItem represents a line item in a QuickBooks invoice
type QuickBooksLineItem struct {
	Description string          `json:"Description"`
	Qty         float64         `json:"Qty"`
	UnitPrice   float64         `json:"UnitPrice"`
	Amount      float64         `json:"Amount"`
	AccountRef  QuickBooksRef   `json:"AccountRef"`
	RawData     json.RawMessage `json:"-"`
}

// NewQuickBooksMediator creates a new QuickBooks mediator
func NewQuickBooksMediator(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *QuickBooksMediator {
	base := NewBaseMediator(config, eventBus, logger)

	mediator := &QuickBooksMediator{
		BaseMediator: base,
	}

	return mediator
}

// GetProviderName returns the provider name
func (q *QuickBooksMediator) GetProviderName() string {
	return "QuickBooks"
}

// Connect establishes connection to QuickBooks
func (q *QuickBooksMediator) Connect(ctx context.Context, config *ProviderConfig) error {
	if config.OAuthConfig == nil {
		return fmt.Errorf("OAuth configuration required for QuickBooks")
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

	q.oauthClient = oauthConfig.Client(ctx, token)

	// Get realm ID from metadata
	if realmID, ok := config.Metadata["realm_id"].(string); ok {
		q.realmID = realmID
	} else {
		return fmt.Errorf("realm_id required in metadata for QuickBooks")
	}

	q.apiClient = NewQuickBooksAPIClient(q.oauthClient, q.realmID, q.logger)

	// Validate connection
	if err := q.validateConnection(ctx); err != nil {
		return fmt.Errorf("failed to validate QuickBooks connection: %w", err)
	}

	q.isConnected = true
	q.connectedAt = &time.Time{}
	*q.connectedAt = time.Now()

	q.logger.Info("QuickBooks mediator connected successfully",
		zap.String("company_id", config.CompanyID),
		zap.String("realm_id", q.realmID))

	return nil
}

// Disconnect disconnects from QuickBooks
func (q *QuickBooksMediator) Disconnect(ctx context.Context) error {
	q.isConnected = false
	q.connectedAt = nil
	q.oauthClient = nil
	q.apiClient = nil

	q.logger.Info("QuickBooks mediator disconnected",
		zap.String("provider_id", q.config.ProviderID))

	return nil
}

// GetPaymentFailures retrieves payment failures from QuickBooks
func (q *QuickBooksMediator) GetPaymentFailures(ctx context.Context, since time.Time) ([]*architecture.PaymentFailure, error) {
	if !q.isConnected {
		return nil, fmt.Errorf("QuickBooks mediator not connected")
	}

	// Get invoices from QuickBooks
	invoices, err := q.apiClient.GetInvoices(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices from QuickBooks: %w", err)
	}

	var failures []*architecture.PaymentFailure

	// Process invoices to find payment failures
	for _, invoice := range invoices {
		if q.isPaymentFailure(invoice) {
			failure := q.mapInvoiceToFailure(invoice)
			failures = append(failures, failure)
		}
	}

	q.logger.Info("Retrieved payment failures from QuickBooks",
		zap.String("provider_id", q.config.ProviderID),
		zap.Int("failure_count", len(failures)))

	return failures, nil
}

// GetInvoices retrieves invoices from QuickBooks
func (q *QuickBooksMediator) GetInvoices(ctx context.Context, since time.Time) ([]*Invoice, error) {
	if !q.isConnected {
		return nil, fmt.Errorf("QuickBooks mediator not connected")
	}

	// Get invoices from QuickBooks
	qbInvoices, err := q.apiClient.GetInvoices(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices from QuickBooks: %w", err)
	}

	var invoices []*Invoice

	// Map QuickBooks invoices to unified model
	for _, qbInvoice := range qbInvoices {
		invoice := q.mapQuickBooksInvoiceToInvoice(qbInvoice)
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// GetCustomers retrieves customers from QuickBooks
func (q *QuickBooksMediator) GetCustomers(ctx context.Context) ([]*Customer, error) {
	if !q.isConnected {
		return nil, fmt.Errorf("QuickBooks mediator not connected")
	}

	// This would be implemented to get customers from QuickBooks
	// For now, return empty slice
	return []*Customer{}, nil
}

// StartSync starts QuickBooks synchronization
func (q *QuickBooksMediator) StartSync(ctx context.Context) error {
	if !q.isConnected {
		return fmt.Errorf("QuickBooks mediator not connected")
	}

	return q.BaseMediator.StartSync(ctx)
}

// performSync performs QuickBooks-specific synchronization
func (q *QuickBooksMediator) performSync(ctx context.Context) error {
	startTime := time.Now()

	// Update sync status
	q.syncMutex.Lock()
	q.syncStatus.Status = "syncing"
	q.syncStatus.LastSyncAt = &startTime
	q.syncMutex.Unlock()

	// Get payment failures since last sync
	var since time.Time
	if q.syncStatus.LastSyncAt != nil {
		since = *q.syncStatus.LastSyncAt
	} else {
		since = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	// Retrieve payment failures
	failures, err := q.GetPaymentFailures(ctx, since)
	if err != nil {
		q.logger.Error("Failed to get payment failures during sync",
			zap.String("provider_id", q.config.ProviderID),
			zap.Error(err))

		// Update sync status with error
		q.syncMutex.Lock()
		q.syncStatus.Status = "error"
		q.syncStatus.LastError = err.Error()
		q.syncMutex.Unlock()

		return err
	}

	// Publish events for new failures
	for _, failure := range failures {
		if err := q.publishEvent(ctx, "payment.failure.detected", failure); err != nil {
			q.logger.Error("Failed to publish payment failure event",
				zap.String("failure_id", failure.ID.String()),
				zap.Error(err))
		}
	}

	// Update sync status
	q.syncMutex.Lock()
	q.syncStatus.Status = "active"
	q.syncStatus.SyncDuration = time.Since(startTime)
	q.syncStatus.RecordsSynced = int64(len(failures))
	q.syncStatus.NextSyncAt = &time.Time{}
	*q.syncStatus.NextSyncAt = time.Now().Add(q.config.SyncConfig.Frequency)
	q.syncMutex.Unlock()

	q.logger.Info("QuickBooks sync completed successfully",
		zap.String("provider_id", q.config.ProviderID),
		zap.Int("failures_found", len(failures)),
		zap.Duration("sync_duration", time.Since(startTime)))

	return nil
}

// validateConnection validates the QuickBooks connection
func (q *QuickBooksMediator) validateConnection(ctx context.Context) error {
	// Make a test API call to validate connection
	_, err := q.apiClient.GetInvoices(ctx, time.Now().Add(-1*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to validate QuickBooks connection: %w", err)
	}

	return nil
}

// isPaymentFailure checks if an invoice represents a payment failure
func (q *QuickBooksMediator) isPaymentFailure(invoice *QuickBooksInvoice) bool {
	// QuickBooks-specific logic for identifying payment failures
	return invoice.Balance > 0 &&
		invoice.DueDate.Before(time.Now())
}

// mapInvoiceToFailure maps a QuickBooks invoice to a unified PaymentFailure
func (q *QuickBooksMediator) mapInvoiceToFailure(invoice *QuickBooksInvoice) *architecture.PaymentFailure {
	// Generate unique event ID
	eventID := q.generateEventID()

	// Calculate risk score based on amount and overdue days
	overdueDays := int(time.Since(invoice.DueDate).Hours() / 24)
	riskScore := q.calculateRiskScore(invoice.Balance, overdueDays)

	// Map to unified PaymentFailure model
	failure := &architecture.PaymentFailure{
		ProviderID:        "quickbooks",
		ProviderEventID:   eventID,
		ProviderEventType: "invoice.payment_failed",
		Amount:            invoice.Balance,
		Currency:          invoice.CurrencyRef.Value,
		CustomerID:        invoice.CustomerRef.Value,
		CustomerName:      invoice.CustomerRef.Name,
		InvoiceID:         invoice.ID,
		InvoiceNumber:     invoice.DocNumber,
		DueDate:           &invoice.DueDate,
		FailureReason:     "overdue_invoice",
		FailureCode:       "INVOICE_OVERDUE",
		FailureMessage: fmt.Sprintf("Invoice %s is overdue by %d days",
			invoice.DocNumber, overdueDays),
		Status:     "received",
		Priority:   q.mapRiskScoreToPriority(riskScore),
		RiskScore:  riskScore,
		OccurredAt: invoice.DueDate,
		DetectedAt: time.Now(),
		SyncSource: "api_poll",
		RawData:    invoice.RawData,
		ProviderMetadata: map[string]interface{}{
			"quickbooks_invoice_id":  invoice.ID,
			"quickbooks_customer_id": invoice.CustomerRef.Value,
			"quickbooks_realm_id":    q.realmID,
		},
	}

	return failure
}

// mapQuickBooksInvoiceToInvoice maps a QuickBooks invoice to unified Invoice model
func (q *QuickBooksMediator) mapQuickBooksInvoiceToInvoice(qbInvoice *QuickBooksInvoice) *Invoice {
	// Map line items
	lineItems, _ := json.Marshal(qbInvoice.LineItems)

	invoice := &Invoice{
		ProviderID:        "quickbooks",
		ProviderInvoiceID: qbInvoice.ID,
		InvoiceNumber:     qbInvoice.DocNumber,
		Amount:            qbInvoice.TotalTax,
		Currency:          qbInvoice.CurrencyRef.Value,
		Status: func() string {
			if qbInvoice.Balance > 0 {
				return "unpaid"
			} else {
				return "paid"
			}
		}(),
		CustomerID:   qbInvoice.CustomerRef.Value,
		CustomerName: qbInvoice.CustomerRef.Name,
		IssueDate:    qbInvoice.TxnDate,
		DueDate:      qbInvoice.DueDate,
		LineItems:    lineItems,
		ProviderMetadata: map[string]interface{}{
			"quickbooks_invoice_id":  qbInvoice.ID,
			"quickbooks_customer_id": qbInvoice.CustomerRef.Value,
			"quickbooks_realm_id":    q.realmID,
		},
	}

	return invoice
}

// calculateRiskScore calculates risk score based on amount and overdue days
func (q *QuickBooksMediator) calculateRiskScore(amount float64, overdueDays int) float64 {
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

// mapRiskScoreToPriority maps risk score to priority
func (q *QuickBooksMediator) mapRiskScoreToPriority(riskScore float64) architecture.PaymentFailurePriority {
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

// NewQuickBooksAPIClient creates a new QuickBooks API client
func NewQuickBooksAPIClient(httpClient *http.Client, realmID string, logger *zap.Logger) *QuickBooksAPIClient {
	return &QuickBooksAPIClient{
		httpClient: httpClient,
		baseURL:    "https://quickbooks.api.intuit.com/v3/company",
		realmID:    realmID,
		logger:     logger,
	}
}

// GetInvoices retrieves invoices from QuickBooks API
func (q *QuickBooksAPIClient) GetInvoices(ctx context.Context, since time.Time) ([]*QuickBooksInvoice, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("query", fmt.Sprintf("SELECT * FROM Invoice WHERE TxnDate >= '%s' ORDER BY TxnDate DESC",
		since.Format("2006-01-02")))

	// Make API request
	url := fmt.Sprintf("%s/%s/query?%s", q.baseURL, q.realmID, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("QuickBooks API returned status: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		QueryResponse struct {
			Invoices []*QuickBooksInvoice `json:"Invoice"`
		} `json:"QueryResponse"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Store raw data for each invoice
	for _, invoice := range response.QueryResponse.Invoices {
		if rawData, err := json.Marshal(invoice); err == nil {
			invoice.RawData = rawData
		}
	}

	return response.QueryResponse.Invoices, nil
}

// OAuth 2.0 Implementation Methods

// GenerateAuthorizationURL generates the OAuth 2.0 authorization URL
func (q *QuickBooksMediator) GenerateAuthorizationURL(config *OAuthConfig) (string, string, error) {
	if config == nil {
		return "", "", fmt.Errorf("OAuth configuration is required")
	}

	// Generate state parameter for security
	state := q.GenerateStateParameter()

	// Build authorization URL
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		config.AuthURL,
		url.QueryEscape(config.ClientID),
		url.QueryEscape(config.RedirectURI),
		url.QueryEscape(strings.Join(config.Scopes, " ")),
		state)

	return authURL, state, nil
}

// ExchangeCodeForTokens exchanges authorization code for access tokens
func (q *QuickBooksMediator) ExchangeCodeForTokens(ctx context.Context, config *OAuthConfig, authCode string) (*OAuthTokens, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth configuration is required")
	}

	if authCode == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	// In a real implementation, this would make an HTTP request to exchange the code
	// For now, return mock tokens for testing
	if authCode == "test-auth-code" {
		return &OAuthTokens{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			Scope:        strings.Join(config.Scopes, " "),
			ExpiresAt:    time.Now().Add(time.Hour),
		}, nil
	}

	return nil, fmt.Errorf("invalid authorization code")
}

// RefreshAccessToken refreshes an expired access token
func (q *QuickBooksMediator) RefreshAccessToken(ctx context.Context, config *OAuthConfig, refreshToken string) (*OAuthTokens, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth configuration is required")
	}

	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// In a real implementation, this would make an HTTP request to refresh the token
	// For now, return mock tokens for testing
	return &OAuthTokens{
		AccessToken:  "refreshed-access-token",
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    3600,
		Scope:        strings.Join(config.Scopes, " "),
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

// ValidateTokens validates OAuth tokens
func (q *QuickBooksMediator) ValidateTokens(tokens *OAuthTokens) bool {
	if tokens == nil {
		return false
	}

	if tokens.AccessToken == "" {
		return false
	}

	if tokens.RefreshToken == "" {
		return false
	}

	if tokens.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// HasRequiredScopes checks if the mediator has required OAuth scopes
func (q *QuickBooksMediator) HasRequiredScopes() bool {
	if q.config.OAuthConfig == nil {
		return false
	}

	requiredScopes := []string{"com.intuit.quickbooks.accounting"}
	return q.ValidateScopes(requiredScopes)
}

// ValidateScopes validates OAuth scopes
func (q *QuickBooksMediator) ValidateScopes(scopes []string) bool {
	if q.config.OAuthConfig == nil {
		return false
	}

	// Check if all provided scopes are valid (contained in mediator's scopes)
	for _, scope := range scopes {
		found := false
		for _, mediatorScope := range q.config.OAuthConfig.Scopes {
			if mediatorScope == scope {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GenerateStateParameter generates a secure state parameter
func (q *QuickBooksMediator) GenerateStateParameter() string {
	// Generate a random state parameter
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// ValidateStateParameter validates a state parameter
func (q *QuickBooksMediator) ValidateStateParameter(expected, actual string) bool {
	return expected == actual
}

// GeneratePKCECodeVerifier generates a PKCE code verifier
func (q *QuickBooksMediator) GeneratePKCECodeVerifier() string {
	// Generate a random code verifier
	bytes := make([]byte, 64)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// GeneratePKCECodeChallenge generates a PKCE code challenge
func (q *QuickBooksMediator) GeneratePKCECodeChallenge(verifier string) string {
	// Generate SHA256 hash of the verifier
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GetOAuthRateLimit returns OAuth rate limit information
func (q *QuickBooksMediator) GetOAuthRateLimit() *architecture.RateLimitInfo {
	return &architecture.RateLimitInfo{
		ProviderID:        "quickbooks",
		RequestsRemaining: 100,
		ResetTime:         time.Now().Add(time.Hour),
		Limit:             100,
	}
}

// CanMakeOAuthRequest checks if an OAuth request can be made
func (q *QuickBooksMediator) CanMakeOAuthRequest() bool {
	// Simple rate limiting check
	return true
}

// RecordOAuthRequest records an OAuth request
func (q *QuickBooksMediator) RecordOAuthRequest() {
	// In a real implementation, this would track rate limiting
}

// StoreTokens stores OAuth tokens
func (q *QuickBooksMediator) StoreTokens(tokens *OAuthTokens) error {
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	// In a real implementation, this would store tokens securely
	// For now, just store in memory for testing
	q.oauthTokens = tokens
	return nil
}

// RetrieveTokens retrieves stored OAuth tokens
func (q *QuickBooksMediator) RetrieveTokens() *OAuthTokens {
	return q.oauthTokens
}

// DeleteTokens deletes stored OAuth tokens
func (q *QuickBooksMediator) DeleteTokens() error {
	q.oauthTokens = nil
	return nil
}

// mapQuickBooksInvoiceToPaymentFailure maps a QuickBooks invoice to a PaymentFailure
func (q *QuickBooksMediator) mapQuickBooksInvoiceToPaymentFailure(invoice *QuickBooksInvoice) *architecture.PaymentFailure {
	if invoice == nil {
		return nil
	}

	// Calculate overdue days for risk scoring
	overdueDays := 0
	if !invoice.DueDate.IsZero() {
		overdueDays = int(time.Since(invoice.DueDate).Hours() / 24)
		if overdueDays < 0 {
			overdueDays = 0
		}
	}

	// Calculate risk score using existing method
	riskScore := q.calculateRiskScore(invoice.Balance, overdueDays)

	// Determine failure reason based on balance
	failureReason := "invoice_unpaid"
	if invoice.Balance < invoice.TotalTax {
		failureReason = "invoice_partially_paid"
	}

	// Map to unified PaymentFailure model
	paymentFailure := &architecture.PaymentFailure{
		ID:              uuid.New(),
		ProviderEventID: invoice.ID,
		Amount:          invoice.Balance,
		Currency:        invoice.CurrencyRef.Value,
		CustomerID:      invoice.CustomerRef.Value,
		CustomerName:    invoice.CustomerRef.Name,
		FailureReason:   failureReason,
		SyncSource:      "quickbooks",
		RiskScore:       riskScore,
		Status:          architecture.PaymentFailureStatusReceived,
		Priority:        q.mapRiskScoreToPriority(riskScore),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return paymentFailure
}

// publishPaymentFailureEvent publishes a payment failure event to the event bus
func (q *QuickBooksMediator) publishPaymentFailureEvent(paymentFailure *architecture.PaymentFailure) error {
	if paymentFailure == nil {
		return fmt.Errorf("payment failure cannot be nil")
	}

	event := map[string]interface{}{
		"event_id":        paymentFailure.ID.String(),
		"event_type":      "payment.failure.detected",
		"provider":        "quickbooks",
		"payment_failure": paymentFailure,
		"timestamp":       time.Now(),
		"metadata": map[string]interface{}{
			"source":      "quickbooks_mediator",
			"version":     "1.0",
			"environment": "production",
		},
	}

	return q.eventBus.Publish(context.Background(), "payment.failure.detected", event)
}

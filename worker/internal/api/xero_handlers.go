package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/lexure-intelligence/payment-watchdog/internal/mediators"
)

// XeroHandlers handles Xero-specific API endpoints using the mediator pattern
type XeroHandlers struct {
	xeroMediator *mediators.XeroMediator
	logger       *zap.Logger
}

// NewXeroHandlers creates a new Xero handlers instance
func NewXeroHandlers(xeroMediator *mediators.XeroMediator, logger *zap.Logger) *XeroHandlers {
	return &XeroHandlers{
		xeroMediator: xeroMediator,
		logger:       logger,
	}
}

// XeroOAuthRequest represents the OAuth authorization request
type XeroOAuthRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
}

// XeroOAuthResponse represents the OAuth authorization response
type XeroOAuthResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

// XeroCallbackRequest represents the OAuth callback request
type XeroCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// XeroCallbackResponse represents the OAuth callback response
type XeroCallbackResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
}

// XeroTenantsResponse represents the response for getting Xero tenants
type XeroTenantsResponse struct {
	Success bool         `json:"success"`
	Tenants []XeroTenant `json:"tenants"`
	Error   string       `json:"error,omitempty"`
}

// XeroTenant represents a Xero tenant/organization
type XeroTenant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortCode   string `json:"short_code"`
	IsActive    bool   `json:"is_active"`
	CreatedDate string `json:"created_date"`
}

// XeroOrganizationsResponse represents the response for getting Xero organizations
type XeroOrganizationsResponse struct {
	Success       bool               `json:"success"`
	Organizations []XeroOrganization `json:"organizations"`
	Error         string             `json:"error,omitempty"`
}

// XeroOrganization represents a Xero organization
type XeroOrganization struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	LegalName    string `json:"legal_name"`
	ShortCode    string `json:"short_code"`
	CountryCode  string `json:"country_code"`
	BaseCurrency string `json:"base_currency"`
	IsActive     bool   `json:"is_active"`
}

// XeroPaymentFailuresResponse represents the response for getting payment failures
type XeroPaymentFailuresResponse struct {
	Success         bool                 `json:"success"`
	PaymentFailures []XeroPaymentFailure `json:"payment_failures"`
	TotalCount      int                  `json:"total_count"`
	Error           string               `json:"error,omitempty"`
}

// XeroPaymentFailure represents a payment failure from Xero
type XeroPaymentFailure struct {
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

// GetAuthorizationURL initiates Xero OAuth flow
func (h *XeroHandlers) GetAuthorizationURL(c *gin.Context) {
	var req XeroOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Generate state for security
	state := fmt.Sprintf("xero_%s_%d", req.CompanyID, time.Now().Unix())

	// Get authorization URL from mediator
	authURL, err := h.xeroMediator.GetAuthorizationURL(state)
	if err != nil {
		h.logger.Error("Failed to get authorization URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get authorization URL",
			"details": err.Error(),
		})
		return
	}

	response := XeroOAuthResponse{
		AuthorizationURL: authURL,
		State:            state,
	}

	c.JSON(http.StatusOK, response)
}

// HandleCallback handles Xero OAuth callback
func (h *XeroHandlers) HandleCallback(c *gin.Context) {
	// Get parameters from URL query string (OAuth callback comes as GET request)
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.logger.Error("Authorization code not received")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code not received",
		})
		return
	}

	if state == "" {
		h.logger.Error("State parameter not received")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State parameter not received",
		})
		return
	}

	// Exchange authorization code for tokens using mediator
	tokens, err := h.xeroMediator.ExchangeCodeForTokens(context.Background(), code, state)
	if err != nil {
		h.logger.Error("Failed to exchange code for tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to exchange authorization code",
			"details": err.Error(),
		})
		return
	}

	response := XeroCallbackResponse{
		Success:      true,
		Message:      "Xero OAuth completed successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GetTenants retrieves Xero tenant connections
func (h *XeroHandlers) GetTenants(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	// Get tenants using mediator
	tenants, err := h.xeroMediator.GetTenants(context.Background(), accessToken)
	if err != nil {
		h.logger.Error("Failed to get Xero tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch Xero data",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	xeroTenants := make([]XeroTenant, len(tenants))
	for i, tenant := range tenants {
		xeroTenants[i] = XeroTenant{
			ID:          tenant.ID,
			Name:        tenant.Name,
			ShortCode:   tenant.ShortCode,
			IsActive:    tenant.IsActive,
			CreatedDate: tenant.CreatedDate,
		}
	}

	response := XeroTenantsResponse{
		Success: true,
		Tenants: xeroTenants,
	}

	c.JSON(http.StatusOK, response)
}

// GetOrganizations retrieves Xero organization details
func (h *XeroHandlers) GetOrganizations(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant_id query parameter required",
		})
		return
	}

	// Get organizations using mediator
	organizations, err := h.xeroMediator.GetOrganizations(context.Background(), accessToken, tenantID)
	if err != nil {
		h.logger.Error("Failed to get Xero organizations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch Xero data",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	xeroOrganizations := make([]XeroOrganization, len(organizations))
	for i, org := range organizations {
		xeroOrganizations[i] = XeroOrganization{
			ID:           org.ID,
			Name:         org.Name,
			LegalName:    org.LegalName,
			ShortCode:    org.ShortCode,
			CountryCode:  org.CountryCode,
			BaseCurrency: org.BaseCurrency,
			IsActive:     org.IsActive,
		}
	}

	response := XeroOrganizationsResponse{
		Success:       true,
		Organizations: xeroOrganizations,
	}

	c.JSON(http.StatusOK, response)
}

// GetPaymentFailures retrieves payment failures from Xero
func (h *XeroHandlers) GetPaymentFailures(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant_id query parameter required",
		})
		return
	}

	// Get payment failures using mediator
	failures, err := h.xeroMediator.GetPaymentFailures(context.Background(), accessToken, tenantID)
	if err != nil {
		h.logger.Error("Failed to get Xero payment failures", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch payment failures",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	xeroFailures := make([]XeroPaymentFailure, len(failures))
	for i, failure := range failures {
		xeroFailures[i] = XeroPaymentFailure{
			ID:            failure.ID,
			InvoiceNumber: failure.InvoiceNumber,
			CustomerName:  failure.CustomerName,
			CustomerEmail: failure.CustomerEmail,
			Amount:        failure.Amount,
			Currency:      failure.Currency,
			DueDate:       failure.DueDate,
			DaysOverdue:   failure.DaysOverdue,
			FailureReason: failure.FailureReason,
			Status:        failure.Status,
		}
	}

	response := XeroPaymentFailuresResponse{
		Success:         true,
		PaymentFailures: xeroFailures,
		TotalCount:      len(xeroFailures),
	}

	c.JSON(http.StatusOK, response)
}

// RegisterXeroRoutes registers Xero-specific routes
func RegisterXeroRoutes(router *gin.RouterGroup, xeroHandlers *XeroHandlers) {
	xero := router.Group("/xero")
	{
		// Test endpoint to debug route registration
		xero.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Xero routes are working!"})
		})

		// OAuth endpoints
		xero.POST("/auth/authorize", xeroHandlers.GetAuthorizationURL)
		xero.GET("/auth/callback", xeroHandlers.HandleCallback)

		// Data endpoints
		xero.GET("/tenants", xeroHandlers.GetTenants)
		xero.GET("/organizations", xeroHandlers.GetOrganizations)
		xero.GET("/payment-failures", xeroHandlers.GetPaymentFailures)
	}
}

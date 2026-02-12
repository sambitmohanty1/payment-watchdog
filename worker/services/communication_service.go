package services

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/lexure-intelligence/payment-watchdog/internal/models"
)

// CommunicationService handles customer communications
type CommunicationService struct {
	db           *gorm.DB
	emailService EmailProvider
	smsService   SMSProvider
	tracer       trace.Tracer
}

// CommunicationRequest represents a communication request
type CommunicationRequest struct {
	CompanyID    uuid.UUID              `json:"company_id"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateName string                 `json:"template_name,omitempty"`
	Recipient    string                 `json:"recipient"`
	Subject      string                 `json:"subject,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// CommunicationResult represents the result of a communication
type CommunicationResult struct {
	MessageID    string    `json:"message_id"`
	Status       string    `json:"status"`
	TemplateUsed string    `json:"template_used,omitempty"`
	SentAt       time.Time `json:"sent_at"`
	Provider     string    `json:"provider"`
}

// EmailProvider interface for email services
type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, body string, metadata map[string]interface{}) (*EmailResult, error)
}

// SMSProvider interface for SMS services
type SMSProvider interface {
	SendSMS(ctx context.Context, to, message string, metadata map[string]interface{}) (*SMSResult, error)
}

// EmailResult represents email sending result
type EmailResult struct {
	MessageID string            `json:"message_id"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SMSResult represents SMS sending result
type SMSResult struct {
	MessageID string            `json:"message_id"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// MockEmailProvider for development/testing
type MockEmailProvider struct{}

func (m *MockEmailProvider) SendEmail(ctx context.Context, to, subject, body string, metadata map[string]interface{}) (*EmailResult, error) {
	return &EmailResult{
		MessageID: fmt.Sprintf("email_%d", time.Now().Unix()),
		Status:    "sent",
		Metadata: map[string]string{
			"provider": "mock",
			"to":       to,
		},
	}, nil
}

// MockSMSProvider for development/testing
type MockSMSProvider struct{}

func (m *MockSMSProvider) SendSMS(ctx context.Context, to, message string, metadata map[string]interface{}) (*SMSResult, error) {
	return &SMSResult{
		MessageID: fmt.Sprintf("sms_%d", time.Now().Unix()),
		Status:    "sent",
		Metadata: map[string]string{
			"provider": "mock",
			"to":       to,
		},
	}, nil
}

// NewCommunicationService creates a new communication service
func NewCommunicationService(db *gorm.DB, emailProvider EmailProvider, smsProvider SMSProvider) *CommunicationService {
	// Use mock providers if none provided
	if emailProvider == nil {
		emailProvider = &MockEmailProvider{}
	}
	if smsProvider == nil {
		smsProvider = &MockSMSProvider{}
	}

	return &CommunicationService{
		db:           db,
		emailService: emailProvider,
		smsService:   smsProvider,
		tracer:       otel.Tracer("communication-service"),
	}
}

// SendEmail sends an email using templates
func (c *CommunicationService) SendEmail(ctx context.Context, req *CommunicationRequest) (*CommunicationResult, error) {
	ctx, span := c.tracer.Start(ctx, "send_email")
	defer span.End()

	span.SetAttributes(
		attribute.String("company_id", req.CompanyID.String()),
		attribute.String("template_id", req.TemplateID),
		attribute.String("template_name", req.TemplateName),
		attribute.String("recipient", req.Recipient),
	)

	// Get or create template
	template, err := c.getEmailTemplate(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get email template: %w", err)
	}

	// Render template with variables
	subject, body, err := c.renderEmailTemplate(template, req.Variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email through provider
	emailResult, err := c.emailService.SendEmail(ctx, req.Recipient, subject, body, map[string]interface{}{
		"company_id":  req.CompanyID.String(),
		"template_id": template.ID.String(),
		"context":     req.Context,
	})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	// Update template usage
	c.updateTemplateUsage(ctx, template.ID)

	// Create customer communication record
	communication := &models.CustomerCommunication{
		ID:               uuid.New(),
		CompanyID:        req.CompanyID.String(),
		CommunicationType: "email",
		Recipient:        req.Recipient,
		Subject:          subject,
		Content:          body,
		Status:           "sent",
		SentAt:           &time.Time{},
		TemplateID:       template.ID.String(),
		ExternalID:       emailResult.MessageID,
		Metadata:         req.Context,
	}
	*communication.SentAt = time.Now()

	if err := c.db.WithContext(ctx).Create(communication).Error; err != nil {
		span.RecordError(err)
		// Log error but don't fail the operation
	}

	return &CommunicationResult{
		MessageID:    emailResult.MessageID,
		Status:       emailResult.Status,
		TemplateUsed: template.Name,
		SentAt:       time.Now(),
		Provider:     "email",
	}, nil
}

// SendSMS sends an SMS using templates
func (c *CommunicationService) SendSMS(ctx context.Context, req *CommunicationRequest) (*CommunicationResult, error) {
	ctx, span := c.tracer.Start(ctx, "send_sms")
	defer span.End()

	span.SetAttributes(
		attribute.String("company_id", req.CompanyID.String()),
		attribute.String("template_id", req.TemplateID),
		attribute.String("template_name", req.TemplateName),
		attribute.String("recipient", req.Recipient),
	)

	// Get or create template
	template, err := c.getSMSTemplate(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get SMS template: %w", err)
	}

	// Render template with variables
	message, err := c.renderSMSTemplate(template, req.Variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to render SMS template: %w", err)
	}

	// Send SMS through provider
	smsResult, err := c.smsService.SendSMS(ctx, req.Recipient, message, map[string]interface{}{
		"company_id":  req.CompanyID.String(),
		"template_id": template.ID.String(),
		"context":     req.Context,
	})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Update template usage
	c.updateTemplateUsage(ctx, template.ID)

	// Create customer communication record
	communication := &models.CustomerCommunication{
		ID:               uuid.New(),
		CompanyID:        req.CompanyID.String(),
		CommunicationType: "sms",
		Recipient:        req.Recipient,
		Content:          message,
		Status:           "sent",
		SentAt:           &time.Time{},
		TemplateID:       template.ID.String(),
		ExternalID:       smsResult.MessageID,
		Metadata:         req.Context,
	}
	*communication.SentAt = time.Now()

	if err := c.db.WithContext(ctx).Create(communication).Error; err != nil {
		span.RecordError(err)
		// Log error but don't fail the operation
	}

	return &CommunicationResult{
		MessageID:    smsResult.MessageID,
		Status:       smsResult.Status,
		TemplateUsed: template.Name,
		SentAt:       time.Now(),
		Provider:     "sms",
	}, nil
}

// getEmailTemplate retrieves or creates an email template
func (c *CommunicationService) getEmailTemplate(ctx context.Context, req *CommunicationRequest) (*models.CommunicationTemplate, error) {
	var template models.CommunicationTemplate

	// Try to find by ID first
	if req.TemplateID != "" {
		if id, err := uuid.Parse(req.TemplateID); err == nil {
			if err := c.db.WithContext(ctx).
				Where("id = ? AND company_id = ? AND template_type = ? AND is_active = ?", 
					id, req.CompanyID, "email", true).
				First(&template).Error; err == nil {
				return &template, nil
			}
		}
	}

	// Try to find by name
	if req.TemplateName != "" {
		if err := c.db.WithContext(ctx).
			Where("name = ? AND company_id = ? AND template_type = ? AND is_active = ?", 
				req.TemplateName, req.CompanyID, "email", true).
			First(&template).Error; err == nil {
			return &template, nil
		}
	}

	// Try to find default template
	if err := c.db.WithContext(ctx).
		Where("company_id = ? AND template_type = ? AND is_default = ? AND is_active = ?", 
			req.CompanyID, "email", true, true).
		First(&template).Error; err == nil {
		return &template, nil
	}

	// Create a basic template if none found
	template = models.CommunicationTemplate{
		ID:           uuid.New(),
		CompanyID:    req.CompanyID,
		Name:         "Default Payment Failure Email",
		Description:  "Auto-generated default email template for payment failures",
		TemplateType: "email",
		Subject:      "Payment Issue - Action Required",
		Content: `Dear {{.customer_name}},

We were unable to process your payment of {{.currency}} {{.amount}} due to: {{.failure_reason}}.

Please update your payment method or contact us to resolve this issue.

Transaction ID: {{.transaction_id}}

Thank you for your attention to this matter.`,
		IsActive:  true,
		IsDefault: true,
		CreatedBy: "system",
	}

	if err := c.db.WithContext(ctx).Create(&template).Error; err != nil {
		return nil, fmt.Errorf("failed to create default email template: %w", err)
	}

	return &template, nil
}

// getSMSTemplate retrieves or creates an SMS template
func (c *CommunicationService) getSMSTemplate(ctx context.Context, req *CommunicationRequest) (*models.CommunicationTemplate, error) {
	var template models.CommunicationTemplate

	// Try to find by ID first
	if req.TemplateID != "" {
		if id, err := uuid.Parse(req.TemplateID); err == nil {
			if err := c.db.WithContext(ctx).
				Where("id = ? AND company_id = ? AND template_type = ? AND is_active = ?", 
					id, req.CompanyID, "sms", true).
				First(&template).Error; err == nil {
				return &template, nil
			}
		}
	}

	// Try to find by name
	if req.TemplateName != "" {
		if err := c.db.WithContext(ctx).
			Where("name = ? AND company_id = ? AND template_type = ? AND is_active = ?", 
				req.TemplateName, req.CompanyID, "sms", true).
			First(&template).Error; err == nil {
			return &template, nil
		}
	}

	// Try to find default template
	if err := c.db.WithContext(ctx).
		Where("company_id = ? AND template_type = ? AND is_default = ? AND is_active = ?", 
			req.CompanyID, "sms", true, true).
		First(&template).Error; err == nil {
		return &template, nil
	}

	// Create a basic template if none found
	template = models.CommunicationTemplate{
		ID:           uuid.New(),
		CompanyID:    req.CompanyID,
		Name:         "Default Payment Failure SMS",
		Description:  "Auto-generated default SMS template for payment failures",
		TemplateType: "sms",
		Content:      "Payment of {{.currency}} {{.amount}} failed. Please update your payment method. Ref: {{.transaction_id}}",
		IsActive:     true,
		IsDefault:    true,
		CreatedBy:    "system",
	}

	if err := c.db.WithContext(ctx).Create(&template).Error; err != nil {
		return nil, fmt.Errorf("failed to create default SMS template: %w", err)
	}

	return &template, nil
}

// renderEmailTemplate renders an email template with variables
func (c *CommunicationService) renderEmailTemplate(tmpl *models.CommunicationTemplate, variables map[string]interface{}) (string, string, error) {
	// Render subject
	subject := tmpl.Subject
	if subject == "" {
		subject = "Payment Issue - Action Required"
	}

	subjectTmpl, err := template.New("subject").Parse(subject)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjectBuf strings.Builder
	if err := subjectTmpl.Execute(&subjectBuf, variables); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}

	// Render body
	bodyTmpl, err := template.New("body").Parse(tmpl.Content)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}

	var bodyBuf strings.Builder
	if err := bodyTmpl.Execute(&bodyBuf, variables); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// renderSMSTemplate renders an SMS template with variables
func (c *CommunicationService) renderSMSTemplate(tmpl *models.CommunicationTemplate, variables map[string]interface{}) (string, error) {
	messageTmpl, err := template.New("sms").Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse SMS template: %w", err)
	}

	var messageBuf strings.Builder
	if err := messageTmpl.Execute(&messageBuf, variables); err != nil {
		return "", fmt.Errorf("failed to execute SMS template: %w", err)
	}

	return messageBuf.String(), nil
}

// updateTemplateUsage updates template usage statistics
func (c *CommunicationService) updateTemplateUsage(ctx context.Context, templateID uuid.UUID) {
	now := time.Now()
	c.db.WithContext(ctx).Model(&models.CommunicationTemplate{}).
		Where("id = ?", templateID).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + ?", 1),
			"last_used_at": now,
		})
}

// Template Management Methods

// CreateTemplate creates a new communication template
func (c *CommunicationService) CreateTemplate(ctx context.Context, template *models.CommunicationTemplate) error {
	ctx, span := c.tracer.Start(ctx, "create_template")
	defer span.End()

	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	if err := c.db.WithContext(ctx).Create(template).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetTemplates retrieves communication templates for a company
func (c *CommunicationService) GetTemplates(ctx context.Context, companyID uuid.UUID, templateType string, page, limit int) ([]models.CommunicationTemplate, int64, error) {
	var templates []models.CommunicationTemplate
	var total int64

	query := c.db.WithContext(ctx).Model(&models.CommunicationTemplate{}).
		Where("company_id = ? AND is_active = ?", companyID, true)

	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// UpdateTemplate updates an existing communication template
func (c *CommunicationService) UpdateTemplate(ctx context.Context, templateID uuid.UUID, updates map[string]interface{}) error {
	ctx, span := c.tracer.Start(ctx, "update_template")
	defer span.End()

	updates["updated_at"] = time.Now()

	if err := c.db.WithContext(ctx).Model(&models.CommunicationTemplate{}).
		Where("id = ?", templateID).
		Updates(updates).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

// DeleteTemplate soft deletes a communication template
func (c *CommunicationService) DeleteTemplate(ctx context.Context, templateID uuid.UUID) error {
	ctx, span := c.tracer.Start(ctx, "delete_template")
	defer span.End()

	if err := c.db.WithContext(ctx).Model(&models.CommunicationTemplate{}).
		Where("id = ?", templateID).
		Update("is_active", false).Error; err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// GetCommunicationHistory retrieves communication history for a company
func (c *CommunicationService) GetCommunicationHistory(ctx context.Context, companyID uuid.UUID, filters map[string]interface{}, page, limit int) ([]models.CustomerCommunication, int64, error) {
	var communications []models.CustomerCommunication
	var total int64

	query := c.db.WithContext(ctx).Model(&models.CustomerCommunication{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if commType, ok := filters["type"].(string); ok && commType != "" {
		query = query.Where("communication_type = ?", commType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if recipient, ok := filters["recipient"].(string); ok && recipient != "" {
		query = query.Where("recipient ILIKE ?", "%"+recipient+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&communications).Error; err != nil {
		return nil, 0, err
	}

	return communications, total, nil
}

package mediators

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// XeroMediatorSimple is a simplified version for testing
type XeroMediatorSimple struct {
	*BaseMediator
	oauthClient *http.Client
}

// NewXeroMediatorSimple creates a new simple Xero mediator
func NewXeroMediatorSimple(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *XeroMediatorSimple {
	base := NewBaseMediator(config, eventBus, logger)

	return &XeroMediatorSimple{
		BaseMediator: base,
	}
}

// GetProviderName returns the provider name
func (x *XeroMediatorSimple) GetProviderName() string {
	return "Xero"
}

// Connect establishes connection to Xero
func (x *XeroMediatorSimple) Connect(ctx context.Context, config *ProviderConfig) error {
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

	x.logger.Info("Xero mediator connected successfully",
		zap.String("company_id", config.CompanyID))

	return nil
}

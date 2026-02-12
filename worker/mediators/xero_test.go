package mediators

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// XeroTest is a test implementation
type XeroTest struct {
	*BaseMediator
	oauthClient *http.Client
}

// NewXeroTest creates a new test mediator
func NewXeroTest(config *ProviderConfig, eventBus EventBus, logger *zap.Logger) *XeroTest {
	base := NewBaseMediator(config, eventBus, logger)

	return &XeroTest{
		BaseMediator: base,
	}
}

// TestOAuth2 tests oauth2 functionality
func (x *XeroTest) TestOAuth2() error {
	config := &oauth2.Config{
		ClientID:     "test",
		ClientSecret: "test",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"test"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:8080/auth",
			TokenURL: "http://localhost:8080/token",
		},
	}

	token := &oauth2.Token{
		AccessToken: "test",
	}

	x.oauthClient = config.Client(context.Background(), token)

	return nil
}

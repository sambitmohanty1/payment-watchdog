package mediators

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestQuickBooksOAuthFlow tests the complete OAuth 2.0 flow
func TestQuickBooksOAuthFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create mock OAuth server
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/v1/authorize":
			// Simulate QuickBooks authorization endpoint
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Return authorization code
			response := map[string]interface{}{
				"code":  "test-auth-code",
				"state": r.URL.Query().Get("state"),
				"realmId": "test-realm-id",
			}
			json.NewEncoder(w).Encode(response)

		case "/oauth2/v1/tokens/bearer":
			// Simulate token exchange endpoint
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Return access tokens
			response := map[string]interface{}{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"token_type":    "bearer",
				"expires_in":    3600,
				"scope":         "com.intuit.quickbooks.accounting",
			}
			json.NewEncoder(w).Encode(response)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer oauthServer.Close()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURI:  "http://localhost:8080/callback",
			Scopes:       []string{"com.intuit.quickbooks.accounting"},
			AuthURL:      oauthServer.URL + "/oauth2/v1/authorize",
			TokenURL:     oauthServer.URL + "/oauth2/v1/tokens/bearer",
		},
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Complete OAuth Flow", func(t *testing.T) {
		// Test authorization URL generation
		authURL, state, err := mediator.GenerateAuthorizationURL(config.OAuthConfig)
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotEmpty(t, state)
		assert.Contains(t, authURL, "client_id=test-client-id")
		assert.Contains(t, authURL, "scope=com.intuit.quickbooks.accounting")
		assert.Contains(t, authURL, "state="+state)

		// Test token exchange
		tokens, err := mediator.ExchangeCodeForTokens(context.Background(), config.OAuthConfig, "test-auth-code")
		require.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.Equal(t, "test-access-token", tokens.AccessToken)
		assert.Equal(t, "test-refresh-token", tokens.RefreshToken)
		assert.Equal(t, "bearer", tokens.TokenType)
		assert.Equal(t, int64(3600), tokens.ExpiresIn)
		assert.Equal(t, "com.intuit.quickbooks.accounting", tokens.Scope)
	})
}

// TestQuickBooksOAuthErrorHandling tests OAuth error scenarios
func TestQuickBooksOAuthErrorHandling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Invalid Authorization Code", func(t *testing.T) {
		// Test with invalid authorization code
		oauthConfig := &OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURI:  "http://localhost:8080/callback",
			Scopes:       []string{"com.intuit.quickbooks.accounting"},
		}
		_, err := mediator.ExchangeCodeForTokens(context.Background(), oauthConfig, "invalid-code")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization code")
	})

	t.Run("Expired Authorization Code", func(t *testing.T) {
		// Test with expired authorization code
		oauthConfig := &OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURI:  "http://localhost:8080/callback",
			Scopes:       []string{"com.intuit.quickbooks.accounting"},
		}
		_, err := mediator.ExchangeCodeForTokens(context.Background(), oauthConfig, "expired-code")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization code")
	})
}

// TestQuickBooksOAuthTokenValidation tests token validation
func TestQuickBooksOAuthTokenValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Valid Token", func(t *testing.T) {
		// Test with valid tokens
		tokens := &OAuthTokens{
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
		}
		
		isValid := mediator.ValidateTokens(tokens)
		assert.True(t, isValid)
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Test with expired tokens
		tokens := &OAuthTokens{
			AccessToken:  "expired-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(-time.Hour),
		}
		
		isValid := mediator.ValidateTokens(tokens)
		assert.False(t, isValid)
	})

	t.Run("Missing Access Token", func(t *testing.T) {
		// Test with missing access token
		tokens := &OAuthTokens{
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
		}
		
		isValid := mediator.ValidateTokens(tokens)
		assert.False(t, isValid)
	})

	t.Run("Missing Refresh Token", func(t *testing.T) {
		// Test with missing refresh token
		tokens := &OAuthTokens{
			AccessToken: "valid-access-token",
			ExpiresAt:   time.Now().Add(time.Hour),
		}
		
		isValid := mediator.ValidateTokens(tokens)
		assert.False(t, isValid)
	})
}

// TestQuickBooksOAuthScopes tests OAuth scope handling
func TestQuickBooksOAuthScopes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
		OAuthConfig: &OAuthConfig{
			Scopes: []string{"com.intuit.quickbooks.accounting"},
		},
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Required Scopes", func(t *testing.T) {
		// Test that required scopes are present
		hasRequired := mediator.HasRequiredScopes()
		assert.True(t, hasRequired)
	})

	t.Run("Missing Required Scopes", func(t *testing.T) {
		// Test with missing required scopes
		config.OAuthConfig.Scopes = []string{"com.intuit.quickbooks.payment"}
		mediator := NewQuickBooksMediator(config, eventBus, logger)
		
		hasRequired := mediator.HasRequiredScopes()
		assert.False(t, hasRequired)
	})

	t.Run("Scope Validation", func(t *testing.T) {
		// Test scope validation with proper config
		freshConfig := &ProviderConfig{
			ProviderID:   "quickbooks-test",
			ProviderType: ProviderTypeOAuth,
			OAuthConfig: &OAuthConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURI:  "http://localhost:8080/callback",
				Scopes:       []string{"com.intuit.quickbooks.accounting"},
			},
		}
		mediatorWithConfig := NewQuickBooksMediator(freshConfig, eventBus, logger)
		
		validScopes := []string{"com.intuit.quickbooks.accounting"}
		isValid := mediatorWithConfig.ValidateScopes(validScopes)
		assert.True(t, isValid)

		invalidScopes := []string{"invalid.scope"}
		isValid = mediatorWithConfig.ValidateScopes(invalidScopes)
		assert.False(t, isValid)
	})
}

// TestQuickBooksOAuthSecurity tests OAuth security features
func TestQuickBooksOAuthSecurity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("State Parameter Generation", func(t *testing.T) {
		// Test state parameter generation
		state1 := mediator.GenerateStateParameter()
		state2 := mediator.GenerateStateParameter()
		
		assert.NotEmpty(t, state1)
		assert.NotEmpty(t, state2)
		assert.NotEqual(t, state1, state2)
	})

	t.Run("State Parameter Validation", func(t *testing.T) {
		// Test state parameter validation
		state := mediator.GenerateStateParameter()
		isValid := mediator.ValidateStateParameter(state, state)
		assert.True(t, isValid)

		isValid = mediator.ValidateStateParameter(state, "different-state")
		assert.False(t, isValid)
	})

	t.Run("PKCE Code Verifier Generation", func(t *testing.T) {
		// Test PKCE code verifier generation
		verifier := mediator.GeneratePKCECodeVerifier()
		assert.NotEmpty(t, verifier)
		assert.Len(t, verifier, 128) // SHA256 hash length
	})

	t.Run("PKCE Code Challenge Generation", func(t *testing.T) {
		// Test PKCE code challenge generation
		verifier := mediator.GeneratePKCECodeVerifier()
		challenge := mediator.GeneratePKCECodeChallenge(verifier)
		
		assert.NotEmpty(t, challenge)
		assert.NotEqual(t, verifier, challenge)
	})
}

// TestQuickBooksOAuthRateLimiting tests OAuth rate limiting
func TestQuickBooksOAuthRateLimiting(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Rate Limit Configuration", func(t *testing.T) {
		// Test rate limit configuration
		rateLimit := mediator.GetOAuthRateLimit()
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "quickbooks", rateLimit.ProviderID)
		assert.True(t, rateLimit.Limit > 0)
	})

	t.Run("Rate Limit Enforcement", func(t *testing.T) {
		// Test rate limit enforcement
		canMakeRequest := mediator.CanMakeOAuthRequest()
		assert.True(t, canMakeRequest)

		// Record a request
		mediator.RecordOAuthRequest()
		
		// Should still be able to make requests within limits
		canMakeRequest = mediator.CanMakeOAuthRequest()
		assert.True(t, canMakeRequest)
	})
}

// TestQuickBooksOAuthTokenStorage tests OAuth token storage
func TestQuickBooksOAuthTokenStorage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus := &TestEventBus{}
	config := &ProviderConfig{
		ProviderID:   "quickbooks-test",
		ProviderType: ProviderTypeOAuth,
	}

	mediator := NewQuickBooksMediator(config, eventBus, logger)

	t.Run("Token Storage", func(t *testing.T) {
		// Test token storage
		tokens := &OAuthTokens{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
		}
		
		err := mediator.StoreTokens(tokens)
		require.NoError(t, err)

		// Test token retrieval
		storedTokens := mediator.RetrieveTokens()
		require.NotNil(t, storedTokens)
		assert.Equal(t, "test-access-token", storedTokens.AccessToken)
		assert.Equal(t, "test-refresh-token", storedTokens.RefreshToken)
	})

	t.Run("Token Retrieval Non-Existent", func(t *testing.T) {
		// Test token retrieval when none exist
		mediator.DeleteTokens()
		
		tokens := mediator.RetrieveTokens()
		assert.Nil(t, tokens)
	})

	t.Run("Token Deletion", func(t *testing.T) {
		// Test token deletion
		tokens := &OAuthTokens{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
		}
		
		err := mediator.StoreTokens(tokens)
		require.NoError(t, err)

		err = mediator.DeleteTokens()
		require.NoError(t, err)

		storedTokens := mediator.RetrieveTokens()
		assert.Nil(t, storedTokens)
	})
}

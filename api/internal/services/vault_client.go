package services

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultClient handles secure secret management using HashiCorp Vault
type VaultClient struct {
	client     *api.Client
	baseURL    string
	token      string
	httpClient *http.Client
}

// VaultSecret represents a secret stored in Vault
type VaultSecret struct {
	Data map[string]interface{} `json:"data"`
}

// NewVaultClient creates a new Vault client
func NewVaultClient(baseURL, token string) (*VaultClient, error) {
	config := &api.Config{
		Address: baseURL,
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultClient{
		client:     client,
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetSecret retrieves a secret from Vault
func (v *VaultClient) GetSecret(path string) (*VaultSecret, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from %s: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret data found at %s", path)
	}

	return &VaultSecret{
		Data: secret.Data,
	}, nil
}

// GetStripeSecrets retrieves Stripe-related secrets from Vault
func (v *VaultClient) GetStripeSecrets(serviceName string) (map[string]string, error) {
	secrets := make(map[string]string)

	// Try to get Stripe secrets from Vault
	secretPath := fmt.Sprintf("complyflow/%s/stripe", serviceName)
	if stripeSecret, err := v.GetSecret(secretPath); err == nil {
		if secretKey, ok := stripeSecret.Data["secret_key"].(string); ok {
			secrets["STRIPE_SECRET_KEY"] = secretKey
		}
		if publishableKey, ok := stripeSecret.Data["publishable_key"].(string); ok {
			secrets["STRIPE_PUBLISHABLE_KEY"] = publishableKey
		}
		if webhookSecret, ok := stripeSecret.Data["webhook_secret"].(string); ok {
			secrets["STRIPE_WEBHOOK_SECRET"] = webhookSecret
		}
	} else {
		log.Printf("Warning: Failed to load Stripe secrets from Vault: %v", err)
	}

	return secrets, nil
}

// GetDatabaseCredentials retrieves database credentials from Vault
func (v *VaultClient) GetDatabaseCredentials(serviceName string) (map[string]string, error) {
	secretPath := fmt.Sprintf("complyflow/%s/database", serviceName)
	secret, err := v.GetSecret(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get database credentials: %w", err)
	}

	creds := make(map[string]string)
	for key, value := range secret.Data {
		if strValue, ok := value.(string); ok {
			creds[key] = strValue
		}
	}

	return creds, nil
}

// GetRedisCredentials retrieves Redis credentials from Vault
func (v *VaultClient) GetRedisCredentials(serviceName string) (map[string]string, error) {
	secretPath := fmt.Sprintf("complyflow/%s/redis", serviceName)
	secret, err := v.GetSecret(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis credentials: %w", err)
	}

	creds := make(map[string]string)
	for key, value := range secret.Data {
		if strValue, ok := value.(string); ok {
			creds[key] = strValue
		}
	}

	return creds, nil
}

// LoadSecretsFromVault loads all secrets for a service from Vault
func (v *VaultClient) LoadSecretsFromVault(serviceName string) (map[string]string, error) {
	secrets := make(map[string]string)

	// Load database credentials
	if dbCreds, err := v.GetDatabaseCredentials(serviceName); err == nil {
		// Build DSN
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			dbCreds["host"], dbCreds["port"], dbCreds["user"],
			dbCreds["password"], dbCreds["name"], dbCreds["ssl_mode"])
		secrets[fmt.Sprintf("%s_DB_DSN", serviceName)] = dsn
	} else {
		log.Printf("Warning: Failed to load database credentials from Vault: %v", err)
	}

	// Load Redis credentials
	if redisCreds, err := v.GetRedisCredentials(serviceName); err == nil {
		secrets["REDIS_HOST"] = redisCreds["host"]
		secrets["REDIS_PORT"] = redisCreds["port"]
		secrets["REDIS_PASSWORD"] = redisCreds["password"]
		secrets["REDIS_DB"] = redisCreds["db"]
	} else {
		log.Printf("Warning: Failed to load Redis credentials from Vault: %v", err)
	}

	// Load Stripe secrets
	if stripeSecrets, err := v.GetStripeSecrets(serviceName); err == nil {
		for key, value := range stripeSecrets {
			secrets[key] = value
		}
	}

	return secrets, nil
}

// HealthCheck checks if Vault is accessible
func (v *VaultClient) HealthCheck() error {
	_, err := v.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("Vault health check failed: %w", err)
	}
	return nil
}

// RenewToken renews the Vault token
func (v *VaultClient) RenewToken() error {
	_, err := v.client.Auth().Token().RenewSelf(0)
	if err != nil {
		return fmt.Errorf("failed to renew Vault token: %w", err)
	}
	return nil
}

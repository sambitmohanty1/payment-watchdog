package main

import (
	"context"
	"flag"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// This script helper adds the 'tenant_id' custom claim to a Firebase user.
// Usage: go run scripts/set_tenant_claims.go -email "user@example.com" -tenant "my-smb-001"
func main() {
	email := flag.String("email", "", "User email to update")
	tenantID := flag.String("tenant", "", "Tenant ID to assign")
	serviceAccount := flag.String("certs", "./secrets/firebase-service-account.json", "Path to Firebase service account JSON")
	flag.Parse()

	if *email == "" || *tenantID == "" {
		log.Fatal("Usage: go run scripts/set_tenant_claims.go -email <email> -tenant <tenant_id>")
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(*serviceAccount)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing app: %v\n", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Auth client: %v\n", err)
	}

	user, err := client.GetUserByEmail(ctx, *email)
	if err != nil {
		log.Fatalf("Error getting user by email %s: %v\n", *email, err)
	}

	// Set custom user claims
	claims := map[string]interface{}{"tenant_id": *tenantID}
	err = client.SetCustomUserClaims(ctx, user.UID, claims)
	if err != nil {
		log.Fatalf("Error setting custom claims for user %s: %v\n", user.UID, err)
	}

	log.Printf("Successfully assigned user %s (%s) to tenant %s\n", user.DisplayName, *email, *tenantID)
}

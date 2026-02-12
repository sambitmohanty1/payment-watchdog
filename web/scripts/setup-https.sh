#!/bin/bash

# 🔒 Complete HTTPS Setup Script for Lexure Intelligence UI
# Sets up HTTPS for both development and production

set -e

echo "🔒 Setting up HTTPS for Lexure Intelligence UI..."
echo "=============================================="

# Step 1: Generate SSL certificates
echo "📝 Step 1: Generating SSL certificates..."
./scripts/generate-ssl-certs.sh
echo ""

# Step 2: Generate Kubernetes TLS secret
echo "📝 Step 2: Generating Kubernetes TLS secret..."
./scripts/generate-k8s-tls-secret.sh
echo ""

# Step 3: Apply Kubernetes resources
echo "📝 Step 3: Applying Kubernetes resources..."
echo "Applying TLS secret..."
kubectl apply -f deployments/kubernetes/tls-secret.yaml

echo "Applying ingress configuration..."
kubectl apply -f deployments/kubernetes/ingress.yaml
echo ""

# Step 4: Verify setup
echo "📝 Step 4: Verifying HTTPS setup..."
echo "Checking TLS secret..."
kubectl get secret lexure-intelligence-ui-tls -n lexure-intelligence-ui

echo ""
echo "Checking ingress..."
kubectl get ingress lexure-intelligence-ui-ingress -n lexure-intelligence-ui
echo ""

echo "✅ HTTPS setup completed successfully!"
echo ""
echo "🚀 Available HTTPS endpoints:"
echo "   Development: https://localhost:3050"
echo "   Kubernetes:  https://localhost:3001 (via ingress)"
echo ""
echo "🧪 Test HTTPS access:"
echo "   curl -k https://localhost:3050/health"
echo ""
echo "🔧 OAuth callback URLs for HTTPS:"
echo "   Xero: https://localhost:3050/auth/xero/callback"
echo "   QuickBooks: https://localhost:3050/auth/quickbooks/callback"
echo "   Stripe: https://localhost:3050/api/v1/webhooks/stripe"

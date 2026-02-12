#!/bin/bash

# Stripe CLI Setup Script for Lexure Intelligence MVP
# This script helps set up Stripe CLI for webhook testing

set -e

echo "🚀 Setting up Stripe CLI for Lexure Intelligence MVP"
echo "=================================================="

# Check if Stripe CLI is installed
if ! command -v stripe &> /dev/null; then
    echo "❌ Stripe CLI is not installed."
    echo ""
    echo "📥 Installing Stripe CLI..."
    
    # Detect OS and install accordingly
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        echo "Installing on macOS..."
        brew install stripe/stripe-cli/stripe
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        echo "Installing on Linux..."
        curl -s https://packages.stripe.dev/api/security/keypair/stripe-cli-gpg/public | gpg --dearmor | sudo tee /usr/share/keyrings/stripe.gpg
        echo "deb [signed-by=/usr/share/keyrings/stripe.gpg] https://packages.stripe.dev/stripe-cli-debian-local stable main" | sudo tee -a /etc/apt/sources.list.d/stripe.list
        sudo apt update
        sudo apt install stripe
    else
        echo "❌ Unsupported OS: $OSTYPE"
        echo "Please install Stripe CLI manually from: https://stripe.com/docs/stripe-cli"
        exit 1
    fi
else
    echo "✅ Stripe CLI is already installed"
fi

echo ""
echo "🔐 Authenticating with Stripe..."
echo "This will open your browser to authenticate with Stripe"

# Login to Stripe
stripe login

echo ""
echo "✅ Stripe CLI setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Start your backend server on port 8085"
echo "2. Run: ./scripts/test_stripe_webhooks.sh"
echo "3. Or manually: stripe listen --forward-to localhost:8085/api/v1/webhooks/stripe?company_id=test_company"
echo ""
echo "🎯 Test webhook events:"
echo "   stripe trigger payment_intent.payment_failed"
echo "   stripe trigger payment_intent.succeeded"
echo "   stripe trigger customer.subscription.deleted"
echo ""
echo "📖 For more info: https://stripe.com/docs/stripe-cli"

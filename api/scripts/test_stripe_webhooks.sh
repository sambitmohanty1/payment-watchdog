#!/bin/bash

# Stripe Webhook Testing Script for Lexure Intelligence MVP
# This script tests the complete webhook processing pipeline

set -e

# Configuration
BACKEND_URL="http://localhost:8085"
COMPANY_ID="test_company_$(date +%s)"
TEST_EVENTS=(
    "payment_intent.payment_failed"
    "payment_intent.succeeded"
    "customer.subscription.deleted"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🧪 Testing Stripe Webhook Integration"
echo "===================================="
echo "Backend URL: $BACKEND_URL"
echo "Company ID: $COMPANY_ID"
echo ""

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "success") echo -e "${GREEN}✅ $message${NC}" ;;
        "error") echo -e "${RED}❌ $message${NC}" ;;
        "warning") echo -e "${YELLOW}⚠️  $message${NC}" ;;
        "info") echo -e "${BLUE}ℹ️  $message${NC}" ;;
    esac
}

# Function to check if backend is running
check_backend() {
    print_status "info" "Checking if backend is running..."
    
    if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
        print_status "success" "Backend is running"
        return 0
    else
        print_status "error" "Backend is not running on $BACKEND_URL"
        print_status "info" "Please start the backend server first"
        return 1
    fi
}

# Function to test webhook endpoint
test_webhook_endpoint() {
    print_status "info" "Testing webhook endpoint..."
    
    local response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Stripe-Signature: test_signature" \
        -d '{"test": true}' \
        "$BACKEND_URL/api/v1/webhooks/stripe?company_id=$COMPANY_ID")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "400" ]; then
        print_status "success" "Webhook endpoint is accessible (expected 400 for invalid signature)"
    else
        print_status "warning" "Unexpected response: HTTP $http_code"
        echo "Response: $body"
    fi
}

# Function to test test webhook endpoint
test_test_webhook() {
    print_status "info" "Testing test webhook endpoint..."
    
    local response=$(curl -s -w "%{http_code}" \
        -X POST \
        "$BACKEND_URL/api/v1/webhooks/test?company_id=$COMPANY_ID")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_status "success" "Test webhook endpoint working"
        echo "Response: $body"
    else
        print_status "error" "Test webhook failed: HTTP $http_code"
        echo "Response: $body"
    fi
}

# Function to test metrics endpoint
test_metrics() {
    print_status "info" "Testing metrics endpoint..."
    
    local response=$(curl -s -w "%{http_code}" \
        "$BACKEND_URL/api/v1/dashboard/quality?company_id=$COMPANY_ID&type=daily")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_status "success" "Metrics endpoint working"
        echo "Response: $body"
    else
        print_status "warning" "Metrics endpoint: HTTP $http_code"
        echo "Response: $body"
    fi
}

# Function to test Stripe CLI webhook forwarding
test_stripe_cli() {
    print_status "info" "Testing Stripe CLI webhook forwarding..."
    
    if ! command -v stripe &> /dev/null; then
        print_status "error" "Stripe CLI not installed. Run ./scripts/stripe_cli_setup.sh first"
        return 1
    fi
    
    print_status "info" "Starting Stripe webhook listener in background..."
    
    # Start webhook listener in background
    stripe listen --forward-to "$BACKEND_URL/api/v1/webhooks/stripe?company_id=$COMPANY_ID" &
    local stripe_pid=$!
    
    # Wait for listener to start
    sleep 3
    
    print_status "info" "Triggering test payment failure event..."
    
    # Trigger a test event
    local trigger_output=$(stripe trigger payment_intent.payment_failed 2>&1)
    
    if echo "$trigger_output" | grep -q "Trigger succeeded"; then
        print_status "success" "Test event triggered successfully"
        
        # Wait for webhook processing
        sleep 2
        
        # Check if event was processed
        local response=$(curl -s "$BACKEND_URL/api/v1/dashboard/stats?company_id=$COMPANY_ID")
        if echo "$response" | grep -q "payment_failures"; then
            print_status "success" "Webhook event processed and data available in dashboard"
        else
            print_status "warning" "Event may not have been processed yet"
        fi
    else
        print_status "error" "Failed to trigger test event"
        echo "Trigger output: $trigger_output"
    fi
    
    # Stop the webhook listener
    kill $stripe_pid 2>/dev/null || true
}

# Function to run comprehensive tests
run_tests() {
    print_status "info" "Running comprehensive webhook tests..."
    echo ""
    
    # Test 1: Backend connectivity
    if ! check_backend; then
        exit 1
    fi
    echo ""
    
    # Test 2: Webhook endpoint
    test_webhook_endpoint
    echo ""
    
    # Test 3: Test webhook endpoint
    test_test_webhook
    echo ""
    
    # Test 4: Metrics endpoint
    test_metrics
    echo ""
    
    # Test 5: Stripe CLI integration
    test_stripe_cli
    echo ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --backend-url URL    Set backend URL (default: http://localhost:8085)"
    echo "  --company-id ID      Set company ID (default: auto-generated)"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run with default settings"
    echo "  $0 --backend-url http://localhost:3000  # Custom backend URL"
    echo "  $0 --company-id my_company            # Custom company ID"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --backend-url)
            BACKEND_URL="$2"
            shift 2
            ;;
        --company-id)
            COMPANY_ID="$2"
            shift 2
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
echo "🚀 Starting Stripe Webhook Integration Tests"
echo "==========================================="
echo "Backend URL: $BACKEND_URL"
echo "Company ID: $COMPANY_ID"
echo ""

run_tests

echo ""
echo "🎉 Webhook testing complete!"
echo ""
echo "📋 Next steps:"
echo "1. Check the dashboard for processed events"
echo "2. Verify data quality metrics"
echo "3. Test with real Stripe webhooks in production"
echo ""
echo "🔗 Useful links:"
echo "- Stripe CLI docs: https://stripe.com/docs/stripe-cli"
echo "- Webhook testing: https://stripe.com/docs/webhooks/test"
echo "- Dashboard: $BACKEND_URL (when UI is running)"

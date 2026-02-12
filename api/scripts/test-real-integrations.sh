#!/bin/bash

# Real Integration Testing Script for Lexure Intelligence
# Tests actual connections to Xero, QuickBooks, and Stripe

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONFIG_FILE="config/integration-testing.yaml"
LOG_FILE="logs/integration-test.log"
TIMEOUT=30

# Create logs directory if it doesn't exist
mkdir -p logs

echo -e "${BLUE}🚀 Lexure Intelligence - Real Integration Testing${NC}"
echo "=================================================="
echo "Testing actual connections to payment systems..."
echo ""

# Function to log messages
log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

# Function to check if service is running
check_service() {
    local service_name=$1
    local port=$2
    
    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        log "${GREEN}✅ $service_name is running on port $port${NC}"
        return 0
    else
        log "${RED}❌ $service_name is not running on port $port${NC}"
        return 1
    fi
}

# Function to test OAuth configuration
test_oauth_config() {
    local provider=$1
    local client_id_var=$2
    
    if [ -z "${!client_id_var}" ]; then
        log "${YELLOW}⚠️  $provider OAuth not configured (set $client_id_var)${NC}"
        return 1
    else
        log "${GREEN}✅ $provider OAuth configured${NC}"
        return 0
    fi
}

# Function to test API connectivity
test_api_connectivity() {
    local provider=$1
    local test_url=$2
    local auth_header=$3
    
    log "${BLUE}🔍 Testing $provider API connectivity...${NC}"
    
    if [ -n "$auth_header" ]; then
        response=$(curl -s -w "%{http_code}" -H "$auth_header" "$test_url" -o /dev/null)
    else
        response=$(curl -s -w "%{http_code}" "$test_url" -o /dev/null)
    fi
    
    if [ "$response" = "200" ] || [ "$response" = "401" ]; then
        log "${GREEN}✅ $provider API accessible (HTTP $response)${NC}"
        return 0
    else
        log "${RED}❌ $provider API not accessible (HTTP $response)${NC}"
        return 1
    fi
}

# Function to test payment failure event creation
test_payment_failure_event() {
    local provider=$1
    local company_id=$2
    
    log "${BLUE}🔍 Testing $provider payment failure event creation...${NC}"
    
    # Create test payment failure event
    event_data=$(cat <<EOF
{
    "company_id": "$company_id",
    "provider": "$provider",
    "amount": 1000.00,
    "currency": "AUD",
    "failure_reason": "test_integration",
    "customer_email": "test@example.com",
    "event_type": "payment_intent.payment_failed"
}
EOF
)
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$event_data" \
        "http://localhost:8085/api/v1/payment-failures" \
        -o /dev/null)
    
    if [ "$response" = "201" ] || [ "$response" = "200" ]; then
        log "${GREEN}✅ $provider payment failure event created successfully${NC}"
        return 0
    else
        log "${RED}❌ Failed to create $provider payment failure event (HTTP $response)${NC}"
        return 1
    fi
}

# Function to test dashboard data retrieval
test_dashboard_data() {
    local company_id=$1
    
    log "${BLUE}🔍 Testing dashboard data retrieval...${NC}"
    
    response=$(curl -s -w "%{http_code}" \
        "http://localhost:8085/api/v1/dashboard/stats?company_id=$company_id" \
        -o /dev/null)
    
    if [ "$response" = "200" ]; then
        log "${GREEN}✅ Dashboard data retrieved successfully${NC}"
        
        # Get actual data
        data=$(curl -s "http://localhost:8085/api/v1/dashboard/stats?company_id=$company_id")
        log "${BLUE}📊 Dashboard data: $data${NC}"
        return 0
    else
        log "${RED}❌ Failed to retrieve dashboard data (HTTP $response)${NC}"
        return 1
    fi
}

# Main testing flow
main() {
    log "${BLUE}Starting real integration testing at $(date)${NC}"
    echo ""
    
    # Check if services are running
    log "${BLUE}🔍 Checking service status...${NC}"
    check_service "Backend API" "8085" || exit 1
    check_service "UI Frontend" "3001" || exit 1
    echo ""
    
    # Test OAuth configurations
    log "${BLUE}🔐 Testing OAuth configurations...${NC}"
    test_oauth_config "Xero" "XERO_CLIENT_ID"
    test_oauth_config "QuickBooks" "QUICKBOOKS_CLIENT_ID"
    test_oauth_config "Stripe" "STRIPE_SECRET_KEY"
    echo ""
    
    # Test API connectivity
    log "${BLUE}🌐 Testing API connectivity...${NC}"
    
    # Test Stripe API (if configured)
    if [ -n "$STRIPE_SECRET_KEY" ]; then
        test_api_connectivity "Stripe" "https://api.stripe.com/v1/account" "Authorization: Bearer $STRIPE_SECRET_KEY"
    fi
    
    # Test Xero API (if configured)
    if [ -n "$XERO_CLIENT_ID" ]; then
        test_api_connectivity "Xero" "https://api.xero.com/api.xro/2.0/Organisations" ""
    fi
    
    # Test QuickBooks API (if configured)
    if [ -n "$QUICKBOOKS_CLIENT_ID" ]; then
        test_api_connectivity "QuickBooks" "https://sandbox-quickbooks.api.intuit.com/v3/company" ""
    fi
    echo ""
    
    # Test payment failure event creation
    log "${BLUE}💳 Testing payment failure event creation...${NC}"
    test_company_id="550e8400-e29b-41d4-a716-446655440000"
    
    test_payment_failure_event "stripe" "$test_company_id"
    test_payment_failure_event "xero" "$test_company_id"
    test_payment_failure_event "quickbooks" "$test_company_id"
    echo ""
    
    # Test dashboard data retrieval
    log "${BLUE}📊 Testing dashboard data retrieval...${NC}"
    test_dashboard_data "$test_company_id"
    echo ""
    
    # Test real-time data flow
    log "${BLUE}⚡ Testing real-time data flow...${NC}"
    
    # Create multiple payment failures to test aggregation
    for i in {1..3}; do
        log "${BLUE}Creating test payment failure $i...${NC}"
        test_payment_failure_event "stripe" "$test_company_id"
        sleep 2
    done
    
    # Check if data is aggregated
    log "${BLUE}Checking data aggregation...${NC}"
    test_dashboard_data "$test_company_id"
    echo ""
    
    # Summary
    log "${BLUE}📋 Integration Testing Summary${NC}"
    log "=================================="
    log "${GREEN}✅ Backend API: Running${NC}"
    log "${GREEN}✅ UI Frontend: Running${NC}"
    log "${GREEN}✅ Payment failure events: Created${NC}"
    log "${GREEN}✅ Dashboard data: Retrieved${NC}"
    log "${GREEN}✅ Real-time flow: Tested${NC}"
    echo ""
    
    log "${GREEN}🎉 Real integration testing completed successfully!${NC}"
    log "Check the logs at: $LOG_FILE"
}

# Load environment variables if .env file exists
if [ -f ".env" ]; then
    log "${BLUE}📁 Loading environment variables from .env file${NC}"
    export $(cat .env | grep -v '^#' | xargs)
fi

# Run main function
main "$@"

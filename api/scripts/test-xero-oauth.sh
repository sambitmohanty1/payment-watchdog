#!/bin/bash

# Xero OAuth Flow Testing Script
# This script tests the complete Xero OAuth integration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:8085"
COMPANY_ID="test-company-$(date +%s)"

echo -e "${BLUE}🧪 Starting Xero OAuth Flow Testing${NC}"
echo "=================================="

# Test 1: Health Check
echo -e "\n${YELLOW}Test 1: Backend Health Check${NC}"
HEALTH_RESPONSE=$(curl --noproxy localhost -s "$BACKEND_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✅ Backend is healthy${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
    echo "Response: $HEALTH_RESPONSE"
    exit 1
fi

# Test 2: Generate Authorization URL
echo -e "\n${YELLOW}Test 2: Generate Authorization URL${NC}"
AUTH_RESPONSE=$(curl --noproxy localhost -X POST "$BACKEND_URL/api/v1/xero/auth/authorize" \
    -H "Content-Type: application/json" \
    -d "{\"company_id\": \"$COMPANY_ID\"}" \
    -s)

if echo "$AUTH_RESPONSE" | grep -q "authorization_url"; then
    echo -e "${GREEN}✅ Authorization URL generated successfully${NC}"
    AUTH_URL=$(echo "$AUTH_RESPONSE" | jq -r '.authorization_url')
    STATE=$(echo "$AUTH_RESPONSE" | jq -r '.state')
    echo "Authorization URL: $AUTH_URL"
    echo "State: $STATE"
    
    # Save for manual testing
    echo "$AUTH_URL" > /tmp/xero_auth_url.txt
    echo "Authorization URL saved to /tmp/xero_auth_url.txt"
else
    echo -e "${RED}❌ Failed to generate authorization URL${NC}"
    echo "Response: $AUTH_RESPONSE"
    exit 1
fi

# Test 3: Test Invalid Callback
echo -e "\n${YELLOW}Test 3: Test Invalid Callback (Error Handling)${NC}"
CALLBACK_RESPONSE=$(curl --noproxy localhost -X POST "$BACKEND_URL/api/v1/xero/auth/callback" \
    -H "Content-Type: application/json" \
    -d '{"code": "invalid_code", "state": "'$STATE'"}' \
    -s)

if echo "$CALLBACK_RESPONSE" | grep -q "error"; then
    echo -e "${GREEN}✅ Error handling works correctly${NC}"
    echo "Error response: $CALLBACK_RESPONSE"
else
    echo -e "${RED}❌ Error handling failed${NC}"
    echo "Response: $CALLBACK_RESPONSE"
fi

echo -e "\n${BLUE}🎯 Manual Testing Required${NC}"
echo "=================================="
echo "1. Open the authorization URL in your browser:"
echo "   $AUTH_URL"
echo ""
echo "2. Complete the OAuth flow in Xero"
echo ""
echo "3. After authorization, test the callback with the real authorization code"
echo ""
echo "4. Use the access token to test API endpoints:"
echo "   - GET /api/v1/xero/tenants"
echo "   - GET /api/v1/xero/organizations"
echo "   - GET /api/v1/xero/payment-failures"
echo ""
echo -e "${GREEN}✅ Automated tests completed successfully!${NC}"

#!/bin/bash

# Test script for Lexure Intelligence MVP webhooks
# This script tests both the test webhook and Stripe webhook endpoints

BASE_URL="http://localhost:8085/api/v1"
COMPANY_ID="test-company-123"

echo "🧪 Testing Lexure Intelligence MVP Webhooks"
echo "=========================================="

# Test 1: Test webhook endpoint
echo ""
echo "📡 Test 1: Testing test webhook endpoint..."
curl -X POST "${BASE_URL}/webhooks/test?company_id=${COMPANY_ID}" \
  -H "Content-Type: application/json" \
  -H "X-Company-ID: ${COMPANY_ID}" \
  -d '{"test": true}' \
  -w "\nHTTP Status: %{http_code}\n"

# Test 2: Test webhook endpoint with header company ID
echo ""
echo "📡 Test 2: Testing test webhook endpoint with header company ID..."
curl -X POST "${BASE_URL}/webhooks/test" \
  -H "Content-Type: application/json" \
  -H "X-Company-ID: ${COMPANY_ID}" \
  -d '{"test": true}' \
  -w "\nHTTP Status: %{http_code}\n"

# Test 3: Test data quality report
echo ""
echo "📊 Test 3: Testing data quality report endpoint..."
curl -X GET "${BASE_URL}/dashboard/quality?company_id=${COMPANY_ID}&type=daily" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n"

# Test 4: Test data quality trends
echo ""
echo "📈 Test 4: Testing data quality trends endpoint..."
curl -X GET "${BASE_URL}/dashboard/quality/trends?company_id=${COMPANY_ID}&days=7" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n"

# Test 5: Test dashboard stats (should now show real data)
echo ""
echo "📊 Test 5: Testing dashboard stats endpoint..."
curl -X GET "${BASE_URL}/dashboard/stats?company_id=${COMPANY_ID}" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n"

# Test 6: Test payment failures (should now show real data)
echo ""
echo "💳 Test 6: Testing payment failures endpoint..."
curl -X GET "${BASE_URL}/failures?company_id=${COMPANY_ID}&page=1&limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "✅ Webhook testing completed!"
echo ""
echo "💡 Next steps:"
echo "   1. Check the logs for webhook processing details"
echo "   2. Verify data is being stored in the database"
echo "   3. Test with actual Stripe webhook events"
echo "   4. Configure real Stripe webhook secret in production"

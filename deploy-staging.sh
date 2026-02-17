#!/bin/bash
# Payment Watchdog Staging Deployment Script (Non-Docker)

echo "🚀 Starting Payment Watchdog Staging Deployment..."

# Check if services are running
echo "📊 Checking service status..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Start API service
echo "⚙️ Starting API service..."
cd /Users/mohants5/Documents/personal/projects/payment-watchdog/api
ENVIRONMENT=staging \
SERVER_PORT=8091 \
DATABASE_HOST=localhost \
DATABASE_PORT=5443 \
DATABASE_NAME=payment_watchdog_staging \
DATABASE_PASSWORD=postgres_staging \
REDIS_URL=redis://localhost:6390 \
./payment-watchdog-api &
API_PID=$!

# Wait for API to be ready
echo "⏳ Waiting for API to be ready..."
sleep 5

# Check API health
echo "🔍 Checking API health..."
curl --noproxy localhost -f http://localhost:8091/health || {
    echo "❌ API health check failed"
    kill $API_PID
    exit 1
}

# Start Web service
echo "🌐 Starting Web service..."
cd /Users/mohants5/Documents/personal/projects/payment-watchdog/web
PORT=3011 NODE_ENV=production npm start &
WEB_PID=$!

# Wait for Web service to be ready
echo "⏳ Waiting for Web service to be ready..."
sleep 5

# Check Web service
echo "🔍 Checking Web service..."
curl --noproxy localhost -f http://localhost:3011 || {
    echo "❌ Web service check failed"
    kill $API_PID $WEB_PID
    exit 1
}

echo "✅ Deployment completed successfully!"
echo "📊 Service URLs:"
echo "   API: http://localhost:8091"
echo "   Web: http://localhost:3011"
echo "   Database: localhost:5443"
echo "   Redis: localhost:6390"

# Save PIDs for cleanup
echo $API_PID > /tmp/payment-watchdog-api.pid
echo $WEB_PID > /tmp/payment-watchdog-web.pid

echo "💾 Process IDs saved. Use 'kill -9 \$(cat /tmp/payment-watchdog-*.pid)' to stop all services."

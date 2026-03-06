#!/bin/bash
# Payment Watchdog Staging Deployment Script (Non-Docker)

echo "🚀 Starting Payment Watchdog Staging Deployment..."

# AC 1.2: Check REGION if SOVEREIGN_MODE is active
if [ "$SOVEREIGN_MODE" = "true" ]; then
    echo "🛡️ Sovereign Mode is ACTIVE. Validating REGION..."
    if [[ "$REGION" != "ap-southeast-2" && "$REGION" != "australia-southeast1" && "$REGION" != "australia-southeast2" && "$REGION" != "ap-sydney-1" && "$REGION" != "ap-melbourne-1" && "$REGION" != "australiaeast" && "$REGION" != "australiasoutheast" ]]; then
        echo "❌ ERROR: Target region '$REGION' is outside of Australia, but SOVEREIGN_MODE is true. Aborting."
        exit 1
    fi
    echo "✅ Region validation passed."
fi

# AC 4.2: Residency Report Generation
if [ "$SOVEREIGN_MODE" = "true" ]; then
    echo "📄 Generating Residency Report..."
    DB_IP=$(dig +short lexure-mvp-postgres.lexure.svc.cluster.local || echo "Internal DNS")
    REDIS_IP=$(dig +short redis-service.payment-watchdog.svc.cluster.local || echo "Internal DNS")
    VAULT_IP=$(dig +short vault.payment-watchdog.svc.cluster.local || echo "N/A")

    cat <<EOF > residency_report.txt
=============================================
SOVEREIGN RESIDENCY REPORT
=============================================
Deployment Region: $REGION
Date: $(date)
---------------------------------------------
Database Endpoint: lexure-mvp-postgres.lexure.svc.cluster.local
Database IP: $DB_IP
Redis Endpoint: redis-service.payment-watchdog.svc.cluster.local
Redis IP: $REDIS_IP
Vault Endpoint: vault.payment-watchdog.svc.cluster.local
Vault IP: $VAULT_IP
=============================================
EOF
    echo "✅ Residency Report saved to residency_report.txt"
fi

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

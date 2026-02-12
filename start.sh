#!/bin/bash

# Payment Watchdog Development Startup Script

set -e

echo "🚀 Starting Payment Watchdog Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ docker-compose is not installed."
    exit 1
fi

# Function to wait for service to be healthy
wait_for_service() {
    local service_name=$1
    local max_attempts=30
    local attempt=1

    echo "⏳ Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        if docker-compose ps $service_name | grep -q "healthy\|running"; then
            echo "✅ $service_name is ready!"
            return 0
        fi

        echo "   Attempt $attempt/$max_attempts: $service_name not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done

    echo "❌ $service_name failed to start within expected time"
    return 1
}

# Start services
echo "🐳 Starting Docker services..."
docker-compose up -d

# Wait for database and Redis to be ready
wait_for_service postgres
wait_for_service redis

# Wait for API to be healthy
wait_for_service api

echo ""
echo "🎉 Payment Watchdog is running!"
echo ""
echo "📊 Services:"
echo "   • API:        http://localhost:8080"
echo "   • Web UI:     http://localhost:3000"
echo "   • Mailhog:    http://localhost:8025"
echo "   • Database:   localhost:5432"
echo "   • Redis:      localhost:6379"
echo ""
echo "📋 Useful commands:"
echo "   • View logs:    docker-compose logs -f"
echo "   • Stop:         docker-compose down"
echo "   • Restart:      docker-compose restart"
echo ""
echo "🔍 Health checks:"
echo "   • API Health:   curl http://localhost:8080/health"
echo "   • API Metrics:  curl http://localhost:8080/metrics"
echo ""

# Show running containers
echo "📦 Running containers:"
docker-compose ps

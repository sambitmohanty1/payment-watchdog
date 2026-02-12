#!/bin/bash

# Lexure Intelligence MVP - Development Setup Script

set -e

echo "🚀 Setting up Lexure Intelligence MVP development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose and try again."
    exit 1
fi

# Navigate to project directory
cd "$(dirname "$0")/.."

echo "📦 Building and starting services..."

# Build and start services
docker-compose up -d --build

echo "⏳ Waiting for services to be ready..."

# Wait for PostgreSQL to be ready
echo "🔄 Waiting for PostgreSQL..."
until docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; do
    sleep 2
done

# Wait for Redis to be ready
echo "🔄 Waiting for Redis..."
until docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
    sleep 2
done

echo "✅ All services are ready!"

# Show service status
echo "📊 Service Status:"
docker-compose ps

# Show service URLs
echo ""
echo "🌐 Service URLs:"
echo "  - Application: http://localhost:8080"
echo "  - Health Check: http://localhost:8080/health"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"

echo ""
echo "📝 Useful Commands:"
echo "  - View logs: docker-compose logs -f app"
echo "  - Stop services: docker-compose down"
echo "  - Restart services: docker-compose restart"
echo "  - Rebuild: docker-compose up -d --build"

echo ""
echo "🎉 Development environment is ready!"
echo "You can now start developing the MVP service."

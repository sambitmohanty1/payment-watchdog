#!/bin/bash

# Start PostgreSQL test database
echo "Starting PostgreSQL test database..."
docker-compose -f docker-compose.test.yml up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 10

# Check if database is ready
if docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U postgres -d payment_watchdog_test; then
    echo "✅ Test database is ready!"
    echo "You can now run: go test -v ./services -run TestGetRecoveryMetrics"
else
    echo "❌ Database failed to start properly"
    docker-compose -f docker-compose.test.yml logs
fi

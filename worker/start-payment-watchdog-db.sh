#!/bin/bash

# Start PostgreSQL database for Payment Watchdog on port 5569
echo "Starting Payment Watchdog PostgreSQL database on port 5569..."

docker-compose -f docker-compose.pg5569.yml up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 10

# Check if database is ready
if docker-compose -f docker-compose.pg5569.yml exec -T postgres-payment-watchdog pg_isready -U postgres -d payment_watchdog; then
    echo "✅ Payment Watchdog database is ready!"
    echo "Connection details:"
    echo "  Host: localhost"
    echo "  Port: 5569"
    echo "  Database: payment_watchdog"
    echo "  User: postgres"
    echo "  Password: payment_watchdog_2024"
    echo ""
    echo "To connect: psql -h localhost -p 5569 -U postgres -d payment_watchdog"
    echo "To stop: docker-compose -f docker-compose.pg5569.yml down"
else
    echo "❌ Database failed to start properly"
    docker-compose -f docker-compose.pg5569.yml logs
fi

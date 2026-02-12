#!/bin/bash

# Port Conflict Checker for Lexure Intelligence MVP

echo "🔍 Checking for potential port conflicts..."

# Define the ports we want to use
MVP_PORTS=("8085" "5403" "6382")

# Check if ports are already in use
conflicts_found=false

for port in "${MVP_PORTS[@]}"; do
    if lsof -i :$port >/dev/null 2>&1; then
        echo "❌ Port $port is already in use:"
        lsof -i :$port
        conflicts_found=true
    else
        echo "✅ Port $port is available"
    fi
done

echo ""

# Check for common conflicting services
echo "🔍 Checking for common conflicting services..."

# Check if lexure-compliance is running
if pgrep -f "lexure-compliance" >/dev/null; then
    echo "⚠️  lexure-compliance is running (typically uses ports 8080, 5432)"
fi

# Check if other services are using our ports
if netstat -tulpn 2>/dev/null | grep -E ':(8080|5432)' >/dev/null; then
    echo "⚠️  Services detected on potentially conflicting ports:"
    netstat -tulpn 2>/dev/null | grep -E ':(8080|5432)'
fi

echo ""

# Summary
if [ "$conflicts_found" = true ]; then
    echo "❌ Port conflicts detected! Please resolve before deploying."
    echo ""
    echo "💡 Solutions:"
    echo "  - Stop conflicting services"
    echo "  - Change ports in config files"
    echo "  - Use different port ranges"
    exit 1
else
    echo "✅ No port conflicts detected!"
    echo "🚀 Ready to deploy Lexure Intelligence MVP"
    echo ""
    echo "📋 Port Configuration:"
    echo "  - Application: 8085"
    echo "  - PostgreSQL: 5435"
    echo "  - Redis: 6382"
fi

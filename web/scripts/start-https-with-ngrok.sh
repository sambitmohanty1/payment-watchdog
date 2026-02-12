#!/bin/bash

# 🔒 Start HTTPS Development Server with ngrok
# Uses ngrok to provide HTTPS tunneling for OAuth callbacks

set -e

echo "🔒 Starting HTTPS development server with ngrok..."

# Check if ngrok is installed
if ! command -v ngrok &> /dev/null; then
    echo "❌ ngrok is not installed. Installing..."
    
    # Detect OS and install ngrok
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install ngrok
        else
            echo "Please install Homebrew first: https://brew.sh/"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null
        echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | sudo tee /etc/apt/sources.list.d/ngrok.list
        sudo apt update && sudo apt install ngrok
    else
        echo "❌ Unsupported OS. Please install ngrok manually: https://ngrok.com/"
        exit 1
    fi
fi

echo "✅ ngrok is installed"

# Check if ngrok is authenticated
if ! ngrok config check &> /dev/null; then
    echo "🔑 Please authenticate ngrok:"
    echo "1. Go to https://dashboard.ngrok.com/get-started/your-authtoken"
    echo "2. Copy your authtoken"
    echo "3. Run: ngrok config add-authtoken YOUR_TOKEN"
    echo ""
    read -p "Press Enter after you've added your authtoken..."
fi

# Start the regular development server in background
echo "🚀 Starting Next.js development server on port 3050..."
npm run dev &
DEV_PID=$!

# Wait for the dev server to start
sleep 5

# Start ngrok tunnel
echo "🌐 Starting ngrok tunnel..."
ngrok http 3050 --log=stdout &
NGROK_PID=$!

# Wait for ngrok to start
sleep 3

# Get the HTTPS URL
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[] | select(.proto=="https") | .public_url')

if [ -z "$NGROK_URL" ]; then
    echo "❌ Failed to get ngrok URL. Check if ngrok is running."
    kill $DEV_PID $NGROK_PID 2>/dev/null || true
    exit 1
fi

echo "✅ HTTPS development server started successfully!"
echo ""
echo "🔗 HTTPS URL: $NGROK_URL"
echo "🔗 Local URL: http://localhost:3050"
echo ""
echo "📋 Update your OAuth app settings:"
echo "   Xero: $NGROK_URL/auth/xero/callback"
echo "   QuickBooks: $NGROK_URL/auth/quickbooks/callback"
echo "   Stripe: $NGROK_URL/api/v1/webhooks/stripe"
echo ""
echo "🧪 Test HTTPS access:"
echo "   curl \"$NGROK_URL/health\""
echo ""
echo "⏹️  Press Ctrl+C to stop both servers"

# Keep script running
trap "echo '🛑 Stopping servers...'; kill $DEV_PID $NGROK_PID 2>/dev/null || true; exit 0" INT
wait $NGROK_PID

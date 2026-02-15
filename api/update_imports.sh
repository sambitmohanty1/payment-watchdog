#!/bin/bash

# 1. First update the main module path
echo "Updating main module path..."
find . -name "*.go" -type f -exec sed -i '' "s|github.com/payment-watchdog|github.com/sambitmohanty1/payment-watchdog/api|g" {} +

# 2. Then update any lexure-intelligence references
echo "Updating lexure-intelligence references..."
find . -name "*.go" -type f -exec sed -i '' "s|github.com/lexure-intelligence/payment-watchdog|github.com/sambitmohanty1/payment-watchdog/api|g" {} +

echo "All imports updated."

# 3. Tidy up go.mod and go.sum
echo "Running go mod tidy..."
go mod tidy

echo "Fix complete. Try running the build again."

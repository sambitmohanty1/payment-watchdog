#!/bin/bash

# 1. Update module imports from generic to specific
OLD_MODULE="github.com/payment-watchdog"
NEW_MODULE="github.com/sambitmohanty1/payment-watchdog/api"

echo "Checking for old import paths..."
grep -r "$OLD_MODULE" .

echo "Replacing '$OLD_MODULE' with '$NEW_MODULE' in api/ directory..."

# Find all .go files and replace the string (macOS compatible)
find . -name "*.go" -type f -exec sed -i '' "s|$OLD_MODULE|$NEW_MODULE|g" {} +

echo "Imports updated."

# 2. Tidy up go.mod and go.sum
echo "Running go mod tidy..."
go mod tidy

echo "Fix complete. Try running the build again."
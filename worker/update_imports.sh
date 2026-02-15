#!/bin/bash

# This script updates all import paths from the old module path to the new one
OLD_MODULE="github.com/lexure-intelligence/payment-watchdog"
NEW_MODULE="github.com/sambitmohanty1/payment-watchdog"

# Find all .go files and update import paths
find . -type f -name "*.go" -exec sed -i '' "s|$OLD_MODULE|$NEW_MODULE|g" {} \;

# Update go.mod
echo "Updating go.mod..."
sed -i '' "s|$OLD_MODULE|$NEW_MODULE|g" go.mod

echo "Import paths have been updated from $OLD_MODULE to $NEW_MODULE"

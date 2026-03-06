#!/bin/bash
# Script to audit kubernetes manifests for US/external cloud resources to maintain Sovereign Compliance

echo "🛡️ Auditing K8s manifests for Sovereign Compliance..."

MANIFESTS_DIR="api/deployments/kubernetes"
VIOLATIONS=0

# Define a list of US or disallowed regions/endpoints
DISALLOWED_STRINGS=(
  "us-east"
  "us-west"
  "us-central"
  "datadoghq.com"
  "newrelic.com"
  ".us-"
  "amazonaws.com" # Unless specified with ap-southeast-2
)

for string in "${DISALLOWED_STRINGS[@]}"; do
    echo "🔍 Checking for: $string"

    # Exclude the script itself and search yaml files
    grep -r -n -i "$string" "$MANIFESTS_DIR" --include="*.yaml" > "violations_$string.tmp" || true

    # Filter exceptions (like allowing .ap-southeast-2.amazonaws.com)
    if [ "$string" = "amazonaws.com" ]; then
        grep -v ".ap-southeast-2.amazonaws.com" "violations_$string.tmp" > "filtered_$string.tmp"
        mv "filtered_$string.tmp" "violations_$string.tmp"
    fi

    if [ -s "violations_$string.tmp" ]; then
        echo "❌ VIOLATION FOUND! Hardcoded external or non-AU resource detected: $string"
        cat "violations_$string.tmp"
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
    rm -f "violations_$string.tmp"
done

if [ $VIOLATIONS -gt 0 ]; then
    echo "🚨 Audit FAILED. $VIOLATIONS violation(s) found. Please update the manifests to ensure local-only/AU endpoints are used."
    exit 1
else
    echo "✅ Audit PASSED. No hardcoded external or non-AU resources found."
    exit 0
fi

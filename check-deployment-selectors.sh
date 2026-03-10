#!/bin/bash

echo "🔍 Checking current deployment selectors in production namespace..."

kubectl config use-context context-ch4w6uqbtwa
kubectl config set-context --current --namespace=production

echo "📋 Current deployment selectors:"
echo "================================"

deployments=("lexure-postgres-prod" "lexure-redis-prod" "payment-watchdog-api-prod" "payment-watchdog-ui-prod" "recovery-orchestration-prod")

for deployment in "${deployments[@]}"; do
    echo "🔍 Checking $deployment..."
    kubectl get deployment $deployment -o jsonpath='{.spec.selector.matchLabels}' 2>/dev/null || echo "  ❌ Deployment not found"
    echo ""
done

echo "📋 Expected selectors from kustomize build:"
echo "========================================"
cd api/deployments/kubernetes/overlays/production
kustomize build . | grep -A 10 "kind: Deployment" | grep -A 5 "selector:"

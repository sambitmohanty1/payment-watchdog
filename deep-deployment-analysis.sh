#!/bin/bash

echo "🔍 Deep Deployment Analysis for Production Namespace"
echo "=================================================="

# Set context
kubectl config use-context context-ch4w6uqbtwa
kubectl config set-context --current --namespace=production

echo ""
echo "📋 CURRENT DEPLOYMENT SELECTORS IN CLUSTER:"
echo "=========================================="

deployments=("lexure-postgres-prod" "lexure-redis-prod" "payment-watchdog-api-prod" "payment-watchdog-ui-prod" "recovery-orchestration-prod")

for deployment in "${deployments[@]}"; do
    echo "🔍 $deployment:"
    if kubectl get deployment $deployment -o jsonpath='{.spec.selector.matchLabels}' 2>/dev/null; then
        echo ""
        echo "  Template labels:"
        kubectl get deployment $deployment -o jsonpath='{.spec.template.metadata.labels}' 2>/dev/null
    else
        echo "  ❌ Deployment not found"
    fi
    echo ""
done

echo "📋 EXPECTED SELECTORS FROM KUSTOMIZE:"
echo "=================================="

cd api/deployments/kubernetes/overlays/production
kustomize build . | grep -A 5 "kind: Deployment" | grep -A 3 "selector:" | head -20

echo ""
echo "📋 HPA SCALE TARGET REF:"
echo "======================"
kustomize build . | grep -A 10 "scaleTargetRef:"

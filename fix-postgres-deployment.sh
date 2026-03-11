#!/bin/bash

echo "🔧 Fix remaining postgres deployment selector issue..."

kubectl config use-context context-ch4w6uqbtwa
kubectl config set-context --current --namespace=production

echo "🗑️ Deleting problematic postgres deployment..."
kubectl delete deployment lexure-postgres-prod --ignore-not-found=true

echo "⏳ Waiting for deletion to complete..."
kubectl wait --for=delete deployment/lexure-postgres-prod --timeout=60s || true

echo "🔄 Recreating postgres deployment with correct selector..."
cd api/deployments/kubernetes/overlays/production
kustomize build . | grep -A 1000 "kind: Deployment" | grep -B 1000 "kind: Service" | head -n -1 | grep -A 50 "lexure-postgres-prod" | kubectl apply -f -

echo "✅ Postgres deployment fixed!"

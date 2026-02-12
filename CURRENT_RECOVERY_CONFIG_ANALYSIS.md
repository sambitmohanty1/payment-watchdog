# Current Recovery Orchestration Configuration Analysis

## 📊 Directory Comparison Summary

### apps/recovery-orchestration/ (Simple Implementation)
**Files:** 7 (deployment, service, hpa, configmap, secret, service-account, kustomization)
**Key Features:**
- Basic deployment with 2 replicas
- Simple HPA (2-10 replicas)
- Port 8086
- Uses `payment-watchdog-mvp:latest` image
- Namespace: `lexure-mvp` (needs update)
- ConfigMap with YAML structure
- Basic secret management

### recovery-orchestration/ (Complex Implementation)
**Files:** 7 (deployment, service, hpa, config, hpa-patch, kustomization)
**Key Features:**
- Advanced deployment with tolerations and node selectors
- Sophisticated HPA with scale up/down policies
- Port 8085 (conflicts with main API)
- Uses `lexure/recovery-orchestration:latest` image
- Namespace: `lexure-mvp` (needs update)
- OpenTelemetry integration
- ConfigMap and Secret generators
- PodDisruptionBudget support
- Provider secrets (Stripe, Xero, QuickBooks)

### base/recovery-orchestration/ (Template)
**Files:** 3 (deployment, service, kustomization)
**Key Features:**
- Basic template structure
- Port 8085
- Uses `recovery-orchestration:latest` image
- No namespace specified
- Minimal configuration

## 🔍 Key Differences Found

### Port Conflicts:
- apps/recovery-orchestration: 8086 ✅
- recovery-orchestration: 8085 ❌ (conflicts with main API)
- base/recovery-orchestration: 8085 ❌ (conflicts with main API)

### Image Naming:
- apps: `payment-watchdog-mvp:latest` (old branding)
- recovery-orchestration: `lexure/recovery-orchestration:latest` (old branding)
- base: `recovery-orchestration:latest` (generic)

### Namespace Issues:
- All use `lexure-mvp` (should be `lexure`)

### Health Check Paths:
- apps: `/health` ✅
- recovery-orchestration: `/healthz` and `/readyz` ❌ (inconsistent)
- base: `/healthz` and `/readyz` ❌ (inconsistent)

## 🎯 Recommended Target Configuration

Based on analysis, target should be:
- **Port**: 8086 (avoid conflicts)
- **Image**: `payment-watchdog/recovery-orchestration:latest`
- **Namespace**: `lexure`
- **Health Check**: `/health` (consistent)
- **Features**: Advanced (from complex version)

## ⚠️ Blocking Issues Identified

1. **Port Conflict**: Complex version uses 8085 which conflicts with main API
2. **Branding**: All versions use old naming conventions
3. **Namespace**: All use `lexure-mvp` instead of `lexure`
4. **Service Addresses**: Complex version has hardcoded observability addresses

## 📋 Questions for User

1. **Backup Strategy**: Git version control vs additional backups?
2. **Current Working Version**: Which implementation is currently deployed?
3. **Observability Stack**: Are the OpenTelemetry addresses correct?
4. **Secret Values**: Use placeholders or real values?

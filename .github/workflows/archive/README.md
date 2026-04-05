# Archive: deployment-patch.yml

## Historical Context
This file contains the original deployment manifests for Payment Watchdog services. It was used before the current CI/CD pipeline refactoring (commit 9547e26e).

## Purpose
- Served as reference templates for manual deployments
- Contains production-ready Kubernetes deployment configurations
- Used for staging, sovereign-au, and production deployments

## Key Features
- **API Service**: payment-watchdog-api with PostgreSQL and Redis configuration
- **Web UI**: payment-watchdog-ui with proper image pull policies
- **Recovery Orchestration**: recovery-orchestration service configuration

## Migration Notes
- Replaced by individual deployment workflows in ci-cd-pipeline.yml
- No longer needed for current CI/CD architecture
- Preserved for historical reference and potential future use

## Status
- **Archived**: Moved to `archive/` folder on 2026-04-05
- **Reason**: CI/CD pipeline now uses separate deployment workflows
- **Action**: Safe to keep as reference

This file is kept for documentation purposes and historical context.

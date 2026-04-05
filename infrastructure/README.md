# Infrastructure

This directory contains all infrastructure-as-code and deployment configurations for Payment Watchdog.

## Structure

```
infrastructure/
├── kubernetes/          # Kubernetes manifests and Kustomize configurations
│   ├── base/           # Base configurations shared across environments
│   │   ├── components/ # Shared infrastructure components (postgres, redis, monitoring)
│   │   └── services/   # Service-specific base configurations
│   ├── overlays/       # Environment-specific configurations
│   └── scripts/        # Deployment and utility scripts
├── terraform/          # Terraform modules for cloud infrastructure
├── helm/              # Helm charts (optional alternative to Kustomize)
└── docker-compose/    # Local development configurations
```

## Usage

### Development
```bash
# Deploy to development environment
kubectl apply -k infrastructure/kubernetes/overlays/development

# Local development with Docker Compose
docker-compose -f infrastructure/docker-compose/development.yml up
```

### Production
```bash
# Deploy to sovereign-au environment
kubectl apply -k infrastructure/kubernetes/overlays/sovereign-au

# Validate configuration before deployment
kustomize build infrastructure/kubernetes/overlays/sovereign-au | kubeval
```

## Principles

1. **Infrastructure as Code** - All infrastructure is versioned and reviewed
2. **Environment Parity** - All environments use the same base with controlled variations
3. **Sovereign Compliance** - Australian data residency and security requirements built-in
4. **GitOps Ready** - Configurations structured for automated deployment pipelines
5. **Validation First** - All configurations validated before deployment

## Ownership

- **Platform Engineering**: Infrastructure components and deployment automation
- **Development Teams**: Service-specific configurations
- **Security**: Compliance validation and policy enforcement

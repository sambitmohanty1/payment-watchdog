# Payment Watchdog

A payment recovery platform for Australian businesses that automatically detects and recovers failed payments through intelligent retry logic and multi-provider integration.

## Status

**Current State**: Development in progress  
**Last Updated**: April 2026  
**Go Version**: 1.25.0  
**Next.js**: 14.2.3  

## Quick Start

```bash
# Clone and setup
git clone https://github.com/sambitmohanty1/payment-watchdog.git
cd payment-watchdog

# Start development environment
docker-compose up -d

# Access services
# API: http://localhost:8080
# Web: http://localhost:4896
# Database: localhost:5432
# Redis: localhost:6379
```

### **Infrastructure Access (Sovereign AU)**
- **API (ClusterIP)**: `10.96.158.63` (Internal)
- **API (LoadBalancer)**: `207.211.158.1` (Public)
- **Web Interface**: `168.138.21.140` (Public)
- **Database**: `postgres.sovereign-au.svc.cluster.local` (Port 5432)
- **Redis**: `redis.sovereign-au.svc.cluster.local` (Port 6379)
```

## What It Does

Payment Watchdog monitors payment failures and automatically attempts recovery through:

- **Smart Retry Logic**: Exponential backoff with provider-specific routing
- **Multi-Provider Support**: Stripe, PayTo, BECS, Xero integration  
- **Analytics Engine**: Pattern detection and failure prediction
- **Australian Data Residency**: Sovereign mode for AU compliance
- **Real-time Dashboard**: Payment metrics and recovery monitoring

## Development

### Prerequisites
- Go 1.25.0+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Building Services
```bash
# Build API
cd api && go build -o payment-watchdog-api ./cmd/main.go

# Build Worker  
cd worker && go build -o payment-watchdog-worker ./cmd/main.go

# Build Web
cd web && npm run build
```

### Testing
```bash
# Run all tests with coverage
go test ./... -v -race -cover

# Run specific service tests
cd api && go test ./... -v
cd worker && go test ./... -v
```

### Deployment

#### Environments
| Environment | Config File | API Port | Web Port | Database |
|------------|--------------|-----------|-----------|----------|
| Development | docker-compose.yml | 8080 | 4896 | payment_watchdog |
| Staging | docker-compose.staging.yml | 8091 | 3011 | payment_watchdog_staging |

#### Local Deployment
```bash
# Standard development
docker-compose up -d

# Staging environment  
docker-compose -f docker-compose.staging.yml up -d

# Production (requires Kubernetes setup)
kubectl apply -f infrastructure/kubernetes/
```

---

## Documentation Structure

### **📋 Strategic Documents**
- **[📊 FUTURE_STATE.md](FUTURE_STATE.md)** - Strategic roadmap and market analysis
- **[🎯 FEATURE_BACKLOG.md](FEATURE_BACKLOG.md)** - Product features and user stories
- **[🏢 PLATFORM_BACKLOG.md](PLATFORM_BACKLOG.md)** - Technical debt and platform improvements

### **🏗️ Architecture Documentation**
- **System Design**: [SYSTEM_DESIGN.md](docs/ARCHITECTURE/SYSTEM_DESIGN.md)
- **Low-Level Design**: [LOW_LEVEL_DESIGN.md](docs/ARCHITECTURE/LOW_LEVEL_DESIGN.md)
- **API Documentation**: [API.md](docs/ARCHITECTURE/API_SPECIFICATION.md)
- **Worker Documentation**: [WORKER.md](docs/WORKER.md)
- **[SYSTEM_DESIGN.md](docs/ARCHITECTURE/SYSTEM_DESIGN.md)** - High-level system architecture
- **[API_SPECIFICATION.md](docs/ARCHITECTURE/API_SPECIFICATION.md)** - Complete API documentation
- **[BUSINESS_REQUIREMENTS.md](docs/STRATEGIC/BUSINESS_REQUIREMENTS.md)** - Business requirements and user stories

### **🔧 Operations Documentation**
- **[SECURITY.md](SECURITY.md)** - Security policies and compliance
- **[LOGGING_GUIDELINES.md](docs/OPERATIONS/LOGGING_GUIDELINES.md)** - Logging standards and best practices
- **[ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)** - Configuration reference

### **📚 Reference Documentation**
- **[CURRENCY_GUIDELINES.md](api/CURRENCY_GUIDELINES.md)** - Currency field usage guidelines

### **🗃️ Archived Documentation**
- **[ARCHIVED/TECH_DEBT_BACKLOG_OLD.md](docs/ARCHIVED/TECH_DEBT_BACKLOG_OLD.md)** - Previous technical debt backlog
- **[ARCHIVED/LOGGING_IMPROVEMENTS_OLD.md](docs/ARCHIVED/LOGGING_IMPROVEMENTS_OLD.md)** - Previous logging improvements

---

## Architecture

```mermaid
graph TB
    subgraph "Frontend"
        WEB[Next.js Dashboard]
    end
    
    subgraph "Backend Services"
        API[Go API Service]
        WORKER[Background Worker]
        WEBHOOK[Webhook Service]
        ANALYTICS[Analytics Service]
    end
    
    subgraph "Data Layer"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
    end
    
    subgraph "External Services"
        STRIPE[Stripe API]
        PAYTO[PayTo/NPP API]
        XERO[Xero API]
        BNPL[BNPL Providers]
    end
    
    WEB --> API
    API --> POSTGRES
    API --> REDIS
    WORKER --> POSTGRES
    WORKER --> REDIS
    API --> STRIPE
    WORKER --> PAYTO
    WORKER --> XERO
```

### Service Responsibilities

#### API Service (Port 8080)
- RESTful API endpoints
- Webhook ingestion (Stripe, Xero)
- Payment failure processing
- Workflow orchestration
- Health checks and metrics

#### Worker Service
- Background job processing
- Smart retry logic with cost optimization
- BECS rail routing for micro-transactions
- Cross-method reconciliation
- Analytics processing

#### Web Interface (Port 4896)
- Payment metrics dashboard
- Recovery workflow monitoring
- Configuration management

### Database Schema
- `payment_failures`: Failed payment attempts
- `recovery_attempts`: Retry execution logs
- `analytics`: Failure patterns and trend metrics
- `users`: System user accounts
- `vip_customers`: Priority customer configurations
- `reconciliation_records`: Cross-method payment matching

---

## API Documentation

### Core Endpoints
- `GET /health` - Service health check
- `POST /api/v1/auth/login` - Authentication
- `GET /api/v1/payment-failures` - Payment failure data
- `POST /api/v1/recovery/start` - Manual recovery trigger
- `GET /api/v1/analytics/*` - Analytics and reporting

### Documentation
- **Interactive Docs**: Available at `http://localhost:8080/docs` (development)
- **OpenAPI Spec**: `/api/v1/openapi.json`

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature-name`)
3. Make changes with tests
4. Follow Go best practices (`gofmt`, `go vet`)
5. Submit pull request with description

### Code Standards
- **Go**: Standard formatting, comprehensive tests
- **TypeScript**: Strict mode, type safety
- **Testing**: Maintain 80%+ coverage
- **Documentation**: Update relevant sections

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

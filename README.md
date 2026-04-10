# Payment Watchdog
## 🇦🇺 Australian Payment Recovery Platform

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflows/payment-watchdog-ci.yml/badge/main)](https://github.com/sambitmohanty1/payment-watchdog/actions)
[![Coverage](https://img.shields.io/codecov/c/github/sambitmohanty1/payment-watchdog)](https://codecov.io/gh/sambitmohanty1/payment-watchdog)

---

### 🛡️ Sovereign AU Status: **ACTIVE & INTELLIGENT**

**Phase 1 (Stabilization): COMPLETE** ✅
- Resolved `CrashLoopBackOff` across API/Worker services.
- Finalized explicit K8s environment variable injection.
- Optimized persistent storage for OCI Sydney.

**Phase 2 (Core Logic Activation): COMPLETE** ✅
- **Intelligence Driven**: Rule-based failure classification active in Webhook pipeline.
- **Dynamic Dispatch**: Implemented `ProviderRegistry` for zero-downtime failover to intent-recording.
- **Service Mastery**: Resolved all `nil` dependencies; full business logic is now wired and persistent.
- **Visibility**: Communication tracking and manual retry orchestration fully operational.

---

## � Quick Start

### **Prerequisites**
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### **Development Setup**

```bash
# Clone repository
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

### **Environment Configuration**

| Environment | Docker Compose File | Ports | Database |
|-------------|-------------------|-------|----------|
| Development | `docker-compose.yml` | API:8080, Web:4896, DB:5432, Redis:6379 | `payment_watchdog` |
| Staging | `docker-compose.staging.yml` | API:8091, Web:3011, DB:5443, Redis:6390 | `payment_watchdog_staging` |
| Local | `docker-compose.local.yml` | API:8096, Web:4896, DB:5432, Redis:6379 | `payment_watchdog_local` |

### **🏗️ Building Services**

```bash
# Build API
cd api && go build -o payment-watchdog-api ./cmd/main.go

# Build Worker  
cd worker && go build -o payment-watchdog-worker ./cmd/main.go

# Build Web
cd web && npm run build
```

### **🧪 Testing**

```bash
# Run all tests
go test ./... -v -race -cover

# Run service-specific tests
cd api && go test ./... -v
cd worker && go test ./... -v
```

### **🚀 Deployment**

#### **Development**
```bash
docker-compose up -d
```

#### **Staging**
```bash
docker-compose -f docker-compose.staging.yml up -d
```

#### **Production**
```bash
# Manual deployment via GitHub Actions
# See .github/workflows/ci-cd-pipeline.yml
```

### **�️ Sovereign Compliance**

- Australian data residency enforced
- Network isolation policies
- PCI-DSS compliance
- Secure secret management

---

## 📚 Documentation Structure

### **📋 Strategic Documents**
- **[📊 FUTURE_STATE.md](FUTURE_STATE.md)** - Strategic roadmap and market analysis
- **[🎯 FEATURE_BACKLOG.md](FEATURE_BACKLOG.md)** - Product features and user stories
- **[🏢 PLATFORM_BACKLOG.md](PLATFORM_BACKLOG.md)** - Technical debt and platform improvements

### **🏗️ Architecture Documentation**
- **System Design**: [SYSTEM_DESIGN.md](docs/ARCHITECTURE/SYSTEM_DESIGN.md)
- **Low-Level Design**: [LOW_LEVEL_DESIGN.md](docs/ARCHITECTURE/LOW_LEVEL_DESIGN.md)
- **API Documentation**: [API.md](docs/ARCHITECTURE/API_SPECIFICATION.md)
- **Worker Documentation**: [WORKER.md](docs/WORKER.md)
- **[📐 SYSTEM_DESIGN.md](docs/ARCHITECTURE/SYSTEM_DESIGN.md)** - High-level system architecture
- **[🔌 API_SPECIFICATION.md](docs/ARCHITECTURE/API_SPECIFICATION.md)** - Complete API documentation
- **[📋 BUSINESS_REQUIREMENTS.md](docs/STRATEGIC/BUSINESS_REQUIREMENTS.md)** - Business requirements and user stories

### **🔧 Operations Documentation**
- **[🔒 SECURITY.md](SECURITY.md)** - Security policies and compliance
- **[📝 LOGGING_GUIDELINES.md](docs/OPERATIONS/LOGGING_GUIDELINES.md)** - Logging standards and best practices
- **[📄 ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)** - Configuration reference

### **📚 Reference Documentation**
- **[💱 CURRENCY_GUIDELINES.md](api/CURRENCY_GUIDELINES.md)** - Currency field usage guidelines

### **🗃️ Archived Documentation**
- **[📁 ARCHIVED/TECH_DEBT_BACKLOG_OLD.md](docs/ARCHIVED/TECH_DEBT_BACKLOG_OLD.md)** - Previous technical debt backlog
- **[📁 ARCHIVED/LOGGING_IMPROVEMENTS_OLD.md](docs/ARCHIVED/LOGGING_IMPROVEMENTS_OLD.md)** - Previous logging improvements

### 🏗️ Architecture Overview

### **📐 Technology Stack**
- **Backend**: Go 1.24, Gin, GORM, PostgreSQL, Redis
- **Frontend**: Next.js 14, TypeScript, Tailwind CSS
- **Infrastructure**: Docker, Kubernetes, AWS
- **Payment Rails**: PayTo, NPP, BECS, Stripe, PayPal
- **Monitoring**: Prometheus, Grafana, ELK Stack

### **🚀 Microservices Architecture**
```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Next.js Dashboard]
        API_CLIENTS[API Clients]
    end
    
    subgraph "API Gateway"
        GATEWAY[API Gateway]
        LB[Load Balancer]
    end
    
    subgraph "Application Services"
        API[Payment API Service]
        WORKER[Background Worker Service]
        WEBHOOK[Webhook Service]
        ANALYTICS[Analytics Service]
    end
    
    subgraph "Data Layer"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
        TIMESERIES[(Time Series)]
    end
    
    subgraph "External Services"
        STRIPE[Stripe API]
        PAYTO[PayTo/NPP API]
        XERO[Xero API]
        BNPL[BNPL Providers]
    end
    
    WEB --> GATEWAY
    API_CLIENTS --> GATEWAY
    GATEWAY --> API
    GATEWAY --> WEBHOOK
    
    API --> POSTGRES
    API --> REDIS
    WORKER --> POSTGRES
    WORKER --> REDIS
    
    API --> STRIPE
    WORKER --> PAYTO
    WORKER --> XERO
```

### **🔒 Security Architecture**
- **Authentication**: JWT-based authentication with MFA
- **Authorization**: Role-based access control (RBAC)
- **Data Protection**: AES-256 encryption, TLS 1.3
- **Compliance**: Australian data residency, PCI-DSS Level 1
- **Monitoring**: Security scanning, vulnerability management

---

## 🚀 Deployment

### **🏗️ Production Deployment**
Payment Watchdog is deployed using Kubernetes with automated CI/CD pipelines.

#### **📋 Deployment Environments**
- **Development**: `docker-compose.yml` (localhost)
- **Staging**: `docker-compose.staging.yml` (staging servers)
- **Production**: Kubernetes clusters (AWS/Azure)

#### **🔄 CI/CD Pipeline**
- **Build**: Automated build and test execution
- **Security**: Security scanning and vulnerability assessment
- **Deploy**: Automated deployment to staging and production
- **Monitor**: Health checks and rollback capabilities

#### **🔧 Deployment Commands**
```bash
# Build Docker images
make build-prod

# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-production

# Check deployment status
make deploy-status
```

### **📊 Monitoring and Observability**
- **Health Checks**: `/health` and `/health/detailed` endpoints
- **Metrics**: Prometheus metrics collection
- **Logging**: Structured logging with ELK stack
- **Alerting**: Grafana dashboards and alerting

---

## 📊 API Documentation

### **🔌 API Reference**
Complete API documentation is available at:
- **Interactive**: [Swagger UI](http://localhost:8080/docs) (development)
- **Specification**: [API Specification](docs/ARCHITECTURE/API_SPECIFICATION.md)
- **OpenAPI**: [OpenAPI JSON](http://localhost:8080/api/v1/openapi.json)

### **📋 Key Endpoints**
- **Health**: `GET /health` - Basic health check
- **Authentication**: `POST /api/v1/auth/login` - User authentication
- **Payment Failures**: `GET /api/v1/payment-failures` - Payment failure data
- **Workflows**: `GET /api/v1/workflows` - Recovery workflows
- **Analytics**: `GET /api/v1/analytics/*` - Analytics and reporting

---

## 🤝 Contributing

### **📋 Development Workflow**
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit changes (`git commit -m "Add amazing feature"`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Create Pull Request
8. Wait for review and merge

### **📝 Code Standards**
- **Go**: Follow Go best practices and `gofmt` formatting
- **TypeScript**: Use TypeScript with strict mode
- **Testing**: Maintain 90%+ test coverage
- **Documentation**: Update relevant documentation

### **🧪 Testing**
```bash
# Run all tests
go test ./... -v -race -cover

# Run specific service tests
go test ./services -v

# Run integration tests
go test ./integration -v

# Run benchmarks
go test -bench=.
```

---

## 📞 Support

### **🆘 Getting Help**
- **Documentation**: [docs/](https://github.com/sambitmohanty1/payment-watchdog/docs)
- **Issues**: [GitHub Issues](https://github.com/sambitmohanty1/payment-watchdog/issues)
- **Discussions**: [GitHub Discussions](https://github.com/sambitmohanty1/payment-watchdog/discussions)
- **Email**: support@payment-watchdog.com.au

### **📞 Community**
- **Slack**: [Join our Slack community](https://payment-watchdog.slack.com)
- **Twitter**: [@paymentwatchdog](https://twitter.com/paymentwatchdog)
- **LinkedIn**: [Payment Watchdog](https://linkedin.com/company/payment-watchdog)

### **📚 Resources**
- **Website**: [https://payment-watchdog.com.au](https://payment-watchdog.com.au)
- **Blog**: [Blog posts and tutorials](https://blog.payment-watchdog.com.au)
- **Status**: [System status page](https://status.payment-watchdog.com.au)
- **Changelog**: [Release notes and updates](https://github.com/sambitmohanty1/payment-watchdog/releases)

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24+
- Node.js 18+

### Local Development
```bash
# Clone the repository
git clone https://github.com/sambitmohanty1/payment-watchdog.git
cd payment-watchdog

# Start all services
docker-compose up -d

# API will be available at http://localhost:8080
# Web interface at http://localhost:4896
# Database at localhost:5432
# Redis at localhost:6379
```

### 🚀 Staging Environment
```bash
# Start staging environment
docker-compose -f docker-compose.staging.yml up -d

# API will be available at http://localhost:8091
# Web interface at http://localhost:3011
# Database at localhost:5443
# Redis at localhost:6390
```

### 🏢 Local Deployment (Isolated)
```bash
# Start isolated local deployment
make local-start

# Access services
# Web Interface: http://localhost:3016
# API Endpoint: http://localhost:8096
# MailHog: http://localhost:8041
```

### Development Commands
```bash
# Code formatting
./format.sh

# Sovereignty audit
./scripts/AUDIT_SOVEREIGNTY.sh

# API Service
cd api && go run cmd/main.go

# Worker Service
cd worker && go run cmd/main.go

# Web Interface
cd web && npm run dev
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific service tests
go test ./services -v

# Test with coverage
go test ./... -cover
```

---

## � Development Environment

### **🔧 Development Tools**
- **IDE**: VS Code, GoLand, WebStorm
- **Database**: PostgreSQL 15+, Redis 7+
- **Container**: Docker Desktop, Docker Compose
- **Version Control**: Git, GitHub
- **CI/CD**: GitHub Actions

### **📋 Environment Variables**
Refer to [📄 ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) for complete configuration reference.

### **🔍 Debugging**
```bash
# View logs
docker-compose logs -f payment-api

# Check service status
docker-compose ps

# Access database shell
docker-compose exec postgres psql -U postgres -d payment_watchdog

# Access Redis CLI
docker-compose exec redis redis-cli
```

---

## 🚀 Deployment

For isolated local testing and development, separate from CI/CD flows:

### Quick Start
```bash
# Start isolated local deployment
make local-start
# Or
./scripts/local-deploy.sh start

# Access services
# Web Interface: http://localhost:3016
# API Endpoint: http://localhost:8096
# MailHog: http://localhost:8041
```

### Management Commands
```bash
# Stop services
make local-stop

# Restart services  
make local-restart

# Check status
make local-status

# View logs
make local-logs

# Run health checks
make local-health

# Clean all data (WARNING: destroys database)
make local-clean
```

### Port Configuration
| Service | Development | Local |
|---------|-------------|-------|
| API | 8080 | 8096 |
| Web | 4896 | 3016 |
| PostgreSQL | 5432 | 5448 |
| Redis | 6379 | 6395 |

---

## AU/NZ Rail Orchestrator

### PW-101: PayTo Failover Mediator
**Technical Implementation**: `PayToExecutor` in `step_executors.go`
- **Trigger**: Payment failures with reason `insufficient_funds` or `card_declined`
- **Action**: Submits PayTo agreement request job via retry service
- **Flow**: 
  1. Validate failure reason
  2. Create recovery action record with type `payto_agreement_requested`
  3. Submit job to retry service with PayTo provider
  4. Track execution via workflow context

### PW-102: Cross-Method Xero Reconciliation  
**Technical Implementation**: `HandleCrossMethodReconciliation` in `recovery_orchestration_service.go`
- **Trigger**: Manual bank transfer detection in Xero
- **Action**: Finds matching payment failures and resolves them
- **Flow**:
  1. Query payment failures by invoice ID, amount, and company
  2. Update status to `resolved` with reason `cross_method_reconciliation`
  3. Cancel all active workflow executions for the payment failure
  4. Log reconciliation reference

### PW-103: Micro-Transaction Cost Logic
**Technical Implementation**: Cost-aware routing in `PaymentRetryExecutor`
- **Trigger**: Transactions < 10,000 cents ($100) on high-fee providers
- **Action**: Switch provider to `becs` before retry execution
- **Flow**:
  1. Check if `amount_cents < 10000` and provider is `stripe` or `international_card`
  2. Override `config.Provider = "becs"`
  3. Execute retry with local BECS rails
  4. Log cost optimization event

---

## �️ Sovereign Data Compliance

Payment Watchdog supports "Sovereign Mode" to ensure all application data and telemetry remain strictly within isolated Australian cloud regions (AWS, GCP, OCI, Azure) to comply with data residency laws.

### Enabling Sovereign Mode
To enable, set `SOVEREIGN_MODE=true` as an environment variable in both the `api` and `worker` services. The application will actively validate its database connection string upon startup and **abort** if a non-AU endpoint is detected.

### Sovereign Kubernetes Deployment
A Kustomize overlay is provided to enforce "Air-Gapped" network policies and use local cluster resources (like Prometheus and Redis) instead of global ones.

```bash
# Apply Sovereign Kubernetes configuration
kubectl apply -k api/deployments/kubernetes/overlays/sovereign-au
```

### Infrastructure Auditing
You can generate a Residency Report during deployments or manually run the manifest auditor to ensure no US-based or global external resources are hardcoded:

```bash
# Run the manifest auditor
./scripts/AUDIT_SOVEREIGNTY.sh
```

### GCP Sydney Setup
For sovereign Australian infrastructure deployment:

```bash
# Setup GCP Sydney region infrastructure
./deployment/gcp-sydney-setup.sh
```

---

## API Documentation

### Base URLs
- **Local**: `http://localhost:8080`

### Key Endpoints
- `GET /health` - Service health check
- `GET /api/v1/status` - API status
- `GET /metrics` - Basic metrics

---

## Docker Services

```yaml
services:
  api:           # Go REST API (Port 8080)
  worker:        # Background processing
  web:           # Next.js dashboard (Port 4896)
  postgres:      # Database (Port 5432)
  redis:         # Cache & Queue (Port 6379)
  mailhog:       # Email testing (Port 8025)
```

---

## Deployment

### Local Development
```bash
docker-compose up -d
```

### Production Deployment
```bash
# Build production images
docker build -f api/Dockerfile.production -t payment-watchdog/api ./api
docker build -f worker/Dockerfile.production -t payment-watchdog/worker ./worker
docker build -f web/Dockerfile.production -t payment-watchdog/web ./web

# Deploy to staging
docker-compose -f docker-compose.staging.yml up -d

# Deploy to production (Kubernetes)
kubectl apply -f api/deployments/kubernetes/

# Deploy to sovereign AU environment
kubectl apply -k api/deployments/kubernetes/overlays/sovereign-au
```

---

## Architecture Overview

Payment Watchdog follows a microservices architecture with the following components:

```mermaid
graph TB
    subgraph "Frontend"
        WEB[Next.js Dashboard]
    end
    
    subgraph "Backend Services"
        API[Go API Service]
        WORKER[Background Worker]
    end
    
    subgraph "Data Layer"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
    end
    
    subgraph "External Services"
        MAILHOG[Email Service]
    end
    
    WEB --> API
    API --> POSTGRES
    API --> REDIS
    WORKER --> POSTGRES
    WORKER --> REDIS
    WORKER --> MAILHOG
```

### Service Responsibilities

#### API Service (Port 8080)
- RESTful API endpoints
- Webhook ingestion (Stripe, Xero)
- Payment failure processing
- Advanced workflow orchestration
- VIP customer detection
- PayTo failover execution
- Cross-method reconciliation handling
- Health checks and metrics

#### Worker Service
- Background job processing
- Smart retry logic with cost optimization
- BECS rail routing for micro-transactions
- Cross-method reconciliation
- AI-powered pattern detection
- Advanced analytics processing
- Distributed locking coordination

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

## 📊 Class Diagrams

### Payment Recovery Orchestration Flow

```mermaid
classDiagram
    class RecoveryOrchestrationService {
        -db: *gorm.DB
        -retryService: *RetryService
        -communicationService: *CommunicationService
        -analyticsService: *AnalyticsService
        -stepExecutors: map[string]StepExecutor
        -tracer: trace.Tracer
        -logger: *zap.Logger
        -redisClient: *redis.Client
        -activeExecutions: map[uuid.UUID]*WorkflowExecution
        +ExecuteWorkflow(ctx, workflowID, paymentFailureID) error
        +CancelWorkflow(executionID) error
        +GetExecutionStatus(executionID) (*WorkflowExecution, error)
        +ProcessWorkflowStep(execution *WorkflowExecution) error
    }

    class WorkflowExecution {
        +ID: uuid.UUID
        +WorkflowID: uuid.UUID
        +PaymentFailureID: uuid.UUID
        +CompanyID: uuid.UUID
        +Status: string
        +CurrentStepIndex: int
        +Context: map[string]interface{}
        +StartedAt: time.Time
        +CancelFunc: context.CancelFunc
    }

    class StepExecutor {
        <<interface>>
        +Execute(ctx, step) error
        +Validate(step) error
        +GetEstimatedTime() time.Duration
    }

    class PayToExecutor {
        -service: *RecoveryOrchestrationService
        -tracer: trace.Tracer
        +GetStepType() string
        +Execute(ctx, execution, step) (*StepResult, error)
    }

    class PaymentRetryExecutor {
        -service: *RecoveryOrchestrationService
        -tracer: trace.Tracer
        +GetStepType() string
        +Execute(ctx, execution, step) (*StepResult, error)
    }

    class PaymentFailureService {
        -db: *gorm.DB
        -logger: *zap.Logger
        +GetPaymentFailures(ctx, companyID, filters, page, limit) ([]PaymentFailureEvent, int64, error)
        +CreatePaymentFailure(ctx, failure) error
        +UpdatePaymentFailure(ctx, id, updates) error
        +GetFailureByID(ctx, id) (*PaymentFailureEvent, error)
    }

    RecoveryOrchestrationService --> WorkflowExecution : manages
    RecoveryOrchestrationService --> StepExecutor : uses
    RecoveryOrchestrationService --> PaymentFailureService : queries
    RecoveryOrchestrationService --> PayToExecutor : registers
    RecoveryOrchestrationService --> PaymentRetryExecutor : registers
    PayToExecutor --|> StepExecutor : implements
    PaymentRetryExecutor --|> StepExecutor : implements
```

### Webhook Processing Flow

```mermaid
classDiagram
    class WebhookProcessor {
        -MaxRetries: int
        -RetryDelay: time.Duration
        -DeadLetterQueue: chan WebhookEvent
        -RateLimiter: *rate.Limiter
        +ProcessWebhook(ctx, event) error
        +ValidateWebhook(event) error
        +RetryFailedEvent(event) error
        +SendToDeadLetterQueue(event) error
    }

    class WebhookEvent {
        +CompanyID: string
        +Event: *stripe.Event
        +RawBody: []byte
        +Headers: http.Header
        +Timestamp: time.Time
    }

    class WebhookError {
        +Type: string
        +Severity: string
        +Message: string
        +Retryable: bool
        +CompanyID: string
        +EventID: string
        +Timestamp: time.Time
        +RetryCount: int
    }

    class WebhookMetrics {
        +ProcessedCount: int64
        +FailedCount: int64
        +RetryCount: int64
        +AverageProcessingTime: time.Duration
        +LastProcessedAt: time.Time
    }

    WebhookProcessor --> WebhookEvent : processes
    WebhookProcessor --> WebhookError : generates
    WebhookProcessor --> WebhookMetrics : tracks
```

### Analytics Engine Flow

```mermaid
classDiagram
    class AnalyticsService {
        -db: *gorm.DB
        -analyticsEngine: *AnalyticsEngine
        -patternDetector: PatternDetector
        -trendAnalyzer: TrendAnalyzer
        -failurePredictor: FailurePredictor
        -logger: *zap.Logger
        +GenerateFailureReport(ctx, companyID, timeRange) (*FailureReport, error)
        +DetectPatterns(ctx, companyID) ([]Pattern, error)
        +PredictFailures(ctx, companyID) (*FailurePrediction, error)
        +AnalyzeTrends(ctx, companyID, period) (*TrendAnalysis, error)
    }

    class AnalyticsEngine {
        -patternDetector: PatternDetector
        -trendAnalyzer: TrendAnalyzer
        -failurePredictor: FailurePredictor
        -logger: *zap.Logger
        +ProcessAnalytics(ctx, request) (*AnalyticsResult, error)
        +AggregateMetrics(ctx, metrics) (*AggregatedMetrics, error)
        +GenerateInsights(ctx, data) ([]Insight, error)
    }

    class PatternDetector {
        <<interface>>
        +DetectPatterns(ctx, data) ([]Pattern, error)
        +ValidatePattern(pattern) error
        +GetPatternConfidence(pattern) float64
    }

    class TrendAnalyzer {
        <<interface>>
        +AnalyzeTrend(ctx, data, period) (*Trend, error)
        +PredictNextPeriod(ctx, trend) (*Prediction, error)
        +CalculateTrendStrength(trend) float64
    }

    class FailurePredictor {
        <<interface>>
        +PredictFailure(ctx, customer, history) (*FailureRisk, error)
        +GetRiskFactors(ctx, customer) ([]RiskFactor, error)
        +UpdateModel(ctx, trainingData) error
    }

    AnalyticsService --> AnalyticsEngine : uses
    AnalyticsEngine --> PatternDetector : delegates
    AnalyticsEngine --> TrendAnalyzer : delegates
    AnalyticsEngine --> FailurePredictor : delegates
```

### Distributed Locking Flow

```mermaid
classDiagram
    class DistributedLockService {
        -redisClient: *redis.Client
        -logger: *zap.Logger
        -prefix: string
        -defaultTTL: time.Duration
        -retryDelay: time.Duration
        -maxRetries: int
        +AcquireLock(ctx, resourceKey) (*Lock, error)
        +ReleaseLock(lock) error
        +ExtendLock(ctx, lock, duration) error
        +IsLocked(ctx, resourceKey) (bool, error)
    }

    class Lock {
        -key: string
        -value: string
        -acquiredAt: time.Time
        -ttl: time.Duration
        -service: *DistributedLockService
        +Extend(ctx, duration) error
        +Release() error
        +IsValid() bool
        +GetRemainingTTL() time.Duration
    }

    class LockManager {
        -lockService: *DistributedLockService
        -activeLocks: map[string]*Lock
        -mu: sync.RWMutex
        +AcquireResourceLock(ctx, resource) (*Lock, error)
        +ReleaseResourceLock(resource) error
        +GetActiveLocks() map[string]*Lock
        +CleanupExpiredLocks() error
    }

    DistributedLockService --> Lock : creates
    DistributedLockService --> Lock : manages
    LockManager --> DistributedLockService : uses
    LockManager --> Lock : tracks
```

### Xero Mediator Flow

```mermaid
classDiagram
    class XeroMediator {
        -oauthClient: *http.Client
        -apiClient: *XeroAPIClient
        -oauthConfig: *OAuthConfig
        +Authenticate(ctx) (*OAuthTokens, error)
        +GetInvoice(ctx, invoiceID) (*XeroInvoice, error)
        +CreatePayment(ctx, payment) (*XeroPayment, error)
        +GetBankTransactions(ctx, since) ([]XeroBankTransaction, error)
        +ReconcilePayment(ctx, invoiceID, paymentID) error
    }

    class XeroAPIClient {
        -httpClient: *http.Client
        -baseURL: string
        -logger: *zap.Logger
        +Get(ctx, endpoint) (*http.Response, error)
        +Post(ctx, endpoint, body) (*http.Response, error)
        +Put(ctx, endpoint, body) (*http.Response, error)
        +RefreshToken(ctx) error
    }

    class OAuthTokens {
        +AccessToken: string
        +RefreshToken: string
        +TokenType: string
        +ExpiresIn: int64
        +Scope: string
        +ExpiresAt: time.Time
        +IsExpired() bool
        +RefreshIfNeeded(ctx) error
    }

    class XeroInvoice {
        +ID: string
        +InvoiceNumber: string
        +Contact: XeroContact
        +LineItems: []XeroLineItem
        +Status: string
        +AmountDue: float64
        +Date: time.Time
    }

    class XeroBankTransaction {
        +ID: string
        +Type: string
        +Contact: XeroContact
        +LineItems: []XeroLineItem
        +Amount: float64
        +Date: time.Time
        +Reference: string
    }

    XeroMediator --> XeroAPIClient : uses
    XeroMediator --> OAuthTokens : manages
    XeroAPIClient --> OAuthTokens : authenticates
    XeroMediator --> XeroInvoice : retrieves
    XeroMediator --> XeroBankTransaction : retrieves
```

### Service Integration Flow

```mermaid
classDiagram
    class APIGateway {
        -recoveryService: *RecoveryOrchestrationService
        -failureService: *PaymentFailureService
        -analyticsService: *AnalyticsService
        -webhookProcessor: *WebhookProcessor
        +HandleWebhook(ctx, request) error
        +GetRecoveryStatus(ctx, id) (*RecoveryStatus, error)
        +GetAnalytics(ctx, companyID) (*AnalyticsData, error)
    }

    class WorkerService {
        -orchestrationService: *RecoveryOrchestrationService
        -lockService: *DistributedLockService
        -analyticsService: *AnalyticsService
        -mediator: PaymentProvider
        +ProcessRecoveryWorkflow(ctx, workflow) error
        +ExecuteRetryLogic(ctx, payment) error
        +UpdateAnalytics(ctx, event) error
    }

    class DatabaseLayer {
        -postgres: *gorm.DB
        -redis: *redis.Client
        +SavePaymentFailure(ctx, failure) error
        +GetRecoveryAttempts(ctx, paymentID) ([]RecoveryAttempt, error)
        +CacheAnalytics(ctx, data) error
        +GetCachedData(ctx, key) (interface{}, error)
    }

    APIGateway --> RecoveryOrchestrationService : delegates
    APIGateway --> PaymentFailureService : queries
    APIGateway --> AnalyticsService : requests
    APIGateway --> WebhookProcessor : processes
    WorkerService --> RecoveryOrchestrationService : executes
    WorkerService --> DistributedLockService : coordinates
    WorkerService --> AnalyticsService : updates
    RecoveryOrchestrationService --> DatabaseLayer : persists
    AnalyticsService --> DatabaseLayer : queries
    DistributedLockService --> DatabaseLayer : locks
```

---

## 🔄 Sequence Diagrams

### Payment Failure Recovery Workflow

```mermaid
sequenceDiagram
    participant Stripe as Stripe Webhook
    participant API as API Gateway
    participant WP as WebhookProcessor
    participant PFS as PaymentFailureService
    participant ROS as RecoveryOrchestrationService
    participant DLS as DistributedLockService
    participant Worker as WorkerService
    participant Xero as XeroMediator
    participant DB as Database
    participant Analytics as AnalyticsService

    Stripe->>API: POST /webhook/stripe
    API->>WP: ProcessWebhook(event)
    WP->>WP: ValidateWebhook()
    WP->>PFS: CreatePaymentFailure(failure)
    PFS->>DB: INSERT payment_failures
    PFS-->>WP: PaymentFailure created
    WP->>ROS: ExecuteWorkflow(workflowID, failureID)
    ROS->>DLS: AcquireLock("recovery_" + failureID)
    DLS->>DB: SET redis lock
    DLS-->>ROS: Lock acquired
    ROS->>Worker: ProcessRecoveryWorkflow(workflow)
    
    Worker->>Xero: GetBankTransactions()
    Xero->>Xero: Authenticate()
    Xero-->>Worker: BankTransactions
    Worker->>Worker: ReconcilePayment()
    
    alt Payment Found
        Worker->>PFS: UpdatePaymentFailure(status: "recovered")
        PFS->>DB: UPDATE payment_failures
        Worker->>Analytics: UpdateAnalytics(recovery)
        Analytics->>DB: INSERT analytics
    else No Payment Found
        Worker->>ROS: ExecuteRetryLogic()
        ROS->>ROS: ScheduleRetry()
    end
    
    Worker-->>ROS: Workflow completed
    ROS->>DLS: ReleaseLock()
    DLS->>DB: DELETE redis lock
    ROS-->>API: Recovery status
    API-->>Stripe: 200 OK
```

### Analytics Processing Flow

```mermaid
sequenceDiagram
    participant UI as Web Interface
    participant API as API Gateway
    participant AS as AnalyticsService
    participant AE as AnalyticsEngine
    participant PD as PatternDetector
    participant TA as TrendAnalyzer
    participant FP as FailurePredictor
    participant DB as Database

    UI->>API: GET /analytics/failure-report
    API->>AS: GenerateFailureReport(companyID, timeRange)
    AS->>DB: Query payment_failures
    DB-->>AS: Failure data
    AS->>AE: ProcessAnalytics(request)
    
    par Pattern Detection
        AE->>PD: DetectPatterns(data)
        PD->>PD: AnalyzeFailurePatterns()
        PD-->>AE: Patterns found
    and Trend Analysis
        AE->>TA: AnalyzeTrend(data, period)
        TA->>TA: CalculateTrends()
        TA-->>AE: Trend analysis
    and Failure Prediction
        AE->>FP: PredictFailures(data)
        FP->>FP: ApplyMLModel()
        FP-->>AE: Risk predictions
    end
    
    AE-->>AS: Aggregated insights
    AS->>DB: Cache analytics results
    AS-->>API: FailureReport
    API-->>UI: JSON response
```

### Distributed Locking Coordination

```mermaid
sequenceDiagram
    participant W1 as Worker 1
    participant W2 as Worker 2
    participant DLS as DistributedLockService
    participant Redis as Redis
    participant DB as PostgreSQL

    W1->>DLS: AcquireLock("payment_123")
    DLS->>Redis: SETNX lock_key_123
    Redis-->>DLS: 1 (success)
    DLS->>Redis: EXPIRE lock_key_123 30
    DLS-->>W1: Lock acquired
    
    Note over W1,W2: Concurrent processing attempt
    W2->>DLS: AcquireLock("payment_123")
    DLS->>Redis: SETNX lock_key_123
    Redis-->>DLS: 0 (already locked)
    DLS->>W2: Lock denied
    
    W1->>DB: BEGIN TRANSACTION
    W1->>DB: UPDATE payment_failures
    W1->>DB: COMMIT
    W1->>DLS: ReleaseLock(lock)
    DLS->>Redis: DELETE lock_key_123
    DLS-->>W1: Lock released
    
    W2->>DLS: AcquireLock("payment_123")
    DLS->>Redis: SETNX lock_key_123
    Redis-->>DLS: 1 (success)
    DLS-->>W2: Lock acquired
    W2->>DB: Process payment
```

### Xero Integration Flow

```mermaid
sequenceDiagram
    participant Worker as WorkerService
    participant XM as XeroMediator
    participant XAC as XeroAPIClient
    participant OAuth as OAuthTokens
    participant Xero as Xero API
    participant DB as Database

    Worker->>XM: ReconcilePayment(invoiceID, paymentID)
    XM->>OAuth: IsExpired()
    alt Token Expired
        XM->>XAC: RefreshToken()
        XAC->>Xero: POST /oauth2/token
        Xero-->>XAC: New tokens
        XAC-->>OAuth: Updated tokens
    end
    
    XM->>XAC: GetInvoice(invoiceID)
    XAC->>OAuth: GetAccessToken()
    OAuth-->>XAC: Bearer token
    XAC->>Xero: GET /invoices/{id}
    Xero-->>XAC: Invoice data
    XAC-->>XM: XeroInvoice
    
    XM->>XAC: GetBankTransactions(since)
    XAC->>Xero: GET /banktransactions
    Xero-->>XAC: Transactions
    XAC-->>XM: BankTransactions
    
    XM->>XM: MatchPaymentToInvoice()
    alt Match Found
        XM->>XAC: CreatePayment(payment)
        XAC->>Xero: POST /payments
        Xero-->>XAC: Payment created
        XM->>DB: Update reconciliation_records
    else No Match
        XM->>DB: Log reconciliation failure
    end
    
    XM-->>Worker: Reconciliation result
```

### Webhook Processing with Retry Logic

```mermaid
sequenceDiagram
    participant Stripe as Stripe
    participant API as API Gateway
    participant WP as WebhookProcessor
    participant DLQ as DeadLetterQueue
    participant RateLimit as RateLimiter
    participant DB as Database

    Stripe->>API: POST /webhook/stripe
    API->>RateLimit: AllowRequest()
    RateLimit-->>API: Allowed
    API->>WP: ProcessWebhook(event)
    
    WP->>WP: ValidateSignature()
    alt Invalid Signature
        WP->>API: 401 Unauthorized
    else Valid Signature
        WP->>WP: ParseEvent()
        
        alt Processing Fails
            WP->>WP: ShouldRetry()?
            alt Retryable and MaxRetries not reached
                WP->>WP: IncrementRetryCount()
                WP->>WP: ScheduleRetry()
                Note over WP: Exponential backoff
                WP->>WP: ProcessWebhook(event)
            else MaxRetries reached
                WP->>DLQ: SendToDeadLetterQueue(event)
                DLQ->>DB: INSERT dead_letter_events
                WP->>API: 500 Server Error
            end
        else Processing Succeeds
            WP->>DB: SaveProcessedEvent()
            WP->>API: 200 OK
        end
    end
    
    Note over Stripe,API: Async retry processing
    loop Retry scheduled events
        WP->>WP: ProcessRetryEvent()
        WP->>WP: ProcessWebhook(retry_event)
    end
```

### PayTo Failover Sequence (PW-101)

```mermaid
sequenceDiagram
    participant Stripe as Stripe Webhook
    participant API as API Gateway
    participant ROS as RecoveryOrchestrationService
    participant PayTo as PayToExecutor
    participant Retry as RetryService
    participant DB as Database

    Stripe->>API: POST /webhook/stripe (insufficient_funds)
    API->>ROS: ExecuteWorkflow(workflowID, failureID)
    ROS->>ROS: CreateWorkflowExecution()
    ROS->>PayTo: Execute("payto_agreement")
    
    PayTo->>PayTo: ValidateFailureReason()
    alt failure_reason is insufficient_funds or card_declined
        PayTo->>PayTo: CreateRecoveryAction(type: "payto_agreement_requested")
        PayTo->>DB: INSERT recovery_actions
        PayTo->>Retry: SubmitJob("payto_agreement_request", jobData)
        Retry-->>PayTo: Job ID
        PayTo-->>ROS: StepResult(Success: true, ExternalID: jobID)
    else other failure reason
        PayTo-->>ROS: StepResult(Success: true, Skipped: true)
    end
    
    ROS->>DB: UPDATE workflow_executions
    ROS-->>API: Workflow status
```

### Micro-Transaction Cost Optimization (PW-103)

```mermaid
sequenceDiagram
    participant Worker as WorkerService
    participant Retry as PaymentRetryExecutor
    participant DB as Database

    Worker->>Retry: Execute(paymentFailure, retryConfig)
    
    Retry->>Retry: CheckAmountAndProvider()
    alt amount_cents < 10000 AND provider in ["stripe", "international_card"]
        Retry->>Retry: ApplyCostLogic()
        Note over Retry: Override config.Provider = "becs"
        Retry->>DB: Log cost optimization event
        Retry->>Retry: ExecuteWithBECS()
    else amount >= 10000 OR provider is low-cost
        Retry->>Retry: ExecuteStandardRetry()
    end
    
    Retry-->>Worker: StepResult with provider used
```

### Multi-Service Recovery Orchestration

```mermaid
sequenceDiagram
    participant Web as Web Interface
    participant API as API Gateway
    participant ROS as RecoveryOrchestrationService
    participant Lock as DistributedLockService
    participant Worker1 as Worker Service 1
    participant Worker2 as Worker Service 2
    participant Analytics as AnalyticsService
    participant Alert as AlertService
    participant Xero as XeroMediator
    participant DB as Database

    Web->>API: POST /api/v1/recovery/start
    API->>ROS: ExecuteWorkflow(workflowID, paymentID)
    ROS->>Lock: AcquireLock("workflow_" + workflowID)
    Lock-->>ROS: Lock acquired
    
    ROS->>ROS: CreateWorkflowExecution()
    ROS->>DB: INSERT workflow_executions
    
    par Step 1: Payment Analysis
        ROS->>Worker1: ExecuteStep("analyze_payment")
        Worker1->>DB: Query payment history
        Worker1-->>ROS: Analysis result
    and Step 2: Cross-Method Reconciliation
        ROS->>Worker2: ExecuteStep("reconcile_payment")
        Worker2->>Xero: GetBankTransactions()
        Xero-->>Worker2: Transaction data
        Worker2-->>ROS: Reconciliation result
    end
    
    ROS->>Analytics: UpdateWorkflowMetrics()
    Analytics->>DB: UPDATE analytics
    
    alt Recovery Successful
        ROS->>Alert: SendSuccessNotification()
        Alert-->>ROS: Notification sent
        ROS->>DB: UPDATE workflow_executions (completed)
    else Recovery Failed
        ROS->>Alert: SendFailureAlert()
        Alert-->>ROS: Alert sent
        ROS->>ROS: ScheduleNextRetry()
    end
    
    ROS->>Lock: ReleaseLock()
    ROS-->>API: Workflow status
    API-->>Web: JSON response
```

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines
- Follow Go best practices
- Use conventional commits
- Write comprehensive tests
- Update documentation

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

For support, please open an issue on GitHub.
# Trigger worker build

# Payment Watchdog - Solution Design Diagrams

## Database-Agnostic Architecture

### Database Detection and Adaptation Flow

```mermaid
graph TB
    subgraph "Service Layer"
        SVC[Recovery Analytics Service]
        DETECT[Database Detection]
        PGSQL[PostgreSQL Handler]
        SQLT[SQLite Handler]
    end
    
    subgraph "Database Layer"
        PG[(PostgreSQL<br/>Production)]
        SQLITE[(SQLite<br/>In-Memory<br/>Testing)]
    end
    
    subgraph "Query Examples"
        PGQ[EXTRACT(HOUR FROM created_at)<br/>hour::int<br/>recovered::float]
        SQLQ[CAST(strftime('%H', created_at) AS INTEGER)<br/>hour<br/>CAST(recovered AS REAL)]
    end
    
    SVC --> DETECT
    DETECT -->|Production| PGSQL
    DETECT -->|Testing| SQLT
    PGSQL --> PG
    SQLT --> SQLITE
    PGSQL --> PGQ
    SQLT --> SQLQ
    
    style PG fill:#e1f5fe
    style SQLITE fill:#f3e5f5
    style PGQ fill:#e8f5e8
    style SQLQ fill:#fff3e0
```

### Benefits of Database-Agnostic Design

- **Production Ready**: PostgreSQL with advanced features (JSONB, EXTRACT, etc.)
- **Testing Friendly**: SQLite in-memory for fast, ephemeral testing
- **No Race Conditions**: Eliminated sqlmock concurrency issues
- **CI/CD Ready**: Tests run anywhere without external dependencies
- **Flexible**: Easy to add support for other databases

---

## Payment Provider Integration Architecture

### Overall System Integration Diagram

```mermaid
graph TB
    subgraph "External Payment Providers"
        STRIPE[Stripe<br/>Webhooks & API]
        PAYPAL[PayPal<br/>Webhooks & API]
        BRAINTREE[Braintree<br/>Webhooks & API]
        SQUARE[Square<br/>Webhooks & API]
        ADOBE[Adobe Commerce<br/>Webhooks & API]
    end
    
    subgraph "Accounting Systems"
        XERO[Xero<br/>API Integration]
        QUICKBOOKS[QuickBooks<br/>API Integration]
        NETSUITE[NetSuite<br/>API Integration]
        SAGE[Sage<br/>API Integration]
    end
    
    subgraph "Communication Channels"
        EMAIL[Email Services<br/>SendGrid/SES]
        SMS[SMS Services<br/>Twilio/Vonage]
        WEBHOOK[Customer Webhooks<br/>Custom Endpoints]
        SLACK[Slack<br/>Team Notifications]
    end
    
    subgraph "Payment Watchdog Platform"
        subgraph "Ingress Layer"
            LB[Load Balancer<br/>AWS ALB/GCP LB]
            GW[API Gateway<br/>Kong/Nginx]
            WAF[Web Application Firewall<br/>Cloud WAF]
        end
        
        subgraph "Core Services"
            API[API Service<br/>Port 8080]
            RO[Recovery Orchestration<br/>Port 8086]
            WRK[Worker Service<br/>Background Jobs]
            WEB[Web Dashboard<br/>Port 4896]
            AUTH[Auth Service<br/>Port 8081]
        end
        
        subgraph "Data Layer"
            PG[(PostgreSQL<br/>Production DB)]
            SQLITE[(SQLite<br/>Testing DB)]
            RD[(Redis<br/>Cache & Queue)]
            ES[(Elasticsearch<br/>Search & Analytics)]
            S3[(Object Storage<br/>S3/GCS)]
        end
        
        subgraph "Monitoring & Observability"
            PROM[Prometheus<br/>Metrics Collection]
            GRAF[Grafana<br/>Dashboards]
            ELK[ELK Stack<br/>Logging]
            JAEGER[Jaeger<br/>Distributed Tracing]
        end
    end
    
    %% External to Platform Connections
    STRIPE --> WAF
    PAYPAL --> WAF
    BRAINTREE --> WAF
    SQUARE --> WAF
    ADOBE --> WAF
    
    XERO --> GW
    QUICKBOOKS --> GW
    NETSUITE --> GW
    SAGE --> GW
    
    EMAIL --> WRK
    SMS --> WRK
    WEBHOOK --> WRK
    SLACK --> WRK
    
    %% Internal Platform Connections
    WAF --> LB
    LB --> GW
    GW --> API
    GW --> WEB
    GW --> AUTH
    
    API --> PG
    API --> RD
    API --> ES
    
    RO --> PG
    RO --> RD
    RO --> STRIPE
    RO --> PAYPAL
    RO --> BRAINTREE
    RO --> XERO
    RO --> QUICKBOOKS
    
    WRK --> PG
    WRK --> RD
    WRK --> EMAIL
    WRK --> SMS
    WRK --> WEBHOOK
    WRK --> SLACK
    
    WEB --> API
    AUTH --> PG
    
    %% Monitoring Connections
    API --> PROM
    RO --> PROM
    WRK --> PROM
    WEB --> PROM
    
    PROM --> GRAF
    API --> ELK
    RO --> ELK
    WRK --> ELK
    
    API --> JAEGER
    RO --> JAEGER
    WRK --> JAEGER
```

## Complete Payment Lifecycle Flow

### 1. Payment Processing Flow

```mermaid
sequenceDiagram
    participant C as Customer
    participant M as Merchant
    participant PP as Payment Provider
    participant PW as Payment Watchdog
    participant ACC as Accounting System
    participant COM as Communication
    
    Note over C,COM: Initial Payment Attempt
    C->>M: Initiates purchase
    M->>PP: Payment request
    PP->>C: Payment form/redirect
    C->>PP: Submits payment
    PP->>M: Payment success/failure
    
    alt Payment Success
        PP->>PW: payment.succeeded webhook
        PW->>PW: Log transaction
        PW->>ACC: Sync successful payment
        PW->>COM: Send success receipt
        COM->>C: Payment confirmation
    else Payment Failure
        PP->>PW: payment.failed webhook
        PW->>PW: Trigger recovery workflow
        Note over PW: Start Recovery Process
        PW->>PP: Get payment details
        PW->>PW: Analyze failure reason
        PW->>PW: Determine retry strategy
        
        alt Retry Attempt
            PW->>PP: Retry payment
            PP->>PW: Retry result
            
            alt Retry Success
                PW->>PW: Update workflow status
                PW->>ACC: Sync recovered payment
                PW->>COM: Send recovery notification
                COM->>C: Payment recovered notification
            else Retry Failed
                PW->>PW: Schedule next retry
                PW->>PW: Update retry count
                PW->>COM: Send retry attempt notification
            end
        else Manual Intervention Required
            PW->>COM: Alert team
            PW->>WEB: Create manual task
            COM->>M: Manual intervention required
            M->>C: Contact customer support
        end
    end
```

### 2. Recovery Orchestration Workflow

```mermaid
flowchart TD
    START([Payment Failure Detected]) --> ANALYZE{Analyze Failure}
    
    ANALYZE -->|Insufficient Funds| INSUFFICIENT[Insufficient Funds Flow]
    ANALYZE -->|Card Declined| DECLINED[Card Declined Flow]
    ANALYZE -->|Technical Error| TECHNICAL[Technical Error Flow]
    ANALYZE -->|Expired Card| EXPIRED[Expired Card Flow]
    ANALYZE -->|Unknown| UNKNOWN[Unknown Error Flow]
    
    INSUFFICIENT --> WAIT1[Wait 24h]
    WAIT1 --> RETRY1[Retry Payment]
    RETRY1 --> SUCCESS1{Retry Success?}
    SUCCESS1 -->|Yes| SYNC1[Sync to Accounting]
    SUCCESS1 -->|No| WAIT2[Wait 48h]
    WAIT2 --> RETRY2[Retry with Different Method]
    RETRY2 --> SUCCESS2{Retry Success?}
    SUCCESS2 -->|Yes| SYNC1
    SUCCESS2 -->|No| MANUAL1[Manual Intervention]
    
    DECLINED --> CHECK1[Check Bank Authorization]
    CHECK1 --> AUTHORIZED{Bank Authorized?}
    AUTHORIZED -->|Yes| RETRY3[Retry Immediately]
    AUTHORIZED -->|No| NOTIFY1[Notify Customer]
    RETRY3 --> SUCCESS3{Retry Success?}
    SUCCESS3 -->|Yes| SYNC1
    SUCCESS3 -->|No| WAIT3[Wait 1h]
    WAIT3 --> RETRY4[Retry with Different Method]
    RETRY4 --> SUCCESS4{Retry Success?}
    SUCCESS4 -->|Yes| SYNC1
    SUCCESS4 -->|No| MANUAL2[Manual Intervention]
    NOTIFY1 --> WAIT3
    
    TECHNICAL --> LOG[Log Technical Details]
    LOG --> DIAGNOSE[Diagnose Issue]
    DIAGNOSE --> FIXABLE{Can Auto-Fix?}
    FIXABLE -->|Yes| AUTO_FIX[Apply Automatic Fix]
    FIXABLE -->|No| ESCALATE[Escalate to Engineering]
    AUTO_FIX --> RETRY5[Retry Payment]
    RETRY5 --> SUCCESS5{Retry Success?}
    SUCCESS5 -->|Yes| SYNC1
    SUCCESS5 -->|No| WAIT4[Wait 15m]
    WAIT4 --> RETRY6[Retry Again]
    RETRY6 --> SUCCESS6{Retry Success?}
    SUCCESS6 -->|Yes| SYNC1
    SUCCESS6 -->|No| ESCALATE
    
    EXPIRED --> UPDATE[Request Updated Card]
    UPDATE --> PROVIDED{Card Provided?}
    PROVIDED -->|Yes| RETRY7[Retry with New Card]
    PROVIDED -->|No| WAIT5[Wait 48h]
    RETRY7 --> SUCCESS7{Retry Success?}
    SUCCESS7 -->|Yes| SYNC1
    SUCCESS7 -->|No| WAIT5
    WAIT5 --> REMIND[Send Reminder]
    REMIND --> PROVIDED
    
    UNKNOWN --> INVESTIGATE[Investigate Unknown Error]
    INVESTIGATE --> PATTERN{Pattern Recognized?}
    PATTERN -->|Yes| APPLY_PATTERN[Apply Known Solution]
    PATTERN -->|No| MANUAL3[Manual Investigation]
    APPLY_PATTERN --> RETRY8[Retry Payment]
    RETRY8 --> SUCCESS8{Retry Success?}
    SUCCESS8 -->|Yes| SYNC1
    SUCCESS8 -->|No| MANUAL3
    MANUAL3 --> ESCALATE
    
    SYNC1 --> NOTIFY_SUCCESS[Notify Success]
    MANUAL1 --> NOTIFY_MANUAL[Notify Manual Required]
    MANUAL2 --> NOTIFY_MANUAL
    MANUAL3 --> NOTIFY_MANUAL
    ESCALATE --> NOTIFY_MANUAL
    
    NOTIFY_SUCCESS --> END([Workflow Complete])
    NOTIFY_MANUAL --> END
```

### 3. Real-time Data Flow Architecture

```mermaid
graph LR
    subgraph "Payment Provider Events"
        PE1[Stripe Events]
        PE2[PayPal Events]
        PE3[Braintree Events]
        PE4[Square Events]
    end
    
    subgraph "Event Ingestion"
        WAF[Web Application Firewall]
        LB[Load Balancer]
        GW[API Gateway]
        VALIDATOR[Event Validator]
    end
    
    subgraph "Event Processing Pipeline"
        QUEUE1[High Priority Queue]
        QUEUE2[Normal Priority Queue]
        QUEUE3[Low Priority Queue]
        
        WORKER1[Worker 1]
        WORKER2[Worker 2]
        WORKER3[Worker 3]
        WORKER4[Worker 4]
    end
    
    subgraph "Business Logic Layer"
        DETECTOR[Failure Detector]
        ANALYZER[Failure Analyzer]
        ORCHESTRATOR[Recovery Orchestrator]
        PREDICTOR[Predictive Engine]
    end
    
    subgraph "Data Storage"
        CACHE[Redis Cache]
        DB[PostgreSQL]
        SEARCH[Elasticsearch]
        BLOB[Object Storage]
    end
    
    subgraph "Output Systems"
        DASHBOARD[Real-time Dashboard]
        ALERTS[Alert System]
        REPORTS[Analytics Reports]
        API_RESPONSE[API Responses]
    end
    
    PE1 --> WAF
    PE2 --> WAF
    PE3 --> WAF
    PE4 --> WAF
    
    WAF --> LB
    LB --> GW
    GW --> VALIDATOR
    
    VALIDATOR -->|Critical| QUEUE1
    VALIDATOR -->|High| QUEUE2
    VALIDATOR -->|Low| QUEUE3
    
    QUEUE1 --> WORKER1
    QUEUE2 --> WORKER2
    QUEUE2 --> WORKER3
    QUEUE3 --> WORKER4
    
    WORKER1 --> DETECTOR
    WORKER2 --> DETECTOR
    WORKER3 --> DETECTOR
    WORKER4 --> DETECTOR
    
    DETECTOR --> ANALYZER
    ANALYZER --> ORCHESTRATOR
    ANALYZER --> PREDICTOR
    
    ORCHESTRATOR --> CACHE
    ORCHESTRATOR --> DB
    PREDICTOR --> SEARCH
    
    CACHE --> DASHBOARD
    DB --> DASHBOARD
    SEARCH --> REPORTS
    
    ORCHESTRATOR --> ALERTS
    DETECTOR --> API_RESPONSE
```

### 4. Multi-Provider Integration Pattern

```mermaid
graph TB
    subgraph "Payment Provider Adapters"
        subgraph "Stripe Adapter"
            STRIPE_WEBHOOK[Webhook Handler]
            STRIPE_API[API Client]
            STRIPE_PARSER[Event Parser]
            STRIPE_FORMAT[Data Formatter]
        end
        
        subgraph "PayPal Adapter"
            PAYPAL_WEBHOOK[Webhook Handler]
            PAYPAL_API[API Client]
            PAYPAL_PARSER[Event Parser]
            PAYPAL_FORMAT[Data Formatter]
        end
        
        subgraph "Braintree Adapter"
            BRAINTREE_WEBHOOK[Webhook Handler]
            BRAINTREE_API[API Client]
            BRAINTREE_PARSER[Event Parser]
            BRAINTREE_FORMAT[Data Formatter]
        end
    end
    
    subgraph "Unified Payment Interface"
        UNIFIED_WEBHOOK[Unified Webhook Interface]
        UNIFIED_API[Unified API Interface]
        UNIFIED_EVENTS[Standardized Event Model]
        UNIFIED_RETRY[Standardized Retry Logic]
    end
    
    subgraph "Core Business Logic"
        WORKFLOW[Workflow Engine]
        RULES[Business Rules Engine]
        DECISION[Decision Matrix]
        SCHEDULER[Retry Scheduler]
    end
    
    %% Adapter to Unified Interface
    STRIPE_WEBHOOK --> UNIFIED_WEBHOOK
    PAYPAL_WEBHOOK --> UNIFIED_WEBHOOK
    BRAINTREE_WEBHOOK --> UNIFIED_WEBHOOK
    
    STRIPE_PARSER --> STRIPE_FORMAT
    PAYPAL_PARSER --> PAYPAL_FORMAT
    BRAINTREE_PARSER --> BRAINTREE_FORMAT
    
    STRIPE_FORMAT --> UNIFIED_EVENTS
    PAYPAL_FORMAT --> UNIFIED_EVENTS
    BRAINTREE_FORMAT --> UNIFIED_EVENTS
    
    UNIFIED_API --> STRIPE_API
    UNIFIED_API --> PAYPAL_API
    UNIFIED_API --> BRAINTREE_API
    
    %% Unified Interface to Business Logic
    UNIFIED_EVENTS --> WORKFLOW
    UNIFIED_RETRY --> RULES
    RULES --> DECISION
    DECISION --> SCHEDULER
    
    SCHEDULER --> UNIFIED_API
```

### 5. Error Handling and Recovery Patterns

```mermaid
stateDiagram-v2
    [*] --> PaymentReceived
    
    PaymentReceived --> ValidateEvent
    ValidateEvent --> EventValid: Valid
    ValidateEvent --> EventInvalid: Invalid
    
    EventInvalid --> LogError
    LogError --> [*]
    
    EventValid --> DetermineFailureType
    
    DetermineFailureType --> InsufficientFunds: Insufficient Funds
    DetermineFailureType --> CardDeclined: Card Declined
    DetermineFailureType --> TechnicalError: Technical Error
    DetermineFailureType --> ExpiredCard: Expired Card
    DetermineFailureType --> UnknownError: Unknown
    
    InsufficientFunds --> ScheduleRetry
    CardDeclined --> CheckBankAuth
    TechnicalError --> DiagnoseIssue
    ExpiredCard --> RequestNewCard
    UnknownError --> InvestigatePattern
    
    CheckBankAuth --> BankAuthorized: Authorized
    CheckBankAuth --> BankNotAuthorized: Not Authorized
    
    BankAuthorized --> ImmediateRetry
    BankNotAuthorized --> NotifyCustomer
    
    DiagnoseIssue --> AutoFixable: Can Auto-Fix
    DiagnoseIssue --> NotAutoFixable: Cannot Auto-Fix
    
    AutoFixable --> ApplyFix
    NotAutoFixable --> EscalateToEngineering
    
    ApplyFix --> RetryPayment
    ImmediateRetry --> RetryPayment
    RequestNewCard --> WaitForNewCard
    WaitForNewCard --> CardProvided: Card Received
    WaitForNewCard --> CardNotProvided: No Card Received
    
    CardProvided --> RetryPayment
    CardNotProvided --> ScheduleFollowUp
    
    NotifyCustomer --> RetryAfterDelay
    RetryAfterDelay --> RetryPayment
    ScheduleFollowUp --> ManualIntervention
    
    RetryPayment --> CheckResult
    CheckResult --> RetrySuccess: Success
    CheckResult --> RetryFailed: Failed
    
    RetrySuccess --> SyncToAccounting
    RetryFailed --> CheckRetryCount
    
    CheckRetryCount --> WithinLimit: Within Limit
    CheckRetryCount --> ExceedsLimit: Exceeds Limit
    
    WithinLimit --> ScheduleNextRetry
    ExceedsLimit --> ManualIntervention
    
    ScheduleNextRetry --> WaitForRetry
    WaitForRetry --> RetryPayment
    
    SyncToAccounting --> NotifySuccess
    NotifySuccess --> [*]
    EscalateToEngineering --> [*]
    ManualIntervention --> [*]
    ScheduleFollowUp --> [*]
```

### 6. Monitoring and Observability Flow

```mermaid
graph TB
    subgraph "Application Metrics"
        API_METRICS[API Response Times]
        WORKFLOW_METRICS[Workflow Success Rates]
        PAYMENT_METRICS[Payment Recovery Rates]
        ERROR_METRICS[Error Rates by Type]
    end
    
    subgraph "Infrastructure Metrics"
        CPU_METRICS[CPU Utilization]
        MEMORY_METRICS[Memory Usage]
        NETWORK_METRICS[Network Traffic]
        DISK_METRICS[Disk I/O]
    end
    
    subgraph "Business Metrics"
        REVENUE_METRICS[Revenue Recovery]
        CUSTOMER_METRICS[Customer Satisfaction]
        PROVIDER_METRICS[Provider Performance]
        CONVERSION_METRICS[Conversion Rates]
    end
    
    subgraph "Monitoring Stack"
        PROMETHEUS[Prometheus Collection]
        GRAFANA[Grafana Dashboards]
        ALERTMANAGER[Alert Manager]
        JAEGER[Jaeger Tracing]
    end
    
    subgraph "Alert Channels"
        EMAIL_ALERTS[Email Alerts]
        SLACK_ALERTS[Slack Notifications]
        PAGERDUTY[PagerDuty Escalation]
        WEBHOOK_ALERTS[Custom Webhooks]
    end
    
    API_METRICS --> PROMETHEUS
    WORKFLOW_METRICS --> PROMETHEUS
    PAYMENT_METRICS --> PROMETHEUS
    ERROR_METRICS --> PROMETHEUS
    
    CPU_METRICS --> PROMETHEUS
    MEMORY_METRICS --> PROMETHEUS
    NETWORK_METRICS --> PROMETHEUS
    DISK_METRICS --> PROMETHEUS
    
    REVENUE_METRICS --> PROMETHEUS
    CUSTOMER_METRICS --> PROMETHEUS
    PROVIDER_METRICS --> PROMETHEUS
    CONVERSION_METRICS --> PROMETHEUS
    
    PROMETHEUS --> GRAFANA
    PROMETHEUS --> ALERTMANAGER
    
    ALERTMANAGER --> EMAIL_ALERTS
    ALERTMANAGER --> SLACK_ALERTS
    ALERTMANAGER --> PAGERDUTY
    ALERTMANAGER --> WEBHOOK_ALERTS
```

## Integration Technical Specifications

### Payment Provider API Integration Details

#### Stripe Integration
```yaml
stripe:
  webhook_endpoints:
    - payment_intent.succeeded
    - payment_intent.payment_failed
    - invoice.payment_failed
    - customer.subscription.deleted
  
  api_endpoints:
    - GET /v1/payment_intents/{id}
    - POST /v1/payment_intents/{id}/confirm
    - GET /v1/customers/{id}
  
  authentication:
    type: api_key
    header: Authorization
    format: Bearer sk_test_...
  
  retry_logic:
    max_attempts: 3
    backoff_strategy: exponential
    base_delay: 1h
    max_delay: 24h
```

#### PayPal Integration
```yaml
paypal:
  webhook_events:
    - PAYMENT.SALE.COMPLETED
    - PAYMENT.SALE.DENIED
    - PAYMENT.SALE.FAILED
    - BILLING.SUBSCRIPTION.CANCELLED
  
  api_methods:
    - POST /v1/payments/sale/{id}/retry
    - GET /v1/payments/sale/{id}
    - GET /v1/customers/{id}
  
  authentication:
    type: oauth2
    client_id: ${PAYPAL_CLIENT_ID}
    client_secret: ${PAYPAL_CLIENT_SECRET}
  
  rate_limits:
    requests_per_second: 10
    burst: 100
```

#### Braintree Integration
```yaml
braintree:
  webhooks:
    - subscription_went_past_due
    - subscription_charged_successfully
    - transaction_failed
    - disbursement
  
  api:
    - POST /transactions/{id}/retry
    - GET /transactions/{id}
    - GET /customers/{id}
  
  authentication:
    type: credentials
    merchant_id: ${BRAINTREE_MERCHANT_ID}
    public_key: ${BRAINTREE_PUBLIC_KEY}
    private_key: ${BRAINTREE_PRIVATE_KEY}
```

### Data Flow Specifications

#### Event Processing Pipeline
```yaml
pipeline:
  ingestion:
    rate_limit: 10000_events_per_second
    validation:
      - signature_verification
      - schema_validation
      - duplicate_detection
    
  processing:
    priority_levels:
      - critical: payment_failures
      - high: recovery_workflows
      - normal: analytics_events
      - low: periodic_tasks
    
    workers:
      count: 20
      batch_size: 100
      timeout: 30s
  
  storage:
    hot_data: redis_cluster
    cold_data: postgresql_replica
    archive: s3_glacier
```

#### Recovery Workflow Engine
```yaml
workflow_engine:
  supported_providers:
    - stripe
    - paypal
    - braintree
    - square
    - adobe_commerce
  
  retry_strategies:
    exponential_backoff:
      base_delay: 1h
      multiplier: 2
      max_delay: 24h
      jitter: true
    
    linear_backoff:
      base_delay: 6h
      increment: 6h
      max_delay: 48h
    
    immediate_retry:
      max_attempts: 2
      delay: 30s
  
  decision_matrix:
    failure_types:
      insufficient_funds:
        strategy: exponential_backoff
        max_attempts: 3
        notify_customer: true
      
      card_declined:
        strategy: linear_backoff
        max_attempts: 5
        require_auth: true
      
      technical_error:
        strategy: immediate_retry
        max_attempts: 2
        escalate_after: true
```

## Security and Compliance Flow

### Payment Data Security
```mermaid
graph LR
    subgraph "Payment Provider"
        PP_ENCRYPT[Encrypted Payment Data]
    end
    
    subgraph "Ingress Security"
        WAF[Web Application Firewall]
        DDoS[DDoS Protection]
        RATE_LIMIT[Rate Limiting]
        VALIDATION[Input Validation]
    end
    
    subgraph "Application Security"
        AUTH[Authentication]
        AUTHZ[Authorization]
        ENCRYPTION[Data Encryption]
        MASKING[Data Masking]
    end
    
    subgraph "Data Security"
        ENCRYPTION_AT_REST[AES-256 at Rest]
        ENCRYPTION_IN_TRANSIT[TLS 1.3 in Transit]
        KEY_MANAGEMENT[Cloud KMS]
        ACCESS_CONTROL[RBAC/ABAC]
    end
    
    PP_ENCRYPT --> WAF
    WAF --> DDoS
    DDoS --> RATE_LIMIT
    RATE_LIMIT --> VALIDATION
    VALIDATION --> AUTH
    AUTH --> AUTHZ
    AUTHZ --> ENCRYPTION
    ENCRYPTION --> MASKING
    MASKING --> ENCRYPTION_AT_REST
    ENCRYPTION_AT_REST --> ENCRYPTION_IN_TRANSIT
    ENCRYPTION_IN_TRANSIT --> KEY_MANAGEMENT
    KEY_MANAGEMENT --> ACCESS_CONTROL
```

---

**Document Version**: 1.0  
**Created**: 2026-02-12  
**Author**: Solution Architecture Team  
**Purpose**: Technical visualization of Payment Watchdog integration flows

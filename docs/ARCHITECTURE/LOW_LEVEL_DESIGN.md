# Payment Watchdog - Low-Level Design Document

## 📋 Overview
This document provides the low-level design for the Worker-API decoupling implementation, focusing on the technical details of event-driven architecture, shared interfaces, and independent deployment strategies.

## 🎯 Objectives

### **Primary Goals**
1. **Decouple Worker from API**: Remove all internal API package dependencies
2. **Event-Driven Communication**: Implement Redis-based event bus for async communication
3. **Shared Interfaces**: Create common interfaces for database, event bus, and business logic
4. **Independent Deployment**: Enable worker to build and deploy without API dependencies
5. **Business Logic Preservation**: Maintain all essential business functionality

## 🏗️ Architecture Decisions

### **Event-Driven Microservices**
- **Communication Pattern**: Publisher-Subscriber via Redis event bus
- **Event Types**: Standardized payment events with JSON schemas
- **Async Processing**: Non-blocking event handling with error recovery
- **Event Ordering**: Redis Streams for guaranteed ordering

### **Shared Interface Design**
- **Location**: `shared/interfaces/` package
- **Scope**: Database, Event Bus, Business Logic interfaces
- **Implementation**: Both API and Worker implement same interfaces
- **Versioning**: Semantic versioning for backward compatibility

### **Configuration Management**
- **Worker Config**: `worker/config/` with shared interfaces
- **API Config**: `api/config/` updated for event publishing
- **Environment**: Development, Staging, Production configurations
- **Sovereign Mode**: Australian data residency compliance

## 📊 Component Design

### **Shared Package Structure**
```
shared/
├── events/
│   ├── types.go              # Payment event definitions
│   └── constructors.go        # Event creation helpers
├── interfaces/
│   ├── payment.go           # Business logic interfaces
│   ├── config.go            # Configuration interfaces
│   └── database.go          # Database interfaces
└── go.mod                   # Shared module definition
```

### **Worker Service Structure**
```
worker/
├── cmd/
│   └── main.go              # Event-driven entry point
├── config/
│   └── config.go            # Worker-specific configuration
├── internal/
│   ├── services/
│   │   ├── event_processor.go    # Main event orchestrator
│   │   ├── analytics_service.go  # Payment analytics
│   │   ├── rules_service.go       # Business rules engine
│   │   └── mediators_service.go  # External integrations
│   ├── database/
│   │   └── postgres.go         # PostgreSQL implementation
│   └── eventbus/
│       └── redis_eventbus.go   # Redis event bus
└── go.mod                     # Worker module dependencies
```

## 🔧 Technical Implementation

### **Identity & Global Bridge**
- **Sovereign Identity**: Firebase Client SDK (UI) + Admin SDK (API)
- **Identity Linkage**: OAuth/Email identifies user; Custom Claims identify `tenant_id`.
- **Provisioning Engine**: Automated GORM migration runner for isolated schemas.
- **Tenant Context**: Injected into every Go handler context via middleware.

### **Event Bus Implementation**
- **Technology**: Redis Streams with go-redis/v8
- **Pattern**: Publisher-Subscriber with topic-based routing
- **Error Handling**: Retry logic with exponential backoff
- **Health Checks**: Connection monitoring and status reporting

### **Database & Tenant Isolation**
- **Isolation Model**: Schema-per-tenant (PostgreSQL)
- **Middleware**: `TenantIsolationMiddleware` intercepts every request to run `SET search_path TO tenant_<id>, public`.
- **Implementation**: Scoped GORM DB instances injected into context.
- **Connection Pooling**: Shared pool across schemas to optimize resource usage on OCI.
- **Health Monitoring**: Connection status and per-tenant query performance.

### **Business Services**
- **Analytics Engine**: Pattern detection and ML-based predictions
- **Rules Engine**: Configurable business rule evaluation
- **Mediator Service**: External accounting system integration
- **Event Processor**: Orchestrates all business services

## 📋 Data Flow

### **Sovereign Onboarding Flow**
1. **User Login**: Firebase identifies the admin.
2. **Identity Guard**: `useAuth` hook detects missing `tenant_id` and redirects to `/onboarding`.
3. **Provisioning Request**: UI sends company details to `POST /api/onboarding/provision`.
4. **Schema Creation**: API initiates `CREATE SCHEMA tenant_<id>` and runs core migrations.
5. **Claim Injection**: Firebase Admin SDK updates user claims with `tenant_id`.
6. **Isolated Access**: Future requests are automatically scoped to the private AU schema.

```json
{
  "id": "event-uuid",
  "event_type": "payment.failure.detected",
  "tenant_id": "smb-001",
  "company_id": "company-uuid",
  "payment_id": "payment-uuid", 
  "amount": 99.99,
  "currency": "AUD",
  "status": "failed",
  "timestamp": "2026-03-31T10:30:00Z"
}
```

## 🔒 Security & Compliance

### **Data Residency**
- **Sovereign Mode**: Configurable Australian data residency
- **Local Processing**: No external data transmission in sovereign mode
- **Compliance**: Australian privacy regulations adherence

### **Event Security**
- **Authentication**: Redis AUTH for event bus access
- **Authorization**: Role-based access control
- **Encryption**: TLS for all Redis connections
- **Validation**: Event schema validation before processing

## 🚀 Deployment Strategy

### **Independent Worker Deployment**
- **Containerization**: Docker image with minimal dependencies
- **Configuration**: Environment-based config management
- **Health Checks**: Readiness and liveness probes
- **Scaling**: Horizontal pod scaling with load balancing

### **API Service Updates**
- **Event Publishing**: Replace direct service calls with event publishing
- **Shared Interfaces**: Adopt shared business logic interfaces
- **Backward Compatibility**: Maintain existing API contracts during transition

## 📈 Performance Considerations

### **Event Processing**
- **Throughput**: Target 10,000+ events/second
- **Latency**: Sub-100ms event processing
- **Error Rate**: < 0.1% event processing failures
- **Scalability**: Linear scaling with Redis cluster

### **Resource Management**
- **Memory**: < 512MB per worker instance
- **CPU**: < 0.5 vCPU per worker instance
- **Connections**: Redis connection pooling (max 10 per worker)

## 🧪 Testing Strategy

### **Unit Tests**
- **Business Services**: 90%+ coverage for analytics, rules, mediators
- **Event Bus**: Redis event bus testing with mock Redis
- **Database**: PostgreSQL integration testing with testcontainers

### **Integration Tests**
- **Event Flow**: End-to-end payment failure processing
- **Multi-Service**: API → Redis → Worker flow validation
- **Performance**: Load testing with concurrent event processing

### **Contract Testing**
- **Event Schemas**: Validate event structure compatibility
- **Interface Compliance**: Ensure shared interface implementations
- **Version Compatibility**: Test backward compatibility scenarios

## 📝 Migration Path

### **Phase 1: Foundation**
- ✅ Shared package creation
- ✅ Worker service implementation
- ✅ Event-driven architecture

### **Phase 2: Multi-Tenant SaaS (Sovereign-AU)**
- ✅ Firebase Identity Bridge (Client + Admin SDK)
- ✅ Automated Onboarding Wizard and Provisioner
- ✅ Isolated Schema-per-tenant Database Fabric
- ✅ **96%+ Test Coverage (Vitest)** on core auth logic

### **Phase 3: Optimization** (Future)
- 📈 Performance optimization and load testing
- 📈 Multi-region AU deployment support (Sydney/Melbourne)

---

**Document Status**: ✅ **Complete**  
**Last Updated**: March 31, 2026  
**Version**: 1.0.0

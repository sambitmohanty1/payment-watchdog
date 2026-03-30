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

### **Event Bus Implementation**
- **Technology**: Redis Streams with go-redis/v8
- **Pattern**: Publisher-Subscriber with topic-based routing
- **Error Handling**: Retry logic with exponential backoff
- **Health Checks**: Connection monitoring and status reporting

### **Database Interface**
- **Implementation**: PostgreSQL with GORM
- **Connection Pooling**: Configurable pool sizes
- **Migration Support**: Database schema versioning
- **Health Monitoring**: Connection status and query performance

### **Business Services**
- **Analytics Engine**: Pattern detection and ML-based predictions
- **Rules Engine**: Configurable business rule evaluation
- **Mediator Service**: External accounting system integration
- **Event Processor**: Orchestrates all business services

## 📋 Data Flow

### **Event Flow**
1. **API Service** detects payment failure
2. **Publishes** `payment.failure.detected` event to Redis
3. **Worker Service** subscribes to event topic
4. **Processes** through analytics → rules → mediators
5. **Publishes** `payment.failure.processed` event back to Redis
6. **API Service** can optionally subscribe to processed events

### **Event Schema**
```json
{
  "id": "event-uuid",
  "event_type": "payment.failure.detected",
  "company_id": "company-uuid",
  "payment_id": "payment-uuid", 
  "amount": 99.99,
  "currency": "AUD",
  "status": "failed",
  "timestamp": "2026-03-31T10:30:00Z",
  "metadata": {
    "provider": "stripe",
    "reason": "insufficient_funds"
  }
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

### **Phase 1: Foundation** (Current)
- ✅ Shared package creation
- ✅ Worker service implementation
- ✅ Event-driven architecture
- ✅ Independent deployment capability

### **Phase 2: API Integration** (Next)
- 🔄 Update API service to publish events
- 🔄 Adopt shared business interfaces
- 🔄 Maintain backward compatibility
- 🔄 Add integration tests

### **Phase 3: Optimization** (Future)
- 📈 Performance optimization and load testing
- 📈 Enhanced monitoring and observability
- 📈 Advanced error handling and recovery
- 📈 Multi-region deployment support

---

**Document Status**: ✅ **Complete**  
**Last Updated**: March 31, 2026  
**Version**: 1.0.0

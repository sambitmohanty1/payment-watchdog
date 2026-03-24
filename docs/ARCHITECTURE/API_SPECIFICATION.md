# Payment Watchdog - API Specification

## 📋 Overview

This document provides comprehensive API specifications for the Payment Watchdog platform. All APIs follow RESTful principles with JSON request/response formats and proper HTTP status codes.

---

## 🎯 API Architecture

### **📐 API Design Principles**
- **RESTful Design**: Proper HTTP methods and status codes
- **JSON Format**: Standardized request/response format
- **Versioning**: API versioning in URL path (`/api/v1/`)
- **Authentication**: JWT-based authentication
- **Rate Limiting**: Request throttling and rate limiting
- **Error Handling**: Consistent error response format
- **Documentation**: OpenAPI/Swagger specification

### **🔐 Authentication & Authorization**
- **JWT Tokens**: Bearer token authentication
- **API Keys**: Service-to-service authentication
- **RBAC**: Role-based access control
- **CORS**: Cross-origin resource sharing configuration

---

## 📚 API Endpoints

### **🏠 Base URL**
```
Production: https://api.payment-watchdog.com.au/api/v1
Staging: https://staging-api.payment-watchdog.com.au/api/v1
Development: http://localhost:8091/api/v1
```

### **🔍 Health Check Endpoints**

#### **GET /health**
Basic health check for load balancers and monitoring systems.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-03-24T23:30:00Z",
  "version": "2.0.0"
}
```

#### **GET /health/detailed**
Comprehensive health check with service status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-03-24T23:30:00Z",
  "version": "2.0.0",
  "services": {
    "api": {
      "status": "healthy",
      "uptime": "72h30m15s",
      "response_time": "45ms"
    },
    "database": {
      "status": "connected",
      "host": "localhost",
      "connections": 5,
      "latency": "2ms"
    },
    "redis": {
      "status": "connected",
      "host": "localhost",
      "memory_used": "125MB",
      "connections": 3
    },
    "workers": {
      "status": "active",
      "count": 3,
      "last_run": "2025-03-24T23:29:45Z",
      "failed_jobs": 0
    }
  }
}
```

---

## 💳 Payment Failure Management

### **📋 Payment Failures**

#### **GET /payment-failures**
Retrieve payment failures with filtering and pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 20, max: 100) - Items per page
- `company_id` (string, optional) - Filter by company ID
- `status` (string, optional) - Filter by status
- `provider` (string, optional) - Filter by payment provider
- `from_date` (string, optional) - Filter from date (ISO 8601)
- `to_date` (string, optional) - Filter to date (ISO 8601)

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "company_id": "123e4567-e89b-12d3-a456-426614174000",
      "provider_id": "stripe",
      "provider_event_id": "pi_1234567890",
      "amount_cents": 250000,
      "currency": "AUD",
      "customer_id": "cus_1234567890",
      "customer_name": "John Doe",
      "failure_reason": "insufficient_funds",
      "status": "received",
      "priority": "medium",
      "risk_score": 65.5,
      "occurred_at": "2025-03-24T22:00:00Z",
      "detected_at": "2025-03-24T22:05:00Z",
      "processed_at": null,
      "created_at": "2025-03-24T22:05:00Z",
      "updated_at": "2025-03-24T22:05:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

#### **GET /payment-failures/{id}**
Retrieve a specific payment failure by ID.

**Path Parameters:**
- `id` (string, required) - Payment failure UUID

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "company_id": "123e4567-e89b-12d3-a456-426614174000",
  "provider_id": "stripe",
  "provider_event_id": "pi_1234567890",
  "amount_cents": 250000,
  "currency": "AUD",
  "customer_id": "cus_1234567890",
  "customer_name": "John Doe",
  "failure_reason": "insufficient_funds",
  "status": "processed",
  "priority": "medium",
  "risk_score": 65.5,
  "occurred_at": "2025-03-24T22:00:00Z",
  "detected_at": "2025-03-24T22:05:00Z",
  "processed_at": "2025-03-24T22:10:00Z",
  "created_at": "2025-03-24T22:05:00Z",
  "updated_at": "2025-03-24T22:10:00Z",
  "recovery_attempts": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "attempt_number": 1,
      "method": "stripe_retry",
      "status": "failed",
      "attempted_at": "2025-03-24T22:07:00Z",
      "error_message": "Card declined: insufficient funds"
    }
  ]
}
```

#### **POST /payment-failures**
Create a new payment failure record (typically used by webhook processor).

**Request Body:**
```json
{
  "provider_id": "stripe",
  "provider_event_id": "pi_1234567890",
  "amount_cents": 250000,
  "currency": "AUD",
  "customer_id": "cus_1234567890",
  "customer_name": "John Doe",
  "failure_reason": "insufficient_funds",
  "occurred_at": "2025-03-24T22:00:00Z"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "received",
  "created_at": "2025-03-24T22:05:00Z"
}
```

---

## 🔄 Recovery Workflows

### **📋 Workflow Management**

#### **GET /workflows**
Retrieve recovery workflows for a company.

**Query Parameters:**
- `company_id` (string, optional) - Filter by company ID
- `is_active` (boolean, optional) - Filter by active status
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 20) - Items per page

**Response:**
```json
{
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "company_id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Micro-Merchant Recovery",
      "description": "Optimized recovery for micro-transactions under $100",
      "priority": "high",
      "is_active": true,
      "steps": [
        {
          "id": "880e8400-e29b-41d4-a716-446655440000",
          "step_type": "retry_payment",
          "order": 1,
          "config": {
            "max_attempts": 3,
            "delay_minutes": 60
          }
        },
        {
          "id": "990e8400-e29b-41d4-a716-446655440000",
          "step_type": "payto_request",
          "order": 2,
          "config": {
            "amount_threshold_cents": 10000
          }
        }
      ],
      "created_at": "2025-03-24T20:00:00Z",
      "updated_at": "2025-03-24T22:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

#### **POST /workflows**
Create a new recovery workflow.

**Request Body:**
```json
{
  "company_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Custom Recovery Workflow",
  "description": "Custom workflow for specific business needs",
  "priority": "medium",
  "steps": [
    {
      "step_type": "retry_payment",
      "order": 1,
      "config": {
        "max_attempts": 2,
        "delay_minutes": 30
      }
    },
    {
      "step_type": "send_notification",
      "order": 2,
      "config": {
        "channel": "email",
        "template": "payment_failure_notice"
      }
    }
  ]
}
```

**Response:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "status": "created",
  "created_at": "2025-03-24T22:30:00Z"
}
```

### **📊 Workflow Executions**

#### **GET /workflow-executions**
Retrieve workflow execution history.

**Query Parameters:**
- `workflow_id` (string, optional) - Filter by workflow ID
- `payment_failure_id` (string, optional) - Filter by payment failure ID
- `status` (string, optional) - Filter by execution status
- `from_date` (string, optional) - Filter from date
- `to_date` (string, optional) - Filter to date
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 20) - Items per page

**Response:**
```json
{
  "data": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440000",
      "workflow_id": "770e8400-e29b-41d4-a716-446655440000",
      "payment_failure_id": "550e8400-e29b-41d4-a716-446655440000",
      "company_id": "123e4567-e89b-12d3-a456-426614174000",
      "status": "completed",
      "current_step_index": 2,
      "context": {
        "retry_attempts": 2,
        "payto_agreement_id": "payto_123456"
      },
      "started_at": "2025-03-24T22:10:00Z",
      "completed_at": "2025-03-24T22:25:00Z",
      "created_at": "2025-03-24T22:10:00Z",
      "updated_at": "2025-03-24T22:25:00Z",
      "step_executions": [
        {
          "id": "101e8400-e29b-41d4-a716-446655440000",
          "step_type": "retry_payment",
          "status": "completed",
          "started_at": "2025-03-24T22:10:00Z",
          "completed_at": "2025-03-24T22:12:00Z",
          "result": {
            "success": false,
            "error": "Card declined"
          }
        },
        {
          "id": "102e8400-e29b-41d4-a716-446655440000",
          "step_type": "payto_request",
          "status": "completed",
          "started_at": "2025-03-24T22:15:00Z",
          "completed_at": "2025-03-24T22:25:00Z",
          "result": {
            "success": true,
            "payto_agreement_id": "payto_123456"
          }
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## 📈 Analytics & Reporting

### **📊 Recovery Analytics**

#### **GET /analytics/recovery-summary**
Get recovery analytics summary for a company.

**Query Parameters:**
- `company_id` (string, optional) - Filter by company ID
- `from_date` (string, required) - From date (ISO 8601)
- `to_date` (string, required) - To date (ISO 8601)
- `group_by` (string, optional) - Group by period (day, week, month)

**Response:**
```json
{
  "period": {
    "from_date": "2025-03-01T00:00:00Z",
    "to_date": "2025-03-24T23:59:59Z"
  },
  "summary": {
    "total_failures": 1250,
    "total_recovered": 875,
    "recovery_rate": 70.0,
    "total_recovered_amount_cents": 12500000,
    "average_recovery_time_hours": 2.5
  },
  "by_provider": [
    {
      "provider": "stripe",
      "failures": 1000,
      "recovered": 700,
      "recovery_rate": 70.0,
      "recovered_amount_cents": 10000000
    },
    {
      "provider": "paypal",
      "failures": 250,
      "recovered": 175,
      "recovery_rate": 70.0,
      "recovered_amount_cents": 2500000
    }
  ],
  "by_reason": [
    {
      "reason": "insufficient_funds",
      "failures": 500,
      "recovered": 400,
      "recovery_rate": 80.0,
      "recovered_amount_cents": 6000000
    },
    {
      "reason": "card_declined",
      "failures": 400,
      "recovered": 250,
      "recovery_rate": 62.5,
      "recovered_amount_cents": 3000000
    }
  ],
  "by_amount_range": [
    {
      "range": "0-100",
      "failures": 800,
      "recovered": 600,
      "recovery_rate": 75.0,
      "recovered_amount_cents": 4000000
    },
    {
      "range": "100-1000",
      "failures": 350,
      "recovered": 225,
      "recovery_rate": 64.3,
      "recovered_amount_cents": 7000000
    },
    {
      "range": "1000+",
      "failures": 100,
      "recovered": 50,
      "recovery_rate": 50.0,
      "recovered_amount_cents": 1500000
    }
  ]
}
```

#### **GET /analytics/trends**
Get recovery trends over time.

**Query Parameters:**
- `company_id` (string, optional) - Filter by company ID
- `from_date` (string, required) - From date (ISO 8601)
- `to_date` (string, required) - To date (ISO 8601)
- `granularity` (string, default: "day") - Data granularity (hour, day, week, month)

**Response:**
```json
{
  "period": {
    "from_date": "2025-03-01T00:00:00Z",
    "to_date": "2025-03-24T23:59:59Z"
  },
  "granularity": "day",
  "data": [
    {
      "date": "2025-03-01T00:00:00Z",
      "failures": 45,
      "recovered": 32,
      "recovery_rate": 71.1,
      "recovered_amount_cents": 125000
    },
    {
      "date": "2025-03-02T00:00:00Z",
      "failures": 52,
      "recovered": 38,
      "recovery_rate": 73.1,
      "recovered_amount_cents": 145000
    }
  ]
}
```

---

## 🏢 Company Management

### **📋 Company Operations**

#### **GET /companies**
Retrieve companies with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 20) - Items per page
- `is_active` (boolean, optional) - Filter by active status

**Response:**
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Acme Corporation",
      "email": "billing@acme.com",
      "phone": "+61-2-1234-5678",
      "address": {
        "street": "123 Business St",
        "city": "Sydney",
        "state": "NSW",
        "postal_code": "2000",
        "country": "AU"
      },
      "is_active": true,
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-03-24T15:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

#### **POST /companies**
Create a new company.

**Request Body:**
```json
{
  "name": "New Company",
  "email": "billing@newcompany.com",
  "phone": "+61-2-9876-5432",
  "address": {
    "street": "456 Commerce Ave",
    "city": "Melbourne",
    "state": "VIC",
    "postal_code": "3000",
    "country": "AU"
  }
}
```

**Response:**
```json
{
  "id": "124e4567-e89b-12d3-a456-426614174001",
  "status": "created",
  "created_at": "2025-03-24T22:45:00Z"
}
```

---

## 🔔 Webhook Management

### **📋 Webhook Endpoints**

#### **POST /webhooks/stripe**
Stripe webhook endpoint for payment event processing.

**Headers:**
- `Stripe-Signature` (string, required) - Stripe signature verification

**Request Body**: Raw Stripe webhook JSON payload

**Response:**
```json
{
  "status": "processed",
  "event_id": "evt_1234567890",
  "event_type": "payment_intent.payment_failed",
  "processed_at": "2025-03-24T22:50:00Z"
}
```

#### **GET /webhooks/configurations**
Retrieve webhook configurations for a company.

**Query Parameters:**
- `company_id` (string, optional) - Filter by company ID

**Response:**
```json
{
  "data": [
    {
      "id": "135e8400-e29b-41d4-a716-446655440000",
      "company_id": "123e4567-e89b-12d3-a456-426614174000",
      "provider": "stripe",
      "endpoint_url": "https://api.payment-watchdog.com.au/api/v1/webhooks/stripe",
      "secret": "whsec_1234567890abcdef",
      "events": [
        "payment_intent.payment_failed",
        "payment_intent.succeeded",
        "invoice.payment_failed"
      ],
      "is_active": true,
      "created_at": "2025-03-24T20:00:00Z",
      "updated_at": "2025-03-24T22:00:00Z"
    }
  ]
}
```

---

## 🛡️ Error Handling

### **📋 Standard Error Response Format**

All API errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "amount_cents",
      "issue": "must be a positive integer"
    },
    "timestamp": "2025-03-24T22:55:00Z",
    "request_id": "req_1234567890"
  }
}
```

### **🔢 HTTP Status Codes**

| Status Code | Meaning | Usage |
|-------------|---------|--------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 204 | No Content | Successful deletion |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 422 | Unprocessable Entity | Validation error |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### **🚨 Error Codes**

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| VALIDATION_ERROR | 400 | Request validation failed |
| AUTHENTICATION_ERROR | 401 | Authentication required |
| AUTHORIZATION_ERROR | 403 | Insufficient permissions |
| NOT_FOUND_ERROR | 404 | Resource not found |
| CONFLICT_ERROR | 409 | Resource conflict |
| RATE_LIMIT_ERROR | 429 | Rate limit exceeded |
| INTERNAL_ERROR | 500 | Internal server error |
| SERVICE_UNAVAILABLE | 503 | Service temporarily unavailable |

---

## 📊 Rate Limiting

### **🔢 Rate Limit Rules**

| Endpoint | Rate Limit | Time Window |
|---------|-------------|-------------|
| GET /payment-failures | 1000 requests | 1 hour |
| POST /payment-failures | 100 requests | 1 minute |
| GET /analytics/* | 500 requests | 1 hour |
| POST /webhooks/* | 1000 requests | 1 minute |
| Other endpoints | 500 requests | 1 hour |

### **📋 Rate Limit Headers**

Rate limited responses include these headers:
- `X-RateLimit-Limit`: Maximum requests per time window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when rate limit resets (Unix timestamp)

---

## 🔍 Pagination

### **📄 Pagination Parameters**

- `page` (integer, default: 1) - Page number (1-indexed)
- `limit` (integer, default: 20, max: 100) - Items per page

### **📋 Pagination Response**

Paginated responses include pagination metadata:
```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## 🧪 Testing & Examples

### **🔧 Authentication Example**

```bash
# Get JWT token
curl -X POST https://api.payment-watchdog.com.au/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# Use token for authenticated requests
curl -X GET https://api.payment-watchdog.com.au/api/v1/payment-failures \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### **📊 API Usage Examples**

```bash
# Get payment failures with filtering
curl -X GET "https://api.payment-watchdog.com.au/api/v1/payment-failures?company_id=123e4567-e89b-12d3-a456-426614174000&status=received&page=1&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Create a payment failure
curl -X POST https://api.payment-watchdog.com.au/api/v1/payment-failures \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "stripe",
    "provider_event_id": "pi_1234567890",
    "amount_cents": 250000,
    "currency": "AUD",
    "customer_id": "cus_1234567890",
    "customer_name": "John Doe",
    "failure_reason": "insufficient_funds",
    "occurred_at": "2025-03-24T22:00:00Z"
  }'

# Get analytics summary
curl -X GET "https://api.payment-watchdog.com.au/api/v1/analytics/recovery-summary?company_id=123e4567-e89b-12d3-a456-426614174000&from_date=2025-03-01T00:00:00Z&to_date=2025-03-24T23:59:59Z" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 📚 OpenAPI Specification

### **📋 Swagger UI**

Interactive API documentation available at:
- **Production**: https://api.payment-watchdog.com.au/docs
- **Staging**: https://staging-api.payment-watchdog.com.au/docs
- **Development**: http://localhost:8091/docs

### **📄 OpenAPI JSON**

Raw OpenAPI specification available at:
- **Production**: https://api.payment-watchdog.com.au/api/v1/openapi.json
- **Staging**: https://staging-api.payment-watchdog.com.au/api/v1/openapi.json
- **Development**: http://localhost:8091/api/v1/openapi.json

---

## 🔄 API Versioning

### **📅 Version History**

| Version | Release Date | Changes |
|---------|-------------|---------|
| 2.0.0 | 2025-03-24 | Current version with enhanced analytics and workflow management |
| 1.2.0 | 2025-02-15 | Added webhook management and improved error handling |
| 1.1.0 | 2025-01-30 | Added recovery workflows and analytics endpoints |
| 1.0.0 | 2024-12-01 | Initial API release |

### **🔄 Versioning Strategy**

- **URL Versioning**: `/api/v1/`, `/api/v2/`
- **Backward Compatibility**: Maintain previous versions for at least 6 months
- **Deprecation**: 3-month deprecation notice for breaking changes
- **Documentation**: Version-specific documentation for each API version

---

## 🎯 Best Practices

### **✅ Recommended Practices**

1. **Authentication**: Always include JWT token in `Authorization: Bearer` header
2. **Error Handling**: Check HTTP status codes and error response format
3. **Pagination**: Use pagination for large datasets to avoid timeouts
4. **Rate Limiting**: Respect rate limits and implement exponential backoff
5. **Time Zones**: Use ISO 8601 format with UTC timestamps
6. **Idempotency**: Use idempotent requests for safe retries
7. **Validation**: Validate request parameters before sending
8. **Logging**: Include request IDs for debugging and support

### **⚠️ Common Pitfalls**

1. **Missing Authentication**: Always include valid JWT token
2. **Invalid Timestamps**: Use proper ISO 8601 format
3. **Large Payloads**: Keep request payloads under 1MB
5. **Rate Limiting**: Don't exceed rate limits
6. **Invalid JSON**: Ensure proper JSON formatting
7. **Missing Parameters**: Check required parameters in documentation
8. **Wrong HTTP Methods**: Use correct HTTP methods for endpoints

---

## 📞 Support & Documentation

### **📚 Documentation Resources**
- **API Reference**: This document
- **OpenAPI Specification**: Interactive API documentation
- **Integration Guides**: Step-by-step integration tutorials
- **Code Examples**: Sample code in multiple languages

### **🆘 Support Channels**
- **Email**: api-support@payment-watchdog.com.au
- **Documentation**: https://docs.payment-watchdog.com.au
- **Status Page**: https://status.payment-watchdog.com.au
- **GitHub Issues**: https://github.com/payment-watchdog/issues

### **🔄 API Updates**
- **Change Log**: Documented in release notes
- **Deprecation Notices**: 3-month advance notice
- **Breaking Changes**: Major version increments
- **New Features**: Minor version increments

---

## 🎯 Last Updated
- **Date**: 2025-03-24
- **Version**: 2.0.0
- **Author**: API Architecture Team
- **Review**: Architecture Committee

---

**🚨 This document serves as the authoritative source for Payment Watchdog API specifications and usage guidelines.**

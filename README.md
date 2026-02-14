# ⚡ **Payment Watchdog**
## AI-Powered SaaS Payment Failure Intelligence Platform

[![Go Version](https://img.shields.io/badge/Go-1.24.4-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-success.svg)](README.md)

---

## 📋 **Project Overview**

**Payment Watchdog** is a comprehensive AI-powered payment failure intelligence platform that transforms how SaaS companies handle payment failures. Never get blindsided by failed payments again — detect, recover, and prevent lost cashflow with one smart platform.

### **🎯 Mission**
**"Never get blindsided by failed payments again — detect, recover, and prevent lost cashflow with one smart platform."**

### **🚀 Vision**
Transform how SaaS companies handle payment failures by providing:
- **Real-time Detection**: Instant payment failure detection across all providers
- **Intelligent Recovery**: Automated retry mechanisms and customer communication
- **Predictive Intelligence**: Prevent failures before they happen
- **Unified Platform**: Single view across all payment methods and SaaS platforms

**Market Coverage**: 60-70% of SaaS companies (Stripe + PayPal + Braintree)
**Revenue Potential**: $2.4B addressable market in SaaS payment failures

---

## 🏗️ **Architecture**

### **Microservices Architecture**
- **API Service**: REST API with Go + Gin framework
- **Worker Service**: Background processing with concurrency controls
- **Recovery Orchestration**: Advanced payment recovery workflows with multi-provider integration
- **Web Interface**: Next.js dashboard with Tailwind CSS
- **Database**: Database-agnostic design supporting PostgreSQL (production) and SQLite (testing)
- **Cache/Queue**: Redis for event processing
- **Event Bus**: Redis-based asynchronous processing

### **Technology Stack**
- **Backend**: Go 1.23, Gin, GORM, Redis, PostgreSQL/SQLite
- **Frontend**: Next.js 14, TypeScript, Tailwind CSS
- **Infrastructure**: Docker, Kubernetes, Kong API Gateway
- **CI/CD**: GitHub Actions, Docker Hub, Checkmarx Security
- **Testing**: SQLite in-memory databases for ephemeral testing

### **Database Architecture**
- **Production**: PostgreSQL with advanced analytics and JSONB support
- **Testing**: SQLite in-memory databases for fast, ephemeral testing
- **Database-Agnostic Service**: Automatic detection and adaptation to database type
- **No Race Conditions**: Eliminated sqlmock concurrency issues with real database testing

---

## 🚀 **Quick Start**

### **Prerequisites**
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+

### **Local Development**
```bash
# Clone the repository
git clone https://github.com/payment-watchdog.git
cd payment-watchdog

# Start all services
docker-compose up -d

# API will be available at http://localhost:8080
# Web interface at http://localhost:4896
# Database at localhost:5432
# Redis at localhost:6379
```

### **Development Commands**
```bash
# API Service
cd api && go run cmd/main.go

# Worker Service
cd worker && go run cmd/main.go

# Web Interface
cd web && npm run dev
```

### **Testing**
```bash
# Run all tests with ephemeral SQLite databases
go test ./...

# Run specific service tests
go test ./services -v

# Test recovery analytics (uses SQLite in-memory)
go test ./services -run TestGetRecoveryMetrics

# Test with PostgreSQL (requires database setup)
./start-payment-watchdog-db.sh  # Starts PostgreSQL on port 5569
go test ./services -run TestGetRecoveryMetrics  # Will use PostgreSQL if available
```

### **Testing Architecture**
- **Ephemeral Testing**: Uses SQLite in-memory databases by default
- **No External Dependencies**: Tests run anywhere without database setup
- **Database-Agnostic**: Service automatically adapts to available database
- **Race Condition Free**: Eliminated sqlmock concurrency issues
- **Production-Like**: Optional PostgreSQL testing for production validation

---

## 📊 **Core Features**

### **1. Payment Failure Detection**
- Real-time monitoring of payment webhooks
- Multi-provider support (Stripe, PayPal, Braintree)
- Pattern recognition and anomaly detection

### **2. Intelligent Recovery**
- Smart retry logic with exponential backoff
- Payment method fallback strategies
- Automated customer communication
- **Recovery Orchestration**: Advanced workflow management with OpenTelemetry tracing
- Multi-provider integration (Stripe, Xero, QuickBooks)
- Configurable retry policies and failure prediction

### **3. Analytics Dashboard**
- Real-time failure analytics
- Customer risk scoring
- Revenue impact tracking
- Predictive failure forecasting

### **4. Integration Hub**
- REST API for easy integration
- Webhook support for real-time updates
- Zapier and native integrations

---

## 🔧 **API Documentation**

### **Base URLs**
- **Local**: `http://localhost:8080`
- **Production**: `https://api.payment-watchdog.com`

### **Key Endpoints**
- `GET /health` - Service health check
- `GET /api/v1/dashboard/stats` - Dashboard statistics
- `GET /api/v1/analytics/test` - Analytics test endpoint
- `POST /api/v1/payments/webhook` - Payment webhook processing
- `GET /health` - Recovery Orchestration health (Port 8086)
- `GET /metrics` - Recovery Orchestration metrics (Port 8086)

---

## 🐳 **Docker Services**

```yaml
services:
  api:                    # Go REST API (Port 8080)
  worker:                 # Background processing (No external port)
  recovery-orchestration: # Payment recovery workflows (Port 8086)
  web:                    # Next.js dashboard (Port 4896)
  postgres:              # Database (Port 5432)
  redis:                 # Cache & Queue (Port 6379)
  mailhog:               # Email testing (Port 8025)
```

---

## 🚀 **Deployment**

### **Local Development**
```bash
docker-compose up -d
```

### **Production Deployment**
```bash
# Build and push images
docker build -t payment-watchdog/api ./api
docker build -t payment-watchdog/worker ./worker
docker build -t payment-watchdog/recovery-orchestration ./recovery-orchestration
docker build -t payment-watchdog/web ./web

# Deploy to Kubernetes (Zero Touch Deployment)
kustomize build api/deployments/kubernetes | kubectl apply -f -

# Or deploy individual services
kustomize build api/deployments/kubernetes/apps/recovery-orchestration | kubectl apply -f -
```

### **CI/CD Pipeline**
- Automated testing and building
- Security scanning with Checkmarx
- Docker image building and pushing
- Kubernetes deployment to staging/production

---

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### **Development Guidelines**
- Follow Go best practices
- Use conventional commits
- Write comprehensive tests
- Update documentation

---

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 📞 **Support**

For support, email support@payment-watchdog.com or join our Slack community.

**Payment Watchdog** - Never lose another payment again! ⚡

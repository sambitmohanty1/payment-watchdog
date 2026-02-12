# Payment Watchdog Development Makefile

.PHONY: help start stop restart logs clean build test lint format deps update

# Default target
help: ## Show this help message
	@echo "Payment Watchdog Development Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# Development environment
start: ## Start all services with Docker Compose
	./start.sh

stop: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs from all services
	docker-compose logs -f

logs-api: ## Show API service logs
	docker-compose logs -f api

logs-web: ## Show web service logs
	docker-compose logs -f web

logs-worker: ## Show worker service logs
	docker-compose logs -f worker

# Database operations
db-logs: ## Show database logs
	docker-compose logs -f postgres

db-shell: ## Connect to database shell
	docker-compose exec postgres psql -U postgres -d payment_watchdog

db-reset: ## Reset database (WARNING: destroys all data)
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 10
	./start.sh

# API operations
api-build: ## Build API service
	docker-compose build api

api-test: ## Run API tests
	cd api && go test ./...

api-lint: ## Run API linting
	cd api && golangci-lint run

# Web operations
web-build: ## Build web service
	docker-compose build web

web-install: ## Install web dependencies
	cd web && npm install

web-test: ## Run web tests
	cd web && npm test

web-lint: ## Run web linting
	cd web && npm run lint

# Worker operations
worker-build: ## Build worker service
	docker-compose build worker

worker-test: ## Run worker tests
	cd worker && go test ./...

# Development utilities
clean: ## Clean up Docker resources
	docker-compose down -v --remove-orphans
	docker system prune -f

build: ## Build all services
	docker-compose build

test: ## Run tests for all services
	@echo "Running API tests..."
	cd api && go test ./... || exit 1
	@echo "Running Worker tests..."
	cd worker && go test ./... || exit 1
	@echo "Running Web tests..."
	cd web && npm test || exit 1

lint: ## Run linting for all services
	@echo "Linting API..."
	cd api && golangci-lint run || exit 1
	@echo "Linting Web..."
	cd web && npm run lint || exit 1

format: ## Format code for all services
	@echo "Formatting API..."
	cd api && gofmt -w . && go mod tidy
	@echo "Formatting Web..."
	cd web && npm run format

deps: ## Update dependencies for all services
	@echo "Updating API dependencies..."
	cd api && go mod tidy
	@echo "Updating Web dependencies..."
	cd web && npm update

# Production builds
build-prod: ## Build production images
	docker build -t payment-watchdog-api:latest -f api/Dockerfile .
	docker build -t payment-watchdog-web:latest -f web/Dockerfile .
	docker build -t payment-watchdog-worker:latest -f worker/Dockerfile .

# Kubernetes operations (if using k8s)
k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f api/deployments/kubernetes/

k8s-logs: ## Show Kubernetes logs
	kubectl logs -f deployment/payment-watchdog-api

k8s-status: ## Show Kubernetes status
	kubectl get pods,services,ingress -l app=payment-watchdog

# CI/CD simulation
ci: ## Run full CI pipeline locally
	@echo "Running CI pipeline..."
	make lint
	make test
	make build
	@echo "CI pipeline completed successfully!"

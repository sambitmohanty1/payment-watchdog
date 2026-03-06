#!/bin/bash

# Local Docker Deployment Script for Payment Watchdog
# This script manages the local deployment separate from CI/CD flows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.local.yml"
ENV_FILE=".env.local"
PROJECT_NAME="payment-watchdog-local"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker Desktop."
        exit 1
    fi
}

# Function to check if docker-compose is available
check_docker_compose() {
    if ! command -v docker-compose > /dev/null 2>&1; then
        if ! docker compose version > /dev/null 2>&1; then
            print_error "docker-compose is not available. Please install docker-compose."
            exit 1
        else
            DOCKER_COMPOSE="docker compose"
        fi
    else
        DOCKER_COMPOSE="docker-compose"
    fi
}

# Function to create .env.local if it doesn't exist
setup_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        print_warning "$ENV_FILE not found. Creating with default values..."
        cp .env.local.example "$ENV_FILE" 2>/dev/null || {
            print_warning "No .env.local.example found. Creating basic .env.local..."
            cat > "$ENV_FILE" << EOF
# Local Development Environment
ENVIRONMENT=local
LOG_LEVEL=debug
DATABASE_PASSWORD=postgres_local
EOF
        }
        print_success "Created $ENV_FILE. Please review and update if needed."
    fi
}

# Function to start services
start_services() {
    print_status "Starting local Payment Watchdog services..."
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" up -d
    print_success "Services started successfully!"
}

# Function to stop services
stop_services() {
    print_status "Stopping local Payment Watchdog services..."
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" down
    print_success "Services stopped successfully!"
}

# Function to stop and remove volumes
clean_services() {
    print_status "Stopping and cleaning local Payment Watchdog services..."
    print_warning "This will remove all local data including database and Redis data."
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" down -v
        print_success "Services and volumes cleaned successfully!"
    else
        print_status "Cleanup cancelled."
    fi
}

# Function to show service status
show_status() {
    print_status "Local Payment Watchdog service status:"
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" ps
}

# Function to show logs
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        print_status "Showing logs for all services..."
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" logs -f
    else
        print_status "Showing logs for $service..."
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" logs -f "$service"
    fi
}

# Function to rebuild services
rebuild_services() {
    local service=$1
    if [ -z "$service" ]; then
        print_status "Rebuilding all services..."
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" build --no-cache
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" up -d
    else
        print_status "Rebuilding $service..."
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" build --no-cache "$service"
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" --project-name "$PROJECT_NAME" up -d "$service"
    fi
    print_success "Rebuild completed!"
}

# Function to run health checks
health_check() {
    print_status "Running health checks..."
    
    # Check API health
    if curl -f http://localhost:8096/health > /dev/null 2>&1; then
        print_success "API is healthy"
    else
        print_warning "API health check failed"
    fi
    
    # Check Web interface
    if curl -f http://localhost:3016 > /dev/null 2>&1; then
        print_success "Web interface is accessible"
    else
        print_warning "Web interface health check failed"
    fi
    
    # Check database connectivity
    if docker exec payment-watchdog-local-postgres pg_isready -U postgres > /dev/null 2>&1; then
        print_success "Database is ready"
    else
        print_warning "Database health check failed"
    fi
    
    # Check Redis connectivity
    if docker exec payment-watchdog-local-redis redis-cli ping > /dev/null 2>&1; then
        print_success "Redis is ready"
    else
        print_warning "Redis health check failed"
    fi
}

# Function to show access URLs
show_urls() {
    echo ""
    print_success "Local Payment Watchdog Services:"
    echo "  🌐 Web Interface:     http://localhost:3016"
    echo "  🔌 API Endpoint:      http://localhost:8096"
    echo "  📧 MailHog Web UI:    http://localhost:8041"
    echo "  🗄️  Database Admin:    http://localhost:8097 (Adminer)"
    echo "  🔴 Redis Commander:    http://localhost:8098"
    echo "  📊 API Health:        http://localhost:8096/health"
    echo ""
}

# Main script logic
case "${1:-}" in
    "start")
        check_docker
        check_docker_compose
        setup_env_file
        start_services
        sleep 10
        health_check
        show_urls
        ;;
    "stop")
        check_docker
        check_docker_compose
        stop_services
        ;;
    "restart")
        check_docker
        check_docker_compose
        stop_services
        sleep 5
        start_services
        sleep 10
        health_check
        show_urls
        ;;
    "clean")
        check_docker
        check_docker_compose
        clean_services
        ;;
    "status")
        check_docker
        check_docker_compose
        show_status
        ;;
    "logs")
        check_docker
        check_docker_compose
        show_logs "$2"
        ;;
    "rebuild")
        check_docker
        check_docker_compose
        rebuild_services "$2"
        ;;
    "health")
        check_docker
        health_check
        ;;
    "urls")
        show_urls
        ;;
    *)
        echo "Local Payment Watchdog Deployment Script"
        echo ""
        echo "Usage: $0 {start|stop|restart|clean|status|logs|rebuild|health|urls}"
        echo ""
        echo "Commands:"
        echo "  start     - Start all local services"
        echo "  stop      - Stop all local services"
        echo "  restart   - Restart all local services"
        echo "  clean     - Stop services and remove all data"
        echo "  status    - Show service status"
        echo "  logs      - Show logs (optional: specify service name)"
        echo "  rebuild   - Rebuild services (optional: specify service name)"
        echo "  health    - Run health checks"
        echo "  urls      - Show access URLs"
        echo ""
        echo "Examples:"
        echo "  $0 start                    # Start all services"
        echo "  $0 logs payment-watchdog-local-api  # Show API logs"
        echo "  $0 rebuild payment-watchdog-local-web # Rebuild web service"
        exit 1
        ;;
esac

#!/bin/bash

# Payment Watchdog Configuration Validation Script
# Validates Kubernetes manifests for deployment readiness

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INFRA_DIR="$(dirname "$SCRIPT_DIR")/../kubernetes"
VALIDATION_TMP_DIR="/tmp/payment-watchdog-validation"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validation functions
validate_prerequisites() {
    log_info "Validating prerequisites..."
    
    # Check if kustomize is available
    if ! command -v kustomize &> /dev/null; then
        log_error "kustomize is not installed or not in PATH"
        return 1
    fi
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        return 1
    fi
    
    # Check if kubeval is available (optional)
    if command -v kubeval &> /dev/null; then
        log_info "kubeval found - will perform schema validation"
        KUBEVAL_AVAILABLE=true
    else
        log_warning "kubeval not found - skipping schema validation"
        KUBEVAL_AVAILABLE=false
    fi
    
    log_success "Prerequisites validation completed"
    return 0
}

validate_kustomize_build() {
    local environment=$1
    log_info "Validating kustomize build for environment: $environment"
    
    local overlay_dir="$INFRA_DIR/overlays/$environment"
    
    if [[ ! -d "$overlay_dir" ]]; then
        log_error "Environment overlay directory not found: $overlay_dir"
        return 1
    fi
    
    # Create temporary directory for validation
    mkdir -p "$VALIDATION_TMP_DIR"
    
    # Build kustomization
    if ! kustomize build "$overlay_dir" > "$VALIDATION_TMP_DIR/$environment.yaml"; then
        log_error "Kustomize build failed for environment: $environment"
        return 1
    fi
    
    log_success "Kustomize build successful for environment: $environment"
    return 0
}

validate_kubernetes_schema() {
    local environment=$1
    log_info "Validating Kubernetes schemas for environment: $environment"
    
    if [[ "$KUBEVAL_AVAILABLE" != true ]]; then
        log_warning "Skipping schema validation - kubeval not available"
        return 0
    fi
    
    local manifest_file="$VALIDATION_TMP_DIR/$environment.yaml"
    
    if ! kubeval "$manifest_file"; then
        log_error "Kubernetes schema validation failed for environment: $environment"
        return 1
    fi
    
    log_success "Kubernetes schema validation passed for environment: $environment"
    return 0
}

validate_configuration_consistency() {
    local environment=$1
    log_info "Validating configuration consistency for environment: $environment"
    
    local manifest_file="$VALIDATION_TMP_DIR/$environment.yaml"
    
    # Check for required services
    local required_services=("postgres" "redis" "recovery-orchestration")
    
    for service in "${required_services[@]}"; do
        if ! grep -q "name: $service" "$manifest_file"; then
            log_error "Required service not found: $service"
            return 1
        fi
    done
    
    # Check for namespace consistency
    if [[ "$environment" != "base" ]]; then
        if ! grep -q "namespace: $environment" "$manifest_file"; then
            log_error "Namespace not found in manifests: $environment"
            return 1
        fi
    fi
    
    # Check for required labels
    if ! grep -q "app.kubernetes.io/part-of: payment-watchdog" "$manifest_file"; then
        log_error "Required part-of label not found"
        return 1
    fi
    
    log_success "Configuration consistency validation passed for environment: $environment"
    return 0
}

validate_sovereign_compliance() {
    local environment=$1
    log_info "Validating sovereign compliance for environment: $environment"
    
    if [[ "$environment" != "sovereign-au" ]]; then
        log_info "Skipping sovereign compliance check for non-sovereign environment"
        return 0
    fi
    
    local manifest_file="$VALIDATION_TMP_DIR/$environment.yaml"
    
    # Check for network policy
    if ! grep -q "kind: NetworkPolicy" "$manifest_file"; then
        log_error "NetworkPolicy not found in sovereign deployment"
        return 1
    fi
    
    # Check for namespace annotations
    if ! grep -q "data-residency: australia" "$manifest_file"; then
        log_error "Data residency annotation not found"
        return 1
    fi
    
    # Check for Pod Security Standards
    if ! grep -q "pod-security.kubernetes.io/enforce: restricted" "$manifest_file"; then
        log_error "Pod Security Standards not enforced"
        return 1
    fi
    
    # Check for non-Australian endpoints (basic check)
    local non_au_patterns=(
        "us-east-"
        "us-west-"
        "us-central-"
        "eu-west-"
        "eu-central-"
    )
    
    for pattern in "${non_au_patterns[@]}"; do
        if grep -q "$pattern" "$manifest_file"; then
            log_error "Non-Australian endpoint found in sovereign deployment: $pattern"
            return 1
        fi
    done
    
    log_success "Sovereign compliance validation passed for environment: $environment"
    return 0
}

validate_worker_configuration() {
    local environment=$1
    log_info "Validating worker configuration for environment: $environment"
    
    local manifest_file="$VALIDATION_TMP_DIR/$environment.yaml"
    
    # Check for worker service configuration
    if ! grep -q "name: recovery-orchestration" "$manifest_file"; then
        log_error "Worker service not found"
        return 1
    fi
    
    # Check for proper port configuration
    if ! grep -q "containerPort: 8086" "$manifest_file"; then
        log_error "Worker service port 8086 not found"
        return 1
    fi
    
    # Check for health probes
    if ! grep -q "livenessProbe:" "$manifest_file"; then
        log_error "Worker liveness probe not found"
        return 1
    fi
    
    if ! grep -q "readinessProbe:" "$manifest_file"; then
        log_error "Worker readiness probe not found"
        return 1
    fi
    
    # Check for configuration file
    if ! grep -q "config.yaml:" "$manifest_file"; then
        log_error "Worker configuration file not found"
        return 1
    fi
    
    # Check for environment variables
    local required_env_vars=(
        "SERVICE_NAME"
        "DATABASE_HOST"
        "DATABASE_NAME"
        "REDIS_ADDR"
    )
    
    for env_var in "${required_env_vars[@]}"; do
        if ! grep -q "name: $env_var" "$manifest_file"; then
            log_error "Required environment variable not found: $env_var"
            return 1
        fi
    done
    
    log_success "Worker configuration validation passed for environment: $environment"
    return 0
}

generate_validation_report() {
    local environment=$1
    log_info "Generating validation report for environment: $environment"
    
    local manifest_file="$VALIDATION_TMP_DIR/$environment.yaml"
    local report_file="$VALIDATION_TMP_DIR/${environment}-validation-report.txt"
    
    {
        echo "Payment Watchdog Configuration Validation Report"
        echo "=============================================="
        echo "Environment: $environment"
        echo "Generated: $(date)"
        echo ""
        
        echo "Resource Summary:"
        echo "---------------"
        grep "kind:" "$manifest_file" | sort | uniq -c | sort -nr
        echo ""
        
        echo "Services:"
        echo "--------"
        grep -A 5 "kind: Service" "$manifest_file" | grep "name:" | sed 's/.*name: /- /'
        echo ""
        
        echo "Deployments:"
        echo "-----------"
        grep -A 10 "kind: Deployment" "$manifest_file" | grep "name:" | sed 's/.*name: /- /'
        echo ""
        
        echo "Configuration Maps:"
        echo "-----------------"
        grep -A 5 "kind: ConfigMap" "$manifest_file" | grep "name:" | sed 's/.*name: /- /'
        echo ""
        
        if [[ "$environment" == "sovereign-au" ]]; then
            echo "Sovereign Compliance:"
            echo "-------------------"
            if grep -q "data-residency: australia" "$manifest_file"; then
                echo "✓ Data residency: Australia"
            fi
            if grep -q "kind: NetworkPolicy" "$manifest_file"; then
                echo "✓ Network policy: Applied"
            fi
            if grep -q "pod-security.kubernetes.io/enforce: restricted" "$manifest_file"; then
                echo "✓ Pod Security: Restricted"
            fi
            echo ""
        fi
        
    } > "$report_file"
    
    log_success "Validation report generated: $report_file"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    rm -rf "$VALIDATION_TMP_DIR"
}

main() {
    local environment=${1:-"sovereign-au"}
    
    log_info "Starting Payment Watchdog configuration validation"
    log_info "Environment: $environment"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Run validations
    validate_prerequisites || exit 1
    validate_kustomize_build "$environment" || exit 1
    validate_kubernetes_schema "$environment" || exit 1
    validate_configuration_consistency "$environment" || exit 1
    validate_sovereign_compliance "$environment" || exit 1
    validate_worker_configuration "$environment" || exit 1
    
    # Generate report
    generate_validation_report "$environment"
    
    log_success "All validation checks passed for environment: $environment"
    echo ""
    log_info "Deployment is ready for environment: $environment"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

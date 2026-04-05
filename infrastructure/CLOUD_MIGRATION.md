# Cloud Migration Guide

This document provides guidance for migrating Payment Watchdog infrastructure between cloud providers while maintaining the sovereign-au deployment requirements.

## Current State: Oracle Cloud Infrastructure (OCI)

### Current Configuration
- **Provider**: Oracle Cloud Infrastructure (OCI)
- **Region**: ap-melbourne-1 (Australia)
- **Storage Class**: `oci-bv`
- **Compartment**: PaymentwatchdogSovereign
- **Network**: Private VCN with sovereign compliance

### OCI-Specific Components
- Block Volume storage
- OCI Load Balancer
- OCI Identity and Access Management (IAM)
- OCI Container Engine for Kubernetes (OKE)

## Migration Strategy

### Phase 1: Preparation (1-2 weeks)
1. **Assessment**
   - Document current resource usage
   - Identify OCI-specific features in use
   - Map OCI services to equivalent services on target cloud

2. **Infrastructure Updates**
   - Ensure all configurations use cloud-agnostic settings
   - Update Terraform modules (if used)
   - Validate Kubernetes manifests work across clouds

### Phase 2: Target Cloud Setup (2-3 weeks)
1. **Environment Creation**
   - Create VPC/VNet in target cloud
   - Set up Kubernetes cluster
   - Configure storage classes
   - Set up networking and security

2. **Storage Migration**
   - Export data from OCI Block Volumes
   - Import to target cloud storage
   - Update storage class references

### Phase 3: Application Migration (1-2 weeks)
1. **Configuration Updates**
   - Update cloud provider specific settings
   - Modify storage class references
   - Update load balancer configurations
   - Adjust IAM roles and policies

2. **Deployment**
   - Deploy using new infrastructure structure
   - Validate all services start correctly
   - Run integration tests

## Cloud Provider Specifics

### AWS Migration
```yaml
# Storage Class Update
storageClassName: aws-gp3

# IAM Role Update
serviceAccountAnnotations:
  eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/payment-watchdog-worker

# Load Balancer
service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

### GCP Migration
```yaml
# Storage Class Update
storageClassName: gcp-standard

# Service Account
serviceAccountName: payment-watchdog@PROJECT.iam.gserviceaccount.com

# Load Balancer
cloud.google.com/load-balancer-type: "Internal"
```

### Azure Migration
```yaml
# Storage Class Update
storageClassName: azure-standard

# Managed Identity
aadpodidbinding: payment-watchdog-identity

# Load Balancer
service.beta.kubernetes.io/azure-load-balancer-internal: "true"
```

## Sovereign Compliance Considerations

### Australian Data Residency
All target clouds must provide:
- **Australia-based regions**: ap-southeast-2 (Sydney) or similar
- **Data residency guarantees**: Data stored and processed in Australia
- **Compliance certifications**: ISO 27001, SOC 2, Australian privacy compliance

### Network Security
- **Private endpoints**: Ensure services use private connectivity
- **Network policies**: Maintain current network isolation
- **Data encryption**: Encryption at rest and in transit

### Supported Cloud Providers for Sovereign Deployment

#### AWS Australia
- **Region**: ap-southeast-2 (Sydney)
- **Services**: EKS, RDS, ElastiCache, ALB/NLB
- **Compliance**: ISO 27001, SOC 2, IRAP

#### GCP Australia
- **Region**: australia-southeast1 (Sydney)
- **Services**: GKE, Cloud SQL, Memorystore, Cloud Load Balancing
- **Compliance**: ISO 27001, SOC 2, IRAP

#### Azure Australia
- **Region**: Australia East (New South Wales)
- **Services**: AKS, Azure Database, Redis Cache, Azure Load Balancer
- **Compliance**: ISO 27001, SOC 2, IRAP

## Migration Checklist

### Pre-Migration
- [ ] Current infrastructure documented
- [ ] Data backup completed
- [ ] Target cloud account set up
- [ ] Network connectivity tested
- [ ] Security policies reviewed

### Migration
- [ ] Storage classes updated
- [ ] Data migrated successfully
- [ ] Applications deployed
- [ ] Health checks passing
- [ ] Monitoring configured

### Post-Migration
- [ ] Performance validated
- [ ] Security scanned
- [ ] Documentation updated
- [ ] Old resources decommissioned
- [ ] Team trained on new console

## Rollback Plan

### If Migration Fails
1. **Immediate Actions**
   - Switch DNS back to OCI deployment
   - Scale up OCI services
   - Notify stakeholders

2. **Investigation**
   - Analyze failure points
   - Document lessons learned
   - Plan retry with fixes

3. **Retry Preparation**
   - Fix identified issues
   - Update migration procedures
   - Schedule new migration window

## Tools and Automation

### Migration Scripts
```bash
# Cloud migration validation
./infrastructure/scripts/validate-cloud-migration.sh

# Storage class migration
./infrastructure/scripts/migrate-storage-class.sh

# Configuration updates
./infrastructure/scripts/update-cloud-config.sh
```

### Terraform Considerations
If using Terraform for infrastructure:
```hcl
# Cloud provider variable
variable "cloud_provider" {
  description = "Target cloud provider"
  type        = string
  default     = "oci"
}

# Provider configuration
provider "aws" {
  region = var.aws_region
  alias  = "target"
}

# Storage class selection
resource "kubernetes_storage_class" "main" {
  metadata {
    name = var.cloud_provider == "aws" ? "aws-gp3" : "oci-bv"
  }
  
  # Provider-specific configuration
  dynamic "provisioner" {
    for_each = var.cloud_provider == "aws" ? [1] : []
    content {
      value = "kubernetes.io/aws-ebs"
    }
  }
}
```

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Preparation | 1-2 weeks | Documentation review |
| Target Cloud Setup | 2-3 weeks | Cloud account setup |
| Application Migration | 1-2 weeks | Data migration complete |
| Testing & Validation | 1 week | Application deployed |
| Total | 5-8 weeks | Full migration |

## Support Contacts

### Oracle Cloud Support
- **Current Provider**: Oracle
- **Support Channel**: OCI Support Console
- **Documentation**: OCI Documentation

### Target Cloud Support
- **AWS**: AWS Support Center
- **GCP**: Google Cloud Support
- **Azure**: Microsoft Azure Support

## Conclusion

The new infrastructure structure is designed to be cloud-agnostic while maintaining sovereign compliance requirements. The migration process should be methodical and well-planned to minimize downtime and ensure data integrity.

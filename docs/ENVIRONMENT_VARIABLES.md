# Environment Variables

This document describes the environment variables used by the Payment Watchdog services.

## Database Configuration

### Standard Variables (Preferred)
- `DATABASE_HOST` - Database server hostname
- `DATABASE_USER` - Database username  
- `DATABASE_PASSWORD` - Database password
- `DATABASE_NAME` - Database name
- `DATABASE_PORT` - Database port

### Legacy Variables (Supported for Backward Compatibility)
- `DB_HOST` - Alternative database hostname
- `DB_USER` - Alternative database username
- `DB_PASSWORD` - Alternative database password  
- `DB_NAME` - Alternative database name
- `DB_PORT` - Alternative database port

### Variable Priority
- **Standard variables (`DATABASE_*`) take precedence over legacy variables (`DB_*`)**
- If both standard and legacy variables are set, the standard variable will be used
- A warning will be logged if conflicting variables are detected

### Example Configuration

#### Kubernetes ConfigMap (Recommended)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: payment-watchdog-config
data:
  # Standard variables (preferred)
  DATABASE_HOST: "postgres.example.com"
  DATABASE_NAME: "payment_watchdog"
  DATABASE_USER: "postgres"
  DATABASE_PASSWORD: "your-password"
  DATABASE_PORT: "5432"
  
  # Legacy variables (optional, for backward compatibility)
  # DB_HOST: "postgres.example.com"
  # DB_NAME: "payment_watchdog"
  # DB_USER: "postgres"
  # DB_PASSWORD: "your-password"
  # DB_PORT: "5432"
```

#### Environment File
```bash
# Standard variables
export DATABASE_HOST="postgres.example.com"
export DATABASE_NAME="payment_watchdog"
export DATABASE_USER="postgres"
export DATABASE_PASSWORD="your-password"
export DATABASE_PORT="5432"

# Legacy variables (still supported)
export DB_HOST="postgres.example.com"
export DB_NAME="payment_watchdog"
export DB_USER="postgres"
export DB_PASSWORD="your-password"
export DB_PORT="5432"
```

## Service Configuration

### API Service
- `SERVER_PORT` - API server port (default: 8085)
- `SERVER_HOST` - API server host (default: 0.0.0.0)
- `SERVER_HTTPS` - Enable HTTPS (default: false)
- `SERVER_CERT_FILE` - SSL certificate file path
- `SERVER_KEY_FILE` - SSL private key file path

### Logging
- `LOG_LEVEL` - Log level (debug, info, warn, error) (default: info)

### Sovereign Mode
- `SOVEREIGN_MODE` - Enable sovereign compliance checks (default: false)

## Conflict Detection

The system will detect and warn about conflicting environment variables:

- Both `DATABASE_HOST` and `DB_HOST` set
- Both `DATABASE_USER` and `DB_USER` set
- Both `DATABASE_PASSWORD` and `DB_PASSWORD` set
- Both `DATABASE_NAME` and `DB_NAME` set
- Both `DATABASE_PORT` and `DB_PORT` set

When conflicts are detected, a warning will be logged but the application will continue using the standard variable.

## Migration Guide

### From Legacy to Standard Variables

If you're currently using `DB_*` variables, migrate to `DATABASE_*`:

```bash
# Before (legacy)
export DB_HOST="postgres.example.com"
export DB_NAME="payment_watchdog"
export DB_USER="postgres"
export DB_PASSWORD="your-password"
export DB_PORT="5432"

# After (standard)
export DATABASE_HOST="postgres.example.com"
export DATABASE_NAME="payment_watchdog"
export DATABASE_USER="postgres"
export DATABASE_PASSWORD="your-password"
export DATABASE_PORT="5432"
```

### Kubernetes ConfigMap Migration

```yaml
# Before (legacy)
data:
  DB_HOST: "postgres.example.com"
  DB_NAME: "payment_watchdog"
  DB_USER: "postgres"
  DB_PASSWORD: "your-password"
  DB_PORT: "5432"

# After (standard)
data:
  DATABASE_HOST: "postgres.example.com"
  DATABASE_NAME: "payment_watchdog"
  DATABASE_USER: "postgres"
  DATABASE_PASSWORD: "your-password"
  DATABASE_PORT: "5432"
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check that `DATABASE_HOST` is correct
   - Verify `DATABASE_USER` and `DATABASE_PASSWORD` are set
   - Ensure `DATABASE_NAME` exists

2. **Sovereign Compliance Failed**
   - Check that `DATABASE_HOST` contains `.au` or `svc.cluster.local`
   - Verify `SOVEREIGN_MODE` is set correctly

3. **Variable Conflicts**
   - Check logs for "environment variable conflicts detected" warnings
   - Remove conflicting variables or choose one naming convention

### Debug Commands

```bash
# Check environment variables in pod
kubectl exec <pod-name> -- env | grep DATABASE

# Check configuration loading
kubectl logs <pod-name> | grep -i "config\|environment"

# Test database connection
kubectl exec <pod-name> -- nc -zv <DATABASE_HOST> <DATABASE_PORT>
```

## Support

For questions about environment variables or configuration issues:
- Check the application logs for detailed error messages
- Refer to this documentation for variable naming conventions
- Contact the development team for assistance

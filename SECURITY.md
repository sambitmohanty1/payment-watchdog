# Security Policy

## Security Architecture

Payment Watchdog implements a comprehensive security scanning and monitoring pipeline through GitHub Actions, providing continuous security assessment across all components.

### Security Scanning Pipeline

#### 1. Static Application Security Testing (SAST)
- **GitHub CodeQL**: Advanced SAST analysis for Go and JavaScript/TypeScript
  - Runs on every push/PR to main/develop branches
  - Scheduled daily scans at 2 AM UTC
  - Extended security and quality queries enabled
  - Results uploaded to GitHub Security tab

#### 2. Container Security Scanning
- **Trivy**: Multi-layer vulnerability scanning
  - File system scanning for all code
  - Docker image scanning for API, Worker, and Web services
  - SARIF format reporting with CRITICAL, HIGH, MEDIUM severity levels
  - Automatic artifact retention for 30 days

#### 3. Go-Specific Security Analysis
- **Gosec**: Go static security analysis
  - Scans both API and Worker services
  - SARIF output integration with GitHub Security
- **govulncheck**: Go vulnerability database checking
  - Database of known Go vulnerabilities
  - Detailed vulnerability reporting

#### 4. Node.js Security Analysis
- **npm audit**: Dependency vulnerability scanning
  - High-severity vulnerability detection
  - JSON reporting with artifact retention
  - Package lock file analysis

### Security Gates

#### Deployment Protection
- **Security Gate Check**: Blocks deployment on critical security failures
  - Evaluates CodeQL and Trivy scan results
  - Prevents main branch deployments with critical issues
  - Automated approval for passed security scans

#### Continuous Monitoring
- **Daily Scheduled Scans**: Automated security assessment
- **PR Security Validation**: All pull requests scanned before merge
- **Dependency Updates**: Automated Dependabot security patches

### Security Metrics & Reporting

#### GitHub Security Tab Integration
- All SARIF findings automatically uploaded
- Centralized security vulnerability tracking
- Historical security trend analysis

#### Action Summaries
- Real-time security scan status
- Comprehensive security metrics dashboard
- Artifact access for detailed analysis

### Supported Versions

| Version | Security Support | EOL Status |
|---------|------------------|------------|
| 1.0.x   | ✅ Active        | Current    |
| < 1.0   | ❌ End of Life   | Unsupported |

### Reporting a Vulnerability

#### Security Contact
- **Private Disclosure**: Create a security advisory via GitHub's private vulnerability reporting
- **Email Security Issues**: Use private GitHub security reporting functionality
- **Critical Issues**: Mark as high priority for immediate response

#### Response Process
1. **Acknowledgment**: Within 24 hours of receipt
2. **Assessment**: Initial triage within 48 hours
3. **Remediation**: Patch development based on severity
4. **Disclosure**: Coordinated disclosure after fix deployment

#### Severity Classification
- **Critical**: Immediate deployment blocking, fix within 24 hours
- **High**: Review required, fix within 72 hours
- **Medium**: Track for remediation, fix in next release cycle
- **Low**: Documentation updates, future consideration

### Security Best Practices Implemented

#### Infrastructure Security
- Minimal Docker base images
- Non-root container execution
- Secrets management through environment variables
- Network segmentation between services

#### Code Security
- Input validation and sanitization
- SQL injection prevention via GORM
- Authentication and authorization controls
- Secure dependency management

#### Operational Security
- Regular security scanning pipeline
- Automated dependency updates
- Security-focused code reviews
- Incident response procedures

---

*Last updated: March 2026*

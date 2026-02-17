# GitHub Actions Environment Variables Template
# Copy this to your GitHub repository Settings > Secrets and variables > Actions

# Kubernetes Deployment Secrets
KUBE_CONFIG=base64-encoded-kubeconfig-content
# Get this with: kubectl config view --raw | base64 -w 0

# Slack Notifications (Optional)
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# Database Configuration (for staging)
STAGING_DB_PASSWORD=your_secure_db_password
STRIPE_SECRET_KEY=sk_test_your_stripe_key

# Container Registry (automatically set by GitHub Actions)
# No additional secrets needed for GHCR - uses GITHUB_TOKEN automatically
REGISTRY=ghcr.io
IMAGE_NAME=sambitmohanty1/payment-watchdog
TAG=main

# Kubernetes Configuration
KUBE_CONFIG=base64-encoded-kubeconfig-content

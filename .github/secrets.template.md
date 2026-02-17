# GitHub Actions Environment Variables Template
# Copy this to your GitHub repository Settings > Secrets and variables > Actions

# Required Secrets
STAGING_HOST=your-server-domain.com
STAGING_USER=deploy
STAGING_SSH_KEY=-----BEGIN OPENSSH PRIVATE KEY-----
# Your private SSH key content here
-----END OPENSSH PRIVATE KEY-----

# Database Configuration
STAGING_DB_PASSWORD=your_secure_db_password
STRIPE_SECRET_KEY=sk_test_your_stripe_key

# Container Registry (automatically set by GitHub Actions)
REGISTRY=ghcr.io
IMAGE_NAME=your-username/payment-watchdog
TAG=main

#!/bin/bash
# Script to setup Google Cloud Platform Infrastructure for Sovereign Data Compliance

REGION=${REGION:-australia-southeast1}

if [[ "$REGION" != "australia-southeast1" && "$REGION" != "australia-southeast2" ]]; then
    echo "❌ ERROR: For sovereign compliance, REGION must be 'australia-southeast1' or 'australia-southeast2'."
    exit 1
fi

echo "🚀 Starting GCP Sovereign Infrastructure Provisioning in region: $REGION"

PROJECT_ID="payment-watchdog-au"

# Setup Cloud SQL
echo "Creating Cloud SQL (Postgres) instance in $REGION..."
gcloud sql instances create payment-watchdog-pg \
    --database-version=POSTGRES_14 \
    --cpu=2 --memory=8GiB \
    --region=$REGION \
    --project=$PROJECT_ID \
    --require-ssl \
    --availability-type=REGIONAL

# Setup Memorystore (Redis)
echo "Creating Memorystore (Redis) instance in $REGION..."
gcloud redis instances create payment-watchdog-redis \
    --size=2 \
    --region=$REGION \
    --project=$PROJECT_ID \
    --redis-version=redis_6_x \
    --tier=STANDARD_HA

echo "✅ GCP Sovereign Setup Complete."

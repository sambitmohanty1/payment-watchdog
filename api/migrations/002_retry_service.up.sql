-- Migration 002: Add retry service tables for Sprint 2
-- This migration adds the retry_jobs table for the enhanced retry service

-- Create retry_jobs table for Sprint 2 retry service
CREATE TABLE retry_jobs (
    id VARCHAR(255) PRIMARY KEY,
    job_type VARCHAR(100) NOT NULL,
    company_id VARCHAR(255) NOT NULL,
    data JSONB DEFAULT '{}',
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    status VARCHAR(50) DEFAULT 'pending',
    error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for retry_jobs table
CREATE INDEX idx_retry_jobs_company ON retry_jobs(company_id);
CREATE INDEX idx_retry_jobs_status ON retry_jobs(status);
CREATE INDEX idx_retry_jobs_next_retry ON retry_jobs(next_retry_at);
CREATE INDEX idx_retry_jobs_created ON retry_jobs(created_at);

-- Add comment to table
COMMENT ON TABLE retry_jobs IS 'Sprint 2 retry service jobs with exponential backoff and dead letter queue support';

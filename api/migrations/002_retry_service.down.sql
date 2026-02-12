-- Migration 002: Rollback retry service tables
-- This migration removes the retry_jobs table added in Sprint 2

-- Drop retry_jobs table
DROP TABLE IF EXISTS retry_jobs;

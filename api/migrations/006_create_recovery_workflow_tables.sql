-- Migration: Create Recovery Workflow Tables
-- Description: Creates tables for the Recovery Orchestration Engine
-- Version: 006
-- Date: 2025-09-17

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Recovery Workflows table
CREATE TABLE IF NOT EXISTS recovery_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 1,
    trigger_conditions JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- Recovery Workflow Steps table
CREATE TABLE IF NOT EXISTS recovery_workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES recovery_workflows(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    description TEXT,
    config JSONB,
    conditions JSONB,
    delay_minutes INTEGER DEFAULT 0,
    is_parallel BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Recovery Workflow Executions table
CREATE TABLE IF NOT EXISTS recovery_workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES recovery_workflows(id) ON DELETE CASCADE,
    payment_failure_id UUID NOT NULL REFERENCES payment_failure_events(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step_id UUID REFERENCES recovery_workflow_steps(id),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    paused_at TIMESTAMP WITH TIME ZONE,
    total_steps INTEGER DEFAULT 0,
    completed_steps INTEGER DEFAULT 0,
    failed_steps INTEGER DEFAULT 0,
    successful_steps INTEGER DEFAULT 0,
    execution_log JSONB,
    last_error TEXT,
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Recovery Step Executions table
CREATE TABLE IF NOT EXISTS recovery_step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_execution_id UUID NOT NULL REFERENCES recovery_workflow_executions(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES recovery_workflow_steps(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT,
    result JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    action_type VARCHAR(100),
    action_data JSONB,
    external_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Communication Templates table
CREATE TABLE IF NOT EXISTS communication_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template_type VARCHAR(50) NOT NULL,
    subject VARCHAR(500),
    content TEXT NOT NULL,
    variables JSONB,
    conditions JSONB,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Recovery Actions table
CREATE TABLE IF NOT EXISTS recovery_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    payment_failure_id UUID NOT NULL REFERENCES payment_failure_events(id) ON DELETE CASCADE,
    workflow_execution_id UUID REFERENCES recovery_workflow_executions(id) ON DELETE SET NULL,
    step_execution_id UUID REFERENCES recovery_step_executions(id) ON DELETE SET NULL,
    action_type VARCHAR(100) NOT NULL,
    action_data JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    provider VARCHAR(100),
    external_id VARCHAR(255),
    result JSONB,
    error_message TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Recovery Metrics table
CREATE TABLE IF NOT EXISTS recovery_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    period_type VARCHAR(20) NOT NULL,
    total_workflow_executions INTEGER DEFAULT 0,
    successful_workflow_executions INTEGER DEFAULT 0,
    failed_workflow_executions INTEGER DEFAULT 0,
    average_execution_time_minutes DECIMAL(10,2) DEFAULT 0,
    total_recovery_actions INTEGER DEFAULT 0,
    successful_recovery_actions INTEGER DEFAULT 0,
    recovery_success_rate DECIMAL(5,2) DEFAULT 0,
    total_amount_recovered DECIMAL(15,2) DEFAULT 0,
    emails_sent INTEGER DEFAULT 0,
    sms_sent INTEGER DEFAULT 0,
    communication_response_rate DECIMAL(5,2) DEFAULT 0,
    average_time_to_recovery_hours DECIMAL(10,2) DEFAULT 0,
    first_attempt_success_rate DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_recovery_workflows_company_id ON recovery_workflows(company_id);
CREATE INDEX IF NOT EXISTS idx_recovery_workflows_is_active ON recovery_workflows(is_active);
CREATE INDEX IF NOT EXISTS idx_recovery_workflows_priority ON recovery_workflows(priority);

CREATE INDEX IF NOT EXISTS idx_recovery_workflow_steps_workflow_id ON recovery_workflow_steps(workflow_id);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_steps_step_order ON recovery_workflow_steps(step_order);

CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_workflow_id ON recovery_workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_payment_failure_id ON recovery_workflow_executions(payment_failure_id);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_company_id ON recovery_workflow_executions(company_id);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_status ON recovery_workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_current_step_id ON recovery_workflow_executions(current_step_id);

CREATE INDEX IF NOT EXISTS idx_recovery_step_executions_workflow_execution_id ON recovery_step_executions(workflow_execution_id);
CREATE INDEX IF NOT EXISTS idx_recovery_step_executions_step_id ON recovery_step_executions(step_id);
CREATE INDEX IF NOT EXISTS idx_recovery_step_executions_status ON recovery_step_executions(status);
CREATE INDEX IF NOT EXISTS idx_recovery_step_executions_external_id ON recovery_step_executions(external_id);

CREATE INDEX IF NOT EXISTS idx_communication_templates_company_id ON communication_templates(company_id);
CREATE INDEX IF NOT EXISTS idx_communication_templates_template_type ON communication_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_communication_templates_is_active ON communication_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_communication_templates_is_default ON communication_templates(is_default);

CREATE INDEX IF NOT EXISTS idx_recovery_actions_company_id ON recovery_actions(company_id);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_payment_failure_id ON recovery_actions(payment_failure_id);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_workflow_execution_id ON recovery_actions(workflow_execution_id);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_step_execution_id ON recovery_actions(step_execution_id);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_action_type ON recovery_actions(action_type);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_status ON recovery_actions(status);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_provider ON recovery_actions(provider);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_external_id ON recovery_actions(external_id);

CREATE INDEX IF NOT EXISTS idx_recovery_metrics_company_id ON recovery_metrics(company_id);
CREATE INDEX IF NOT EXISTS idx_recovery_metrics_period_start ON recovery_metrics(period_start);
CREATE INDEX IF NOT EXISTS idx_recovery_metrics_period_end ON recovery_metrics(period_end);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_recovery_workflows_company_active_priority ON recovery_workflows(company_id, is_active, priority DESC);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_steps_workflow_order ON recovery_workflow_steps(workflow_id, step_order);
CREATE INDEX IF NOT EXISTS idx_recovery_workflow_executions_company_status ON recovery_workflow_executions(company_id, status);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_company_type_status ON recovery_actions(company_id, action_type, status);

-- Add triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_recovery_workflows_updated_at BEFORE UPDATE ON recovery_workflows FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recovery_workflow_steps_updated_at BEFORE UPDATE ON recovery_workflow_steps FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recovery_workflow_executions_updated_at BEFORE UPDATE ON recovery_workflow_executions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recovery_step_executions_updated_at BEFORE UPDATE ON recovery_step_executions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_communication_templates_updated_at BEFORE UPDATE ON communication_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recovery_actions_updated_at BEFORE UPDATE ON recovery_actions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recovery_metrics_updated_at BEFORE UPDATE ON recovery_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add constraints for data integrity
ALTER TABLE recovery_workflow_steps ADD CONSTRAINT chk_step_type CHECK (step_type IN ('retry_payment', 'send_email', 'send_sms', 'wait', 'conditional', 'webhook'));
ALTER TABLE recovery_workflow_executions ADD CONSTRAINT chk_execution_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'paused', 'cancelled'));
ALTER TABLE recovery_step_executions ADD CONSTRAINT chk_step_execution_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped'));
ALTER TABLE communication_templates ADD CONSTRAINT chk_template_type CHECK (template_type IN ('email', 'sms', 'in_app', 'webhook'));
ALTER TABLE recovery_actions ADD CONSTRAINT chk_action_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'));
ALTER TABLE recovery_metrics ADD CONSTRAINT chk_period_type CHECK (period_type IN ('daily', 'weekly', 'monthly', 'quarterly', 'yearly'));

-- Add unique constraints
ALTER TABLE recovery_workflow_steps ADD CONSTRAINT uk_workflow_step_order UNIQUE (workflow_id, step_order);
ALTER TABLE communication_templates ADD CONSTRAINT uk_company_template_name UNIQUE (company_id, name);
ALTER TABLE recovery_metrics ADD CONSTRAINT uk_company_period UNIQUE (company_id, period_start, period_end, period_type);

COMMENT ON TABLE recovery_workflows IS 'Automated recovery workflows for payment failures';
COMMENT ON TABLE recovery_workflow_steps IS 'Individual steps within recovery workflows';
COMMENT ON TABLE recovery_workflow_executions IS 'Execution instances of recovery workflows';
COMMENT ON TABLE recovery_step_executions IS 'Execution instances of individual workflow steps';
COMMENT ON TABLE communication_templates IS 'Templates for customer communications';
COMMENT ON TABLE recovery_actions IS 'Individual recovery actions taken';
COMMENT ON TABLE recovery_metrics IS 'Aggregated recovery performance metrics';

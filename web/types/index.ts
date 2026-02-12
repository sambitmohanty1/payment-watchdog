// Core types for Lexure Intelligence MVP

export interface Company {
  id: string;
  name: string;
  domain?: string;
  status: 'active' | 'inactive' | 'suspended';
  stripe_account_id?: string;
  alert_settings: AlertSettings;
  retry_settings: RetrySettings;
  created_at: string;
  updated_at: string;
}

export interface AlertSettings {
  email_enabled: boolean;
  sms_enabled: boolean;
  slack_enabled: boolean;
  alert_throttling: boolean;
  throttle_minutes: number;
}

export interface RetrySettings {
  auto_retry_enabled: boolean;
  max_retry_attempts: number;
  retry_delay_minutes: number;
  retry_strategy: 'immediate' | 'delayed' | 'smart';
}

export interface PaymentFailureEvent {
  id: string;
  company_id: string;
  provider_id: string;
  event_id: string;
  event_type: string;
  payment_intent_id?: string;
  amount: number;
  currency: string;
  customer_id?: string;
  customer_email?: string;
  customer_name?: string;
  failure_reason: string;
  failure_code?: string;
  failure_message?: string;
  status: 'received' | 'processing' | 'resolved' | 'escalated';
  processed_at?: string;
  alerted_at?: string;
  raw_event_data: Record<string, any>;
  normalized_data: Record<string, any>;
  webhook_received_at: string;
  created_at: string;
  updated_at: string;
}

export interface RetryAttempt {
  id: string;
  payment_failure_id: string;
  company_id: string;
  attempt_number: number;
  retry_amount?: number;
  retry_method?: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  provider_retry_id?: string;
  provider_response: Record<string, any>;
  scheduled_at?: string;
  attempted_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CustomerCommunication {
  id: string;
  payment_failure_id: string;
  company_id: string;
  channel: 'email' | 'sms' | 'push' | 'in_app';
  template_id?: string;
  subject?: string;
  content: string;
  status: 'pending' | 'sent' | 'delivered' | 'opened' | 'clicked' | 'failed';
  provider_message_id?: string;
  delivery_response: Record<string, any>;
  sent_at?: string;
  delivered_at?: string;
  opened_at?: string;
  clicked_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DashboardStats {
  payment_failures: {
    total: number;
    total_amount: number;
    by_status: any[];
    by_reason: any[];
    by_provider: any[];
    daily_breakdown: any[];
  };
  alerts: {
    total: number;
    by_status: any[];
    by_channel: any[];
  };
  retries: {
    total: number;
    success_rate: number;
    by_status: any[];
  };
  last_updated: string;
}

export interface Alert {
  id: string;
  company_id: string;
  type: 'payment_failure' | 'retry_success' | 'retry_failure' | 'system';
  title: string;
  message: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: 'unread' | 'read' | 'acknowledged';
  action_required: boolean;
  action_url?: string;
  created_at: string;
  read_at?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface FilterOptions {
  date_from?: string;
  date_to?: string;
  status?: string;
  customer_id?: string;
  failure_reason?: string;
  amount_min?: number;
  amount_max?: number;
}

export interface SortOptions {
  field: string;
  direction: 'asc' | 'desc';
}

export interface RetryAction {
  payment_failure_id: string;
  retry_amount?: number;
  retry_method?: string;
  scheduled_at?: string;
  customer_notification: boolean;
  notification_template?: string;
}

export interface CommunicationTemplate {
  id: string;
  name: string;
  subject: string;
  content: string;
  channel: 'email' | 'sms';
  variables: string[];
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

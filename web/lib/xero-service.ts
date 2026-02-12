/**
 * Xero Service - Uses the backend mediator pattern for Xero integration
 * This service calls the backend API endpoints that use the XeroMediator
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8085';

export interface XeroOAuthRequest {
  company_id: string;
}

export interface XeroOAuthResponse {
  authorization_url: string;
  state: string;
}

export interface XeroCallbackRequest {
  code: string;
  state: string;
}

export interface XeroCallbackResponse {
  success: boolean;
  message: string;
  access_token?: string;
  refresh_token?: string;
  expires_at?: string;
}

export interface XeroTenant {
  id: string;
  name: string;
  short_code: string;
  is_active: boolean;
  created_date: string;
}

export interface XeroTenantsResponse {
  success: boolean;
  tenants: XeroTenant[];
  error?: string;
}

export interface XeroOrganization {
  id: string;
  name: string;
  legal_name: string;
  short_code: string;
  country_code: string;
  base_currency: string;
  is_active: boolean;
}

export interface XeroOrganizationsResponse {
  success: boolean;
  organizations: XeroOrganization[];
  error?: string;
}

export interface XeroPaymentFailure {
  id: string;
  invoice_number: string;
  customer_name: string;
  customer_email: string;
  amount: number;
  currency: string;
  due_date: string;
  days_overdue: number;
  failure_reason: string;
  status: string;
}

export interface XeroPaymentFailuresResponse {
  success: boolean;
  payment_failures: XeroPaymentFailure[];
  total_count: number;
  error?: string;
}

class XeroService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = `${API_BASE_URL}/api/v1/xero`;
  }

  /**
   * Get Xero OAuth authorization URL
   */
  async getAuthorizationUrl(companyId: string): Promise<XeroOAuthResponse> {
    const response = await fetch(`${this.baseUrl}/auth/authorize`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ company_id: companyId }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to get authorization URL');
    }

    return response.json();
  }

  /**
   * Handle Xero OAuth callback
   */
  async handleCallback(code: string, state: string): Promise<XeroCallbackResponse> {
    const response = await fetch(`${this.baseUrl}/auth/callback`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code, state }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to handle OAuth callback');
    }

    return response.json();
  }

  /**
   * Get Xero tenant connections
   */
  async getTenants(accessToken: string): Promise<XeroTenantsResponse> {
    const response = await fetch(`${this.baseUrl}/tenants`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to get tenants');
    }

    return response.json();
  }

  /**
   * Get Xero organization details
   */
  async getOrganizations(accessToken: string, tenantId: string): Promise<XeroOrganizationsResponse> {
    const response = await fetch(`${this.baseUrl}/organizations?tenant_id=${tenantId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to get organizations');
    }

    return response.json();
  }

  /**
   * Get payment failures from Xero
   */
  async getPaymentFailures(accessToken: string, tenantId: string): Promise<XeroPaymentFailuresResponse> {
    const response = await fetch(`${this.baseUrl}/payment-failures?tenant_id=${tenantId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to get payment failures');
    }

    return response.json();
  }

  /**
   * Test Xero API connectivity
   */
  async testConnection(accessToken: string, tenantId: string): Promise<XeroOrganizationsResponse> {
    return this.getOrganizations(accessToken, tenantId);
  }
}

// Export singleton instance
export const xeroService = new XeroService();
export default xeroService;

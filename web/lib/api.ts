import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
  ApiResponse,
  PaginatedResponse,
  PaymentFailureEvent,
  RetryAttempt,
  CustomerCommunication,
  DashboardStats,
  Alert,
  Company,
  FilterOptions,
  SortOptions,
  RetryAction,
  CommunicationTemplate,
} from '@/types';

class ApiClient {
  private client: AxiosInstance;
  private companyId: string | null = null;

  constructor() {
    // Configure base URL with proxy bypass considerations
    const baseURL = process.env.NODE_ENV === 'development' 
      ? '/api/backend' 
      : '/api/backend';

    this.client = axios.create({
      baseURL,
      timeout: 15000, // Increased timeout for proxy issues
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache',
        'Pragma': 'no-cache',
      },
      // Proxy bypass configuration
      proxy: false, // Disable axios proxy
      maxRedirects: 0, // Prevent redirect loops
    });

    // Request interceptor to add company ID and handle proxy issues
    this.client.interceptors.request.use((config) => {
      if (this.companyId && config.url) {
        // Always add company_id to query parameters for API calls
        const separator = config.url.includes('?') ? '&' : '?';
        config.url = `${config.url}${separator}company_id=${this.companyId}`;
      }
      
      // Add proxy bypass headers (excluding unsafe headers that browsers block)
      if (config.headers) {
        config.headers['X-Proxy-Bypass'] = 'localhost,127.0.0.1,::1';
        // Note: 'Connection' header is blocked by browsers for security reasons
      }
      
      return config;
    });

    // Response interceptor for error handling with proxy awareness
    this.client.interceptors.response.use(
      (response: AxiosResponse) => response,
      (error) => {
        console.error('API Error:', error);
        
        // Handle proxy-related errors specifically
        if (error.code === 'ECONNREFUSED' || error.code === 'ENOTFOUND') {
          console.error('Proxy connection error - check proxy configuration');
          error.message = 'Proxy connection error - check proxy configuration';
        }
        
        if (error.response?.status === 401) {
          // Handle unauthorized access
          window.location.href = '/login';
        }
        
        // Add proxy error context
        if (error.message.includes('timeout')) {
          error.message = 'Request timeout - possible proxy issue';
        }
        
        return Promise.reject(error);
      }
    );
  }

  setCompanyId(companyId: string) {
    this.companyId = companyId;
  }

  private async request<T>(config: any): Promise<ApiResponse<T>> {
    try {
      const response = await this.client.request(config);
      // The Next.js API routes return the data directly, so we need to wrap it
      return {
        success: true,
        data: response.data,
      };
    } catch (error: any) {
      // Enhanced error handling for proxy issues
      let errorMessage = error.response?.data?.error || error.message || 'Unknown error occurred';
      
      if (error.code === 'ECONNREFUSED') {
        errorMessage = 'Connection refused - check port forwarding and proxy settings';
      } else if (error.code === 'ENOTFOUND') {
        errorMessage = 'Service not found - check proxy configuration';
      } else if (error.message.includes('timeout')) {
        errorMessage = 'Request timeout - possible proxy interference';
      }
      
      return {
        success: false,
        error: errorMessage,
      };
    }
  }

  // Health check with proxy awareness
  async healthCheck(): Promise<ApiResponse<{ status: string; service: string; timestamp: string }>> {
    return this.request({
      method: 'GET',
      url: '/health',
      timeout: 20000, // Longer timeout for health checks
    });
  }

  // Payment Failures
  async getPaymentFailures(
    filters?: FilterOptions,
    sort?: SortOptions,
    page: number = 1,
    limit: number = 20
  ): Promise<ApiResponse<PaginatedResponse<PaymentFailureEvent>>> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, value.toString());
        }
      });
    }
    if (sort) {
      params.append('sort_field', sort.field);
      params.append('sort_direction', sort.direction);
    }
    params.append('page', page.toString());
    params.append('limit', limit.toString());

    return this.request({
      method: 'GET',
      url: `/failures?${params.toString()}`,
      timeout: 30000, // Longer timeout for data requests
    });
  }

  async getPaymentFailure(id: string): Promise<ApiResponse<PaymentFailureEvent>> {
    return this.request({
      method: 'GET',
      url: `/failures/${id}`,
    });
  }

  async retryPayment(id: string, retryData: RetryAction): Promise<ApiResponse<RetryAttempt>> {
    return this.request({
      method: 'POST',
      url: `/failures/${id}/retry`,
      data: retryData,
    });
  }

  // Dashboard Stats
  async getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
    return this.request({
      method: 'GET',
      url: '/dashboard/stats',
      timeout: 30000,
    });
  }

  async exportData(type: string, filters?: FilterOptions): Promise<Blob> {
    const params = new URLSearchParams();
    params.append('type', type);
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, value.toString());
        }
      });
    }

    const response = await this.client.get(`/dashboard/export?${params.toString()}`, {
      responseType: 'blob',
    });
    return response.data;
  }

  // Alerts
  async getAlerts(
    page: number = 1,
    limit: number = 20
  ): Promise<ApiResponse<PaginatedResponse<Alert>>> {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('limit', limit.toString());

    return this.request({
      method: 'GET',
      url: `/alerts?${params.toString()}`,
      timeout: 30000,
    });
  }

  async getAlert(id: string): Promise<ApiResponse<Alert>> {
    return this.request({
      method: 'GET',
      url: `/api/v1/alerts/${id}`,
    });
  }

  async markAlertAsRead(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request({
      method: 'PATCH',
      url: `/api/v1/alerts/${id}/read`,
    });
  }

  // Company Management
  async getCompany(): Promise<ApiResponse<Company>> {
    return this.request({
      method: 'GET',
      url: '/api/v1/company',
    });
  }

  async updateCompany(companyData: Partial<Company>): Promise<ApiResponse<Company>> {
    return this.request({
      method: 'PUT',
      url: '/api/v1/company',
      data: companyData,
    });
  }

  // Communication Templates
  async getCommunicationTemplates(): Promise<ApiResponse<CommunicationTemplate[]>> {
    return this.request({
      method: 'GET',
      url: '/api/v1/communication/templates',
    });
  }

  async createCommunicationTemplate(
    template: Omit<CommunicationTemplate, 'id' | 'created_at' | 'updated_at'>
  ): Promise<ApiResponse<CommunicationTemplate>> {
    return this.request({
      method: 'POST',
      url: '/api/v1/communication/templates',
      data: template,
    });
  }

  async updateCommunicationTemplate(
    id: string,
    template: Partial<CommunicationTemplate>
  ): Promise<ApiResponse<CommunicationTemplate>> {
    return this.request({
      method: 'PUT',
      url: `/api/v1/communication/templates/${id}`,
      data: template,
    });
  }

  async deleteCommunicationTemplate(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return this.request({
      method: 'DELETE',
      url: `/api/v1/communication/templates/${id}`,
    });
  }

  // Retry Attempts
  async getRetryAttempts(
    paymentFailureId?: string,
    page: number = 1,
    limit: number = 20
  ): Promise<ApiResponse<PaginatedResponse<RetryAttempt>>> {
    const params = new URLSearchParams();
    if (paymentFailureId) {
      params.append('payment_failure_id', paymentFailureId);
    }
    params.append('page', page.toString());
    params.append('limit', limit.toString());

    return this.request({
      method: 'GET',
      url: `/api/v1/retry-attempts?${params.toString()}`,
    });
  }

  // Customer Communications
  async getCustomerCommunications(
    paymentFailureId?: string,
    page: number = 1,
    limit: number = 20
  ): Promise<ApiResponse<PaginatedResponse<CustomerCommunication>>> {
    const params = new URLSearchParams();
    if (paymentFailureId) {
      params.append('payment_failure_id', paymentFailureId);
    }
    params.append('page', page.toString());
    params.append('limit', limit.toString());

    return this.request({
      method: 'GET',
      url: `/api/v1/communications?${params.toString()}`,
    });
  }

  async sendCustomerCommunication(
    communication: Omit<CustomerCommunication, 'id' | 'created_at' | 'updated_at'>
  ): Promise<ApiResponse<CustomerCommunication>> {
    return this.request({
      method: 'POST',
      url: '/api/v1/communications',
      data: communication,
    });
  }
}

export default new ApiClient();

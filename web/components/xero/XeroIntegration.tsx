'use client';

import React, { useState } from 'react';
import { xeroService, XeroTenant, XeroOrganization, XeroPaymentFailure } from '@/lib/xero-service';

interface XeroIntegrationProps {
  companyId: string;
}

export default function XeroIntegration({ companyId }: XeroIntegrationProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [tenants, setTenants] = useState<XeroTenant[]>([]);
  const [organizations, setOrganizations] = useState<XeroOrganization[]>([]);
  const [paymentFailures, setPaymentFailures] = useState<XeroPaymentFailure[]>([]);
  const [selectedTenantId, setSelectedTenantId] = useState<string>('');

  const handleStartOAuth = async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await xeroService.getAuthorizationUrl(companyId);
      
      // Redirect to Xero authorization URL
      window.location.href = response.authorization_url;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start OAuth flow');
    } finally {
      setIsLoading(false);
    }
  };

  const handleGetTenants = async () => {
    if (!accessToken) {
      setError('Access token required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await xeroService.getTenants(accessToken);
      setTenants(response.tenants);
      setSuccess(`Found ${response.tenants.length} tenant(s)`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get tenants');
    } finally {
      setIsLoading(false);
    }
  };

  const handleGetOrganizations = async () => {
    if (!accessToken || !selectedTenantId) {
      setError('Access token and tenant ID required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await xeroService.getOrganizations(accessToken, selectedTenantId);
      setOrganizations(response.organizations);
      setSuccess(`Found ${response.organizations.length} organization(s)`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get organizations');
    } finally {
      setIsLoading(false);
    }
  };

  const handleGetPaymentFailures = async () => {
    if (!accessToken || !selectedTenantId) {
      setError('Access token and tenant ID required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await xeroService.getPaymentFailures(accessToken, selectedTenantId);
      setPaymentFailures(response.payment_failures);
      setSuccess(`Found ${response.payment_failures.length} payment failure(s)`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get payment failures');
    } finally {
      setIsLoading(false);
    }
  };

  const handleTestConnection = async () => {
    if (!accessToken || !selectedTenantId) {
      setError('Access token and tenant ID required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await xeroService.testConnection(accessToken, selectedTenantId);
      setSuccess('Connection test successful!');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection test failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Xero Integration (Mediator Pattern)</h2>
      
      {/* OAuth Section */}
      <div className="mb-8 p-4 border rounded-lg">
        <h3 className="text-lg font-semibold mb-4">OAuth Authentication</h3>
        <button
          onClick={handleStartOAuth}
          disabled={isLoading}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {isLoading ? 'Starting...' : 'Connect to Xero'}
        </button>
      </div>

      {/* Access Token Input */}
      <div className="mb-8 p-4 border rounded-lg">
        <h3 className="text-lg font-semibold mb-4">Access Token</h3>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder="Enter access token from OAuth callback"
            value={accessToken || ''}
            onChange={(e) => setAccessToken(e.target.value)}
            className="flex-1 px-3 py-2 border rounded"
          />
          <button
            onClick={handleGetTenants}
            disabled={isLoading || !accessToken}
            className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 disabled:opacity-50"
          >
            Get Tenants
          </button>
        </div>
      </div>

      {/* Tenants */}
      {tenants.length > 0 && (
        <div className="mb-8 p-4 border rounded-lg">
          <h3 className="text-lg font-semibold mb-4">Xero Tenants</h3>
          <div className="space-y-2">
            {tenants.map((tenant) => (
              <div key={tenant.id} className="flex items-center gap-4 p-2 bg-gray-50 rounded">
                <input
                  type="radio"
                  id={tenant.id}
                  name="tenant"
                  value={tenant.id}
                  checked={selectedTenantId === tenant.id}
                  onChange={(e) => setSelectedTenantId(e.target.value)}
                />
                <label htmlFor={tenant.id} className="flex-1">
                  <strong>{tenant.name}</strong> ({tenant.short_code})
                  {tenant.is_active ? ' - Active' : ' - Inactive'}
                </label>
              </div>
            ))}
          </div>
          <div className="mt-4 flex gap-2">
            <button
              onClick={handleGetOrganizations}
              disabled={isLoading || !selectedTenantId}
              className="bg-purple-600 text-white px-4 py-2 rounded hover:bg-purple-700 disabled:opacity-50"
            >
              Get Organizations
            </button>
            <button
              onClick={handleGetPaymentFailures}
              disabled={isLoading || !selectedTenantId}
              className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 disabled:opacity-50"
            >
              Get Payment Failures
            </button>
            <button
              onClick={handleTestConnection}
              disabled={isLoading || !selectedTenantId}
              className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 disabled:opacity-50"
            >
              Test Connection
            </button>
          </div>
        </div>
      )}

      {/* Organizations */}
      {organizations.length > 0 && (
        <div className="mb-8 p-4 border rounded-lg">
          <h3 className="text-lg font-semibold mb-4">Organizations</h3>
          <div className="space-y-2">
            {organizations.map((org) => (
              <div key={org.id} className="p-3 bg-gray-50 rounded">
                <div className="font-semibold">{org.name}</div>
                <div className="text-sm text-gray-600">
                  Legal Name: {org.legal_name} | Currency: {org.base_currency} | Country: {org.country_code}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Payment Failures */}
      {paymentFailures.length > 0 && (
        <div className="mb-8 p-4 border rounded-lg">
          <h3 className="text-lg font-semibold mb-4">Payment Failures</h3>
          <div className="space-y-2">
            {paymentFailures.map((failure) => (
              <div key={failure.id} className="p-3 bg-red-50 border border-red-200 rounded">
                <div className="font-semibold text-red-800">
                  Invoice #{failure.invoice_number} - {failure.customer_name}
                </div>
                <div className="text-sm text-red-600">
                  Amount: {failure.currency} {failure.amount.toFixed(2)} | 
                  Due: {failure.due_date} | 
                  Overdue: {failure.days_overdue} days | 
                  Status: {failure.status}
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  Reason: {failure.failure_reason}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Status Messages */}
      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded text-red-800">
          <strong>Error:</strong> {error}
        </div>
      )}

      {success && (
        <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded text-green-800">
          <strong>Success:</strong> {success}
        </div>
      )}

      {/* Architecture Note */}
      <div className="mt-8 p-4 bg-blue-50 border border-blue-200 rounded">
        <h4 className="font-semibold text-blue-800 mb-2">Architecture Note</h4>
        <p className="text-blue-700 text-sm">
          This integration now uses the <strong>mediator pattern</strong> as designed in the Lexure Intelligence architecture. 
          The UI calls backend API endpoints that use the XeroMediator, which provides:
        </p>
        <ul className="text-blue-700 text-sm mt-2 ml-4 list-disc">
          <li>Unified OAuth handling</li>
          <li>Event-driven architecture</li>
          <li>Consistent error handling</li>
          <li>Rate limiting and retry logic</li>
          <li>Centralized configuration</li>
        </ul>
      </div>
    </div>
  );
}

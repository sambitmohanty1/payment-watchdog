import React, { useState } from 'react';
import { Card, CardHeader, CardBody } from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { PaymentFailureEvent, RetryAction } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { RefreshCw, Download, Filter, Eye, RotateCcw } from 'lucide-react';

interface PaymentFailuresListProps {
  failures: PaymentFailureEvent[];
  onRetry: (id: string, retryData: RetryAction) => void;
  onView: (id: string) => void;
  onExport: () => void;
  loading?: boolean;
}

const PaymentFailuresList: React.FC<PaymentFailuresListProps> = ({
  failures,
  onRetry,
  onView,
  onExport,
  loading = false,
}) => {
  const [selectedFailure, setSelectedFailure] = useState<string | null>(null);

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-AU', {
      style: 'currency',
      currency: currency || 'AUD',
      minimumFractionDigits: 2,
    }).format(amount);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'received':
        return <Badge variant="info">New</Badge>;
      case 'processing':
        return <Badge variant="warning">Processing</Badge>;
      case 'resolved':
        return <Badge variant="success">Resolved</Badge>;
      case 'escalated':
        return <Badge variant="error">Escalated</Badge>;
      default:
        return <Badge variant="default">{status}</Badge>;
    }
  };

  const getFailureReasonColor = (reason: string) => {
    if (reason.toLowerCase().includes('insufficient')) return 'text-error-600';
    if (reason.toLowerCase().includes('expired')) return 'text-warning-600';
    if (reason.toLowerCase().includes('fraud')) return 'text-error-700';
    return 'text-gray-600';
  };

  const handleRetry = (failure: PaymentFailureEvent) => {
    const retryData: RetryAction = {
      payment_failure_id: failure.id,
      retry_amount: failure.amount,
      retry_method: 'stripe',
      customer_notification: true,
      notification_template: 'default_retry',
    };
    
    onRetry(failure.id, retryData);
    setSelectedFailure(null);
  };

  if (loading) {
    return (
      <Card>
        <CardBody>
          <div className="flex items-center justify-center py-8">
            <div className="loading-spinner w-8 h-8"></div>
            <span className="ml-2 text-gray-600">Loading payment failures...</span>
          </div>
        </CardBody>
      </Card>
    );
  }

  if (failures.length === 0) {
    return (
      <Card>
        <CardBody>
          <div className="text-center py-8">
            <div className="text-gray-400 mb-2">🎉</div>
            <h3 className="text-lg font-medium text-gray-900 mb-1">No Payment Failures</h3>
            <p className="text-gray-600">All payments are processing successfully!</p>
          </div>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Payment Failures</h3>
            <p className="text-sm text-gray-600">
              {failures.length} failure{failures.length !== 1 ? 's' : ''} requiring attention
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="secondary" size="sm" onClick={onExport}>
              <Download className="w-4 h-4 mr-1" />
              Export
            </Button>
            <Button variant="secondary" size="sm">
              <Filter className="w-4 h-4 mr-1" />
              Filter
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardBody>
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Customer</th>
                <th>Amount</th>
                <th>Failure Reason</th>
                <th>Status</th>
                <th>Time</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {failures.map((failure) => (
                <tr key={failure.id} className="hover:bg-gray-50">
                  <td>
                    <div>
                      <div className="font-medium text-gray-900">
                        {failure.customer_name || 'Unknown Customer'}
                      </div>
                      {failure.customer_email && (
                        <div className="text-sm text-gray-500">{failure.customer_email}</div>
                      )}
                    </div>
                  </td>
                  <td>
                    <div className="font-medium text-gray-900">
                      {formatCurrency(failure.amount, failure.currency)}
                    </div>
                  </td>
                  <td>
                    <div className={`text-sm ${getFailureReasonColor(failure.failure_reason)}`}>
                      {failure.failure_reason}
                    </div>
                    {failure.failure_code && (
                      <div className="text-xs text-gray-500">Code: {failure.failure_code}</div>
                    )}
                  </td>
                  <td>{getStatusBadge(failure.status)}</td>
                  <td>
                    <div className="text-sm text-gray-600">
                      {formatDistanceToNow(new Date(failure.created_at), { addSuffix: true })}
                    </div>
                  </td>
                  <td>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onView(failure.id)}
                      >
                        <Eye className="w-4 h-4" />
                      </Button>
                      {failure.status === 'received' && (
                        <Button
                          variant="success"
                          size="sm"
                          onClick={() => handleRetry(failure)}
                        >
                          <RotateCcw className="w-4 h-4 mr-1" />
                          Retry
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardBody>
    </Card>
  );
};

export default PaymentFailuresList;

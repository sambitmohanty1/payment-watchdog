import React from 'react';
import { Card, CardHeader, CardBody } from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { PaymentFailureEvent, Alert } from '@/types';
import { AlertTriangle, DollarSign, Clock, User } from 'lucide-react';

interface UrgentActionsCardProps {
  criticalFailures: PaymentFailureEvent[];
  urgentAlerts: Alert[];
  onRetry: (id: string, retryData?: any) => void;
  onViewAlert: (id: string) => void;
}

const UrgentActionsCard: React.FC<UrgentActionsCardProps> = ({
  criticalFailures,
  urgentAlerts,
  onRetry,
  onViewAlert,
}) => {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-AU', {
      style: 'currency',
      currency: 'AUD',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const getPriorityIcon = (priority: 'critical' | 'high' | 'medium' | 'low') => {
    switch (priority) {
      case 'critical':
        return <AlertTriangle className="w-5 h-5 text-error-600" />;
      case 'high':
        return <DollarSign className="w-5 h-5 text-warning-600" />;
      case 'medium':
        return <Clock className="w-5 h-5 text-primary-600" />;
      default:
        return <User className="w-5 h-5 text-gray-600" />;
    }
  };

  const getPriorityColor = (priority: 'critical' | 'high' | 'medium' | 'low') => {
    switch (priority) {
      case 'critical':
        return 'border-error-200 bg-error-50';
      case 'high':
        return 'border-warning-200 bg-warning-50';
      case 'medium':
        return 'border-primary-200 bg-primary-50';
      default:
        return 'border-gray-200 bg-gray-50';
    }
  };

  const getPriorityBadge = (priority: 'critical' | 'high' | 'medium' | 'low') => {
    switch (priority) {
      case 'critical':
        return <Badge variant="error">Critical</Badge>;
      case 'high':
        return <Badge variant="warning">High</Badge>;
      case 'medium':
        return <Badge variant="info">Medium</Badge>;
      default:
        return <Badge variant="default">Low</Badge>;
    }
  };

  const totalUrgentItems = criticalFailures.length + urgentAlerts.length;

  if (totalUrgentItems === 0) {
    return (
      <Card>
        <CardHeader>
          <h3 className="text-lg font-semibold text-gray-900">Urgent Actions</h3>
        </CardHeader>
        <CardBody>
          <div className="text-center py-6">
            <div className="text-success-500 text-4xl mb-2">✅</div>
            <h4 className="text-lg font-medium text-gray-900 mb-1">All Clear!</h4>
            <p className="text-gray-600">No urgent actions required at this time.</p>
          </div>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-5 h-5 text-error-600" />
            <h3 className="text-lg font-semibold text-gray-900">Urgent Actions</h3>
            <Badge variant="error">{totalUrgentItems}</Badge>
          </div>
        </div>
      </CardHeader>
      <CardBody>
        <div className="space-y-4">
          {/* Critical Payment Failures */}
          {criticalFailures.map((failure) => (
            <div
              key={failure.id}
              className={`p-4 rounded-lg border ${getPriorityColor('critical')}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-3">
                  {getPriorityIcon('critical')}
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-1">
                      <h4 className="font-medium text-gray-900">Payment Failure</h4>
                      {getPriorityBadge('critical')}
                    </div>
                    <p className="text-sm text-gray-700 mb-2">
                      {failure.customer_name || 'Unknown Customer'} - {formatCurrency(failure.amount)}
                    </p>
                    <p className="text-sm text-gray-600">
                      {failure.failure_reason}
                    </p>
                  </div>
                </div>
                <Button
                  variant="success"
                  size="sm"
                  onClick={() => onRetry(failure.id)}
                >
                  Retry Payment
                </Button>
              </div>
            </div>
          ))}

          {/* Urgent Alerts */}
          {urgentAlerts.map((alert) => (
            <div
              key={alert.id}
              className={`p-4 rounded-lg border ${getPriorityColor(alert.severity)}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-3">
                  {getPriorityIcon(alert.severity)}
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-1">
                      <h4 className="font-medium text-gray-900">{alert.title}</h4>
                      {getPriorityBadge(alert.severity)}
                    </div>
                    <p className="text-sm text-gray-700 mb-2">{alert.message}</p>
                    {alert.action_required && (
                      <p className="text-xs text-gray-600 font-medium">
                        ⚠️ Action required
                      </p>
                    )}
                  </div>
                </div>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => onViewAlert(alert.id)}
                >
                  View Details
                </Button>
              </div>
            </div>
          ))}
        </div>

        {totalUrgentItems > 0 && (
          <div className="mt-4 p-3 bg-gray-50 rounded-lg">
            <p className="text-sm text-gray-600 text-center">
              {totalUrgentItems} item{totalUrgentItems !== 1 ? 's' : ''} requiring immediate attention
            </p>
          </div>
        )}
      </CardBody>
    </Card>
  );
};

export default UrgentActionsCard;

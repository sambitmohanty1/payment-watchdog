import React from 'react';
import { Card, CardBody } from '@/components/ui/Card';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { DashboardStats } from '@/types';

interface KeyMetricsCardProps {
  stats: DashboardStats;
}

const KeyMetricsCard: React.FC<KeyMetricsCardProps> = ({ stats }) => {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-AU', {
      style: 'currency',
      currency: 'AUD',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${value.toFixed(1)}%`;
  };

  type Trend = 'improving' | 'declining' | 'stable';

  const getTrendIcon = (trend: Trend) => {
    switch (trend) {
      case 'improving':
        return <TrendingUp className="w-4 h-4 text-success-600" />;
      case 'declining':
        return <TrendingDown className="w-4 h-4 text-error-600" />;
      default:
        return <Minus className="w-4 h-4 text-gray-400" />;
    }
  };

  const getTrendColor = (trend: Trend) => {
    switch (trend) {
      case 'improving':
        return 'text-success-600';
      case 'declining':
        return 'text-error-600';
      default:
        return 'text-gray-500';
    }
  };

  const metrics: { label: string; value: string; trend: Trend; description: string }[] = [
    {
      label: 'Cashflow at Risk',
      value: formatCurrency(stats.payment_failures.total_amount),
      trend: 'stable' as const,
      description: 'Total amount at risk from failed payments',
    },
    {
      label: 'Recovery Rate',
      value: formatPercentage(stats.retries.success_rate),
      trend: 'stable' as const,
      description: 'Percentage of failed payments successfully recovered',
    },
    {
      label: 'Failures Today',
      value: stats.payment_failures.total.toString(),
      trend: 'stable' as const,
      description: 'Number of payment failures today',
    },
    {
      label: 'Active Alerts',
      value: stats.alerts.total.toString(),
      trend: 'stable' as const,
      description: 'Number of unread alerts requiring attention',
    },
  ];

  return (
    <Card>
      <CardBody>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Key Metrics</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {metrics.map((metric, index) => (
            <div key={index} className="text-center">
              <div className="flex items-center justify-center mb-2">
                {getTrendIcon(metric.trend)}
                <span className={`ml-1 text-xs font-medium ${getTrendColor(metric.trend)}`}>
                  {metric.trend === 'improving' ? '↗' : metric.trend === 'declining' ? '↘' : '→'}
                </span>
              </div>
              <div className="text-2xl font-bold text-gray-900 mb-1">
                {metric.value}
              </div>
              <div className="text-sm font-medium text-gray-600 mb-1">
                {metric.label}
              </div>
              <div className="text-xs text-gray-500">
                {metric.description}
              </div>
            </div>
          ))}
        </div>
      </CardBody>
    </Card>
  );
};

export default KeyMetricsCard;

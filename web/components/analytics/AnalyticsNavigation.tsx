import React from 'react';
import { Card, CardContent } from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import { BarChart3, TrendingUp, Target, AlertTriangle } from 'lucide-react';

export function AnalyticsNavigation() {
  const navItems = [
    { icon: BarChart3, label: 'Overview', active: true },
    { icon: TrendingUp, label: 'Trends', active: false },
    { icon: Target, label: 'Patterns', active: false },
    { icon: AlertTriangle, label: 'Predictions', active: false },
  ];

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex space-x-2">
          {navItems.map((item) => (
            <Button
              key={item.label}
              variant={item.active ? 'primary' : 'secondary'}
              size="sm"
              className="flex items-center space-x-2"
            >
              <item.icon className="h-4 w-4" />
              <span>{item.label}</span>
            </Button>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

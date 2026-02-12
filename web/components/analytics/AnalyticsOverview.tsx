import React from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';

export function AnalyticsOverview() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Analytics Overview</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500">
          <div className="text-4xl mb-4">📊</div>
          <p>Analytics dashboard coming soon</p>
          <p className="text-sm mt-2">Real-time payment failure intelligence</p>
        </div>
      </CardContent>
    </Card>
  );
}

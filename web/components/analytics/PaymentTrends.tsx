import React from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';

export function PaymentTrends() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Payment Trends</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-gray-500">
          <div className="text-4xl mb-4">📈</div>
          <p>Trend analysis coming soon</p>
          <p className="text-sm mt-2">Historical payment patterns</p>
        </div>
      </CardContent>
    </Card>
  );
}

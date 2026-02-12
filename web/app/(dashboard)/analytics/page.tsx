import { Suspense } from 'react'
import { AnalyticsOverview } from '@/components/analytics/AnalyticsOverview'
import { PaymentTrends } from '@/components/analytics/PaymentTrends'
import { FailurePatterns } from '@/components/analytics/FailurePatterns'
import { PredictiveInsights } from '@/components/analytics/PredictiveInsights'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <Suspense fallback={<LoadingSpinner />}>
        <AnalyticsOverview />
      </Suspense>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<LoadingSpinner />}>
          <PaymentTrends />
        </Suspense>
        
        <Suspense fallback={<LoadingSpinner />}>
          <FailurePatterns />
        </Suspense>
      </div>
      
      <Suspense fallback={<LoadingSpinner />}>
        <PredictiveInsights />
      </Suspense>
    </div>
  )
}

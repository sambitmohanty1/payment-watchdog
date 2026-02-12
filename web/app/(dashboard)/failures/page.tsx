import { Suspense } from 'react'
import { FailuresDashboard } from '@/components/failures/FailuresDashboard'
import { FailuresTable } from '@/components/failures/FailuresTable'
import { FailureFilters } from '@/components/failures/FailureFilters'
import { RiskScoring } from '@/components/failures/RiskScoring'
import { BulkActions } from '@/components/failures/BulkActions'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function FailuresPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Payment Failures Management</h2>
        <BulkActions />
      </div>
      
      <Suspense fallback={<LoadingSpinner />}>
        <FailuresDashboard />
      </Suspense>
      
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3">
          <FailureFilters />
          <Suspense fallback={<LoadingSpinner />}>
            <FailuresTable />
          </Suspense>
        </div>
        
        <div>
          <Suspense fallback={<LoadingSpinner />}>
            <RiskScoring />
          </Suspense>
        </div>
      </div>
    </div>
  )
}

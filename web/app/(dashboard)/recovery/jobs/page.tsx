import { Suspense } from 'react'
import { FailedPaymentsTable, JobFilters, JobActions } from '@/components/recovery'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function RecoveryJobsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Recovery Jobs</h2>
        <JobActions />
      </div>
      
      <JobFilters />
      
      <Suspense fallback={<LoadingSpinner />}>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium">Failed Payments</h3>
            <p className="text-sm text-gray-500">{/* Add count here when API is connected */}</p>
          </div>
          <FailedPaymentsTable />
        </div>
      </Suspense>
    </div>
  )
}

import { Suspense } from 'react'
import { RecoveryDashboard } from '@/components/recovery/RecoveryDashboard'
import { ActiveJobsOverview } from '@/components/recovery/ActiveJobsOverview'
import { RecoveryMetrics } from '@/components/recovery/RecoveryMetrics'
import { WorkflowStatus } from '@/components/recovery/WorkflowStatus'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function RecoveryPage() {
  return (
    <div className="space-y-6">
      <Suspense fallback={<LoadingSpinner />}>
        <RecoveryMetrics />
      </Suspense>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<LoadingSpinner />}>
          <ActiveJobsOverview />
        </Suspense>
        
        <Suspense fallback={<LoadingSpinner />}>
          <WorkflowStatus />
        </Suspense>
      </div>
      
      <Suspense fallback={<LoadingSpinner />}>
        <RecoveryDashboard />
      </Suspense>
    </div>
  )
}

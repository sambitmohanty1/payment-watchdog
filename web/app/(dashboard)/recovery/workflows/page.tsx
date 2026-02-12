import { Suspense } from 'react'
import { WorkflowBuilder } from '@/components/recovery/WorkflowBuilder'
import { WorkflowTemplates } from '@/components/recovery/WorkflowTemplates'
import { WorkflowList } from '@/components/recovery/WorkflowList'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function WorkflowsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Recovery Workflows</h2>
      </div>
      
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <Suspense fallback={<LoadingSpinner />}>
            <WorkflowBuilder />
          </Suspense>
        </div>
        
        <div className="space-y-6">
          <Suspense fallback={<LoadingSpinner />}>
            <WorkflowTemplates />
          </Suspense>
          
          <Suspense fallback={<LoadingSpinner />}>
            <WorkflowList />
          </Suspense>
        </div>
      </div>
    </div>
  )
}

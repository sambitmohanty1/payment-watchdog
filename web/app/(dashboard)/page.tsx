import { Suspense } from 'react'
import { DashboardOverview } from '@/components/dashboard/DashboardOverview'
import { KeyMetricsGrid } from '@/components/dashboard/KeyMetricsGrid'
import { RecentActivity } from '@/components/dashboard/RecentActivity'
import { QuickActions } from '@/components/dashboard/QuickActions'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <QuickActions />
      </div>
      
      <Suspense fallback={<LoadingSpinner />}>
        <KeyMetricsGrid />
      </Suspense>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<LoadingSpinner />}>
          <DashboardOverview />
        </Suspense>
        
        <Suspense fallback={<LoadingSpinner />}>
          <RecentActivity />
        </Suspense>
      </div>
    </div>
  )
}

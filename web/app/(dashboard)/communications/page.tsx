import { Suspense } from 'react'
import { CommunicationsDashboard } from '@/components/communications/CommunicationsDashboard'
import { TemplateManager } from '@/components/communications/TemplateManager'
import { CommunicationHistory } from '@/components/communications/CommunicationHistory'
import { CampaignOverview } from '@/components/communications/CampaignOverview'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

export default function CommunicationsPage() {
  return (
    <div className="space-y-6">
      <Suspense fallback={<LoadingSpinner />}>
        <CommunicationsDashboard />
      </Suspense>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<LoadingSpinner />}>
          <TemplateManager />
        </Suspense>
        
        <Suspense fallback={<LoadingSpinner />}>
          <CampaignOverview />
        </Suspense>
      </div>
      
      <Suspense fallback={<LoadingSpinner />}>
        <CommunicationHistory />
      </Suspense>
    </div>
  )
}

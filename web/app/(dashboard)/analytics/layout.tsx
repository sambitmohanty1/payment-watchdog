import { Metadata } from 'next'
import { AnalyticsNavigation } from '@/components/analytics/AnalyticsNavigation'

export const metadata: Metadata = {
  title: 'Analytics - Lexure Intelligence',
  description: 'Advanced payment failure analytics and insights',
}

export default function AnalyticsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
      </div>
      
      <AnalyticsNavigation />
      
      <div className="mt-6">
        {children}
      </div>
    </div>
  )
}

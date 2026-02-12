import { Metadata } from 'next'
import { CommunicationsNavigation } from '@/components/communications/CommunicationsNavigation'

export const metadata: Metadata = {
  title: 'Communications - Lexure Intelligence',
  description: 'Customer communication management and templates',
}

export default function CommunicationsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Communications</h1>
      </div>
      
      <CommunicationsNavigation />
      
      <div className="mt-6">
        {children}
      </div>
    </div>
  )
}

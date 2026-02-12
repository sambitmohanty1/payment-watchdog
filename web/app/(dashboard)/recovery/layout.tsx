import { Metadata } from 'next'
import { RecoveryNavigation } from '@/components/recovery/RecoveryNavigation'

export const metadata: Metadata = {
  title: 'Recovery Orchestration - Lexure Intelligence',
  description: 'Advanced payment failure recovery workflows and automation',
}

export default function RecoveryLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Recovery Orchestration</h1>
      </div>
      
      <RecoveryNavigation />
      
      <div className="mt-6">
        {children}
      </div>
    </div>
  )
}

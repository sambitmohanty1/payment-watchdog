import { Metadata } from 'next'
import { FailuresNavigation } from '@/components/failures/FailuresNavigation'

export const metadata: Metadata = {
  title: 'Payment Failures - Lexure Intelligence',
  description: 'Intelligent payment failure management with ML insights',
}

export default function FailuresLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Payment Failures</h1>
      </div>
      
      <FailuresNavigation />
      
      <div className="mt-6">
        {children}
      </div>
    </div>
  )
}

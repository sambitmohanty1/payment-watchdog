import { Metadata } from 'next'
import { Sidebar } from '@/components/layout/Sidebar'
import { Header } from '@/components/layout/Header'
import { Providers } from '@/components/providers/Providers'

export const metadata: Metadata = {
  title: 'Lexure Intelligence - Payment Failure Intelligence Platform',
  description: 'Advanced payment failure detection, recovery, and prevention for Australian SMEs',
}

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <Providers>
      <div className="flex h-screen bg-background">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header />
          <main className="flex-1 overflow-auto p-6">
            {children}
          </main>
        </div>
      </div>
    </Providers>
  )
}

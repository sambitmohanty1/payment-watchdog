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
      <div className="flex h-screen bg-transparent">
        <Sidebar className="border-r border-white/5 bg-slate-950/40 backdrop-blur-md" />
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header className="border-b border-white/5 bg-slate-950/20 backdrop-blur-md" />
          <main className="flex-1 overflow-auto">
            {children}
          </main>
        </div>
      </div>
    </Providers>
  )
}

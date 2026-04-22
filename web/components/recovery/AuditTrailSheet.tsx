'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { format } from 'date-fns'
import { MaterialIcon, Button } from '@/components/ui'

interface TimelineEvent {
  id: string
  type: 'system' | 'communication' | 'action' | 'error'
  title: string
  description: string
  timestamp: Date
  metadata?: Record<string, any>
}

interface AuditTrailSheetProps {
  isOpen: boolean
  onClose: () => void
  paymentId: string | null
  customerName: string
}

export function AuditTrailSheet({ isOpen, onClose, paymentId, customerName }: AuditTrailSheetProps) {
  // Mock data for the principal-engineer demonstration
  // In a real scenario, this would be a useQuery hook fetching from /api/payment-failures/:id/history
  const events: TimelineEvent[] = [
    {
      id: '1',
      type: 'system',
      title: 'Webhook Received',
      description: 'Incoming payment failure notification from Stripe (ap_123456789).',
      timestamp: new Date(Date.now() - 3600000),
      metadata: { provider: 'Stripe', event: 'payment_intent.payment_failed' }
    },
    {
      id: '2',
      type: 'system',
      title: 'AI Classification',
      description: 'Rules engine classified failure as "Insufficient Funds" (ISF). Priority: High.',
      timestamp: new Date(Date.now() - 3550000),
    },
    {
      id: '3',
      type: 'communication',
      title: 'Customer Alert Sent',
      description: 'Automated "Payment Failed" email dispatched to customer via SendGrid.',
      timestamp: new Date(Date.now() - 3500000),
      metadata: { template: 'recovery_initial_contact', channel: 'email' }
    },
    {
      id: '4',
      type: 'action',
      title: 'Recovery Strategy Active',
      description: 'Smart Retry Strategy "Australian Business Hours" (ABH) engaged.',
      timestamp: new Date(Date.now() - 3400000),
    },
    {
      id: '5',
      type: 'error',
      title: 'Gateway Timeout',
      description: 'Internal system timeout during NPP availability check in Melbourne region.',
      timestamp: new Date(Date.now() - 1800000),
    }
  ]

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'system': return { name: 'settings', color: 'text-blue-400 bg-blue-400/10' }
      case 'communication': return { name: 'mail', color: 'text-amber-400 bg-amber-400/10' }
      case 'action': return { name: 'play_arrow', color: 'text-emerald-400 bg-emerald-400/10' }
      case 'error': return { name: 'error_outline', color: 'text-rose-400 bg-rose-400/10' }
      default: return { name: 'circle', color: 'text-slate-400 bg-slate-400/10' }
    }
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50"
          />

          {/* Sheet */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 25, stiffness: 200 }}
            className="fixed right-0 top-0 h-full w-full max-w-md bg-slate-900 border-l border-slate-800 shadow-2xl z-50 flex flex-col"
          >
            {/* Header */}
            <div className="p-6 border-b border-slate-800 flex items-center justify-between">
              <div>
                <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2">
                  <MaterialIcon name="history" className="text-blue-400" />
                  Audit Trail
                </h2>
                <div className="text-xs text-slate-500 mt-1 uppercase tracking-widest font-bold">
                  ID: {paymentId} • {customerName}
                </div>
              </div>
              <Button variant="ghost" size="icon" aria-label="Close audit trail" onClick={onClose} className="text-slate-400 hover:text-white">
                <MaterialIcon name="close" />
              </Button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-6 space-y-8">
              <div className="relative">
                {/* Vertical Line */}
                <div className="absolute left-4 top-2 bottom-2 w-px bg-slate-800" />

                <div className="space-y-8">
                  {events.map((event) => {
                    const iconConfig = getEventIcon(event.type)
                    return (
                      <div key={event.id} className="relative pl-12 group">
                        {/* Dot/Icon */}
                        <div className={`absolute left-0 top-0 h-8 w-8 rounded-full border border-slate-700 flex items-center justify-center z-10 transition-transform group-hover:scale-110 ${iconConfig.color}`}>
                          <MaterialIcon name={iconConfig.name} className="text-sm" />
                        </div>

                        <div className="space-y-1">
                          <div className="flex items-center justify-between">
                            <h4 className="text-sm font-bold text-slate-200 uppercase tracking-tight">{event.title}</h4>
                            <span className="text-[10px] font-mono text-slate-500">
                              {format(event.timestamp, 'HH:mm:ss')}
                            </span>
                          </div>
                          <p className="text-xs text-slate-400 leading-relaxed">
                            {event.description}
                          </p>
                          
                          {event.metadata && (
                            <div className="mt-2 p-2 rounded bg-slate-800/50 border border-slate-700/50">
                               <pre className="text-[9px] font-mono text-blue-300 overflow-x-auto">
                                 {JSON.stringify(event.metadata, null, 2)}
                               </pre>
                            </div>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="p-6 border-t border-slate-800 bg-slate-900/80 backdrop-blur-md space-y-3">
              <Button className="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-6">
                 <MaterialIcon name="send" className="mr-2" />
                 Manually Intervene
              </Button>
              <div className="text-[10px] text-center text-slate-600 uppercase tracking-widest">
                Data residency enforced: Australia/Sydney
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

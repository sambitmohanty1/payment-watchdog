'use client'

import { useState, useEffect } from 'react'
import { format } from 'date-fns'
import api from '@/lib/api'
import { 
  Button, 
  Badge, 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow, 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger,
  MaterialIcon 
} from '@/components/ui'

type PaymentStatus = 'pending' | 'succeeded' | 'failed' | 'needs_attention'

interface FailedPayment {
  id: string
  customer: {
    name: string
    email: string
  }
  amount: number
  status: PaymentStatus
  provider: string
  lastAttempt: Date
  nextRetry: Date | null
  retryCount: number
  maxRetries: number
}

const statusConfig = {
  pending: { icon: 'schedule', color: 'bg-amber-500/10 text-amber-500 border-amber-500/20' },
  succeeded: { icon: 'check_circle', color: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' },
  failed: { icon: 'error', color: 'bg-rose-500/10 text-rose-500 border-rose-500/20' },
  needs_attention: { icon: 'warning', color: 'bg-orange-500/10 text-orange-500 border-orange-500/20' },
}

import { useAppStore } from '@/lib/store'
import { AuditTrailSheet } from './AuditTrailSheet'

export function FailedPaymentsTable() {
  const { isPrivacyMode } = useAppStore()
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(true)
  
  const maskPII = (text: string) => {
    if (!isPrivacyMode || !text) return text
    if (text.includes('@')) {
      const [name, domain] = text.split('@')
      return `${name[0]}***@${domain}`
    }
    return `${text[0]}${'*'.repeat(text.length - 1)}`
  }
  
  const [error, setError] = useState<string | null>(null)
  const [failedPayments, setFailedPayments] = useState<FailedPayment[]>([])
  
  // Audit Trail State
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const [focusedPayment, setFocusedPayment] = useState<{id: string, name: string} | null>(null)

  const openAuditTrail = (payment: FailedPayment) => {
    setFocusedPayment({ id: payment.id, name: maskPII(payment.customer.name) })
    setIsSheetOpen(true)
  }
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true)
        const response = await api.getPaymentFailures()
        
        if (response.success && response.data) {
          const mappedData: FailedPayment[] = response.data.data.map((item: any) => ({
            id: item.id,
            customer: { 
              name: item.customer_name || 'Anonymous', 
              email: item.customer_email || 'no-email@provided.com' 
            },
            amount: item.amount,
            status: item.status === 'resolved' ? 'succeeded' : item.status === 'escalated' ? 'failed' : 'pending',
            provider: item.provider_id || 'stripe',
            lastAttempt: new Date(item.created_at),
            nextRetry: null, 
            retryCount: 0,
            maxRetries: 5,
          }))
          setFailedPayments(mappedData)
        }
        setIsLoading(false)
      } catch (err) {
        setError('Failed to load live payment data. Check API connectivity.')
        setIsLoading(false)
      }
    }
    
    fetchData()
  }, [])

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-AU', {
      style: 'currency',
      currency: 'AUD',
    }).format(amount / 100)
  }

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center p-20 space-y-4">
        <MaterialIcon name="sync" className="text-4xl text-blue-500 animate-spin" />
        <span className="text-slate-400 font-medium animate-pulse">Synchronizing with Australian Gateway...</span>
      </div>
    )
  }

  return (
    <>
    <div className="rounded-2xl border border-slate-800 bg-slate-900/40 backdrop-blur-sm overflow-hidden shadow-2xl">
      <Table>
        <TableHeader className="bg-slate-800/50">
          <TableRow className="border-slate-800 hover:bg-transparent">
            <TableHead className="w-12 text-center text-[10px] text-slate-500 font-black uppercase tracking-tighter">
              Audit
            </TableHead>
            <TableHead className="text-slate-300 font-semibold">Customer</TableHead>
            <TableHead className="text-slate-300 font-semibold">Amount</TableHead>
            <TableHead className="text-slate-300 font-semibold">Status</TableHead>
            <TableHead className="text-slate-300 font-semibold">Gateway</TableHead>
            <TableHead className="text-slate-300 font-semibold">Timestamp</TableHead>
            <TableHead className="text-right text-slate-300 font-semibold">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {failedPayments.map((payment) => {
            const config = statusConfig[payment.status] || statusConfig.pending
            const isHighValue = payment.amount >= 100000 // $1,000
            
            return (
              <TableRow 
                key={payment.id} 
                className={cn(
                  "border-slate-800 hover:bg-slate-800/30 transition-colors group cursor-pointer",
                  isHighValue && "bg-blue-500/5"
                )}
                onClick={() => openAuditTrail(payment)}
              >
                <TableCell className="text-center relative">
                  {isHighValue && (
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" />
                  )}
                  <MaterialIcon 
                    name={isHighValue ? "priority_high" : "manage_search"} 
                    className={cn(
                      "transition-colors",
                      isHighValue ? "text-blue-400 font-bold" : "text-slate-600 hover:text-blue-400"
                    )}
                  />
                </TableCell>
                <TableCell>
                  <div className="flex flex-col">
                    <div className="flex items-center gap-2">
                       <span className="font-bold text-slate-100">{maskPII(payment.customer.name)}</span>
                       {isHighValue && (
                         <span className="text-[9px] bg-blue-500/20 text-blue-400 px-1 rounded font-black tracking-tighter border border-blue-500/30">
                           HEAVYWEIGHT
                         </span>
                       )}
                    </div>
                    <span className="text-xs text-slate-500 group-hover:text-slate-400 transition-colors">{maskPII(payment.customer.email)}</span>
                  </div>
                </TableCell>
                <TableCell className="font-mono font-bold text-slate-200">
                  {formatCurrency(payment.amount)}
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className={`${config.color} py-1 px-2 flex items-center gap-1.5 w-fit border`}>
                    <MaterialIcon name={config.icon} className="text-sm" />
                    <span className="capitalize">{payment.status.replace('_', ' ')}</span>
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px] font-black uppercase tracking-tighter border border-slate-700">
                      {payment.provider}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="text-slate-400 text-sm">
                  <div className="flex items-center gap-1.5 font-medium">
                    <MaterialIcon name="event" className="text-sm text-slate-600" />
                    {format(payment.lastAttempt, 'dd MMM, HH:mm')}
                  </div>
                </TableCell>
                <TableCell className="text-right whitespace-nowrap">
                  <div className="flex items-center justify-end gap-2" onClick={(e) => e.stopPropagation()}>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-blue-400">
                      <MaterialIcon name="send" className="text-lg" />
                    </Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400 hover:text-slate-100">
                          <MaterialIcon name="more_vert" className="text-lg" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="bg-slate-900 border-slate-800 text-slate-200">
                        <DropdownMenuItem className="hover:bg-slate-800">
                          <MaterialIcon name="refresh" className="mr-2 text-sm" />
                          Manual Retry
                        </DropdownMenuItem>
                        <DropdownMenuItem className="hover:bg-slate-800">
                          <MaterialIcon name="person" className="mr-2 text-sm" />
                          Contact Client
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-rose-400 hover:bg-rose-400/10">
                          <MaterialIcon name="delete" className="mr-2 text-sm" />
                          Mark for Deletion
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
      
      {failedPayments.length === 0 && (
         <div className="py-20 text-center flex flex-col items-center justify-center space-y-4">
            <MaterialIcon name="verified" className="text-6xl text-slate-800" />
            <div className="text-slate-500 font-medium">No payment failures detected in this cycle.</div>
         </div>
      )}
      
      <div className="bg-slate-800/30 px-6 py-4 flex items-center justify-between">
        <div className="text-xs text-slate-500 uppercase font-black tracking-widest">
            {failedPayments.length} Active Records detected in OCI Melbourne
        </div>
        <div className="flex gap-2">
           <Button variant="outline" size="sm" className="h-8 text-[10px] bg-slate-900 border-slate-700 text-slate-400 hover:text-white uppercase font-bold tracking-tighter">
             <MaterialIcon name="navigate_before" className="mr-1" /> Prev
           </Button>
           <Button variant="outline" size="sm" className="h-8 text-[10px] bg-slate-900 border-slate-700 text-slate-400 hover:text-white uppercase font-bold tracking-tighter">
             Next <MaterialIcon name="navigate_next" className="ml-1" />
           </Button>
        </div>
      </div>
    </div>

    <AuditTrailSheet 
        isOpen={isSheetOpen} 
        onClose={() => setIsSheetOpen(false)} 
        paymentId={focusedPayment?.id || null} 
        customerName={focusedPayment?.name || ''} 
    />
    </>
  )
}


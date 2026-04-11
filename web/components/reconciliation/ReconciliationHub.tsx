'use client'

import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { MaterialIcon } from '@/components/ui/MaterialIcon'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui'

interface Match {
  id: string
  reference: string
  amount: number
  customer: string
  date: string
  status: 'resolved' | 'matched' | 'pending'
}

export function ReconciliationHub() {
  const [isSyncing, setIsSyncing] = useState(false)
  const [matches, setMatches] = useState<Match[]>([
    { id: '1', reference: 'INV-2024-001', amount: 1250.00, customer: 'Acme Corp', date: '2024-04-10', status: 'resolved' },
    { id: '2', reference: 'INV-2024-005', amount: 450.50, customer: 'Global Tech', date: '2024-04-09', status: 'resolved' },
  ])

  const triggerSync = async () => {
    setIsSyncing(true)
    // Simulate API call to /api/xero/reconcile
    try {
      await new Promise(resolve => setTimeout(resolve, 2000))
      // In a real app, we would fetch fresh matches here
    } finally {
      setIsSyncing(false)
    }
  }

  return (
    <section className="mt-12">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-white flex items-center space-x-2">
            <MaterialIcon name="verified_user" className="w-6 h-6 text-indigo-400" />
            <span>Cross-Method <span className="text-indigo-400">Reconciliation</span></span>
          </h2>
          <p className="text-sm text-slate-400 mt-1">Automatic matching of bank transfers against credit card failures.</p>
        </div>
        
        <Button 
          onClick={triggerSync}
          disabled={isSyncing}
          className="premium-glass bg-indigo-600/20 hover:bg-indigo-600/40 text-indigo-400 border-indigo-500/30 px-6 py-2 rounded-full flex items-center space-x-2 border"
        >
          <MaterialIcon name="sync" className={cn("w-4 h-4", isSyncing && "animate-spin")} />
          <span className="text-xs font-bold uppercase tracking-wider">Trigger Xero Sync</span>
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {[
          { label: 'Auto-Matched', value: '85%', sub: 'Last 30 days', color: 'text-emerald-400', icon: 'bolt' },
          { label: 'Pending Sync', value: '12', sub: 'Detected in Xero', color: 'text-indigo-400', icon: 'search' },
          { label: 'Manual Action', value: '2', sub: 'Requires Review', color: 'text-rose-400', icon: 'warning' },
        ].map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="premium-glass rounded-2xl p-6 relative overflow-hidden group"
          >
            <div className="absolute top-0 right-0 w-24 h-24 bg-indigo-500/5 blur-3xl rounded-full group-hover:bg-indigo-500/10 transition-all"></div>
            <MaterialIcon name={stat.icon} className={cn("w-5 h-5 mb-4", stat.color)} />
            <h3 className="text-sm font-bold text-slate-500 uppercase tracking-widest leading-none">{stat.label}</h3>
            <div className={cn("text-3xl font-bold mt-2", stat.color)}>{stat.value}</div>
            <p className="text-[10px] text-slate-500 font-bold uppercase mt-1 tracking-tighter">{stat.sub}</p>
          </motion.div>
        ))}
      </div>

      <div className="premium-glass rounded-3xl overflow-hidden border border-white/5">
        <div className="p-6 border-b border-white/5 bg-white/5 flex items-center justify-between">
          <h3 className="text-sm font-bold text-white uppercase tracking-widest">Recent Auto-Resolutions</h3>
          <span className="text-[10px] font-bold text-indigo-400 bg-indigo-400/10 px-2 py-1 rounded-md uppercase">Live Feed</span>
        </div>
        
        <div className="divide-y divide-white/5">
          {matches.map((match) => (
            <div key={match.id} className="p-4 hover:bg-white/5 transition-colors group flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
                  <MaterialIcon name="check_circle" className="w-5 h-5 text-emerald-400" />
                </div>
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-bold text-white tracking-tight">{match.reference}</span>
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 font-bold uppercase">{match.customer}</span>
                  </div>
                  <div className="text-[10px] text-slate-500 font-bold uppercase tracking-wider mt-0.5">
                    Match Found • ${match.amount.toLocaleString()} • {match.date}
                  </div>
                </div>
              </div>
              
              <Button variant="ghost" className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-white">
                <span className="text-[10px] font-bold uppercase tracking-widest mr-2">Audit Trace</span>
                <MaterialIcon name="chevron_right" className="w-4 h-4" />
              </Button>
            </div>
          ))}
        </div>
        
        <div className="p-4 bg-slate-900/40 text-center">
          <button className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] hover:text-indigo-400 transition-colors">
            View All Reconciliation History
          </button>
        </div>
      </div>
    </section>
  )
}

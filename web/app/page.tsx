'use client'

import { useState } from 'react'
import { useSystemStatus } from '../hooks/useSystemStatus'
import { StatusCard, StatusDetail } from '../components/ui/StatusCard'
import { Button } from '../components/ui'
import { ReconciliationHub } from '../components/reconciliation/ReconciliationHub'
import { motion, AnimatePresence } from 'framer-motion'
import { MaterialIcon } from '@/components/ui'
import { cn } from '@/lib/utils'

interface DetailModal {
  title: string
  status: any
}

export default function HomePage() {
  const [detailModal, setDetailModal] = useState<DetailModal | null>(null)
  
  const { 
    status, 
    loading, 
    error, 
    lastUpdate, 
    isRefreshing, 
    refresh, 
    toggleAutoRefresh, 
    autoRefreshEnabled 
  } = useSystemStatus()

  const handleStatusClick = (title: string, statusData: any) => {
    setDetailModal({ title, status: statusData })
  }

  const closeDetail = () => {
    setDetailModal(null)
  }

  return (
    <div className="min-h-screen mesh-gradient pb-20 selection:bg-indigo-500/30">
      {/* Premium Header */}
      <header className="border-b border-white/5 bg-slate-950/40 backdrop-blur-xl sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-20 items-center">
            <div className="flex items-center space-x-4">
              <div className="relative group">
                <div className="absolute -inset-1 bg-gradient-to-r from-indigo-500 to-purple-600 rounded-xl blur opacity-40 group-hover:opacity-100 transition duration-1000 group-hover:duration-200"></div>
                <div className="relative w-11 h-11 bg-slate-950 rounded-xl flex items-center justify-center border border-white/10 group-hover:border-white/20 transition-colors">
                  <MaterialIcon name="shield" className="text-2xl text-indigo-400 font-bold" />
                </div>
              </div>
              <div>
                <h1 className="text-xl font-bold text-white tracking-tight flex items-center">
                  <span>Payment</span>
                  <span className="text-indigo-400 ml-1.5 font-black text-glow">Watchdog</span>
                </h1>
                <div className="flex items-center space-x-2 mt-0.5">
                  <div className="status-pulse-emerald">
                    <span></span>
                    <span></span>
                  </div>
                  <span className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em]">
                    Sovereign AU Platform
                  </span>
                </div>
              </div>
            </div>
            
            <div className="flex items-center space-x-6">
              <div className="hidden lg:flex items-center space-x-6 mr-6">
                <div className="flex flex-col items-end">
                  <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Region</span>
                  <div className="flex items-center space-x-1">
                    <MaterialIcon name="public" className="text-sm text-slate-400" />
                    <span className="text-xs font-semibold text-slate-300">Australia East</span>
                  </div>
                </div>
                <div className="h-4 w-px bg-white/10"></div>
                <div className="flex flex-col items-end">
                  <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Environment</span>
                  <span className="text-xs font-semibold text-indigo-400 capitalize">{status?.environment || 'Production'}</span>
                </div>
              </div>

              <div className="flex items-center space-x-3">
                <button
                  onClick={toggleAutoRefresh}
                  className={cn(
                    "flex items-center space-x-2 px-4 py-2 rounded-full border transition-all duration-500",
                    autoRefreshEnabled 
                      ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400 shadow-[0_0_15px_rgba(52,211,153,0.1)]' 
                      : 'bg-slate-900 border-white/10 text-slate-400 hover:border-white/20'
                  )}
                >
                  <div className={cn("h-1.5 w-1.5 rounded-full shadow-sm", autoRefreshEnabled ? 'bg-emerald-400 animate-pulse' : 'bg-slate-600')}></div>
                  <span className="text-[10px] font-black uppercase tracking-widest">
                    {autoRefreshEnabled ? 'Live Pulse' : 'On Demand'}
                  </span>
                </button>
                
                <button
                  onClick={refresh}
                  disabled={isRefreshing}
                  className="w-10 h-10 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl shadow-lg shadow-indigo-600/20 disabled:opacity-50 transition-all active:scale-95 flex items-center justify-center border border-indigo-400/30"
                >
                  <MaterialIcon name="sync" className={cn("text-xl", isRefreshing && "animate-spin")} />
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>
      
      <main className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        {/* Welcome Section */}
        <section className="mb-12 relative flex items-center justify-between">
          <div className="relative">
            <div className="absolute -left-12 -top-12 w-32 h-32 bg-indigo-600/10 blur-[80px] rounded-full"></div>
            <div className="flex items-center space-x-2 text-indigo-400 mb-3 text-[10px] font-black uppercase tracking-[0.3em]">
              <MaterialIcon name="bolt" className="text-sm" />
              <span>Operational Intelligence</span>
            </div>
            <h2 className="text-4xl font-extrabold text-white mb-2 leading-tight">
              Recovery <span className="text-indigo-400">Orchestrator</span>
            </h2>
            <p className="text-slate-400 max-w-xl font-medium text-sm leading-relaxed">
              Monitoring AI-driven recovery across Australian payment rails. Automated cross-method reconciliation is currently active.
            </p>
          </div>

          <div className="hidden md:flex space-x-4">
             <div className="premium-glass p-1 rounded-2xl flex">
               <div className="px-4 py-2 bg-indigo-600/10 text-indigo-400 rounded-xl border border-indigo-500/20">
                 <span className="text-xs font-bold uppercase tracking-widest">Overview</span>
               </div>
               <div className="px-4 py-2 text-slate-500 hover:text-slate-300 transition-colors cursor-pointer">
                 <span className="text-xs font-bold uppercase tracking-widest">History</span>
               </div>
             </div>
          </div>
        </section>

        {error && (
          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="premium-glass border-rose-500/30 bg-rose-500/5 rounded-3xl p-6 mb-12 flex items-center space-x-6"
          >
            <div className="w-14 h-14 bg-rose-500/10 rounded-2xl flex items-center justify-center border border-rose-500/20 text-rose-500">
              <MaterialIcon name="electric_bolt" className="text-3xl" />
            </div>
            <div className="flex-1">
              <h3 className="text-rose-400 font-bold uppercase tracking-widest text-xs">Connectivity Incident</h3>
              <p className="text-slate-400 text-sm mt-1">{error}</p>
            </div>
            <button
               onClick={refresh}
               className="px-6 py-2.5 bg-rose-500 hover:bg-rose-600 text-white text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-lg shadow-rose-500/20 active:scale-95"
            >
              Restore Sync
            </button>
          </motion.div>
        )}

        {/* Core Infrastructure Grid */}
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-16">
          <StatusCard
            title="API Cluster"
            status={status?.api || { status: 'unknown', lastCheck: 'Never' }}
            icon="API"
            onClick={() => handleStatusClick('API Core Cluster', status?.api)}
            loading={loading}
          />
          
          <StatusCard
            title="Database"
            status={status?.database || { status: 'unknown', lastCheck: 'Never' }}
            icon="SQL"
            onClick={() => handleStatusClick('PostgreSQL Fabric', status?.database)}
            loading={loading}
          />
          
          <StatusCard
            title="Workers"
            status={status?.workers || { status: 'unknown', lastCheck: 'Never' }}
            icon="EVT"
            onClick={() => handleStatusClick('Recovery Engine', status?.workers)}
            loading={loading}
          />
          
          <StatusCard
            title="Edge Cache"
            status={status?.redis || { status: 'unknown', lastCheck: 'Never' }}
            icon="MEM"
            onClick={() => handleStatusClick('Global Cache Layer', status?.redis)}
            loading={loading}
          />
        </div>

        {/* NEW: Reconciliation Hub Section */}
        <div className="relative">
          <div className="absolute -right-24 -bottom-24 w-64 h-64 bg-emerald-500/5 blur-[100px] rounded-full pointer-events-none"></div>
          <ReconciliationHub />
        </div>

        {/* Modal handling */}
        <AnimatePresence>
          {detailModal && (
            <div className="fixed inset-0 z-[100] flex items-center justify-center px-4">
              <motion.div 
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={closeDetail}
                className="absolute inset-0 bg-slate-950/60 backdrop-blur-sm"
              />
              <motion.div 
                initial={{ opacity: 0, scale: 0.9, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.9, y: 20 }}
                className="relative w-full max-w-lg premium-glass rounded-[40px] p-8 overflow-hidden"
              >
                <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/10 blur-3xl rounded-full"></div>
                <div className="flex items-center justify-between mb-8">
                  <h3 className="text-2xl font-bold text-white tracking-tight">
                    {detailModal.title}
                  </h3>
                  <button 
                    onClick={closeDetail} 
                    className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-white/5 text-slate-400 transition-colors"
                  >
                    ×
                  </button>
                </div>

                <div className="space-y-6">
                  <StatusDetail 
                    status={detailModal.status} 
                    title={detailModal.title}
                    onClose={closeDetail}
                  />
                </div>

                <Button
                  onClick={closeDetail}
                  className="w-full mt-8 bg-indigo-600 hover:bg-indigo-500 text-white font-bold py-4 rounded-2xl shadow-xl shadow-indigo-600/20"
                >
                  Dismiss Overlay
                </Button>
              </motion.div>
            </div>
          )}
        </AnimatePresence>
      </main>
    </div>
  )
}

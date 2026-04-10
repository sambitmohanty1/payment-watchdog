'use client'

import { useState } from 'react'
import { useSystemStatus } from '../hooks/useSystemStatus'
import { StatusCard, StatusDetail } from '../components/ui/StatusCard'

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
    <div className="min-h-screen mesh-gradient pb-20">
      <header className="border-b border-white/5 bg-slate-950/20 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-20 items-center">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
                <span className="text-white font-bold text-xl leading-none">P</span>
              </div>
              <div>
                <h1 className="text-xl font-bold text-white tracking-tight">
                  Payment Watchdog
                </h1>
                <p className="text-[10px] font-bold text-indigo-400 uppercase tracking-[0.2em] leading-none mt-1">
                  Sovereign AU Instance
                </p>
              </div>
            </div>
            
            <div className="flex items-center space-x-6">
              <div className="hidden md:flex flex-col items-end">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Environment</span>
                <span className="text-sm font-medium text-slate-300">{status?.environment || 'Detecting...'}</span>
              </div>
              
              <div className="h-8 w-px bg-white/10 hidden md:block"></div>

              <div className="flex items-center space-x-3">
                <button
                  onClick={toggleAutoRefresh}
                  className={`
                    flex items-center space-x-2 px-4 py-2 rounded-full border transition-all duration-300
                    ${autoRefreshEnabled 
                      ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20' 
                      : 'bg-slate-800 border-slate-700 text-slate-400 hover:bg-slate-700'}
                  `}
                >
                  <span className={`h-2 w-2 rounded-full ${autoRefreshEnabled ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></span>
                  <span className="text-xs font-bold uppercase tracking-wider">
                    Auto-refresh: {autoRefreshEnabled ? 'LIVE' : 'OFF'}
                  </span>
                </button>
                
                <button
                  onClick={refresh}
                  disabled={isRefreshing}
                  className="p-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-full shadow-lg shadow-indigo-500/20 disabled:opacity-50 transition-all"
                >
                  <svg className={`w-5 h-5 ${isRefreshing ? 'animate-spin' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>
      
      <main className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        {/* Hero Section */}
        <section className="mb-12 relative">
          <div className="absolute -left-4 -top-4 w-24 h-24 bg-indigo-500/10 blur-3xl rounded-full"></div>
          <h2 className="text-3xl font-bold text-white mb-2 leading-tight">
            System <span className="text-indigo-400">Pulse</span> Overview
          </h2>
          <p className="text-slate-400 max-w-2xl font-medium">
            Real-time monitoring of recovery orchestration services, database health, and payment rail connectivity across the Sovereign Australian region.
          </p>
        </section>

        {error && (
          <div className="glass-card border-rose-500/50 bg-rose-500/5 rounded-2xl p-6 mb-12 flex items-center space-x-4">
            <div className="p-3 bg-rose-500/20 rounded-xl">
              <span className="text-2xl">⚠️</span>
            </div>
            <div className="flex-1">
              <h3 className="text-rose-400 font-bold uppercase tracking-wider text-xs">Connectivity Error</h3>
              <p className="text-slate-300 text-sm mt-1">{error}</p>
            </div>
            <button
               onClick={refresh}
               className="px-4 py-2 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold rounded-lg transition-colors"
            >
              Retry Connection
            </button>
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <StatusCard
            title="API Core"
            status={status?.api || { status: 'unknown', lastCheck: 'Never' }}
            icon="API"
            onClick={() => handleStatusClick('API Core Service', status?.api)}
            loading={loading}
          />
          
          <StatusCard
            title="Persistence"
            status={status?.database || { status: 'unknown', lastCheck: 'Never' }}
            icon="SQL"
            onClick={() => handleStatusClick('PostgreSQL Cluster', status?.database)}
            loading={loading}
          />
          
          <StatusCard
            title="Orchestrator"
            status={status?.workers || { status: 'unknown', lastCheck: 'Never' }}
            icon="EVT"
            onClick={() => handleStatusClick('Recovery Worker Pool', status?.workers)}
            loading={loading}
          />
          
          <StatusCard
            title="Caching"
            status={status?.redis || { status: 'unknown', lastCheck: 'Never' }}
            icon="MEM"
            onClick={() => handleStatusClick('Redis Cache Layer', status?.redis)}
            loading={loading}
          />
        </div>

        {detailModal && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
              <div className="mt-3">
                <h3 className="text-lg font-medium text-gray-900">
                  {detailModal.title} Details
                </h3>
                <div className="mt-2 px-7 py-3">
                  <StatusDetail 
                    status={detailModal.status} 
                    title={detailModal.title}
                    onClose={closeDetail}
                  />
                </div>
                <div className="items-center px-4 py-3">
                  <button
                    onClick={closeDetail}
                    className="px-4 py-2 bg-blue-500 text-white text-base font-medium rounded-md w-full shadow-sm hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-300"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

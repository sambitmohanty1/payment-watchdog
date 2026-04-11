import React from 'react'
import { ServiceStatus } from '../../hooks/useSystemStatus'
import { cn } from '@/lib/utils'
import { MaterialIcon } from '@/components/ui/MaterialIcon'

interface StatusCardProps {
  title: string
  status: ServiceStatus
  icon: string
  onClick?: () => void
  className?: string
  loading?: boolean
}

export const StatusCard: React.FC<StatusCardProps> = ({
  title,
  status,
  icon,
  onClick,
  className = '',
  loading = false,
}) => {
  const isDown = status.status === 'down';
  const isHealthy = status.status === 'healthy';
  const isDegraded = status.status === 'degraded';

  return (
    <div 
      className={cn(
        "premium-glass rounded-3xl p-6 relative group cursor-pointer overflow-hidden",
        className
      )}
      onClick={onClick}
      role="button"
      tabIndex={0}
    >
      <div className="absolute top-0 right-0 w-24 h-24 bg-white/5 blur-3xl rounded-full -mr-8 -mt-8 group-hover:bg-white/10 transition-all duration-700"></div>
      
      <div className="flex flex-col h-full justify-between relative z-10">
        <div className="flex items-start justify-between">
          <div className="relative">
            <div className="absolute -inset-2 bg-indigo-500/10 rounded-full blur-lg opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div className="relative p-3 rounded-2xl bg-white/5 border border-white/10 group-hover:border-white/20 transition-all duration-300">
              {loading ? (
                <div className="animate-spin rounded-full h-6 w-6 border-2 border-indigo-500/20 border-t-indigo-500"></div>
              ) : (
                <span className="text-xl font-black bg-gradient-to-br from-indigo-300 to-indigo-500 bg-clip-text text-transparent tracking-tighter">
                  {icon}
                </span>
              )}
            </div>
          </div>
          <div className="flex items-center space-x-2 bg-white/5 px-2 py-1 rounded-full border border-white/5">
            <span className="status-pulse-emerald">
              <span className={cn(isHealthy ? 'bg-emerald-400' : isDegraded ? 'bg-amber-400' : isDown ? 'bg-rose-500 ' : 'bg-slate-500')}></span>
              <span className={cn("relative inline-flex h-2 w-2 rounded-full", isHealthy ? 'bg-emerald-500' : isDegraded ? 'bg-amber-500' : isDown ? 'bg-rose-500' : 'bg-slate-500')}></span>
            </span>
            <span className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-400 pr-1">Active</span>
          </div>
        </div>

        <div className="mt-8">
          <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em]">{title}</h3>
          <div className="flex items-baseline mt-2">
            <span className={cn(
              "text-3xl font-black tracking-tight transition-colors duration-500",
              isHealthy ? 'text-white text-glow' : isDegraded ? 'text-amber-400' : isDown ? 'text-rose-400' : 'text-slate-500'
            )}>
              {loading ? '...' : (isHealthy ? 'Optimal' : status.status)}
            </span>
          </div>
        </div>

        <div className="mt-6 pt-6 border-t border-white/5 space-y-3">
          {status.error ? (
            <div className="flex items-center space-x-2 text-rose-400">
              <MaterialIcon name="warning" className="w-3 h-3" />
              <p className="text-[10px] font-bold uppercase truncate">{status.error}</p>
            </div>
          ) : (
             <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                   <MaterialIcon name="bolt" className="w-3 h-3 text-indigo-400" />
                   <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                     {status.responseTime ? `${status.responseTime}ms` : 'Syncing'}
                   </span>
                </div>
                <div className="h-1 w-12 bg-white/5 rounded-full overflow-hidden">
                   <div className="h-full bg-indigo-500 w-2/3"></div>
                </div>
             </div>
          )}
          <div className="flex justify-between items-center text-[9px] text-slate-500 font-black tracking-widest uppercase">
            <span>Audit Secured</span>
            <span>{new Date(status.lastCheck).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

interface StatusDetailProps {
  title: string
  status: ServiceStatus
  onClose: () => void
}

export const StatusDetail: React.FC<StatusDetailProps> = ({
  title,
  status,
  onClose,
}) => {
  return (
    <div className="space-y-6">
      <div className={cn(
        "p-6 rounded-[32px] border bg-white/5 flex items-center space-x-6",
        status.status === 'healthy' ? 'border-emerald-500/20 shadow-[0_0_20px_rgba(16,185,129,0.05)]' : 'border-rose-500/20'
      )}>
        <div className={cn(
          "w-12 h-12 rounded-2xl flex items-center justify-center border",
          status.status === 'healthy' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : 'bg-rose-500/10 border-rose-500/20 text-rose-400'
        )}>
          {status.status === 'healthy' ? <MaterialIcon name="verified_user" className="w-6 h-6" /> : <MaterialIcon name="warning" className="w-6 h-6" />}
        </div>
        <div>
          <p className="font-black text-slate-500 uppercase tracking-[0.2em] text-[10px]">Service Operational Integrity</p>
          <p className={cn("text-2xl font-black tracking-tight", status.status === 'healthy' ? 'text-emerald-400' : 'text-rose-400')}>
            {status.status.toUpperCase()}
          </p>
        </div>
      </div>
      
      <div className="grid grid-cols-2 gap-4">
        <div className="premium-glass p-5 rounded-3xl border border-white/5">
          <p className="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] mb-2">Network Latency</p>
          <div className="flex items-baseline space-x-1">
            <span className="text-3xl font-black text-indigo-400 tracking-tighter">{status.responseTime || '0'}</span>
            <span className="text-[10px] font-black text-slate-600 uppercase tracking-widest">ms</span>
          </div>
        </div>
        <div className="premium-glass p-5 rounded-3xl border border-white/5">
          <p className="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] mb-2">Cycle Analysis</p>
          <p className="text-sm font-bold text-slate-300">{new Date(status.lastCheck).toLocaleTimeString()}</p>
        </div>
      </div>
      
      {status.error && (
        <div className="bg-rose-500/5 border border-rose-500/20 p-5 rounded-3xl">
          <div className="flex items-center space-x-2 mb-2">
            <MaterialIcon name="bolt" className="w-3 h-3 text-rose-500" />
            <p className="text-[10px] font-black text-rose-500 uppercase tracking-[0.2em]">Fault Log</p>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed font-mono bg-black/20 p-3 rounded-xl border border-white/5">{status.error}</p>
        </div>
      )}
      
      <div className="bg-indigo-500/5 p-5 rounded-3xl border border-indigo-500/10">
        <div className="flex items-center justify-center space-x-2 mb-3">
          <MaterialIcon name="monitoring" className="w-3 h-3 text-indigo-400" />
          <p className="text-[10px] font-black text-indigo-400 uppercase tracking-[0.2em]">Sovereignty Intelligence</p>
        </div>
        <div className="text-xs text-slate-400 leading-relaxed text-center italic px-4">
          {status.status === 'healthy' 
            ? "Environment stability is within nominal Australian region constraints. Audit records secured and hashed." 
            : "Sovereignty threshold violation detected. Verify VPC endpoint connectivity and local persistence layer health."}
        </div>
      </div>
    </div>
  )
}

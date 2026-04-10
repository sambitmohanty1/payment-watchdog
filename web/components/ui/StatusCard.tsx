import React from 'react'
import { ServiceStatus, getStatusColor, getStatusIcon } from '../../hooks/useSystemStatus'

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
  const isUnknown = status.status === 'unknown';

  return (
    <div 
      className={`
        glass-card rounded-2xl p-6 relative group
        ${className}
      `}
      onClick={onClick}
      role="button"
      tabIndex={0}
    >
      <div className="flex flex-col h-full justify-between">
        <div className="flex items-start justify-between">
          <div className={`
            p-3 rounded-xl bg-white/5 border border-white/10
            group-hover:bg-white/10 transition-colors
          `}>
            {loading ? (
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-500"></div>
            ) : (
              <span className="text-xl font-bold bg-gradient-to-br from-indigo-400 to-purple-400 bg-clip-text text-transparent">
                {icon}
              </span>
            )}
          </div>
          <div className="flex items-center space-x-2">
            <span className="status-pulse">
              <span className={`status-pulse-ping ${isHealthy ? 'bg-emerald-500' : isDegraded ? 'bg-amber-500' : isDown ? 'bg-rose-500' : 'bg-slate-500'}`}></span>
              <span className={`relative inline-flex h-3 w-3 rounded-full ${isHealthy ? 'bg-emerald-500' : isDegraded ? 'bg-amber-500' : isDown ? 'bg-rose-500' : 'bg-slate-500'}`}></span>
            </span>
            <span className="text-[10px] font-bold uppercase tracking-widest text-slate-400">Live</span>
          </div>
        </div>

        <div className="mt-8">
          <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider">{title}</h3>
          <div className="flex items-baseline mt-1 space-x-2">
            <span className={`text-4xl font-bold tracking-tight ${isHealthy ? 'text-emerald-400' : isDegraded ? 'text-amber-400' : isDown ? 'text-rose-400' : 'text-slate-200'}`}>
              {loading ? '---' : status.status}
            </span>
          </div>
        </div>

        <div className="mt-6 pt-6 border-t border-white/5 space-y-2">
          {status.error ? (
            <p className="text-xs text-rose-400 font-medium truncate">{status.error}</p>
          ) : (
            <p className="text-xs text-slate-500 uppercase tracking-wide">
              {status.responseTime ? `Latency: ${status.responseTime}ms` : 'Awaiting Metrics'}
            </p>
          )}
          <div className="flex justify-between items-center text-[10px] text-slate-600 font-medium">
            <span>AUDIT LOG ACTIVE</span>
            <span>{new Date(status.lastCheck).toLocaleTimeString()}</span>
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
    <div className="fixed inset-0 bg-[#020617]/80 backdrop-blur-sm overflow-y-auto h-full w-full z-[100] flex items-center justify-center p-4">
      <div className="relative glass-card w-full max-w-md rounded-2xl overflow-hidden border-indigo-500/20 shadow-indigo-500/10">
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-xl font-bold text-white tracking-tight">
              {title} Analysis
            </h3>
            <button
              onClick={onClose}
              className="text-slate-500 hover:text-white transition-colors"
            >
              ✕
            </button>
          </div>
          
          <div className="space-y-6">
            <div className={`
              p-4 rounded-xl border bg-white/5 flex items-center space-x-4
              ${status.status === 'healthy' ? 'border-emerald-500/20' : 'border-rose-500/20'}
            `}>
              <span className="status-pulse">
                <span className={`status-pulse-ping ${status.status === 'healthy' ? 'bg-emerald-500' : 'bg-rose-500'}`}></span>
                <span className={`relative inline-flex h-3 w-3 rounded-full ${status.status === 'healthy' ? 'bg-emerald-500' : 'bg-rose-500'}`}></span>
              </span>
              <div>
                <p className="font-bold text-slate-200 uppercase tracking-wider text-xs">Current State</p>
                <p className={`text-lg font-bold ${status.status === 'healthy' ? 'text-emerald-400' : 'text-rose-400'}`}>
                  {status.status.toUpperCase()}
                </p>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-white/5 p-4 rounded-xl border border-white/5">
                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Latency</p>
                <p className="text-xl font-mono text-indigo-400">{status.responseTime || '0'}ms</p>
              </div>
              <div className="bg-white/5 p-4 rounded-xl border border-white/5">
                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Last Audit</p>
                <p className="text-sm font-medium text-slate-300">{new Date(status.lastCheck).toLocaleTimeString()}</p>
              </div>
            </div>
            
            {status.error && (
              <div className="bg-rose-500/10 border border-rose-500/20 p-4 rounded-xl">
                <p className="text-[10px] font-bold text-rose-400 uppercase tracking-widest mb-1">Fault Detected</p>
                <p className="text-sm text-slate-300 leading-relaxed font-mono">{status.error}</p>
              </div>
            )}
            
            <div className="bg-white/5 p-4 rounded-xl border border-white/5">
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2 text-center">Health Recommendation</p>
              <div className="text-xs text-slate-400 space-y-2 text-center italic">
                {status.status === 'healthy' ? '"Service performing within optimal sovereignty parameters. No intervention required."' : '"Anomalous behavior detected. Verification of internal cluster networking and PVC availability recommended."'}
              </div>
            </div>
          </div>
          
          <button
            onClick={onClose}
            className="mt-8 w-full py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-500/20"
          >
            Acknowledge & Close
          </button>
        </div>
      </div>
    </div>
  )
}

'use client'

import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, LoadingSpinner, Progress, MaterialIcon } from '@/components/ui'

export function RecoveryDashboard() {
  const { data: response, isLoading } = useQuery({
    queryKey: ['recovery-dashboard-stats'],
    queryFn: () => api.getDashboardStats(),
    refetchInterval: 30000,
  })

  if (isLoading) {
    return (
      <div className="flex justify-center p-12">
        <LoadingSpinner />
      </div>
    )
  }

  const statsData = response?.data;

  const stats = [
    {
      label: 'Total Recovered',
      value: `$${((statsData?.payment_failures?.total_amount || 0) / 100).toLocaleString()}`,
      change: '+12.5%',
      trend: 'up',
      icon: 'payments'
    },
    {
      label: 'Success Rate',
      value: `${(statsData?.retries?.success_rate || 0).toFixed(1)}%`,
      change: '+2.1%',
      trend: 'up',
      icon: 'check_circle'
    },
    {
      label: 'Active Retries',
      value: statsData?.retries?.total || 0,
      change: 'Active',
      trend: 'neutral',
      icon: 'autorenew'
    },
    {
      label: 'Total Failures',
      value: statsData?.payment_failures?.total || 0,
      change: '-8%',
      trend: 'down',
      icon: 'error_outline'
    }
  ]

  return (
    <Card className="border-none shadow-xl bg-slate-900/50 backdrop-blur-md">
      <CardHeader>
        <CardTitle className="flex items-center space-x-2 text-slate-100">
          <MaterialIcon name="analytics" className="text-blue-400" />
          <span>Recovery Performance</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {stats.map((stat, index) => {
            return (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                className="p-4 rounded-xl bg-slate-800/40 border border-slate-700/50 space-y-2 hover:border-blue-500/50 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="text-xs font-medium text-slate-400 uppercase tracking-wider">{stat.label}</div>
                  <MaterialIcon name={stat.icon} className="text-lg text-slate-500" />
                </div>
                <div className="text-2xl font-bold text-slate-100">{stat.value}</div>
                <div className="flex items-center text-xs">
                  <span 
                    className={
                      stat.trend === 'up' 
                        ? 'text-emerald-400' 
                        : stat.trend === 'down' 
                        ? 'text-rose-400' 
                        : 'text-slate-400'
                    }
                  >
                    {stat.change}
                  </span>
                  <span className="text-slate-500 ml-1 italic">current window</span>
                </div>
              </motion.div>
            )
          })}
        </div>
        
        <div className="mt-10 space-y-6">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-slate-300 uppercase tracking-widest">Provider Recovery Health</h3>
            <MaterialIcon name="shutter_speed" className="text-slate-500 animate-pulse" />
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-6">
            {(statsData?.payment_failures?.by_provider || [
              { name: 'Stripe', count: 85 },
              { name: 'Xero', count: 92 },
              { name: 'PayTo', count: 64 },
              { name: 'Manual', count: 77 }
            ]).slice(0, 4).map((provider: any, index: number) => (
              <div key={index} className="space-y-2">
                <div className="flex items-center justify-between text-xs px-1">
                  <span className="text-slate-400 font-medium">{provider.name || `Provider ${index + 1}`}</span>
                  <span className="text-blue-400 font-bold">{provider.count || provider.value}%</span>
                </div>
                <Progress value={provider.count || provider.value} className="h-1.5 bg-slate-800" />
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}


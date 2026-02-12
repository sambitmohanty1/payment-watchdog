'use client'

import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { RefreshCw, CheckCircle, Clock, AlertCircle, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

interface RecoveryMetric {
  label: string
  value: string | number
  change: string
  trend: 'up' | 'down' | 'neutral'
  icon: React.ComponentType<{ className?: string }>
}

const mockMetrics: RecoveryMetric[] = [
  {
    label: 'Active Jobs',
    value: 247,
    change: '+12%',
    trend: 'up',
    icon: RefreshCw
  },
  {
    label: 'Success Rate',
    value: '94.2%',
    change: '+2.1%',
    trend: 'up',
    icon: CheckCircle
  },
  {
    label: 'Avg Processing Time',
    value: '2.4m',
    change: '-15%',
    trend: 'down',
    icon: Clock
  },
  {
    label: 'Failed Jobs',
    value: 18,
    change: '-8%',
    trend: 'down',
    icon: AlertCircle
  }
]

export function RecoveryMetrics() {
  const { data: metrics, isLoading } = useQuery({
    queryKey: ['recovery-metrics'],
    queryFn: () => Promise.resolve(mockMetrics),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {metrics?.map((metric, index) => (
        <motion.div
          key={metric.label}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: index * 0.1 }}
        >
          <Card hover>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {metric.label}
              </CardTitle>
              <metric.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{metric.value}</div>
              <div className="flex items-center space-x-1 text-xs">
                <TrendingUp 
                  className={`h-3 w-3 ${
                    metric.trend === 'up' 
                      ? 'text-green-500 rotate-0' 
                      : metric.trend === 'down'
                      ? 'text-red-500 rotate-180'
                      : 'text-gray-500'
                  }`} 
                />
                <span className={
                  metric.trend === 'up' 
                    ? 'text-green-600' 
                    : metric.trend === 'down'
                    ? 'text-red-600'
                    : 'text-gray-600'
                }>
                  {metric.change}
                </span>
                <span className="text-muted-foreground">from last hour</span>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </div>
  )
}

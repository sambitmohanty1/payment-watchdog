'use client'

import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Activity, AlertCircle, Clock, Zap, CheckCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'
import { Progress } from '@/components/ui/progress'

interface RecoveryStat {
  label: string
  value: string | number
  change: string
  trend: 'up' | 'down' | 'neutral'
  icon: React.ComponentType<{ className?: string }>
}

const mockStats: RecoveryStat[] = [
  {
    label: 'Total Recovered',
    value: '$281,450',
    change: '+12.5%',
    trend: 'up',
    icon: Zap
  },
  {
    label: 'Success Rate',
    value: '94.2%',
    change: '+2.1%',
    trend: 'up',
    icon: CheckCircle
  },
  {
    label: 'Avg Recovery Time',
    value: '3.2h',
    change: '-15%',
    trend: 'down',
    icon: Clock
  },
  {
    label: 'Failures',
    value: '18',
    change: '-8%',
    trend: 'down',
    icon: AlertCircle
  }
]

export function RecoveryDashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['recovery-dashboard-stats'],
    queryFn: () => Promise.resolve(mockStats),
    refetchInterval: 30000,
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Activity className="h-5 w-5" />
          <span>Recovery Performance</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {stats?.map((stat, index) => {
            const Icon = stat.icon
            return (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                className="space-y-2"
              >
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">{stat.label}</div>
                  <Icon className="h-4 w-4 text-muted-foreground" />
                </div>
                <div className="text-2xl font-bold">{stat.value}</div>
                <div className="flex items-center text-xs">
                  <span 
                    className={
                      stat.trend === 'up' 
                        ? 'text-green-500' 
                        : stat.trend === 'down' 
                        ? 'text-red-500' 
                        : 'text-gray-500'
                    }
                  >
                    {stat.change}
                  </span>
                  <span className="text-muted-foreground ml-1">vs last period</span>
                </div>
              </motion.div>
            )
          })}
        </div>
        
        <div className="mt-8 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium">Workflow Performance</h3>
            <Button variant="outline" size="sm">View All</Button>
          </div>
          
          <div className="space-y-4">
            {[70, 85, 60, 90].map((value, index) => (
              <div key={index} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span>Workflow {index + 1}</span>
                  <span className="font-medium">{value}%</span>
                </div>
                <Progress value={value} className="h-2" />
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

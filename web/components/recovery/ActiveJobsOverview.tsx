'use client'

import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Play, Pause, CheckCircle, XCircle, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/Badge'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

interface RecoveryJob {
  id: string
  type: 'retry' | 'dunning' | 'communication'
  status: 'running' | 'paused' | 'completed' | 'failed'
  customer: string
  amount: number
  progress: number
  startTime: string
  estimatedCompletion: string
}

const mockJobs: RecoveryJob[] = [
  {
    id: 'job-001',
    type: 'retry',
    status: 'running',
    customer: 'Acme Corp',
    amount: 2450.00,
    progress: 65,
    startTime: '2024-01-15T10:30:00Z',
    estimatedCompletion: '2024-01-15T11:15:00Z'
  },
  {
    id: 'job-002',
    type: 'dunning',
    status: 'paused',
    customer: 'TechStart Ltd',
    amount: 890.50,
    progress: 30,
    startTime: '2024-01-15T09:45:00Z',
    estimatedCompletion: '2024-01-15T12:00:00Z'
  },
  {
    id: 'job-003',
    type: 'communication',
    status: 'completed',
    customer: 'Global Solutions',
    amount: 5200.00,
    progress: 100,
    startTime: '2024-01-15T08:00:00Z',
    estimatedCompletion: '2024-01-15T10:30:00Z'
  }
]

const statusIcons = {
  running: Play,
  paused: Pause,
  completed: CheckCircle,
  failed: XCircle
}

const statusColors = {
  running: 'bg-blue-500',
  paused: 'bg-yellow-500',
  completed: 'bg-green-500',
  failed: 'bg-red-500'
}

export function ActiveJobsOverview() {
  const { data: jobs, isLoading } = useQuery({
    queryKey: ['active-jobs'],
    queryFn: () => Promise.resolve(mockJobs),
    refetchInterval: 5000, // Refresh every 5 seconds for real-time updates
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          Active Recovery Jobs
          <Badge variant="secondary">{jobs?.length || 0} jobs</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {jobs?.map((job, index) => {
            const StatusIcon = statusIcons[job.status]
            
            return (
              <motion.div
                key={job.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-accent/50 transition-colors"
              >
                <div className="flex items-center space-x-4">
                  <div className={`p-2 rounded-full ${statusColors[job.status]}`}>
                    <StatusIcon className="h-4 w-4 text-white" />
                  </div>
                  
                  <div>
                    <div className="font-medium">{job.customer}</div>
                    <div className="text-sm text-muted-foreground">
                      {job.type} • ${job.amount.toLocaleString()}
                    </div>
                  </div>
                </div>
                
                <div className="flex items-center space-x-4">
                  {job.status === 'running' && (
                    <div className="flex items-center space-x-2">
                      <div className="w-24 bg-muted rounded-full h-2">
                        <motion.div
                          className="bg-primary h-2 rounded-full"
                          initial={{ width: 0 }}
                          animate={{ width: `${job.progress}%` }}
                          transition={{ duration: 0.5 }}
                        />
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {job.progress}%
                      </span>
                    </div>
                  )}
                  
                  <div className="flex items-center space-x-2">
                    {job.status === 'running' && (
                      <Button variant="ghost" size="sm">
                        <Pause className="h-4 w-4" />
                      </Button>
                    )}
                    {job.status === 'paused' && (
                      <Button variant="ghost" size="sm">
                        <Play className="h-4 w-4" />
                      </Button>
                    )}
                    <Button variant="ghost" size="sm">
                      <Clock className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </motion.div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}

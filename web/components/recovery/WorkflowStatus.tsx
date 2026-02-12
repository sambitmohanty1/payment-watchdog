'use client'

import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { GitBranch, Zap, Users, DollarSign } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Progress } from '@/components/ui/progress'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

interface WorkflowMetric {
  name: string
  activeJobs: number
  successRate: number
  totalRecovered: number
  avgTime: string
  color: string
}

const mockWorkflows: WorkflowMetric[] = [
  {
    name: 'Smart Retry',
    activeJobs: 45,
    successRate: 87,
    totalRecovered: 125000,
    avgTime: '2.3h',
    color: 'bg-blue-500'
  },
  {
    name: 'Dunning Sequence',
    activeJobs: 23,
    successRate: 72,
    totalRecovered: 89000,
    avgTime: '4.1h',
    color: 'bg-green-500'
  },
  {
    name: 'Customer Outreach',
    activeJobs: 12,
    successRate: 94,
    totalRecovered: 67000,
    avgTime: '1.8h',
    color: 'bg-purple-500'
  }
]

export function WorkflowStatus() {
  const { data: workflows, isLoading } = useQuery({
    queryKey: ['workflow-status'],
    queryFn: () => Promise.resolve(mockWorkflows),
    refetchInterval: 30000,
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <GitBranch className="h-5 w-5" />
          <span>Workflow Performance</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {workflows?.map((workflow, index) => (
            <motion.div
              key={workflow.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="space-y-3"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className={`w-3 h-3 rounded-full ${workflow.color}`} />
                  <span className="font-medium">{workflow.name}</span>
                </div>
                <div className="text-sm text-muted-foreground">
                  {workflow.activeJobs} active
                </div>
              </div>
              
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div className="flex items-center space-x-2">
                  <Zap className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Success:</span>
                  <span className="font-medium">{workflow.successRate}%</span>
                </div>
                
                <div className="flex items-center space-x-2">
                  <DollarSign className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Recovered:</span>
                  <span className="font-medium">${workflow.totalRecovered.toLocaleString()}</span>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Users className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Avg Time:</span>
                  <span className="font-medium">{workflow.avgTime}</span>
                </div>
              </div>
              
              <Progress value={workflow.successRate} className="h-2" />
              
              {index < workflows.length - 1 && (
                <div className="border-b border-border mt-4" />
              )}
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

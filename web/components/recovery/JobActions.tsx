'use client'

import { Button, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui'
import { MaterialIcon } from '@/components/ui/MaterialIcon'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'

export function JobActions({
  onRefresh,
  onExport,
  selectedCount,
}: {
  onRefresh: () => void
  onExport: (selectedOnly: boolean) => void
  selectedCount: number
}) {
  const router = useRouter()
  
  const handleNewJob = () => {
    router.push('/recovery/new')
  }
  
  const handleRetrySelected = () => {
    if (selectedCount === 0) {
      toast.warning('No payments selected')
      return
    }
    toast.info(`Retrying ${selectedCount} selected payments`)
    // TODO: Implement actual retry logic
  }
  
  const handleExport = (selectedOnly = false) => {
    const count = selectedOnly ? selectedCount : 'all'
    toast.info(`Exporting ${count} payments`)
    onExport(selectedOnly)
  }

  return (
    <div className="flex items-center space-x-2">
      <Button 
        variant="outline" 
        size="sm" 
        className="h-8"
        onClick={onRefresh}
      >
        <MaterialIcon name="sync" className="mr-2 h-4 w-4" />
        Refresh
      </Button>
      
      <Button 
        variant="outline" 
        size="sm" 
        className="h-8"
        onClick={() => handleExport(false)}
      >
        <MaterialIcon name="download" className="mr-2 h-4 w-4" />
        Export
      </Button>
      
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" className="h-8">
            <MaterialIcon name="filter_alt" className="mr-2 h-4 w-4" />
            Actions
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleNewJob}>
            <MaterialIcon name="add" className="mr-2 h-4 w-4" />
            New Recovery Job
          </DropdownMenuItem>
          <DropdownMenuItem 
            onClick={handleRetrySelected}
            disabled={selectedCount === 0}
          >
            <MaterialIcon name="sync" className="mr-2 h-4 w-4" />
            Retry Selected
          </DropdownMenuItem>
          <DropdownMenuItem 
            onClick={() => handleExport(true)}
            disabled={selectedCount === 0}
          >
            <MaterialIcon name="download" className="mr-2 h-4 w-4" />
            Export Selected
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

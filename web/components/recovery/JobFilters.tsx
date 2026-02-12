'use client'

import { Search } from 'lucide-react'
import { Input } from '@/components/ui/Input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type FilterValues = {
  search: string
  status: string
  provider: string
  dateRange: string
}

export function JobFilters({
  values,
  onChange,
}: {
  values: FilterValues
  onChange: (values: FilterValues) => void
}) {
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...values, search: e.target.value })
  }

  const handleStatusChange = (value: string) => {
    onChange({ ...values, status: value })
  }

  const handleProviderChange = (value: string) => {
    onChange({ ...values, provider: value })
  }

  const handleDateRangeChange = (value: string) => {
    onChange({ ...values, dateRange: value })
  }

  return (
    <div className="flex flex-col space-y-4 md:flex-row md:items-center md:justify-between md:space-y-0">
      <div className="relative w-full md:max-w-sm">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search payments..."
          className="w-full rounded-lg bg-background pl-8"
          value={values.search}
          onChange={handleSearchChange}
        />
      </div>
      
      <div className="flex items-center space-x-2">
        <Select value={values.status} onValueChange={handleStatusChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="succeeded">Succeeded</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="needs_attention">Needs Attention</SelectItem>
          </SelectContent>
        </Select>
        
        <Select value={values.provider} onValueChange={handleProviderChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Provider" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Providers</SelectItem>
            <SelectItem value="stripe">Stripe</SelectItem>
            <SelectItem value="xero">Xero</SelectItem>
            <SelectItem value="quickbooks">QuickBooks</SelectItem>
          </SelectContent>
        </Select>
        
        <Select value={values.dateRange} onValueChange={handleDateRangeChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Date Range" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="today">Today</SelectItem>
            <SelectItem value="yesterday">Yesterday</SelectItem>
            <SelectItem value="last7">Last 7 days</SelectItem>
            <SelectItem value="last30">Last 30 days</SelectItem>
            <SelectItem value="custom">Custom Range</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}

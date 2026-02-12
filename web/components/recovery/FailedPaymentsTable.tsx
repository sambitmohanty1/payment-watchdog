'use client'

import { useState, useEffect } from 'react'
import { format } from 'date-fns'
import { MoreHorizontal, Clock, CheckCircle2, XCircle, AlertCircle, Loader2, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import Badge from '@/components/ui/Badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

type PaymentStatus = 'pending' | 'succeeded' | 'failed' | 'needs_attention'

interface FailedPayment {
  id: string
  customer: {
    name: string
    email: string
  }
  amount: number
  status: PaymentStatus
  provider: 'stripe' | 'xero' | 'quickbooks'
  lastAttempt: Date
  nextRetry: Date | null
  retryCount: number
  maxRetries: number
}

const statusIcons = {
  pending: Clock,
  succeeded: CheckCircle2,
  failed: XCircle,
  needs_attention: AlertCircle,
}

const statusColors = {
  pending: 'bg-yellow-100 text-yellow-800',
  succeeded: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  needs_attention: 'bg-orange-100 text-orange-800',
}

const providerColors = {
  stripe: 'bg-purple-100 text-purple-800',
  xero: 'bg-blue-100 text-blue-800',
  quickbooks: 'bg-indigo-100 text-indigo-800',
}

export function FailedPaymentsTable() {
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [failedPayments, setFailedPayments] = useState<FailedPayment[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [itemsPerPage] = useState(10)
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true)
        // TODO: Replace with actual API call
        // const response = await fetch('/api/failed-payments')
        // const data = await response.json()
        // setFailedPayments(data)
        
        // Mock API call with timeout
        await new Promise(resolve => setTimeout(resolve, 1000))
        setFailedPayments([
          {
            id: 'inv_12345',
            customer: { name: 'Acme Corp', email: 'billing@acmecorp.com' },
            amount: 2999,
            status: 'pending',
            provider: 'stripe',
            lastAttempt: new Date('2023-06-15T14:30:00'),
            nextRetry: new Date('2023-06-16T09:00:00'),
            retryCount: 2,
            maxRetries: 5,
          },
          {
            id: 'inv_67890',
            customer: {
              name: 'Globex Corporation',
              email: 'payments@globex.com',
            },
            amount: 4500,
            status: 'failed',
            provider: 'xero',
            lastAttempt: new Date('2023-06-14T10:15:00'),
            nextRetry: null,
            retryCount: 5,
            maxRetries: 5,
          },
          {
            id: 'inv_54321',
            customer: {
              name: 'Initech',
              email: 'accounts@initech.com',
            },
            amount: 1299,
            status: 'succeeded',
            provider: 'quickbooks',
            lastAttempt: new Date('2023-06-16T08:45:00'),
            nextRetry: null,
            retryCount: 1,
            maxRetries: 5,
          },
          {
            id: 'inv_98765',
            customer: {
              name: 'Umbrella Corp',
              email: 'finance@umbrella.com',
            },
            amount: 8999,
            status: 'needs_attention',
            provider: 'stripe',
            lastAttempt: new Date('2023-06-13T16:20:00'),
            nextRetry: new Date('2023-06-17T11:30:00'),
            retryCount: 3,
            maxRetries: 5,
          },
          {
            id: 'inv_13579',
            customer: {
              name: 'Wayne Enterprises',
              email: 'accounts@wayne.com',
            },
            amount: 6500,
            status: 'pending',
            provider: 'xero',
            lastAttempt: new Date('2023-06-16T09:30:00'),
            nextRetry: new Date('2023-06-17T14:00:00'),
            retryCount: 1,
            maxRetries: 3,
          },
        ])
        setIsLoading(false)
      } catch (err) {
        setError('Failed to load payment data. Please try again.')
        setIsLoading(false)
      }
    }
    
    fetchData()
  }, [])

  const toggleRowSelection = (id: string) => {
    const newSelection = new Set(selectedRows)
    if (newSelection.has(id)) {
      newSelection.delete(id)
    } else {
      newSelection.add(id)
    }
    setSelectedRows(newSelection)
  }

  const selectAllRows = () => {
    if (selectedRows.size === failedPayments.length) {
      setSelectedRows(new Set())
    } else {
      setSelectedRows(new Set(failedPayments.map(payment => payment.id)))
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount / 100)
  }

  // Pagination logic
  const indexOfLastItem = currentPage * itemsPerPage
  const indexOfFirstItem = indexOfLastItem - itemsPerPage
  const currentItems = failedPayments.slice(indexOfFirstItem, indexOfLastItem)
  const totalPages = Math.ceil(failedPayments.length / itemsPerPage)

  const paginate = (pageNumber: number) => setCurrentPage(pageNumber)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">Loading payment data...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">
        {error}
        <Button 
          variant="outline" 
          size="sm" 
          className="mt-4"
          onClick={() => window.location.reload()}
        >
          Retry
        </Button>
      </div>
    )
  }

  if (failedPayments.length === 0) {
    return (
      <div className="p-8 text-center text-gray-500">
        No failed payments found. All clear!
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[40px]">
              <input
                type="checkbox"
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                checked={selectedRows.size === failedPayments.length && failedPayments.length > 0}
                onChange={selectAllRows}
              />
            </TableHead>
            <TableHead>Customer</TableHead>
            <TableHead>Amount</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Provider</TableHead>
            <TableHead>Last Attempt</TableHead>
            <TableHead>Next Retry</TableHead>
            <TableHead>Retry Count</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {currentItems.map((payment) => {
            const StatusIcon = statusIcons[payment.status]
            const statusColor = statusColors[payment.status]
            const providerColor = providerColors[payment.provider]
            
            return (
              <TableRow key={payment.id}>
                <TableCell>
                  <input
                    type="checkbox"
                    className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    checked={selectedRows.has(payment.id)}
                    onChange={() => toggleRowSelection(payment.id)}
                  />
                </TableCell>
                <TableCell>
                  <div className="font-medium">{payment.customer.name}</div>
                  <div className="text-sm text-gray-500">{payment.customer.email}</div>
                </TableCell>
                <TableCell className="font-medium">{formatCurrency(payment.amount)}</TableCell>
                <TableCell>
                  <Badge className={`${statusColor} flex items-center gap-1`}>
                    <StatusIcon className="h-3.5 w-3.5" />
                    {payment.status.replace('_', ' ')}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge className={providerColor}>
                    {payment.provider}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div>{format(payment.lastAttempt, 'MMM d, yyyy')}</div>
                  <div className="text-sm text-gray-500">
                    {format(payment.lastAttempt, 'h:mm a')}
                  </div>
                </TableCell>
                <TableCell>
                  {payment.nextRetry ? (
                    <>
                      <div>{format(payment.nextRetry, 'MMM d, yyyy')}</div>
                      <div className="text-sm text-gray-500">
                        {format(payment.nextRetry, 'h:mm a')}
                      </div>
                    </>
                  ) : (
                    <span className="text-gray-400">-</span>
                  )}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <div className="w-16 bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-blue-600 h-2 rounded-full" 
                        style={{ width: `${(payment.retryCount / payment.maxRetries) * 100}%` }}
                      />
                    </div>
                    <span className="text-sm text-gray-600">
                      {payment.retryCount}/{payment.maxRetries}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontal className="h-4 w-4" />
                        <span className="sr-only">Open menu</span>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem>Retry Now</DropdownMenuItem>
                      <DropdownMenuItem>View Details</DropdownMenuItem>
                      <DropdownMenuItem>Contact Customer</DropdownMenuItem>
                      <DropdownMenuItem className="text-red-600">Cancel Payment</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
      
      {/* Pagination controls */}
      <div className="flex items-center justify-between px-4 py-3 border-t">
        <div className="text-sm text-muted-foreground">
          Showing {indexOfFirstItem + 1}-{Math.min(indexOfLastItem, failedPayments.length)} of {failedPayments.length} payments
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => paginate(currentPage - 1)}
            disabled={currentPage === 1}
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          <div className="flex items-center space-x-1">
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((number) => (
              <Button
                key={number}
                variant={currentPage === number ? "default" : "outline"}
                size="sm"
                onClick={() => paginate(number)}
              >
                {number}
              </Button>
            ))}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => paginate(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}

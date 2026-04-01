'use client'

import { useState } from 'react'
import { useSystemStatus } from '../hooks/useSystemStatus'
import { StatusCard, StatusDetail } from '../components/ui/StatusCard'

interface DetailModal {
  title: string
  status: any
}

export default function HomePage() {
  const [detailModal, setDetailModal] = useState<DetailModal | null>(null)
  
  const { 
    status, 
    loading, 
    error, 
    lastUpdate, 
    isRefreshing, 
    refresh, 
    toggleAutoRefresh, 
    autoRefreshEnabled 
  } = useSystemStatus()

  const handleStatusClick = (title: string, statusData: any) => {
    setDetailModal({ title, status: statusData })
  }

  const closeDetail = () => {
    setDetailModal(null)
  }

  return (
    <div className="space-y-6">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-semibold text-gray-900">
                Payment Watchdog
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-500">
                Environment: {status?.environment || 'Loading...'}
              </span>
              <span className="inline-flex h-2 w-2 rounded-full bg-green-400"></span>
              <button
                onClick={toggleAutoRefresh}
                className={`px-2 py-1 text-xs rounded ${
                  autoRefreshEnabled 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                Auto-refresh: {autoRefreshEnabled ? 'ON' : 'OFF'}
              </button>
              <button
                onClick={refresh}
                disabled={isRefreshing}
                className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded disabled:opacity-50"
              >
                {isRefreshing ? 'Refreshing...' : 'Refresh'}
              </button>
            </div>
          </div>
        </div>
      </header>
      
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
            <div className="flex">
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Error fetching system status
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  {error}
                </div>
                <div className="mt-3">
                  <button
                    onClick={refresh}
                    className="bg-red-100 text-red-800 px-3 py-1 rounded text-sm"
                  >
                    Retry
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h2 className="text-lg font-medium text-gray-900">
              Payment Watchdog Dashboard
            </h2>
            <div className="mt-2 max-w-xl text-sm text-gray-500">
              Monitor and manage payment recovery workflows
            </div>
            {lastUpdate && (
              <div className="mt-2 text-xs text-gray-400">
                Last updated: {new Date(lastUpdate).toLocaleString()}
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          <StatusCard
            title="API Service"
            status={status?.api || { status: 'unknown', lastCheck: 'Never' }}
            icon="API"
            onClick={() => handleStatusClick('API Service', status?.api)}
            loading={loading}
          />
          
          <StatusCard
            title="Database"
            status={status?.database || { status: 'unknown', lastCheck: 'Never' }}
            icon="DB"
            onClick={() => handleStatusClick('Database', status?.database)}
            loading={loading}
          />
          
          <StatusCard
            title="Worker Service"
            status={status?.workers || { status: 'unknown', lastCheck: 'Never' }}
            icon="WRK"
            onClick={() => handleStatusClick('Worker Service', status?.workers)}
            loading={loading}
          />
          
          <StatusCard
            title="Redis Cache"
            status={status?.redis || { status: 'unknown', lastCheck: 'Never' }}
            icon="RDS"
            onClick={() => handleStatusClick('Redis Cache', status?.redis)}
            loading={loading}
          />
        </div>

        {detailModal && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
              <div className="mt-3">
                <h3 className="text-lg font-medium text-gray-900">
                  {detailModal.title} Details
                </h3>
                <div className="mt-2 px-7 py-3">
                  <StatusDetail status={detailModal.status} />
                </div>
                <div className="items-center px-4 py-3">
                  <button
                    onClick={closeDetail}
                    className="px-4 py-2 bg-blue-500 text-white text-base font-medium rounded-md w-full shadow-sm hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-300"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

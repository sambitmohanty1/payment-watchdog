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
  const colorClasses = getStatusColor(status.status)
  const statusIcon = getStatusIcon(status.status)
  
  return (
    <div 
      className={`
        bg-white overflow-hidden shadow rounded-lg transition-all duration-200
        hover:shadow-md cursor-pointer border-2 border-transparent
        ${colorClasses}
        ${className}
      `}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          onClick?.()
        }
      }}
      aria-label={`View details for ${title} - Status: ${status.status}`}
    >
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`
              w-8 h-8 rounded-md flex items-center justify-center text-sm font-medium
              ${status.status === 'healthy' ? 'bg-green-500 text-white' : ''}
              ${status.status === 'degraded' ? 'bg-yellow-500 text-white' : ''}
              ${status.status === 'down' ? 'bg-red-500 text-white' : ''}
              ${status.status === 'unknown' ? 'bg-gray-500 text-white' : ''}
            `}>
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              ) : (
                <span>{icon}</span>
              )}
            </div>
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">
                {title}
              </dt>
              <dd className="mt-1 text-3xl font-semibold">
                <span className={`
                  ${status.status === 'healthy' ? 'text-green-900' : ''}
                  ${status.status === 'degraded' ? 'text-yellow-900' : ''}
                  ${status.status === 'down' ? 'text-red-900' : ''}
                  ${status.status === 'unknown' ? 'text-gray-900' : ''}
                `}>
                  {loading ? 'Loading...' : status.status}
                </span>
              </dd>
              <dd className="mt-1 text-sm text-gray-500">
                {status.error ? (
                  <span className="text-red-600">{status.error}</span>
                ) : (
                  <span>
                    {status.responseTime && `Response: ${status.responseTime}ms`}
                    {!status.responseTime && 'Checking status...'}
                  </span>
                )}
              </dd>
              <dd className="mt-1 text-xs text-gray-400">
                Last check: {new Date(status.lastCheck).toLocaleTimeString()}
              </dd>
            </dl>
          </div>
          <div className="flex-shrink-0 ml-2">
            <span className="text-2xl" role="img" aria-label={`Status: ${status.status}`}>
              {statusIcon}
            </span>
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
  const colorClasses = getStatusColor(status.status)
  const statusIcon = getStatusIcon(status.status)
  
  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="mt-3">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              {title} Details
            </h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
              aria-label="Close details"
            >
              ✕
            </button>
          </div>
          
          <div className="mt-4 space-y-4">
            <div className={`
              p-3 rounded-md border ${colorClasses}
            `}>
              <div className="flex items-center">
                <span className="text-2xl mr-3">{statusIcon}</span>
                <div>
                  <p className="font-medium">Status: {status.status}</p>
                  <p className="text-sm opacity-75">
                    Last check: {new Date(status.lastCheck).toLocaleString()}
                  </p>
                </div>
              </div>
            </div>
            
            {status.responseTime && (
              <div className="bg-gray-50 p-3 rounded-md">
                <p className="font-medium">Response Time</p>
                <p className="text-sm text-gray-600">{status.responseTime}ms</p>
              </div>
            )}
            
            {status.error && (
              <div className="bg-red-50 border border-red-200 p-3 rounded-md">
                <p className="font-medium text-red-800">Error Details</p>
                <p className="text-sm text-red-600">{status.error}</p>
              </div>
            )}
            
            <div className="bg-gray-50 p-3 rounded-md">
              <p className="font-medium">Recommended Actions</p>
              <ul className="text-sm text-gray-600 mt-1 space-y-1">
                {status.status === 'healthy' && (
                  <li>• No action required - service is operating normally</li>
                )}
                {status.status === 'degraded' && (
                  <>
                    <li>• Monitor service performance closely</li>
                    <li>• Check system resources</li>
                    <li>• Review recent logs for warnings</li>
                  </>
                )}
                {status.status === 'down' && (
                  <>
                    <li>• Immediate investigation required</li>
                    <li>• Check service logs</li>
                    <li>• Verify system resources</li>
                    <li>• Consider service restart</li>
                  </>
                )}
                {status.status === 'unknown' && (
                  <>
                    <li>• Verify service connectivity</li>
                    <li>• Check network configuration</li>
                    <li>• Review health check endpoint</li>
                  </>
                )}
              </ul>
            </div>
          </div>
          
          <div className="mt-6 flex justify-end space-x-3">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

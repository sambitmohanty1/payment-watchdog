import { useState, useEffect, useCallback, useRef } from 'react'
import axios, { AxiosError } from 'axios'

// Types for system status
export interface ServiceStatus {
  status: 'healthy' | 'degraded' | 'down' | 'unknown'
  lastCheck: string
  responseTime?: number
  error?: string
}

export interface SystemStatus {
  api: ServiceStatus
  database: ServiceStatus
  workers: ServiceStatus
  redis: ServiceStatus
  environment: string
  version: string
  timestamp: string
}

export interface UseSystemStatusOptions {
  refreshInterval?: number
  retryAttempts?: number
  retryDelay?: number
  enableAutoRefresh?: boolean
}

export interface UseSystemStatusReturn {
  status: SystemStatus | null
  loading: boolean
  error: string | null
  lastUpdate: Date | null
  isRefreshing: boolean
  refresh: () => Promise<void>
  toggleAutoRefresh: () => void
  autoRefreshEnabled: boolean
}

const DEFAULT_OPTIONS: UseSystemStatusOptions = {
  refreshInterval: 30000, // 30 seconds
  retryAttempts: 3,
  retryDelay: 1000, // 1 second
  enableAutoRefresh: true,
}

// API client configuration
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8091',
  timeout: 10000, // 10 seconds timeout
  headers: {
    'Content-Type': 'application/json',
  },
})

// Retry logic with exponential backoff
const retryWithBackoff = async <T>(
  fn: () => Promise<T>,
  attempts: number,
  delay: number
): Promise<T> => {
  try {
    return await fn()
  } catch (error) {
    if (attempts <= 1) throw error
    
    // Wait before retry
    await new Promise(resolve => setTimeout(resolve, delay))
    
    // Exponential backoff
    return retryWithBackoff(fn, attempts - 1, delay * 2)
  }
}

// Transform API response to our interface
const transformApiResponse = (apiResponse: any): SystemStatus => {
  const timestamp = new Date().toISOString()
  
  return {
    api: {
      status: apiResponse.status === 'healthy' ? 'healthy' : 
             apiResponse.status === 'degraded' ? 'degraded' : 
             apiResponse.status === 'unhealthy' ? 'down' : 'unknown',
      lastCheck: timestamp,
      responseTime: apiResponse.responseTime,
    },
    database: {
      status: apiResponse.database?.status === 'connected' ? 'healthy' :
             apiResponse.database?.status === 'disconnected' ? 'down' :
             apiResponse.database?.status === 'error' ? 'degraded' : 'unknown',
      lastCheck: timestamp,
      responseTime: apiResponse.database?.responseTime,
      error: apiResponse.database?.error,
    },
    workers: {
      status: apiResponse.workers?.status === 'active' ? 'healthy' :
             apiResponse.workers?.status === 'inactive' ? 'down' :
             apiResponse.workers?.status === 'error' ? 'degraded' : 'unknown',
      lastCheck: timestamp,
      responseTime: apiResponse.workers?.responseTime,
      error: apiResponse.workers?.error,
    },
    redis: {
      status: apiResponse.redis?.status === 'connected' ? 'healthy' :
             apiResponse.redis?.status === 'disconnected' ? 'down' :
             apiResponse.redis?.status === 'error' ? 'degraded' : 'unknown',
      lastCheck: timestamp,
      responseTime: apiResponse.redis?.responseTime,
      error: apiResponse.redis?.error,
    },
    environment: apiResponse.environment || 'unknown',
    version: apiResponse.version || 'unknown',
    timestamp,
  }
}

export const useSystemStatus = (options: UseSystemStatusOptions = {}): UseSystemStatusReturn => {
  const config = { ...DEFAULT_OPTIONS, ...options }
  
  // State management
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState(config.enableAutoRefresh)
  
  // Refs for cleanup
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const mountedRef = useRef(true)
  
  // Fetch system status
  const fetchStatus = useCallback(async (showRefreshing = false) => {
    if (!mountedRef.current) return
    
    try {
      if (showRefreshing) setIsRefreshing(true)
      setError(null)
      
      const response = await retryWithBackoff(
        () => apiClient.get('/api/health'),
        config.retryAttempts!,
        config.retryDelay!
      )
      
      if (mountedRef.current) {
        const transformedStatus = transformApiResponse(response.data)
        setStatus(transformedStatus)
        setLastUpdate(new Date())
        setLoading(false)
        setError(null)
      }
    } catch (err) {
      if (mountedRef.current) {
        const error = err as AxiosError
        
        // Provide user-friendly error messages
        let errorMessage = 'Failed to fetch system status'
        
        if (error.code === 'ECONNABORTED') {
          errorMessage = 'Request timeout - please check your connection'
        } else if (error.response?.status === 404) {
          errorMessage = 'Health check endpoint not available'
        } else if (error.response?.status && error.response.status >= 500) {
          errorMessage = 'Server error - please try again later'
        } else if (error.code === 'NETWORK_ERROR') {
          errorMessage = 'Network error - please check your connection'
        }
        
        setError(errorMessage)
        setLoading(false)
        
        // Set degraded status when API is unavailable
        setStatus({
          api: { status: 'down', lastCheck: new Date().toISOString(), error: errorMessage },
          database: { status: 'unknown', lastCheck: new Date().toISOString() },
          workers: { status: 'unknown', lastCheck: new Date().toISOString() },
          redis: { status: 'unknown', lastCheck: new Date().toISOString() },
          environment: 'unknown',
          version: 'unknown',
          timestamp: new Date().toISOString(),
        })
      }
    } finally {
      if (mountedRef.current) {
        setIsRefreshing(false)
      }
    }
  }, [config.retryAttempts, config.retryDelay])
  
  // Manual refresh function
  const refresh = useCallback(async () => {
    await fetchStatus(true)
  }, [fetchStatus])
  
  // Toggle auto-refresh
  const toggleAutoRefresh = useCallback(() => {
    setAutoRefreshEnabled(prev => !prev)
  }, [])
  
  // Setup auto-refresh
  useEffect(() => {
    if (autoRefreshEnabled && config.refreshInterval! > 0) {
      intervalRef.current = setInterval(() => {
        fetchStatus(false)
      }, config.refreshInterval!)
      
      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current)
          intervalRef.current = null
        }
      }
    }
  }, [autoRefreshEnabled, config.refreshInterval, fetchStatus])
  
  // Initial fetch
  useEffect(() => {
    fetchStatus(false)
    
    // Cleanup
    return () => {
      mountedRef.current = false
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [fetchStatus])
  
  // Pause auto-refresh when page is not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Page is hidden, pause auto-refresh
        if (intervalRef.current) {
          clearInterval(intervalRef.current)
          intervalRef.current = null
        }
      } else if (autoRefreshEnabled) {
        // Page is visible, resume auto-refresh
        fetchStatus(false)
      }
    }
    
    document.addEventListener('visibilitychange', handleVisibilityChange)
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [autoRefreshEnabled, fetchStatus])
  
  return {
    status,
    loading,
    error,
    lastUpdate,
    isRefreshing,
    refresh,
    toggleAutoRefresh,
    autoRefreshEnabled: autoRefreshEnabled || false,
  }
}

// Utility function to get status color
export const getStatusColor = (status: ServiceStatus['status']): string => {
  switch (status) {
    case 'healthy':
      return 'text-green-600 bg-green-100 border-green-200'
    case 'degraded':
      return 'text-yellow-600 bg-yellow-100 border-yellow-200'
    case 'down':
      return 'text-red-600 bg-red-100 border-red-200'
    case 'unknown':
    default:
      return 'text-gray-600 bg-gray-100 border-gray-200'
  }
}

// Utility function to get status icon
export const getStatusIcon = (status: ServiceStatus['status']): string => {
  switch (status) {
    case 'healthy':
      return '✅'
    case 'degraded':
      return '⚠️'
    case 'down':
      return '❌'
    case 'unknown':
    default:
      return '❓'
  }
}

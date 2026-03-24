# Payment Watchdog - Technical Debt Backlog

## 🎯 Overview
This document tracks technical debt identified during the Sovereign deployment review. Items are prioritized by business impact and security requirements.

---

## 🚨 P0 - Critical Security & Production Issues

### [P0-001] Implement Dynamic Status Dashboard (CRITICAL)
**Priority**: Critical  
**Impact**: Dashboard displays hardcoded fake status values, creating false confidence in system health  
**Estimated Effort**: 2 days  
**Owner**: Frontend Team + Backend Team  

**Description**: The UI currently shows hardcoded status values ("Healthy", "Connected", "Active") regardless of actual system state. This creates a dangerous false sense of security and prevents operators from detecting real issues.

**Business Impact**:
- **False Confidence**: Users think system is healthy when it's not
- **Operational Blindness**: Cannot detect real system failures
- **Trust Issues**: Dashboard becomes unreliable and ignored
- **Incident Response**: Delayed detection of actual problems

**Current State Analysis**:
```tsx
// CURRENT: All hardcoded values
<dd className="mt-1 text-3xl font-semibold text-gray-900">Healthy</dd>     // API
<dd className="mt-1 text-3xl font-semibold text-gray-900">Connected</dd>   // Database  
<dd className="mt-1 text-3xl font-semibold text-gray-900">Active</dd>      // Workers
<span>Environment: Staging</span>  // Environment
```

**Acceptance Criteria**:
- [ ] **Real-time API Status**: Fetch actual API health from `/api/health` endpoint
- [ ] **Database Connectivity**: Show real database connection status and metrics
- [ ] **Worker Activity**: Display actual worker status and last execution time
- [ ] **Environment Detection**: Auto-detect and display current environment
- [ ] **Error States**: Show appropriate error states when services are down
- [ ] **Auto-refresh**: Update status every 30 seconds automatically
- [ ] **Loading States**: Show loading indicators during data fetch
- [ ] **Error Handling**: Graceful error handling with retry mechanisms
- [ ] **Visual Indicators**: Color-coded status (green=healthy, yellow=degraded, red=down)
- [ ] **Timestamps**: Show last update time for each status component

**Technical Implementation Design**:

#### **Backend API Enhancement**:
```go
// File: api/internal/handlers/health.go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
)

type HealthStatus struct {
    API         APIStatus         `json:"api"`
    Database    DatabaseStatus    `json:"database"`
    Redis       RedisStatus       `json:"redis"`
    Workers     WorkerStatus      `json:"workers"`
    Environment EnvironmentStatus `json:"environment"`
    System      SystemStatus      `json:"system"`
    Timestamp   time.Time         `json:"timestamp"`
}

type APIStatus struct {
    Status    string    `json:"status"`    // healthy, degraded, down
    Version   string    `json:"version"`
    Uptime    string    `json:"uptime"`
    LastCheck time.Time `json:"last_check"`
    Response  string    `json:"response_time"`
}

type DatabaseStatus struct {
    Status     string `json:"status"`      // connected, disconnected, error
    Host       string `json:"host"`
    Connections int   `json:"connections"`
    Latency    string `json:"latency"`
    LastQuery  string `json:"last_query"`
}

type RedisStatus struct {
    Status      string `json:"status"`       // connected, disconnected, error
    Host        string `json:"host"`
    Connections  int    `json:"connections"`
    Memory      string `json:"memory_used"`
    LastCommand string `json:"last_command"`
}

type WorkerStatus struct {
    Status    string    `json:"status"`    // active, idle, error
    Count     int       `json:"count"`
    LastRun   time.Time `json:"last_run"`
    NextRun   time.Time `json:"next_run"`
    Running   []string  `json:"running_jobs"`
    Failed    int       `json:"failed_jobs"`
}

type EnvironmentStatus struct {
    Name      string `json:"name"`       // staging, production
    Version   string `json:"version"`    // sovereign-au
    Region    string `json:"region"`     // ap-melbourne-1
    Namespace string `json:"namespace"`  // sovereign-au
}

type SystemStatus struct {
    CPU    string `json:"cpu_usage"`
    Memory string `json:"memory_usage"`
    Disk   string `json:"disk_usage"`
    Load   string `json:"system_load"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    // Check API health
    apiStatus := checkAPIHealth()
    
    // Check database connectivity
    dbStatus := checkDatabaseHealth()
    
    // Check Redis connectivity
    redisStatus := checkRedisHealth()
    
    // Check worker status
    workerStatus := checkWorkerHealth()
    
    // Get environment info
    envStatus := getEnvironmentStatus()
    
    // Get system metrics
    systemStatus := getSystemMetrics()
    
    health := HealthStatus{
        API:         apiStatus,
        Database:    dbStatus,
        Redis:       redisStatus,
        Workers:     workerStatus,
        Environment: envStatus,
        System:      systemStatus,
        Timestamp:   time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    
    // Return appropriate HTTP status based on overall health
    overallStatus := getOverallHealth(health)
    if overallStatus == "down" {
        w.WriteHeader(http.StatusServiceUnavailable)
    } else if overallStatus == "degraded" {
        w.WriteHeader(http.StatusOK) // But with degraded status
    }
    
    json.NewEncoder(w).Encode(health)
}

func checkAPIHealth() APIStatus {
    // Check if API is responding
    start := time.Now()
    
    // Database ping test
    dbErr := database.Ping()
    
    // Redis ping test
    redisErr := redis.Ping()
    
    responseTime := time.Since(start)
    
    status := "healthy"
    if dbErr != nil || redisErr != nil {
        status = "degraded"
    }
    
    return APIStatus{
        Status:    status,
        Version:   os.Getenv("APP_VERSION"),
        Uptime:    getUptime(),
        LastCheck: time.Now(),
        Response:  responseTime.String(),
    }
}

func checkDatabaseHealth() DatabaseStatus {
    // Test database connectivity
    start := time.Now()
    err := database.Ping()
    latency := time.Since(start)
    
    status := "connected"
    if err != nil {
        status = "disconnected"
    }
    
    // Get connection pool stats
    stats := database.Stats()
    
    return DatabaseStatus{
        Status:      status,
        Host:        getDatabaseHost(),
        Connections: stats.OpenConnections,
        Latency:     latency.String(),
        LastQuery:   getLastQueryTime(),
    }
}

func checkRedisHealth() RedisStatus {
    // Test Redis connectivity
    start := time.Now()
    err := redis.Ping()
    latency := time.Since(start)
    
    status := "connected"
    if err != nil {
        status = "disconnected"
    }
    
    // Get Redis info
    info := redis.Info()
    
    return RedisStatus{
        Status:      status,
        Host:        getRedisHost(),
        Connections: getRedisConnections(),
        Memory:      getRedisMemoryUsage(info),
        LastCommand: getLastRedisCommand(),
    }
}

func checkWorkerHealth() WorkerStatus {
    // Check if workers are running
    workers := getWorkerStatus()
    
    status := "active"
    if len(workers.Running) == 0 {
        status = "idle"
    }
    if workers.FailedJobs > 10 {
        status = "error"
    }
    
    return WorkerStatus{
        Status:    status,
        Count:     workers.Count,
        LastRun:   workers.LastRun,
        NextRun:   workers.NextRun,
        Running:   workers.Running,
        Failed:    workers.FailedJobs,
    }
}

func getEnvironmentStatus() EnvironmentStatus {
    return EnvironmentStatus{
        Name:      os.Getenv("ENVIRONMENT"),
        Version:   os.Getenv("APP_VERSION"),
        Region:    os.Getenv("OCI_REGION"),
        Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
    }
}

func getSystemMetrics() SystemStatus {
    // Get system metrics from /proc or Kubernetes metrics
    return SystemStatus{
        CPU:    getCPUUsage(),
        Memory: getMemoryUsage(),
        Disk:   getDiskUsage(),
        Load:   getSystemLoad(),
    }
}
```

#### **Frontend React Hook**:
```tsx
// File: ui/hooks/useSystemStatus.ts
import { useState, useEffect, useCallback } from 'react'
import axios, { AxiosError } from 'axios'

interface SystemStatus {
  api: {
    status: 'healthy' | 'degraded' | 'down'
    version: string
    uptime: string
    last_check: string
    response_time: string
  }
  database: {
    status: 'connected' | 'disconnected' | 'error'
    host: string
    connections: number
    latency: string
    last_query: string
  }
  redis: {
    status: 'connected' | 'disconnected' | 'error'
    host: string
    connections: number
    memory_used: string
    last_command: string
  }
  workers: {
    status: 'active' | 'idle' | 'error'
    count: number
    last_run: string
    next_run: string
    running_jobs: string[]
    failed_jobs: number
  }
  environment: {
    name: string
    version: string
    region: string
    namespace: string
  }
  system: {
    cpu_usage: string
    memory_usage: string
    disk_usage: string
    system_load: string
  }
  timestamp: string
}

interface UseSystemStatusReturn {
  status: SystemStatus | null
  loading: boolean
  error: string | null
  lastUpdate: Date | null
  refresh: () => void
  isHealthy: boolean
  overallStatus: 'healthy' | 'degraded' | 'down'
}

export function useSystemStatus(refreshInterval: number = 30000): UseSystemStatusReturn {
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)

  const fetchStatus = useCallback(async () => {
    try {
      setError(null)
      setLoading(true)
      
      const response = await axios.get<SystemStatus>('/api/health', {
        timeout: 10000, // 10 second timeout
        headers: {
          'Cache-Control': 'no-cache',
        },
      })
      
      setStatus(response.data)
      setLastUpdate(new Date())
    } catch (err) {
      const axiosError = err as AxiosError
      
      if (axiosError.code === 'ECONNABORTED') {
        setError('Request timeout - server not responding')
      } else if (axiosError.response) {
        setError(`Server error: ${axiosError.response.status}`)
      } else if (axiosError.request) {
        setError('Network error - unable to reach server')
      } else {
        setError('Unknown error occurred')
      }
      
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }, [])

  const refresh = useCallback(() => {
    fetchStatus()
  }, [fetchStatus])

  useEffect(() => {
    fetchStatus()
    
    if (refreshInterval > 0) {
      const interval = setInterval(fetchStatus, refreshInterval)
      return () => clearInterval(interval)
    }
  }, [fetchStatus, refreshInterval])

  const isHealthy = status ? 
    status.api.status === 'healthy' && 
    status.database.status === 'connected' && 
    status.redis.status === 'connected' : false

  const overallStatus = status ? 
    (isHealthy ? 'healthy' : 
     status.api.status === 'down' || status.database.status === 'disconnected' ? 'down' : 'degraded') : 
    'down'

  return {
    status,
    loading,
    error,
    lastUpdate,
    refresh,
    isHealthy,
    overallStatus,
  }
}
```

#### **Updated Dashboard Component**:
```tsx
// File: ui/app/page.tsx
'use client'

import { useSystemStatus } from '@/hooks/useSystemStatus'

function StatusCard({ 
  title, 
  status, 
  icon, 
  details, 
  color = 'blue' 
}: {
  title: string
  status: string
  icon: string
  details?: string
  color?: string
}) {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'healthy':
      case 'connected':
      case 'active':
        return 'bg-green-500 text-green-900'
      case 'degraded':
      case 'idle':
        return 'bg-yellow-500 text-yellow-900'
      case 'down':
      case 'disconnected':
      case 'error':
        return 'bg-red-500 text-red-900'
      default:
        return 'bg-gray-500 text-gray-900'
    }
  }

  const statusColor = getStatusColor(status)

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`w-8 h-8 ${statusColor.includes('green') ? 'bg-green-500' : 
                           statusColor.includes('yellow') ? 'bg-yellow-500' : 
                           statusColor.includes('red') ? 'bg-red-500' : 'bg-gray-500'} 
                           rounded-md flex items-center justify-center`}>
              <span className="text-white text-sm font-medium">{icon}</span>
            </div>
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">
                {title}
              </dt>
              <dd className={`mt-1 text-3xl font-semibold ${statusColor}`}>
                {status}
              </dd>
              {details && (
                <dd className="mt-1 text-sm text-gray-500">
                  {details}
                </dd>
              )}
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function HomePage() {
  const { status, loading, error, lastUpdate, refresh, isHealthy, overallStatus } = useSystemStatus()

  if (loading && !status) {
    return (
      <div className="space-y-6">
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-2 text-gray-600">Loading system status...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <div className="w-5 h-5 bg-red-400 rounded-full"></div>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Unable to fetch system status
              </h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{error}</p>
              </div>
              <div className="mt-3">
                <button
                  onClick={refresh}
                  className="bg-red-100 text-red-800 px-3 py-1 rounded-md text-sm font-medium hover:bg-red-200"
                >
                  Retry
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-medium text-gray-900">
                Payment Watchdog Dashboard
              </h2>
              <div className="mt-2 max-w-xl text-sm text-gray-500">
                Monitor and manage payment recovery workflows
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className="text-right">
                <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  overallStatus === 'healthy' ? 'bg-green-100 text-green-800' :
                  overallStatus === 'degraded' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-red-100 text-red-800'
                }`}>
                  {overallStatus === 'healthy' ? 'All Systems Operational' :
                   overallStatus === 'degraded' ? 'Some Issues Detected' :
                   'System Down'}
                </div>
                {lastUpdate && (
                  <div className="text-xs text-gray-500 mt-1">
                    Last updated: {lastUpdate.toLocaleTimeString()}
                  </div>
                )}
              </div>
              <button
                onClick={refresh}
                disabled={loading}
                className="inline-flex items-center px-3 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                {loading ? 'Refreshing...' : 'Refresh'}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <StatusCard
          title="API Status"
          status={status?.api.status || 'Unknown'}
          icon="API"
          details={`${status?.api.response_time || 'N/A'} response time`}
        />
        
        <StatusCard
          title="Database"
          status={status?.database.status || 'Unknown'}
          icon="DB"
          details={`${status?.database.connections || 0} connections`}
        />
        
        <StatusCard
          title="Redis Cache"
          status={status?.redis.status || 'Unknown'}
          icon="R"
          details={status?.redis.memory_used || 'N/A'}
        />
        
        <StatusCard
          title="Workers"
          status={status?.workers.status || 'Unknown'}
          icon="W"
          details={`${status?.workers.running_jobs?.length || 0} running`}
        />
      </div>

      {/* Environment Info */}
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Environment Information
          </h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">Environment</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.environment.name || 'Unknown'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Version</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.environment.version || 'Unknown'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Region</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.environment.region || 'Unknown'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Namespace</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.environment.namespace || 'Unknown'}</dd>
            </div>
          </div>
        </div>
      </div>

      {/* System Metrics */}
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            System Metrics
          </h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">CPU Usage</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.system.cpu_usage || 'N/A'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Memory Usage</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.system.memory_usage || 'N/A'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Disk Usage</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.system.disk_usage || 'N/A'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">System Load</dt>
              <dd className="mt-1 text-sm text-gray-900">{status?.system.system_load || 'N/A'}</dd>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Quick Actions
          </h3>
          <div className="space-y-3">
            <button 
              onClick={refresh}
              disabled={loading}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Testing...' : 'Test API Connection'}
            </button>
            <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50">
              View Recovery Workflows
            </button>
            <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50">
              Download System Logs
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
```

**Implementation Steps**:
1. **Day 1**: Backend API health endpoint implementation
2. **Day 1**: Frontend React hook and error handling
3. **Day 2**: Dashboard component updates and styling
4. **Day 2**: Testing, validation, and deployment

**Success Metrics**:
- [ ] Dashboard shows real system status (100% accuracy)
- [ ] Status updates every 30 seconds automatically
- [ ] Error states displayed clearly when services are down
- [ ] Loading states shown during data fetch
- [ ] Color-coded status indicators (green/yellow/red)
- [ ] Timestamps show last update time

---

### [P0-002] Implement Authentication & Authorization
**Priority**: Critical  
**Impact**: Security vulnerability - services exposed without authentication  
**Estimated Effort**: 3 days  
**Owner**: Security Team  

**Description**: External services (UI, API, Recovery) are accessible without any authentication mechanism.

**Acceptance Criteria**:
- [ ] Implement JWT-based authentication for API
- [ ] Add OAuth2 integration for UI
- [ ] Create role-based access control (RBAC)
- [ ] Implement API key authentication for Recovery service
- [ ] Add rate limiting to prevent abuse

**Technical Implementation**:
```yaml
# Add authentication middleware
# Implement JWT tokens
# Configure OAuth2 providers
# Create user management system
```

---

### [P0-003] Implement External Access Monitoring
**Priority**: Critical  
**Impact**: No visibility into external service usage and potential security breaches  
**Estimated Effort**: 2 days  
**Owner**: DevOps Team  

**Description**: Current monitoring only covers internal services. External access patterns are not tracked.

**Acceptance Criteria**:
- [ ] Configure Prometheus metrics for external access
- [ ] Set up Grafana dashboards for external service monitoring
- [ ] Implement alerting for unusual access patterns
- [ ] Add request/response logging for external endpoints
- [ ] Create security incident detection rules

**Technical Implementation**:
```yaml
# Add ServiceMonitor for LoadBalancer services
# Configure external access metrics
# Implement security alerting
# Create access pattern analysis
```

---

### [P0-004] Implement Cost Controls
**Priority**: Critical  
**Impact**: Oracle Cloud costs could exceed free tier without monitoring  
**Estimated Effort**: 1 day  
**Owner**: DevOps Team  

**Description**: No cost monitoring or alerts for LoadBalancer usage and potential cost overruns.

**Acceptance Criteria**:
- [ ] Configure Oracle Cloud cost alerts
- [ ] Set up LoadBalancer usage monitoring
- [ ] Create cost optimization recommendations
- [ ] Implement automated cost reporting
- [ ] Add budget limits and notifications

**Technical Implementation**:
```yaml
# Configure OCI cost monitoring
# Set up budget alerts
# Create cost optimization scripts
# Implement automated reporting
```

---

## 🎯 P1 - High Priority Technical Improvements

### [P1-001] Standardize Service Port Configuration
**Priority**: High  
**Impact**: Inconsistent port configuration creates operational complexity  
**Estimated Effort**: 1 day  
**Owner**: Backend Team  

**Description**: Services use different ports (8085, 8086, 3001) creating confusion in configuration.

**Acceptance Criteria**:
- [ ] Standardize all services to use port 80 internally
- [ ] Update LoadBalancer configurations
- [ ] Update health check configurations
- [ ] Update documentation
- [ ] Validate all external access still works

**Technical Implementation**:
```yaml
# Standard port configuration:
apiVersion: v1
kind: Service
spec:
  ports:
    - name: http
      port: 80
      targetPort: 8080  # Standardized internal port
```

---

### [P1-002] Implement Network Policies
**Priority**: High  
**Impact**: Removed network policy creates security exposure  
**Estimated Effort**: 2 days  
**Owner**: Security Team  

**Description**: Network policy was removed for quick fix, leaving services exposed.

**Acceptance Criteria**:
- [ ] Create granular network policies for each service
- [ ] Allow LoadBalancer traffic only
- [ ] Implement inter-service communication rules
- [ ] Add database/Redis access controls
- [ ] Create security policy documentation

**Technical Implementation**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: payment-watchdog-network-policy
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
```

---

### [P1-003] Fix Health Check Endpoints
**Priority**: High  
**Impact**: API and Recovery services in CrashLoopBackOff due to health check failures  
**Estimated Effort**: 2 days  
**Owner**: Backend Team  

**Description**: Health check endpoints are not properly configured, causing service restarts.

**Acceptance Criteria**:
- [ ] Fix API service to respond to /health endpoint
- [ ] Fix Recovery service to respond to /health endpoint
- [ ] Standardize health check response format
- [ ] Add detailed health check endpoints
- [ ] Implement readiness/liveness probe tuning

**Technical Implementation**:
```go
// Add health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

---

### [P1-004] Implement Service Discovery Configuration
**Priority**: High  
**Impact**: Hardcoded service FQDNs create deployment inflexibility  
**Estimated Effort**: 1 day  
**Owner**: Backend Team  

**Description**: Service endpoints are hardcoded in deployment configurations.

**Acceptance Criteria**:
- [ ] Create ConfigMap for service endpoints
- [ ] Update all services to use ConfigMap values
- [ ] Implement environment-specific configuration
- [ ] Add service discovery validation
- [ ] Update deployment documentation

**Technical Implementation**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-endpoints
data:
  DATABASE_HOST: "lexure-postgres-sovereign-au.sovereign-au.svc.cluster.local"
  REDIS_HOST: "lexure-redis-sovereign-au.sovereign-au.svc.cluster.local"
```

---

## 🔧 P2 - Medium Priority Infrastructure Improvements

### [P2-001] Optimize Resource Allocations
**Priority**: Medium  
**Impact**: Generic resource limits may not match actual usage patterns  
**Estimated Effort**: 2 days  
**Owner**: DevOps Team  

**Description**: Current resource allocations are generic and not optimized for actual usage.

**Acceptance Criteria**:
- [ ] Analyze current resource usage patterns
- [ ] Create service-specific resource profiles
- [ ] Implement resource requests/limits optimization
- [ ] Add resource usage monitoring
- [ ] Create auto-scaling policies

**Technical Implementation**:
```yaml
# Resource profiles for different services
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

---

### [P2-002] Implement Backup and Disaster Recovery
**Priority**: Medium  
**Impact**: No backup strategy for data and configurations  
**Estimated Effort**: 3 days  
**Owner**: DevOps Team  

**Description**: No automated backup or disaster recovery procedures in place.

**Acceptance Criteria**:
- [ ] Configure automated database backups
- [ ] Implement configuration backup
- [ ] Create disaster recovery procedures
- [ ] Test backup restoration procedures
- [ ] Create backup monitoring and alerting

**Technical Implementation**:
```yaml
# Add backup CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
```

---

### [P2-003] Add SSL/TLS Termination
**Priority**: Medium  
**Impact**: External services use HTTP instead of HTTPS  
**Estimated Effort**: 2 days  
**Owner**: DevOps Team  

**Description**: External services are not secured with SSL/TLS encryption.

**Acceptance Criteria**:
- [ ] Configure SSL certificates for LoadBalancers
- [ ] Implement HTTPS redirection
- [ ] Add certificate management and renewal
- [ ] Update external URLs to use HTTPS
- [ ] Test SSL configuration and security

**Technical Implementation**:
```yaml
# Add SSL annotation to LoadBalancer
metadata:
  annotations:
    service.beta.kubernetes.io/oci-load-balancer-ssl-certificates: "ocid1.certificate.oc1.phx.aaaa..."
```

---

## 📊 P3 - Low Priority Optimizations

### [P3-001] Standardize Image Naming Convention
**Priority**: Low  
**Impact**: Inconsistent image naming creates confusion  
**Estimated Effort**: 1 day  
**Owner**: DevOps Team  

**Description**: Image names and tags are inconsistent across services.

**Acceptance Criteria**:
- [ ] Standardize image naming convention
- [ ] Update all deployment configurations
- [ ] Implement image versioning strategy
- [ ] Add image scanning and vulnerability checks
- [ ] Update CI/CD pipeline for consistent naming

**Technical Implementation**:
```yaml
# Standard image naming
image: ghcr.io/sambitmohanty1/payment-watchdog/api:latest
image: ghcr.io/sambitmohanty1/payment-watchdog/web:latest
image: ghcr.io/sambitmohanty1/payment-watchdog/worker:latest
```

---

### [P3-002] Implement Auto-scaling
**Priority**: Low  
**Impact**: Manual scaling required for traffic changes  
**Estimated Effort**: 2 days  
**Owner**: DevOps Team  

**Description**: Services don't automatically scale based on load.

**Acceptance Criteria**:
- [ ] Configure Horizontal Pod Autoscaler (HPA)
- [ ] Set up metrics collection for auto-scaling
- [ ] Implement load-based scaling policies
- [ ] Test auto-scaling behavior
- [ ] Create scaling documentation

**Technical Implementation**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: payment-watchdog-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: payment-watchdog-api
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

---

### [P3-003] Add Comprehensive Documentation
**Priority**: Low  
**Impact**: Limited operational documentation  
**Estimated Effort**: 2 days  
**Owner**: Technical Writer  

**Description**: Documentation is incomplete for operational procedures.

**Acceptance Criteria**:
- [ ] Create deployment procedures documentation
- [ ] Add troubleshooting guides
- [ ] Document external access configuration
- [ ] Create runbook for common issues
- [ ] Add architecture diagrams

---

## 📈 Success Metrics

### **Security Metrics**
- [ ] Authentication implemented: 100%
- [ ] External access monitoring: 100%
- [ ] Network policies active: 100%
- [ ] SSL/TLS coverage: 100%

### **Operational Metrics**
- [ ] Service uptime: >99.9%
- [ ] Health check success rate: >99%
- [ ] Resource utilization: <80%
- [ ] Backup success rate: 100%

### **Cost Metrics**
- [ ] Monthly cloud costs: <$100 (free tier)
- [ ] Cost monitoring alerts: Active
- [ ] Resource optimization: 20% reduction

---

## 🚀 Implementation Timeline

### **Week 1: Critical Security**
- [P0-001] Authentication & Authorization
- [P0-002] External Access Monitoring
- [P0-003] Cost Controls

### **Week 2: High Priority**
- [P1-001] Port Standardization
- [P1-002] Network Policies
- [P1-003] Health Check Fixes
- [P1-004] Service Discovery

### **Week 3-4: Medium Priority**
- [P2-001] Resource Optimization
- [P2-002] Backup & DR
- [P2-003] SSL/TLS Implementation

### **Month 2: Low Priority**
- [P3-001] Image Naming
- [P3-002] Auto-scaling
- [P3-003] Documentation

---

## 📋 Review Process

### **Monthly Review**
- Assess technical debt reduction progress
- Prioritize new technical debt items
- Review success metrics
- Adjust implementation timeline

### **Quarterly Review**
- Long-term architectural planning
- Major technical debt resolution
- Technology stack evaluation
- Process improvement planning

---

**Last Updated**: 2026-03-23  
**Next Review**: 2026-04-23  
**Owner**: Principal Engineer  
**Review Cadence**: Monthly

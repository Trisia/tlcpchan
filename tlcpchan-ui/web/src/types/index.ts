export interface Instance {
  name: string
  type: 'server' | 'client' | 'http-server' | 'http-client'
  protocol: 'auto' | 'tlcp' | 'tls'
  auth: 'none' | 'one-way' | 'mutual'
  listen: string
  target: string
  enabled: boolean
  status: 'created' | 'running' | 'stopped' | 'error'
  uptime?: number
  config?: InstanceConfig
  stats?: InstanceStats
}

export interface InstanceConfig {
  certificates?: {
    tlcp?: { cert: string; key: string }
    tls?: { cert: string; key: string }
  }
  client_ca?: string[]
  tlcp?: TLCPConfig
  tls?: TLSConfig
}

export interface TLCPConfig {
  min_version?: string
  max_version?: string
  cipher_suites?: string[]
  session_tickets?: boolean
}

export interface TLSConfig {
  min_version?: string
  max_version?: string
  cipher_suites?: string[]
}

export interface InstanceStats {
  connections_total: number
  connections_active: number
  bytes_received: number
  bytes_sent: number
  requests_total?: number
  errors?: number
  latency_avg_ms?: number
}

export interface Certificate {
  name: string
  type: 'tlcp' | 'tls'
  subject: string
  issuer: string
  not_before: string
  not_after: string
  is_ca: boolean
  serial_number: string
  dns_names?: string[]
  ip_addresses?: string[]
  public_key_algorithm: string
  signature_algorithm: string
}

export interface SystemInfo {
  version: string
  go_version: string
  os: string
  arch: string
  uptime: number
  start_time: string
  pid: number
  goroutines: number
  memory: {
    alloc_mb: number
    sys_mb: number
  }
}

export interface HealthInfo {
  status: string
  instances: {
    total: number
    running: number
    stopped: number
  }
  certificates: {
    total: number
    expired: number
    expiring_soon: number
  }
}

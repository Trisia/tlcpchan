export interface Instance {
  name: string
  status: 'created' | 'running' | 'stopped' | 'error'
  config: InstanceConfig
  enabled: boolean
  uptime?: number
}

export interface InstanceConfig {
  name: string
  type: 'server' | 'client' | 'http-server' | 'http-client'
  listen: string
  target: string
  protocol: 'auto' | 'tlcp' | 'tls'
  auth?: 'none' | 'one-way' | 'mutual'
  enabled: boolean
  clientCa?: string[]
  serverCa?: string[]
  tlcp?: TLCPConfig
  tls?: TLSConfig
  http?: HTTPConfig
  sni?: string
  bufferSize?: number
}

export interface KeyStoreConfig {
  name?: string
  type: string
  params: Record<string, string>
}

export interface TLCPConfig {
  auth?: 'none' | 'one-way' | 'mutual'
  minVersion?: string
  maxVersion?: string
  cipherSuites?: string[]
  curvePreferences?: string[]
  sessionTickets?: boolean
  sessionCache?: boolean
  insecureSkipVerify?: boolean
  keystore?: KeyStoreConfig
}

export interface TLSConfig {
  auth?: 'none' | 'one-way' | 'mutual'
  minVersion?: string
  maxVersion?: string
  cipherSuites?: string[]
  curvePreferences?: string[]
  sessionTickets?: boolean
  sessionCache?: boolean
  insecureSkipVerify?: boolean
  keystore?: KeyStoreConfig
}

export interface HTTPConfig {
  requestHeaders?: HeadersConfig
  responseHeaders?: HeadersConfig
}

export interface HeadersConfig {
  add?: Record<string, string>
  remove?: string[]
  set?: Record<string, string>
}

export interface InstanceStats {
  totalConnections: number
  activeConnections: number
  bytesReceived: number
  bytesSent: number
  requestsTotal?: number
  errors?: number
  avgLatencyMs?: number
}

export interface KeyStoreInfo {
  name: string
  type: string
  loaderType: string
  params: Record<string, string>
  protected: boolean
  createdAt: string
  updatedAt: string
}

export interface GenerateKeyStoreRequest {
  name: string
  type: string
  protected: boolean
  certConfig: {
    commonName: string
    country?: string
    stateOrProvince?: string
    locality?: string
    org?: string
    orgUnit?: string
    emailAddress?: string
    years?: number
    days?: number
    keyAlgorithm?: string
    keyBits?: number
    dnsNames?: string[]
    ipAddresses?: string[]
  }
  signerKeyStore?: string
}

export interface RootCertInfo {
  filename: string
  subject: string
  issuer: string
  notAfter: string
}

export interface GenerateRootCARequest {
  commonName: string
  country?: string
  stateOrProvince?: string
  locality?: string
  org?: string
  orgUnit?: string
  emailAddress?: string
  years?: number
  days?: number
}

export interface SystemInfo {
  os: string
  arch: string
  numCpu: number
  numGoroutine: number
  memAllocMb: number
  memTotalMb: number
  memSysMb: number
  startTime: string
  uptime: string
  version?: string
  pid?: number
  memory?: {
    allocMb: number
    sysMb: number
  }
}

export interface HealthStatus {
  status: string
  version: string
  instances?: {
    total: number
    running: number
    stopped: number
  }
  certificates?: {
    total: number
    expired: number
    expiringSoon: number
  }
}

export interface HealthCheckResult {
  protocol: string
  success: boolean
  latencyMs: number
  error?: string
}

export interface InstanceHealthResponse {
  instance: string
  results: HealthCheckResult[]
}

export enum CertType {
  TLCP = 'tlcp',
  TLS = 'tls'
}

export interface VersionInfo {
  version: string
}

export interface Config {
  server: {
    api: { address: string }
    ui: { enabled: boolean; address: string; path: string }
    log?: {
      level: string
      file: string
      maxSize: number
      maxBackups: number
      maxAge: number
      compress: boolean
      enabled: boolean
    }
  }
  keystores: KeyStoreConfig[]
  instances: InstanceConfig[]
}

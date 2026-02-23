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
  tlcp: TLCPConfig
  tls: TLSConfig
  http?: HTTPConfig
  sni?: string
  bufferSize?: number
  stats?: StatsConfig
}

export interface LogConfig {
  level: 'debug' | 'info' | 'warn' | 'error'
  file: string
  maxSize: number
  maxBackups: number
  maxAge: number
  compress: boolean
  enabled: boolean
}

export interface StatsConfig {
  enabled: boolean
}

export interface KeyStoreConfig {
  name?: string
  type: string
  params: Record<string, string>
}

// 前端 UI 使用的 keystore 配置，支持 named 和 file 类型
export interface KeystoreConfigUI {
  type: 'named' | 'file'
  name?: string
  params: Record<string, string>
}

export interface TLCPConfig {
  auth?: 'none' | 'one-way' | 'mutual'
  clientAuthType?: 'no-client-cert' | 'request-client-cert' | 'require-any-client-cert' | 'verify-client-cert-if-given' | 'require-and-verify-client-cert'
  minVersion?: string
  maxVersion?: string
  cipherSuites?: string[]
  sessionTickets?: boolean
  sessionCache?: boolean
  insecureSkipVerify?: boolean
  keystore?: KeyStoreConfig
  keystoreConfig?: KeystoreConfigUI
}

export interface TLSConfig {
  auth?: 'none' | 'one-way' | 'mutual'
  clientAuthType?: 'no-client-cert' | 'request-client-cert' | 'require-any-client-cert' | 'verify-client-cert-if-given' | 'require-and-verify-client-cert'
  minVersion?: string
  maxVersion?: string
  cipherSuites?: string[]
  sessionTickets?: boolean
  sessionCache?: boolean
  insecureSkipVerify?: boolean
  keystore?: KeyStoreConfig
  keystoreConfig?: KeystoreConfigUI
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
  notBefore: string
  notAfter: string
  keyType: string
  serialNumber: string
  version: number
  isCA: boolean
  keyUsage: string[]
}

export interface GenerateRootCARequest {
  type?: string
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

export const CertType = {
  TLCP: 'tlcp',
  TLS: 'tls'
} as const

export type CertType = typeof CertType[keyof typeof CertType]

export interface VersionInfo {
  version: string
}

export interface Config {
  server: {
    api: { address: string }
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
  mcp?: {
    apiKey: string
  }
  keystores: KeyStoreConfig[]
  instances: InstanceConfig[]
}

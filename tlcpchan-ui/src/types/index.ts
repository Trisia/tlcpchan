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

// 前端 UI 使用的 keystore 配置，支持 named（引用已有密钥）、file（证书密钥文件）、ibc-file（IBC 材料文件）三种类型
export interface KeystoreConfigUI {
  type: 'named' | 'file' | 'ibc-file'
  name?: string
  params: Record<string, string>
}

export interface TLCPConfig {
  auth?: 'none' | 'one-way' | 'mutual'
  clientAuthType?: 'no-client-cert' | 'request-client-cert' | 'require-any-client-cert' | 'verify-client-cert-if-given' | 'require-and-verify-client-cert'
  minVersion?: string
  maxVersion?: string
  cipherSuites?: string[]
  sessionCache?: boolean
  insecureSkipVerify?: boolean
  keystore?: KeyStoreConfig
  // IBC(SM9) 身份密钥存储，与 keystore 并列，供 IBC/IBSDH 套件使用
  // 后端 JSON 字段名为 ibcKeystore
  ibcKeystore?: KeyStoreConfig
  keystoreConfig?: KeystoreConfigUI
  // 仅前端使用：IBC 身份 keystore 的表单态，提交前删除
  ibcKeystoreConfig?: KeystoreConfigUI
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

// IBC(SM9) keystore 的只读元信息，字段名与后端 JSON 标签一致
export interface IBCKeyStoreInfo {
  // 本端标识可读形式，如 server@tlcpchan.local
  identity: string
  // 是否提供本端 KGC 公共参数
  hasParams: boolean
  // KGC 属地区域名
  districtName: string
  // 同一区域下的 KGC 序号
  districtSerial: number
  // 公共参数生效时间（ISO8601，零值表示未提供/不限）
  notBefore: string
  // 公共参数失效时间（ISO8601，零值表示未提供/不限）
  notAfter: string
  // 是否提供签名私钥（hid=0x01）
  hasSignKey: boolean
  // 是否提供加密私钥（hid=0x03）
  hasEncryptKey: boolean
  // 是否提供密钥交换私钥（hid=0x02）
  hasKeyExchangeKey: boolean
}

export interface KeyStoreInfo {
  name: string
  type: string
  loaderType: string
  params: Record<string, string>
  protected: boolean
  createdAt: string
  updatedAt: string
  // 仅 type=ibc 时存在
  ibc?: IBCKeyStoreInfo
}

// IBC 信任池中的 KGC 公共参数条目
export interface IBCParamInfo {
  // 文件名
  filename: string
  // KGC 属地区域名
  districtName: string
  // 同一区域下的 KGC 序号
  districtSerial: number
  // 公共参数生效时间（零值表示不限）
  notBefore: string
  // 公共参数失效时间（零值表示不限）
  notAfter: string
  // 公共参数颁发者标识
  issuerIdentity: string
  // 签名主公钥 SM3 指纹（HEX）
  signKeyFingerprint: string
  // 加密主公钥 SM3 指纹（HEX）
  encKeyFingerprint: string
}

// 生成测试 KGC 公共参数请求（仅测试用途）
export interface GenerateIBCParamRequest {
  // KGC 属地区域名
  districtName: string
  // 同一区域下的 KGC 序号
  districtSerial: number
  // 有效期（年）
  years: number
}

export interface GenerateKeyStoreRequest {
  name: string
  type: string
  protected: boolean
  certConfig?: {
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
  // 以下三项仅 type=ibc 时使用
  // 本端标识，如 server@tlcpchan.local
  identity?: string
  // KGC 属地区域名
  districtName?: string
  // 同一区域下的 KGC 序号
  districtSerial?: number
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

export interface KeystoreInstance {
  name: string
  status: 'created' | 'running' | 'stopped' | 'error'
  protocol: 'auto' | 'tlcp' | 'tls'
}

export const CertType = {
  TLCP: 'tlcp',
  TLS: 'tls',
  // 标识密码（IBC/SM9），不使用 X.509 证书
  IBC: 'ibc'
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

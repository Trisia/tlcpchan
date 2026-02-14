/**
 * 代理实例信息
 */
export interface Instance {
  /** 实例名称，全局唯一标识符 */
  name: string
  /** 代理类型
   * - server: TCP服务端代理
   * - client: TCP客户端代理
   * - http-server: HTTP服务端代理
   * - http-client: HTTP客户端代理
   */
  type: 'server' | 'client' | 'http-server' | 'http-client'
  /** 协议类型
   * - auto: 自动检测
   * - tlcp: 国密TLCP协议
   * - tls: 标准TLS协议
   */
  protocol: 'auto' | 'tlcp' | 'tls'
  /** 认证模式
   * - none: 无认证
   * - one-way: 单向认证
   * - mutual: 双向认证
   */
  auth: 'none' | 'one-way' | 'mutual'
  /** 监听地址，格式: "host:port" 或 ":port" */
  listen: string
  /** 目标地址，格式: "host:port" */
  target: string
  /** 是否启用 */
  enabled: boolean
  /** 运行状态
   * - created: 已创建
   * - running: 运行中
   * - stopped: 已停止
   * - error: 错误
   */
  status: 'created' | 'running' | 'stopped' | 'error'
  /** 运行时长，单位: 秒 */
  uptime?: number
  /** 实例详细配置 */
  config?: InstanceConfig
  /** 实例统计信息 */
  stats?: InstanceStats
}

/**
 * 实例详细配置
 */
export interface InstanceConfig {
  /** 证书配置 */
  certificates?: {
    /** TLCP协议证书 */
    tlcp?: { cert: string; key: string }
    /** TLS协议证书 */
    tls?: { cert: string; key: string }
  }
  /** 客户端CA证书路径列表 */
  client_ca?: string[]
  /** TLCP协议配置 */
  tlcp?: TLCPConfig
  /** TLS协议配置 */
  tls?: TLSConfig
}

/**
 * TLCP协议配置（国密）
 */
export interface TLCPConfig {
  /** 最低协议版本，可选值: "1.1" */
  min_version?: string
  /** 最高协议版本，可选值: "1.1" */
  max_version?: string
  /** 密码套件列表
   * - ECC_SM4_CBC_SM3
   * - ECC_SM4_GCM_SM3
   * - ECC_SM4_CCM_SM3
   * - ECDHE_SM4_CBC_SM3
   * - ECDHE_SM4_GCM_SM3
   * - ECDHE_SM4_CCM_SM3
   */
  cipher_suites?: string[]
  /** 是否启用会话票据 */
  session_tickets?: boolean
}

/**
 * TLS协议配置
 */
export interface TLSConfig {
  /** 最低协议版本，可选值: "1.0", "1.1", "1.2", "1.3" */
  min_version?: string
  /** 最高协议版本，可选值: "1.0", "1.1", "1.2", "1.3" */
  max_version?: string
  /** 密码套件列表 */
  cipher_suites?: string[]
}

/**
 * 实例统计信息
 */
export interface InstanceStats {
  /** 总连接数 */
  connections_total: number
  /** 活跃连接数 */
  connections_active: number
  /** 接收字节数，单位: 字节 */
  bytes_received: number
  /** 发送字节数，单位: 字节 */
  bytes_sent: number
  /** 总请求数（HTTP代理） */
  requests_total?: number
  /** 错误数 */
  errors?: number
  /** 平均延迟，单位: 毫秒 */
  latency_avg_ms?: number
}

/**
 * 证书信息
 */
export interface Certificate {
  /** 证书名称 */
  name: string
  /** 证书类型
   * - tlcp: 国密证书
   * - tls: 标准TLS证书
   */
  type: 'tlcp' | 'tls'
  /** 证书主题（DN） */
  subject: string
  /** 颁发者（DN） */
  issuer: string
  /** 生效时间，ISO 8601格式 */
  not_before: string
  /** 过期时间，ISO 8601格式 */
  not_after: string
  /** 是否为CA证书 */
  is_ca: boolean
  /** 序列号 */
  serial_number: string
  /** DNS名称列表（SAN） */
  dns_names?: string[]
  /** IP地址列表（SAN） */
  ip_addresses?: string[]
  /** 公钥算法，如: SM2, RSA, ECDSA */
  public_key_algorithm: string
  /** 签名算法，如: SM3SM2, SHA256-RSA */
  signature_algorithm: string
}

/**
 * 系统信息
 */
export interface SystemInfo {
  /** 版本号 */
  version: string
  /** Go版本 */
  go_version: string
  /** 操作系统 */
  os: string
  /** 架构 */
  arch: string
  /** 运行时长，单位: 秒 */
  uptime: number
  /** 启动时间，ISO 8601格式 */
  start_time: string
  /** 进程ID */
  pid: number
  /** Goroutine数量 */
  goroutines: number
  /** 内存信息 */
  memory: {
    /** 已分配内存，单位: MB */
    alloc_mb: number
    /** 系统内存，单位: MB */
    sys_mb: number
  }
}

/**
 * 健康检查信息
 */
export interface HealthInfo {
  /** 健康状态 */
  status: string
  /** 实例统计 */
  instances: {
    /** 总实例数 */
    total: number
    /** 运行中实例数 */
    running: number
    /** 已停止实例数 */
    stopped: number
  }
  /** 证书统计 */
  certificates: {
    /** 总证书数 */
    total: number
    /** 已过期证书数 */
    expired: number
    /** 即将过期证书数（30天内） */
    expiring_soon: number
  }
}

/**
 * 版本信息
 */
export interface VersionInfo {
  /** 版本号 */
  version: string
  /** Go版本 */
  go_version: string
}

/**
 * UI版本信息
 */
export interface UIVersionInfo {
  /** 版本号 */
  version: string
  /** Go版本 */
  go_version: string
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

export interface VersionInfo {
  version: string
  go_version: string
}

export interface UIVersionInfo {
  version: string
  go_version: string
}

# TLCP Channel 设计文档

## 1. 概述

### 1.1 项目背景

TLCP（Transport Layer Cryptography Protocol，传输层密码协议）是中国国家密码管理局发布的国密SSL协议。本项目旨在提供一个功能完善的TLCP/TLS代理工具，支持双协议并行工作，实现协议转换、流量统计、可视化管理等功能。

### 1.2 设计目标

- 支持TLCP和TLS双协议代理
- 提供服务端代理（TLCP/TLS → TCP）和客户端代理（TCP → TLCP/TLS）
- 支持HTTP高级代理，可配置请求/响应头
- 提供RESTful API接口管理
- 提供Web UI可视化管理界面
- 支持多平台部署（Linux、Windows、macOS）
- 支持多处理器架构（x86_64、arm64、loongarch）
- 开箱即用，首次启动自动生成测试证书

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        TLCP Channel 系统                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ tlcpchan-cli│  │ tlcpchan-ui │  │      tlcpchan (内核)     │ │
│  │  (CLI工具)   │  │  (Web UI)   │  │                         │ │
│  └──────┬──────┘  └──────┬──────┘  │  ┌───────────────────┐  │ │
│         │                │         │  │    Controller     │  │ │
│         │                │         │  │   (RESTful API)   │  │ │
│         │                │         │  └─────────┬─────────┘  │ │
│         │                │         │            │            │ │
│         │                │         │  ┌─────────▼─────────┐  │ │
│         │                │         │  │ Instance Manager  │  │ │
│         │                │         │  └─────────┬─────────┘  │ │
│         │                │         │            │            │ │
│         │                │         │  ┌─────────▼─────────┐  │ │
│         │                │         │  │    Proxy Engine   │  │ │
│         │                │         │  │ ┌───────┐┌───────┐│  │ │
│         │                │         │  │ │Server ││Client ││  │ │
│         │                │         │  │ │Proxy  ││Proxy  ││  │ │
│         │                │         │  │ └───────┘└───────┘│  │ │
│         │                │         │  │ ┌─────────────────┐│  │ │
│         │                │         │  │ │  HTTP Proxy     ││  │ │
│         │                │         │  │ └─────────────────┘│  │ │
│         │                │         │  └───────────────────┘  │ │
│         │                │         │            │            │ │
│         │                │         │  ┌─────────▼─────────┐  │ │
│         │                │         │  │   Cert Manager    │  │ │
│         │                │         │  │   Key Manager     │  │ │
│         │                │         │  │   Stats Module    │  │ │
│         │                │         │  │   Logger Module   │  │ │
│         │                │         │  │   Config Module   │  │ │
│         │                │         │  └───────────────────┘  │ │
└─────────┴────────────────┴─────────┴─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
         │                │                      │
         └────────────────┴──────────────────────┘
                    HTTP RESTful API
```

### 2.2 项目目录结构

```
tlcpchan/                          # 内核主程序
├── main.go                        # 程序入口
├── config/                        # 配置管理
│   ├── config.go                  # 配置结构定义、加载/保存逻辑
│   ├── config_test.go             # 配置测试
│   └── config.yaml                # 默认配置文件
├── logger/                        # 日志管理
│   └── logger.go                  # 日志管理器
├── cert/                          # 证书管理
│   ├── loader.go                  # 证书加载器
│   ├── loader_test.go             # 加载器测试
│   ├── manager.go                 # 证书管理器
│   ├── generator.go               # 证书生成器
│   └── embedded.go                # 嵌入证书资源
├── key/                           # 密钥存储管理
│   ├── key.go                     # 密钥结构
│   ├── manager.go                 # 密钥管理器
│   ├── generator.go               # 密钥生成器
│   ├── model.go                   # 数据模型
│   ├── store.go                   # 存储实现
│   └── validator.go               # 验证器
├── proxy/                         # 代理核心
│   ├── server.go                  # 服务端代理(TLCP/TLS→TCP)
│   ├── client.go                  # 客户端代理(TCP→TLCP/TLS)
│   ├── http_server.go             # HTTP服务端代理
│   ├── http_client.go             # HTTP客户端代理
│   ├── adapter.go                 # 协议适配器
│   ├── adapter_test.go            # 适配器测试
│   ├── conn.go                    # 连接处理
│   └── vars.go                    # 变量定义
├── instance/                      # 实例管理
│   ├── manager.go                 # 实例管理器
│   ├── instance.go                # 实例实现
│   └── types.go                   # 类型定义
├── controller/                    # API控制器
│   ├── server.go                  # API服务器
│   ├── router.go                  # 路由定义
│   ├── instance.go                # 实例管理API
│   ├── config.go                  # 配置管理API
│   ├── cert.go                    # 证书管理API
│   ├── keystore.go                # 密钥存储API
│   ├── health.go                  # 健康检查API
│   ├── system.go                  # 系统信息API
│   ├── middleware.go              # 中间件
│   └── response.go                # 响应处理
├── stats/                         # 流量统计
│   └── collector.go               # 统计收集器
├── release/                       # 打包相关
│   └── systemd/                   # systemd服务文件
├── docs/
│   └── config-examples.md         # 配置示例
├── bin/                           # 编译输出目录
└── go.mod

tlcpchan-cli/                      # CLI工具
├── main.go                        # 程序入口
├── commands/                      # 命令实现
│   ├── root.go                    # 根命令
│   ├── instance.go                # 实例命令
│   ├── config.go                  # 配置命令
│   ├── cert.go                    # 证书命令
│   ├── keystore.go                # 密钥存储命令
│   ├── system.go                  # 系统命令
│   └── version.go                 # 版本命令
├── client/                        # API客户端
│   └── client.go                  # HTTP客户端封装
├── bin/                           # 编译输出目录
└── go.mod

tlcpchan-ui/                       # Web前端界面
├── main.go                        # UI服务主入口
├── server/                        # UI服务器
├── proxy/                         # 代理实现
├── web/                           # 前端项目
│   ├── src/                       # 前端源代码
│   │   ├── api/                   # API调用封装
│   │   ├── assets/                # 静态资源
│   │   ├── components/            # 通用组件
│   │   ├── layouts/               # 页面布局
│   │   ├── router/                # 路由配置
│   │   ├── stores/                # 状态管理(Pinia)
│   │   ├── types/                 # TypeScript类型定义
│   │   ├── views/                 # 页面视图
│   │   │   ├── Dashboard.vue      # 仪表板
│   │   │   ├── Instances.vue      # 实例列表
│   │   │   ├── InstanceDetail.vue # 实例详情
│   │   │   ├── Certificates.vue   # 证书管理
│   │   │   ├── KeyStores.vue      # 密钥存储
│   │   │   ├── Logs.vue           # 日志查看
│   │   │   └── Settings.vue       # 系统设置
│   │   ├── App.vue                 # 根组件
│   │   ├── main.ts                 # 入口文件
│   │   └── style.css               # 全局样式
│   ├── public/                     # 公共资源
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── dist/                          # 前端构建产物
│   ├── assets/
│   ├── index.html
│   └── version.txt
├── bin/                           # 编译输出目录
└── go.mod
```

## 3. 核心模块设计

### 3.1 配置管理模块

#### 3.1.1 配置结构

```go
type Config struct {
    Server    ServerConfig    `yaml:"server"`
    Instances []InstanceConfig `yaml:"instances"`
}

type ServerConfig struct {
    API  APIConfig  `yaml:"api"`
    UI   UIConfig   `yaml:"ui"`
    Log  LogConfig  `yaml:"log"`
}

type APIConfig struct {
    Address string `yaml:"address"` // API服务地址，默认 :30080
}

type UIConfig struct {
    Enabled bool   `yaml:"enabled"` // 是否启动UI服务
    Address string `yaml:"address"` // UI服务地址，默认 :30000
    Path    string `yaml:"path"`    // UI静态文件路径
}

type LogConfig struct {
    Level      string `yaml:"level"`       // 日志级别：debug/info/warn/error
    File       string `yaml:"file"`        // 日志文件路径
    MaxSize    int    `yaml:"max_size"`    // 单文件最大大小(MB)
    MaxBackups int    `yaml:"max_backups"` // 最大备份数
    MaxAge     int    `yaml:"max_age"`     // 最大保留天数
    Compress   bool   `yaml:"compress"`    // 是否压缩
}

type InstanceConfig struct {
    Name     string            `yaml:"name"`     // 实例名称
    Type     string            `yaml:"type"`     // 类型：server/client/http-server/http-client
    Listen   string            `yaml:"listen"`   // 监听地址
    Target   string            `yaml:"target"`   // 目标地址
    Protocol string            `yaml:"protocol"` // 协议：auto/tlcp/tls
    Auth     string            `yaml:"auth"`     // 认证：none/one-way/mutual
    Enabled  bool              `yaml:"enabled"`  // 是否启用
    
    // TLS/TLCP 配置
    TLCP     TLCPConfig        `yaml:"tlcp"`
    TLS      TLSConfig         `yaml:"tls"`
    
    // 证书配置
    Certificates CertificatesConfig `yaml:"certificates"`
    ClientCA     []string           `yaml:"client_ca"`     // 客户端CA证书（服务端验证）
    ServerCA     []string           `yaml:"server_ca"`     // 服务端CA证书（客户端验证）
    
    // HTTP代理配置
    HTTP      HTTPConfig        `yaml:"http"`
    
    // 日志配置
    Log       *LogConfig        `yaml:"log"`
    
    // 统计配置
    Stats     *StatsConfig      `yaml:"stats"`
}

type TLCPConfig struct {
    MinVersion       uint16   `yaml:"min_version"`       // 最低协议版本
    MaxVersion       uint16   `yaml:"max_version"`       // 最高协议版本
    CipherSuites     []uint16 `yaml:"cipher_suites"`     // 密码套件
    CurvePreferences []uint16 `yaml:"curve_preferences"` // 曲线偏好
    SessionTickets   bool     `yaml:"session_tickets"`   // 会话票据
    SessionCache     bool     `yaml:"session_cache"`     // 会话缓存
    InsecureSkipVerify bool   `yaml:"insecure_skip_verify"` // 跳过验证（客户端）
}

type TLSConfig struct {
    MinVersion       uint16   `yaml:"min_version"`
    MaxVersion       uint16   `yaml:"max_version"`
    CipherSuites     []uint16 `yaml:"cipher_suites"`
    CurvePreferences []uint16 `yaml:"curve_preferences"`
    SessionTickets   bool     `yaml:"session_tickets"`
    SessionCache     bool     `yaml:"session_cache"`
    InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type CertificatesConfig struct {
    TLCP CertConfig `yaml:"tlcp"` // TLCP证书配置
    TLS  CertConfig `yaml:"tls"`  // TLS证书配置
}

type CertConfig struct {
    Cert string `yaml:"cert"` // 证书文件路径
    Key  string `yaml:"key"`  // 私钥文件路径
}

type HTTPConfig struct {
    RequestHeaders  HeadersConfig `yaml:"request_headers"`
    ResponseHeaders HeadersConfig `yaml:"response_headers"`
}

type HeadersConfig struct {
    Add    map[string]string `yaml:"add"`    // 添加头部
    Remove []string          `yaml:"remove"` // 移除头部
    Set    map[string]string `yaml:"set"`    // 设置头部
}

type StatsConfig struct {
    Enabled  bool          `yaml:"enabled"`  // 是否启用统计
    Interval time.Duration `yaml:"interval"` // 统计间隔
}
```

### 3.2 代理模块

#### 3.2.1 服务端代理（TLCP/TLS → TCP）

服务端代理接收TLCP/TLS加密流量，解密后转发到目标TCP服务。

```
客户端 ──[TLCP/TLS]──> 代理服务端 ──[TCP]──> 目标服务
```

**协议自动适配流程：**

```
┌─────────────┐     ┌─────────────────────────────────┐     ┌─────────────┐
│   客户端     │     │           代理服务端              │     │   目标服务   │
│             │     │                                 │     │             │
│   发送      │     │  1. 接收ClientHello             │     │             │
│ ClientHello │────>│                                 │     │             │
│             │     │  2. 解析协议类型                 │     │             │
│             │     │     - TLCP: 特定扩展/密码套件    │     │             │
│             │     │     - TLS: 标准ALPN/SNI         │     │             │
│             │     │                                 │     │             │
│             │     │  3. 选择对应证书和配置           │     │             │
│             │     │                                 │     │             │
│             │     │  4. 完成握手                     │     │             │
│             │     │                                 │     │             │
│   加密数据   │────>│  5. 解密并转发 ─────────────────>│────>│   明文数据   │
│             │     │                                 │     │             │
│             │<────│  6. 接收响应并加密 <─────────────│<────│   响应数据   │
│             │     │                                 │     │             │
└─────────────┘     └─────────────────────────────────┘     └─────────────┘
```

**协议检测方法：**
- TLCP：检查ClientHello中的国密密码套件或TLCP特定扩展
- TLS：标准TLS握手协议

#### 3.2.2 客户端代理（TCP → TLCP/TLS）

客户端代理接收明文TCP流量，加密后转发到目标TLCP/TLS服务。

```
客户端 ──[TCP]──> 代理客户端 ──[TLCP/TLS]──> 目标服务
```

**自动协议检测（客户端模式）：**

当protocol设置为auto时，代理会：
1. 首次连接目标服务时，发送协议探测请求
2. 根据目标服务响应判断协议类型
3. 缓存协议类型，后续连接复用
4. 创建对应的http.Client实例

#### 3.2.3 HTTP代理

HTTP代理在TCP代理基础上增加HTTP协议解析和头部处理能力。

```
HTTP客户端 ──[HTTP/HTTPS]──> HTTP代理 ──[HTTP/HTTPS]──> 目标服务
```

**请求处理流程：**
1. 接收HTTP请求
2. 根据配置添加/删除/修改请求头
3. 转发到目标服务
4. 接收响应
5. 根据配置修改响应头
6. 返回给客户端

### 3.3 证书管理模块

#### 3.3.1 证书加载器接口

```go
type CertLoader interface {
    Load(certFile, keyFile string) (*CertBundle, error)
    Watch(certFile, keyFile string, onChange func(*CertBundle)) error
}

type CertBundle struct {
    TLCP *tlcp.Certificate  // TLCP证书
    TLS  *tls.Certificate   // TLS证书
}
```

#### 3.3.2 热更新机制

使用`GetCertificate`回调函数实现证书热更新：

```go
func (l *TLCPListener) GetCertificate(hello *tlcp.ClientHelloInfo) (*tlcp.Certificate, error) {
    // 动态加载最新证书
    return l.certLoader.LoadCertificate()
}
```

#### 3.3.3 证书生成器

首次启动时自动生成：
- 根CA证书（SM2/RSA）
- 服务端证书（SM2/RSA）
- 客户端证书（SM2/RSA）

### 3.4 密钥存储管理模块

密钥存储模块提供密钥和证书的集中管理功能，支持 TLCP（国密双证书）和 TLS 两种类型。

#### 3.4.1 密钥存储类型

```go
type KeyStoreType string

const (
    KeyStoreTypeTLCP KeyStoreType = "tlcp"
    KeyStoreTypeTLS  KeyStoreType = "tls"
)
```

**TLCP 密钥存储**：包含签名证书/密钥和加密证书/密钥（国密双证书体系）
**TLS 密钥存储**：包含单一证书/密钥

#### 3.4.2 密钥管理器

```go
type Manager struct {
    // 密钥存储管理
}

func (m *Manager) List() ([]*KeyStoreInfo, error)
func (m *Manager) GetInfo(name string) (*KeyStoreInfo, error)
func (m *Manager) Create(name string, keyType KeyStoreType, params KeyParams, 
    signCert, signKey, encCert, encKey []byte) (*KeyStore, error)
func (m *Manager) UpdateCertificates(name string, signCert, encCert []byte) (*KeyStore, error)
func (m *Manager) Delete(name string) error
func (m *Manager) Reload(name string) error
func (m *Manager) Exists(name string) bool
```

#### 3.4.3 密钥存储信息

```go
type KeyStoreInfo struct {
    Name        string                 // 密钥名称
    Type        KeyStoreType           // 类型（tlcp/tls）
    KeyParams   KeyParams              // 密钥参数
    HasSignCert bool                   // 是否有签名证书
    HasSignKey  bool                   // 是否有签名密钥
    HasEncCert  bool                   // 是否有加密证书（仅国密）
    HasEncKey   bool                   // 是否有加密密钥（仅国密）
    CreatedAt   time.Time              // 创建时间
    UpdatedAt   time.Time              // 更新时间
}

type KeyParams struct {
    Algorithm string                   // 算法（SM2/RSA/ECDSA）
    Length    int                      // 密钥长度
    Type      string                   // 类型
}
```

### 3.5 实例管理模块

```go
type Instance interface {
    Name() string
    Type() InstanceType
    Start() error
    Stop() error
    Reload(config *InstanceConfig) error
    Status() InstanceStatus
    Stats() *InstanceStats
}

type InstanceManager struct {
    instances map[string]Instance
    mu        sync.RWMutex
}

func (m *InstanceManager) Create(config *InstanceConfig) (Instance, error)
func (m *InstanceManager) Get(name string) (Instance, bool)
func (m *InstanceManager) List() []Instance
func (m *InstanceManager) Delete(name string) error
```

### 3.6 统计模块

```go
type Metrics struct {
    ConnectionsTotal   int64         // 总连接数
    ConnectionsActive  int64         // 活跃连接数
    BytesReceived      int64         // 接收字节数
    BytesSent          int64         // 发送字节数
    RequestsTotal      int64         // 总请求数（HTTP）
    Errors             int64         // 错误数
    LatencyAvg         time.Duration // 平均延迟
    LastUpdateTime     time.Time     // 最后更新时间
}
```

## 4. API设计

### 4.1 API服务器

使用Go标准库`net/http`实现RESTful API，无需第三方框架。

### 4.2 路由设计

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/instances | 获取实例列表 |
| POST | /api/v1/instances | 创建实例 |
| GET | /api/v1/instances/:name | 获取实例详情 |
| PUT | /api/v1/instances/:name | 更新实例配置 |
| DELETE | /api/v1/instances/:name | 删除实例 |
| POST | /api/v1/instances/:name/start | 启动实例 |
| POST | /api/v1/instances/:name/stop | 停止实例 |
| POST | /api/v1/instances/:name/reload | 重载实例 |
| GET | /api/v1/instances/:name/stats | 获取统计信息 |
| GET | /api/v1/instances/:name/logs | 获取日志 |
| POST | /api/v1/config/reload | 重载全局配置 |
| GET | /api/v1/config | 获取当前配置 |
| GET | /api/v1/certificates | 获取证书列表 |
| POST | /api/v1/certificates/reload | 热更新证书 |
| GET | /api/v1/keystores | 获取密钥列表 |
| GET | /api/v1/keystores/:name | 获取密钥详情 |
| POST | /api/v1/keystores | 创建密钥 |
| POST | /api/v1/keystores/:name/certificates | 更新密钥证书 |
| DELETE | /api/v1/keystores/:name | 删除密钥 |
| POST | /api/v1/keystores/:name/reload | 重载密钥 |
| GET | /api/v1/system/info | 系统信息 |
| GET | /api/v1/system/health | 健康检查 |

## 5. UI设计

### 5.1 技术栈

- Vue 3 + TypeScript
- Vite 构建工具
- Element Plus UI组件库
- Pinia 状态管理
- Vue Router 路由

### 5.2 页面设计

| 页面 | 路由 | 描述 |
|------|------|------|
| 仪表盘 | / | 系统概览、流量统计图表 |
| 实例管理 | /instances | 实例列表、创建、编辑 |
| 实例详情 | /instances/:name | 单个实例详情和监控 |
| 证书管理 | /certificates | 证书列表、上传、生成 |
| 日志查看 | /logs | 日志实时查看 |
| 系统设置 | /settings | 系统配置 |

### 5.3 UI服务架构

```
tlcpchan-ui/
├── main.go                        # UI服务主入口
├── server/                        # UI服务器实现
├── proxy/                         # 代理实现
├── web/                           # Vue前端项目
│   ├── src/                       # 前端源代码
│   ├── public/                    # 公共资源
│   └── package.json
├── dist/                          # 前端构建产物（嵌入Go二进制）
└── bin/                           # 编译输出目录
```

UI服务作为独立进程运行，提供：
1. 静态资源服务 - 托管 Vue 前端
2. API 代理 - 转发请求到 tlcpchan 核心服务
3. 独立部署 - 可与核心服务分离部署

前端技术栈：
- Vue 3 + TypeScript
- Vite 构建工具
- Element Plus UI组件库
- Pinia 状态管理
- Vue Router 路由

## 6. 部署设计

### 6.1 systemd服务

```ini
[Unit]
Description=TLCP Channel - TLCP/TLS Proxy Server
After=network.target

[Service]
Type=simple
User=tlcpchan
Group=tlcpchan
WorkingDirectory=/opt/tlcpchan
ExecStart=/usr/bin/tlcpchan
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### 6.2 多平台打包

| 平台 | 架构 | 格式 |
|------|------|------|
| Linux | x86_64 | .deb, .rpm, .tar.gz |
| Linux | arm64 | .deb, .rpm, .tar.gz |
| Linux | loongarch | .deb, .rpm, .tar.gz |
| Windows | x86_64 | .msi, .zip |
| Windows | arm64 | .msi, .zip |
| macOS | x86_64 | .pkg, .tar.gz |
| macOS | arm64 | .pkg, .tar.gz |

### 6.3 Docker支持

```dockerfile
FROM alpine:3.18
COPY tlcpchan /usr/bin/tlcpchan
COPY tlcpchan-ui/ui-server /usr/bin/tlcpchan-ui
COPY config /etc/tlcpchan
EXPOSE 30080 30000
ENTRYPOINT ["/usr/bin/tlcpchan"]
```

## 10. 热加载机制设计

### 10.1 概述

热加载机制允许在不重启服务的情况下更新证书和配置，确保服务的高可用性。系统支持多层次的热加载：

- **证书热重载**：更新端证书和 CA 证书池
- **配置热重载**：更新代理实例配置
- **API 触发**：通过 RESTful API 手动触发重载

### 10.2 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                        热加载架构                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐    ┌─────────────────┐    ┌───────────────┐ │
│  │   API 触发      │    │  定时器自动     │    │  手动调用      │ │
│  │  /reload-certs  │───▶│  HotReloader    │◀───│  Reload()      │ │
│  │  /reload        │    │                 │    │               │ │
│  └────────┬────────┘    └────────┬────────┘    └───────┬───────┘ │
│           │                        │                     │          │
│           └────────────────────────┼─────────────────────┘          │
│                                    │                                │
│  ┌─────────────────────────────────▼──────────────────────────────┐  │
│  │                      TLCPAdapter                               │  │
│  │  ┌──────────────────────────────────────────────────────────┐  │  │
│  │  │ ReloadCertificates() - 仅重载证书                        │  │  │
│  │  │  - tlcpCertRef.ReloadFromPath()                         │  │  │
│  │  │  - tlsCertRef.ReloadFromPath()                          │  │  │
│  │  │  - clientCAPool.Reload()                                 │  │  │
│  │  │  - serverCAPool.Reload()                                 │  │  │
│  │  └──────────────────────────────────────────────────────────┘  │  │
│  │  ┌──────────────────────────────────────────────────────────┐  │  │
│  │  │ ReloadConfig() - 完全重载配置                            │  │  │
│  │  │  - 重新加载证书（如路径变更）                             │  │  │
│  │  │  - 重建 CA 证书池                                         │  │  │
│  │  │  - 更新协议配置                                           │  │  │
│  │  └──────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                    │                                │
│  ┌─────────────────────────────────┼───────────────────────────────┐  │
│  │                                 │                               │  │
│  ▼                                 ▼                               ▼  │
│  ┌──────────────┐        ┌─────────────────┐        ┌───────────┐  │
│  │ Certificate  │        │   HotCertPool   │        │  Instance │  │
│  │  - Reload()  │        │   - Reload()    │        │ - Reload()│  │
│  │  - ReloadF.. │        │   - Pool()      │        │           │  │
│  └──────────────┘        └─────────────────┘        └───────────┘  │
└───────────────────────────────────────────────────────────────────────┘
```

### 10.3 核心组件

#### 10.3.1 Certificate 热重载

`Certificate` 结构体支持两种重载方式：

```go
type Certificate struct {
    Certificate []*x509.Certificate
    PrivateKey  crypto.PrivateKey
    certPEM     []byte      // 内存中的 PEM 数据
    keyPEM      []byte
    certPath    string      // 文件路径
    keyPath     string
    mu          sync.RWMutex
}

// Reload 从内存 PEM 数据重载
func (c *Certificate) Reload() error

// ReloadFromPath 从文件路径重载
func (c *Certificate) ReloadFromPath() error
```

**关键特性**：
- 使用读写锁保证并发安全
- 原子更新证书和私钥
- 新连接使用新证书，已有连接不受影响

#### 10.3.2 HotCertPool CA 证书池热重载

```go
type HotCertPool struct {
    mu         sync.RWMutex
    paths      []string
    pool       *x509.CertPool
    smPool     *smx509.CertPool
    lastReload time.Time
}

func (h *HotCertPool) Load() error
func (h *HotCertPool) Reload() error
func (h *HotCertPool) Pool() *x509.CertPool      // 读操作，加读锁
func (h *HotCertPool) SMPool() *smx509.CertPool  // 读操作，加读锁
```

**设计要点**：
- 读写分离：读操作加读锁，写操作加写锁
- 支持国密（SM）和标准 x509 双证书池
- 记录上次重载时间

#### 10.3.3 TLCPAdapter 协议适配器热重载

适配器是热加载的核心协调者：

```go
type TLCPAdapter struct {
    mu           sync.RWMutex
    tlcpConfig   *tlcp.Config
    tlsConfig    *tls.Config
    clientCAPool *cert.HotCertPool
    serverCAPool *cert.HotCertPool
    tlcpCertRef  *cert.Certificate  // 证书引用
    tlsCertRef   *cert.Certificate
    // ...
}

// ReloadCertificates 仅重载证书（轻量级）
func (a *TLCPAdapter) ReloadCertificates() error {
    // 重载端证书
    a.tlcpCertRef.ReloadFromPath()
    a.tlsCertRef.ReloadFromPath()
    // 重载 CA 池
    clientCAPool.Reload()
    serverCAPool.Reload()
}

// ReloadConfig 完全重载配置（重量级）
func (a *TLCPAdapter) ReloadConfig(cfg *config.InstanceConfig) error
```

**关键设计**：
- 区分轻量级（仅证书）和重量级（完整配置）重载
- 使用 `GetCertificate` 回调实现证书动态获取
- 新配置原子替换，避免服务中断

#### 10.3.4 HotReloader 定时器自动重载

```go
type HotReloader struct {
    loader    *Loader
    interval  time.Duration
    stopCh    chan struct{}
    stoppedCh chan struct{}
}

func NewHotReloader(loader *Loader, interval time.Duration) *HotReloader
func (h *HotReloader) Start()  // 启动定时重载
func (h *HotReloader) Stop()   // 停止定时重载
```

**工作流程**：
1. 启动后台 goroutine
2. 按设定间隔周期性调用 `loader.ReloadAll()`
3. 支持优雅停止

### 10.4 热加载流程

#### 10.4.1 证书热重载流程

```
用户/API 触发
      │
      ▼
┌─────────────────┐
│ ReloadCertificates() │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────┐
│ TLCP    │ │ TLS     │
│ 证书    │ │ 证书    │
│重载     │ │重载     │
└────┬────┘ └────┬────┘
     │           │
     └─────┬─────┘
           │
           ▼
    ┌──────────────┐
    │  CA 证书池    │
    │   重载       │
    └──────┬───────┘
           │
           ▼
      完成！新连接
      自动使用新证书
```

#### 10.4.2 完整配置热重载流程

```
用户/API 触发
      │
      ▼
┌─────────────────┐
│  ReloadConfig() │
└────────┬────────┘
         │
    ┌────▼────┐
    │ 检查配置  │
    │ 变更     │
    └────┬────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────┐
│证书路径 │ │ 其他配置 │
│变更？   │ │ 更新     │
└────┬────┘ └────┬────┘
     │           │
     ▼           │
┌─────────┐      │
│重新加载 │      │
│证书     │      │
└────┬────┘      │
     │           │
     └─────┬─────┘
           │
           ▼
    ┌──────────────┐
    │  重建配置    │
    │  原子替换    │
    └──────┬───────┘
           │
           ▼
        完成！
```

### 10.5 API 接口

系统提供以下热加载 API：

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/instances/:name/reload` | 重载实例完整配置 |
| POST | `/api/v1/instances/:name/reload-certs` | 仅重载实例证书 |
| POST | `/api/v1/certificates/reload` | 重载所有证书 |
| POST | `/api/v1/keystores/:name/reload` | 重载指定密钥 |
| POST | `/api/v1/config/reload` | 重载全局配置 |

### 10.6 并发安全设计

热加载机制在设计上充分考虑了并发安全：

| 组件 | 并发机制 |
|------|----------|
| `Certificate` | `sync.RWMutex` 保护证书和私钥 |
| `HotCertPool` | `sync.RWMutex` 保护证书池 |
| `Manager.certs` | `sync.Map` 提升并发读性能 |
| `Manager.certDir` | `atomic.Value` 无锁访问 |
| `TLCPAdapter` | `sync.RWMutex` 保护配置更新 |

### 10.7 最佳实践

1. **证书更新**：优先使用 `ReloadCertificates()`，轻量高效
2. **配置变更**：仅在必要时使用 `ReloadConfig()`
3. **定时重载**：合理设置 `HotReloader` 间隔，避免频繁 I/O
4. **监控**：关注重载失败日志，及时处理证书问题

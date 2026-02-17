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
│         │                │         │  │  Security Module  │  │ │
│         │                │         │  │ ┌───────────────┐ │  │ │
│         │                │         │  │ │KeyStore Manager│ │  │ │
│         │                │         │  │ └───────────────┘ │  │ │
│         │                │         │  │ ┌───────────────┐ │  │ │
│         │                │         │  │ │RootCert Manager│ │  │ │
│         │                │         │  │ └───────────────┘ │  │ │
│         │                │         │  └───────────────────┘  │ │
│         │                │         │  ┌───────────────────┐  │ │
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
    Enabled  bool              `yaml:"enabled"`  // 是否启用
    
    // TLS/TLCP 配置
    TLCP     TLCPConfig        `yaml:"tlcp"`
    TLS      TLSConfig         `yaml:"tls"`
    
    // 安全配置（详见 security.md）
    ClientCA []string          `yaml:"client-ca,omitempty"` // 客户端CA证书（服务端验证）
    ServerCA []string          `yaml:"server-ca,omitempty"` // 服务端CA证书（客户端验证）
    
    // HTTP代理配置
    HTTP      *HTTPConfig      `yaml:"http,omitempty"`
    
    // 日志配置
    Log       *LogConfig        `yaml:"log,omitempty"`
    
    // 统计配置
    Stats     *StatsConfig      `yaml:"stats,omitempty"`
}

type TLCPConfig struct {
    Auth             string              `yaml:"auth,omitempty"`              // 认证模式：none/one-way/mutual
    MinVersion       string              `yaml:"min-version,omitempty"`       // 最低协议版本
    MaxVersion       string              `yaml:"max-version,omitempty"`       // 最高协议版本
    CipherSuites     []string            `yaml:"cipher-suites,omitempty"`     // 密码套件
    CurvePreferences []string            `yaml:"curve-preferences,omitempty"` // 曲线偏好
    SessionTickets   bool                `yaml:"session-tickets,omitempty"`   // 会话票据
    SessionCache     bool                `yaml:"session-cache,omitempty"`     // 会话缓存
    InsecureSkipVerify bool              `yaml:"insecure-skip-verify,omitempty"` // 跳过验证（客户端）
    Keystore         *config.KeyStoreConfig `yaml:"keystore,omitempty"`     // Keystore 配置（详见 security.md）
}

type TLSConfig struct {
    Auth             string              `yaml:"auth,omitempty"`              // 认证模式：none/one-way/mutual
    MinVersion       string              `yaml:"min-version,omitempty"`       // 最低协议版本
    MaxVersion       string              `yaml:"max-version,omitempty"`       // 最高协议版本
    CipherSuites     []string            `yaml:"cipher-suites,omitempty"`     // 密码套件
    CurvePreferences []string            `yaml:"curve-preferences,omitempty"` // 曲线偏好
    SessionTickets   bool                `yaml:"session-tickets,omitempty"`   // 会话票据
    SessionCache     bool                `yaml:"session-cache,omitempty"`     // 会话缓存
    InsecureSkipVerify bool              `yaml:"insecure-skip-verify,omitempty"` // 跳过验证（客户端）
    Keystore         *config.KeyStoreConfig `yaml:"keystore,omitempty"`     // Keystore 配置（详见 security.md）
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

### 3.3 安全参数管理模块

安全参数（Keystore、根证书）的详细配置和管理方法请参考 [security.md](./security.md)。

安全参数管理模块统一管理 Keystore 和根证书，提供抽象的接口设计，支持多种加载方式。

安全参数管理模块统一管理 Keystore 和根证书，提供抽象的接口设计，支持多种加载方式。

#### 3.3.1 核心概念

| 概念 | 说明 |
|------|------|
| **Keystore** | 密钥存储，包含签名/加密证书和密钥 |
| **RootCert** | 根证书，用于验证对端证书 |
| **Loader** | Keystore 加载器，支持多种加载方式 |

> **详细文档**：安全参数的完整配置和管理方法请参考 [security.md](./security.md)。

### 3.4 实例管理模块

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

### 3.5 统计模块

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
| GET | /api/instances | 获取实例列表 |
| POST | /api/instances | 创建实例 |
| GET | /api/instances/:name | 获取实例详情 |
| PUT | /api/instances/:name | 更新实例配置 |
| DELETE | /api/instances/:name | 删除实例 |
| POST | /api/instances/:name/start | 启动实例 |
| POST | /api/instances/:name/stop | 停止实例 |
| POST | /api/instances/:name/reload | 重载实例 |
| GET | /api/instances/:name/stats | 获取统计信息 |
| GET | /api/instances/:name/logs | 获取日志 |
| POST | /api/config/reload | 重载全局配置 |
| GET | /api/config | 获取当前配置 |
| GET | /api/security/keystores | 获取 keystore 列表 |
| POST | /api/security/keystores | 创建 keystore |
| GET | /api/security/keystores/:name | 获取 keystore 详情 |
| DELETE | /api/security/keystores/:name | 删除 keystore |
| POST | /api/security/keystores/:name/reload | 重载 keystore |
| GET | /api/security/rootcerts | 获取根证书列表 |
| POST | /api/security/rootcerts | 添加根证书 |
| GET | /api/security/rootcerts/:name | 获取根证书详情 |
| DELETE | /api/security/rootcerts/:name | 删除根证书 |
| POST | /api/security/rootcerts/reload | 重载所有根证书 |
| GET | /api/system/info | 系统信息 |
| GET | /api/system/health | 健康检查 |

> **安全参数 API 详细文档**：请参考 [security.md](./security.md)

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
| 安全参数 | /security | Keystore 和根证书管理 |
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


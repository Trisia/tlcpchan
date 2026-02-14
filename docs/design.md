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
│         │                │         │  │   Cert Loader     │  │ │
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
├── internal/
│   ├── config/                    # 配置管理
│   │   ├── config.go              # 配置结构定义
│   │   ├── loader.go              # YAML配置加载器
│   │   └── validator.go           # 配置验证器
│   ├── logger/                    # 日志管理
│   │   ├── logger.go              # 日志管理器
│   │   └── writer.go              # 多输出Writer
│   ├── cert/                      # 证书管理
│   │   ├── loader.go              # 证书加载器接口
│   │   ├── pem_loader.go          # PEM格式加载器
│   │   ├── tlcp.go                # TLCP证书处理
│   │   ├── tls.go                 # TLS证书处理
│   │   └── generator.go           # 证书生成器
│   ├── proxy/                     # 代理核心
│   │   ├── server.go              # 服务端代理(TLCP/TLS→TCP)
│   │   ├── client.go              # 客户端代理(TCP→TLCP/TLS)
│   │   ├── http_server.go         # HTTP服务端代理
│   │   ├── http_client.go         # HTTP客户端代理
│   │   ├── adapter.go             # 协议适配器
│   │   └── conn.go                # 连接处理
│   ├── instance/                  # 实例管理
│   │   ├── manager.go             # 实例管理器
│   │   └── instance.go            # 实例定义
│   ├── controller/                # API控制器
│   │   ├── server.go              # API服务器
│   │   ├── router.go              # 路由定义
│   │   ├── instance.go            # 实例管理API
│   │   ├── config.go              # 配置管理API
│   │   ├── cert.go                # 证书管理API
│   │   ├── stats.go               # 统计查询API
│   │   └── system.go              # 系统信息API
│   └── stats/                     # 流量统计
│       ├── collector.go           # 统计收集器
│       └── metrics.go             # 指标定义
├── release/                       # 打包相关
│   ├── systemd/                   # systemd服务文件
│   └── scripts/                   # 构建脚本
├── config/                        # 默认配置
│   └── config.yaml                # 默认配置文件
└── go.mod

tlcpchan-cli/                      # CLI工具
├── main.go                        # 程序入口
├── internal/
│   ├── commands/                  # 命令实现
│   │   ├── root.go                # 根命令
│   │   ├── start.go               # 启动命令
│   │   ├── stop.go                # 停止命令
│   │   ├── status.go              # 状态命令
│   │   ├── config.go              # 配置命令
│   │   └── cert.go                # 证书命令
│   └── client/                    # API客户端
│       └── client.go              # HTTP客户端封装
├── release/                       # 打包相关
└── go.mod

tlcpchan-ui/                       # Web前端
├── src/                           # 源代码
│   ├── views/                     # 页面组件
│   ├── components/                # 通用组件
│   ├── api/                       # API封装
│   ├── stores/                    # 状态管理
│   └── router/                    # 路由配置
├── package.json
└── vite.config.ts
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
├── ui-server/                     # Go静态资源服务器
│   ├── main.go                    # 服务入口
│   └── handler.go                 # 静态文件处理
├── dist/                          # 构建产物（嵌入Go）
└── src/                           # Vue源代码
```

UI服务作为独立进程运行，通过环境变量或配置文件连接后端API。

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

## 7. 安全设计

### 7.1 证书安全

- 私钥文件权限设置为600
- 支持证书热更新，无需重启
- 证书文件变更监控

### 7.2 网络安全

- API服务默认绑定127.0.0.1
- 支持通过配置绑定其他地址
- 日志中不记录敏感信息

### 7.3 日志安全

- 不记录证书私钥
- 可配置日志脱敏规则
- 日志文件权限控制

## 8. 性能设计

### 8.1 连接池

- 复用后端连接
- 连接健康检查
- 连接超时控制

### 8.2 并发处理

- 每连接一个goroutine
- 协程池可选（高并发场景）
- 优雅关闭

### 8.3 内存管理

- 流量缓冲区复用
- 统计数据周期清理
- 大文件传输流式处理

## 9. 扩展性设计

### 9.1 插件机制

预留插件接口，支持：
- 自定义认证
- 流量处理
- 日志输出

### 9.2 配置扩展

配置文件支持：
- 环境变量引用
- 配置文件包含
- 动态配置加载

## 10. 开发计划

### 阶段0：文档准备
- [x] 设计文档
- [ ] 需求文档
- [ ] API文档

### 阶段1：核心功能
- [ ] 配置管理
- [ ] 日志管理
- [ ] 证书加载
- [ ] 服务端代理
- [ ] 客户端代理
- [ ] 实例管理

### 阶段2：API服务
- [ ] API服务器
- [ ] 实例管理API
- [ ] 配置管理API
- [ ] 统计查询API

### 阶段3：HTTP代理
- [ ] HTTP服务器代理
- [ ] HTTP客户端代理
- [ ] 头部处理

### 阶段4：证书工具
- [ ] 证书生成
- [ ] 首次初始化

### 阶段5：UI前端
- [ ] 项目框架
- [ ] 仪表盘
- [ ] 实例管理
- [ ] 证书管理
- [ ] 日志查看

### 阶段6：CLI工具
- [ ] 命令实现
- [ ] API客户端

### 阶段7：打包发布
- [ ] 多平台构建
- [ ] 安装包制作
- [ ] GitHub CI
- [ ] Docker镜像

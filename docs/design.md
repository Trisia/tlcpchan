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

### 2.2 运行时目录结构

TLCP Channel 包含三个独立的可执行文件，推荐部署在同一目录下便于管理。

#### 完整目录结构

```
/opt/tlcpchan/ (推荐部署目录)
│
├── tlcpchan                      # [核心] tlcpchan 可执行文件
├── tlcpchan-cli                  # [CLI] 命令行工具可执行文件
├── tlcpchan-ui                   # [UI] Web 界面可执行文件
│
├── config.yaml                   # tlcpchan 主配置文件
│
├── keystores/                    # Keystore 证书文件
│   ├── tlcpchan-root-ca.crt
│   ├── tlcpchan-root-ca.key
│   ├── default-tlcp-sign.crt
│   ├── default-tlcp-sign.key
│   ├── default-tlcp-enc.crt
│   ├── default-tlcp-enc.key
│   ├── default-tls.crt
│   └── default-tls.key
│
├── rootcerts/                    # 根证书目录
│   └── tlcpchan-root-ca.crt
│
├── logs/                         # 日志目录
│   └── tlcpchan.log
│
├── ui/                           # [UI] 前端静态文件目录
│   ├── index.html
│   ├── assets/
│   └── version.txt
│
└── .tlcpchan-initialized         # 初始化标志文件
```

#### 模块说明

| 模块 | 可执行文件 | 默认端口 | 说明 |
|------|-----------|---------|------|
| **tlcpchan** | tlcpchan | 30080 | 核心代理服务、API 服务 |
| **tlcpchan-cli** | tlcpchan-cli | - | 命令行管理工具 |
| **tlcpchan-ui** | tlcpchan-ui | 30000 | Web 管理界面 |

#### 默认生成文件说明

首次启动 tlcpchan 时会自动生成以下默认文件：

| 文件名 | 路径 | 类型 | 有效期 | 说明 |
|--------|------|------|--------|------|
| tlcpchan-root-ca.crt | keystores/ | 根 CA 证书 | 10 年 | 自签名根 CA，用于签发其他证书 |
| tlcpchan-root-ca.key | keystores/ | 根 CA 私钥 | 10 年 | 根 CA 私钥，需保密 |
| tlcpchan-root-ca.crt | rootcerts/ | 根 CA 证书 | 10 年 | 根 CA 证书副本，用于信任链验证 |
| default-tlcp-sign.crt | keystores/ | TLCP 签名证书 | 5 年 | 由根 CA 签发，用于身份认证 |
| default-tlcp-sign.key | keystores/ | TLCP 签名私钥 | 5 年 | 签名证书对应的私钥 |
| default-tlcp-enc.crt | keystores/ | TLCP 加密证书 | 5 年 | 由根 CA 签发，用于密钥交换 |
| default-tlcp-enc.key | keystores/ | TLCP 加密私钥 | 5 年 | 加密证书对应的私钥 |
| default-tls.crt | keystores/ | TLS 证书 | 5 年 | 由根 CA 签发，用于 TLS 协议 |
| default-tls.key | keystores/ | TLS 私钥 | 5 年 | TLS 证书对应的私钥 |
| config.yaml | ./ | 配置文件 | - | 主配置文件，包含 keystores 和 auto-proxy 实例 |
| .tlcpchan-initialized | ./ | 标志文件 | - | 初始化完成标志 |

#### tlcpchan-ui 运行参数

```bash
tlcpchan-ui [选项]

选项:
  -listen string    监听地址 (默认 ":30000")
  -api string       后端 API 地址 (默认 "http://localhost:30080")
  -static string    前端静态文件目录 (默认 "./ui")
  -version          显示版本信息
```

### 2.3 体系结构分层设计

TLCP Channel 采用清晰的分层架构，各层职责明确：

```
┌─────────────────────────────────────────────────────────────┐
│                     接入层 (Access Layer)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │     CLI      │  │   Web UI     │  │  REST API    │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  控制层 (Controller Layer)                   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Instance API | Security API | System API | Config API│ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   业务层 (Business Layer)                    │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │ Instance Manager │  │  Security Module │                │
│  └──────────────────┘  └──────────────────┘                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    引擎层 (Engine Layer)                     │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              Proxy Engine (Server/Client/HTTP)         │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   基础设施层 (Infrastructure)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  Config  │  │  Logger  │  │  Stats   │  │  CertGen │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 2.4 安全模块架构详解

安全模块是 TLCP Channel 的核心模块之一，负责管理所有安全相关的参数：

```
┌─────────────────────────────────────────────────────────────┐
│                      Security Module                          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   KeyStore Manager                     │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  Loaders: file | named | skf | sdf             │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  KeyStores: tlcp | tls                         │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                   │
│                           ▼                                   │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  RootCert Manager                      │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  Cert Pool: x509 (TLS) + smx509 (TLCP)        │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  Formats: PEM | DER | Base64 | Hex             │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### 2.5 Keystore 与根证书说明

#### Keystore（密钥存储）

**定义：** Keystore 是用于存储证书和私钥的容器，用于 TLCP/TLS 握手时的身份认证。

**类型：**
- **TLCP 类型**：包含双证书（签名证书 + 加密证书）和对应的私钥
  - 签名证书：用于身份认证和数字签名
  - 加密证书：用于密钥交换和数据加密
- **TLS 类型**：包含单证书和对应的私钥，同时用于签名和加密

**用途：**
- 服务端代理：向客户端证明自身身份
- 客户端代理：向服务端证明自身身份（双向认证时）

**存储位置：**
- 配置信息：`config.yaml` 的 `keystores` 字段
- 证书文件：`keystores/` 目录

#### RootCert（根证书）

**定义：** 根证书是受信任的 CA 证书，用于验证对端证书的合法性。

**用途：**
- 服务端代理：验证客户端证书（双向认证时）
- 客户端代理：验证服务端证书

**双证书池设计：**
- `x509.CertPool`：用于标准 TLS 协议的证书验证
- `smx509.CertPool`：用于国密 TLCP 协议的证书验证
- 两个证书池保持同步，包含相同的根证书

**存储位置：**
- 证书文件：`rootcerts/` 目录
- 支持格式：`.pem`, `.cer`, `.crt`, `.der`

#### Keystore 与 RootCert 的关系

```
┌─────────────────────────────────────────────────────────────┐
│                     双向认证场景                               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │   客户端     │         │   服务端     │                  │
│  └──────┬───────┘         └──────┬───────┘                  │
│         │                           │                          │
│         │  1. ClientHello          │                          │
│         │  (携带随机数)             │                          │
│         ├──────────────────────────>│                          │
│         │                           │                          │
│         │  2. ServerHello          │                          │
│         │  + 证书 (Keystore)       │                          │
│         │<──────────────────────────┤                          │
│         │                           │                          │
│         │  3. 验证服务端证书        │                          │
│         │     (使用 RootCert)       │                          │
│         │                           │                          │
│         │  4. 客户端证书            │                          │
│         │     (Keystore)            │                          │
│         ├──────────────────────────>│                          │
│         │                           │                          │
│         │                           │  5. 验证客户端证书       │
│         │                           │     (使用 RootCert)      │
│         │                           │                          │
│         │  6. 密钥交换              │                          │
│         │<─────────────────────────>│                          │
│         │                           │                          │
│         │  7. 加密通信              │                          │
│         │<─────────────────────────>│                          │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**关键区别：**

| 特性 | Keystore | RootCert |
|------|----------|----------|
| **包含内容** | 证书 + 私钥 | 仅证书（公钥） |
| **用途** | 证明自身身份 | 验证对方身份 |
| **保密性** | 私钥必须保密 | 可公开 |
| **数量** | 每个实体一个 | 可包含多个 CA |
| **存储位置** | keystores/ 目录 | rootcerts/ 目录 |

### 2.6 源代码目录结构

```
tlcpchan/
├── main.go                 # 主程序入口
├── config/                 # 配置管理模块
├── initialization/         # 初始化模块
├── security/               # 安全模块
│   ├── keystore/          # Keystore 管理
│   ├── rootcert/          # 根证书管理
│   └── certgen/           # 证书生成
├── instance/              # 实例管理模块
├── proxy/                 # 代理引擎
├── controller/            # API 控制器
├── logger/                # 日志模块
└── stats/                 # 统计模块
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

#### 3.3.1 核心概念

| 概念 | 说明 |
|------|------|
| **Keystore** | 密钥存储，包含签名/加密证书和密钥 |
| **RootCert** | 根证书，用于验证对端证书 |
| **Loader** | Keystore 加载器，支持多种加载方式 |

> **详细文档**：安全参数的完整配置和管理方法请参考 [security.md](./security.md)。

#### 3.3.2 初始化流程

系统首次启动时会自动执行初始化流程，生成测试证书和默认配置。

**初始化检查：**
1. 检查配置文件是否存在
2. 检查必要的 keystores 是否存在（tlcpchan-root-ca、default-tlcp、default-tls）
3. 检查 auto-proxy 实例是否存在
4. 检查关键证书文件是否存在

**初始化流程图：**
```
开始
  │
  ▼
检查是否已初始化？
  │
  ├─ 是 ──────────────────┐
  │                        │
  否                       │
  │                        │
  ▼                        │
生成根 CA 证书 (10年有效期)
  │
  ▼
保存到 keystores/ 和 rootcerts/
  │
  ▼
生成 TLCP 双证书 (签名+加密，5年有效期)
  │
  ▼
生成 TLS 单证书 (5年有效期)
  │
  ▼
配置 keystores 到 config.yaml
  │
  ▼
配置 auto-proxy 实例
  │
  ▼
保存配置文件
  │
  ▼
创建初始化标志文件
  │
  ▼
初始化完成
```

**初始化生成的内容：**
- `tlcpchan-root-ca`：根 CA 证书（用于签发其他证书）
- `default-tlcp`：TLCP 双证书（签名证书 + 加密证书）
- `default-tls`：TLS 单证书
- `auto-proxy`：默认代理实例（监听 :30443，转发到 API 服务 :30080）

#### 3.3.3 Keystore 管理

Keystore 管理器负责管理所有密钥存储，提供统一的访问接口。

**Manager 数据结构：**
```go
type Manager struct {
    keyStores    map[string]KeyStore      // 已加载的 keystore 实例
    keyStoreInfo map[string]*KeyStoreInfo // keystore 元信息
    loaders      map[LoaderType]Loader    // 加载器映射
    mu           sync.RWMutex             // 读写锁
}
```

**加载器类型：**
| 类型 | 说明 |
|------|------|
| `file` | 从文件系统加载（默认） |
| `named` | 通过名称引用已存在的 keystore |
| `skf` | SKF 硬件接口（预留） |
| `sdf` | SDF 硬件接口（预留） |

**核心接口：**
- `LoadFromConfigs(configs []ConfigEntry)` - 从配置批量加载
- `Create(name, loaderType, params, protected)` - 创建新 keystore
- `Delete(name)` - 删除 keystore
- `Get(name)` - 获取 keystore 元信息
- `GetKeyStore(name)` - 获取 keystore 实例
- `List()` - 列出所有 keystore
- `Reload(name)` - 重新加载指定 keystore
- `ReloadAll()` - 重新加载所有 keystore

**受保护 Keystore 机制：**
- 实例配置直接创建的 keystore 会被标记为 `protected: true`
- 受保护的 keystore 不允许通过 API 删除
- 命名规则：`instance-<实例名>`

**内存管理与持久化分离：**
- Keystore Manager 仅负责内存中的 keystore 管理
- 持久化由控制器层通过 `config.Config.KeyStores` 负责
- 配置变更后需要调用 `Reload()` 生效

#### 3.3.4 根证书管理

根证书管理器负责管理所有信任的根证书，提供证书验证功能。

**Manager 数据结构：**
```go
type Manager struct {
    baseDir    string
    certs      map[string]*RootCert
    certPool   *x509.CertPool        // 标准 TLS 证书池
    smCertPool *smx509.CertPool      // 国密 TLCP 证书池
    mu         sync.RWMutex
}
```

**证书格式支持：**
- PEM 格式（.pem, .cer, .crt）
- DER 格式（.der）
- Base64 编码
- Hex 编码

**双证书池设计：**
- `certPool`：标准 x509 证书池，用于 TLS 协议
- `smCertPool`：国密 smx509 证书池，用于 TLCP 协议
- 两个证书池保持同步，包含相同的根证书

**核心接口：**
- `Initialize()` - 初始化并加载所有根证书
- `Add(filename, certData)` - 添加根证书
- `Delete(filename)` - 删除根证书
- `Get(filename)` - 获取根证书
- `List()` - 列出所有根证书
- `GetPool()` - 获取根证书池
- `Reload()` - 重新加载所有根证书

**目录扫描与自动加载：**
- 扫描 `rootcerts/` 目录
- 自动识别支持的证书扩展名
- 解析并加载所有有效证书
- 忽略无效证书文件并记录日志

#### 3.3.5 热更新机制

**Keystore 热更新：**
```bash
# 重载单个 keystore
POST /api/security/keystores/:name/reload
```
- 清空内存中的 keystore 缓存
- 下次访问时自动重新加载
- 更新 `UpdatedAt` 时间戳

**根证书热更新：**
```bash
# 重载所有根证书
POST /api/security/rootcerts/reload
```
- 重新扫描 `rootcerts/` 目录
- 重建两个证书池
- 新连接使用更新后的证书池

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

### 3.6 初始化模块

初始化模块负责系统首次启动时的初始化工作，包括生成测试证书、创建默认配置等。

**Manager 数据结构：**
```go
type Manager struct {
    cfg        *config.Config
    configPath string
    workDir    string
}
```

**核心接口：**
- `CheckInitialized() bool` - 检查是否已初始化
- `Initialize() error` - 执行完整初始化流程

**检查初始化状态的逻辑：**
1. 检查配置文件是否存在
2. 读取配置并检查必要的 keystores
   - `tlcpchan-root-ca`：根 CA
   - `default-tlcp`：默认 TLCP 证书
   - `default-tls`：默认 TLS 证书
3. 检查 auto-proxy 实例是否存在
4. 检查关键证书文件是否存在于文件系统

**完整初始化步骤：**
1. **生成根 CA 证书**
   - 使用 SM2 算法生成密钥对
   - 自签名 CA 证书，有效期 10 年
   - 保存到 `keystores/tlcpchan-root-ca.crt/.key`
   - 同时复制到 `rootcerts/` 目录供 RootCertManager 使用

2. **生成 TLCP 证书对**
   - 签名证书：用于身份认证
   - 加密证书：用于密钥交换
   - 由根 CA 签发，有效期 5 年
   - 保存到 `keystores/default-tlcp-sign.crt/.key` 和 `default-tlcp-enc.crt/.key`

3. **生成 TLS 证书**
   - 单证书模式（同时用于签名和加密）
   - 由根 CA 签发，有效期 5 年
   - 保存到 `keystores/default-tls.crt/.key`

4. **配置 keystores**
   - 在 config.yaml 中配置三个 keystores
   - 使用 file 类型加载器

5. **配置 auto-proxy 实例**
   - 类型：server
   - 监听：:30443
   - 目标：127.0.0.1:30080（API 服务）
   - 协议：auto（自动检测 TLCP/TLS）

6. **保存配置文件**
   - 写入 config.yaml
   - 创建初始化标志文件 `.tlcpchan-initialized`

**启动流程集成：**
```
main() 启动
  │
  ▼
加载或创建默认配置
  │
  ▼
CheckInitialized()?
  │
  ├─ 否 → Initialize() → 继续
  │
  └─ 是 → 继续
  │
  ▼
初始化日志
  │
  ▼
加载 keystores
  │
  ▼
初始化根证书管理器
  │
  ▼
创建并启动实例
  │
  ▼
启动 API 服务
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
├── ui/                            # 前端构建产物（运行时使用）
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


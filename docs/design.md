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
│  ┌─────────────┐              ┌─────────────────────────┐     │
│  │ tlcpchan-cli│              │      tlcpchan (内核)     │     │
│  │  (CLI工具)   │              │                         │     │
│  └──────┬──────┘              │  ┌───────────────────┐  │     │
│         │                     │  │    Controller     │  │     │
│         │                     │  │   (RESTful API)   │  │     │
│         │                     │  └─────────┬─────────┘  │     │
│         │                     │            │            │     │
│         │                     │  ┌─────────▼─────────┐  │     │
│         │                     │  │ Instance Manager  │  │     │
│         │                     │  └─────────┬─────────┘  │     │
│         │                     │            │            │     │
│         │                     │  ┌─────────▼─────────┐  │     │
│         │                     │  │    Proxy Engine   │  │     │
│         │                     │  │ ┌───────┐┌───────┐│  │     │
│         │                     │  │ │Server ││Client ││  │     │
│         │                     │  │ │Proxy  ││Proxy  ││  │     │
│         │                     │  │ └───────┘└───────┘│  │     │
│         │                     │  │ ┌─────────────────┐│  │     │
│         │                     │  │ │  HTTP Proxy     ││  │     │
│         │                     │  │ └─────────────────┘│  │     │
│         │                     │  └───────────────────┘  │     │
│         │                     │            │            │     │
│         │                     │  ┌─────────▼─────────┐  │     │
│         │                     │  │  Security Module  │  │     │
│         │                     │  │ ┌───────────────┐ │  │     │
│         │                     │  │ │KeyStore Manager│ │  │     │
│         │                     │  │ └───────────────┘ │  │     │
│         │                     │  │ ┌───────────────┐ │  │     │
│         │                     │  │ │RootCert Manager│ │  │     │
│         │                     │  │ └───────────────┘ │  │     │
│         │                     │  └───────────────────┘  │     │
│         │                     │  ┌───────────────────┐  │     │
│         │                     │  │   Stats Module    │  │     │
│         │                     │  │   Logger Module   │  │     │
│         │                     │  │   Config Module   │  │     │
│         │                     │  │   UI Static Files │  │     │
│         │                     │  └───────────────────┘  │     │
└─────────┴─────────────────────┴─────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
         │                                   │
         └───────────────────────────────────┘
                    HTTP RESTful API / UI
```

### 2.2 运行时目录结构

TLCP Channel 包含两个独立的可执行文件，推荐部署在同一目录下便于管理。

**工作目录：**
- Linux/Unix 默认：`/etc/tlcpchan`
- Windows：程序所在目录

#### 完整目录结构

```
/opt/tlcpchan/ (推荐部署目录)
│
├── tlcpchan                      # [核心] tlcpchan 可执行文件（提供 API 和 UI）
├── tlcpchan-cli                  # [CLI] 命令行工具可执行文件
│
├── config.yaml                   # tlcpchan 主配置文件
│
├── keystores/                    # Keystore 证书文件
│   ├── tlcpchan-tlcp-root-ca.crt   # TLCP 根 CA 证书（SM2）
│   ├── tlcpchan-tlcp-root-ca.key   # TLCP 根 CA 私钥（SM2）
│   ├── tlcpchan-tls-root-ca.crt    # TLS 根 CA 证书（RSA 2048）
│   ├── tlcpchan-tls-root-ca.key    # TLS 根 CA 私钥（RSA 2048）
│   ├── default-tlcp-sign.crt
│   ├── default-tlcp-sign.key
│   ├── default-tlcp-enc.crt
│   ├── default-tlcp-enc.key
│   ├── default-tls.crt
│   └── default-tls.key
│
├── rootcerts/                    # 根证书目录
│   ├── tlcpchan-tlcp-root-ca.crt   # TLCP 根 CA 证书
│   └── tlcpchan-tls-root-ca.crt    # TLS 根 CA 证书
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
| **tlcpchan** | tlcpchan | 20080 | 核心代理服务、API 服务、Web UI |
| **tlcpchan-cli** | tlcpchan-cli | - | 命令行管理工具 |

#### 访问路径

- Web UI: `http://host:20080/` 或 `http://host:20080/ui/`
- RESTful API: `http://host:20080/api/`

#### 默认生成文件说明

首次启动 tlcpchan 时会自动生成以下默认文件：

| 文件名 | 路径 | 类型 | 有效期 | 说明 |
|--------|------|------|--------|------|
| tlcpchan-tlcp-root-ca.crt | keystores/ | TLCP 根 CA 证书 | 10 年 | 自签名 SM2 根 CA，用于签发 TLCP 证书 |
| tlcpchan-tlcp-root-ca.key | keystores/ | TLCP 根 CA 私钥 | 10 年 | SM2 根 CA 私钥，需保密 |
| tlcpchan-tlcp-root-ca.crt | rootcerts/ | TLCP 根 CA 证书 | 10 年 | TLCP 根 CA 证书副本，用于信任链验证 |
| tlcpchan-tls-root-ca.crt | keystores/ | TLS 根 CA 证书 | 10 年 | 自签名 RSA 2048 根 CA，用于签发 TLS 证书 |
| tlcpchan-tls-root-ca.key | keystores/ | TLS 根 CA 私钥 | 10 年 | RSA 2048 根 CA 私钥，需保密 |
| tlcpchan-tls-root-ca.crt | rootcerts/ | TLS 根 CA 证书 | 10 年 | TLS 根 CA 证书副本，用于信任链验证 |
| default-tlcp-sign.crt | keystores/ | TLCP 签名证书 | 5 年 | 由 TLCP 根 CA 签发，用于身份认证 |
| default-tlcp-sign.key | keystores/ | TLCP 签名私钥 | 5 年 | 签名证书对应的私钥 |
| default-tlcp-enc.crt | keystores/ | TLCP 加密证书 | 5 年 | 由 TLCP 根 CA 签发，用于密钥交换 |
| default-tlcp-enc.key | keystores/ | TLCP 加密私钥 | 5 年 | 加密证书对应的私钥 |
| default-tls.crt | keystores/ | TLS 证书 | 5 年 | 由 TLS 根 CA 签发（RSA 2048），用于 TLS 协议 |
| default-tls.key | keystores/ | TLS 私钥 | 5 年 | TLS 证书对应的私钥 |
| config.yaml | ./ | 配置文件 | - | 主配置文件，包含 keystores 和 auto-proxy 实例 |
| .tlcpchan-initialized | ./ | 标志文件 | - | 初始化完成标志 |

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
tlcpchan/                      # 项目根目录
├── tlcpchan/                  # [核心] 核心服务（Go）
│   ├── main.go                # 主程序入口
│   ├── config/                # 配置管理模块
│   ├── initialization/        # 初始化模块
│   ├── security/              # 安全模块
│   │   ├── keystore/         # Keystore 管理
│   │   ├── rootcert/         # 根证书管理
│   │   └── certgen/          # 证书生成
│   ├── instance/             # 实例管理模块
│   ├── proxy/                # 代理引擎
│   ├── controller/           # API 控制器
│   ├── logger/               # 日志模块
│   └── stats/                # 统计模块
├── tlcpchan-cli/              # [CLI] 命令行工具（Go）
│   ├── main.go
│   ├── client/
│   └── commands/
├── tlcpchan-ui/               # [UI] Web 前端（Vue/TypeScript）
│   ├── src/
│   │   ├── api/
│   │   ├── views/
│   │   ├── layouts/
│   │   ├── types/
│   │   └── router/
│   └── public/
├── design.md                  # 设计文档
└── AGENTS.md                  # Agent 指南
```

## 3. 核心模块设计

### 3.1 代理模块

#### 3.1.1 服务端代理（TLCP/TLS → TCP）

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

#### 3.1.2 客户端代理（TCP → TLCP/TLS）

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

#### 3.1.3 HTTP代理

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

#### 3.1.4 双层Config架构设计


**解决方案：双层Config架构**

通过引入双层配置结构，解决配置热重载问题：

```
┌─────────────────────────────────────────────────────────────┐
│                    TLCPAdapter 配置结构                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              外层 Config (Outer Config)              │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ outerTLCPConfig (stable reference)           │  │   │
│  │  │   - 稳定持有，listener 可以安全持有          │  │   │
│  │  │   - 通过 GetConfigForClient 回调动态获取      │  │   │
│  │  │   - 仅在首次 reload 时创建                   │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ outerTLSConfig (stable reference)            │  │   │
│  │  │   - 稳定持有，listener 可以安全持有          │  │   │
│  │  │   - 通过 GetConfigForClient 回调动态获取      │  │   │
│  │  │   - 仅在首次 reload 时创建                   │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                   │
│                           │ GetConfigForClient                │
│                           │ 回调                              │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │        原子指针 (Atomic Pointers)                    │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ atomicTLCPConfig (atomic.Value)                │  │   │
│  │  │   - 并发安全读取                               │  │   │
│  │  │   - 每次 reload 后替换                         │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ atomicTLSConfig (atomic.Value)                 │  │   │
│  │  │   - 并发安全读取                               │  │   │
│  │  │   - 每次 reload 后替换                         │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              内层 Config (Inner Config)              │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ tlcpConfig (actual configuration)            │  │   │
│  │  │   - 包含实际配置数据                          │  │   │
│  │  │   - 每次 reload 时重新构建                   │  │   │
│  │  │   - 证书直接设置到 Certificates 字段          │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │ tlsConfig (actual configuration)             │  │   │
          │  │   - 包含实际配置数据                          │  │   │
│  │  │   - 每次 reload 时重新构建                   │  │   │
│  │  │   - 证书直接设置到 Certificates 字段          │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

**核心数据结构：**

```go
type TLCPAdapter struct {

    // 内层 Config (reload 时替换)
    tlcpConfig       *tlcp.Config
    tlsConfig        *tls.Config
    
    // 外层 Config (listener 稳定持有)
    outerTLCPConfig  *tlcp.Config
    outerTLSConfig   *tls.Config
    
    // 原子指针 (并发安全读取)
    atomicTLCPConfig atomic.Value  // 存储 *tlcp.Config
    atomicTLSConfig  atomic.Value  // 存储 *tls.Config
   
}
```

**配置更新流程（Server 类型）：**

```
ReloadConfig() 调用
    │
    ▼
1. 完全构建内层 Config (无锁)
    - 加载证书
    - 创建 Config 对象
    - 设置所有配置项
    - 证书直接设置到 Certificates 字段
    │
    ▼
2. 在锁保护下更新所有引用
    a.mu.Lock()
    │
    ├─> 更新内部引用
    │    a.tlcpConfig = tlcpConfig
    │    a.tlsConfig = tlsConfig
    │
    ├─> 首次 reload 时创建外层 Config
    │    if a.outerTLCPConfig == nil && tlcpConfig != nil {
    │        a.outerTLCPConfig = &tlcp.Config{
    │            GetConfigForClient: func(...) (*tlcp.Config, error) {
    │                return a.atomicTLCPConfig.Load().(*tlcp.Config), nil
    │            },
    │        }
    │    }
    │
    └─> 最后才原子替换 (关键!)
         a.atomicTLCPConfig.Store(tlcpConfig)
         a.atomicTLSConfig.Store(tlsConfig)
         
    a.mu.Unlock()
    │
    ▼
3. 新连接使用最新配置
```

**配置读取路径：**

| 场景 | 配置读取路径 | 说明 |
|------|------------|------|
| **Server Listener 接受新连接** | 外层 Config → GetConfigForClient → atomic.Value → `*tlcp.Config` | Listener 持有外层 Config，通过回调动态获取最新的内层 Config |
| **Client Dial** | 直接从 atomic.Value 读取 `*tlcp.Config` | Client 不创建 listener，直接使用原子指针读取最新配置 |
| **健康检查** | 从 atomic.Value 读取，然后 Clone | 每次检查都 Clone 新的配置对象，不保存引用 |

**关键设计原则：**

1. **先构建，再替换**
   - 内层 Config 必须完全构建好后才能原子替换
   - 确保其他线程读取到的是完整配置

2. **外层 Config 稳定持有**
   - 外层 Config 只创建一次
   - Listener 可以安全持有，无需担心配置变更
   - 通过 GetConfigForClient 回调动态获取最新配置

3. **原子操作保证并发安全**
   - 使用 `atomic.Value` 存储内层 Config 指针
   - 提供无锁的并发读取能力
   - 配置更新时原子的 Store 新配置

4. **证书直接设置**
   - 改用 `Certificates` 字段直接设置证书
   - 移除 `GetCertificate` 回调
   - 简化配置热重载逻辑

**热重载前后对比：**

**重载前（旧设计）：**
```
创建 listener 时
  listener 持有 config 指针 ─────┐
                                 │
ReloadConfig() 调用              │
  替换 tlcpConfig/tlsConfig      │
  但 listener 仍持有旧指针 ──────┼──> 问题：无法感知配置更新
                                 │
新连接                          │
  使用 listener 持有的旧配置 ───┘
```

**重载后（双层设计）：**
```
首次 reload 时
  创建外层 Config (stable)
  listener 持有外层 Config ─────────────┐
  外层 Config.SetGetConfigForClient ─────┤
                                        │
后续 reload                             │
  构建新内层 Config                     │
  atomic.Value.Store(新内层Config) ─────┘

新连接
  外层 Config.GetConfigForClient()
    -> atomic.Value.Load()
    -> 返回最新的内层 Config ✅
```

### 3.2 安全参数管理模块

安全参数（Keystore、根证书）的详细配置和管理方法请参考 [security.md](./security.md)。

安全参数管理模块统一管理 Keystore 和根证书，提供抽象的接口设计，支持多种加载方式。

#### 3.2.1 核心概念

| 概念 | 说明 |
|------|------|
| **Keystore** | 密钥存储，包含签名/加密证书和密钥 |
| **RootCert** | 根证书，用于验证对端证书 |
| **Loader** | Keystore 加载器，支持多种加载方式 |

> **详细文档**：安全参数的完整配置和管理方法请参考 [security.md](./security.md)。

#### 3.2.2 初始化流程

系统首次启动时会自动执行初始化流程，生成测试证书和默认配置。

**初始化检查：**
1. 检查配置文件是否存在
2. 检查必要的 keystores 是否存在（tlcpchan-tlcp-root-ca、tlcpchan-tls-root-ca、default-tlcp、default-tls）
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
生成 TLCP 根 CA 证书 (SM2，10年有效期)
  │
  ▼
生成 TLS 根 CA 证书 (RSA 2048，10年有效期)
  │
  ▼
保存根证书到 keystores/ 和 rootcerts/
  │
  ▼
用 TLCP 根 CA 签发 TLCP 双证书 (签名+加密，5年有效期)
  │
  ▼
用 TLS 根 CA 签发 TLS 单证书 (RSA 2048，5年有效期)
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
- `tlcpchan-tlcp-root-ca`：TLCP 根 CA 证书（SM2，用于签发 TLCP 证书）
- `tlcpchan-tls-root-ca`：TLS 根 CA 证书（RSA 2048，用于签发 TLS 证书）
- `default-tlcp`：TLCP 双证书（签名证书 + 加密证书，由 TLCP 根 CA 签发）
- `default-tls`：TLS 单证书（RSA 2048，由 TLS 根 CA 签发）
- `auto-proxy`：默认代理实例（监听 :20443，转发到 API 服务 :20080）

#### 3.2.3 Keystore 管理

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

**受保护 Keystore 机制：**
- 实例配置直接创建的 keystore 会被标记为 `protected: true`
- 受保护的 keystore 不允许通过 API 删除
- 命名规则：`instance-<实例名>`

**内存管理与持久化分离：**
- Keystore Manager 仅负责内存中的 keystore 管理
- 持久化由控制器层通过 `config.Config.KeyStores` 负责

#### 3.2.4 根证书管理

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

#### 3.2.5 热更新机制

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

### 3.3 实例管理模块

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

### 3.4 统计模块

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

### 3.5 初始化模块

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
1. **生成 TLCP 根 CA 证书**
   - 使用 SM2 算法生成密钥对
   - 自签名 CA 证书，有效期 10 年
   - 保存到 `keystores/tlcpchan-tlcp-root-ca.crt/.key`
   - 同时复制到 `rootcerts/` 目录供 RootCertManager 使用

2. **生成 TLS 根 CA 证书**
   - 使用 RSA 2048 算法生成密钥对
   - 自签名 CA 证书，有效期 10 年
   - 保存到 `keystores/tlcpchan-tls-root-ca.crt/.key`
   - 同时复制到 `rootcerts/` 目录供 RootCertManager 使用

3. **生成 TLCP 证书对**
   - 签名证书：用于身份认证
   - 加密证书：用于密钥交换
   - 由 TLCP 根 CA 签发，有效期 5 年
   - 保存到 `keystores/default-tlcp-sign.crt/.key` 和 `default-tlcp-enc.crt/.key`

4. **生成 TLS 证书**
   - 单证书模式（同时用于签名和加密）
   - 使用 RSA 2048 算法
   - 由 TLS 根 CA 签发，有效期 5 年
   - 保存到 `keystores/default-tls.crt/.key`

5. **配置 keystores**
   - 在 config.yaml 中配置四个 keystores
   - 使用 file 类型加载器

5. **配置 auto-proxy 实例**
   - 类型：server
   - 监听：:20443
   - 目标：127.0.0.1:20080（API 服务）
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

### 4.2 完整API路由表

系统共提供 36 个 RESTful API 接口，分为 5 个主要类别：

#### 4.2.1 Instance API (13个)

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/instances | 获取所有实例列表 | - | 实例数组，包含名称、状态、配置、是否启用 |
| POST | /api/instances | 创建实例 | 实例配置对象 | 创建的实例信息 |
| GET | /api/instances/:name | 获取实例详情 | - | 实例详细信息 |
| PUT | /api/instances/:name | 更新实例配置 | 更新后的实例配置 | 更新后的实例信息 |
| DELETE | /api/instances/:name | 删除实例 | - | 确认删除成功 |
| POST | /api/instances/:name/start | 启动实例 | - | 实例状态 |
| POST | /api/instances/:name/stop | 停止实例 | - | 实例状态 |
| POST | /api/instances/:name/reload | 重载实例 | - | 实例状态 |
| POST | /api/instances/:name/restart | 重启实例 | - | 实例状态 |
| GET | /api/instances/:name/stats | 获取统计信息 | - | 统计数据对象 |
| GET | /api/instances/:name/logs | 获取日志 | - | 日志列表 |
| GET | /api/instances/:name/health | 实例健康检查 | - | 健康检查结果 |

#### 4.2.2 Security API (13个)

**Keystore API (7个):**

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/security/keystores | 获取 keystore 列表 | - | keystore 数组 |
| POST | /api/security/keystores | 创建 keystore | keystore 配置（支持 multipart/form-data） | 创建的 keystore 信息 |
| GET | /api/security/keystores/:name | 获取 keystore 详情 | - | keystore 详细信息 |
| PUT | /api/security/keystores/:name | 更新 keystore 参数 | params 对象 | 更新后的 keystore 信息 |
| POST | /api/security/keystores/:name/upload | 上传更新 keystore 证书和密钥 | multipart/form-data（signCert/signKey/encCert/encKey） | 更新后的 keystore 信息 |
| DELETE | /api/security/keystores/:name | 删除 keystore | - | 确认删除成功 |
| POST | /api/security/keystores/generate | 生成新 keystore | keystore 生成参数 | 生成的 keystore 信息 |
| POST | /api/security/keystores/:name/export-csr | 导出 CSR | CSR 文件（二进制流） | - 文件流下载 |

**RootCert API (6个):**

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/security/rootcerts | 获取根证书列表 | - | 根证书数组（包含主题、颁发者、过期时间等） |
| POST | /api/security/rootcerts | 添加根证书 | multipart/form-data（filename + cert） | 添加的根证书信息 |
| GET | /api/security/rootcerts/:filename | 下载根证书（二进制流） | - | 文件流下载 |
| DELETE | /api/security/rootcerts/:filename | 删除根证书 | - | 确认删除成功 |
| POST | /api/security/rootcerts/generate | 生成根 CA 证书 | 根 CA 生成参数 | 生成的根 CA 信息 |
| POST | /api/security/rootcerts/reload | 重载所有根证书 | - | 确认重载成功 |

#### 4.2.3 System API (3个)

| 方法 | 路径 | 描述 | 响应体 |
|------|------|------|--------|
| GET | /api/system/info | 获取系统信息 | 系统信息对象（操作系统、架构、内存、CPU、Goroutine 等）|
| GET | /api/system/health | 系统健康检查 | 状态和版本信息 |
| GET | /api/system/version | 版本信息 | 版本号 |
| GET | /api/version | 版本信息（别名） | 版本号（同上） |

#### 4.2.4 Config API (4个)

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/config | 获取当前配置 | - | 完整配置对象 |
| POST | /api/config | 更新配置 | 配置对象 | 更新后的配置 |
| POST | /api/config/reload | 重载配置 | - | 确认重载成功 |
| POST | /api/config/validate | 验证配置文件 | 验证结果（支持指定或默认文件） | 验证结果 |

#### 4.2.5 Logs API (4个)

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/system/logs | 列出日志文件 | - | 日志文件数组（名称、大小、修改时间、是否当前）|
| GET | /api/system/logs/content | 读取日志内容 | - | 日志行数组（支持行数和级别过滤）|
| GET | /api/system/logs/download/:filename | 下载单个日志文件 | - | 文件流下载 |
| GET | /api/system/logs/download-all | 打包下载所有日志 | - | ZIP 文件流下载 |

**总计：37 个 API 接口**

### 4.3 配置管理设计理念

#### 4.3.1 为什么不提供 config update API？

TLCP Channel 不提供 `POST /api/config` 更新 API，而是要求用户通过编辑配置文件后使用 `config reload` 重载配置。这个设计基于以下考虑：

**1. 原子性考虑**
- 配置更新是复杂的多步骤操作（读取 → 验证 → 更新 → 写入 → 重载），通过 API 很难保证原子性和一致性
- YAML 配置文件格式复杂，直接编辑更直观，可避免 API 部分更新导致配置损坏
- 用户回滚能力：编辑配置文件后，用户可以手动回滚，API 模式增加了复杂性

**2. 数据完整性**
- 配置文件是唯一的数据源，避免 API 更新与文件状态不一致
- 简化配置管理，减少数据不一致风险

**3. 用户习惯**
- 配置文件编辑是用户熟悉的运维方式
- 与系统管理工具（systemd, puppet, ansible）集成方便

**4. 简化设计**
- 减少一个 API 接口，降低维护成本
- 配置文件即真理，API 只是配置文件的视图

#### 4.3.2 配置更新的正确流程

```
┌─────────────────────────────────────────────────────────────┐
│      配置更新的正确流程               │
├─────────────────────────────────────────────────────┤
│                                          │
│  1. 使用编辑器编辑 config.yaml     │
│                                          │
│   2. 验证配置有效性              │
│     → config validate                │
│                                          │
│  3. 重载配置                     │
│     → config reload                 │
│                                          │
│  4. 验证配置已生效              │
│     → system info                  │
│     → config show                  │
│                                          │
└─────────────────────────────────────────────────────────────┘
```

**关键操作：**
1. 编辑 `config.yaml`
2. `tlcpchan-cli config validate` - 验证配置文件格式正确性
3. `tlcpchan-cli config reload` - 重载配置使配置生效
4. `tlcpchan-cli config show` - 确认配置已更新

#### 4.3.3 配置热重载机制

TLCP Channel 提供配置热重载功能，无需重启服务即可使配置变更生效。

**重载触发时机：**
- 收到 `SIGHUP` 信号（推荐，Linux）
- 手动调用 `POST /api/config/reload`
- CLI 命令 `config reload`

**重载范围：**
- keystores 和根证书：重新扫描并重新加载到内存
- 实例配置：根据配置重新启动/停止实例
- 系统日志：重新初始化日志系统

**热重载对运行中连接的影响：**
- 新连接使用更新后的配置
- 现有连接继续使用旧配置直到完成当前请求
- 不会中断现有连接
- 证书热更新后，新连接使用新证书

**双层 Config 架构设计**

```
┌─────────────────────────────────────────────────────────────────────┐
│                    TLCPAdapter 配置结构                      │
├─────────────────────────────────────────────────────────────────────┤
│  │ 外层 Config (Outer Config)              │
│ │   │ outerTLCPConfig (stable reference)    │     │
│ │   - 稳定持有，listener 可以安全持有          │
│ │   - 通过 GetConfigForClient 回调动态获取      │
│ │   - 仅在首次 reload 时创建                   │
│ │                                     │
│ │   └ outerTLSConfig (stable reference)     │     │
│   └─────────────────────────────────────────────────────────────┤
│                                     │
│         │ 原子指针 (并发安全读取）              │
│   │ atomicTLCPConfig (atomic.Value)     │     │
│   atomicTLSConfig  atomic.Value      │     │
│                                    │
└─────────────────────────────────────────────────────────────┘

配置读取路径：│
┌─────────────────────────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────────┐
│ 场景               │ 配置读取路径              │
├─────────────────────────────────────────────────────────────┤
│ Server Listener  │ 外层 Config → GetConfigForClient │     │
│ (新连接)          │      │              │
│                  │      │              │
│                  │      │              │
│   └───────────┐              │     │
│     │      │              │
│     │   innerTLCPConfig ────────────┐
│     │      │              │
└─────────────────────────────────────────────────┘
│ (动态获取)          │              │
│     │      │              │
│ Client Dial          │      │              │
│                  │ 直接从 atomic.Value 读取最新配置       │
│                  │      │              │
└─────────────────────────────────────────────────────────┘

健康检查：          │              │
│ 从 atomic.Value 读取，然后 Clone │     │
│ 每次检查都 Clone 新的配置对象，不保存引用  │
```

**关键设计原则：**
1. **先构建，再替换**：内层 Config 必须完全构建好后才能原子替换
2. **外层 Config 稳定持有**：外层 Config 只创建一次，Listener 可以安全持有
3. **原子操作保证并发安全**：使用 `atomic.Value` 提供无锁的并发读取能力
4. **证书直接设置**：改用 `Certificates` 字段直接设置证书，移除 `GetCertificate` 回调

### 4.4 错误处理和状态码规范

#### 4.4.1 HTTP 状态码

| 状态码 | 说明 | 常见原因 |
|--------|------|----------|
| 200 OK | 请求成功 | - |
| 201 Created | 资源创建成功 | - |
| 202 Accepted | 请求已接受 | - |
| 204 No Content | 无内容（删除成功，无返回数据）| - |
| 400 Bad Request | 请求参数错误或格式不正确 | - |
| 401 Unauthorized | 未授权（需要 API Key 或认证）| - |
| 403 Forbidden | 权限访问（权限不足）| - |
| 404 Not Found | 资源不存在（文件、证书、实例等）| - |
| 409 Conflict | 资源冲突（实例名、端口等）| - |
| 422 Unprocessable Entity | 无法处理的请求体格式 | - |
| 500 Internal Server Error | 服务器内部错误 | - |
| 502 Bad Gateway | 网关错误 | - |

#### 4.4.2 错误响应格式

**JSON 格式：**
```json
{
  "error": "错误描述",
  "code": "错误码",
  "details": "详细信息（可选）"
}
```

**文本格式：**
```
错误描述: 无效的请求体
```

#### 4.4.3 特殊错误场景

| 场景 | HTTP状态码 | 错误码 | 错误描述 |
|------|-----------|----------|----------|
| 端口冲突 | 409 | 端口已被占用 | - |
| 实例名重复 | 409 | 实例名已存在 | - |
| Keystore 不存在 | 404 | Keystore 不存在 | - |
| 证书无效 | 400 | 证书格式错误或损坏 | - |
| 配置验证失败 | 400 | 配置文件格式错误 | - |
| 配置重载失败 | 500 | 配置重载失败 | - |

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
├── src/                           # Vue前端项目源代码
├── public/                        # 公共资源
├── package.json
├── vite.config.ts
├── tsconfig.json
└── ui/                            # 前端构建产物（运行时使用）
```

UI 作为纯前端静态文件，由 tlcpchan 核心服务直接提供：
1. 静态资源服务 - tlcpchan 核心服务托管 Vue 前端
2. 访问路径：
   - `/` → 重定向到 `/ui/`
   - `/ui/` → UI 界面
   - `/api/` → RESTful API
3. SPA 路由支持 - 前端路由回退到 index.html

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
COPY config /etc/tlcpchan
COPY ui /etc/tlcpchan/ui
EXPOSE 20080 20443
ENTRYPOINT ["/usr/bin/tlcpchan"]
```


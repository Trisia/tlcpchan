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
│   ├── tlcpchan-ibc-kgc-master.key # 内置测试 KGC 主密钥（SM9 主密钥对，0600）
│   ├── default-tlcp-sign.crt
│   ├── default-tlcp-sign.key
│   ├── default-tlcp-enc.crt
│   ├── default-tlcp-enc.key
│   ├── default-tls.crt
│   ├── default-tls.key
│   ├── default-ibc-server-identity.txt  # IBC 服务端标识（裸字节文本，0644）
│   ├── default-ibc-server-params.pem    # IBC 服务端本端 KGC 公共参数（0644）
│   ├── default-ibc-server-sign.key      # IBC 签名私钥 hid=0x01（0600）
│   ├── default-ibc-server-enc.key       # IBC 加密私钥 hid=0x03（0600）
│   ├── default-ibc-server-kex.key       # IBC 密钥交换私钥 hid=0x02（0600）
│   ├── default-ibc-client-identity.txt  # IBC 客户端标识（0644）
│   ├── default-ibc-client-params.pem
│   ├── default-ibc-client-sign.key
│   ├── default-ibc-client-enc.key
│   └── default-ibc-client-kex.key
│
├── rootcerts/                    # 根证书目录
│   ├── tlcpchan-tlcp-root-ca.crt   # TLCP 根 CA 证书
│   └── tlcpchan-tls-root-ca.crt    # TLS 根 CA 证书
│
├── ibcparams/                    # IBC 信任池目录（只存放信任的 KGC 公共参数）
│   └── tlcpchan-ibc-kgc.pem          # 内置测试 KGC 公共参数（0644，入信任池）
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
| default-ibc-server-identity.txt | keystores/ | IBC 标识 | - | 服务端标识裸字节文本 `server@tlcpchan.local`（0644） |
| default-ibc-server-params.pem | keystores/ | IBC KGC 公共参数 | 10 年 | 本端公共参数，源自内置测试 KGC（0644） |
| default-ibc-server-sign.key | keystores/ | IBC 签名私钥 | 10 年 | 签名用户私钥，hid=0x01（0600） |
| default-ibc-server-enc.key | keystores/ | IBC 加密私钥 | 10 年 | 加密用户私钥，hid=0x03（0600） |
| default-ibc-server-kex.key | keystores/ | IBC 密钥交换私钥 | 10 年 | 密钥交换用户私钥，hid=0x02（0600） |
| default-ibc-client-identity.txt | keystores/ | IBC 标识 | - | 客户端标识裸字节文本 `client@tlcpchan.local`（0644） |
| default-ibc-client-params.pem | keystores/ | IBC KGC 公共参数 | 10 年 | 本端公共参数（0644） |
| default-ibc-client-sign.key | keystores/ | IBC 签名私钥 | 10 年 | 签名用户私钥，hid=0x01（0600） |
| default-ibc-client-enc.key | keystores/ | IBC 加密私钥 | 10 年 | 加密用户私钥，hid=0x03（0600） |
| default-ibc-client-kex.key | keystores/ | IBC 密钥交换私钥 | 10 年 | 密钥交换用户私钥，hid=0x02（0600） |
| tlcpchan-ibc-kgc.pem | ibcparams/ | KGC 公共参数 | 10 年 | 内置测试 KGC（`tlcpchan.local#1`）公共参数，写入全局信任池（0644） |
| tlcpchan-ibc-kgc-master.key | keystores/ | KGC 主密钥 | 10 年 | 测试 KGC 主密钥，与其它密钥材料同目录、不通过任何 API 下发（0600） |
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
│  │  │  Loaders: file | named | skf | sdf | ibc-file  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  KeyStores: tlcp | tls | ibc                   │  │  │
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
│                           │                                   │
│                           ▼                                   │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                 IBCParams Manager                      │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  Directory: ibcparams/ (子目录不扫描)           │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │  Ext: .pem | .der | .ibcparams → tlcp.IBCPool  │  │  │
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
- `x509.CertPool`：用于标准 TLS 协议的证书验证，仅包含标准库 `crypto/x509` 能解析的证书（如 RSA/ECDSA）
- `smx509.CertPool`：用于国密 TLCP 协议的证书验证，包含全部根证书（SM2 与 RSA/ECDSA）
- 两个证书池由同一份根证书文件同步加载；由于标准库不支持 SM2 曲线，SM2 根证书只会进入 `smx509.CertPool`

**存储位置：**
- 证书文件：`rootcerts/` 目录
- 支持格式：`.pem`, `.cer`, `.crt`, `.der`

#### IBC（SM9）扩展

除 X.509 证书身份外，系统还支持 **IBC（SM9 标识密码）身份**，用于 TLCP 的 IBC/IBSDH 密码套件。两者相互独立，可只配置其一，也可同时配置以实现同端口混合协商。

**IBC 身份（IBC Keystore）：**
- `type: ibc`，`loaderType: ibc-file`，与证书 Keystore 并列存放于 `config.yaml` 的 `keystores` 字段
- 不使用 X.509 证书，材料由 5 项组成：本端标识、本端 KGC 公共参数、三把 SM9 用户私钥
- 私钥按派生用途（hid）区分：签名私钥 `hid=0x01`、密钥交换私钥 `hid=0x02`、加密私钥 `hid=0x03`；误装其他用途的私钥不会在装载期报错，须由导入来源保证
- 装载时仅对签名私钥做一次 SM3 摘要自签自验（`sm9.SignASN1` + `sm9.VerifyASN1`），`kex-key` 无法在装载期反推校验

**IBC 信任池（IBCParams）：**
- KGC 公共参数（`IBCSysParams`）的全局信任锚，等价于证书体系中的根证书库
- 独立目录 `ibcparams/`，只存放信任的 KGC 公共参数；KGC 主密钥存放在 `keystores/`（0600，不通过任何 API 下发）
- 客户端与服务端共用同一份信任列表，不进入实例配置；**默认拒绝**：信任池中不存在对端 KGC 时 IBC 握手一定失败
- 详细设计见 [3.2.5 IBC 信任池管理](#325-ibc-信任池管理) 与 [3.6 IBC（SM9）支持设计](#36-ibcsm9支持设计)

**与证书身份的对照：**

| 特性 | 证书身份（Keystore） | IBC 身份（IBC Keystore） |
|------|----------------------|--------------------------|
| **类型** | `tlcp` / `tls` | `ibc` |
| **加载器** | `file` / `named` / `skf` / `sdf` | `ibc-file` |
| **身份凭据** | X.509 证书（SM2/RSA/ECDSA） | 标识字节串 + KGC 公共参数 |
| **私钥** | 签名私钥 + 加密私钥 | 签名私钥（0x01）+ 加密私钥（0x03）+ 密钥交换私钥（0x02） |
| **信任锚** | `rootcerts/`（根证书） | `ibcparams/`（KGC 公共参数） |
| **可用套件** | ECC_SM4_* / ECDHE_SM4_* | IBC_SM4_* / IBSDH_SM4_* |
| **存储位置** | `keystores/` | `keystores/` |

> **安全提示**：`InsecureSkipVerify=true` 会同时跳过 X.509 证书验证与 IBC 公共参数校验，仅供测试使用，UI 中有明确警示文案。

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
│   │   ├── keystore/         # Keystore 管理（含 ibc-file 加载器）
│   │   ├── rootcert/         # 根证书管理
│   │   ├── ibcparams/        # IBC 信任池（KGC 公共参数）管理
│   │   └── certgen/          # 证书与 IBC 身份生成
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
| **Keystore** | 密钥存储，包含签名/加密证书和密钥（X.509/SM2 证书身份） |
| **RootCert** | 根证书，用于验证对端证书 |
| **Loader** | Keystore 加载器，支持多种加载方式 |
| **IBCKeystore** | IBC（SM9）身份密钥存储，包含本端标识、本端 KGC 公共参数与三把 SM9 用户私钥，不使用 X.509 证书 |
| **IBCParams** | IBC 信任池条目（KGC 公共参数），用于校验对端 IBC 身份，等价于证书体系的根证书 |

> **详细文档**：安全参数的完整配置和管理方法请参考 [security.md](./security.md)。

#### 3.2.2 初始化流程

系统首次启动时会自动执行初始化流程，生成测试证书和默认配置。

**初始化检查：**
1. 检查配置文件是否存在
2. 检查必要的 keystores 是否存在（tlcpchan-tlcp-root-ca、tlcpchan-tls-root-ca、default-tlcp、default-tls）
3. 检查 auto-proxy 实例是否存在
4. 检查关键证书文件是否存在

> **向后兼容约束**：`CheckInitialized` **不检查任何 IBC 产物**（既不看 `default-ibc-server` / `default-ibc-client` keystore，也不看 `ibcparams/` 目录与文件）。否则已运行的老安装会被判定为"未初始化"并重新初始化、覆盖 `config.yaml`；老安装应通过「生成」或「导入」接口补齐 IBC 材料。

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
生成内置测试 KGC 公共参数 (SM9，tlcpchan.local#1，10年有效期)
  │
  ▼
公共参数写入 ibcparams/ (入信任池，0644)
主密钥写入 keystores/ (0600)
  │
  ▼
派生 default-ibc-server / default-ibc-client 两组身份 (各含标识+公共参数+三把私钥)
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
- `tlcpchan-ibc-kgc`：内置测试 KGC 公共参数（`tlcpchan.local#1`，10 年有效期），写入 `ibcparams/` 信任池；主密钥写入 `keystores/tlcpchan-ibc-kgc-master.key`（0600）
- `default-ibc-server`：IBC 服务端身份（`server@tlcpchan.local`，`ibc-file` 类型）
- `default-ibc-client`：IBC 客户端身份（`client@tlcpchan.local`，`ibc-file` 类型）
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
| `ibc-file` | 从文件系统加载 IBC（SM9）身份材料（`type=ibc`），参数键为 `identity` / `params` / `sign-key` / `enc-key` / `kex-key` |

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

**RootCert 数据结构：**
```go
type RootCert struct {
    Filename     string               // 证书文件名
    Cert         *smx509.Certificate  // 解析后的证书对象（使用 smx509 统一解析）
    NotBefore    time.Time            // 证书生效时间
    NotAfter     time.Time            // 证书过期时间
    Subject      string               // 证书主题
    Issuer       string               // 证书颁发者
    KeyType      string               // 密钥类型（"SM2"、"RSA-2048"、"ECDSA-P256" 等）
    SerialNumber string               // 证书序列号（十六进制）
    Version      int                  // 证书版本
    IsCA         bool                 // 是否为 CA 证书
    KeyUsage     []string             // 密钥用途
}
```

**证书格式支持：**
- PEM 格式（.pem, .cer, .crt）
- DER 格式（.der）
- Base64 编码
- Hex 编码

**证书解析：**
- 统一使用 `smx509.ParseCertificate` 解析根证书；smx509 是标准库 `crypto/x509` 的超集，同时支持 SM2/PQC 与 RSA/ECDSA 等算法
- 解析结果以 `*smx509.Certificate` 保存在 `RootCert.Cert` 中
- 自 gmsm v0.44.0 起 `smx509.Certificate` 与 `x509.Certificate` 是相互独立的结构体（`ToX509` 已移除），两者不能互转，因此不再持有标准库证书类型
- 证书元数据（主题、颁发者、密钥类型、密钥用途、序列号等）均由该对象提取

**双证书池设计：**
- `certPool`：标准 x509 证书池，用于 TLS 协议，仅包含标准库能解析的证书（如 RSA/ECDSA）
- `smCertPool`：国密 smx509 证书池，用于 TLCP 协议，包含全部根证书（SM2 与 RSA/ECDSA）
- 两个证书池由同一份根证书文件同步追加；由于标准库不支持 SM2 曲线，SM2 根证书只会进入 `smCertPool`

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

#### 3.2.5 IBC 信任池管理

IBC 信任池管理器（`security/ibcparams`）负责管理受信任的 KGC 公共参数（`IBCSysParams`），其定位与根证书管理器对等：**是 IBC/IBSDH 套件的全局信任锚**。

**Manager 数据结构：**
```go
// Manager IBC 信任池管理器
type Manager struct {
    baseDir string               // 信任池目录，默认 <workDir>/ibcparams
    params  map[string]*IBCParam // 文件名 -> 参数元信息
    pool    *tlcp.IBCPool        // 全量信任池，Reload 时重建
    mu      sync.RWMutex         // 读写锁
}
```

**IBCParam 数据结构：**
```go
// IBCParam KGC 公共参数条目
type IBCParam struct {
    Filename           string    // 文件名，条目在信任池中的唯一标识
    DistrictName       string    // KGC 属地区域名，与 DistrictSerial 共同唯一标识 KGC
    DistrictSerial     int       // 同一区域下的 KGC 序号
    NotBefore          time.Time // 公共参数生效时间（零值表示不限）
    NotAfter           time.Time // 公共参数失效时间（零值表示不限）
    IssuerIdentity     string    // 公共参数颁发者标识（KGC 标识）可读形式
    SignKeyFingerprint string    // 签名主公钥 SM3 指纹（HEX 小写）
    EncKeyFingerprint  string    // 加密主公钥 SM3 指纹（HEX 小写）
}
```

> `IBCParam` 仅包含只读元信息，解析后的 `*tlcp.IBCSysParams` 仅供内部装载与比对使用，不对外序列化。

**存储位置与目录约定：**
- 信任池目录：`<workDir>/ibcparams/`（默认 `/etc/tlcpchan/ibcparams`，与 `rootcerts/` 同级）
- KGC 主密钥**不在**本目录：内置测试 KGC 主密钥与生成接口产出的主密钥统一存放在 `<workDir>/keystores/`（0600）
- 通过管理接口（`Add`）写入信任池的文件权限统一为 `0600`；初始化预置的测试 KGC 公共参数为 `0644`（公共参数不含私钥，便于分发给对端）
- 子目录（如 `master/`）**不参与扫描**，主密钥绝不通过任何 API 下发

**支持格式：**
- 扩展名白名单：`.pem`, `.der`, `.ibcparams`（大写扩展名按小写匹配）
- 内容编码：PEM / DER / HEX / Base64，解析时统一归一化为 DER 后交给 `tlcp.ParseIBCSysParams`

**核心接口：**
- `Initialize()` - 初始化并扫描加载全部 KGC 公共参数（目录不存在视为空池）
- `Add(filename, data)` - 添加 KGC 公共参数（写入目录后重新加载）
- `Delete(filename)` - 删除 KGC 公共参数
- `Get(filename)` - 获取条目元信息
- `List()` - 列出所有条目元信息
- `GetPool()` - 获取全量 `*tlcp.IBCPool`
- `Reload()` - 重新扫描目录并重建信任池
- `ReadFile(filename)` - 读取信任池中文件的原始内容（供下载接口使用）

**目录扫描与自动加载：**
- 扫描 `ibcparams/` 目录，跳过所有子目录
- 仅处理扩展名白名单内的文件
- 无效文件、解析失败的文件与重复 KGC 会被跳过并记录日志，不影响其余条目
- 同一 KGC 由 `(DistrictName, DistrictSerial)` 唯一标识：**同一 KGC 使用不同文件名重复添加会被拒绝**（加载期同样跳过重复项并告警）

**信任语义（默认拒绝）：**
- 客户端与服务端**共用同一份信任列表**，不做实例级子集选择
- 适配器在装配 `tlcp.Config` 时**总是显式注入非 nil 的信任池（即使为空）**：
  - 服务端 → `ClientIBCSysParams`
  - 客户端 → `RootIBCSysParams`
- 显式注入空池用于压掉 gotlcp"本端公共参数退化为默认信任池"的行为，因此**信任池中不存在对端 KGC 时 IBC 握手一定失败**，与"没有根证书就验不过证书"对等
- 修改信任池后需重载相关实例才会生效，详见 [3.2.6 热更新机制](#326-热更新机制)

**安全模型：**
- GM/T 0024-2023 的 Certificate 消息只传**裸** `IBCSysParams`，不带签名保护，中间人可替换加密主公钥以解密预主密钥
- 唯一缓解手段是带外预置信任锚：**必须通过可信渠道获取 KGC 公共参数后再入库**，绝不直接信任对端下发的参数
- 信任池等价于根证书库；SM9 的"无证书"不等于"无信任问题"，只是把 CA 的信任问题平移到 KGC 公共参数的带外分发
- `InsecureSkipVerify=true` 会同时跳过 X.509 与 IBC 公共参数校验，仅供测试使用

#### 3.2.6 热更新机制

**Keystore 热更新：**

keystore **没有独立的"重载"接口**。材料变更通过以下接口完成后立即生效（写入内存）：

- 证书/密钥材料：`POST /api/security/keystores/:name/upload`
- 参数变更：`PUT /api/security/keystores/:name`
- IBC 材料：同样走 `upload`（identity / params / signKey / encKey / kexKey 字段）

服务端处理时重新装载 `KeyStore` 并通过 `Manager.Set` 替换内存实例、刷新 `UpdatedAt`，无需额外重载动作；但**已启动的实例仍持有旧的 TLS/TLCP 配置**，必须 `POST /api/instances/:name/reload` 才会用新材料重建配置（Web UI 在密钥库详情页提供"重载关联实例"）。

**根证书热更新：**
```bash
# 重载所有根证书
POST /api/security/rootcerts/reload
```
- 重新扫描 `rootcerts/` 目录
- 重建两个证书池
- 新连接使用更新后的证书池

**IBC 信任池热更新：**
```bash
# 重载 IBC 信任池
POST /api/security/ibcparams/reload
```
- 重新扫描 `ibcparams/` 目录并重建 `tlcp.IBCPool`
- 与根证书一致：适配器在实例 `ReloadConfig` 时重新调用 `GetPool()`，因此**修改信任池后必须重载相关实例才会生效**（CLI 与 Web UI 均会给出该提示）：
  ```bash
  tlcpchan-cli instance reload <实例名>
  ```
- IBC keystore 的材料更新同样通过 `upload` / `PUT` 接口完成，无独立重载接口；更新后同样需要重载引用它的实例

> **实例重载提醒**：信任池对象在每次 `Reload()` 时被替换为新对象，已启动实例仍持有旧池引用；仅调用信任池重载接口不会让运行中的实例感知变化。

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

### 3.6 IBC（SM9）支持设计

IBC（Identity-Based Cryptography，标识密码）以 SM9 算法为基础，用"标识 + KGC 公共参数"取代 X.509 证书体系。本节说明 TLCP Channel 对 IBC 的完整设计：keystore 层、信任池、实例配置与混合协商、密码套件与配置校验、能力诊断、生成与导入。信任池的详细管理接口见 [3.2.5 IBC 信任池管理](#325-ibc-信任池管理)，热更新见 [3.2.6 热更新机制](#326-热更新机制)。

> 依据：GM/T 0024-2023；依赖 `gitee.com/Trisia/gotlcp`（IBC 能力）与 `github.com/emmansun/gmsm`（SM9 实现）。

#### 3.6.1 支持范围与密码套件

**支持的套件：**

| 套件名 | 密钥交换 | 加密 | 校验 | 值 |
|--------|---------|------|------|-----|
| `IBC_SM4_GCM_SM3` | IBC | SM4-GCM | SM3 | 0xE057 |
| `IBC_SM4_CBC_SM3` | IBC | SM4-CBC | SM3 | 0xE017 |
| `IBSDH_SM4_GCM_SM3` | IBSDH | SM4-GCM | SM3 | 0xE055 |
| `IBSDH_SM4_CBC_SM3` | IBSDH | SM4-CBC | SM3 | 0xE015 |

- 4 个套件**默认全部关闭**：必须显式列入 `tlcp.cipher-suites`，**且本端配置了 IBC 身份（`tlcp.ibc-keystore`）才会参与协商**
- **IBC 与 IBSDH 的差别**：IBC 由客户端用服务端标识 + 加密主公钥单向加密预主密钥（无前向安全）；IBSDH 执行 SM9 密钥交换（有前向安全，且**自动强制客户端认证**）
- IBC/IBSDH 套件与证书套件（ECC/ECDHE）可在同一 `tlcp.Config` 中共存，实现**同端口混合协商**
- 会话重用沿用现有 `SessionCache`，IBC 套件同样支持

**非目标（本期不做）：**

- DTLCP（UDP）的 IBC 支持
- IRL 标识吊销
- 生产 KGC 主密钥托管（仅提供"内置测试 KGC"用于本地联调）
- 兼容 GB/T 38636-2020 的 IBC 报文（本实现以 GM/T 0024-2023 为唯一依据）

#### 3.6.2 IBC 身份 keystore

**类型与接口：**

| 项 | 值 |
|---|---|
| `KeyStoreType` | `KeyStoreTypeIBC = "ibc"` |
| `LoaderType` | `LoaderTypeIBCFile = "ibc-file"` |
| 可选接口 | `IBCKeyStore`（在 `KeyStore` 之上提供 `IBCIdentity()` 与 `IBCInfo()`） |

```go
// IBCKeyStore IBC（SM9）身份密钥存储的可选能力接口
// 仅 ibc 类型的 keystore 实现；调用方通过类型断言判断本端是否具备 IBC 能力
type IBCKeyStore interface {
    KeyStore
    // IBCIdentity 返回装载完成的 IBC 身份（标识 + KGC 公共参数 + 用户私钥）
    IBCIdentity() (*tlcp.IBCIdentity, error)
    // IBCInfo 返回 IBC 身份的只读元信息，供 API 与 UI 展示
    IBCInfo() *IBCInfo
}
```

普通 `file` / `named` keystore 无需实现该接口；适配器通过类型断言取用，断言失败即视为无 IBC 能力。

**元信息：**

`KeyStoreInfo` 增加只读字段（不含任何私钥内容）：

```go
// IBCInfo IBC（SM9）keystore 的只读元信息，用于 API 与 UI 展示
type IBCInfo struct {
    Identity          string    // 本端标识可读形式，如 server@tlcpchan.local
    HasParams         bool      // 是否提供本端 KGC 公共参数
    DistrictName      string    // KGC 属地区域名
    DistrictSerial    int       // 同区域下的 KGC 序号
    NotBefore         time.Time // 公共参数生效时间（零值表示未提供/不限）
    NotAfter          time.Time // 公共参数失效时间（零值表示未提供/不限）
    HasSignKey        bool      // 是否提供签名私钥（hid=0x01）
    HasEncryptKey     bool      // 是否提供加密私钥（hid=0x03）
    HasKeyExchangeKey bool      // 是否提供密钥交换私钥（hid=0x02）
}

type KeyStoreInfo struct {
    // ... 既有字段 ...
    IBC *IBCInfo `json:"ibc,omitempty" yaml:"ibc,omitempty"` // 仅 type=ibc 时非空
}
```

**材料参数与文件命名（对齐初始化 TLCP）：**

| params key | 含义 | 磁盘文件名 | 权限 | 格式 | 必需性 |
|---|---|---|---|---|---|
| `identity` | 本端标识，如 `server@tlcpchan.local` | `<name>-identity.txt` | 0644 | 裸字节串文本（亦兼容 PEM / `Identifier` DER） | 必需（IBSDH 强制要求非空） |
| `params` | 本端 KGC 公共参数 `IBCSysParams` | `<name>-params.pem` | 0644 | PEM（`IBC PARAMETERS`）/ DER / HEX / Base64 | 服务端必需；仅服务端认证的客户端可省 |
| `sign-key` | 签名用户私钥 PKCS#8（hid=0x01） | `<name>-sign.key` | 0600 | PEM（`PRIVATE KEY`）/ DER / HEX / Base64 | 服务端必需；双向认证客户端必需 |
| `enc-key` | 加密用户私钥 PKCS#8（hid=0x03） | `<name>-enc.key` | 0600 | 同上 | 使用 IBC 套件的服务端必需 |
| `kex-key` | 密钥交换用户私钥 PKCS#8（hid=0x02） | `<name>-kex.key` | 0600 | 同上 | 使用 IBSDH 套件必需 |

**命名对齐说明：**

| 项 | TLCP 初始化（证书身份） | IBC |
|---|---|---|
| keystore 名 | `default-tlcp` | `default-ibc-server` / `default-ibc-client` |
| 签名私钥 | `default-tlcp-sign.key` | `default-ibc-server-sign.key` |
| 加密私钥 | `default-tlcp-enc.key` | `default-ibc-server-enc.key` |
| 密钥交换私钥 | — | `default-ibc-server-kex.key` |
| 信任锚 | `tlcpchan-tlcp-root-ca.crt`（rootcerts/） | `tlcpchan-ibc-kgc.pem`（ibcparams/） |
| 测试主密钥 | — | `tlcpchan-ibc-kgc-master.key`（keystores/，0600） |

- KGC 公共参数与 TLCP 根 CA 一样"两份"：一份入信任池（`ibcparams/`），一份作为各 IBC keystore 的本端 `-params.pem`
- 日志与 API 响应不得输出任何私钥内容

**装载流程**（实现位置：`tlcpchan/security/keystore/ibc_loader.go`）：

1. `der.Any2DER` 宽松解码公共参数与三把私钥（PEM / DER / HEX / Base64）
2. 标识按"优先解 PEM / `Identifier` DER，否则按裸字节串"处理
3. 三把私钥由 `smx509.ParsePKCS8PrivateKey` 解析，支持 `*sm9.SignPrivateKey` / `*sm9.EncryptPrivateKey`
4. `tlcp.LoadIBCIdentity(identity, paramsDER, signKeyDER, encKeyDER, kexKeyDER)` 装载身份
5. 对上一步得到的**签名私钥**执行 SM3 摘要自签自验（`sm9.SignASN1` + `sm9.VerifyASN1`）
6. 生成 `IBCInfo` 元信息（标识、`districtName#serial`、有效期、三把私钥齐备性）
7. 惰性装载 + 缓存（与现有 `FileKeyStore` 一致，`sync.RWMutex` 保护）；加载器在创建时立即装载一次，使导入阶段即可发现材料错误

**校验边界（重要）：**

- gotlcp 的 `LoadIBCIdentity` **不校验** `kex-key` 的派生用途（应为 hid=0x02）。误装其他 hid 的私钥不会在装载期报错，而是在握手的 `Finished` 阶段以 `bad record MAC` 失败且难以定位
- 对策：装载时对**签名私钥**执行自检（用本端标识与公共参数中的签名主公钥）；`kex-key` 无法反推校验，只能由导入来源保证，UI 与文档明确提示用途
- 只做结构解析与自检，不阻止"服务端认证场景下客户端不携带私钥"等合法配置

#### 3.6.3 IBC 信任池与默认拒绝

IBC 信任池是 IBC/IBSDH 套件的全局信任锚，管理方式与根证书一致（独立目录 + 目录扫描 + 列表管理 + 重载），**不进入实例配置**，客户端与服务端共用同一列表。数据结构、格式、接口与安全模型见 [3.2.5 IBC 信任池管理](#325-ibc-信任池管理)。

**默认拒绝语义（关键）：** 适配器在服务端设置 `ClientIBCSysParams`、客户端设置 `RootIBCSysParams` 时**总是注入非 nil 的信任池（即使为空池）**，从而压掉 gotlcp"本端公共参数退化为默认信任池"的行为。池中不存在对端 KGC 时 IBC 握手直接失败（`handshake_failure(40)`），与"没有根证书就验不过证书"对等。

- 两套信任锚互不替代：`RootCAs` / `ClientCAs` 只作用于证书套件；IBC 全局信任池只作用于 IBC 套件
- `InsecureSkipVerify=true` 会同时跳过 X.509 验证与 IBC 公共参数校验（仅测试用，UI 有警示文案）

#### 3.6.4 实例配置与混合协商

`TLCPConfig` **只新增一个字段**（信任池不进入实例配置）：

```go
type TLCPConfig struct {
    // ... 既有字段不变 ...
    // Keystore 证书身份密钥存储（X.509/SM2 证书），供 ECC/ECDHE 套件使用
    Keystore *KeyStoreConfig `yaml:"keystore,omitempty" json:"keystore,omitempty"`
    // IBCKeystore IBC(SM9) 身份密钥存储（标识 + KGC 公共参数 + 三把用户私钥），供 IBC/IBSDH 套件使用
    // 与 Keystore 相互独立、可只配其一，也可同时配置以实现同端口混合协商
    IBCKeystore *KeyStoreConfig `yaml:"ibc-keystore,omitempty" json:"ibcKeystore,omitempty"`
}
```

**配置示例：**

```yaml
instances:
  - name: tlcp-hybrid
    type: server
    protocol: tlcp
    tlcp:
      cipher-suites: [ECDHE_SM4_GCM_SM3, IBC_SM4_GCM_SM3, IBSDH_SM4_GCM_SM3]
      keystore:     { type: named, params: { name: default-tlcp } }        # 证书身份（ECC/ECDHE）
      ibc-keystore: { type: named, params: { name: default-ibc-server } }  # IBC 身份（IBC/IBSDH）
```

**装配矩阵（proxy adapter）：**

| 配置组合 | 写入 `tlcp.Config` | 可用套件 |
|---|---|---|
| 仅 `keystore` | `Certificates` / `GetClientCertificate` / `GetClientKECertificate` / `ClientCAs` / `RootCAs` | ECC、ECDHE |
| 仅 `ibc-keystore` | `IBCIdentity` + `ClientIBCSysParams` / `RootIBCSysParams` | IBC、IBSDH（服务端无证书亦可启动） |
| 两者同时 | 上述两组字段同时写入同一个 `tlcp.Config` | 四类套件同端口混合协商 |

- 实例有效性判定由「TLCP 证书不能为空」改为 `Keystore != nil || IBCKeystore != nil`
- `ibc-keystore` 为 `nil` 时，IBC/IBSDH 套件在库的套件选择阶段被跳过（本端无 IBC 能力）
- 客户端认证策略（`client-auth-type`）对证书身份与 IBC 身份同样生效；IBSDH 套件下库会强制要求客户端提供 IBC 身份

#### 3.6.5 密码套件与配置校验

- `config.TLCPCipherSuiteNames` 新增 4 个 IBC 套件名；新增判定函数 `IsTLCPIBCSuite`（IBC + IBSDH）与 `IsTLCPIBSDHSuite`（仅 IBSDH）
- `config.Validate` **只校验套件名是否合法**（TLCP 用 `TLCPCipherSuiteNames`、TLS 用 `TLSCipherSuiteNames`），**故意不校验"勾选了 IBC 套件但本端无 IBC 能力"**，该情况由 [3.6.6 能力诊断](#366-能力诊断与日志告警)记录日志
- 适配器中 `ParseCipherSuites` 的错误不再静默忽略，改为返回错误并记录日志
- UI：未配置 IBC 身份时，IBC/IBSDH 勾选框 `disabled`，提示"需先配置 IBC 身份 keystore"，并在 IBC 身份被清空时自动取消已勾选项
- UI：ECDHE 套件要求服务端认证客户端身份，服务端角色下客户端认证类型不是 `require-any-client-cert` / `require-and-verify-client-cert` 时 ECDHE 勾选框 `disabled` 并提示，认证类型改变后自动恢复可选或清理已勾选项，详见 [5.2 页面设计](#52-页面设计)

#### 3.6.6 能力诊断与日志告警

手工编辑 `config.yaml` 可能配置出"有 IBC 套件、无 IBC 身份"或"IBC 材料不满足套件需要"的组合。为避免静默失败，实例启动/重载时执行能力诊断并输出日志，**不阻断实例启动**（库会自动跳过不可用套件）。实现位置：`tlcpchan/proxy/ibc_diagnose.go`。

诊断共 8 类场景，同一场景按本端角色可能产生不同级别（共 11 条提示文案）：

| 场景 | 级别 | 日志要点 |
|---|---|---|
| 配了 IBC/IBSDH 套件但未配可用的 `ibc-keystore` | Error | 列出套件名，说明这些套件不会参与协商（诊断到此终止） |
| IBC 身份装载失败 | Error | 说明该身份不生效、IBC/IBSDH 套件不会参与协商（诊断到此终止） |
| 标识为空且勾选 IBSDH 套件 | Warn | 提示 IBSDH 握手时对端将报 `identity_need(205)` |
| 标识为空但未勾选 IBSDH 套件 | Warn | 提示对端可能无法确定本端标识 |
| 缺本端 KGC 公共参数（服务端，或已有签名私钥的客户端） | Error | 提示将报 `bad_ibcparam(203)`，需导入本端公共参数 |
| 缺本端 KGC 公共参数（无签名私钥的客户端） | Warn | 提示仅服务端单向认证场景可省略 |
| 缺签名私钥（服务端） | Error | 提示无法对 signed_params 签名，握手将失败 |
| 缺签名私钥（客户端） | Warn | 提示若服务端要求客户端认证将失败 |
| 服务端使用 `IBC_SM4_*` 但缺加密私钥（hid=0x03） | Error | 提示无法解密预主密钥，握手将失败 |
| 使用 `IBSDH_SM4_*` 但缺密钥交换私钥（hid=0x02） | Error | 提示这些套件不会参与协商 |
| 全局 IBC 信任池为空 | Error | 引导"请先在「IBC 信任池」中添加 KGC 公共参数" |

> 未勾选任何 IBC/IBSDH 套件时**不产生任何诊断**：按设计决策 D5，"配了 IBC 身份但未勾选 IBC 套件"是正常配置，无告警。

**套件 ↔ 材料能力矩阵：**

| 本端角色 | `IBC_SM4_*` | `IBSDH_SM4_*` |
|---|---|---|
| 服务端 | `params`、`sign-key`(0x01)、`enc-key`(0x03) | `params`、`sign-key`、`kex-key`(0x02) |
| 客户端 | `params`（双向认证另需 `sign-key`） | `params`、`sign-key`、`kex-key`（IBSDH 强制双向认证） |
| 共通 | 全局信任池中存在对端 KGC 参数；IBSDH 还要求 `identity` 非空 | 同左 |

**诊断函数签名（纯函数、仅产出诊断结论，由调用方输出日志）：**

```go
// diagnoseIBCSuites 诊断实例的 IBC 套件配置与本端 IBC 能力是否匹配，仅输出日志不阻断启动
// 参数：
//   - instanceName: 实例名称，用于日志定位
//   - isServer: 本端是否为服务端，决定必需材料集合
//   - suites: 已解析的 TLCP 密码套件数值列表
//   - ibcKS: 已加载的 IBC keystore，nil 表示未配置 ibc-keystore
//   - ident: 已装载的 IBC 身份，nil 表示身份不可用
//   - poolSize: 全局 IBC 信任池条目数，0 表示无可信 KGC
// 返回：[]ibcDiagnosis（Level 取 error / warn），由 logIBCDiagnoses 输出
func diagnoseIBCSuites(instanceName string, isServer bool, suites []uint16,
    ibcKS security.KeyStore, ident *tlcp.IBCIdentity, poolSize int) []ibcDiagnosis
```

**为什么 `Validate` 不硬失败**：`config.Validate` 在 `config.Load` 与 API 保存时都会执行，硬失败会导致手工编辑过的配置文件让整个服务无法启动（连 UI 管理入口都不可用），也会卡死"先配套件、后补材料"的正常流程。因此能力匹配全部交给上述日志诊断，UI 侧做前置预防。

#### 3.6.7 生成、导入与初始化预置

三条落地路径：

1. **初始化预置**：首次初始化生成一套内置测试 KGC（`districtName=tlcpchan.local`、`districtSerial=1`、有效期 10 年）
   - 公共参数写入 `ibcparams/tlcpchan-ibc-kgc.pem`（0644，入信任池）
   - 主密钥写入 `keystores/tlcpchan-ibc-kgc-master.key`（0600，不提供下载）
   - 由该 KGC 派生两份身份并登记为 `ibc-file` keystore：`default-ibc-server`（`server@tlcpchan.local`）与 `default-ibc-client`（`client@tlcpchan.local`）
   - `CheckInitialized` **明确不检查任何 IBC 产物**，防止老安装被重新初始化（见 [3.2.2 初始化流程](#322-初始化流程)）
2. **生成**（仅使用初始化内置的测试 KGC）
   - `POST /api/security/keystores/generate`，body `{name, type:"ibc", protected, identity, districtName, districtSerial}`
   - 入参 `identity` 为本端标识，本端公共参数取自 IBC 信任池中的目标 KGC
   - 一次性派生 `sign-key`(0x01)、`enc-key`(0x03)、`kex-key`(0x02)，落盘命名与导入路径完全一致
   - `POST /api/security/ibcparams/generate`，body `{districtName, districtSerial, years}`（默认 `tlcpchan.local` / 1 / 10），返回 `{filename, masterFile, param}`
   - 主密钥来源固定为 `keystores/tlcpchan-ibc-kgc-master.key`；生产环境应由外部 KGC 派生后导入
3. **导入**
   - `POST /api/security/keystores`（multipart，`loaderType=ibc-file`，字段 `identity` / `params` / `signKey` / `encKey` / `kexKey`）
   - `POST /api/security/keystores/:name/upload`（替换材料，同时放开类型门槛以支持 `ibc-file`）

> `POST /api/security/keystores/:name/export-csr` 对 `ibc` 类型返回 400（SM9 无 CSR 概念）。

#### 3.6.8 管理入口

**HTTP API：** IBC 信任池新增 6 个接口，keystore 接口扩展 IBC 分支，详见 [4.2 完整 API 路由表](#42-完整api路由表)。

**MCP 工具（4 个）：**

| 工具名 | 说明 |
|--------|------|
| `list_ibc_params` | 列出信任池中的 KGC 公共参数 |
| `add_ibc_params` | 添加 KGC 公共参数 |
| `remove_ibc_params` | 删除 KGC 公共参数 |
| `reload_ibc_params` | 重载 IBC 信任池 |

**CLI：** 新增 `tlcpchan-cli ibcparams` 命令组，与 `rootcert` 命令逐项对应：

```bash
tlcpchan-cli ibcparams list                    # 列出所有 KGC 公共参数
tlcpchan-cli ibcparams add [选项]              # 添加 KGC 公共参数
tlcpchan-cli ibcparams generate [选项]         # 生成测试 KGC 公共参数
tlcpchan-cli ibcparams download <filename>     # 下载 KGC 公共参数文件
tlcpchan-cli ibcparams delete <filename>       # 删除 KGC 公共参数
tlcpchan-cli ibcparams reload                  # 重载 IBC 信任池
```

`keystore` 命令组的列表与详情会展示 `type=ibc` 及 IBC 标识、KGC 区域/序号等元信息。

**Web UI：**
- 新增独立页面「IBC 信任池」（路由 `/ibcparams`，页面 `src/views/IBCParams.vue`），与「信任证书」并列
- keystore 创建 / 生成 / 更新支持 IBC 类型（含三把私钥的用途标注）
- 密钥管理页面（`src/views/KeyStores.vue`）按分类 Tab 展示：PKI（`tlcp`、`tls`）与 IBC（`ibc`），IBC Tab 展示标识、KGC 区域/序号、公共参数有效期与三把用户私钥状态，详见 [5.2 页面设计](#52-页面设计)
- 实例详情页的 TLCP 配置区在引用密钥库时补齐 IBC 标识、KGC 区域/序号、公共参数有效期与三把用户私钥状态
- 实例创建 / 编辑表单在 TLCP 配置卡片内按「PKI 配置 / IBC 配置 / TLCP 协议配置」三块排列（详见 [5.2 页面设计](#52-页面设计)）；未配置 IBC 身份时置灰 IBC/IBSDH 套件复选框并提示"需先配置 IBC 身份 keystore"

#### 3.6.9 兼容性与安全提示

| 风险 / 约束 | 说明 | 对策 |
|---|---|---|
| `ibc_parameter` 无签名保护 | 中间人可替换加密主公钥以解密预主密钥 | 信任池默认拒绝 + 带外预置 + UI 明确提示 |
| `kex-key` 的 hid 无法校验 | 装载期无法反推，误装要到 `Finished` 才以 `bad record MAC` 失败 | 签名私钥自检 + UI/文档明确用途 |
| `InsecureSkipVerify` | 客户端开启后同时跳过 X.509 验证与 IBC 公共参数校验 | UI 文案警示，默认关闭 |
| 会话重用 | 复用会话不重新校验公共参数有效期 | 文档说明 |
| 公共参数有效期 | 完整握手强制校验 `IBCSysParams.validity`，失败报 `unsupported_ibcparam(204)` | 文档与日志提示 |
| 老安装重初始化 | 若把 IBC 材料纳入 `CheckInitialized` 会重刷配置 | 明确不纳入（D10） |
| 信任池改动不生效 | 信任池对象在 `Reload()` 时被替换，运行中实例仍持有旧池 | 修改后重载相关实例（CLI/UI 提示） |

## 4. API设计

### 4.1 API服务器

使用Go标准库`net/http`实现RESTful API，无需第三方框架。

### 4.2 完整API路由表

系统共提供 44 个 RESTful API 接口（不含 `/api/version` 别名路由），分为 5 个主要类别：

#### 4.2.1 Instance API (12个)

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

#### 4.2.2 Security API (21个)

**Keystore API (9个):**

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/security/keystores | 获取 keystore 列表 | - | keystore 数组（`type=ibc` 时含 `ibc` 元信息） |
| POST | /api/security/keystores | 创建 keystore | keystore 配置（支持 multipart/form-data；`ibc-file` 用 identity/params/signKey/encKey/kexKey 字段） | 创建的 keystore 信息 |
| GET | /api/security/keystores/:name | 获取 keystore 详情 | - | keystore 详细信息（含 `ibc` 元信息） |
| PUT | /api/security/keystores/:name | 更新 keystore 参数 | params 对象 | 更新后的 keystore 信息 |
| POST | /api/security/keystores/:name/upload | 上传更新 keystore 证书和密钥 | multipart/form-data（证书：signCert/signKey/encCert/encKey；IBC：identity/params/signKey/encKey/kexKey） | 更新后的 keystore 信息 |
| GET | /api/security/keystores/:name/instances | 查询引用该 keystore 的实例 | - | 实例列表 |
| DELETE | /api/security/keystores/:name | 删除 keystore | - | 确认删除成功 |
| POST | /api/security/keystores/generate | 生成新 keystore | keystore 生成参数（`type=ibc` 时为 `{name, type, protected, identity, districtName, districtSerial}`） | 生成的 keystore 信息 |
| POST | /api/security/keystores/:name/export-csr | 导出 CSR | CSR 请求参数（`ibc` 类型返回 400） | CSR 文件（二进制流） |

**RootCert API (6个):**

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/security/rootcerts | 获取根证书列表 | - | 根证书数组（包含主题、颁发者、过期时间等） |
| POST | /api/security/rootcerts | 添加根证书 | multipart/form-data（filename + cert） | 添加的根证书信息 |
| GET | /api/security/rootcerts/:filename | 下载根证书（二进制流） | - | 文件流下载 |
| DELETE | /api/security/rootcerts/:filename | 删除根证书 | - | 确认删除成功 |
| POST | /api/security/rootcerts/generate | 生成根 CA 证书 | 根 CA 生成参数 | 生成的根 CA 信息 |
| POST | /api/security/rootcerts/reload | 重载所有根证书 | - | 确认重载成功 |

**IBCParams API (6个):**

| 方法 | 路径 | 描述 | 请求体 | 响应体 |
|------|------|------|--------|--------|
| GET | /api/security/ibcparams | 获取 IBC 信任池列表 | - | KGC 公共参数元信息数组 |
| POST | /api/security/ibcparams | 添加 KGC 公共参数 | multipart/form-data（filename + params，扩展名不在白名单时规范化为 `.pem`） | 添加的条目元信息 |
| POST | /api/security/ibcparams/generate | 生成测试 KGC 公共参数 | `{districtName, districtSerial, years}`（默认 `tlcpchan.local` / 1 / 10） | `{filename, masterFile, param}` |
| GET | /api/security/ibcparams/:filename | 下载 KGC 公共参数（PEM） | - | 文件流下载 |
| DELETE | /api/security/ibcparams/:filename | 删除 KGC 公共参数 | - | 确认删除成功 |
| POST | /api/security/ibcparams/reload | 重载 IBC 信任池 | - | 确认重载成功 |

#### 4.2.3 System API (3个)

| 方法 | 路径 | 描述 | 响应体 |
|------|------|------|--------|
| GET | /api/system/info | 获取系统信息 | 系统信息对象（操作系统、架构、内存、CPU、Goroutine 等）|
| GET | /api/system/health | 系统健康检查 | 状态和版本信息 |
| GET | /api/system/version | 版本信息 | 版本号 |

> 另有 `GET /api/version`（与 `/api/system/version` 等价）与 `GET /version` 两个别名路由，不计入接口总数。

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

**总计：44 个 API 接口**

| 分类 | 数量 |
|------|------|
| Instance API | 12 |
| Security API - Keystore | 9 |
| Security API - RootCert | 6 |
| Security API - IBCParams | 6 |
| System API | 3 |
| Config API | 4 |
| Logs API | 4 |
| **合计** | **44** |

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
| 密钥管理 | /keystores | 证书类密钥（PKI）与 IBC(SM9) 身份密钥管理，按分类 Tab 展示 |
| 信任根证书 | /trusted | 信任根证书列表与管理 |
| IBC 信任池 | /ibcparams | KGC 公共参数列表、上传、下载、删除、重载、生成测试 KGC |
| 日志查看 | /logs | 日志实时查看 |
| 系统设置 | /settings | 系统配置 |

**密钥管理页面（`src/views/KeyStores.vue`）的分类 Tab：**

页面通过两个分类 Tab 展示 keystore，避免证书类密钥与 IBC(SM9) 标识身份密钥混在同一张表中：

| Tab（query 取值） | 标签 | 包含的 keystore 类型 | 顶部操作 |
|-------------------|------|----------------------|----------|
| pki | PKI | `tlcp`（SM2 双证书）、`tls`（RSA/ECC 证书） | 生成密钥、导入密钥 |
| ibc | IBC | `ibc`（SM9 标识身份，不使用 X.509 证书） | 生成 IBC 身份、导入 IBC 身份 |

- Tab 状态与 URL query 同步（`/keystores?tab=pki`、`/keystores?tab=ibc`）；缺省或非法取值回退到 `pki`，刷新、分享链接与浏览器前进/后退均保持当前分类。
- PKI Tab 列：名称、类型、签名证书/密钥、加密证书/密钥（列表中存在 TLCP 密钥时才展示）、创建时间；操作：详情、导出 CSR、更新证书、删除。
- IBC Tab 列：名称、标识、KGC 区域、KGC 序号、公共参数状态、公共参数有效期、三把用户私钥状态（hid=0x01 签名 / hid=0x03 加密 / hid=0x02 密钥交换）、创建时间；操作：详情、替换材料、删除。IBC 不涉及 X.509，因此不提供导出 CSR 与更新证书。
- 顶部按钮通过 `?type=tlcp|tls|ibc` 预设目标页面的密钥类型；生成、导入、替换材料完成后返回列表页时携带对应分类（`keystoreListLocation`）。
- 分类判定与过滤集中在 `tlcpchan-ui/src/constants/keystoreTab.ts`；后端列表接口 `GET /api/security/keystores` 已返回 IBC 只读元信息（`ibc` 字段），页面无需额外请求。
- 时间展示统一走 `tlcpchan-ui/src/constants/datetime.ts`：IBC 未提供 KGC 公共参数时后端返回 Go 零值时间 `0001-01-01T00:00:00Z`，页面按“不限”展示。

**实例管理页面（`src/views/Instances.vue`）与实例详情页的 IBC 展示：**

- 实例列表只展示名称、类型、协议、监听地址、目标地址、状态与操作（进入管理），不展示认证类型、IBC 身份与 IBC 套件；相关细节统一在实例详情页查看。
- 实例详情页的 TLCP 配置区（`src/components/ProtocolConfigDetail.vue`）在配置层信息之外补齐 IBC 身份元信息：
  - `ibcKeystore.type = named` 时按密钥库名称额外请求 `GET /api/security/keystores/{name}`，展示标识、KGC 区域/序号、公共参数状态与有效期、三把用户私钥状态（hid=0x01 签名 / hid=0x03 加密 / hid=0x02 密钥交换）；密钥库已被删除或无 IBC 元信息时给出明确提示。
  - `ibcKeystore.type = ibc-file` 时配置中只有文件路径，展示各材料文件路径并提示“标识、KGC 区域与私钥用途以文件内容为准”。

**协议配置的三段式分组（PKI 配置 / IBC 配置 / 协议配置）：**

TLCP 配置涉及“证书身份、IBC 身份、协议参数”三类互相独立的内容，创建实例页、编辑实例页与实例详情页统一按三块依次排列，块内字段归属固定如下：

| 分组 | 归属字段 | 出现范围 |
|------|----------|----------|
| PKI 配置 | 证书 keystore（`named` 引用名称；`file` 的签名证书/密钥、加密证书/密钥路径） | TLCP 与 TLS |
| IBC 配置 | IBC 身份 keystore 类型、引用名称或各材料文件路径、IBC 身份只读元信息（标识、KGC 区域/序号、公共参数状态与有效期、三把用户私钥状态） | 仅 TLCP |
| TLCP 协议配置 | 客户端认证类型、最低/最高版本、密码套件（含 IBC/IBSDH 套件）、握手重用、跳过证书验证 | TLCP（TLS 侧对应「TLS 协议配置」，另含会话票据） |

- 创建实例页（`src/views/CreateInstance.vue`）与编辑实例页（`src/views/EditInstance.vue`）在 TLCP 配置卡片内用 `el-divider`（`content-position="left"`）分隔三块，`IBC 配置` 块保留“证书身份与 IBC 身份相互独立”的说明提示。
- 实例详情页（`src/components/ProtocolConfigDetail.vue`）用三个带左侧色条标题的区块承载同一分组，每块一个 `el-descriptions`；证书 keystore 或 IBC 身份未配置时该块展示“未配置”，避免出现空白表格。
- **IBC/IBSDH 套件归于协议配置块**（它是 `cipherSuites` 的取值，属于协议参数），不放在 IBC 配置块。
- 实例列表页不展示这三类配置，只保留进入管理入口。

**IBC 信任池页面（`src/views/IBCParams.vue`）：**
- 表格列：文件名、KGC 区域、序号、生效时间、失效时间、颁发者标识、签名主公钥指纹、加密主公钥指纹
- 操作：上传 KGC 公共参数、下载、删除、重载信任池、生成测试 KGC（主密钥保存在 `keystores/`，不提供下载）
- 页面提示：信任池修改后需重载相关实例才会生效；信任池中不存在对端 KGC 时 IBC 握手一定失败

### 5.3 实例表单的认证与 keystore 联动规则

实例创建/编辑表单需要根据实例角色与客户端认证类型决定 keystore 是否必填，并限制依赖证书身份的密码套件，避免用户提交出无法正常握手的组合。

**角色划分：**

| 实例类型 | 角色 | 证书身份用途 |
|----------|------|--------------|
| server / http-server | 服务端 | 向客户端证明自身身份 |
| client / http-client | 客户端 | 双向认证时向服务端证明自身身份 |

**keystore 必填规则：**

| 场景 | 证书 keystore | 说明 |
|------|---------------|------|
| 服务端代理，协议为 tlcp / tls | 必填 | 服务端必须提供自身证书 |
| 服务端代理，协议为 auto | 必填（tlcp 与 tls 至少其一） | 允许混合协商 |
| 客户端代理，认证类型为 no-client-cert / request-client-cert | 可选 | 单向认证，本端无需向服务端出示证书 |
| 客户端代理，认证类型为 require-any-client-cert / verify-client-cert-if-given / require-and-verify-client-cert | 必填 | 双向认证，本端需要出示客户端证书 |
| 已配置 IBC(SM9) 身份 keystore | 可选 | IBC 与证书身份相互独立，可只配其一 |

**密码套件与身份的依赖：**

| 套件前缀 | 依赖 | 未满足时的交互 |
|----------|------|----------------|
| ECC_SM4_* | 无（客户端侧仅验证服务端身份） | — |
| ECDHE_SM4_* | ① TLCP 证书 keystore；② 服务端角色的客户端认证类型必须是 require-any-client-cert 或 require-and-verify-client-cert | 置灰不可选：缺 keystore 时提示“需先配置 TLCP 证书 keystore”，服务端认证类型不合格时提示“ECDHE 套件要求认证客户端身份，请将客户端认证类型设为「要求证书」或「要求并验证」”，并在条件不再满足时自动取消已勾选项 |
| IBC_SM4_* / IBSDH_SM4_* | IBC 身份 keystore | 置灰不可选，提示“需先配置 IBC 身份 keystore”，并在 IBC 身份被清空时自动取消已勾选项 |

**ECDHE 套件的客户端认证类型约束：**

ECDHE 套件要求对客户端做身份认证，因此服务端必须“要求”客户端出示证书：`require-any-client-cert`（要求证书）与 `require-and-verify-client-cert`（要求并验证）满足要求；`no-client-cert`、`request-client-cert`（仅请求）、`verify-client-cert-if-given`（提供才验证，不强制出示）均不满足。

- 该约束**只作用于服务端角色**（`server` / `http-server`）：客户端代理的 `client-auth-type` 描述的是本端是否向服务端出示证书，对端是否要求本端证书不由本端配置决定，因此客户端角色下 ECDHE 只受证书 keystore 约束。
- 客户端角色（`client` / `http-client`）的 ECDHE 可选条件即“已配置完整的 TLCP 证书 keystore”（客户端密钥），与证书 keystore 的必填规则衔接：单向认证下 keystore 可不配，此时 ECDHE 置灰。
- 判定函数 `isClientAuthAllowedForECDHE` 位于 `tlcpchan-ui/src/constants/cipherSuite.ts`，由创建与编辑表单共用；两个表单都以 `watch` 监听“证书 keystore 与认证类型”两个条件，条件不再满足时过滤掉已勾选的 ECDHE 套件（置灰中的复选框无法被点击取消，必须主动清理）。
- 编辑表单在配置加载完成后额外执行一次兜底清理：历史配置可能残留“已勾选 ECDHE 但条件不满足”的组合，该组合在界面上会呈现为置灰且勾选，用户无法主动取消。

**客户端认证类型的展示文案：**

前端统一以“中文（英文）”呈现，取值与后端 `ValidClientAuthValues` 一致：

| 取值 | 显示文案 |
|------|----------|
| no-client-cert | 不要求证书（no-client-cert） |
| request-client-cert | 请求证书（request-client-cert） |
| require-any-client-cert | 要求证书（require-any-client-cert） |
| verify-client-cert-if-given | 提供则验证（verify-client-cert-if-given） |
| require-and-verify-client-cert | 要求并验证（require-and-verify-client-cert） |

实现位置：文案映射与单向认证判定集中在 `tlcpchan-ui/src/constants/clientAuth.ts`，由实例列表、实例详情与创建/编辑表单共用。

### 5.4 UI服务架构

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
Description=TLCP Channel Proxy Service
After=network.target

[Service]
Type=simple
User=tlcpchan
WorkingDirectory=/etc/tlcpchan
ExecStart=/etc/tlcpchan/tlcpchan
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

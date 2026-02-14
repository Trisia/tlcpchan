# TLCPChan 配置示例

本文档提供各种场景的配置示例，帮助快速上手。

## 目录

- [基础TLCP代理](#基础tlcp代理)
- [基础TLS代理](#基础tls代理)
- [双向认证代理](#双向认证代理)
- [自动协议检测](#自动协议检测)
- [HTTP代理](#http代理)
- [负载均衡场景](#负载均衡场景)

---

## 基础TLCP代理

将 TLCP 客户端请求代理到后端 HTTP 服务。

```yaml
server:
  api:
    address: ":30080"
  ui:
    enabled: true
    address: ":30000"
  log:
    level: "info"
    file: "./logs/tlcpchan.log"

instances:
  - name: "tlcp-basic"
    type: "server"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
    tlcp:
      min_version: "1.1"
      max_version: "1.1"
      cipher_suites:
        - "ECDHE_SM4_GCM_SM3"
```

### 证书文件要求

```
certs/tlcp/
├── server-sm2.crt    # SM2 签名证书
├── server-sm2.key    # SM2 私钥
└── server-sm2.enc    # SM2 加密证书（可选）
```

---

## 基础TLS代理

将 TLS 客户端请求代理到后端服务。

```yaml
instances:
  - name: "tls-basic"
    type: "server"
    protocol: "tls"
    auth: "one-way"
    listen: ":8443"
    target: "127.0.0.1:8080"
    enabled: true
    certificates:
      tls:
        cert: "server-rsa"
        key: "server-rsa"
    tls:
      min_version: "1.2"
      max_version: "1.3"
```

### 证书文件要求

```
certs/tls/
├── server-rsa.crt    # RSA 或 ECDSA 证书
└── server-rsa.key    # 私钥文件
```

---

## 双向认证代理

客户端和服务器双向证书认证（mTLS）。

### TLCP 双向认证

```yaml
instances:
  - name: "tlcp-mtls"
    type: "server"
    protocol: "tlcp"
    auth: "mutual"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
    client_ca:
      - "ca-sm2"
    tlcp:
      min_version: "1.1"
      cipher_suites:
        - "ECDHE_SM4_GCM_SM3"
        - "ECC_SM4_GCM_SM3"
```

### TLS 双向认证

```yaml
instances:
  - name: "tls-mtls"
    type: "server"
    protocol: "tls"
    auth: "mutual"
    listen: ":8443"
    target: "127.0.0.1:8080"
    enabled: true
    certificates:
      tls:
        cert: "server-rsa"
        key: "server-rsa"
    client_ca:
      - "ca-rsa"
    tls:
      min_version: "1.2"
```

---

## 自动协议检测

自动识别客户端使用 TLCP 还是 TLS 协议。

```yaml
instances:
  - name: "auto-protocol"
    type: "server"
    protocol: "auto"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
      tls:
        cert: "server-rsa"
        key: "server-rsa"
    tlcp:
      min_version: "1.1"
    tls:
      min_version: "1.2"
```

工作原理：
1. 监听同一端口
2. 根据 ClientHello 自动识别协议
3. 使用对应协议的证书完成握手

---

## HTTP代理

提供 HTTP/HTTPS 代理功能，支持请求/响应头修改。

```yaml
instances:
  - name: "http-proxy"
    type: "http-server"
    protocol: "tlcp"
    listen: ":8443"
    target: "127.0.0.1:8000"
    enabled: true
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
    http:
      request_headers:
        add:
          X-Proxy: "tlcpchan"
          X-Real-IP: "$remote_addr"
          X-Forwarded-Proto: "https"
        remove:
          - "X-Forwarded-For"
      response_headers:
        add:
          Server: "TLCPChan/1.0"
        remove:
          - "X-Powered-By"
```

### 可用变量

| 变量 | 说明 |
|------|------|
| `$remote_addr` | 客户端 IP 地址 |
| `$remote_port` | 客户端端口 |
| `$server_addr` | 服务器地址 |
| `$server_port` | 服务器端口 |
| `$scheme` | 协议 (http/https) |
| `$host` | 请求主机名 |

---

## 负载均衡场景

### 多实例配置

```yaml
instances:
  # TLCP 入口
  - name: "tlcp-frontend"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8001"
    enabled: true
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"

  # TLS 入口
  - name: "tls-frontend"
    type: "server"
    protocol: "tls"
    listen: ":8443"
    target: "127.0.0.1:8002"
    enabled: true
    certificates:
      tls:
        cert: "server-rsa"
        key: "server-rsa"

  # 后端客户端代理（访问外部服务）
  - name: "backend-api"
    type: "client"
    protocol: "tls"
    listen: ":9001"
    target: "api.example.com:443"
    enabled: true
    server_ca:
      - "ca-rsa"

  - name: "backend-auth"
    type: "client"
    protocol: "tls"
    listen: ":9002"
    target: "auth.example.com:443"
    enabled: true
    server_ca:
      - "ca-rsa"
```

### 生产环境推荐配置

```yaml
server:
  api:
    address: ":30080"
  ui:
    enabled: false
  log:
    level: "warn"
    file: "./logs/tlcpchan.log"
    max_size: 100
    max_backups: 10
    max_age: 30
    compress: true

instances:
  - name: "production-tlcp"
    type: "server"
    protocol: "tlcp"
    auth: "mutual"
    listen: ":443"
    target: "10.0.0.10:8080"
    enabled: true
    certificates:
      tlcp:
        cert: "prod-server-sm2"
        key: "prod-server-sm2"
    client_ca:
      - "prod-ca-sm2"
    tlcp:
      min_version: "1.1"
      cipher_suites:
        - "ECDHE_SM4_GCM_SM3"
```

---

## 配置项说明

### 全局配置 (server)

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `api.address` | 管理 API 监听地址 | `:30080` |
| `ui.enabled` | 是否启用 Web UI | `true` |
| `ui.address` | UI 监听地址 | `:30000` |
| `ui.path` | UI 静态文件路径 | `./ui` |
| `log.level` | 日志级别 | `info` |
| `log.file` | 日志文件路径 | `./logs/tlcpchan.log` |
| `log.max_size` | 单文件最大 MB | `100` |
| `log.max_backups` | 保留文件数 | `5` |
| `log.max_age` | 保留天数 | `30` |
| `log.compress` | 是否压缩 | `true` |

### 实例配置 (instances)

| 字段 | 说明 | 可选值 |
|------|------|--------|
| `name` | 实例名称 | 唯一标识 |
| `type` | 代理类型 | `server`, `client`, `http-server` |
| `protocol` | 协议类型 | `tlcp`, `tls`, `auto` |
| `auth` | 认证方式 | `none`, `one-way`, `mutual` |
| `listen` | 监听地址 | `:port` 或 `ip:port` |
| `target` | 目标地址 | `host:port` |
| `enabled` | 是否启用 | `true`, `false` |

### TLCP 配置 (tlcp)

| 字段 | 说明 | 可选值 |
|------|------|--------|
| `min_version` | 最低版本 | `1.1` |
| `max_version` | 最高版本 | `1.1` |
| `cipher_suites` | 密码套件 | `ECDHE_SM4_GCM_SM3`, `ECC_SM4_GCM_SM3` |

### TLS 配置 (tls)

| 字段 | 说明 | 可选值 |
|------|------|--------|
| `min_version` | 最低版本 | `1.0`, `1.1`, `1.2`, `1.3` |
| `max_version` | 最高版本 | `1.0`, `1.1`, `1.2`, `1.3` |
| `cipher_suites` | 密码套件 | 参考 Go TLS 文档 |
| `insecure_skip_verify` | 跳过证书验证 | `true`, `false` |

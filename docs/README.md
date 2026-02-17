# TLCP Channel 使用指南

## 项目简介

TLCP Channel 是一款 TLCP/TLS 协议代理工具，支持双协议并行工作。它能够：

- 将 TLCP/TLS 加密流量转换为明文 TCP 流量（服务端代理）
- 将明文 TCP 流量转换为 TLCP/TLS 加密流量（客户端代理）
- 自动检测客户端协议类型（TLCP 或 TLS）

本工具基于 [gotlcp](https://github.com/Trisia/gotlcp) 库实现 TLCP 协议支持，使用国密 SM2 算法进行证书签名和密钥交换。

## 功能特性

### 协议支持

- **TLCP 1.1**：完整支持国密传输层密码协议
- **TLS 1.0/1.1/1.2/1.3**：标准 TLS 协议支持
- **自动检测**：同一端口自动识别 TLCP/TLS 协议

### 代理模式

| 模式 | 说明 |
|------|------|
| 服务端代理 | TLCP/TLS → TCP，解密后转发到后端服务 |
| 客户端代理 | TCP → TLCP/TLS，加密后连接目标服务 |
| HTTP 服务端 | HTTP/HTTPS 代理，支持头部修改 |
| HTTP 客户端 | 明文 HTTP 转发为加密 HTTPS |

### 认证方式

- **无认证（none）**：不验证对端证书
- **单向认证（one-way）**：仅验证服务端证书
- **双向认证（mutual）**：双向证书验证

### 管理功能

- RESTful API 接口
- Web UI 管理界面
- 命令行工具（tlcpchan-cli）
- 配置热重载
- 证书热更新
- 实时流量统计

## 安装部署

详细的编译、部署和运行说明请参考 [编译部署运行指南](./build-deploy.md)。

## 基本配置

配置文件位于工作目录的 `config/config.yaml`：

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
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    keystore: "my-server-keystore"
```

## 配置说明

### 配置文件结构

```yaml
server:           # 全局服务配置
  api:            # API 服务配置
  ui:             # Web UI 配置
  log:            # 日志配置

instances:        # 代理实例列表
  - name: ""      # 实例名称
    type: ""      # 实例类型
    # ...
```

### 实例配置详解

#### 基本字段

| 字段 | 说明 | 可选值 |
|------|------|--------|
| `name` | 实例名称，唯一标识 | 字符串 |
| `type` | 代理类型 | `server`, `client`, `http-server`, `http-client` |
| `protocol` | 协议类型 | `tlcp`, `tls`, `auto` |
| `auth` | 认证方式 | `none`, `one-way`, `mutual` |
| `listen` | 监听地址 | `:port` 或 `ip:port` |
| `target` | 目标地址 | `host:port` |
| `enabled` | 是否启用 | `true`, `false` |

#### TLCP 配置

```yaml
tlcp:
  min_version: "1.1"      # 最低版本
  max_version: "1.1"      # 最高版本
  cipher_suites:          # 密码套件
    - "ECDHE_SM4_GCM_SM3"
    - "ECC_SM4_GCM_SM3"
  session_tickets: true   # 会话票据
```

#### TLS 配置

```yaml
tls:
  min_version: "1.2"
  max_version: "1.3"
  cipher_suites:
    - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
```

### 安全参数配置

```yaml
keystore: "my-server-keystore"  # 通过名称引用 keystore

# 或直接配置文件路径
keystore:
  sign-cert: "/path/to/sign.crt"
  sign-key: "/path/to/sign.key"
  enc-cert: "/path/to/enc.crt"  # TLCP 可选
  enc-key: "/path/to/enc.key"    # TLCP 可选

root-certs: ["ca1", "ca2"]  # 根证书名称列表
```

## 使用示例

### TLCP 服务端代理

将 TLCP 客户端请求代理到后端 HTTP 服务：

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    keystore: "my-server-keystore"
    tlcp:
      min_version: "1.1"
      cipher_suites:
        - "ECDHE_SM4_GCM_SM3"
```

### TLS 服务端代理

```yaml
instances:
  - name: "tls-server"
    type: "server"
    protocol: "tls"
    auth: "one-way"
    listen: ":8443"
    target: "127.0.0.1:8080"
    enabled: true
    keystore: "my-tls-keystore"
    tls:
      min_version: "1.2"
```

### 客户端代理

将明文请求加密后转发到 TLCP/TLS 服务：

```yaml
instances:
  - name: "tlcp-client"
    type: "client"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":9000"
    target: "tlcp-server.example.com:443"
    enabled: true
    root-certs: ["ca-sm2"]
```

### HTTP 代理

```yaml
instances:
  - name: "http-proxy"
    type: "http-server"
    protocol: "tlcp"
    listen: ":8443"
    target: "127.0.0.1:8000"
    enabled: true
    keystore: "my-server-keystore"
    http:
      request_headers:
        add:
          X-Proxy: "tlcpchan"
          X-Real-IP: "$remote_addr"
        remove:
          - "X-Forwarded-For"
      response_headers:
        add:
          Server: "TLCPChan/1.0"
```

### 协议自动检测

同一端口支持 TLCP 和 TLS 客户端：

```yaml
instances:
  - name: "auto-protocol"
    type: "server"
    protocol: "auto"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    enabled: true
    keystore: "my-server-keystore"
```

## API 使用

### 基础 URL

```
http://localhost:30080/api
```

### 常用接口

```bash
# 获取实例列表
curl http://localhost:30080/api/instances

# 创建实例
curl -X POST http://localhost:30080/api/instances \
  -H "Content-Type: application/json" \
  -d '{"name":"test","type":"server","protocol":"tlcp","listen":":443","target":"127.0.0.1:8080"}'

# 启动实例
curl -X POST http://localhost:30080/api/instances/test/start

# 健康检查
curl http://localhost:30080/api/system/health
```

完整 API 文档请参考 [API 使用指南](./api-usage.md) 和 [API 接口文档](./api.md)。

## CLI 使用

### 安装 CLI

详细的 CLI 安装和使用说明请参考 [编译部署运行指南](./build-deploy.md)。

### 常用命令

```bash
# 查看版本
tlcpchan-cli version

# 实例管理
tlcpchan-cli instance list
tlcpchan-cli instance show <name>
tlcpchan-cli instance start <name>
tlcpchan-cli instance stop <name>
tlcpchan-cli instance create -f config.json
tlcpchan-cli instance delete <name>

# 证书管理
tlcpchan-cli cert list
tlcpchan-cli cert show <name>
tlcpchan-cli cert reload

# 配置管理
tlcpchan-cli config show
tlcpchan-cli config reload

# 系统信息
tlcpchan-cli system info
tlcpchan-cli system health
```

### 全局选项

```bash
tlcpchan-cli -api http://localhost:30080 -output json instance list
```

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-api` | API 服务地址 | `http://localhost:30080` |
| `-output` | 输出格式 | `table` |

## 常见问题

### Q: 首次启动报错找不到证书？

A: 首次启动时会自动生成测试证书。确保程序有写入工作目录的权限。

### Q: TLCP 连接握手失败？

A: 检查以下几点：
1. 证书是否为 SM2 算法签名
2. 密码套件是否正确配置
3. 客户端是否支持国密算法

### Q: 如何查看实例运行状态？

```bash
tlcpchan-cli instance show <name>
curl http://localhost:30080/api/instances/<name>
```

### Q: 如何更新证书而不重启服务？

```bash
# 替换证书文件后执行
tlcpchan-cli cert reload

# 或通过 API
curl -X POST http://localhost:30080/api/certificates/reload
```

### Q: 如何启用双向认证？

```yaml
instances:
  - name: "mtls-server"
    type: "server"
    protocol: "tlcp"
    auth: "mutual"          # 设置为 mutual
    # ...
    client_ca:              # 配置客户端 CA
      - "client-ca-sm2"
```

## 故障排查

### 日志查看

```bash
# 查看日志文件
tail -f /etc/tlcpchan/logs/tlcpchan.log

# 通过 API 查看实例日志
curl "http://localhost:30080/api/instances/<name>/logs?lines=100"

# 通过 CLI 查看
tlcpchan-cli instance logs <name>
```

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| 端口已被占用 | 监听端口被其他进程使用 | 更换端口或停止占用进程 |
| 证书加载失败 | 证书文件不存在或格式错误 | 检查证书路径和格式 |
| 连接被拒绝 | 目标服务未启动 | 确认目标服务运行正常 |
| 证书验证失败 | CA 证书不匹配 | 检查 client_ca/server_ca 配置 |

### 调试模式

```yaml
server:
  log:
    level: "debug"  # 开启调试日志
```

或通过命令行（如支持）：

```bash
tlcpchan -log-level debug
```

### 性能问题排查

1. 查看系统资源使用：
```bash
curl http://localhost:30080/api/system/info
```

2. 查看实例统计：
```bash
curl http://localhost:30080/api/instances/<name>/stats
```

3. 检查连接数：
```bash
ss -tnp | grep tlcpchan
```

# TLCP Channel API 接口文档

## 1. 概述

### 1.1 基本信息

- 基础URL：`http://localhost:8080/api/v1`
- 内容类型：`application/json`
- 字符编码：`UTF-8`

### 1.2 响应格式

**成功响应**：直接返回数据对象，无额外包装

```json
{
  "instances": [],
  "total": 0
}
```

**错误响应**：HTTP状态码 + 纯文本错误信息

```
HTTP 400 Bad Request
Content-Type: text/plain

实例名称不能为空
```

### 1.3 HTTP状态码

| 状态码 | 描述 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如已存在） |
| 500 | 服务器内部错误 |

## 2. 实例管理接口

### 2.1 获取实例列表

**请求**

```
GET /api/v1/instances
```

**响应**

```json
{
  "instances": [
    {
      "name": "tlcp-server",
      "type": "server",
      "protocol": "tlcp",
      "listen": ":443",
      "target": "127.0.0.1:8080",
      "status": "running",
      "enabled": true,
      "uptime": 3600
    }
  ],
  "total": 1
}
```

### 2.2 创建实例

**请求**

```
POST /api/v1/instances
Content-Type: application/json

{
  "name": "tlcp-server",
  "type": "server",
  "protocol": "tlcp",
  "auth": "mutual",
  "listen": ":443",
  "target": "127.0.0.1:8080",
  "enabled": true,
  "certificates": {
    "tlcp": {
      "cert": "server-sm2",
      "key": "server-sm2"
    }
  },
  "client_ca": ["ca-sm2"],
  "tlcp": {
    "min_version": "1.1",
    "max_version": "1.1",
    "cipher_suites": ["ECC_SM4_GCM_SM3", "ECDHE_SM4_GCM_SM3"],
    "session_tickets": true
  }
}
```

**响应**

```
HTTP 201 Created
```

```json
{
  "name": "tlcp-server",
  "status": "created"
}
```

### 2.3 获取实例详情

**请求**

```
GET /api/v1/instances/:name
```

**响应**

```json
{
  "name": "tlcp-server",
  "type": "server",
  "protocol": "tlcp",
  "auth": "mutual",
  "listen": ":443",
  "target": "127.0.0.1:8080",
  "status": "running",
  "enabled": true,
  "uptime": 3600,
  "config": {
    "certificates": {
      "tlcp": {"cert": "server-sm2", "key": "server-sm2"}
    },
    "client_ca": ["ca-sm2"],
    "tlcp": {
      "min_version": "1.1",
      "cipher_suites": ["ECC_SM4_GCM_SM3"]
    }
  },
  "stats": {
    "connections_total": 1000,
    "connections_active": 10,
    "bytes_received": 1048576,
    "bytes_sent": 2097152
  }
}
```

### 2.4 更新实例配置

**请求**

```
PUT /api/v1/instances/:name
Content-Type: application/json

{
  "target": "127.0.0.1:9090",
  "enabled": true
}
```

**响应**

```json
{
  "name": "tlcp-server",
  "status": "updated"
}
```

### 2.5 删除实例

**请求**

```
DELETE /api/v1/instances/:name
```

**响应**

```
HTTP 204 No Content
```

### 2.6 启动实例

**请求**

```
POST /api/v1/instances/:name/start
```

**响应**

```json
{
  "name": "tlcp-server",
  "status": "running"
}
```

### 2.7 停止实例

**请求**

```
POST /api/v1/instances/:name/stop
```

**响应**

```json
{
  "name": "tlcp-server",
  "status": "stopped"
}
```

### 2.8 重载实例

**请求**

```
POST /api/v1/instances/:name/reload
```

**响应**

```json
{
  "name": "tlcp-server",
  "status": "reloaded"
}
```

### 2.9 获取实例统计

**请求**

```
GET /api/v1/instances/:name/stats
```

**查询参数**

| 参数 | 类型 | 描述 |
|------|------|------|
| period | string | 统计周期：1m/5m/15m/1h/1d |

**响应**

```json
{
  "connections_total": 1000,
  "connections_active": 10,
  "bytes_received": 1048576,
  "bytes_sent": 2097152,
  "requests_total": 500,
  "errors": 2,
  "latency_avg_ms": 5.2
}
```

### 2.10 获取实例日志

**请求**

```
GET /api/v1/instances/:name/logs
```

**查询参数**

| 参数 | 类型 | 描述 |
|------|------|------|
| lines | integer | 返回行数，默认100，最大1000 |
| level | string | 日志级别：debug/info/warn/error |

**响应**

```json
{
  "logs": [
    {
      "time": "2024-01-01T00:00:00Z",
      "level": "info",
      "message": "connection accepted from 192.168.1.1:12345"
    }
  ]
}
```

## 3. 配置管理接口

### 3.1 获取当前配置

**请求**

```
GET /api/v1/config
```

**响应**

```json
{
  "server": {
    "api": {"address": ":8080"},
    "ui": {"enabled": true, "address": ":3000"},
    "log": {"level": "info", "file": "./logs/tlcpchan.log"}
  },
  "instances": []
}
```

### 3.2 重载配置

**请求**

```
POST /api/v1/config/reload
```

**响应**

```json
{
  "reloaded": true,
  "changes": {
    "instances": {
      "added": [],
      "removed": ["old-instance"],
      "modified": ["tlcp-server"]
    }
  }
}
```

### 3.3 验证配置

**请求**

```
POST /api/v1/config/validate
Content-Type: application/json

{
  "instances": [
    {
      "name": "test-instance",
      "type": "server",
      "protocol": "tlcp",
      "listen": ":443",
      "target": "127.0.0.1:8080"
    }
  ]
}
```

**响应**

```json
{
  "valid": true
}
```

**错误响应**

```
HTTP 400 Bad Request

配置验证失败：实例 test-instance 监听地址格式错误
```

## 4. 证书管理接口

### 4.1 获取证书列表

**请求**

```
GET /api/v1/certificates
```

**响应**

```json
{
  "certificates": [
    {
      "name": "server-sm2",
      "type": "tlcp",
      "subject": "CN=localhost",
      "issuer": "CN=TLCPChan Root CA",
      "not_before": "2024-01-01T00:00:00Z",
      "not_after": "2025-01-01T00:00:00Z",
      "public_key_algorithm": "SM2",
      "signature_algorithm": "SM2-SM3"
    }
  ],
  "total": 1
}
```

### 4.2 获取证书详情

**请求**

```
GET /api/v1/certificates/:name
```

**响应**

```json
{
  "name": "server-sm2",
  "type": "tlcp",
  "subject": "CN=localhost",
  "issuer": "CN=TLCPChan Root CA",
  "not_before": "2024-01-01T00:00:00Z",
  "not_after": "2025-01-01T00:00:00Z",
  "is_ca": false,
  "serial_number": "01",
  "dns_names": ["localhost", "*.example.com"],
  "ip_addresses": ["127.0.0.1"],
  "public_key_algorithm": "SM2",
  "signature_algorithm": "SM2-SM3"
}
```

### 4.3 热更新证书

**请求**

```
POST /api/v1/certificates/reload
```

**响应**

```json
{
  "reloaded": true,
  "updated": ["server-sm2", "client-sm2"]
}
```

### 4.4 生成证书

**请求**

```
POST /api/v1/certificates/generate
Content-Type: application/json

{
  "type": "tlcp",
  "name": "new-server",
  "common_name": "example.com",
  "dns_names": ["example.com", "*.example.com"],
  "ip_addresses": ["192.168.1.1"],
  "days": 365,
  "key_usage": ["digital_signature", "key_encipherment"],
  "ext_key_usage": ["server_auth"],
  "ca_name": "ca-sm2"
}
```

**响应**

```json
{
  "name": "new-server",
  "not_before": "2024-01-01T00:00:00Z",
  "not_after": "2025-01-01T00:00:00Z"
}
```

### 4.5 删除证书

**请求**

```
DELETE /api/v1/certificates/:name
```

**响应**

```
HTTP 204 No Content
```

## 5. 系统接口

### 5.1 获取系统信息

**请求**

```
GET /api/v1/system/info
```

**响应**

```json
{
  "version": "1.0.0",
  "go_version": "go1.21.0",
  "os": "linux",
  "arch": "amd64",
  "uptime": 86400,
  "start_time": "2024-01-01T00:00:00Z",
  "pid": 12345,
  "goroutines": 25,
  "memory": {
    "alloc_mb": 10,
    "sys_mb": 50
  }
}
```

### 5.2 健康检查

**请求**

```
GET /api/v1/system/health
```

**响应**

```json
{
  "status": "healthy",
  "instances": {
    "total": 2,
    "running": 2,
    "stopped": 0
  },
  "certificates": {
    "total": 4,
    "expired": 0,
    "expiring_soon": 0
  }
}
```

### 5.3 关闭服务

**请求**

```
POST /api/v1/system/shutdown
```

**响应**

```
HTTP 200 OK

shutting down
```

## 6. 数据模型

### 6.1 Instance 实例模型

```typescript
interface Instance {
  name: string;
  type: 'server' | 'client' | 'http-server' | 'http-client';
  protocol: 'auto' | 'tlcp' | 'tls';
  auth: 'none' | 'one-way' | 'mutual';
  listen: string;
  target: string;
  enabled: boolean;
  status: 'created' | 'running' | 'stopped' | 'error';
  uptime?: number;
  config?: InstanceConfig;
  stats?: InstanceStats;
}
```

### 6.2 Certificate 证书模型

```typescript
interface Certificate {
  name: string;
  type: 'tlcp' | 'tls';
  subject: string;
  issuer: string;
  not_before: string;
  not_after: string;
  is_ca: boolean;
  serial_number: string;
  dns_names?: string[];
  ip_addresses?: string[];
  public_key_algorithm: string;
  signature_algorithm: string;
}
```

## 7. 密码套件配置

### 7.1 TLCP密码套件

支持以下配置格式：

| 名称 | 16进制 | 数字 | 描述 |
|------|--------|------|------|
| ECC_SM4_CBC_SM3 | 0xC011 | 49169 | SM2密钥交换+SM4_CBC+SM3 |
| ECC_SM4_GCM_SM3 | 0xC012 | 49170 | SM2密钥交换+SM4_GCM+SM3 |
| ECC_SM4_CCM_SM3 | 0xC019 | 49177 | SM2密钥交换+SM4_CCM+SM3 |
| ECDHE_SM4_CBC_SM3 | 0xC013 | 49171 | ECDHE密钥交换+SM4_CBC+SM3 |
| ECDHE_SM4_GCM_SM3 | 0xC014 | 49172 | ECDHE密钥交换+SM4_GCM+SM3 |
| ECDHE_SM4_CCM_SM3 | 0xC01A | 49178 | ECDHE密钥交换+SM4_CCM+SM3 |

**配置示例：**

```yaml
tlcp:
  min_version: "1.1"
  max_version: "1.1"
  cipher_suites:
    - "ECDHE_SM4_GCM_SM3"
    - "0xC014"
    - 49172
```

### 7.2 TLS密码套件

支持以下配置格式：

| 名称 | 16进制 | 描述 |
|------|--------|------|
| TLS_RSA_WITH_AES_128_GCM_SHA256 | 0x009C | RSA+AES128_GCM |
| TLS_RSA_WITH_AES_256_GCM_SHA384 | 0x009D | RSA+AES256_GCM |
| TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 | 0xC02B | ECDHE+ECDSA+AES128_GCM |
| TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 | 0xC02C | ECDHE+ECDSA+AES256_GCM |
| TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 | 0xC02F | ECDHE+RSA+AES128_GCM |
| TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 | 0xC030 | ECDHE+RSA+AES256_GCM |

## 8. 协议版本配置

### 8.1 TLCP版本

| 名称 | 16进制 | 数字 | 描述 |
|------|--------|------|------|
| 1.1 | 0x0101 | 257 | TLCP 1.1 |

### 8.2 TLS版本

| 名称 | 16进制 | 数字 | 描述 |
|------|--------|------|------|
| 1.0 | 0x0301 | 769 | TLS 1.0 |
| 1.1 | 0x0302 | 770 | TLS 1.1 |
| 1.2 | 0x0303 | 771 | TLS 1.2 |
| 1.3 | 0x0304 | 772 | TLS 1.3 |

**配置示例：**

```yaml
tls:
  min_version: "1.2"
  max_version: "1.3"
```

## 9. 证书名称引用

证书通过名称引用，不直接使用文件路径：

```yaml
certificates:
  tlcp:
    cert: "server-sm2"
    key: "server-sm2"
  tls:
    cert: "server-rsa"
    key: "server-rsa"

client_ca:
  - "ca-sm2"
  - "ca-rsa"
```

证书文件存储在预定义目录中：
- TLCP证书：`./certs/tlcp/`
- TLS证书：`./certs/tls/`

## 10. 错误响应示例

**参数错误**

```
HTTP 400 Bad Request
Content-Type: text/plain

实例名称不能为空
```

**资源不存在**

```
HTTP 404 Not Found
Content-Type: text/plain

实例 unknown-instance 不存在
```

**资源冲突**

```
HTTP 409 Conflict
Content-Type: text/plain

实例 tlcp-server 已存在
```

**服务器错误**

```
HTTP 500 Internal Server Error
Content-Type: text/plain

启动实例失败：端口已被占用
```

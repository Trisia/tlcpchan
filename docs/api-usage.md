# API 使用指南

本文档提供 TLCP Channel RESTful API 的详细使用示例。

## 概述

### 基本信息

- 基础 URL：`http://localhost:8080/api/v1`
- 内容类型：`application/json`
- 字符编码：`UTF-8`

### 响应格式

**成功响应**：直接返回数据对象

```json
{
  "instances": [],
  "total": 0
}
```

**错误响应**：HTTP 状态码 + 纯文本错误信息

```
HTTP 400 Bad Request
Content-Type: text/plain

实例名称不能为空
```

### HTTP 状态码

| 状态码 | 描述 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

## 实例管理

### 获取实例列表

```bash
curl -s http://localhost:8080/api/v1/instances | jq
```

响应：

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

### 创建实例

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-tlcp-server",
    "type": "server",
    "protocol": "tlcp",
    "auth": "one-way",
    "listen": ":443",
    "target": "127.0.0.1:8080",
    "enabled": true,
    "certificates": {
      "tlcp": {
        "cert": "server-sm2",
        "key": "server-sm2"
      }
    },
    "tlcp": {
      "min_version": "1.1",
      "cipher_suites": ["ECDHE_SM4_GCM_SM3"]
    }
  }'
```

响应：

```
HTTP 201 Created

{
  "name": "my-tlcp-server",
  "status": "created"
}
```

### 获取实例详情

```bash
curl -s http://localhost:8080/api/v1/instances/my-tlcp-server | jq
```

响应：

```json
{
  "name": "my-tlcp-server",
  "type": "server",
  "protocol": "tlcp",
  "auth": "one-way",
  "listen": ":443",
  "target": "127.0.0.1:8080",
  "status": "running",
  "enabled": true,
  "uptime": 3600,
  "config": {
    "certificates": {
      "tlcp": {"cert": "server-sm2", "key": "server-sm2"}
    },
    "tlcp": {
      "min_version": "1.1",
      "cipher_suites": ["ECDHE_SM4_GCM_SM3"]
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

### 更新实例配置

```bash
curl -X PUT http://localhost:8080/api/v1/instances/my-tlcp-server \
  -H "Content-Type: application/json" \
  -d '{
    "target": "127.0.0.1:9090",
    "enabled": true
  }'
```

响应：

```json
{
  "name": "my-tlcp-server",
  "status": "updated"
}
```

### 删除实例

```bash
curl -X DELETE http://localhost:8080/api/v1/instances/my-tlcp-server
```

响应：

```
HTTP 204 No Content
```

### 启动实例

```bash
curl -X POST http://localhost:8080/api/v1/instances/my-tlcp-server/start
```

响应：

```json
{
  "name": "my-tlcp-server",
  "status": "running"
}
```

### 停止实例

```bash
curl -X POST http://localhost:8080/api/v1/instances/my-tlcp-server/stop
```

响应：

```json
{
  "name": "my-tlcp-server",
  "status": "stopped"
}
```

### 重载实例

```bash
curl -X POST http://localhost:8080/api/v1/instances/my-tlcp-server/reload
```

响应：

```json
{
  "name": "my-tlcp-server",
  "status": "reloaded"
}
```

### 获取实例统计

```bash
curl -s "http://localhost:8080/api/v1/instances/my-tlcp-server/stats" | jq
```

查询参数：

| 参数 | 类型 | 描述 |
|------|------|------|
| period | string | 统计周期：1m/5m/15m/1h/1d |

响应：

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

### 获取实例日志

```bash
curl -s "http://localhost:8080/api/v1/instances/my-tlcp-server/logs?lines=50&level=info" | jq
```

查询参数：

| 参数 | 类型 | 描述 |
|------|------|------|
| lines | integer | 返回行数，默认100，最大1000 |
| level | string | 日志级别：debug/info/warn/error |

响应：

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

## 配置管理

### 获取当前配置

```bash
curl -s http://localhost:8080/api/v1/config | jq
```

响应：

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

### 重载配置

```bash
curl -X POST http://localhost:8080/api/v1/config/reload
```

响应：

```json
{
  "reloaded": true,
  "changes": {
    "instances": {
      "added": ["new-instance"],
      "removed": ["old-instance"],
      "modified": ["tlcp-server"]
    }
  }
}
```

### 验证配置

```bash
curl -X POST http://localhost:8080/api/v1/config/validate \
  -H "Content-Type: application/json" \
  -d '{
    "instances": [
      {
        "name": "test-instance",
        "type": "server",
        "protocol": "tlcp",
        "listen": ":443",
        "target": "127.0.0.1:8080"
      }
    ]
  }'
```

成功响应：

```json
{
  "valid": true
}
```

错误响应：

```
HTTP 400 Bad Request

配置验证失败：实例 test-instance 监听地址格式错误
```

## 证书管理

### 获取证书列表

```bash
curl -s http://localhost:8080/api/v1/certificates | jq
```

响应：

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

### 获取证书详情

```bash
curl -s http://localhost:8080/api/v1/certificates/server-sm2 | jq
```

响应：

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

### 热更新证书

```bash
curl -X POST http://localhost:8080/api/v1/certificates/reload
```

响应：

```json
{
  "reloaded": true,
  "updated": ["server-sm2", "client-sm2"]
}
```

### 生成证书

```bash
curl -X POST http://localhost:8080/api/v1/certificates/generate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "tlcp",
    "name": "new-server",
    "common_name": "example.com",
    "dns_names": ["example.com", "*.example.com"],
    "ip_addresses": ["192.168.1.1"],
    "days": 365,
    "key_usage": ["digital_signature", "key_encipherment"],
    "ext_key_usage": ["server_auth"],
    "ca_name": "ca-sm2"
  }'
```

响应：

```json
{
  "name": "new-server",
  "not_before": "2024-01-01T00:00:00Z",
  "not_after": "2025-01-01T00:00:00Z"
}
```

### 删除证书

```bash
curl -X DELETE http://localhost:8080/api/v1/certificates/old-cert
```

响应：

```
HTTP 204 No Content
```

## 系统接口

### 获取系统信息

```bash
curl -s http://localhost:8080/api/v1/system/info | jq
```

响应：

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

### 健康检查

```bash
curl -s http://localhost:8080/api/v1/system/health | jq
```

响应：

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

### 关闭服务

```bash
curl -X POST http://localhost:8080/api/v1/system/shutdown
```

响应：

```
HTTP 200 OK

shutting down
```

## 错误处理

### 常见错误示例

**参数错误**

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{"name": ""}'
```

响应：

```
HTTP 400 Bad Request
Content-Type: text/plain

实例名称不能为空
```

**资源不存在**

```bash
curl http://localhost:8080/api/v1/instances/unknown
```

响应：

```
HTTP 404 Not Found
Content-Type: text/plain

实例 unknown 不存在
```

**资源冲突**

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{"name": "existing-instance", "type": "server", "listen": ":443", "target": "127.0.0.1:8080"}'
```

响应：

```
HTTP 409 Conflict
Content-Type: text/plain

实例 existing-instance 已存在
```

**服务器错误**

```bash
curl -X POST http://localhost:8080/api/v1/instances/test/start
```

响应：

```
HTTP 500 Internal Server Error
Content-Type: text/plain

启动实例失败：端口已被占用
```

### 错误处理最佳实践

```bash
#!/bin/bash

# 检查响应状态码
response=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/instances/test)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
    echo "成功: $body"
elif [ "$http_code" -eq 404 ]; then
    echo "资源不存在"
else
    echo "错误 ($http_code): $body"
fi
```

## 最佳实践

### 使用 jq 处理 JSON

```bash
# 格式化输出
curl -s http://localhost:8080/api/v1/instances | jq

# 提取特定字段
curl -s http://localhost:8080/api/v1/instances | jq '.instances[].name'

# 过滤数据
curl -s http://localhost:8080/api/v1/instances | jq '.instances[] | select(.status == "running")'

# 统计
curl -s http://localhost:8080/api/v1/instances | jq '.total'
```

### 批量操作脚本

```bash
#!/bin/bash

API_URL="http://localhost:8080/api/v1"

# 批量创建实例
create_instances() {
    local configs=(
        '{"name":"server1","type":"server","protocol":"tlcp","listen":":443","target":"127.0.0.1:8080"}'
        '{"name":"server2","type":"server","protocol":"tls","listen":":8443","target":"127.0.0.1:8080"}'
    )
    
    for config in "${configs[@]}"; do
        curl -s -X POST "$API_URL/instances" \
            -H "Content-Type: application/json" \
            -d "$config"
        echo
    done
}

# 批量启动实例
start_all() {
    local names=$(curl -s "$API_URL/instances" | jq -r '.instances[].name')
    for name in $names; do
        curl -s -X POST "$API_URL/instances/$name/start"
        echo "Started: $name"
    done
}

# 健康检查
health_check() {
    curl -s "$API_URL/system/health" | jq '.status'
}

create_instances
start_all
health_check
```

### 监控脚本

```bash
#!/bin/bash

API_URL="http://localhost:8080/api/v1"

while true; do
    echo "=== $(date) ==="
    
    # 实例状态
    curl -s "$API_URL/instances" | jq -r '.instances[] | "\(.name): \(.status)"'
    
    # 系统资源
    curl -s "$API_URL/system/info" | jq '{uptime, memory}'
    
    echo
    sleep 10
done
```

### 配置文件示例

创建实例配置文件 `instance.json`：

```json
{
  "name": "production-tlcp",
  "type": "server",
  "protocol": "tlcp",
  "auth": "mutual",
  "listen": ":443",
  "target": "10.0.0.10:8080",
  "enabled": true,
  "certificates": {
    "tlcp": {
      "cert": "prod-server-sm2",
      "key": "prod-server-sm2"
    }
  },
  "client_ca": ["prod-ca-sm2"],
  "tlcp": {
    "min_version": "1.1",
    "cipher_suites": ["ECDHE_SM4_GCM_SM3"]
  }
}
```

使用配置文件创建：

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d @instance.json
```

### Python 调用示例

```python
import requests
import json

API_URL = "http://localhost:8080/api/v1"

class TLCPChanClient:
    def __init__(self, api_url=API_URL):
        self.api_url = api_url
        self.session = requests.Session()
    
    def list_instances(self):
        resp = self.session.get(f"{self.api_url}/instances")
        resp.raise_for_status()
        return resp.json()
    
    def create_instance(self, config):
        resp = self.session.post(
            f"{self.api_url}/instances",
            json=config
        )
        resp.raise_for_status()
        return resp.json()
    
    def start_instance(self, name):
        resp = self.session.post(f"{self.api_url}/instances/{name}/start")
        resp.raise_for_status()
        return resp.json()
    
    def get_stats(self, name):
        resp = self.session.get(f"{self.api_url}/instances/{name}/stats")
        resp.raise_for_status()
        return resp.json()

# 使用示例
client = TLCPChanClient()

# 列出实例
for inst in client.list_instances()['instances']:
    print(f"{inst['name']}: {inst['status']}")

# 创建实例
config = {
    "name": "test-instance",
    "type": "server",
    "protocol": "tlcp",
    "listen": ":443",
    "target": "127.0.0.1:8080"
}
client.create_instance(config)

# 启动实例
client.start_instance("test-instance")
```

## 密码套件配置

### TLCP 密码套件

| 名称 | 16进制 | 描述 |
|------|--------|------|
| ECC_SM4_CBC_SM3 | 0xC011 | SM2密钥交换+SM4_CBC+SM3 |
| ECC_SM4_GCM_SM3 | 0xC012 | SM2密钥交换+SM4_GCM+SM3 |
| ECDHE_SM4_CBC_SM3 | 0xC013 | ECDHE密钥交换+SM4_CBC+SM3 |
| ECDHE_SM4_GCM_SM3 | 0xC014 | ECDHE密钥交换+SM4_GCM+SM3 |

配置示例：

```json
{
  "tlcp": {
    "min_version": "1.1",
    "cipher_suites": ["ECDHE_SM4_GCM_SM3", "0xC012"]
  }
}
```

### TLS 密码套件

| 名称 | 16进制 | 描述 |
|------|--------|------|
| TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 | 0xC02F | ECDHE+RSA+AES128_GCM |
| TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 | 0xC030 | ECDHE+RSA+AES256_GCM |
| TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 | 0xC02B | ECDHE+ECDSA+AES128_GCM |

配置示例：

```json
{
  "tls": {
    "min_version": "1.2",
    "max_version": "1.3",
    "cipher_suites": ["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"]
  }
}
```

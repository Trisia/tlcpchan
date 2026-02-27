# TLCP Channel MCP 使用指南

## 1. 概述

### 1.1 MCP 是什么

MCP (Model Context Protocol) 是一种开放协议，用于 AI 助手与外部工具和数据源之间的标准化通信。通过 MCP，AI 模型可以安全、高效地调用各种功能，而无需知道底层的实现细节。

### 1.2 TLCP Channel 的 MCP 功能

TLCP Channel 实现了 MCP 服务器端，提供了完整的 TLCP/TLS 代理管理能力。通过 MCP，您可以：

- 管理代理实例（创建、启动、停止、删除）
- 查询系统信息和统计
- 管理密钥存储
- 读取系统日志
- 管理配置文件

### 1.3 支持的工具列表

TLCP Channel MCP 提供了以下 5 类共 18 个工具：

#### 配置管理工具 (3 个)
- `get_config` - 获取当前系统配置
- `update_config` - 更新系统配置
- `reload_config` - 重新加载配置文件

#### 密钥存储管理工具 (5 个)
- `list_keystores` - 获取所有密钥存储的列表信息
- `get_keystore` - 获取指定名称的密钥存储的详细信息
- `create_keystore` - 创建新的密钥存储
- `update_keystore` - 更新指定密钥存储的参数
- `delete_keystore` - 删除指定的密钥存储

#### 日志管理工具 (1 个)
- `get_system_logs` - 获取系统日志（历史日志文件）

#### 系统信息工具 (2 个)
- `get_system_info` - 获取系统信息（版本、Go版本、操作系统、架构、运行时长）
- `get_system_stats` - 获取系统统计信息（CPU使用率、内存使用、总连接数、活跃实例数）

#### 实例管理工具 (10 个)
- `list_instances` - 获取所有代理实例的列表信息
- `get_instance` - 获取指定实例的详细信息
- `create_instance` - 创建新的代理实例
- `update_instance` - 更新实例配置，支持热重载
- `delete_instance` - 删除指定实例
- `start_instance` - 启动指定实例
- `stop_instance` - 停止指定实例
- `restart_instance` - 重启指定实例
- `get_instance_stats` - 获取实例运行统计信息
- `check_instance_health` - 检查实例健康状态

## 2. 配置

### 2.1 启用 MCP 服务

在配置文件中添加 MCP 配置块：

```yaml
mcp:
  enabled: true
  api_key: ""  # 留空则开放访问
  server_info:
    name: "tlcpchan-mcp"
    version: "1.0.0"
```

### 2.2 配置参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `enabled` | boolean | 是 | 是否启用 MCP 服务 |
| `api_key` | string | 否 | API 密钥，留空则无需认证 |
| `server_info.name` | string | 否 | 服务器名称，默认 "tlcpchan-mcp" |
| `server_info.version` | string | 否 | 版本号，默认 "1.0.0" |

### 2.3 完整配置示例

```yaml
server:
  api:
    address: ":20080"
  log:
    level: "info"
    file: "./logs/tlcpchan.log"
    max-size: 100
    max-backups: 5
    max-age: 30
    compress: true
    enabled: true

mcp:
  enabled: true
  api_key: "your-secret-api-key"  # 生产环境建议设置
  server_info:
    name: "tlcpchan-mcp"
    version: "1.0.0"

instances:
  - name: "proxy-example"
    type: "server"
    listen: ":8443"
    target: "backend.example.com:443"
    protocol: "auto"
    enabled: true
    tlcp:
      keystore:
        name: "tlcp-keystore"
```

## 3. 连接方式

### 3.1 SSE 连接

TLCP Channel MCP 使用 Server-Sent Events (SSE) 作为传输层，连接端点为：

```
http://localhost:20080/api/mcp/sse
```

如果配置了不同的 API 地址，请相应调整端点。

### 3.2 认证方式

如果配置了 `api_key`，需要在请求头中携带 Bearer Token：

```
Authorization: Bearer your-secret-api-key
```

如果 `api_key` 为空，则无需认证，可以直接连接。

### 3.3 连接流程

1. 客户端发送 GET 请求到 `/api/mcp/sse` 端点
2. 服务器建立 SSE 连接
3. 客户端发送初始化消息
4. 服务器返回服务器信息和可用工具列表
5. 客户端可以开始调用工具

## 4. 客户端示例

### 4.1 JavaScript/TypeScript

使用 `@modelcontextprotocol/sdk-js` 客户端：

```typescript
import { Client } from '@modelcontextprotocol/sdk-js';

async function connectToMCP() {
  // 创建客户端
  const client = new Client({
    "name": "tlcpchan-client",
    "version": "1.0.0"
  }, {
    capabilities: {}
  });

  // 连接到 SSE 服务器
  const url = "http://localhost:20080/api/mcp/sse";
  const apiKey = "your-secret-api-key"; // 如果需要认证

  await client.connect(
    url,
    {
      headers: {
        "Authorization": `Bearer ${apiKey}`
      }
    }
  );

  // 获取实例列表
  const instances = await client.callTool({
    name: "list_instances",
    arguments: {}
  });

  console.log("实例列表:", instances);

  // 创建新实例
  const newInstance = await client.callTool({
    name: "create_instance",
    arguments: {
      config: {
        name: "my-proxy",
        type: "server",
        listen: ":8443",
        target: "example.com:443",
        protocol: "auto",
        enabled: true
      }
    }
  });

  console.log("创建的实例:", newInstance);

  // 关闭连接
  await client.close();
}

connectToMCP().catch(console.error);
```

### 4.2 Python

使用原生 HTTP 和 SSE 客户端：

```python
import json
import requests
import sseclient

class MCPClient:
    def __init__(self, base_url, api_key=None):
        self.base_url = base_url
        self.api_key = api_key
        self.session_id = None

    def _get_headers(self):
        headers = {'Accept': 'text/event-stream'}
        if self.api_key:
            headers['Authorization'] = f'Bearer {self.api_key}'
        return headers

    def call_tool(self, tool_name, arguments):
        """调用 MCP 工具"""
        # 构建请求消息
        request_id = 1
        message = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }

        # 发送 SSE 请求
        url = f"{self.base_url}/api/mcp/sse"
        response = requests.post(
            url,
            headers=self._get_headers(),
            json=message,
            stream=True
        )

        # 解析 SSE 响应
        client = sseclient.SSEClient(response)
        for event in client.events():
            if event.event == 'message':
                data = json.loads(event.data)
                if 'result' in data:
                    return data['result']
                elif 'error' in data:
                    raise Exception(data['error'])
        return None

# 使用示例
if __name__ == "__main__":
    client = MCPClient(
        "http://localhost:20080",
        api_key="your-secret-api-key"
    )

    # 获取实例列表
    instances = client.call_tool("list_instances", {})
    print("实例列表:", json.dumps(instances, indent=2, ensure_ascii=False))

    # 获取系统信息
    system_info = client.call_tool("get_system_info", {})
    print("系统信息:", json.dumps(system_info, indent=2, ensure_ascii=False))

    # 检查实例健康状态
    health = client.call_tool("check_instance_health", {
        "name": "proxy-example",
        "timeout": 10
    })
    print("健康检查:", json.dumps(health, indent=2, ensure_ascii=False))
```

### 4.3 cURL

使用 cURL 进行快速测试：

```bash
# 获取实例列表
curl -X GET "http://localhost:20080/api/mcp/sse" \
  -H "Authorization: Bearer your-secret-api-key" \
  -H "Accept: text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "list_instances",
      "arguments": {}
    }
  }'

# 获取系统信息
curl -X GET "http://localhost:20080/api/mcp/sse" \
  -H "Authorization: Bearer your-secret-api-key" \
  -H "Accept: text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_system_info",
      "arguments": {}
    }
  }'

# 检查实例健康
curl -X GET "http://localhost:20080/api/mcp/sse" \
  -H "Authorization: Bearer your-secret-api-key" \
  -H "Accept: text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "check_instance_health",
      "arguments": {
        "name": "proxy-example",
        "timeout": 10
      }
    }
  }'
```

## 5. 工具使用示例

### 5.1 获取实例列表

```json
{
  "name": "list_instances",
  "arguments": {}
}
```

响应示例：

```json
{
  "instances": [
    {
      "name": "proxy-example",
      "status": "running",
      "enabled": true,
      "config": {
        "name": "proxy-example",
        "type": "server",
        "listen": ":8443",
        "target": "backend.example.com:443",
        "protocol": "auto",
        "enabled": true
      }
    }
  ]
}
```

### 5.2 创建实例

```json
{
  "name": "create_instance",
  "arguments": {
    "config": {
      "name": "my-proxy",
      "type": "server",
      "listen": ":8444",
      "target": "example.com:443",
      "protocol": "auto",
      "enabled": true,
      "tlcp": {
        "keystore": {
          "name": "tlcp-keystore"
        }
      }
    }
  }
}
```

响应示例：

```json
{
  "name": "my-proxy",
  "status": "stopped"
}
```

### 5.3 检查健康状态

```json
{
  "name": "check_instance_health",
  "arguments": {
    "name": "proxy-example",
    "timeout": 10
  }
}
```

响应示例：

```json
{
  "instance": "proxy-example",
  "results": [
    {
      "protocol": "tlcp",
      "success": true,
      "latencyMs": 45,
      "error": ""
    },
    {
      "protocol": "tls",
      "success": true,
      "latencyMs": 38,
      "error": ""
    }
  ]
}
```

### 5.4 获取系统日志

```json
{
  "name": "get_system_logs",
  "arguments": {
    "lines": 100,
    "level": "ERROR"
  }
}
```

响应示例：

```json
{
  "logs": [
    {
      "timestamp": "2025/02/27 10:30:45.123",
      "level": "ERROR",
      "message": "连接失败: connection refused"
    },
    {
      "timestamp": "2025/02/27 10:30:46.456",
      "level": "ERROR",
      "message": "握手超时"
    }
  ]
}
```

### 5.5 更新配置

```json
{
  "name": "update_config",
  "arguments": {
    "config": {
      "server": {
        "api": {
          "address": ":20080"
        }
      },
      "mcp": {
        "enabled": true,
        "api_key": "new-api-key"
      }
    }
  }
}
```

## 6. 安全建议

### 6.1 API Key 管理

#### 生产环境
- **必须设置强密钥**：使用至少 32 位的随机字符串
- **定期更换**：建议每 3-6 个月更换一次 API Key
- **安全存储**：使用密钥管理系统或环境变量存储，不要硬编码在代码中

#### 开发环境
- 可以留空 `api_key` 以便快速测试
- 确保开发环境不暴露到公网

#### API Key 生成示例

```bash
# 生成 32 字节的随机 API Key
openssl rand -hex 32
```

### 6.2 网络配置

#### 监听地址
- **生产环境**：监听本地回环地址 `127.0.0.1:20080`，使用反向代理（如 Nginx）处理外部访问
- **开发环境**：可以监听 `:20080` 以便从其他机器访问

#### 反向代理配置示例 (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name tlcpchan.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location /api/mcp/ {
        proxy_pass http://127.0.0.1:20080/api/mcp/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE 特定配置
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400;
    }
}
```

### 6.3 权限控制

#### 文件权限
- 配置文件应设置为 `600`（仅所有者可读写）
- 日志目录应设置为 `700`（仅所有者可访问）

```bash
chmod 600 /etc/tlcpchan/config/config.yaml
chmod 700 /etc/tlcpchan/logs
```

#### 运行用户
- 不要使用 root 用户运行服务
- 创建专用用户运行 TLCP Channel

```bash
# 创建专用用户
sudo useradd -r -s /bin/false tlcpchan

# 使用专用用户运行
sudo -u tlcpchan ./tlcpchan -c /etc/tlcpchan/config/config.yaml
```

### 6.4 日志审计

- 启用日志记录所有 MCP 调用
- 定期检查日志文件，发现异常访问
- 配置日志轮转，避免日志文件过大

### 6.5 速率限制

如果使用反向代理，建议配置速率限制：

```nginx
limit_req_zone $binary_remote_addr zone=mcp_limit:10m rate=10r/s;

location /api/mcp/ {
    limit_req zone=mcp_limit burst=20;
    # ... 其他配置
}
```

## 7. 故障排除

### 7.1 连接失败

**问题**：无法连接到 MCP 服务端点

**检查项**：
1. 确认 `mcp.enabled` 设置为 `true`
2. 检查 TLCP Channel 服务是否正在运行
3. 验证 API 地址配置是否正确
4. 检查防火墙规则

**解决方案**：
```bash
# 检查服务状态
curl http://localhost:20080/api/mcp/sse

# 查看服务日志与其他相关日志
tail -f /etc/tlcpchan/logs/tlcpchan.log
```

### 7.2 认证失败

**问题**：返回 401 Unauthorized

**检查项**：
1. 确认 `api_key` 配置正确
2. 检查 Authorization 头格式：`Bearer <api_key>`
3. 验证 API Key 是否匹配

**解决方案**：
```bash
# 测试认证
curl -X GET "http://localhost:20080/api/mcp/sse" \
  -H "Authorization: Bearer your-api-key"
```

### 7.3 工具调用失败

**问题**：工具调用返回错误

**检查项**：
1. 确认工具名称拼写正确
2. 检查参数格式是否符合 schema
3. 查看服务日志获取详细错误信息

**常见错误**：
- `实例不存在`：指定的实例名称不存在
- `参数不能为空`：缺少必需参数
- `端口冲突`：监听端口已被占用

## 8. 进阶用法

### 8.1 与 AI 助手集成

将 TLCP Channel MCP 集成到支持 MCP 协议的 AI 助手（如 Claude Desktop）：

```json
{
  "mcpServers": {
    "tlcpchan": {
      "command": "node",
      "args": ["-e", "require('@modelcontextprotocol/sdk-node').connectSSE({ url: 'http://localhost:20080/api/mcp/sse' })"]
    }
  }
}
```

### 8.2 批量操作

使用 MCP 工具批量管理实例：

```javascript
// 批量启动所有实例
async function startAllInstances(client) {
  const instances = await client.callTool({
    name: "list_instances",
    arguments: {}
  });

  for (const instance of instances.instances) {
    if (instance.status !== "running" && instance.enabled) {
      await client.callTool({
        name: "start_instance",
        arguments: { name: instance.name }
      });
      console.log(`已启动实例: ${instance.name}`);
    }
  }
}
```

### 8.3 健康检查监控

定期检查所有实例的健康状态：

```python
import asyncio
import mcp_client

async def health_monitor():
    client = mcp_client.MCPClient("http://localhost:20080", "api-key")
    
    while True:
        instances = await client.call_tool("list_instances", {})
        
        for inst in instances["instances"]:
            if inst["status"] == "running":
                health = await client.call_tool("check_instance_health", {
                    "name": inst["name"],
                    "timeout": 5
                })
                
                for result in health["results"]:
                    if not result["success"]:
                        print(f"实例 {inst['name']} 健康检查失败: {result['error']}")
        
        await asyncio.sleep(60)  # 每分钟检查一次

asyncio.run(health_monitor())
```

## 9. 参考资料

### 9.1 相关文档

- [TLCP Channel 设计文档](./design.md)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [go-mcp SDK](https://github.com/modelcontextprotocol/go-sdk)
- [gotlcp 库](https://github.com/Trisia/gotlcp)

### 9.2 支持与反馈

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [tlcpchan/issues](https://github.com/Trisia/tlcpchan/issues)
- 邮件: support@example.com

---

**最后更新**: 2025-02-27
**版本**: 1.0.0

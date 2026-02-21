# TLCP Channel MCP 服务集成

## 概述

本项目为 TLCP Channel 添加了 MCP (Model Context Protocol) 服务支持，使 LLM 可以通过自然语言控制 TLCP 代理。

## 已完成的功能

### 1. 依赖添加
- ✅ 添加了 `github.com/gorilla/websocket` 依赖

### 2. 配置扩展
在 `config/config.go` 中添加了 `MCPConfig` 结构：
```yaml
mcp:
  enabled: true  # 开关
  api_key: "sk-tlcp-xxxx"  # 可选API密钥
```

### 3. 核心模块创建

#### mcp/protocol/
- ✅ `protocol.go` - JSON-RPC 2.0 和 MCP 协议定义

#### mcp/connection/
- ✅ `connection.go` - WebSocket 连接管理

#### mcp/tools/
- ✅ `base.go` - 工具基类和工具注册表
- ✅ `keystore_manager.go` - keystore 管理工具（核心功能）
- ✅ `certificate_manager.go` - 证书管理工具

#### mcp/
- ✅ `server.go` - MCP 服务器核心实现

### 4. 控制器集成
在 `controller/server.go` 中：
- ✅ 集成了 MCP 服务器
- ✅ 注册了 `/mcp/ws` 端点
- ✅ 注册了 `keystore_manager` 工具
- ✅ 注册了 `certificate_manager` 工具

## 使用方式

### 配置示例
```yaml
# config.yaml
mcp:
  enabled: true
  api_key: ""  # 留空表示无需认证
```

### WebSocket 连接
连接地址：`ws://host:20080/mcp/ws`

如果配置了 API 密钥：
`ws://host:20080/mcp/ws?api_key=your-key-here`

## MCP 工具

### 1. keystore_manager
管理 TLCP/TLS 密钥存储库的完整生命周期。

#### 方法列表
- `list` - 列出所有可用的 keystores
- `get(name)` - 获取指定 keystore 的详细信息
- `generate(name, type, commonName, ...)` - 生成包含证书的 keystore
- `delete(name)` - 删除指定的 keystore
- `reload(name)` - 重载指定的 keystore

#### 使用示例
```json
{
  "method": "tools/call",
  "params": {
    "name": "keystore_manager",
    "arguments": {
      "method": "generate",
      "params": {
        "name": "my-tlcp-cert",
        "type": "tlcp",
        "commonName": "example.com",
        "org": "My Org",
        "years": 5
      }
    }
  }
}
```

### 2. certificate_manager
管理国密/TLS证书的操作，包括生成、导入、验证等。

#### 方法列表
- `generate_root_ca(type, commonName, ...)` - 生成根CA证书
- `import_certificate(filename, certData, format)` - 导入证书（支持PEM/DER/Base64）
- `list_certificates()` - 列出所有根证书
- `get_certificate(filename)` - 获取指定证书详情
- `delete_certificate(filename)` - 删除指定证书
- `validate_certificate(certData)` - 验证证书有效性

#### 使用示例
```json
{
  "method": "tools/call",
  "params": {
    "name": "certificate_manager",
    "arguments": {
      "method": "generate_root_ca",
      "params": {
        "type": "tlcp",
        "commonName": "My Root CA",
        "org": "My Org",
        "years": 10
      }
    }
  }
}
```

## 下一步计划

1. 完成 `certificate_manager` 工具
2. 添加 `instance_lifecycle` 工具
3. 完善错误处理
4. 添加更多的工具方法
5. 实现自然语言解析增强

## 注意事项

- 现有代码中的 `proxy/proxy_test.go` 有一个未定义的函数错误，这是原来就存在的问题，与本次修改无关
- 当前 keystore 生成功能需要完善，需要正确处理根 CA 的加载和使用

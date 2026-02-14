# AGENTS.md - TLCP Channel 项目指南

本文档为 agentic coding agents 提供项目上下文和代码规范。

## 项目概述

TLCP Channel 是一款 TLCP/TLS 协议代理工具，支持双协议并行工作。

- `tlcpchan/` - Go 核心代理服务
- `tlcpchan-ui/` - Web 前端界面

## 语言偏好

- 始终使用中文（简体）思考和回复
- 代码注释、提交信息使用中文

## 构建命令

### Go 项目 (tlcpchan/)

```bash
go build -o bin/tlcpchan ./cmd/tlcpchan  # 构建
go run ./cmd/tlcpchan                     # 运行
go test ./...                             # 运行所有测试
go test -v -run TestFunctionName ./path   # 运行单个测试
go test -cover ./...                      # 测试覆盖率
go fmt ./...                              # 格式化
golangci-lint run                         # 静态检查
go mod tidy                               # 整理依赖
```

### 前端项目 (tlcpchan-ui/)

```bash
npm install                               # 安装依赖
npm run dev                               # 开发模式
npm run build                             # 构建
npm run typecheck                         # 类型检查
npm run lint                              # 代码检查
npm test -- --grep "test name"            # 运行单个测试
```

## 代码风格指南

### Go 代码规范

#### 导入顺序

```go
import (
    // 标准库
    "context"
    "fmt"
    // 第三方库
    "github.com/Trisia/gotlcp"
    // 本项目内部包
    "github.com/Trisia/tlcpchan/proxy"
)
```

#### 命名约定

- 包名：小写单词，如 `proxy`, `tlcp`
- 导出函数/类型：大驼峰，如 `NewProxyServer`
- 私有函数/变量：小驼峰，如 `parseConfig`
- 接口：动词或名词+er，如 `Handler`, `ConnectionReader`

#### 错误处理

```go
if err != nil {
    return fmt.Errorf("连接失败: %w", err)
}
```

#### 日志规范

```go
logger.Info("代理服务启动", zap.String("address", addr))
logger.Error("连接失败", zap.Error(err), zap.String("remote", remoteAddr))
```

### TypeScript 代码规范

#### 文件命名

- 组件：大驼峰，如 `ProxyConfig.tsx`
- 工具函数：小驼峰，如 `formatBytes.ts`

#### 导入顺序

```typescript
import React, { useState } from 'react';        // React
import { Box } from '@mui/material';            // 第三方库
import { ConfigForm } from './ConfigForm';      // 内部组件
import type { ProxyConfig } from '@/types';     // 类型
```

#### 组件结构

```typescript
interface Props { config: ProxyConfig; onChange: (c: ProxyConfig) => void; }

export const ProxyConfig: React.FC<Props> = ({ config, onChange }) => {
    const [local, setLocal] = useState(config);
    return <Box>{/* 组件内容 */}</Box>;
};
```

## 项目特定约定

### TLCP 协议

- 使用 `github.com/Trisia/gotlcp` 库处理 TLCP 协议
- 证书使用国密 SM2 算法
- 双协议模式同时支持 TLCP 和 TLS

### 配置管理

- 配置文件使用 YAML 格式，路径 `./config/config.yaml`
- 支持环境变量覆盖配置项

### API 设计

- RESTful API 路由前缀: `/api/v1`
- WebSocket 路由: `/ws`
- 响应格式: `{"code": 0, "message": "success", "data": {}}`

## 测试规范

### Go 测试

```go
func TestProxyConnection(t *testing.T) {
    tests := []struct { name string; input string; wantErr bool }{
        {"正常连接", "localhost:8080", false},
        {"连接超时", "invalid:9999", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* 测试逻辑 */ })
    }
}
```

### 前端测试

- 使用 Vitest，组件测试使用 Testing Library

## 提交规范

```
类型: 简短描述

详细说明（可选）
```

类型：`feat` `fix` `docs` `refactor` `test` `chore`

## 注意事项

1. 不提交敏感信息（密钥、证书、密码等）
2. 使用 context.Context 进行超时和取消控制

## 注释规范

### 函数注释规范

除非函数功能特别简单直观（如简单的 getter/setter），否则必须为函数编写详细注释。

后端（Go）函数必须严格遵循注释规范，前端（TypeScript/React）函数可根据场景灵活处理。

注释应包括：

1. **函数功能**：简要描述函数的作用
2. **参数说明**：每个参数的含义和约束
3. **返回值说明**：返回值的含义，特别是错误情况
4. **注意事项**：使用时需要注意的事项（如果有）

#### Go 函数注释示例

```go
// NewProxyServer 创建新的代理服务器实例
// 参数:
//   - cfg: 代理配置，不能为 nil
//   - logger: 日志记录器，若为 nil 则使用默认日志
// 返回:
//   - *ProxyServer: 代理服务器实例
//   - error: 配置验证失败时返回错误
// 注意: 调用前需确保证书文件存在且可读
func NewProxyServer(cfg *Config, logger *zap.Logger) (*ProxyServer, error) {
    // ...
}

// Start 启动代理服务器
// 参数:
//   - ctx: 上下文，用于控制启动超时
// 返回:
//   - error: 端口占用或证书加载失败时返回错误
// 注意: 该方法会阻塞直到服务器停止或ctx取消
func (s *ProxyServer) Start(ctx context.Context) error {
    // ...
}

// parseAddress 解析地址字符串为host和port
// 参数:
//   - addr: 地址字符串，格式为 "host:port" 或 ":port"
// 返回:
//   - host: 主机地址，若输入为 ":port" 则返回 ""
//   - port: 端口号
//   - error: 地址格式无效时返回错误
func parseAddress(addr string) (host string, port int, err error) {
    // ...
}
```

#### TypeScript 函数注释示例

```typescript
/**
 * 格式化字节数为人类可读格式
 * @param bytes 字节数，必须为非负整数
 * @param decimals 小数位数，默认为2
 * @returns 格式化后的字符串，如 "1.5 MB"
 */
export function formatBytes(bytes: number, decimals = 2): string {
    // ...
}
```

### 类型/结构体注释规范

所有DTO（数据传输对象）、配置对象和关键结构体必须添加详细的注释说明：

#### 枚举类型注释

必须说明所有可选值及其含义：

```go
// Auth 认证模式
type Auth string

const (
    // AuthNone 无认证
    AuthNone Auth = "none"
    // AuthOneWay 单向认证，仅验证对端证书
    AuthOneWay Auth = "one-way"
    // AuthMutual 双向认证，双方互相验证证书
    AuthMutual Auth = "mutual"
)
```

#### 数值类型注释

必须明确说明数值单位：

```go
type LogConfig struct {
    // MaxSize 单个日志文件最大大小，单位: MB
    MaxSize int `yaml:"max_size"`
    // MaxBackups 保留的旧日志文件最大数量，单位: 个
    MaxBackups int `yaml:"max_backups"`
    // MaxAge 保留旧日志文件的最大天数，单位: 天
    MaxAge int `yaml:"max_age"`
    // AvgLatencyMs 平均延迟，单位: 毫秒
    AvgLatencyMs float64 `json:"avg_latency_ms"`
}
```

#### 字符串类型注释

若有特殊意义需说明，若有明确构成成分应举例说明：

```go
type APIConfig struct {
    // Address API服务监听地址
    // 格式: "host:port" 或 ":port"
    // 示例: ":8080" 表示监听所有网卡的8080端口
    // 示例: "127.0.0.1:8080" 表示仅监听本地回环地址
    Address string `yaml:"address"`
}

type CertConfig struct {
    // Cert 证书文件路径，支持PEM格式
    // 示例: "server.crt" 或 "./certs/server.pem"
    Cert string `yaml:"cert,omitempty"`
}
```

#### TypeScript 类型注释

```typescript
/**
 * 代理实例信息
 */
export interface Instance {
  /** 实例名称，全局唯一标识符 */
  name: string
  /** 代理类型
   * - server: TCP服务端代理
   * - client: TCP客户端代理
   */
  type: 'server' | 'client'
  /** 运行时长，单位: 秒 */
  uptime?: number
  /** 平均延迟，单位: 毫秒 */
  latency_avg_ms?: number
}
```

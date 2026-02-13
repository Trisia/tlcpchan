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
2. 代码中不添加不必要的注释
3. 使用 context.Context 进行超时和取消控制

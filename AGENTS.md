# AGENTS.md - TLCP Channel 项目指南


本文档为 agentic coding agents 提供项目上下文和代码规范。

## 构建/测试/格式化命令

### 后端（Go）

```bash
# 在 tlcpchan/ 目录下执行
cd tlcpchan
go build -o tlcpchan ./cmd/tlcpchan          # 构建核心服务
go test ./... -v                              # 运行所有测试
go test ./path/to/package -run TestName -v   # 运行单个测试
go fmt ./...                                  # 格式化代码
go vet ./...                                  # 静态检查
```

### 前端（Vue/TypeScript）

```bash
# 在 tlcpchan-ui/ 目录下执行
cd tlcpchan-ui
npm install          # 安装依赖
npm run dev          # 开发模式启动
npm run build        # 生产构建
npm run test         # 运行所有测试
npm run test:watch   # 监听模式运行测试
```

### CLI（Go）

```bash
# 在 tlcpchan-cli/ 目录下执行
cd tlcpchan-cli
go build -o tlcpchan-cli   # 构建命令行工具
```

## 规约

- 在进行开发前必须先把开发的内容系统的整理并加入设计文档
- 保证当前项目设计与设计文档一致
- 对应对于所有开发的功能尽可能的提供测试
- 对于测试产生的临时文件，应该在测试结束后进行清理，除非这些文件后续仍然需要

## 特殊约定

- 若opencode运行用户为root时，需要使用chrome-dev-tools 时，总是以 --no-sandbox 启动google-chrome 
- Windows环境中如果有git-bash.exe 优先使用这个最为shell的执行器（默认路径`C:\Program Files\Git\git-bash.exe`）

## 项目概述

TLCP Channel 是一款 TLCP/TLS 协议代理工具，支持双协议并行工作。

- `tlcpchan/` - Go 核心代理服务
- `tlcpchan-cli/` - 命令行工具
- `tlcpchan-ui/` - Web 前端界面

## 语言偏好

- 始终使用中文（简体）思考和回复
- 代码注释、提交信息使用中文

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

- 包名：小写单词，如 `proxy`, `cert`
- 导出函数/类型：大驼峰，如 `NewProxyServer`
- 私有函数/变量：小驼峰，如 `parseConfig`
- 接口：动词或名词+er，如 `Handler`, `ConnectionReader`
- 常量：大驼峰或全大写，如 `AuthNone` 或 `MAX_SIZE`

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
- 类型定义：大驼峰，如 `Instance.ts`

## 项目特定约定

### TLCP 协议

- 使用 `github.com/Trisia/gotlcp` 库处理 TLCP 协议
- 证书使用国密 SM2 算法
- 双协议模式同时支持 TLCP 和 TLS

### 配置管理

- 配置文件使用 YAML 格式，默认路径 `/etc/tlcpchan/config/config.yaml`
- 支持环境变量覆盖配置项
- 工作目录：Linux 为 `/etc/tlcpchan`，Windows 为程序所在目录

### API 设计

- RESTful API 路由前缀: `/api`
- 响应格式: HTTP RESTful，状态使用 HTTP status code 返回，内容直接在 body 中返回，例如 `code 500 body:系统内部错误`
- API 服务默认地址: `:20080`

## 测试规范

### Go 测试

```go
func TestProxyConnection(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"正常连接", "localhost:8080", false},
        {"连接超时", "invalid:9999", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
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
3. 工作目录结构：certs/（证书）、trusted/（信任证书）、logs/（日志）、config/（配置）

## 注释规范

### 函数注释规范

无论是修改，还是创建都应该遵守函数注释规范！

除非函数功能特别简单直观（如简单的 getter/setter），否则必须为函数编写详细注释。

后端（Go）函数必须严格遵循注释规范，前端（TypeScript/React）函数可根据场景灵活处理。

注释应包括：
1. **函数功能**：简要描述函数的作用
2. **参数说明**：每个参数的含义和约束
3. **返回值说明**：返回值的含义，特别是错误情况
4. **注意事项**：使用时需要注意的事项（如果有）

### 类型/结构体注释规范

所有 DTO（数据传输对象）、配置对象和关键结构体必须添加详细的注释说明：

- 枚举类型：说明所有可选值及其含义
- 数值类型：明确说明数值单位
- 字符串类型：如有特殊意义需说明，如有明确构成成分应举例说明

## 接口文档规范

- 在创建、修改、编辑 HTTP接口时，总是更新接口上的apidoc文档
- 接口参数和响应值都应该遵循驼峰命名，特殊名称或缩写除外
- 当对公开接口进行修改后必须检查关联的文档、CLI、UI涉及并进行修改

## 前端开发规范

- 实现可能简洁，不要过度封装
- 枚举类型的元素若元素数量少于10，下拉或其他需要循环显示的情况，直接以多行代码的方式实现，不要动态生成
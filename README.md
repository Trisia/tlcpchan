# TLCP Channel

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![TLCP](https://img.shields.io/badge/TLCP-1.1-green.svg)](https://github.com/Trisia/gotlcp)

一款 TLCP/TLS 协议代理工具，支持双协议并行工作，基于国密算法实现安全通信。

## 功能特性

- **双协议支持** - 同时支持 TLCP 1.1 和 TLS 1.0-1.3 协议
- **自动协议检测** - 同一端口自动识别 TLCP/TLS 客户端
- **多种代理模式** - 服务端代理、客户端代理、HTTP 代理
- **国密算法** - 支持 SM2/SM3/SM4 国密算法套件
- **灵活认证** - 支持单向认证、双向认证
- **Web 管理界面** - Vue3 + Element Plus 现代化管理界面
- **RESTful API** - 完整的 API 接口支持
- **命令行工具** - tlcpchan-cli 命令行管理工具
- **证书热更新** - 无需重启即可更新证书
- **流量统计** - 实时连接数、流量、延迟统计

## 快速开始

### Docker 部署（推荐）

```bash
# 使用 docker-compose
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan
docker compose up -d

# 访问服务
# API: http://localhost:8080
# UI:  http://localhost:3000
```

### 二进制部署

```bash
# 下载
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-linux-amd64.tar.gz
tar -xzf tlcpchan-linux-amd64.tar.gz

# 启动
./tlcpchan

# 后台运行
./tlcpchan -daemon
```

### 从源码构建

```bash
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan

# 构建主程序
cd tlcpchan && go build -o bin/tlcpchan .

# 构建前端
cd ../tlcpchan-ui/web && npm install && npm run build
cd .. && go build -o bin/tlcpchan-ui .
```

## 使用示例

### TLCP 服务端代理

将 TLCP 加密流量解密后转发到后端服务：

```yaml
# config/config.yaml
instances:
  - name: "tlcp-server"
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
      cipher_suites:
        - "ECDHE_SM4_GCM_SM3"
```

### 客户端代理

将明文流量加密后连接 TLCP/TLS 服务：

```yaml
instances:
  - name: "tlcp-client"
    type: "client"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":9000"
    target: "tlcp-server.example.com:443"
    enabled: true
    server_ca:
      - "ca-sm2"
```

### 协议自动检测

同一端口支持 TLCP 和 TLS 客户端：

```yaml
instances:
  - name: "auto-server"
    type: "server"
    protocol: "auto"
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
```

## 项目结构

```
tlcpchan/
├── tlcpchan/          # 主程序
│   ├── cmd/           # 入口
│   ├── config/        # 配置管理
│   ├── cert/          # 证书管理
│   ├── proxy/         # 代理核心
│   ├── instance/      # 实例管理
│   ├── controller/    # API 控制器
│   └── stats/         # 流量统计
├── tlcpchan-cli/      # CLI 工具
├── tlcpchan-ui/       # Web UI
│   └── web/           # Vue3 前端
├── docs/              # 文档
├── Dockerfile
└── docker-compose.yml
```

## 文档

- [使用指南](docs/README.md) - 详细使用说明
- [安装指南](docs/installation.md) - 安装部署指南
- [API 文档](docs/api.md) - RESTful API 接口
- [证书管理](docs/certificates.md) - 证书配置说明
- [配置示例](docs/config-examples.md) - 常用配置示例

## 代理模式

| 模式 | 说明 | 典型场景 |
|------|------|----------|
| server | TLCP/TLS → TCP | 后端服务国密改造 |
| client | TCP → TLCP/TLS | 访问国密服务 |
| http-server | HTTPS → HTTP | HTTP 服务国密化 |
| http-client | HTTP → HTTPS | 客户端国密适配 |

## 认证方式

| 方式 | 说明 | 适用场景 |
|------|------|----------|
| none | 不验证证书 | 测试环境 |
| one-way | 验证服务端证书 | 常规场景 |
| mutual | 双向证书验证 | 高安全要求 |

## CLI 使用

```bash
# 实例管理
tlcpchan-cli instance list
tlcpchan-cli instance start <name>
tlcpchan-cli instance stop <name>

# 证书管理
tlcpchan-cli cert list
tlcpchan-cli cert reload

# 系统信息
tlcpchan-cli system info
tlcpchan-cli system health
```

## API 示例

```bash
# 健康检查
curl http://localhost:8080/api/v1/system/health

# 获取实例列表
curl http://localhost:8080/api/v1/instances

# 创建实例
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{"name":"test","type":"server","protocol":"tlcp","listen":":443","target":"127.0.0.1:8080"}'

# 启动实例
curl -X POST http://localhost:8080/api/v1/instances/test/start
```

## 技术栈

- **后端**: Go 1.21+, [gotlcp](https://github.com/Trisia/gotlcp)
- **前端**: Vue 3, TypeScript, Element Plus
- **协议**: TLCP 1.1, TLS 1.0-1.3
- **算法**: SM2/SM3/SM4, RSA/ECDSA/AES

## 开发

```bash
# 运行测试
cd tlcpchan && go test ./...

# 构建所有组件
go build -o bin/tlcpchan ./tlcpchan
go build -o bin/tlcpchan-cli ./tlcpchan-cli
go build -o bin/tlcpchan-ui ./tlcpchan-ui
```

## 许可证

[Apache License 2.0](LICENSE)

## 致谢

- [gotlcp](https://github.com/Trisia/gotlcp) - TLCP 协议 Go 实现
- [gmsm](https://github.com/emmansun/gmsm) - 国密算法库

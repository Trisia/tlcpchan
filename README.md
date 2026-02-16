# TLCP Channel

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![TLCP](https://img.shields.io/badge/TLCP-1.1-green.svg)](https://github.com/Trisia/gotlcp)

**该项目目前处于开发中...**

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

## 技术栈

- **后端**: Go 1.21+, [gotlcp](https://github.com/Trisia/gotlcp)
- **前端**: Vue 3, TypeScript, Element Plus
- **协议**: TLCP 1.1, TLS 1.0-1.3
- **算法**: SM2/SM3/SM4, RSA/ECDSA/AES

## 文档

- [设计文档](docs/design.md) - 系统架构设计和技术方案
- [编译部署运行指南](docs/build-deploy.md) - 详细的安装、编译、部署说明
- [使用指南](docs/README.md) - 详细使用说明
- [API 文档](docs/api.md) - RESTful API 接口
- [证书管理](docs/certificates.md) - 证书配置说明
- [配置示例](docs/config-examples.md) - 常用配置示例

## 快速开始

详细的安装部署指南请参考 [编译部署运行指南](docs/build-deploy.md)。

### Docker 部署（推荐）

#### 使用 Docker Hub 官方镜像

```bash
docker pull trisia/tlcpchan:latest
docker run -d \
  -p 30080:30080 \
  -p 30000:30000 \
  -p 30443:30443 \
  -v tlcpchan-config:/etc/tlcpchan/config \
  -v tlcpchan-certs:/etc/tlcpchan/certs \
  -v tlcpchan-logs:/etc/tlcpchan/logs \
  --name tlcpchan \
  trisia/tlcpchan:latest
```

#### 使用 docker-compose

```bash
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan
docker compose up -d
```

#### 自行构建 Docker 镜像

```bash
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan
docker build -t tlcpchan:latest .
```

### 服务访问

- API: http://localhost:30080
- UI:  http://localhost:30000


## 许可证

[Apache License 2.0](LICENSE)

## 致谢

- [gotlcp](https://github.com/Trisia/gotlcp) - TLCP 协议 Go 实现
- [gmsm](https://github.com/emmansun/gmsm) - 国密算法库

# TLCP Channel

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![TLCP](https://img.shields.io/badge/TLCP-1.1-green.svg)](https://github.com/Trisia/gotlcp)

## 一句话介绍

TLCP Channel 让你在 30 秒内完成国密改造，无需修改现有应用。一款功能强大的 TLCP/TLS 协议代理工具，支持双协议并行工作，基于国密算法实现安全通信。

![Web 管理界面](docs/img/README-dashboard.png)

*仪表板展示实时连接数、流量统计和实例状态*

## 快速上手

### 30 秒体验

```bash
# 1. 启动服务
./tlcpchan --config config.yaml

# 2. 访问 Web 管理界面
http://localhost:20080

# 3. 完成国密改造！
```

就这么简单，无需修改任何现有应用代码。

---

## 功能特性

- **双协议支持** - 同时支持 TLCP 1.1 和 TLS 1.0-1.3 协议
- **自动协议检测** - 同一端口自动识别 TLCP/TLS 客户端
- **多种代理模式** - 服务端代理、客户端代理、HTTP 代理
- **国密算法** - 支持 SM2/SM3/SM4 国密密钥库（包含 TLCP 1.1 的 ECC 证书）
- **灵活认证** - 支持单向认证、双向认证
- **Web 管理界面** - Vue3 + Element Plus 现代化管理界面
- **RESTful API** - 完整的 API 接口支持
- **命令行工具** - tlcpchan-cli 命令行管理工具
- **证书热更新** - 无需重启即可更新证书
- **流量统计** - 实时连接数、流量、延迟统计

## 代理模式

| 模式 | 说明 | 典型场景 |
|------|------|----------|
| server | TLCP/TLS → TCP | 后端服务国密改造，将现有 TCP 服务包装为国密服务 |
| client | TCP → TLCP/TLS | 访问国密服务，让普通应用连接国密服务 |
| http-server | HTTPS → HTTP | HTTP 服务国密化，将 Web 服务升级为 HTTPS（支持国密） |
| http-client | HTTP → HTTPS | 客户端国密适配，让 HTTP 客户端访问国密 HTTPS 服务 |

## 认证方式

| 方式 | 说明 | 适用场景 |
|------|------|----------|
| none | 不验证证书 | 测试环境、开发环境 |
| one-way | 验证服务端证书 | 常规生产环境，客户端验证服务端身份 |
| mutual | 双向证书验证 | 高安全要求场景，双向身份验证 |

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan

# 构建后端服务
cd tlcpchan
go build -o tlcpchan ./cmd/tlcpchan

# 构建命令行工具
cd ../tlcpchan-cli
go build -o tlcpchan-cli



# 构建前端（可选）
cd ../tlcpchan-ui
npm install
npm run build
```

### 使用预编译版本（待提供）

## 快速启动

### 1. 创建配置文件

创建 `config.yaml`：

```yaml
server:
  address: ":20080"  # Web 管理界面端口

instances:
  - name: "我的 TLCP 服务"
    mode: "server"
    listen: ":8443"   # 监听端口
    target: "localhost:8080"  # 目标服务
    protocol: "tlcp"  # 协议类型：tlcp 或 tls
    auth: "one-way"   # 认证方式：none, one-way, mutual
    keystore: "/path/to/keystore.p12"  # 密钥库路径
    keystorePassword: "your-password"  # 密钥库密码
```

### 2. 启动服务

```bash
./tlcpchan --config config.yaml
```

### 3. 访问管理界面

打开浏览器访问：`http://localhost:20080`

## 配置说明

### 代理实例配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| name | string | 实例名称，用于标识和管理 |
| mode | string | 代理模式：server/client/http-server/http-client |
| listen | string | 监听地址，如 ":8443" |
| target | string | 目标地址，如 "localhost:8080" |
| protocol | string | 协议类型：tlcp 或 tls |
| auth | string | 认证方式：none/one-way/mutual |
| keystore | string | 密钥库文件路径（.p12 格式） |
| keystorePassword | string | 密钥库密码 |
| trustedCerts | []string | 信任的证书列表（双向认证时） |

### 密钥库管理

支持通过 Web 界面或 CLI 管理密钥库：

```bash
# 生成新的密钥库
tlcpchan-cli keystore generate -o keystore.p12 -p password

# 导入现有密钥库
tlcpchan-cli keystore import -i cert.pem -k key.pem -o keystore.p12 -p password

# 查看密钥库信息
tlcpchan-cli keystore info -i keystore.p12
```

## 使用示例

### 场景 1：后端服务国密改造

将现有的 TCP 服务（如 gRPC、自定义协议）升级为国密服务：

```yaml
instances:
  - name: "gRPC 服务国密化"
    mode: "server"
    listen: ":8443"
    target: "localhost:8080"  # 原 gRPC 服务
    protocol: "tlcp"
    auth: "one-way"
    keystore: "server.p12"
    keystorePassword: "password"
```

客户端连接 `8443` 端口即可使用国密通信，无需修改原有 gRPC 服务。

### 场景 2：HTTP 服务国密化

将 Web 服务升级为支持国密的 HTTPS 服务：

```yaml
instances:
  - name: "Web 服务国密化"
    mode: "http-server"
    listen: ":8443"
    target: "localhost:80"  # 原 HTTP 服务
    protocol: "tlcp"
    auth: "mutual"
    keystore: "server.p12"
    keystorePassword: "password"
    trustedCerts:
      - "client-cert.pem"
```

### 场景 3：访问国密服务

让普通应用访问国密服务：

```yaml
instances:
  - name: "访问国密 gRPC"
    mode: "client"
    listen: ":8080"  # 普通应用连接此端口
    target: "grpc-server:8443"  # 国密 gRPC 服务
    protocol: "tlcp"
    auth: "one-way"
    keystore: "client.p12"
    keystorePassword: "password"
    trustedCerts:
      - "server-cert.pem"
```

## 常见问题

### Q: TLCP 和 TLS 有什么区别？

TLCP（Transport Layer Cryptography Protocol）是中国国家密码管理局制定的传输层密码协议，基于国密算法（SM2/SM3/SM4）。TLS（Transport Layer Security）是国际标准协议，通常使用 RSA/AES 算法。TLCP Channel 同时支持两种协议，可以满足不同场景需求。

### Q: 支持 Windows 吗？

支持。TLCP Channel 使用 Go 开发，支持 Linux、Windows、macOS 等多平台。

### Q: 如何生成国密证书？

可以使用 Web 管理界面生成证书，或使用 `tlcpchan-cli` 命令行工具：

```bash
tlcpchan-cli cert generate -o cert.pem -k key.pem
```

也可以使用其他国密工具生成证书，然后导入到密钥库中。

### Q: 性能如何？

TLCP Channel 使用高效的 Go 实现和零拷贝技术，性能接近原生连接。在生产环境中，单实例可以轻松处理数千并发连接。

### Q: 如何部署到生产环境？

建议：
1. 使用 systemd 或 Docker 进行进程管理
2. 配置日志轮转
3. 启用访问控制
4. 定期备份密钥库和配置文件

## 技术栈

- **后端**: Go 1.26+, [gotlcp](https://github.com/Trisia/gotlcp)
- **前端**: Vue 3, TypeScript, Element Plus, Vite
- **协议**: TLCP 1.1, TLS 1.0-1.3
- **算法**: SM2/SM3/SM4, RSA/ECDSA/AES
- **配置**: YAML
- **API**: RESTful

## 相关资源

- **文档**: [详细使用文档](docs/)
- **UI 用户手册**: [Web 管理界面使用指南](tlcpchan-ui/docs/README.md)
- **API 文档**: [RESTful API 文档](docs/api.md)
- **示例**: [配置示例](examples/)
- **Demo**: [在线演示](https://demo.tlcpchan.com) (待提供)

## 致谢

- [gotlcp](https://github.com/Trisia/gotlcp) - TLCP 协议 Go 实现
- [gmsm](https://github.com/emmansun/gmsm) - 国密算法库
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 UI 组件库

## 许可证

本项目采用 [Apache 2.0](LICENSE) 许可证开源。

## 联系方式

- **项目主页**: https://github.com/Trisia/tlcpchan
- **问题反馈**: https://github.com/Trisia/tlcpchan/issues
- **讨论区**: https://github.com/Trisia/tlcpchan/discussions

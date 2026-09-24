# TLCP Channel

<div align="center">
  <img src="icon.png" width="128" alt="TLCP Channel Logo">
</div>

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![TLCP](https://img.shields.io/badge/TLCP-1.1-green.svg)](https://github.com/Trisia/gotlcp)
[![Documentation](https://pkg.go.dev/badge/github.com/Trisia/tlcpchan.svg)](https://pkg.go.dev/github.com/Trisia/tlcpchan)
[![Release](https://img.shields.io/github/release/Trisia/tlcpchan/all.svg)](https://github.com/Trisia/tlcpchan/releases)
[![Linux](https://img.shields.io/badge/Linux-FCC634?style=flat&logo=linux&logoColor=black)]()
[![Windows](https://img.shields.io/badge/Windows-0078D6?style=flat&logo=windows&logoColor=white)]()
[![macOS](https://img.shields.io/badge/macOS-000000?style=flat&logo=apple&logoColor=white)]()
[![x86_64](https://img.shields.io/badge/x86__64-blue.svg)]()
[![ARM64](https://img.shields.io/badge/ARM64-green.svg)]()
[![LoongArch64](https://img.shields.io/badge/LoongArch64-orange.svg)]()
[![MCP AI Agent](https://img.shields.io/badge/MCP%20AI%20Agent-Supported-purple.svg)](docs/mcp-usage-guide.md)



## 介绍

TLCP Channel 传输通道国密改造，无需修改现有应用。一款功能强大的 TLCP/TLS 协议代理工具，支持双协议在同一个端口并行工作，基于国密算法实现安全通信。

> ⚠️ **重要提示**：使用本项目前请先阅读 [免责声明](DISCLAIMER.md)

![demo](docs/img/demo.gif)


## 功能特性

- **双协议支持** - 同时支持 TLCP 1.1 和 TLS 1.0-1.3 协议
- **自动协议检测** - 同一端口自动识别 TLCP/TLS 客户端
- **多种代理模式** - 服务端代理、客户端代理、HTTP 代理
- **国密算法** - 支持 SM2/SM3/SM4 国密密钥库（包含 TLCP 1.1 的 ECC 证书）
- **传输通道身份认证** - 支持单向认证、双向认证
- **Web 管理界面** - Vue3 + Element Plus 现代化管理界面
- **RESTful API** - 完整的 API 接口支持
- **MCP 协议支持** - 通过 Model Context Protocol 支持 AI 助手集成
- **命令行工具** - tlcpchan-cli 命令行管理工具
- **证书热更新** - 无需重启即可更新证书
- **流量统计** - 实时连接数、流量、延迟统计



## 快速试用

### 使用 AI Agent 安装（推荐）

让 AI Agent 自动完成安装配置，无需手动执行命令。



将以下提示语复制并粘贴给您的 AI Agent（Claude Code、OpenCode、Cursor 等）：

```
使用 AI Agent 为 TLCP Channel 项目安装并配置，请遵循以下指南：
https://raw.githubusercontent.com/Trisia/tlcpchan/main/docs/guide/installation.md
```

或查看详细的 [Agent 安装指南](docs/guide/installation.md)。


### 对于人类用户 一键安装（手动）

#### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/Trisia/tlcpchan/main/install.sh | sudo bash
```

安装完成后，Linux使用 systemd 启动服务：

```bash
sudo systemctl start tlcpchan
```

MACOS 用户请到工作目录中运行启动

#### Windows

以管理员身份运行 PowerShell：

```powershell
irm https://raw.githubusercontent.com/Trisia/tlcpchan/main/install.ps1 | iex
```

安装完成后，直接运行：

```powershell
cd "C:\Program Files\TLCP Channel"
.\tlcpchan.exe
```

### 使用 Docker 快速启动

```bash
# 拉取镜像并启动服务
docker run -d \
  --name tlcpchan \
  --restart unless-stopped \
  -p 20080:20080 \
  -p 20443:20443 \
  -v tlcpchan-keystores:/etc/tlcpchan/keystores \
  -v tlcpchan-logs:/etc/tlcpchan/logs \
  tlcpchan/tlcpchan:latest
```

访问服务：

- **Web 管理界面**: http://localhost:20080
- **服务状态**: `docker logs -f tlcpchan`

从 GitHub Releases 下载最新版本安装包和二进制程序：[https://github.com/Trisia/tlcpchan/releases](https://github.com/Trisia/tlcpchan/releases)


## 相关文档

- **免责声明**: [免责声明](DISCLAIMER.md)
- **设计文档**: [详细设计文档](docs/design.md)
- **UI 用户手册**: [Web 管理界面使用指南](tlcpchan-ui/README.md)
- **CLI 用户手册**: [CLI 管理使用指南](tlcpchan-cli/README.md)
- **MCP 使用指南**: [Model Context Protocol 使用指南](docs/mcp-usage-guide.md)

## 代理模式

| 模式 | 说明 | 典型场景 |
|------|------|----------|
| server | TLCP/TLS → TCP | 后端服务国密改造，将现有 TCP 服务包装为国密服务 |
| client | TCP → TLCP/TLS | 访问国密服务，让普通应用连接国密服务 |
| http-server | HTTPS → HTTP | HTTP 服务国密化，将 Web 服务升级为 HTTPS（支持国密） |
| http-client | HTTP → HTTPS | 客户端国密适配，让 HTTP 客户端访问国密 HTTPS 服务 |


详细使用方法请参考 [MCP 使用指南](docs/mcp-usage-guide.md)。


## 系统适配

| 操作系统 | 支持的架构（CPU厂家） | 支持说明 |
|---------|---------------------|---------|
| **统信UOS** 20 | **飞腾ARM**、**龙芯LoongArch64**、x86_64 (Intel/AMD) | ✓ 国产化环境完整适配 |
| **银河麒麟** V10 | **飞腾ARM**、**龙芯LoongArch64**、x86_64 (Intel/AMD) | ✓ 国产化环境完整适配 |
| Ubuntu 18.04+ | x86_64 (Intel/AMD)、ARM64 (ARM) | ✓ |
| CentOS 7+ | x86_64 (Intel/AMD)、ARM64 (ARM) | ✓ |
| Windows 10+ | x86_64 (Intel/AMD) | ✓ |
| macOS 12 | x86_64 (Intel)、ARM64 (Apple Silicon) | ✓ |

下载最新版本安装包和二进制程序：

- [https://github.com/Trisia/tlcpchan/releases](https://github.com/Trisia/tlcpchan/releases)

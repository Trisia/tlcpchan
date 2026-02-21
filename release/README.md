# TLCP Channel 部署安装手册

## 1. 平台架构支持情况

TLCP Channel 支持多种平台和架构，满足不同部署环境的需求。项目的编译基于 Go 语言，因此支持平台以 Go 编译时支持的最低操作系统版本为准。

### 支持平台表格

| 操作系统 | 架构 | 打包格式 | 适用场景 |
|---------|------|---------|----------|
| Linux | amd64 | tar.gz, deb, rpm | 服务器部署 |
| Linux | arm64 | tar.gz, deb, rpm | ARM主机 |
| Linux | loong64 | tar.gz, deb, rpm | 龙芯平台 |
| macOS | amd64 | tar.gz, app | Mac 桌面应用 |
| macOS | arm64 | tar.gz, app | Apple Silicon |
| Windows | amd64 | zip | Windows 服务器/桌面 |

### 国产操作系统支持

- **RPM 包支持**：银河麒麟、统信 UOS、中标麒麟、红旗 Linux
- **DEB 包支持**：统信 UOS、深度 Deepin

## 2. 安装文件清单

每个安装包包含以下核心文件和目录：

```
tlcpchan               # 核心服务可执行程序
tlcpchan-cli           # 命令行管理工具
ui/                    # Web 管理界面静态文件
├── index.html
├── assets/
└── version.txt
rootcerts/             # 预置国密CA信任证书
├── ca-sm2-root.crt
├── 上海市数字证书认证中心有限公司_CN=SHECA SM2,O=UniTrust,C=CN.pem
└── ... 其他50+个证书
keystores/             # 用户证书存储目录（初始为空）
logs/                  # 日志目录（初始为空）
config.yaml            # 配置文件（首次启动时生成）
```

## 3. 关键路径说明

安装完成后，系统会创建以下关键目录结构：

### Linux 系统路径

```
/etc/tlcpchan/                 # 主工作目录
├── tlcpchan                   # 核心服务程序
├── tlcpchan-cli               # CLI工具
├── config.yaml                # 配置文件
├── keystores/                 # 证书存储
├── rootcerts/                 # 信任证书
├── logs/                      # 日志文件
├── ui/                        # 前端文件
└── .tlcpchan-initialized      # 初始化标记
```

### 符号链接

系统会自动创建以下符号链接，方便命令行使用：

```
/usr/bin/tlcpchan → /etc/tlcpchan/tlcpchan
/usr/bin/tlcpchan-cli → /etc/tlcpchan/tlcpchan-cli
/usr/bin/tlcpc → /etc/tlcpchan/tlcpchan-cli
```

## 4. 系统资源需求

TLCP Channel 设计为轻量级代理服务，对系统资源要求较低。实际资源消耗根据运行时负载和配置调整。

### 资源需求参考

- **内存**：空闲时 50-100 MB，高负载时根据连接数增加
- **CPU**：单核心基本够用，高并发建议多核心
- **存储**：安装包体积的 1.5-2 倍（包含证书和日志增长）
- **网络**：无特殊要求，根据代理流量调整

### 端口需求

- **20080**：API 服务和 Web UI（必须开放）
- **20443**：默认代理端口（必须开放）
- **其他端口**：根据 `instances` 配置自定义（需手动开放）

## 5. 安装方式详细说明

### 5.1 Linux 系统包安装（推荐）

对于 Linux 系统，推荐使用系统包管理器安装，便于版本管理和自动更新。

#### Debian/Ubuntu 系统（.deb 包）

下载对应平台的 deb 包后，执行以下安装命令：

```bash
sudo dpkg -i tlcpchan_1.0.0_linux_amd64.deb
```

如果安装过程中出现依赖缺失问题，执行以下命令解决：

```bash
sudo apt-get install -f
```

**预期输出**：
```
正在选中未选择的软件包 tlcpchan。
(正在读取数据库 ... 系统当前共安装有 123456 个文件和目录。)
准备解压 tlcpchan_1.0.0_linux_amd64.deb  ...
正在解压 tlcpchan (1.0.0) ...
正在设置 tlcpchan (1.0.0) ...
正在创建系统用户 tlcpchan...
正在创建目录结构...
正在复制文件...
正在设置权限...
正在创建符号链接...
正在配置 systemd 服务...
安装完成。
```

#### RHEL/CentOS 系统（.rpm 包）

下载对应平台的 rpm 包后，执行以下安装命令：

```bash
sudo rpm -i tlcpchan_1.0.0_linux_amd64.rpm
```

### 5.2 二进制压缩包安装（通用方法）

如果系统包不可用，可以使用二进制压缩包安装。此方法适用于所有支持平台。

#### Linux/macOS 系统

下载对应平台的压缩包后，首先解压：

```bash
tar -xzf tlcpchan_1.0.0_linux_amd64.tar.gz
cd tlcpchan_1.0.0_linux_amd64
```

查看解压后的文件内容：

```bash
ls -la
```

**预期输出**：
```
-rwxr-xr-x 1 user group 15M Jan  1 12:00 tlcpchan
-rwxr-xr-x 1 user group 2.5M Jan  1 12:00 tlcpchan-cli
drwxr-xr-x 4 user group 4096 Jan  1 12:00 ui
drwxr-xr-x 2 user group 4096 Jan  1 12:00 rootcerts
drwxr-xr-x 2 user group 4096 Jan  1 12:00 keystores
drwxr-xr-x 2 user group 4096 Jan  1 12:00 logs
-rw-r--r-- 1 user group  100 Jan  1 12:00 config.yaml
```

#### 手动部署步骤

创建必要的目录结构：

```bash
sudo mkdir -p /etc/tlcpchan
sudo mkdir -p /etc/tlcpchan/keystores
sudo mkdir -p /etc/tlcpchan/logs
```

复制程序和资源文件：

```bash
sudo cp tlcpchan tlcpchan-cli /etc/tlcpchan/
sudo cp -r ui rootcerts /etc/tlcpchan/
```

设置程序执行权限：

```bash
sudo chmod +x /etc/tlcpchan/tlcpchan
sudo chmod +x /etc/tlcpchan/tlcpchan-cli
```

创建命令行符号链接：

```bash
sudo ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
sudo ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
sudo ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc
```

### 5.3 Docker 安装

Docker 安装提供容器化部署方案，便于环境隔离和快速部署。

#### Docker Compose 方式（推荐）

创建 `docker-compose.yml` 文件，内容如下：

```yaml
version: '3.8'
services:
  tlcpchan:
    image: tlcpchan:latest
    container_name: tlcpchan
    restart: unless-stopped
    ports:
      - "20080:20080"
      - "20443:20443"
    volumes:
      - tlcpchan-keystores:/etc/tlcpchan/keystores
      - tlcpchan-logs:/etc/tlcpchan/logs

volumes:
  tlcpchan-keystores:
    driver: local
  tlcpchan-logs:
    driver: local
```

启动服务：

```bash
docker-compose up -d
```

查看容器运行状态：

```bash
docker-compose ps
```

**预期输出**：
```
      Name                    Command               State            Ports
--------------------------------------------------------------------------------
tlcpchan   tlcpchan                    Up      0.0.0.0:20080->20080/tcp, 0.0.0.0:20443->20443/tcp
```

#### 直接 Docker 命令

使用 docker run 命令直接运行容器：

```bash
docker run -d \
  --name tlcpchan \
  -p 20080:20080 \
  -p 20443:20443 \
  -v tlcpchan-keystores:/etc/tlcpchan/keystores \
  -v tlcpchan-logs:/etc/tlcpchan/logs \
  tlcpchan:latest
```

### 5.4 Windows 安装

Windows 版本支持可执行文件运行，不支持安装为 Windows 服务。

#### 压缩包安装

解压下载的压缩包到指定目录：

```batch
mkdir "C:\Program Files\TLCP Channel"
tar -xzf tlcpchan_1.0.0_windows_amd64.zip -C "C:\Program Files\TLCP Channel"
```

运行程序：

```batch
cd "C:\Program Files\TLCP Channel"
.\tlcpchan.exe
```

#### 手动添加到 PATH

将 TLCP Channel 的安装路径添加到系统环境变量，方便命令行调用：

```batch
setx PATH "%PATH%;C:\Program Files\TLCP Channel" /M
```

## 6. 服务控制和管理

### 6.1 Linux systemd 服务管理

Linux 系统安装后会自动配置为 systemd 服务，便于统一管理。

#### 启动服务

使用 systemctl 命令启动 TLCP Channel 服务：

```bash
sudo systemctl start tlcpchan
```

**预期输出**：
无输出表示启动成功，或：
```
Job for tlcpchan.service has begun.
```

#### 查看服务状态

查看服务当前运行状态：

```bash
sudo systemctl status tlcpchan
```

**预期输出**：
```
● tlcpchan.service - TLCP Channel Proxy Service
   Loaded: loaded (/usr/lib/systemd/system/tlcpchan.service; enabled; vendor preset: enabled)
   Active: active (running) since Mon 2024-01-01 12:00:00 CST; 10s ago
 Main PID: 12345 (tlcpchan)
    Tasks: 10 (limit: 4915)
   Memory: 45.2M
   CGroup: /system.slice/tlcpchan.service
           └─12345 /etc/tlcpchan/tlcpchan
```

#### 设置开机自启

配置服务在系统启动时自动运行：

```bash
sudo systemctl enable tlcpchan
```

#### 停止服务

停止正在运行的服务：

```bash
sudo systemctl stop tlcpchan
```

### 6.2 日志管理

TLCP Channel 采用日志轮转策略，避免日志文件无限增长占用磁盘空间。默认配置下，日志文件会自动轮转、压缩和清理。

#### 日志配置位置

日志配置在 `/etc/tlcpchan/config.yaml` 中的 `server.log` 部分：

```yaml
server:
  log:
    level: info                    # 日志级别: debug, info, warn, error
    file: ./logs/tlcpchan.log      # 日志文件路径
    max-size: 100                  # 单个日志文件最大大小(MB)
    max-backups: 5                 # 最多保留的日志文件数
    max-age: 30                    # 日志文件保留天数
    compress: true                 # 是否压缩旧日志
    enabled: true                  # 是否启用文件日志
```

#### 日志轮转策略说明

1. **大小限制**：当日志文件达到 100MB 时，会自动轮转
2. **数量限制**：最多保留 5 个历史日志文件
3. **时间限制**：超过 30 天的日志文件自动删除
4. **压缩处理**：历史日志文件自动压缩为 .gz 格式

#### 日志文件示例

```
/etc/tlcpchan/logs/
├── tlcpchan.log          # 当前活跃日志文件
├── tlcpchan.log.1.gz     # 最近的历史日志（已压缩）
├── tlcpchan.log.2.gz
├── tlcpchan.log.3.gz
└── tlcpchan.log.4.gz
```

#### 查看日志

查看实时日志输出：

```bash
tail -f /etc/tlcpchan/logs/tlcpchan.log
```

查看历史日志文件内容：

```bash
zcat /etc/tlcpchan/logs/tlcpchan.log.1.gz | head -20
```

#### 日志内容示例

```
2024-01-01T12:00:00+08:00 [INFO] proxy: 代理服务启动 address=:20443
2024-01-01T12:00:05+08:00 [INFO] api: API服务启动 address=:20080
2024-01-01T12:00:10+08:00 [INFO] init: 初始化完成，生成默认配置文件
2024-01-01T12:01:00+08:00 [INFO] connection: 新连接建立 remote=192.168.1.100:54321
```

## 7. 初始化和配置

### 7.1 首次启动初始化

首次启动 TLCP Channel 时，系统会自动执行初始化流程，包括生成配置文件、创建证书目录等。

#### 自动执行步骤

1. 检查配置文件是否存在，不存在则生成默认配置
2. 创建必要的目录结构（keystores、logs、ui 等）
3. 生成初始配置文件 `/etc/tlcpchan/config.yaml`
4. 标记初始化完成，创建 `.tlcpchan-initialized` 文件

#### 启动命令

使用 systemd 服务启动：

```bash
sudo systemctl start tlcpchan
```

或直接运行可执行文件：

```bash
sudo /etc/tlcpchan/tlcpchan
```

#### 预期输出（控制台）

```
[INFO] 正在初始化 TLCP Channel...
[INFO] 生成默认配置文件: /etc/tlcpchan/config.yaml
[INFO] 初始化完成，启动服务...
[INFO] API服务启动 address=:20080
[INFO] 代理服务启动 address=:20443
[INFO] Web UI访问地址: http://localhost:20080/ui/
```

### 7.2 配置检查

安装完成后，可以使用以下命令验证安装是否成功。

#### 版本检查

查看当前安装的 TLCP Channel 版本：

```bash
tlcpchan -version
```

**预期输出**：
```
TLCP Channel version 1.0.0
```

#### 配置检查

验证配置文件的语法和内容是否正确：

```bash
tlcpchan-cli config validate
```

**预期输出**：
```
配置文件验证通过: /etc/tlcpchan/config.yaml
```

## 8. 故障排除

### 8.1 服务健康检查

TLCP Channel 提供健康检查接口，用于监控服务状态。

#### 健康检查命令

使用 curl 调用健康检查 API：

```bash
curl http://localhost:20080/api/system/health
```

**预期输出**：
```
{"status": "ok", "timestamp": "2024-01-01T12:00:00+08:00"}
```

#### Web UI 检查

检查 Web UI 是否正常响应：

```bash
curl http://localhost:20080/ui/
```

**预期输出**：
```
<html>...</html>  # 返回 HTML 页面内容
```

### 8.2 常见问题

#### 端口被占用

检查 20080 或 20443 端口是否被其他程序占用：

```bash
sudo netstat -tlnp | grep -E '20080|20443'
```

#### 证书加载失败

检查证书文件的权限是否正确：

```bash
ls -la /etc/tlcpchan/keystores/
```

查看与证书相关的错误日志：

```bash
grep -i "cert" /etc/tlcpchan/logs/tlcpchan.log
```

#### 服务无法启动

查看详细的错误信息和日志：

```bash
sudo journalctl -u tlcpchan -n 50
```

## 9. 卸载和清理

### 9.1 Linux 系统卸载

对于使用系统包安装的 TLCP Channel，推荐使用包管理器卸载。对于手动安装，需要手动清理相关文件。

#### Debian/Ubuntu 卸载

使用 dpkg 命令卸载 TLCP Channel：

```bash
sudo dpkg -r tlcpchan
```

**预期输出**：
```
正在删除 tlcpchan (1.0.0) ...
正在停止 tlcpchan 服务...
正在删除文件...
正在删除符号链接...
正在删除 systemd 配置...
tlcpchan 已成功删除。
```

#### RHEL/CentOS 卸载

使用 rpm 命令卸载 TLCP Channel：

```bash
sudo rpm -e tlcpchan
```

#### 手动安装卸载步骤

首先停止服务并禁用自启动：

```bash
sudo systemctl stop tlcpchan
sudo systemctl disable tlcpchan
```

删除命令行符号链接：

```bash
sudo rm /usr/bin/tlcpchan /usr/bin/tlcpchan-cli /usr/bin/tlcpc
```

删除 systemd 服务配置：

```bash
sudo rm /usr/lib/systemd/system/tlcpchan.service
sudo systemctl daemon-reload
```

删除安装目录（可选，会删除所有数据）：

```bash
sudo rm -rf /etc/tlcpchan/
```

### 9.2 Docker 卸载

Docker 安装的 TLCP Channel 卸载相对简单，但需要注意数据备份。

#### 卸载容器

使用 Docker Compose 停止并删除容器：

```bash
docker-compose down
```

或使用 Docker 命令直接操作：

```bash
docker stop tlcpchan && docker rm tlcpchan
```

删除 Docker 镜像：

```bash
docker rmi tlcpchan:latest
```

删除数据卷（会删除所有数据，谨慎操作）：

```bash
docker volume rm tlcpchan-keystores tlcpchan-logs
```

#### 数据备份

在卸载前，建议先备份重要的证书数据：

```bash
docker cp tlcpchan:/etc/tlcpchan/keystores ./backup/keystores/
```

## 10. 未来扩展计划

为了提升部署便利性和企业级支持，计划未来将 TLCP Channel 发布到以下官方仓库：

### Linux 包管理仓库

- **yum/dnf**：RHEL、CentOS、Fedora
- **apt**：Debian、Ubuntu、统信 UOS
- **国产操作系统官方仓库**：银河麒麟、中标麒麟

### Docker 官方仓库

- **Docker Hub**：`tlcpchan/tlcpchan`
- **阿里云容器镜像服务**
- **华为云容器镜像仓库**

### Windows 渠道

- **Chocolatey 包管理器**
- **Scoop 包管理器**
- **Microsoft Store**（可选）

这些扩展将使用户能够使用熟悉的包管理器直接安装，无需手动下载和配置。

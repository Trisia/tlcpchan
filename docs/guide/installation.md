# TLCP Channel 安装指南

本指南为 LLM Agent 提供详细的 TLCP Channel 安装步骤。

## 概述

TLCP Channel 是一款功能强大的 TLCP/TLS 协议代理工具，支持双协议在同一个端口并行工作。

**安装方式**：
1. **包管理器安装**（Linux，推荐）：apt/yum/dnf/zypper
2. **二进制安装**（Linux/macOS）：tar.gz 压缩包
3. **Docker 安装**（跨平台）：Docker 容器
4. **源码编译**（开发者）：从源码构建

## 安装前检查

在开始安装前，必须检查以下环境：

### 1. 检查操作系统和架构

```bash
# 检测操作系统
os=$(uname -s)
case "$os" in
    Linux)
        echo "检测到 Linux 系统"
        ;;
    Darwin)
        echo "检测到 macOS 系统（需要特殊处理）"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        echo "检测到 Windows 系统（Git Bash 环境下）"
        ;;
    *)
        echo "不支持的操作系统: $os"
        exit 1
        ;;
esac

# 检测 CPU 架构
arch=$(uname -m)
case "$arch" in
    x86_64)
        echo "检测到 x86_64 架构"
        ;;
    aarch64|arm64)
        echo "检测到 ARM64 架构"
        ;;
    loongarch64)
        echo "检测到 LoongArch64 架构（国产 CPU）"
        ;;
    *)
        echo "不支持的架构: $arch"
        exit 1
        ;;
esac
```

### 2. 检查必要的命令

必须确保以下命令可用：
- Linux/macOS: `curl`, `tar`, `chmod`, `mkdir`
- Windows: `powershell`, `Invoke-RestMethod` 或 `curl`

### 3. 检查权限

**Linux/macOS**：需要 root 权限或使用 sudo
```bash
if [ "$EUID" -ne 0 ]; then
    echo "需要 root 权限，请使用 sudo"
    exit 1
fi
```

**Windows**：需要管理员权限的 PowerShell

### 4. 检查是否已安装

检查 `/etc/tlcpchan`（Linux/macOS）或 `C:\Program Files\TLCP Channel`（Windows）是否存在。

## 方式一：包管理器安装（Linux 推荐）

### 适用场景
- Ubuntu/Debian（apt）
- CentOS/RHEL（yum/dnf）
- openSUSE/SUSE（zypper）

### 安装步骤

#### 1. 获取最新版本

```bash
# 使用 GitHub API 获取最新版本
REPO_OWNER="Trisia"
REPO_NAME="tlcpchan"
GITHUB_API="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}"

version=$(curl -s "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
echo "最新版本: v${version}"
```

#### 2. 检测包管理器

```bash
detect_package_manager() {
    if command -v apt >/dev/null 2>&1; then
        echo "apt"
    elif command -v dnf >/dev/null 2>&1; then
        echo "dnf"
    elif command -v yum >/dev/null 2>&1; then
        echo "yum"
    elif command -v zypper >/dev/null 2>&1; then
        echo "zypper"
    else
        echo "none"
    fi
}

pkg_manager=$(detect_package_manager)
```

#### 3. 下载安装包

```bash
# 构建包 URL
os="linux"
arch=$(uname -m)
case "$arch" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    loongarch64) arch="loong64" ;;
esac

# 确定包格式
case "$pkg_manager" in
    apt)
        package_format="deb"
        ;;
    dnf|yum|zypper)
        package_format="rpm"
        ;;
    *)
        echo "不支持的包管理器"
        exit 1
        ;;
esac

# 构建下载 URL
filename="tlcpchan_${version}_${os}_${arch}.${package_format}"
download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${filename}"

# 下载包
echo "正在下载 ${filename}..."
curl -# -L -o "/tmp/${filename}" "$download_url"
```

#### 4. 安装包

```bash
case "$pkg_manager" in
    apt)
        apt install -y "/tmp/${filename}"
        ;;
    dnf)
        dnf install -y "/tmp/${filename}"
        ;;
    yum)
        yum install -y "/tmp/${filename}"
        ;;
    zypper)
        zypper install --non-interactive "/tmp/${filename}"
        ;;
esac
```

#### 5. 清理临时文件

```bash
rm -f "/tmp/${filename}"
```

### 安装后验证

```bash
# 验证安装
if command -v tlcpchan >/dev/null 2>&1; then
    echo "✅ TLCP Channel 安装成功"
    tlcpchan -version
else
    echo "❌ 安装验证失败"
    exit 1
fi
```

## 方式二：二进制安装（Linux/macOS）

### 适用场景
- 没有 apt/yum/dnf/zypper 的系统
- macOS 系统
- 需要自定义安装路径

### 安装步骤

#### 1. 获取最新版本

（同方式一，步骤1）

#### 2. 下载二进制包

```bash
# 检测操作系统
os=$(uname -s)
case "$os" in
    Linux)
        os="linux"
        ;;
    Darwin)
        os="darwin"
        ;;
    *)
        echo "不支持的操作系统: $os"
        exit 1
        ;;
esac

# 检测架构
arch=$(uname -m)
case "$arch" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    loongarch64) arch="loong64" ;;
esac

# 构建下载 URL
filename="tlcpchan_${version}_${os}_${arch}.tar.gz"
download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${filename}"

echo "正在下载 ${filename}..."
curl -# -L -o "/tmp/${filename}" "$download_url"
```

#### 3. 解压并安装

```bash
# 创建安装目录
INSTALL_DIR="/etc/tlcpchan"
mkdir -p "$INSTALL_DIR"
mkdir -p "${INSTALL_DIR}/keystores"
mkdir -p "${INSTALL_DIR}/logs"

# 解压到临时目录
tmp_extract_dir="${INSTALL_DIR}_extract"
rm -rf "$tmp_extract_dir"
mkdir -p "$tmp_extract_dir"

echo "正在解压安装包..."
tar -xzf "/tmp/${filename}" -C "$tmp_extract_dir"

# 移动文件到安装目录
if [ -d "${tmp_extract_dir}/tlcpchan" ]; then
    mv "${tmp_extract_dir}/tlcpchan"/* "$INSTALL_DIR/"
else
    mv "$tmp_extract_dir"/* "$INSTALL_DIR/" 2>/dev/null || true
fi

# 清理临时目录
rm -rf "$tmp_extract_dir"
rm -f "/tmp/${filename}"

# 设置可执行权限
chmod +x "${INSTALL_DIR}/tlcpchan"
chmod +x "${INSTALL_DIR}/tlcpchan-cli"
```

#### 4. 验证安装

```bash
# 验证安装
if "${INSTALL_DIR}/tlcpchan" -version; then
    echo "✅ TLCP Channel 安装成功"
else
    echo "❌ 安装验证失败"
    exit 1
fi
```

## 方式三：Docker 安装（跨平台）

### 适用场景
- 快速试用
- 容器化部署
- 不想修改系统文件

### 安装步骤

#### 1. 检查 Docker 是否安装

```bash
if ! command -v docker >/dev/null 2>&1; then
    echo "Docker 未安装，请先安装 Docker"
    exit 1
fi
```

#### 2. 拉取镜像

```bash
echo "正在拉取 TLCP Channel Docker 镜像..."
docker pull tlcpchan/tlcpchan:latest
```

#### 3. 启动容器

```bash
# 检查是否已存在容器
if docker ps -a | grep -q tlcpchan; then
    echo "容器已存在，正在删除旧容器..."
    docker rm -f tlcpchan
fi

# 启动容器
docker run -d \
  --name tlcpchan \
  --restart unless-stopped \
  -p 20080:20080 \
  -p 20443:20443 \
  -v tlcpchan-keystores:/etc/tlcpchan/keystores \
  -v tlcpchan-logs:/etc/tlcpchan/logs \
  tlcpchan/tlcpchan:latest
```

#### 4. 验证安装

```bash
# 等待服务启动
sleep 5

# 检查容器状态
if docker ps | grep -q tlcpchan; then
    echo "✅ TLCP Channel Docker 容器启动成功"
    echo ""
    echo "访问 Web 界面: http://localhost:20080"
else
    echo "❌ 容器启动失败"
    echo "查看日志: docker logs tlcpchan"
    exit 1
fi
```

## 方式四：源码编译（开发者）

### 适用场景
- 需要修改源码
- 需要最新的开发版本
- 贡献代码

### 安装前准备

#### 1. 安装 Go

```bash
# 检查 Go 版本（需要 1.26+）
if ! command -v go >/dev/null 2>&1; then
    echo "Go 未安装，请先安装 Go 1.26 或更高版本"
    exit 1
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
echo "检测到 Go 版本: $go_version"
```

#### 2. 安装 Node.js（用于构建 UI）

```bash
# 检查 Node.js 版本（需要 18+）
if ! command -v node >/dev/null 2>&1; then
    echo "Node.js 未安装，请先安装 Node.js 18 或更高版本"
    exit 1
fi

node_version=$(node --version)
echo "检测到 Node.js 版本: $node_version"
```

### 编译步骤

#### 1. 克隆仓库

```bash
# 克隆仓库（如果不存在）
if [ ! -d "tlcpchan" ]; then
    git clone https://github.com/Trisia/tlcpchan.git
    cd tlcpchan
else
    cd tlcpchan
    git pull origin main
fi
```

#### 2. 构建项目

```bash
# 使用项目自带的构建脚本
echo "正在构建项目..."
./build.sh

# 或手动构建
# cd tlcpchan && go build -o ../target/tlcpchan
# cd ../tlcpchan-ui && npm install && npm run build
# cd ../tlcpchan-cli && go build -o ../target/tlcpchan-cli
```

#### 3. 安装二进制文件

```bash
# 创建安装目录
INSTALL_DIR="/etc/tlcpchan"
mkdir -p "$INSTALL_DIR"
mkdir -p "${INSTALL_DIR}/keystores"
mkdir -p "${INSTALL_DIR}/logs

# 复制二进制文件
cp target/tlcpchan "$INSTALL_DIR/"
cp target/tlcpchan-cli "$INSTALL_DIR/"

# 设置可执行权限
chmod +x "${INSTALL_DIR}/tlcpchan"
chmod +x "${INSTALL_DIR}/tlcpchan-cli"
```

#### 4. 验证安装

```bash
# 验证安装
if "${INSTALL_DIR}/tlcpchan" -version; then
    echo "✅ TLCP Channel 编译安装成功"
else
    echo "❌ 安装验证失败"
    exit 1
fi
```

## 安装后配置

### 1. 配置文件

配置文件位于：
- Linux/macOS: `/etc/tlcpchan/config.yaml`
- Docker: 挂载到 `/etc/tlcpchan/config.yaml`

### 2. 启动服务

#### systemd（Linux）

```bash
# 启动服务
sudo systemctl start tlcpchan

# 设置开机自启
sudo systemctl enable tlcpchan

# 查看状态
sudo systemctl status tlcpchan

# 查看日志
sudo journalctl -u tlcpchan -f
```

#### 手动启动

```bash
# Linux/macOS
cd /etc/tlcpchan
./tlcpchan

# Docker
docker logs -f tlcpchan
```

### 3. 访问 Web 界面

打开浏览器访问：`http://localhost:20080`

默认端口：
- Web 管理界面: `20080`
- 代理服务端口: `20443`（可配置）

## 验证安装

### 健康检查

```bash
# 检查服务是否运行
curl http://localhost:20080/api/health

# 检查版本
curl http://localhost:20080/api/version
```

### 测试代理功能

1. 访问 Web 界面
2. 创建代理实例
3. 测试连接

## 卸载

### Linux/macOS

```bash
# 停止服务
sudo systemctl stop tlcpchan
sudo systemctl disable tlcpchan

# 删除文件
sudo rm -rf /etc/tlcpchan

# 删除 systemd 服务文件
sudo rm -f /etc/systemd/system/tlcpchan.service
sudo systemctl daemon-reload
```

### Docker

```bash
# 停止并删除容器
docker stop tlcpchan
docker rm tlcpchan

# 删除镜像
docker rmi tlcpchan/tlcpchan:latest

# 删除数据卷
docker volume rm tlcpchan-keystores tlcpchan-logs
```

## 常见问题

### 1. 权限错误

**问题**：`Permission denied`

**解决**：
- Linux/macOS：使用 `sudo` 执行安装命令
- Windows：以管理员身份运行 PowerShell

### 2. 下载失败

**问题**：无法下载安装包

**解决**：
- 检查网络连接
- 检查防火墙设置
- 尝试使用代理
- 手动下载后本地安装

### 3. 端口被占用

**问题**：端口 20080 或 20443 已被占用

**解决**：
- 修改配置文件中的端口
- 或停止占用端口的服务

### 4. 服务启动失败

**问题**：服务无法启动

**解决**：
```bash
# 查看日志
sudo journalctl -u tlcpchan -n 50

# 检查配置文件
sudo cat /etc/tlcpchan/config.yaml

# 手动启动查看详细错误
sudo /etc/tlcpchan/tlcpchan
```

## 获取帮助

- **文档**：https://github.com/Trisia/tlcpchan
- **Issue**：https://github.com/Trisia/tlcpchan/issues
- **MCP 使用指南**：https://github.com/Trisia/t/tlcpchan/blob/main/docs/mcp-usage-guide.md

## 总结

安装 TLCP Channel 的推荐顺序：
1. **Linux**：包管理器安装 → 二进制安装 → Docker 安装
2. **macOS**：二进制安装 → Docker 安装
3. **Windows**：Docker 安装
4. **开发者**：源码编译

安装完成后，访问 Web 界面 http://localhost:20080 开始使用。

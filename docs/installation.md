# 安装指南

本文档详细说明 TLCP Channel 在各平台的安装方法。

## 系统要求

### 操作系统

| 平台 | 最低版本 |
|------|----------|
| Linux | 内核 3.10+ |
| Windows | Windows 10 / Windows Server 2016+ |
| macOS | macOS 10.15+ |

### 硬件架构

| 架构 | 支持平台 |
|------|----------|
| x86_64 (amd64) | Linux, Windows, macOS |
| arm64 (aarch64) | Linux, Windows, macOS |
| loongarch64 | Linux |

### 资源要求

| 项目 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核 | 2 核+ |
| 内存 | 128 MB | 512 MB+ |
| 磁盘 | 50 MB | 1 GB+（含日志） |

## Linux 安装

### 方式一：DEB 包（Debian/Ubuntu）

```bash
# 下载 DEB 包
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan_1.0.0_amd64.deb

# 安装
sudo dpkg -i tlcpchan_1.0.0_amd64.deb

# 安装依赖（如有缺失）
sudo apt-get install -f

# 启动服务
sudo systemctl start tlcpchan
sudo systemctl enable tlcpchan
```

安装后文件位置：

| 文件 | 路径 |
|------|------|
| 可执行文件 | `/usr/bin/tlcpchan` |
| 配置文件 | `/etc/tlcpchan/config.yaml` |
| 证书目录 | `/var/lib/tlcpchan/certs/` |
| 日志目录 | `/var/log/tlcpchan/` |

### 方式二：RPM 包（RHEL/CentOS/Fedora）

```bash
# 下载 RPM 包
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-1.0.0.x86_64.rpm

# 安装
sudo rpm -i tlcpchan-1.0.0.x86_64.rpm

# 或使用 yum/dnf
sudo yum install ./tlcpchan-1.0.0.x86_64.rpm

# 启动服务
sudo systemctl start tlcpchan
sudo systemctl enable tlcpchan
```

### 方式三：TAR.GZ 压缩包

```bash
# 下载压缩包
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-linux-amd64.tar.gz

# 解压
tar -xzf tlcpchan-linux-amd64.tar.gz

# 安装到系统路径
sudo mv tlcpchan /usr/local/bin/

# 创建工作目录
sudo mkdir -p /opt/tlcpchan
sudo chown $USER:$USER /opt/tlcpchan
```

### 方式四：手动安装 systemd 服务

```bash
# 创建用户
sudo useradd -r -s /bin/false tlcpchan

# 创建目录
sudo mkdir -p /opt/tlcpchan /var/log/tlcpchan
sudo chown tlcpchan:tlcpchan /opt/tlcpchan /var/log/tlcpchan

# 创建 systemd 服务文件
sudo tee /etc/systemd/system/tlcpchan.service << 'EOF'
[Unit]
Description=TLCP Channel - TLCP/TLS Proxy Server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=tlcpchan
Group=tlcpchan
WorkingDirectory=/opt/tlcpchan
ExecStart=/usr/local/bin/tlcpchan
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable tlcpchan
sudo systemctl start tlcpchan
```

## Windows 安装

### 方式一：MSI 安装包

1. 下载 MSI 安装包：
   ```
   https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-1.0.0-windows-amd64.msi
   ```

2. 双击运行安装程序，按向导完成安装

3. 安装完成后，程序将自动注册为 Windows 服务

安装后文件位置：

| 文件 | 路径 |
|------|------|
| 可执行文件 | `C:\Program Files\TLCPChan\tlcpchan.exe` |
| 配置文件 | `C:\ProgramData\TLCPChan\config\config.yaml` |
| 证书目录 | `C:\ProgramData\TLCPChan\certs\` |

### 方式二：ZIP 压缩包

```powershell
# 使用 PowerShell 下载
Invoke-WebRequest -Uri "https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-windows-amd64.zip" -OutFile "tlcpchan.zip"

# 解压
Expand-Archive -Path tlcpchan.zip -DestinationPath C:\tlcpchan

# 进入目录
cd C:\tlcpchan

# 运行
.\tlcpchan.exe
```

### 注册为 Windows 服务

使用 NSSM（Non-Sucking Service Manager）：

```powershell
# 下载 NSSM
Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile "nssm.zip"
Expand-Archive -Path nssm.zip -DestinationPath C:\nssm

# 安装服务
C:\nssm\nssm-2.24\win64\nssm.exe install TLCPChan C:\tlcpchan\tlcpchan.exe
C:\nssm\nssm-2.24\win64\nssm.exe set TLCPChan AppDirectory C:\tlcpchan
C:\nssm\nssm-2.24\win64\nssm.exe set TLCPChan DisplayName "TLCP Channel Service"

# 启动服务
net start TLCPChan
```

### 防火墙配置

```powershell
# 允许 API 端口
netsh advfirewall firewall add rule name="TLCPChan API" dir=in action=allow protocol=tcp localport=30080

# 允许代理端口
netsh advfirewall firewall add rule name="TLCPChan Proxy" dir=in action=allow protocol=tcp localport=443
```

## macOS 安装

### 方式一：PKG 安装包

```bash
# 下载 PKG 包
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-1.0.0-darwin-arm64.pkg

# 安装
sudo installer -pkg tlcpchan-1.0.0-darwin-arm64.pkg -target /

# 安装后需要授权
# 系统偏好设置 -> 安全性与隐私 -> 允许
```

### 方式二：TAR.GZ 压缩包

```bash
# 下载（Apple Silicon）
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-darwin-arm64.tar.gz

# 下载（Intel）
wget https://github.com/Trisia/tlcpchan/releases/download/v1.0.0/tlcpchan-darwin-amd64.tar.gz

# 解压并安装
tar -xzf tlcpchan-darwin-arm64.tar.gz
sudo mv tlcpchan /usr/local/bin/

# 创建工作目录
mkdir -p ~/tlcpchan && cd ~/tlcpchan
```

### 方式三：Homebrew（社区维护）

```bash
# 添加 tap
brew tap trisia/tlcpchan

# 安装
brew install tlcpchan

# 启动服务
brew services start tlcpchan
```

### launchd 服务配置

创建 `~/Library/LaunchAgents/com.trisia.tlcpchan.plist`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.trisia.tlcpchan</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/tlcpchan</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/opt/tlcpchan</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/tlcpchan.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/tlcpchan.log</string>
</dict>
</plist>
```

加载服务：

```bash
launchctl load ~/Library/LaunchAgents/com.trisia.tlcpchan.plist
```

## Docker 安装

### 方式一：Docker Hub

```bash
# 拉取镜像
docker pull trisia/tlcpchan:latest

# 运行容器
docker run -d \
  --name tlcpchan \
  -p 30080:30080 \
  -p 443:443 \
  -v /opt/tlcpchan/config:/etc/tlcpchan \
  -v /opt/tlcpchan/certs:/var/lib/tlcpchan/certs \
  -v /opt/tlcpchan/logs:/var/log/tlcpchan \
  trisia/tlcpchan:latest
```

### 方式二：Docker Compose

创建 `docker-compose.yaml`：

```yaml
version: '3.8'

services:
  tlcpchan:
    image: trisia/tlcpchan:latest
    container_name: tlcpchan
    restart: unless-stopped
    ports:
      - "30080:30080"    # API
      - "443:443"      # TLCP/TLS 代理
      - "8443:8443"    # HTTPS 代理（可选）
    volumes:
      - ./config:/etc/tlcpchan
      - ./certs:/var/lib/tlcpchan/certs
      - ./logs:/var/log/tlcpchan
    environment:
      - TLCPCHAN_LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:30080/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

启动：

```bash
docker-compose up -d
```

### 方式三：自行构建镜像

```bash
# 克隆仓库
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan

# 构建镜像
docker build -t tlcpchan:local .

# 运行
docker run -d -p 30080:30080 -p 443:443 tlcpchan:local
```

## 源码编译

### 环境要求

- Go 1.21 或更高版本
- Git
- Make（可选）

### 编译步骤

```bash
# 克隆仓库
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan

# 下载依赖
go mod download

# 编译
go build -o bin/tlcpchan ./tlcpchan

# 或使用 Make
make build

# 运行测试
go test ./...

# 安装到 GOPATH
go install ./tlcpchan
```

### 交叉编译

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o bin/tlcpchan-linux-amd64 ./tlcpchan

# Linux arm64
GOOS=linux GOARCH=arm64 go build -o bin/tlcpchan-linux-arm64 ./tlcpchan

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o bin/tlcpchan-windows-amd64.exe ./tlcpchan

# macOS arm64
GOOS=darwin GOARCH=arm64 go build -o bin/tlcpchan-darwin-arm64 ./tlcpchan
```

### 编译前端

```bash
cd tlcpchan-ui

# 安装依赖
npm install

# 构建
npm run build

# 输出目录：./dist
```

## 验证安装

### 检查版本

```bash
tlcpchan -version
```

### 检查健康状态

```bash
curl http://localhost:30080/api/v1/system/health
```

预期响应：

```json
{
  "status": "healthy",
  "instances": {
    "total": 0,
    "running": 0,
    "stopped": 0
  },
  "certificates": {
    "total": 0,
    "expired": 0,
    "expiring_soon": 0
  }
}
```

### 访问 Web UI

打开浏览器访问 `http://localhost:30000`

## 卸载

### Linux（DEB）

```bash
sudo systemctl stop tlcpchan
sudo apt-get remove tlcpchan
```

### Linux（RPM）

```bash
sudo systemctl stop tlcpchan
sudo rpm -e tlcpchan
```

### Windows（MSI）

通过「控制面板 → 程序和功能」卸载，或：

```powershell
msiexec /x {产品GUID}
```

### Docker

```bash
docker stop tlcpchan
docker rm tlcpchan
docker rmi trisia/tlcpchan
```

## 升级

### 二进制升级

```bash
# 停止服务
sudo systemctl stop tlcpchan

# 备份配置
cp -r /etc/tlcpchan /etc/tlcpchan.bak

# 下载新版本
wget https://github.com/Trisia/tlcpchan/releases/download/v1.1.0/tlcpchan-linux-amd64.tar.gz
tar -xzf tlcpchan-linux-amd64.tar.gz

# 替换二进制文件
sudo mv tlcpchan /usr/bin/tlcpchan

# 启动服务
sudo systemctl start tlcpchan
```

### Docker 升级

```bash
# 拉取新镜像
docker pull trisia/tlcpchan:latest

# 停止并删除旧容器
docker stop tlcpchan
docker rm tlcpchan

# 启动新容器
docker run -d --name tlcpchan ... trisia/tlcpchan:latest
```

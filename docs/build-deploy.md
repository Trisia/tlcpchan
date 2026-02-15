# 编译部署运行指南

本文档详细介绍 TLCP Channel 的编译、部署和运行方法。

## 目录

- [环境要求](#环境要求)
- [从源码编译](#从源码编译)
- [Docker 部署](#docker-部署)
- [二进制部署](#二进制部署)
- [配置说明](#配置说明)
- [运行管理](#运行管理)

## 环境要求

### 编译环境

- Go 1.21 或更高版本
- Node.js 18 或更高版本（前端）
- npm 或 yarn（前端）

### 运行环境

- Linux / macOS / Windows
- 至少 512MB 内存
- 至少 100MB 磁盘空间

## 从源码编译

### 1. 克隆代码

```bash
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan
```

### 2. 编译后端

```bash
cd tlcpchan
go mod tidy
go build -o bin/tlcpchan .
```

编译成功后，二进制文件位于 `tlcpchan/bin/tlcpchan`。

### 3. 编译前端

```bash
cd ../tlcpchan-ui/web
npm install
npm run build
```

前端构建产物位于 `tlcpchan-ui/web/dist`。

### 4. 编译 CLI 工具（可选）

```bash
cd ../../tlcpchan-cli
go build -o bin/tlcpchan-cli .
```

### 5. 一键编译所有组件

```bash
# 在项目根目录
cd tlcpchan && go build -o ../bin/tlcpchan . && cd ..
cd tlcpchan-cli && go build -o ../bin/tlcpchan-cli . && cd ..
cd tlcpchan-ui/web && npm install && npm run build && cd ../..
```

## Docker 部署

### 使用 Docker Hub 官方镜像（推荐）

直接从 Docker Hub 拉取官方镜像，无需自行编译：

```bash
# 拉取最新镜像
docker pull trisia/tlcpchan:latest

# 运行容器
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

### 使用 docker-compose

```bash
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan
docker compose up -d
```

### 自行构建 Docker 镜像

如果需要自定义构建，可以按以下步骤操作：

```bash
# 1. 克隆代码
git clone https://github.com/Trisia/tlcpchan.git
cd tlcpchan

# 2. 构建镜像
docker build -t tlcpchan:latest .

# 3. 运行容器
docker run -d \
  -p 30080:30080 \
  -p 30000:30000 \
  -p 30443:30443 \
  -v tlcpchan-config:/etc/tlcpchan/config \
  -v tlcpchan-certs:/etc/tlcpchan/certs \
  -v tlcpchan-logs:/etc/tlcpchan/logs \
  --name tlcpchan \
  tlcpchan:latest
```

### 服务访问

- API 服务: http://localhost:30080
- Web UI: http://localhost:30000

## 二进制部署

### Linux 系统

```bash
# 1. 创建工作目录
sudo mkdir -p /etc/tlcpchan/{certs,trusted,logs,config}
sudo chown -R $USER:$USER /etc/tlcpchan

# 2. 复制二进制文件
sudo cp bin/tlcpchan /usr/local/bin/
sudo chmod +x /usr/local/bin/tlcpchan

# 3. 复制配置文件
cp tlcpchan/config/config.yaml /etc/tlcpchan/config/

# 4. 启动服务
tlcpchan
```

### macOS 系统

```bash
# 1. 创建工作目录
mkdir -p ~/.tlcpchan/{certs,trusted,logs,config}

# 2. 复制二进制文件
cp bin/tlcpchan /usr/local/bin/
chmod +x /usr/local/bin/tlcpchan

# 3. 复制配置文件
cp tlcpchan/config/config.yaml ~/.tlcpchan/config/

# 4. 启动服务
tlcpchan -config ~/.tlcpchan/config/config.yaml
```

### Windows 系统

```powershell
# 1. 创建工作目录
mkdir C:\tlcpchan\certs
mkdir C:\tlcpchan\trusted
mkdir C:\tlcpchan\logs
mkdir C:\tlcpchan\config

# 2. 复制文件
copy bin\tlcpchan.exe C:\tlcpchan\
copy tlcpchan\config\config.yaml C:\tlcpchan\config\

# 3. 启动服务
cd C:\tlcpchan
tlcpchan.exe
```

## 配置说明

### 配置文件位置

- Linux: `/etc/tlcpchan/config/config.yaml`
- macOS: `~/.tlcpchan/config/config.yaml`
- Windows: 程序所在目录 `config\config.yaml`

### 基本配置

```yaml
server:
  api:
    address: ":30080"       # API 服务监听地址
  ui:
    enabled: true             # 是否启用 Web UI
    address: ":30000"         # Web UI 监听地址
    path: "./ui"              # 前端静态文件路径
  log:
    level: "info"             # 日志级别: debug, info, warn, error
    file: "./logs/tlcpchan.log"  # 日志文件路径
    max_size: 100             # 单个日志文件最大大小(MB)
    max_backups: 5            # 保留的旧日志文件数量
    max_age: 30               # 保留旧日志文件的最大天数
    compress: true            # 是否压缩旧日志
    enabled: true             # 是否启用日志文件

instances:
  # 代理实例配置
  - name: "tlcp-server-demo"
    type: "server"
    protocol: "tlcp"
    listen: ":30443"
    target: "127.0.0.1:8080"
    enabled: false
```

### 代理实例配置

#### 服务端代理（TLCP）

```yaml
- name: "tlcp-server"
  type: "server"
  protocol: "tlcp"
  listen: ":443"
  target: "127.0.0.1:8080"
  enabled: true
  certificates:
    tlcp:
      cert: "server-sm2"
      key: "server-sm2"
  tlcp:
    auth: "one-way"
    min_version: "1.1"
    max_version: "1.1"
    cipher_suites:
      - "ECDHE_SM4_GCM_SM3"
      - "ECC_SM4_GCM_SM3"
```

#### 客户端代理

```yaml
- name: "tlcp-client"
  type: "client"
  protocol: "tlcp"
  listen: ":9000"
  target: "tlcp-server.example.com:443"
  enabled: true
  server_ca:
    - "ca-sm2"
  tlcp:
    auth: "one-way"
```

#### 协议自动检测

```yaml
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

## 运行管理

### 启动服务

```bash
# 使用默认配置
tlcpchan

# 指定配置文件
tlcpchan -config /path/to/config.yaml

# 查看版本
tlcpchan -version
```

### 后台运行

#### Linux/macOS

```bash
# 使用 nohup
nohup tlcpchan > tlcpchan.log 2>&1 &

# 使用 systemd（推荐）
sudo tee /etc/systemd/system/tlcpchan.service > /dev/null << 'EOF'
[Unit]
Description=TLCP Channel Proxy Service
After=network.target

[Service]
Type=simple
User=nobody
ExecStart=/usr/local/bin/tlcpchan
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable tlcpchan
sudo systemctl start tlcpchan
```

#### Windows

```powershell
# 使用任务计划程序
# 或使用 NSSM (Non-Sucking Service Manager)
nssm install tlcpchan C:\tlcpchan\tlcpchan.exe
nssm start tlcpchan
```

### 查看状态

```bash
# 查看服务状态（systemd）
sudo systemctl status tlcpchan

# 查看日志
tail -f /etc/tlcpchan/logs/tlcpchan.log

# 查看端口监听
netstat -tlnp | grep tlcpchan
```

### 停止服务

```bash
# systemd
sudo systemctl stop tlcpchan

# 查找并杀死进程
pkill tlcpchan
# 或
kill -9 $(pgrep tlcpchan)
```

## 常见问题

### Q: 首次启动报错找不到证书？

A: 首次启动时会自动生成测试证书。确保程序有写入工作目录的权限。

### Q: 端口已被占用？

A: 修改配置文件中的监听端口，或停止占用端口的进程。

### Q: 如何查看详细日志？

A: 修改配置文件中的 `log.level` 为 `debug`，然后重启服务。

### Q: 如何更新证书而不重启服务？

A: 使用 API 或 CLI 工具的证书重载功能：

```bash
# 通过 API
curl -X POST http://localhost:30080/api/v1/certificates/reload

# 通过 CLI（如果已安装）
tlcpchan-cli cert reload
```

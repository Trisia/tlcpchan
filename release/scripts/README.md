# TLCP Channel 打包脚本使用说明

## 目录结构

```
release/
├── scripts/                # 构建脚本目录
│   ├── build.sh            # 跨平台编译脚本
│   ├── release.sh          # 统一发布入口
│   ├── linux/              # Linux平台打包脚本
│   │   ├── deb/            # Debian包配置
│   │   │   ├── package.sh   # Debian打包脚本
│   │   │   ├── nfpm.yaml.template
│   │   │   ├── postinst.sh
│   │   │   └── prerm.sh
│   │   ├── rpm/            # RPM包配置
│   │   │   ├── package.sh   # RPM打包脚本
│   │   │   ├── nfpm.yaml.template
│   │   │   ├── postinst.sh
│   │   │   └── prerm.sh
│   │   ├── install.sh      # 安装脚本
│   │   └── uninstall.sh    # 卸载脚本
│   ├── macos/              # macOS打包配置
│   │   ├── build.sh        # macOS应用构建
│   │   ├── Info.plist.template
│   │   ├── com.trisia.tlcpchan.plist
│   │   └── tlcpchan-wrapper
│   └── windows/            # Windows打包配置
│       ├── build.bat       # Windows构建脚本
│       ├── package.bat     # Windows打包脚本
│       └── tlcpchan.wxs    # WiX安装程序配置
└── systemd/               # systemd服务文件
    └── tlcpchan.service   # systemd服务配置
```

## 快速开始

### 方式一：统一发布（推荐）

```bash
# 完整发布（编译 + 打包）
./release/scripts/release.sh
```

### 方式二：分步构建和打包

```bash
# 1. 先编译所有平台
./release/scripts/build.sh

# 2. 按需打包
# Debian包 (需要nfpm)
bash release/scripts/linux/deb/package.sh

# RPM包 (需要nfpm)
bash release/scripts/linux/rpm/package.sh

# macOS应用包
bash release/scripts/macos/build.sh

# Windows MSI需要在Windows环境运行 package.bat
```

## 支持的平台和架构

| 操作系统 | 架构 | 打包格式 | 打包脚本位置 |
|---------|------|---------|-------------|
| Linux | amd64 | tar.gz, deb, rpm | linux/deb/package.sh, linux/rpm/package.sh |
| Linux | arm64 | tar.gz, deb, rpm | 同上 |
| Linux | loong64 | tar.gz, deb, rpm | 同上 |
| macOS | amd64 | tar.gz, app | macos/build.sh |
| macOS | arm64 | tar.gz, app | macos/build.sh |
| Windows | amd64 | zip, msi (需WiX) | windows/package.bat |

## 构建流程

1. **版本解析** - 从 `tlcpchan/main.go` 解析版本号
2. **前端构建** - 编译 Vue.js 前端项目（当前代码中注释，可选）
3. **跨平台编译** - 为多个平台架构编译 Go 二进制文件
4. **资源复制** - 复制前端资源、证书、配置文件
5. **压缩打包** - 生成 .tar.gz 或 .zip 压缩包
6. **系统包生成** - 生成 .deb/.rpm/.app 等系统安装包

## 版本管理

版本号定义在 `tlcpchan/main.go` 文件中，由构建脚本自动解析：

```bash
# 查看当前版本（读取 tlcpchan/main.go）
grep -E 'version\s*=' tlcpchan/main.go

# 更新版本：编辑 tlcpchan/main.go 中的 version 变量
# 例如：var version = "1.0.0"
```

## 清理构建产物

```bash
# 清理 build/ 和 dist/ 目录
./release/scripts/release.sh --clean
```

## 输出目录

- `build/` - 临时构建文件（各平台编译产物）
- `dist/` - 最终发布包（压缩包和系统包）

## 依赖说明

### 基础依赖

- **Go 1.21+** - 用于后端编译
- **Node.js 18+** - 用于前端构建（如果需要）
- **tar, gzip, zip** - 基础打包工具

### 打包工具依赖

```bash
# nfpm: 用于生成 .deb/.rpm 包
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.0

# WiX Toolset: 仅在 Windows 上用于生成 MSI 安装包
# 下载地址: https://wixtoolset.org/
```

## 平台特定打包说明

### Linux 包

#### Debian/Ubuntu 包 (deb)

```bash
# 先编译 Linux 平台
./release/scripts/build.sh

# 然后生成 deb 包（需要 nfpm）
bash release/scripts/linux/deb/package.sh

# 安装
sudo dpkg -i dist/tlcpchan_<version>_linux_<arch>.deb

# 卸载
sudo dpkg -r tlcpchan
```

#### RHEL/CentOS 包 (rpm)

```bash
# 先编译 Linux 平台
./release/scripts/build.sh

# 然后生成 rpm 包（需要 nfpm）
bash release/scripts/linux/rpm/package.sh

# 安装
sudo rpm -i dist/tlcpchan_<version>_linux_<arch>.rpm

# 卸载
sudo rpm -e tlcpchan
```

#### systemd 服务管理

```bash
# 启动服务
sudo systemctl start tlcpchan

# 设置开机自启
sudo systemctl enable tlcpchan

# 查看状态
sudo systemctl status tlcpchan

# 停止服务
sudo systemctl stop tlcpchan
```

### macOS 包

```bash
# 先编译 macOS 平台
./release/scripts/build.sh

# 然后生成 .app 包
bash release/scripts/macos/build.sh

# 输出位置
dist/tlcpchan_<version>_darwin_<arch>.tar.gz
dist/tlcpchan-<version>-macOS.zip
```

### Windows 包

```batch
REM 需要在 Windows 环境下执行
cd release\scripts\windows

REM 构建
build.bat

REM 打包（需要 WiX Toolset）
package.bat
```

## 发布包内容

每个完整包包含：

- `tlcpchan` - 核心服务（包含API服务和Web UI）
- `tlcpchan-cli` - 命令行工具（含 `tlcpc` 别名）
- `ui/` - 前端静态资源
- `rootcerts/` - 预置信任证书
- systemd 服务文件（Linux）
- 安装/卸载脚本（Linux）

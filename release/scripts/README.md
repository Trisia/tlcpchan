# TLCP Channel 打包脚本使用说明

## 目录结构

```
release/
├── VERSION                 # 版本文件
├── .goreleaser.yaml       # GoReleaser 配置
├── scripts/                # 构建脚本目录
│   ├── build.sh            # 跨平台编译脚本
│   ├── package-deb.sh      # Debian 打包
│   ├── package-rpm.sh      # RPM 打包
│   ├── package-macos.sh    # macOS 打包
│   ├── package-windows.sh  # Windows 打包
│   ├── release.sh          # 统一发布入口
│   └── README.md           # 本文档
└── wix/                    # Windows MSI 配置
```

## 快速开始

### 方式一：统一发布（推荐）

```bash
# 使用纯脚本方式（无需额外依赖）
./release/scripts/release.sh

# 或使用 --script 参数
./release/scripts/release.sh --script
```

### 方式二：使用 GoReleaser

```bash
# 安装 GoReleaser（如果未安装）
go install github.com/goreleaser/goreleaser/v2@latest

# 使用 GoReleaser 发布
./release/scripts/release.sh --goreleaser
```

### 方式三：单独编译和打包

```bash
# 1. 先编译所有平台
./release/scripts/build.sh

# 2. 按需打包
./release/scripts/package-deb.sh      # 仅打包 deb
./release/scripts/package-rpm.sh      # 仅打包 rpm
./release/scripts/package-macos.sh    # 仅打包 macOS
./release/scripts/package-windows.sh  # 仅打包 Windows
```

## 支持的平台和架构

| 操作系统 | 架构 | 打包格式 |
|---------|------|---------|
| Linux | amd64 | tar.gz, deb, rpm |
| Linux | arm64 | tar.gz, deb, rpm |
| Linux | loong64 | tar.gz, deb, rpm |
| macOS | amd64 | tar.gz, zip |
| macOS | arm64 | tar.gz, zip |
| Windows | amd64 | zip, msi (需 WiX) |

## 版本管理

版本号通过 release/ 目录下的 `VERSION` 文件管理：

```bash
# 查看当前版本
cat release/VERSION

# 更新版本
echo "0.2.0" > release/VERSION
```

## 清理构建产物

```bash
# 清理 build/ 和 dist/ 目录
./release/scripts/release.sh --clean
```

## 输出目录

- `build/` - 临时构建文件
- `dist/` - 最终发布包

## 依赖说明

### 纯脚本方式

- Go 1.21+
- Node.js 20+（用于前端构建）
- tar, gzip, zip（可选）

### GoReleaser 方式

- 所有纯脚本方式依赖
- GoReleaser
- nfpm（GoReleaser 内置）

### Windows MSI 打包

需要在 Windows 上安装 [WiX Toolset](https://wixtoolset.org/)

## 发布包内容

每个完整包包含：

- `tlcpchan` - 核心服务
- `tlcpchan-cli` - 命令行工具（含 `tlcpc` 别名）
- `tlcpchan-ui` - UI 服务
- `ui/` - 前端静态资源
- `rootcerts/` - 预置信任证书
- `config.yaml.example` - 配置文件模板
- systemd 服务文件（Linux）

## Linux 包安装/卸载

### DEB 包

```bash
# 安装
sudo dpkg -i tlcpchan_<version>_linux_<arch>.deb

# 卸载
sudo dpkg -r tlcpchan
```

### RPM 包

```bash
# 安装
sudo rpm -i tlcpchan_<version>_linux_<arch>.rpm

# 卸载
sudo rpm -e tlcpchan
```

### systemd 服务管理

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

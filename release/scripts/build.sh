#!/bin/bash

# TLCP Channel 跨平台编译脚本
# 支持多平台、多架构编译

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/version/version.go 中解析版本号
VERSION=$(grep -E 'Version\s*=' "$PROJECT_ROOT/tlcpchan/version/version.go" | head -1 | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# 定义目标平台和架构
PLATFORMS=(
    "linux:amd64"
    "linux:arm64"
    "linux:loong64"
    "darwin:amd64"
    "darwin:arm64"
    "windows:amd64"
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理旧的构建文件
cleanup() {
    log_info "清理旧的构建文件..."
    rm -rf "$BUILD_DIR"
    rm -rf "$DIST_DIR"
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
}

# 构建前端资源
build_frontend() {
    log_info "构建前端资源..."
    if [ -d "$PROJECT_ROOT/tlcpchan-ui" ]; then
        cd "$PROJECT_ROOT/tlcpchan-ui"
        if [ ! -d "node_modules" ]; then
            log_info "安装前端依赖..."
            npm ci
        fi
        npm run build
        cd "$PROJECT_ROOT"
    else
        log_warn "前端目录不存在，跳过前端构建"
    fi
}

# 编译单个平台
build_platform() {
    local os=$1
    local arch=$2
    local output_dir="$BUILD_DIR/$os-$arch"
    
    log_info "编译 $os/$arch..."
    
    mkdir -p "$output_dir"
    
    local ext=""
    if [ "$os" = "windows" ]; then
        ext=".exe"
    fi
    
    # 编译 tlcpchan
    cd "$PROJECT_ROOT/tlcpchan"
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "$output_dir/tlcpchan$ext" .
    
    # 编译 tlcpchan-cli
    cd "$PROJECT_ROOT/tlcpchan-cli"
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "$output_dir/tlcpchan-cli$ext" .
    
    cd "$PROJECT_ROOT"
    
    # 复制前端资源
    if [ -d "$PROJECT_ROOT/tlcpchan-ui/ui" ]; then
        cp -r "$PROJECT_ROOT/tlcpchan-ui/ui" "$output_dir/"
    fi
    
    # 复制信任证书
    if [ -d "$PROJECT_ROOT/trustedcerts" ]; then
        cp -r "$PROJECT_ROOT/trustedcerts" "$output_dir/rootcerts"
    fi
    

    
    # 对于 Linux 平台，添加 systemd 服务文件和安装/卸载脚本
    if [ "$os" = "linux" ]; then
        if [ -f "$RELEASE_DIR/systemd/tlcpchan.service" ]; then
            cp "$RELEASE_DIR/systemd/tlcpchan.service" "$output_dir/"
        fi
        
        # 复制安装脚本
        if [ -f "$SCRIPT_DIR/linux/install.sh" ]; then
            cp "$SCRIPT_DIR/linux/install.sh" "$output_dir/install.sh"
            chmod +x "$output_dir/install.sh"
        fi
        
        # 复制卸载脚本
        if [ -f "$SCRIPT_DIR/linux/uninstall.sh" ]; then
            cp "$SCRIPT_DIR/linux/uninstall.sh" "$output_dir/uninstall.sh"
            chmod +x "$output_dir/uninstall.sh"
        fi
    fi
    
    log_info "完成 $os/$arch 编译"
}

# 创建压缩包
create_archive() {
    local os=$1
    local arch=$2
    local source_dir="$BUILD_DIR/$os-$arch"
    local archive_name="tlcpchan_${VERSION}_${os}_${arch}"
    
    log_info "创建 $archive_name 压缩包..."
    
    cd "$BUILD_DIR"
    
    if [ "$os" = "windows" ]; then
        # Windows 使用 zip
        if command -v zip &> /dev/null; then
            zip -r "$DIST_DIR/$archive_name.zip" "$os-$arch"
        else
            log_warn "zip 命令不可用，跳过 Windows zip 打包"
        fi
    else
        # Unix 使用 tar.gz
        tar -czf "$DIST_DIR/$archive_name.tar.gz" -C "$BUILD_DIR" "$os-$arch"
    fi
    
    cd "$PROJECT_ROOT"
}

# 主函数
main() {
    log_info "========================================"
    log_info "  TLCP Channel 编译脚本"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    cleanup
    # 暂时跳过前端构建以快速测试
    # build_frontend
    
    for platform in "${PLATFORMS[@]}"; do
        IFS=":" read -r os arch <<< "$platform"
        build_platform "$os" "$arch"
        create_archive "$os" "$arch"
    done
    
    log_info "========================================"
    log_info "  编译完成！"
    log_info "  输出目录: $DIST_DIR"
    log_info "========================================"
}

main "$@"

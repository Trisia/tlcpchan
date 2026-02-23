#!/bin/bash

# TLCP Channel Debian 打包脚本
# 使用 nfpm 创建 deb 包

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LINUX_DIR="$(dirname "$SCRIPT_DIR")"
SCRIPTS_DIR="$(dirname "$LINUX_DIR")"
RELEASE_DIR="$(dirname "$SCRIPTS_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/version/version.go 中解析版本号
VERSION=$(grep -E 'Version\s*=' "$PROJECT_ROOT/tlcpchan/version/version.go" | head -1 | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

ARCHES=("amd64" "arm64" "loong64")

# GPG 配置（从环境变量读取）
GPG_PRIVATE_KEY="${GPG_PRIVATE_KEY:-}"
GPG_PASSPHRASE="${GPG_PASSPHRASE:-}"

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

# 检查 nfpm 是否可用
check_nfpm() {
    if ! command -v nfpm &> /dev/null; then
        log_info "nfpm 未安装，正在安装..."
        go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.0
    fi
}

# 创建 deb 包配置
create_nfpm_config() {
    local arch=$1
    local package_arch=$arch
    if [ "$arch" = "loong64" ]; then
        package_arch="loongarch64"
    fi
    
    # 创建 GPG 密钥文件（如果提供了 GPG_PRIVATE_KEY）
    local gpg_key_file=""
    local gpg_passphrase=""
    if [ -n "$GPG_PRIVATE_KEY" ]; then
        gpg_key_file="$BUILD_DIR/gpg-key-${arch}.asc"
        echo "$GPG_PRIVATE_KEY" | base64 -d > "$gpg_key_file"
        gpg_passphrase="$GPG_PASSPHRASE"
    fi
    
    # 先替换 linux-{{ARCH}} 为 linux-$arch，再替换其他变量
    sed -e "s|linux-{{ARCH}}|linux-$arch|g" \
        -e "s|{{ARCH}}|$package_arch|g" \
        -e "s|{{VERSION}}|$VERSION|g" \
        -e "s|{{BUILD_DIR}}|$BUILD_DIR|g" \
        -e "s|{{POSTINST_PATH}}|$BUILD_DIR/postinst.sh|g" \
        -e "s|{{PRERM_PATH}}|$BUILD_DIR/prerm.sh|g" \
        -e "s|{{GPG_KEY_FILE}}|$gpg_key_file|g" \
        -e "s|{{GPG_PASSPHRASE}}|$gpg_passphrase|g" \
        "$SCRIPT_DIR/nfpm.yaml.template" > "$BUILD_DIR/nfpm-$arch.yaml"
}

# 创建 postinst 脚本
create_postinst() {
    cp "$SCRIPT_DIR/postinst.sh" "$BUILD_DIR/postinst.sh"
    chmod +x "$BUILD_DIR/postinst.sh"
}

# 创建 prerm 脚本
create_prerm() {
    cp "$SCRIPT_DIR/prerm.sh" "$BUILD_DIR/prerm.sh"
    chmod +x "$BUILD_DIR/prerm.sh"
}

# 构建 deb 包
build_deb() {
    local arch=$1
    log_info "构建 $arch deb 包..."
    
    create_nfpm_config "$arch"
    nfpm package -f "$BUILD_DIR/nfpm-$arch.yaml" -p deb -t "$DIST_DIR"
    
    local package_arch=$arch
    if [ "$arch" = "loong64" ]; then
        package_arch="loongarch64"
    fi
    mv "$DIST_DIR/tlcpchan_${VERSION}_${package_arch}.deb" "$DIST_DIR/tlcpchan_${VERSION}_linux_${arch}.deb"
}

main() {
    log_info "========================================"
    log_info "  TLCP Channel Debian 打包"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    check_nfpm
    create_postinst
    create_prerm
    
    for arch in "${ARCHES[@]}"; do
        if [ -d "$BUILD_DIR/linux-$arch" ]; then
            build_deb "$arch"
        else
            log_info "跳过 $arch，未找到编译产物"
        fi
    done
    
    log_info "========================================"
    log_info "  deb 打包完成！"
    log_info "  输出目录: $DIST_DIR"
    log_info "========================================"
}

main "$@"

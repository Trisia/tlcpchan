#!/bin/bash

# TLCP Channel RPM 打包脚本
# 使用 nfpm 创建 rpm 包

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

# GPG 签名配置
# 通过环境变量控制签名：
# - GPG_PRIVATE_KEY: base64 编码的 GPG 私钥内容
# - GPG_PASSPHRASE: GPG 密钥密码（会被自动映射到 NFPM_PASSPHRASE）
# 如果不设置 GPG_PRIVATE_KEY，则不会对包进行签名
GPG_PRIVATE_KEY="${GPG_PRIVATE_KEY:-}"

# 将 GPG_PASSPHRASE 映射到 NFPM_PASSPHRASE，兼容 GitHub Actions 的配置
if [ -n "$GPG_PASSPHRASE" ]; then
    export NFPM_PASSPHRASE="$GPG_PASSPHRASE"
fi

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

# 创建 rpm 包配置
create_nfpm_config() {
    local arch=$1
    local package_arch=$arch
    if [ "$arch" = "amd64" ]; then
        package_arch="x86_64"
    elif [ "$arch" = "arm64" ]; then
        package_arch="aarch64"
    elif [ "$arch" = "loong64" ]; then
        package_arch="loongarch64"
    fi
    
    # 创建 GPG 密钥文件（如果提供了 GPG_PRIVATE_KEY）
    # 并设置环境变量让 nfpm 使用
    local export_sign_key=""
    if [ -n "$GPG_PRIVATE_KEY" ]; then
        local gpg_key_file="$BUILD_DIR/gpg-key-${arch}.asc"
        echo "$GPG_PRIVATE_KEY" | base64 -d > "$gpg_key_file"
        export_sign_key="export NFPM_SIGN_KEY_FILE=\"$gpg_key_file\""
    fi
    
    # 保存签名环境变量到临时文件
    echo "$export_sign_key" > "$BUILD_DIR/nfpm-env-rpm-$arch.sh"
    
    # 替换模板变量（不替换签名相关变量，由环境变量控制）
    sed -e "s|linux-{{ARCH}}|linux-$arch|g" \
        -e "s|{{PACKAGE_ARCH}}|$package_arch|g" \
        -e "s|{{VERSION}}|$VERSION|g" \
        -e "s|{{BUILD_DIR}}|$BUILD_DIR|g" \
        -e "s|{{POSTINST_RPM_PATH}}|$BUILD_DIR/postinst-rpm.sh|g" \
        -e "s|{{PRERM_RPM_PATH}}|$BUILD_DIR/prerm-rpm.sh|g" \
        "$SCRIPT_DIR/nfpm.yaml.template" > "$BUILD_DIR/nfpm-rpm-$arch.yaml"
}

# 创建 postinst 脚本
create_postinst() {
    cp "$SCRIPT_DIR/postinst.sh" "$BUILD_DIR/postinst-rpm.sh"
    chmod +x "$BUILD_DIR/postinst-rpm.sh"
}

# 创建 prerm 脚本
create_prerm() {
    cp "$SCRIPT_DIR/prerm.sh" "$BUILD_DIR/prerm-rpm.sh"
    chmod +x "$BUILD_DIR/prerm-rpm.sh"
}

# 构建 rpm 包
build_rpm() {
    local arch=$1
    log_info "构建 $arch rpm 包..."
    
    create_nfpm_config "$arch"
    
    # 加载签名环境变量（如果存在）
    if [ -f "$BUILD_DIR/nfpm-env-rpm-$arch.sh" ]; then
        source "$BUILD_DIR/nfpm-env-rpm-$arch.sh"
        
        # 检查是否启用了签名
        if [ -n "$NFPM_SIGN_KEY_FILE" ]; then
            log_info "  - 启用 GPG 签名（密钥文件: $NFPM_SIGN_KEY_FILE）"
        else
            log_info "  - 不进行 GPG 签名"
        fi
    fi
    
    nfpm package -f "$BUILD_DIR/nfpm-rpm-$arch.yaml" -p rpm -t "$DIST_DIR"
    
    local package_arch=$arch
    if [ "$arch" = "amd64" ]; then
        package_arch="x86_64"
    elif [ "$arch" = "arm64" ]; then
        package_arch="aarch64"
    elif [ "$arch" = "loong64" ]; then
        package_arch="loongarch64"
    fi
    mv "$DIST_DIR/tlcpchan-${VERSION}-1.${package_arch}.rpm" "$DIST_DIR/tlcpchan_${VERSION}_linux_${arch}.rpm"
}

main() {
    log_info "========================================"
    log_info "  TLCP Channel RPM 打包"
    log_info "  版本: $VERSION"
    
    # 检查是否启用 GPG 签名
    if [ -n "$GPG_PRIVATE_KEY" ]; then
        log_info "  GPG 签名: 已启用"
    else
        log_info "  GPG 签名: 未启用"
    fi
    log_info "========================================"
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    check_nfpm
    create_postinst
    create_prerm
    
    for arch in "${ARCHES[@]}"; do
        if [ -d "$BUILD_DIR/linux-$arch" ]; then
            build_rpm "$arch"
        else
            log_info "跳过 $arch，未找到编译产物"
        fi
    done
    
    log_info "========================================"
    log_info "  rpm 打包完成！"
    log_info "  输出目录: $DIST_DIR"
    log_info "========================================"
}

main "$@"

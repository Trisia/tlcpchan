#!/bin/bash

# TLCP Channel RPM 打包脚本
# 使用 nfpm 创建 rpm 包

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/main.go 中解析版本号
VERSION=$(grep -E '^var\s+version\s*=' "$PROJECT_ROOT/tlcpchan/main.go" | head -1 | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/')
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

ARCHES=("amd64" "arm64" "loong64")

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

# 检查 nfpm 是否可用
check_nfpm() {
    if ! command -v nfpm &> /dev/null; then
        log_info "nfpm 未安装，正在安装..."
        go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
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
    
    cat > "$BUILD_DIR/nfpm-rpm-$arch.yaml" << EOF
name: "tlcpchan"
arch: "$package_arch"
platform: "linux"
version: "$VERSION"
release: "1"
section: "default"
priority: "optional"
maintainer: "TLCP Channel Team"
description: |
  TLCP/TLS 协议代理工具，支持双协议并行工作
vendor: "TLCP Channel"
homepage: "https://github.com/Trisia/tlcpchan"
license: "Apache-2.0"
contents:
  - src: "$BUILD_DIR/linux-$arch/tlcpchan"
    dst: "/etc/tlcpchan/tlcpchan"
    file_info:
      mode: 0755
  - src: "$BUILD_DIR/linux-$arch/tlcpchan-cli"
    dst: "/etc/tlcpchan/tlcpchan-cli"
    file_info:
      mode: 0755
  - src: "$BUILD_DIR/linux-$arch/tlcpchan-ui"
    dst: "/etc/tlcpchan/tlcpchan-ui"
    file_info:
      mode: 0755
  - src: "$BUILD_DIR/linux-$arch/ui"
    dst: "/etc/tlcpchan/ui"
    type: tree
  - src: "$BUILD_DIR/linux-$arch/rootcerts"
    dst: "/etc/tlcpchan/rootcerts"
    type: tree
  - src: "$BUILD_DIR/linux-$arch/config.yaml.example"
    dst: "/etc/tlcpchan/config.yaml.example"
    type: config
  - src: "$BUILD_DIR/linux-$arch/tlcpchan.service"
    dst: "/usr/lib/systemd/system/tlcpchan.service"
    file_info:
      mode: 0644
  - dst: "/etc/tlcpchan/keystores"
    type: dir
    file_info:
      mode: 0755
  - dst: "/etc/tlcpchan/logs"
    type: dir
    file_info:
      mode: 0755
scripts:
  postinstall: "$BUILD_DIR/postinst-rpm.sh"
  preremove: "$BUILD_DIR/prerm-rpm.sh"
EOF
}

# 创建 postinst 脚本
create_postinst() {
    cat > "$BUILD_DIR/postinst-rpm.sh" << 'EOF'
#!/bin/bash
set -e

# 创建 tlcpchan 用户
if ! getent passwd tlcpchan > /dev/null; then
    useradd -r -s /bin/false -d /etc/tlcpchan tlcpchan
fi

# 设置权限
chown -R tlcpchan:tlcpchan /etc/tlcpchan/keystores 2>/dev/null || true
chown -R tlcpchan:tlcpchan /etc/tlcpchan/logs 2>/dev/null || true

# 创建软链接到 /usr/bin
ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
ln -sf /etc/tlcpchan/tlcpchan-ui /usr/bin/tlcpchan-ui
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 重新加载 systemd
systemctl daemon-reload 2>/dev/null || true

echo "TLCP Channel 安装成功！"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
EOF
    chmod +x "$BUILD_DIR/postinst-rpm.sh"
}

# 创建 prerm 脚本
create_prerm() {
    cat > "$BUILD_DIR/prerm-rpm.sh" << 'EOF'
#!/bin/bash
set -e

# 停止服务
if systemctl is-active --quiet tlcpchan 2>/dev/null; then
    systemctl stop tlcpchan
fi

# 禁用服务
if systemctl is-enabled --quiet tlcpchan 2>/dev/null; then
    systemctl disable tlcpchan
fi

# 重新加载 systemd
systemctl daemon-reload 2>/dev/null || true

# 删除软链接
rm -f /usr/bin/tlcpchan
rm -f /usr/bin/tlcpchan-cli
rm -f /usr/bin/tlcpchan-ui
rm -f /usr/bin/tlcpc
EOF
    chmod +x "$BUILD_DIR/prerm-rpm.sh"
}

# 构建 rpm 包
build_rpm() {
    local arch=$1
    log_info "构建 $arch rpm 包..."
    
    create_nfpm_config "$arch"
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

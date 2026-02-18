#!/bin/bash

# TLCP Channel 统一发布脚本
# 支持使用 GoReleaser 或纯脚本方式发布

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/main.go 中解析版本号
VERSION=$(grep -E '^var\s+version\s*=' "$PROJECT_ROOT/tlcpchan/main.go" | head -1 | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/')

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

log_warn() {
    echo -e "\033[1;33m[WARN]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

usage() {
    cat << EOF
TLCP Channel 发布脚本 v${VERSION}

用法:
  $0 [选项]

选项:
  --goreleaser    使用 GoReleaser 发布（推荐）
  --script        使用纯脚本方式发布
  --clean         清理构建产物
  -h, --help      显示帮助信息

示例:
  $0 --goreleaser    # 使用 GoReleaser 完整发布
  $0 --script         # 使用纯脚本方式发布
  $0 --clean          # 清理构建文件
EOF
}

check_goreleaser() {
    if command -v goreleaser &> /dev/null; then
        return 0
    else
        return 1
    fi
}

install_goreleaser() {
    log_info "正在安装 GoReleaser..."
    go install github.com/goreleaser/goreleaser/v2@latest
}

release_goreleaser() {
    log_info "========================================"
    log_info "  使用 GoReleaser 发布"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    if ! check_goreleaser; then
        install_goreleaser
    fi
    
    # 创建 GoReleaser 需要的脚本
    create_postinstall_for_goreleaser
    create_preremove_for_goreleaser
    
    # 运行 GoReleaser
    cd "$PROJECT_ROOT"
    goreleaser release --snapshot --clean -f "$RELEASE_DIR/.goreleaser.yaml"
    
    log_info "========================================"
    log_info "  GoReleaser 发布完成！"
    log_info "========================================"
}

release_script() {
    log_info "========================================"
    log_info "  使用纯脚本方式发布"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    # 编译所有平台
    "$SCRIPT_DIR/build.sh"
    
    # 打包各格式
    if [ -d "$PROJECT_ROOT/build/linux-amd64" ] || [ -d "$PROJECT_ROOT/build/linux-arm64" ] || [ -d "$PROJECT_ROOT/build/linux-loong64" ]; then
        "$SCRIPT_DIR/package-deb.sh"
        "$SCRIPT_DIR/package-rpm.sh"
    fi
    
    if [ -d "$PROJECT_ROOT/build/darwin-amd64" ] || [ -d "$PROJECT_ROOT/build/darwin-arm64" ]; then
        "$SCRIPT_DIR/package-macos.sh"
    fi
    
    if [ -d "$PROJECT_ROOT/build/windows-amd64" ]; then
        "$SCRIPT_DIR/package-windows.sh"
    fi
    
    log_info "========================================"
    log_info "  纯脚本发布完成！"
    log_info "  输出目录: $PROJECT_ROOT/dist"
    log_info "========================================"
}

cleanup() {
    log_info "清理构建产物..."
    rm -rf "$PROJECT_ROOT/build"
    rm -rf "$PROJECT_ROOT/dist"
    log_info "清理完成"
}

create_postinstall_for_goreleaser() {
    cat > "$RELEASE_DIR/scripts/postinstall.sh" << 'EOF'
#!/bin/bash
set -e

if ! getent passwd tlcpchan > /dev/null; then
    useradd -r -s /bin/false -d /etc/tlcpchan tlcpchan
fi

chown -R tlcpchan:tlcpchan /etc/tlcpchan/keystores 2>/dev/null || true
chown -R tlcpchan:tlcpchan /etc/tlcpchan/logs 2>/dev/null || true

# 创建软链接到 /usr/bin
ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
ln -sf /etc/tlcpchan/tlcpchan-ui /usr/bin/tlcpchan-ui
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

systemctl daemon-reload 2>/dev/null || true

echo "TLCP Channel 安装成功！"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
EOF
    chmod +x "$RELEASE_DIR/scripts/postinstall.sh"
}

create_preremove_for_goreleaser() {
    cat > "$RELEASE_DIR/scripts/preremove.sh" << 'EOF'
#!/bin/bash
set -e

if systemctl is-active --quiet tlcpchan 2>/dev/null; then
    systemctl stop tlcpchan
fi

if systemctl is-enabled --quiet tlcpchan 2>/dev/null; then
    systemctl disable tlcpchan
fi

systemctl daemon-reload 2>/dev/null || true

# 删除软链接
rm -f /usr/bin/tlcpchan
rm -f /usr/bin/tlcpchan-cli
rm -f /usr/bin/tlcpchan-ui
rm -f /usr/bin/tlcpc
EOF
    chmod +x "$RELEASE_DIR/scripts/preremove.sh"
}

# 主函数
main() {
    case "$1" in
        --goreleaser)
            release_goreleaser
            ;;
        --script)
            release_script
            ;;
        --clean)
            cleanup
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_info "未指定发布方式，默认使用纯脚本方式"
            release_script
            ;;
    esac
}

main "$@"

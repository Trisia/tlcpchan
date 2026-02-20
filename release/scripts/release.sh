#!/bin/bash

# TLCP Channel 统一发布脚本
# 使用纯脚本方式发布

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/main.go 中解析版本号
VERSION=$(grep -E 'version\s*=' "$PROJECT_ROOT/tlcpchan/main.go" | head -1 | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/')

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
  --clean         清理构建产物
  -h, --help      显示帮助信息

示例:
  $0                # 完整发布
  $0 --clean        # 清理构建文件
EOF
}

cleanup() {
    log_info "清理构建产物..."
    rm -rf "$PROJECT_ROOT/build"
    rm -rf "$PROJECT_ROOT/dist"
    log_info "清理完成"
}

release() {
    log_info "========================================"
    log_info "  TLCP Channel 发布"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    # 编译所有平台
    "$SCRIPT_DIR/build.sh"
    
    # 打包各格式
    if [ -d "$PROJECT_ROOT/build/linux-amd64" ] || [ -d "$PROJECT_ROOT/build/linux-arm64" ] || [ -d "$PROJECT_ROOT/build/linux-loong64" ]; then
        if [ -f "$SCRIPT_DIR/linux/deb/package.sh" ]; then
            "$SCRIPT_DIR/linux/deb/package.sh"
        fi
        if [ -f "$SCRIPT_DIR/linux/rpm/package.sh" ]; then
            "$SCRIPT_DIR/linux/rpm/package.sh"
        fi
    fi
    
    if [ -d "$PROJECT_ROOT/build/darwin-amd64" ] || [ -d "$PROJECT_ROOT/build/darwin-arm64" ]; then
        if [ -f "$SCRIPT_DIR/macos/build.sh" ]; then
            "$SCRIPT_DIR/macos/build.sh"
        fi
    fi
    
    if [ -d "$PROJECT_ROOT/build/windows-amd64" ]; then
        log_warn "Windows MSI 包需要在 Windows 环境下使用 package.bat 生成"
    fi
    
    log_info "========================================"
    log_info "  发布完成！"
    log_info "  输出目录: $PROJECT_ROOT/dist"
    log_info "========================================"
}

# 主函数
main() {
    case "$1" in
        --clean)
            cleanup
            ;;
        -h|--help)
            usage
            ;;
        *)
            release
            ;;
    esac
}

main "$@"

#!/usr/bin/env bash
# ============================================
# TLCP Channel 一键安装脚本 (Linux/macOS)
# GitHub: https://github.com/Trisia/tlcpchan
# ============================================

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目配置
REPO_OWNER="Trisia"
REPO_NAME="tlcpchan"
GITHUB_API="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}"
INSTALL_DIR="/etc/tlcpchan"

# 全局变量
USE_BINARY=false

# 打印错误信息
print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# 打印成功信息
print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

# 打印普通信息
print_info() {
    echo -e "  $1"
}

# 打印警告信息
print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

# 清理临时文件
cleanup() {
    if [ -n "${tmp_dir:-}" ] && [ -d "$tmp_dir" ]; then
        rm -rf "$tmp_dir"
    fi
}

# 设置退出时清理
trap cleanup EXIT

# 检测操作系统
detect_os() {
    local os
    os=$(uname -s)
    case "$os" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        *)
            print_error "不支持的操作系统: $os"
            exit 1
            ;;
    esac
}

# 检测 CPU 架构
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        loongarch64)
            echo "loong64"
            ;;
        *)
            print_error "不支持的架构: $arch"
            exit 1
            ;;
    esac
}

# 检查必要命令
check_commands() {
    local commands=("tar" "curl" "chmod" "mkdir" "grep" "sed")
    for cmd in "${commands[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            print_error "缺少必要命令: $cmd"
            exit 1
        fi
    done
}

# 检查是否已安装
check_installed() {
    if [ -d "$INSTALL_DIR" ]; then
        print_error "TLCP Channel 已经安装在 $INSTALL_DIR"
        echo ""
        print_info "如需重新安装，请先卸载旧版本"
        echo ""
        print_info "卸载方式:"
        print_info "  - 通过包管理器: sudo apt remove tlcpchan 或 sudo yum remove tlcpchan"
        print_info "  - 手动删除: sudo rm -rf $INSTALL_DIR"
        exit 1
    fi
}

# 检查权限
check_permissions() {
    if [ "$EUID" -ne 0 ]; then
        print_error "需要 root 权限或使用 sudo 执行此脚本"
        print_info "请使用: sudo bash $0"
        exit 1
    fi
}

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -s "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    if [ -z "$version" ]; then
        print_error "无法获取最新版本信息"
        exit 1
    fi
    echo "$version"
}

# 检测包管理器
detect_package_manager() {
    # 检测顺序：apt -> dnf -> yum -> zypper
    if command -v apt >/dev/null 2>&1; then
        echo "apt"
    elif command -v dnf >/dev/null 2>&1; then
        echo "dnf"
    elif command -v yum >/dev/null 2>&1; then
        echo "yum"
    elif command -v zypper >/dev/null 2>&1; then
        echo "zypper"
    else
        echo "none"
    fi
}

# 构建包下载 URL (deb 或 rpm)
build_package_url() {
    local os=$1
    local arch=$2
    local version=$3

    local package_format url

    # 根据包管理器确定包格式
    case "$(detect_package_manager)" in
        apt)
            package_format="deb"
            ;;
        dnf|yum|zypper)
            package_format="rpm"
            ;;
        *)
            return 1
            ;;
    esac

    local filename="tlcpchan_${version}_${os}_${arch}.${package_format}"
    url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${filename}"
    echo "$url"
    return 0
}

# 构建二进制包下载 URL (tar.gz)
build_binary_url() {
    local os=$1
    local arch=$2
    local version=$3

    local filename="tlcpchan_${version}_${os}_${arch}.tar.gz"
    local url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${filename}"
    echo "$url"
}

# 下载安装包
download_package() {
    local url=$1
    local output_dir=$2

    local filename=$(basename "$url")
    local output_path="${output_dir}/${filename}"

    echo -e "📥 正在下载安装包..."
    if ! curl -# -L -o "$output_path" "$url"; then
        print_error "下载安装包失败"
        return 1
    fi

    # 检查文件大小
    if [ ! -s "$output_path" ]; then
        print_error "下载的文件为空"
        rm -f "$output_path"
        return 1
    fi

    echo ""
    return 0
}

# 使用包管理器安装
install_via_package_manager() {
    local version=$1
    local arch=$2
    local pkg_manager=$3

    local package_url filename output_path

    # 构建包 URL
    if ! package_url=$(build_package_url "linux" "$arch" "$version"); then
        return 1
    fi

    filename=$(basename "$package_url")
    output_path="${tmp_dir}/${filename}"

    # 下载包
    if ! download_package "$package_url" "$tmp_dir"; then
        return 1
    fi

    # 直接使用 dpkg 或 rpm 安装（无依赖）
    echo -e "📦 正在安装..."
    echo ""

    case "$pkg_manager" in
        apt)
            if ! dpkg -i "$output_path"; then
                print_error "dpkg 安装失败"
                return 1
            fi
            ;;
        dnf|yum|zypper)
            if ! rpm -i "$output_path"; then
                print_error "rpm 安装失败"
                return 1
            fi
            ;;
        *)
            return 1
            ;;
    esac

    echo ""
    return 0
}

# 使用二进制方式安装
install_via_binary() {
    local version=$1
    local arch=$2
    local os=$3
    local install_dir=$4

    local binary_url filename output_path

    # 构建二进制包 URL
    binary_url=$(build_binary_url "$os" "$arch" "$version")
    filename=$(basename "$binary_url")
    output_path="${tmp_dir}/${filename}"

    # 下载二进制包
    if ! download_package "$binary_url" "$tmp_dir"; then
        return 1
    fi

    # 解压和安装
    echo -e "📦 正在安装到 $install_dir..."
    echo ""

    # 创建安装目录
    mkdir -p "$install_dir"
    mkdir -p "${install_dir}/keystores"
    mkdir -p "${install_dir}/logs"

    # 解压到临时目录
    local tmp_extract_dir="${install_dir}_extract"
    rm -rf "$tmp_extract_dir"
    mkdir -p "$tmp_extract_dir"

    if ! tar -xzf "$output_path" -C "$tmp_extract_dir"; then
        print_error "解压安装包失败"
        rm -rf "$tmp_extract_dir"
        return 1
    fi

    # 移动文件到安装目录
    if [ -d "${tmp_extract_dir}/tlcpchan" ]; then
        mv "${tmp_extract_dir}/tlcpchan"/* "$install_dir/"
    else
        mv "$tmp_extract_dir"/* "$install_dir/" 2>/dev/null || true
    fi

    # 清理临时解压目录
    rm -rf "$tmp_extract_dir"

    # 设置可执行权限
    chmod +x "${install_dir}/tlcpchan"
    chmod +x "${install_dir}/tlcpchan-cli"

    echo ""
    return 0
}

# 警告：回退到二进制安装
warn_binary_fallback() {
    echo ""
    print_warning "未检测到包管理器（apt/yum/dnf/zypper）"
    echo ""
    print_info "建议:"
    print_info "  - 使用系统包管理器可以更好地管理依赖和更新"
    print_info "  - 将使用二进制安装包（tar.gz）进行安装"
    echo ""
    print_info "如需使用包管理器安装，请先安装对应的包管理工具"
    echo ""
}

# 询问用户：是否使用二进制安装
ask_binary_fallback() {
    echo ""
    print_warning "包管理器安装失败"
    echo ""
    print_info "是否使用二进制安装包（tar.gz）继续安装？"
    echo ""
    echo -n "  输入 y/N: "

    local answer
    read -r answer

    case "$answer" in
        [yY]|[yY][eE][sS])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# 验证安装
verify_install() {
    local install_dir=$1

    echo -e "🔍 正在验证安装..."

    local version_output
    if ! version_output=$("${install_dir}/tlcpchan" -version 2>&1); then
        print_error "安装验证失败"
        return 1
    fi

    local installed_version
    installed_version=$(echo "$version_output" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)

    if [ -z "$installed_version" ]; then
        print_error "无法获取安装的版本信息"
        return 1
    fi

    echo ""
    echo "✓ 版本: $installed_version"
    echo ""
    return 0
}

# 显示安装成功信息
show_success_info() {
    local version=$1
    local install_method=$2

    echo -e "${GREEN}✅ TLCP Channel 安装成功！${NC}"
    echo ""
    echo -e "📦 版本信息"
    print_info "版本: $version"
    print_info "安装方式: $install_method"
    echo ""
    echo -e "📂 安装位置"
    print_info "安装目录: $INSTALL_DIR"
    print_info "可执行文件: $INSTALL_DIR/tlcpchan"
    print_info "CLI 工具: $INSTALL_DIR/tlcpchan-cli"
    echo ""
    echo -e "🚀 启动服务"
    print_info "启动: sudo systemctl start tlcpchan"
    print_info "查看状态: sudo systemctl status tlcpchan"
    print_info "停止: sudo systemctl stop tlcpchan"
    print_info "开机自启: sudo systemctl enable tlcpchan"
    print_info "查看日志: sudo journalctl -u tlcpchan -f"
    echo ""
    echo -e "🌐 访问 Web 界面"
    print_info "http://localhost:20080"
    echo ""
    echo -e "📖 更多信息"
    print_info "文档: https://github.com/Trisia/tlcpchan"
}

# 主函数
main() {
    # 检查必要命令
    check_commands

    # 检测系统环境
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)

    echo -e "🔍 正在检测系统环境..."
    echo ""
    print_info "操作系统: $os"
    print_info "架构: $arch"
    echo ""

    # 检查是否已安装
    check_installed

    # 检查权限
    check_permissions

    # 获取最新版本
    local version
    echo -e "🔍 正在获取最新版本信息..."
    version=$(get_latest_version)
    print_info "最新版本: v$version"
    echo ""

    # 创建临时目录
    tmp_dir=$(mktemp -d)

    local install_result=0
    local install_method=""

    if [ "$os" = "linux" ]; then
        # 检测包管理器
        local pkg_manager
        pkg_manager=$(detect_package_manager)

        if [ "$pkg_manager" = "none" ]; then
            # 包管理器不存在 → 警告后使用二进制
            warn_binary_fallback
            install_method="二进制安装 (tar.gz)"
            
            if ! install_via_binary "$version" "$arch" "$os" "$INSTALL_DIR"; then
                install_result=1
            fi
        else
            # 尝试包管理器安装
            echo -e "📦 检测到包管理器: ${CYAN}${pkg_manager}${NC}"
            echo ""
            
            if install_via_package_manager "$version" "$arch" "$pkg_manager"; then
                install_result=0
                install_method="包管理器安装 (${pkg_manager})"
            else
                # 包管理器安装失败 → 询问是否使用二进制
                if ask_binary_fallback; then
                    install_method="二进制安装 (tar.gz)"
                    
                    if ! install_via_binary "$version" "$arch" "$os" "$INSTALL_DIR"; then
                        install_result=1
                    fi
                else
                    echo ""
                    print_info "安装已取消"
                    exit 1
                fi
            fi
        fi
    else
        # macOS → 使用二进制安装
        install_method="二进制安装 (tar.gz)"
        
        if ! install_via_binary "$version" "$arch" "$os" "$INSTALL_DIR"; then
            install_result=1
        fi
    fi

    # 检查安装结果
    if [ $install_result -ne 0 ]; then
        print_error "安装失败，请检查错误信息"
        exit 1
    fi

    # 验证安装
    if ! verify_install "$INSTALL_DIR"; then
        print_error "安装验证失败"
        exit 1
    fi

    # 显示成功信息
    show_success_info "$version" "$install_method"
}

# 执行主函数
main

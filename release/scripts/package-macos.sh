#!/bin/bash

# TLCP Channel macOS 打包脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/main.go 中解析版本号
VERSION=$(grep -E '^var\s+version\s*=' "$PROJECT_ROOT/tlcpchan/main.go" | head -1 | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/')
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

ARCHES=("amd64" "arm64")

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

# 创建 macOS 应用包结构
create_app_bundle() {
    local arch=$1
    local app_name="TLCP Channel"
    local app_dir="$BUILD_DIR/macos-$arch/$app_name.app"
    
    log_info "创建 $arch macOS 应用包..."
    
    mkdir -p "$app_dir/Contents/MacOS"
    mkdir -p "$app_dir/Contents/Resources"
    
    # 创建 Info.plist
    cat > "$app_dir/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>tlcpchan-wrapper</string>
    <key>CFBundleIdentifier</key>
    <string>com.trisia.tlcpchan</string>
    <key>CFBundleName</key>
    <string>TLCP Channel</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
</dict>
</plist>
EOF
    
    # 创建启动脚本
    cat > "$app_dir/Contents/MacOS/tlcpchan-wrapper" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
open -a Terminal.app ./tlcpchan-ui
EOF
    chmod +x "$app_dir/Contents/MacOS/tlcpchan-wrapper"
    
    # 复制可执行文件
    cp "$BUILD_DIR/darwin-$arch/tlcpchan" "$app_dir/Contents/MacOS/"
    cp "$BUILD_DIR/darwin-$arch/tlcpchan-cli" "$app_dir/Contents/MacOS/"
    cp "$BUILD_DIR/darwin-$arch/tlcpchan-ui" "$app_dir/Contents/MacOS/"
    
    # 复制资源文件
    cp -r "$BUILD_DIR/darwin-$arch/ui" "$app_dir/Contents/Resources/"
    cp -r "$BUILD_DIR/darwin-$arch/rootcerts" "$app_dir/Contents/Resources/"
    cp "$BUILD_DIR/darwin-$arch/config.yaml.example" "$app_dir/Contents/Resources/"
    
    # 创建 .tar.gz 和 .zip
    cd "$BUILD_DIR/macos-$arch"
    tar -czf "$DIST_DIR/tlcpchan_${VERSION}_darwin_${arch}.tar.gz" "$app_name.app"
    if command -v zip &> /dev/null; then
        zip -r "$DIST_DIR/tlcpchan_${VERSION}_darwin_${arch}.zip" "$app_name.app"
    fi
    cd "$PROJECT_ROOT"
}

# 创建 launchd plist
create_launchd_plist() {
    local arch=$1
    cat > "$BUILD_DIR/darwin-$arch/com.trisia.tlcpchan.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.trisia.tlcpchan</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/tlcpchan</string>
        <string>-ui</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/tlcpchan.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/tlcpchan.err.log</string>
</dict>
</plist>
EOF
}

main() {
    log_info "========================================"
    log_info "  TLCP Channel macOS 打包"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    for arch in "${ARCHES[@]}"; do
        if [ -d "$BUILD_DIR/darwin-$arch" ]; then
            create_launchd_plist "$arch"
            create_app_bundle "$arch"
        else
            log_info "跳过 $arch，未找到编译产物"
        fi
    done
    
    log_info "========================================"
    log_info "  macOS 打包完成！"
    log_info "  输出目录: $DIST_DIR"
    log_info "========================================"
}

main "$@"

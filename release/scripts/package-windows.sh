#!/bin/bash

# TLCP Channel Windows 打包脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$RELEASE_DIR")"

# 从 tlcpchan/main.go 中解析版本号
VERSION=$(grep -E '^var\s+version\s*=' "$PROJECT_ROOT/tlcpchan/main.go" | head -1 | sed -E 's/.*version\s*=\s*"([^"]+)".*/\1/')
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"
WIX_DIR="$RELEASE_DIR/wix"

ARCHES=("amd64")

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

# 创建 WiX 配置
create_wix_config() {
    log_info "创建 WiX 配置..."
    
    mkdir -p "$WIX_DIR"
    
    cat > "$WIX_DIR/tlcpchan.wxs" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" Name="TLCP Channel" Language="1033" Version="$(var.Version)" Manufacturer="TLCP Channel Team" UpgradeCode="12345678-1234-1234-1234-123456789012">
    <Package InstallerVersion="200" Compressed="yes" InstallScope="perMachine" />
    
    <MajorUpgrade DowngradeErrorMessage="A newer version of [ProductName] is already installed." />
    <MediaTemplate />
    
    <Feature Id="ProductFeature" Title="TLCP Channel" Level="1">
      <ComponentGroupRef Id="ProductComponents" />
    </Feature>
    
    <Property Id="WIXUI_INSTALLDIR" Value="INSTALLFOLDER" />
    <UIRef Id="WixUI_Minimal" />
  </Product>
  
  <Fragment>
    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="TLCP Channel" />
      </Directory>
    </Directory>
  </Fragment>
  
  <Fragment>
    <ComponentGroup Id="ProductComponents" Directory="INSTALLFOLDER">
      <Component Id="tlcpchan_exe" Guid="*">
        <File Id="tlcpchan_exe" Name="tlcpchan.exe" Source="$(var.SourceDir)/tlcpchan.exe" KeyPath="yes" />
      </Component>
      <Component Id="tlcpchan_cli_exe" Guid="*">
        <File Id="tlcpchan_cli_exe" Name="tlcpchan-cli.exe" Source="$(var.SourceDir)/tlcpchan-cli.exe" />
      </Component>
      <Component Id="tlcpchan_ui_exe" Guid="*">
        <File Id="tlcpchan_ui_exe" Name="tlcpchan-ui.exe" Source="$(var.SourceDir)/tlcpchan-ui.exe" />
      </Component>
    </ComponentGroup>
  </Fragment>
</Wix>
EOF
}

# 创建 Windows 安装脚本
create_install_script() {
    log_info "创建 Windows 安装脚本..."
    
    if [ ! -d "$BUILD_DIR/windows-amd64" ]; then
        log_info "未找到 Windows 编译产物，跳过"
        return
    fi
    
    cat > "$BUILD_DIR/windows-amd64/install.bat" << 'EOF'
@echo off
SETLOCAL

set "INSTALL_DIR=%ProgramFiles%\TLCP Channel"
set "CONFIG_DIR=%ALLUSERSPROFILE%\TLCP Channel"

echo Installing TLCP Channel...

REM 创建目录
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"
if not exist "%CONFIG_DIR%\keystores" mkdir "%CONFIG_DIR%\keystores"
if not exist "%CONFIG_DIR%\logs" mkdir "%CONFIG_DIR%\logs"
if not exist "%CONFIG_DIR%\rootcerts" mkdir "%CONFIG_DIR%\rootcerts"

REM 复制文件
xcopy /E /I /Y "%~dp0tlcpchan.exe" "%INSTALL_DIR%\"
xcopy /E /I /Y "%~dp0tlcpchan-cli.exe" "%INSTALL_DIR%\"
xcopy /E /I /Y "%~dp0tlcpchan-ui.exe" "%INSTALL_DIR%\"
xcopy /E /I /Y "%~dp0ui" "%CONFIG_DIR%\ui\"
xcopy /E /I /Y "%~dp0rootcerts" "%CONFIG_DIR%\rootcerts\"
xcopy /E /I /Y "%~dp0config.yaml.example" "%CONFIG_DIR%\"

REM 添加到 PATH
setx PATH "%PATH%;%INSTALL_DIR%" /M

echo Installation complete!
echo Please restart your terminal to update PATH.
echo To start TLCP Channel, run: tlcpchan -ui
pause
EOF
    
    cat > "$BUILD_DIR/windows-amd64/uninstall.bat" << 'EOF'
@echo off
SETLOCAL

set "INSTALL_DIR=%ProgramFiles%\TLCP Channel"
set "CONFIG_DIR=%ALLUSERSPROFILE%\TLCP Channel"

echo Uninstalling TLCP Channel...

REM 删除程序目录
if exist "%INSTALL_DIR%" rmdir /S /Q "%INSTALL_DIR%"

echo Uninstallation complete!
pause
EOF
}

# 创建 zip 包
create_zip_package() {
    log_info "创建 Windows zip 包..."
    
    if [ ! -d "$BUILD_DIR/windows-amd64" ]; then
        log_info "未找到 Windows 编译产物，跳过"
        return
    fi
    
    if command -v zip &> /dev/null; then
        cd "$BUILD_DIR"
        zip -r "$DIST_DIR/tlcpchan_${VERSION}_windows_amd64.zip" "windows-amd64"
        cd "$PROJECT_ROOT"
    else
        log_info "zip 命令不可用，跳过 zip 打包"
    fi
}

main() {
    log_info "========================================"
    log_info "  TLCP Channel Windows 打包"
    log_info "  版本: $VERSION"
    log_info "========================================"
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    create_wix_config
    create_install_script
    create_zip_package
    
    log_info "========================================"
    log_info "  Windows 打包完成！"
    log_info "  输出目录: $DIST_DIR"
    log_info "  注意：MSI 安装包需要在 Windows 上使用 WiX Toolset 构建"
    log_info "========================================"
}

main "$@"

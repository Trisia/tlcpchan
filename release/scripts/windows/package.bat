@echo off
setlocal enabledelayedexpansion

REM TLCP Channel Windows MSI 打包脚本
REM 使用 WiX Toolset 生成 Windows 安装程序

echo ========================================
echo   TLCP Channel Windows MSI 打包脚本
echo ========================================

REM 设置目录
set "SCRIPT_DIR=%~dp0"
set "RELEASE_DIR=%SCRIPT_DIR%.."
for %%i in ("%RELEASE_DIR%") do set "RELEASE_DIR=%%~fi"
for %%i in ("%RELEASE_DIR%\..") do set "PROJECT_ROOT=%%~fi"

REM 从 tlcpchan/version/version.go 中解析版本号
set "VERSION_FILE=%PROJECT_ROOT%\tlcpchan\version\version.go"
if exist "%VERSION_FILE%" (
    for /f "tokens=2 delims==" %%a in ('findstr /r "Version.*=" "%VERSION_FILE%"') do (
        set "VERSION=%%a"
    )
    set "VERSION=%VERSION:"=%"
    set "VERSION=%VERSION: =%"
) else (
    echo [ERROR] version.go not found!
    exit /b 1
)

if "%VERSION%"=="" (
    echo [ERROR] Failed to parse version from version.go!
    exit /b 1
)

set "BUILD_DIR=%PROJECT_ROOT%\build"
set "DIST_DIR=%PROJECT_ROOT%\dist"
set "SOURCE_DIR=%BUILD_DIR%\windows-amd64"
set "WIX_DIR=%SCRIPT_DIR%wix-toolset"
set "WXS_FILE=%SCRIPT_DIR%tlcpchan.wxs"

echo [INFO] 版本: %VERSION%
echo [INFO] 项目根目录: %PROJECT_ROOT%
echo [INFO] 源文件目录: %SOURCE_DIR%
echo [INFO] WiX 目录: %WIX_DIR%

REM 检查源文件是否存在，不存在则运行 build.bat
if not exist "%SOURCE_DIR%\tlcpchan.exe" (
    echo [INFO] 源文件不存在，先运行 build.bat 进行构建...
    call "%SCRIPT_DIR%build.bat"
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] build.bat 执行失败！
        exit /b 1
    )
)

REM 检查 WiX Toolset 是否存在
set "WIX_FOUND=0"
where candle >nul 2>nul
if %ERRORLEVEL% equ 0 (
    set "WIX_FOUND=1"
    set "CANDLE=candle"
    set "LIGHT=light"
) else if exist "%WIX_DIR%\candle.exe" (
    set "WIX_FOUND=1"
    set "CANDLE=%WIX_DIR%\candle.exe"
    set "LIGHT=%WIX_DIR%\light.exe"
)

if %WIX_FOUND% equ 0 (
    echo [INFO] WiX Toolset 未找到，开始下载...
    
    REM 创建 WiX 目录
    if not exist "%WIX_DIR%" mkdir "%WIX_DIR%"
    
    REM 下载 WiX Toolset v3.11.2 (最后一个稳定的 .NET Framework 版本)
    set "WIX_URL=https://github.com/wixtoolset/wix3/releases/download/wix3112rtm/wix311-binaries.zip"
    set "WIX_ZIP=%WIX_DIR%\wix311-binaries.zip"
    
    echo [INFO] 正在从 %WIX_URL% 下载...
    
    REM 使用 PowerShell 下载
    powershell -Command "& {[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%WIX_URL%' -OutFile '%WIX_ZIP%'}"
    
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] WiX Toolset 下载失败！
        exit /b 1
    )
    
    echo [INFO] 解压 WiX Toolset...
    powershell -Command "& {Expand-Archive -Path '%WIX_ZIP%' -DestinationPath '%WIX_DIR%' -Force}"
    
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] WiX Toolset 解压失败！
        exit /b 1
    )
    
    REM 清理 zip 文件
    del "%WIX_ZIP%"
    
    set "CANDLE=%WIX_DIR%\candle.exe"
    set "LIGHT=%WIX_DIR%\light.exe"
    
    echo [INFO] WiX Toolset 安装完成！
)

REM 创建输出目录
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

REM 编译 WiX 源文件
echo [INFO] 编译 WiX 源文件...
"%CANDLE%" -nologo -dVersion=%VERSION% -dSourceDir=%SOURCE_DIR% -out "%BUILD_DIR%\\" "%WXS_FILE%"

if %ERRORLEVEL% neq 0 (
    echo [ERROR] candle.exe 编译失败！
    exit /b 1
)

REM 链接生成 MSI
echo [INFO] 生成 MSI 安装包...
"%LIGHT%" -nologo -out "%DIST_DIR%\tlcpchan_%VERSION%_windows_amd64.msi" "%BUILD_DIR%\tlcpchan.wixobj" -ext WixUIExtension

if %ERRORLEVEL% neq 0 (
    echo [ERROR] light.exe 链接失败！
    exit /b 1
)

REM 清理临时文件
del "%BUILD_DIR%\tlcpchan.wixobj" 2>nul

echo ========================================
echo   MSI 打包完成！
echo   输出文件: %DIST_DIR%\tlcpchan_%VERSION%_windows_amd64.msi
echo ========================================

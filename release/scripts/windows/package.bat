@echo off
setlocal enabledelayedexpansion

REM TLCP Channel Windows MSI 打包脚本
REM 使用 WiX Toolset 生成 Windows 安装程序

echo ========================================
echo   TLCP Channel Windows MSI 打包脚本
echo ========================================

REM 设置目录
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%..\..\.."
for %%i in ("%PROJECT_ROOT%") do set "PROJECT_ROOT=%%~fi"

REM 从 tlcpchan/version/version.go 中解析版本号
set "VERSION_FILE=%PROJECT_ROOT%\tlcpchan\version\version.go"
if exist "%VERSION_FILE%" (
    for /f "usebackq tokens=2 delims==" %%a in (`findstr /r "Version" "%VERSION_FILE%"`) do (
        set "VERSION_LINE=%%a"
    )
    REM 清理版本号：去除引号、空格、分号
    set "VERSION=%VERSION_LINE:"=%"
    set "VERSION=%VERSION: =%"
    set "VERSION=%VERSION:;=%"
) else (
    echo [ERROR] version.go not found at %VERSION_FILE%!
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
    set "HEAT=heat"
) else if exist "%WIX_DIR%\candle.exe" (
    set "WIX_FOUND=1"
    set "CANDLE=%WIX_DIR%\candle.exe"
    set "LIGHT=%WIX_DIR%\light.exe"
    set "HEAT=%WIX_DIR%\heat.exe"
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

REM 检查 ui 和 rootcerts 目录是否存在
set "HAS_UI=0"
set "HAS_ROOTCERTS=0"

if exist "%SOURCE_DIR%\ui" (
    set "HAS_UI=1"
) else (
    echo [WARN] ui 目录不存在：%SOURCE_DIR%\ui，跳过 UI 文件打包
)

if exist "%SOURCE_DIR%\rootcerts" (
    set "HAS_ROOTCERTS=1"
) else (
    echo [WARN] rootcerts 目录不存在：%SOURCE_DIR%\rootcerts，跳过信任证书打包
)

REM 使用 heat 生成 ui 目录结构的 XML 片段
set "UI_WXS_ARG="
if %HAS_UI% equ 1 (
    echo [INFO] 生成 UI 目录的 WiX 片段...
    "%HEAT%" dir "%SOURCE_DIR%\ui" -gg -scom -sreg -sfrag -sw5150 -dr INSTALLFOLDER -cg UiComponents -out "%BUILD_DIR%\ui.wxs"
    
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] heat.exe 生成 ui.wxs 失败！
        exit /b 1
    )
    set "UI_WXS_ARG=%BUILD_DIR%\ui.wxs %BUILD_DIR%\ui.wixobj"
)

REM 使用 heat 生成 rootcerts 目录结构的 XML 片段
set "ROOTCERTS_WXS_ARG="
if %HAS_ROOTCERTS% equ 1 (
    echo [INFO] 生成信任证书目录的 WiX 片段...
    "%HEAT%" dir "%SOURCE_DIR%\rootcerts" -gg -scom -sreg -sfrag -sw5150 -dr INSTALLFOLDER -cg RootCertComponents -out "%BUILD_DIR%\rootcerts.wxs"
    
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] heat.exe 生成 rootcerts.wxs 失败！
        exit /b 1
    )
    set "ROOTCERTS_WXS_ARG=%BUILD_DIR%\rootcerts.wxs %BUILD_DIR%\rootcerts.wixobj"
)

REM 编译 WiX 源文件
echo [INFO] 编译 WiX 源文件...
set "CANDLE_CMD=%CANDLE% -nologo -dVersion=%VERSION% -dSourceDir=%SOURCE_DIR% -out "%BUILD_DIR%\\" "%WXS_FILE%""

if %HAS_UI% equ 1 (
    set "CANDLE_CMD=%CANDLE_CMD% "%BUILD_DIR%\ui.wxs""
)
if %HAS_ROOTCERTS% equ 1 (
    set "CANDLE_CMD=%CANDLE_CMD% "%BUILD_DIR%\rootcerts.wxs""
)

%CANDLE_CMD%

if %ERRORLEVEL% neq 0 (
    echo [ERROR] candle.exe 编译失败！
    exit /b 1
)

REM 链接生成 MSI
echo [INFO] 生成 MSI 安装包...
set "LIGHT_CMD=%LIGHT% -sw1076 -nologo -out "%DIST_DIR%\tlcpchan_%VERSION%_windows_amd64.msi" "%BUILD_DIR%\tlcpchan.wixobj""

if %HAS_UI% equ 1 (
    set "LIGHT_CMD=%LIGHT_CMD% "%BUILD_DIR%\ui.wixobj""
)
if %HAS_ROOTCERTS% equ 1 (
    set "LIGHT_CMD=%LIGHT_CMD% "%BUILD_DIR%\rootcerts.wixobj""
)

set "LIGHT_CMD=%LIGHT_CMD% -ext WixUIExtension"
%LIGHT_CMD%

if %ERRORLEVEL% neq 0 (
    echo [ERROR] light.exe 链接失败！
    exit /b 1
)

REM 清理临时文件
del "%BUILD_DIR%\tlcpchan.wixobj" 2>nul
if %HAS_UI% equ 1 (
    del "%BUILD_DIR%\ui.wxs" 2>nul
    del "%BUILD_DIR%\ui.wixobj" 2>nul
)
if %HAS_ROOTCERTS% equ 1 (
    del "%BUILD_DIR%\rootcerts.wxs" 2>nul
    del "%BUILD_DIR%\rootcerts.wixobj" 2>nul
)

echo ========================================
echo   MSI 打包完成！
echo   输出文件: %DIST_DIR%\tlcpchan_%VERSION%_windows_amd64.msi
echo ========================================

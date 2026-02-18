@echo off
setlocal enabledelayedexpansion

REM TLCP Channel Windows 构建脚本
REM 支持在 Windows 环境下构建 TLCP Channel

echo ========================================
echo   TLCP Channel Windows 构建脚本
echo ========================================

REM 设置目录
set "SCRIPT_DIR=%~dp0"
set "RELEASE_DIR=%SCRIPT_DIR%.."
for %%i in ("%RELEASE_DIR%") do set "RELEASE_DIR=%%~fi"
for %%i in ("%RELEASE_DIR%\..") do set "PROJECT_ROOT=%%~fi"

REM 从 tlcpchan/main.go 中解析版本号
set "MAIN_GO=%PROJECT_ROOT%\tlcpchan\main.go"
if exist "%MAIN_GO%" (
    for /f "tokens=2 delims==" %%a in ('findstr /r "var.*version.*=" "%MAIN_GO%"') do (
        set "VERSION=%%a"
    )
    set "VERSION=%VERSION:"=%"
    set "VERSION=%VERSION: =%"
) else (
    echo [ERROR] main.go not found!
    exit /b 1
)

if "%VERSION%"=="" (
    echo [ERROR] Failed to parse version from main.go!
    exit /b 1
)

set "BUILD_DIR=%PROJECT_ROOT%\build"
set "DIST_DIR=%PROJECT_ROOT%\dist"
set "OUTPUT_DIR=%BUILD_DIR%\windows-amd64"

echo [INFO] 版本: %VERSION%
echo [INFO] 项目根目录: %PROJECT_ROOT%
echo [INFO] 输出目录: %OUTPUT_DIR%

REM 检查 Go 是否安装
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go 未安装，请先安装 Go 1.21+
    exit /b 1
)

REM 检查 Node.js 是否安装（用于前端）
where node >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [WARN] Node.js 未安装，将跳过前端构建
    set "SKIP_FRONTEND=1"
)

REM 创建目录
if exist "%OUTPUT_DIR%" (
    echo [INFO] 清理旧的构建文件...
    rmdir /s /q "%OUTPUT_DIR%"
)
mkdir "%OUTPUT_DIR%"

REM 构建前端
if not defined SKIP_FRONTEND (
    echo [INFO] 构建前端资源...
    if exist "%PROJECT_ROOT%\tlcpchan-ui\web" (
        cd "%PROJECT_ROOT%\tlcpchan-ui\web"
        if not exist "node_modules" (
            echo [INFO] 安装前端依赖...
            call npm ci
        )
        call npm run build
        cd "%PROJECT_ROOT%"
    ) else (
        echo [WARN] 前端目录不存在，跳过前端构建
    )
)

REM 编译 tlcpchan
echo [INFO] 编译 tlcpchan...
cd "%PROJECT_ROOT%\tlcpchan"
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=0"
go build -ldflags="-s -w -X main.version=%VERSION%" -o "%OUTPUT_DIR%\tlcpchan.exe" .
if %ERRORLEVEL% neq 0 (
    echo [ERROR] tlcpchan 编译失败!
    exit /b 1
)

REM 编译 tlcpchan-cli
echo [INFO] 编译 tlcpchan-cli...
cd "%PROJECT_ROOT%\tlcpchan-cli"
go build -ldflags="-s -w -X main.version=%VERSION%" -o "%OUTPUT_DIR%\tlcpchan-cli.exe" .
if %ERRORLEVEL% neq 0 (
    echo [ERROR] tlcpchan-cli 编译失败!
    exit /b 1
)

REM 编译 tlcpchan-ui
echo [INFO] 编译 tlcpchan-ui...
cd "%PROJECT_ROOT%\tlcpchan-ui"
go build -ldflags="-s -w -X main.version=%VERSION%" -o "%OUTPUT_DIR%\tlcpchan-ui.exe" .
if %ERRORLEVEL% neq 0 (
    echo [ERROR] tlcpchan-ui 编译失败!
    exit /b 1
)

cd "%PROJECT_ROOT%"

REM 复制前端资源
if exist "%PROJECT_ROOT%\tlcpchan-ui\ui" (
    echo [INFO] 复制前端资源...
    xcopy /E /I /Y "%PROJECT_ROOT%\tlcpchan-ui\ui" "%OUTPUT_DIR%\ui\"
)

REM 复制信任证书
if exist "%PROJECT_ROOT%\trustedcerts" (
    echo [INFO] 复制信任证书...
    xcopy /E /I /Y "%PROJECT_ROOT%\trustedcerts" "%OUTPUT_DIR%\rootcerts\"
)

REM 创建配置文件模板
echo [INFO] 创建配置文件模板...
(
echo # TLCP Channel 配置文件
echo api:
echo   address: ":30080"
echo ui:
echo   address: ":30000"
echo log:
echo   level: "info"
echo   path: "%%ALLUSERSPROFILE%%\TLCP Channel\logs"
) > "%OUTPUT_DIR%\config.yaml.example"

REM 创建安装脚本
echo [INFO] 创建安装脚本...
(
echo @echo off
echo SETLOCAL
echo.
echo set "INSTALL_DIR=%%ProgramFiles%%\TLCP Channel"
echo set "CONFIG_DIR=%%ALLUSERSPROFILE%%\TLCP Channel"
echo.
echo echo Installing TLCP Channel...
echo.
echo REM 创建目录
echo if not exist "%%INSTALL_DIR%%" mkdir "%%INSTALL_DIR%%"
echo if not exist "%%CONFIG_DIR%%" mkdir "%%CONFIG_DIR%%"
echo if not exist "%%CONFIG_DIR%%\keystores" mkdir "%%CONFIG_DIR%%\keystores"
echo if not exist "%%CONFIG_DIR%%\logs" mkdir "%%CONFIG_DIR%%\logs"
echo if not exist "%%CONFIG_DIR%%\rootcerts" mkdir "%%CONFIG_DIR%%\rootcerts"
echo.
echo REM 复制文件
echo xcopy /E /I /Y "%%~dp0tlcpchan.exe" "%%INSTALL_DIR%%\"
echo xcopy /E /I /Y "%%~dp0tlcpchan-cli.exe" "%%INSTALL_DIR%%\"
echo xcopy /E /I /Y "%%~dp0tlcpchan-ui.exe" "%%INSTALL_DIR%%\"
echo xcopy /E /I /Y "%%~dp0ui" "%%CONFIG_DIR%%\ui\"
echo xcopy /E /I /Y "%%~dp0rootcerts" "%%CONFIG_DIR%%\rootcerts\"
echo xcopy /E /I /Y "%%~dp0config.yaml.example" "%%CONFIG_DIR%%\"
echo.
echo REM 添加到 PATH
echo setx PATH "%%PATH%%;%%INSTALL_DIR%%" /M
echo.
echo echo Installation complete!
echo echo Please restart your terminal to update PATH.
echo echo To start TLCP Channel, run: tlcpchan -ui
echo pause
) > "%OUTPUT_DIR%\install.bat"

REM 创建卸载脚本
(
echo @echo off
echo SETLOCAL
echo.
echo set "INSTALL_DIR=%%ProgramFiles%%\TLCP Channel"
echo set "CONFIG_DIR=%%ALLUSERSPROFILE%%\TLCP Channel"
echo.
echo echo Uninstalling TLCP Channel...
echo.
echo REM 删除程序目录
echo if exist "%%INSTALL_DIR%%" rmdir /S /Q "%%INSTALL_DIR%%"
echo.
echo echo Uninstallation complete!
echo pause
) > "%OUTPUT_DIR%\uninstall.bat"

REM 创建 zip 包
where zip >nul 2>nul
if %ERRORLEVEL% equ 0 (
    echo [INFO] 创建 zip 包...
    if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"
    cd "%BUILD_DIR%"
    zip -r "%DIST_DIR%\tlcpchan_%VERSION%_windows_amd64.zip" "windows-amd64"
    cd "%PROJECT_ROOT%"
) else (
    echo [WARN] zip 命令不可用，跳过 zip 打包
)

echo ========================================
echo   构建完成！
echo   输出目录: %OUTPUT_DIR%
echo ========================================

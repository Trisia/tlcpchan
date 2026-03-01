# TLCP Channel Windows MSI 打包脚本
# 使用 WiX Toolset 生成 Windows 安装程序

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TLCP Channel Windows MSI 打包脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 设置目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = (Resolve-Path (Join-Path $ScriptDir "..\..\")).Path
$VersionFile = Join-Path $ProjectRoot "tlcpchan\version\version.go"

# 解析版本号
if (-not (Test-Path $VersionFile)) {
    Write-Host "[ERROR] version.go not found at $VersionFile!" -ForegroundColor Red
    exit 1
}

$Match = Select-String -Path $VersionFile -Pattern 'Version\s*=\s*"([^"]+)"' -AllMatches | 
         Select-Object -First 1
$Version = if ($Match) { $Match.Matches[0].Groups[1].Value } else { $null }

if (-not $Version) {
    Write-Host "[ERROR] Failed to parse version from version.go!" -ForegroundColor Red
    exit 1
}

$BuildDir = Join-Path $ProjectRoot "build"
$DistDir = Join-Path $ProjectRoot "dist"
$SourceDir = Join-Path $BuildDir "windows-amd64"
$WixDir = Join-Path $ScriptDir "wix-toolset"
$WxsFile = Join-Path $ScriptDir "tlcpchan.wxs"

Write-Host "[INFO] 版本: $Version" -ForegroundColor Green
Write-Host "[INFO] 项目根目录: $ProjectRoot" -ForegroundColor Green
Write-Host "[INFO] 源文件目录: $SourceDir" -ForegroundColor Green
Write-Host "[INFO] WiX 目录: $WixDir" -ForegroundColor Green

# 检查源文件是否存在，不存在则运行 build.bat
$TlcpchanExe = Join-Path $SourceDir "tlcpchan.exe"
if (-not (Test-Path $TlcpchanExe)) {
    Write-Host "[INFO] 源文件不存在，先运行 build.bat 进行构建..." -ForegroundColor Green
    & (Join-Path $ScriptDir "build.bat")
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] build.bat 执行失败！" -ForegroundColor Red
        exit 1
    }
}

# 检查 WiX Toolset 是否存在
$Candle = Get-Command candle -ErrorAction SilentlyContinue
if ($Candle) {
    $CandleExe = "candle"
    $LightExe = "light"
    $HeatExe = "heat"
} elseif (Test-Path (Join-Path $WixDir "candle.exe")) {
    $CandleExe = Join-Path $WixDir "candle.exe"
    $LightExe = Join-Path $WixDir "light.exe"
    $HeatExe = Join-Path $WixDir "heat.exe"
} else {
    Write-Host "[INFO] WiX Toolset 未找到，开始下载..." -ForegroundColor Green
    
    # 创建 WiX 目录
    New-Item -ItemType Directory -Path $WixDir -Force | Out-Null
    
    # 下载 WiX Toolset v3.11.2
    $WixUrl = "https://github.com/wixtoolset/wix3/releases/download/wix3112rtm/wix311-binaries.zip"
    $WixZip = Join-Path $WixDir "wix311-binaries.zip"
    
    Write-Host "[INFO] 正在从 $WixUrl 下载..." -ForegroundColor Green
    
    # 使用 Invoke-WebRequest 下载
    Invoke-WebRequest -Uri $WixUrl -OutFile $WixZip
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] WiX Toolset 下载失败！" -ForegroundColor Red
        Write-Host "[INFO] 请手动下载 WiX Toolset: $WixUrl" -ForegroundColor Yellow
        Write-Host "[INFO] 解压到: $WixDir" -ForegroundColor Yellow
        exit 1
    }
    
    Write-Host "[INFO] 解压 WiX Toolset..." -ForegroundColor Green
    # 使用 Expand-Archive 解压
    Expand-Archive -Path $WixZip -DestinationPath $WixDir -Force
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] WiX Toolset 解压失败！" -ForegroundColor Red
        exit 1
    }
    
    # 清理 zip 文件
    Remove-Item -Path $WixZip -Force
    
    $CandleExe = Join-Path $WixDir "candle.exe"
    $LightExe = Join-Path $WixDir "light.exe"
    
    Write-Host "[INFO] WiX Toolset 安装完成！" -ForegroundColor Green
}

# 创建输出目录
New-Item -ItemType Directory -Path $DistDir -Force | Out-Null

# 检查 ui 和 rootcerts 目录是否存在
$UiDir = Join-Path $SourceDir "ui"
$HasUi = Test-Path $UiDir
if (-not $HasUi) {
    Write-Host "[WARN] ui 目录不存在：$UiDir，跳过 UI 文件打包" -ForegroundColor Yellow
}

$RootCertsDir = Join-Path $SourceDir "rootcerts"
$HasRootCerts = Test-Path $RootCertsDir
if (-not $HasRootCerts) {
    Write-Host "[WARN] rootcerts 目录不存在：$RootCertsDir，跳过信任证书打包" -ForegroundColor Yellow
}

# 使用 heat 生成 ui 目录结构的 XML 片段
$WxsFiles = @()
if ($HasUi) {
    Write-Host "[INFO] 生成 UI 目录的目录树和 WiX 片段..." -ForegroundColor Green
    Get-ChildItem -Path $UiDir -Recurse | Format-Table FullName
    
    $UiWxsFile = Join-Path $BuildDir "ui.wxs"
    & $HeatExe dir $UiDir -gg -scom -sreg -sfrag -sw5150 -dr INSTALLFOLDER -cg UiComponents -var var.SourceDir -srd -out $UiWxsFile
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] heat.exe 生成 ui.wxs 失败！" -ForegroundColor Red
        exit 1
    }
    $WxsFiles += $UiWxsFile
}

# 使用 heat 生成 rootcerts 目录结构的 XML 片段
if ($HasRootCerts) {
    Write-Host "[INFO] 生成信任证书目录的目录树和 WiX 片段..." -ForegroundColor Green
    Get-ChildItem -Path $RootCertsDir -Recurse | Format-Table FullName
    
    $RootCertsWxsFile = Join-Path $BuildDir "rootcerts.wxs"
    & $HeatExe dir $RootCertsDir -gg -scom -sreg -sfrag -sw5150 -dr INSTALLFOLDER -cg RootCertComponents -var var.SourceDir -srd -out $RootCertsWxsFile
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] heat.exe 生成 rootcerts.wxs 失败！" -ForegroundColor Red
        exit 1
    }
    $WxsFiles += $RootCertsWxsFile
}

# 编译 WiX 源文件
Write-Host "[INFO] 编译 WiX 源文件..." -ForegroundColor Green

$WixObjectFiles = @()
& $CandleExe -nologo -dVersion=$Version -dSourceDir=$SourceDir -out "$BuildDir\" $WxsFile $WxsFiles

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] candle.exe 编译失败！" -ForegroundColor Red
    exit 1
}

$WixObjectFiles += Join-Path $BuildDir "tlcpchan.wixobj"
if ($HasUi) {
    $WixObjectFiles += Join-Path $BuildDir "ui.wixobj"
}
if ($HasRootCerts) {
    $WixObjectFiles += Join-Path $BuildDir "rootcerts.wixobj"
}

# 链接生成 MSI
Write-Host "[INFO] 生成 MSI 安装包..." -ForegroundColor Green

$MsiPath = Join-Path $DistDir "tlcpchan_${Version}_windows_amd64.msi"
& $LightExe -sw1076 -nologo -out $MsiPath $WixObjectFiles -ext WixUIExtension

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] light.exe 链接失败！" -ForegroundColor Red
    exit 1
}

# 清理临时文件
Remove-Item -Path (Join-Path $BuildDir "tlcpchan.wixobj") -ErrorAction SilentlyContinue
if ($HasUi) {
    Remove-Item -Path (Join-Path $BuildDir "ui.wxs") -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $BuildDir "ui.wixobj") -ErrorAction SilentlyContinue
}
if ($HasRootCerts) {
    Remove-Item -Path (Join-Path $BuildDir "rootcerts.wxs") -ErrorAction SilentlyContinue
    Remove-Item -Path (Join-Path $BuildDir "rootcerts.wixobj") -ErrorAction SilentlyContinue
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MSI 打包完成！" -ForegroundColor Green
Write-Host "  输出文件: $MsiPath" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

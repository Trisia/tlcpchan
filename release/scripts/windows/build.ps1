# TLCP Channel Windows 构建脚本
# 支持在 Windows 环境下构建 TLCP Channel

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TLCP Channel Windows 构建脚本" -ForegroundColor Cyan
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
$OutputDir = Join-Path $BuildDir "windows-amd64"

Write-Host "[INFO] 版本: $Version" -ForegroundColor Green
Write-Host "[INFO] 项目根目录: $ProjectRoot" -ForegroundColor Green
Write-Host "[INFO] 输出目录: $OutputDir" -ForegroundColor Green

# 检查 Go 是否安装
$GoCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $GoCmd) {
    Write-Host "[ERROR] Go 未安装，请先安装 Go 1.26+" -ForegroundColor Red
    exit 1
}

# 检查 Node.js 是否安装
$NodeCmd = Get-Command node -ErrorAction SilentlyContinue
$SkipFrontend = -not $NodeCmd
if ($SkipFrontend) {
    Write-Host "[WARN] Node.js 未安装，将跳过前端构建" -ForegroundColor Yellow
}

# 创建目录
if (Test-Path $OutputDir) {
    Write-Host "[INFO] 清理旧的构建文件..." -ForegroundColor Green
    Remove-Item -Path $OutputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

# 构建前端
if (-not $SkipFrontend) {
    $FrontendPackageJson = Join-Path $ProjectRoot "tlcpchan-ui\package.json"
    if (Test-Path $FrontendPackageJson) {
        Write-Host "[INFO] 构建前端资源..." -ForegroundColor Green
        Push-Location (Join-Path $ProjectRoot "tlcpchan-ui")
        
        $NodeModules = Join-Path (Get-Location) "node_modules"
        if (-not (Test-Path $NodeModules)) {
            Write-Host "[INFO] 安装前端依赖..." -ForegroundColor Green
            npm ci
            if ($LASTEXITCODE -ne 0) {
                Pop-Location
                Write-Host "[ERROR] npm ci 失败!" -ForegroundColor Red
                exit 1
            }
        }
        
        npm run build
        if ($LASTEXITCODE -ne 0) {
            Pop-Location
            Write-Host "[ERROR] npm run build 失败!" -ForegroundColor Red
            exit 1
        }
        
        Pop-Location
    } else {
        Write-Host "[WARN] 前端目录不存在，跳过前端构建" -ForegroundColor Yellow
    }
}

# 编译 tlcpchan
Write-Host "[INFO] 编译 tlcpchan..." -ForegroundColor Green
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = "0"
Push-Location (Join-Path $ProjectRoot "tlcpchan")
go build -ldflags="-s -w" -o (Join-Path $OutputDir "tlcpchan.exe") .
if ($LASTEXITCODE -ne 0) {
    Pop-Location
    Write-Host "[ERROR] tlcpchan 编译失败!" -ForegroundColor Red
    exit 1
}
Pop-Location

# 编译 tlcpchan-cli
Write-Host "[INFO] 编译 tlcpchan-cli..." -ForegroundColor Green
Push-Location (Join-Path $ProjectRoot "tlcpchan-cli")
go build -ldflags="-s -w" -o (Join-Path $OutputDir "tlcpchan-cli.exe") .
if ($LASTEXITCODE -ne 0) {
    Pop-Location
    Write-Host "[ERROR] tlcpchan-cli 编译失败!" -ForegroundColor Red
    exit 1
}
Pop-Location

# 复制前端资源
$FrontendUiDir = Join-Path $ProjectRoot "tlcpchan-ui\ui"
if (Test-Path $FrontendUiDir) {
    Write-Host "[INFO] 复制前端资源..." -ForegroundColor Green
    Copy-Item -Path $FrontendUiDir -Destination (Join-Path $OutputDir "ui") -Recurse -Force
}

# 复制信任证书
$TrustedCertsDir = Join-Path $ProjectRoot "trustedcerts"
if (Test-Path $TrustedCertsDir) {
    Write-Host "[INFO] 复制信任证书..." -ForegroundColor Green
    Copy-Item -Path $TrustedCertsDir -Destination (Join-Path $OutputDir "rootcerts") -Recurse -Force
}

# 注意：MSI 安装包将自动处理安装和卸载，无需额外的安装/卸载脚本

# 创建 zip 包
Write-Host "[INFO] 创建 zip 包..." -ForegroundColor Green
New-Item -ItemType Directory -Path $DistDir -Force | Out-Null
$ZipPath = Join-Path $DistDir "tlcpchan_${Version}_windows_amd64.zip"
Compress-Archive -Path (Join-Path $BuildDir "windows-amd64") -DestinationPath $ZipPath -Force
if ($LASTEXITCODE -ne 0) {
    Write-Host "[WARN] zip 包创建失败，跳过" -ForegroundColor Yellow
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  构建完成！" -ForegroundColor Green
Write-Host "  输出目录: $OutputDir" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

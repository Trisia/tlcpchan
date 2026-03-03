# TLCP Channel Windows Build Script
# Build TLCP Channel in Windows environment
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TLCP Channel Windows Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 设置目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = (Resolve-Path (Join-Path $ScriptDir "..\..\..\")).Path
$VersionFile = Join-Path $ProjectRoot "tlcpchan\version\version.go"
$IconFile = Join-Path $ProjectRoot "icon.ico"

Write-Host "[INFO] ScriptDir: $ScriptDir" -ForegroundColor Green
Write-Host "[INFO] Project Root: $ProjectRoot" -ForegroundColor Green
Write-Host "[INFO] Version File: $VersionFile" -ForegroundColor Green
Write-Host "[INFO] Icon File: $IconFile" -ForegroundColor Green


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

Write-Host "[INFO] Version: $Version" -ForegroundColor Green
Write-Host "[INFO] Output Directory: $OutputDir" -ForegroundColor Green

# 检查 Go 是否安装
$GoCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $GoCmd) {
    Write-Host "[ERROR] Go is not installed, please install Go 1.26+" -ForegroundColor Red
    exit 1
}

# 检查 Node.js 是否安装
$NodeCmd = Get-Command node -ErrorAction SilentlyContinue
if (-not $NodeCmd) {
    Write-Host "[WARN] Node.js is not installed, please install it first" -ForegroundColor Yellow
    exit 1
}



# 创建目录
if (Test-Path $OutputDir) {
    Write-Host "[INFO] Cleaning old build files..." -ForegroundColor Green
    Remove-Item -Path $OutputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$RsrcExe = Join-Path $ScriptDir "rsrc.exe"
if (-not (Test-Path $RsrcExe)) {
    Write-Host "[INFO] rsrc not found, downloading..." -ForegroundColor Green
    
    $RsrcUrl = "https://github.com/akavel/rsrc/releases/download/v0.10.2/rsrc_windows_amd64.exe"
    Write-Host "[INFO] Downloading rsrc from $RsrcUrl ..." -ForegroundColor Green
    Invoke-WebRequest -Uri $RsrcUrl -OutFile $RsrcExe
    
    if (-not (Test-Path $RsrcExe)) {
        Write-Host "[ERROR] Failed to download rsrc!" -ForegroundColor Red
        exit 1
    }
    Write-Host "[INFO] rsrc downloaded successfully!" -ForegroundColor Green
}

# 检查图标文件
$IconFile = Join-Path $ProjectRoot "icon.ico"
Write-Host $IconFile

if (-not (Test-Path $IconFile)) {
    Write-Host "[WARN] icon.ico not found at $IconFile, skipping icon embedding" -ForegroundColor Yellow
    $HasIcon = $false
} else {
    Write-Host "[INFO] Icon file found: $IconFile" -ForegroundColor Green
    $HasIcon = $true
}

# 构建前端
$FrontendPackageJson = Join-Path $ProjectRoot "tlcpchan-ui\package.json"
if (Test-Path $FrontendPackageJson) {
    Write-Host "[INFO] Building frontend resources..." -ForegroundColor Green
    Push-Location (Join-Path $ProjectRoot "tlcpchan-ui")
        
    $NodeModules = Join-Path (Get-Location) "node_modules"
    if (-not (Test-Path $NodeModules)) {
        Write-Host "[INFO] Installing frontend dependencies..." -ForegroundColor Green
        npm ci
        if ($LASTEXITCODE -ne 0) {
            Pop-Location
            Write-Host "[ERROR] npm ci failed!" -ForegroundColor Red
            exit 1
        }
    }
        
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Pop-Location
        Write-Host "[ERROR] npm run build failed!" -ForegroundColor Red
        exit 1
    }
        
    Pop-Location
}
else {
    Write-Host "[WARN] Frontend directory does not exist, skipping frontend build" -ForegroundColor Yellow
}


# 编译 tlcpchan
Write-Host "[INFO] Compiling tlcpchan..." -ForegroundColor Green
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = "0"
Push-Location (Join-Path $ProjectRoot "tlcpchan")

# 生成资源文件（嵌入图标）
$ManifestFile = Join-Path $ScriptDir "tlcpchan.exe.manifest"
$SysoFile = Join-Path (Get-Location) "rsrc_windows_amd64.syso"
if ($HasIcon -and (Test-Path $ManifestFile)) {
    Write-Host "[INFO] Embedding icon for tlcpchan.exe..." -ForegroundColor Green
    & $RsrcExe -manifest $ManifestFile -ico $IconFile -o $SysoFile
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARN] Failed to embed icon for tlcpchan.exe, continuing without icon" -ForegroundColor Yellow
        if (Test-Path $SysoFile) {
            Remove-Item -Path $SysoFile -Force
        }
    } else {
        Write-Host "[INFO] Icon embedded successfully for tlcpchan.exe" -ForegroundColor Green
    }
}

go build -ldflags="-s -w" -o (Join-Path $OutputDir "tlcpchan.exe") .
if ($LASTEXITCODE -ne 0) {
    Pop-Location
    Write-Host "[ERROR] tlcpchan compilation failed!" -ForegroundColor Red
    exit 1
}

# 清理资源文件
if (Test-Path $SysoFile) {
    Remove-Item -Path $SysoFile -Force
}
Pop-Location

# 编译 tlcpchan-cli
Write-Host "[INFO] Compiling tlcpchan-cli..." -ForegroundColor Green
Push-Location (Join-Path $ProjectRoot "tlcpchan-cli")

# 生成资源文件（嵌入图标）
$CliManifestFile = Join-Path $ScriptDir "tlcpchan-cli.exe.manifest"
$CliSysoFile = Join-Path (Get-Location) "rsrc_windows_amd64.syso"
if ($HasIcon -and (Test-Path $CliManifestFile)) {
    Write-Host "[INFO] Embedding icon for tlcpchan-cli.exe..." -ForegroundColor Green
    & $RsrcExe -manifest $CliManifestFile -ico $IconFile -o $CliSysoFile
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARN] Failed to embed icon for tlcpchan-cli.exe, continuing without icon" -ForegroundColor Yellow
        if (Test-Path $CliSysoFile) {
            Remove-Item -Path $CliSysoFile -Force
        }
    } else {
        Write-Host "[INFO] Icon embedded successfully for tlcpchan-cli.exe" -ForegroundColor Green
    }
}

go build -ldflags="-s -w" -o (Join-Path $OutputDir "tlcpchan-cli.exe") .
if ($LASTEXITCODE -ne 0) {
    Pop-Location
    Write-Host "[ERROR] tlcpchan-cli compilation failed!" -ForegroundColor Red
    exit 1
}

# 清理资源文件
if (Test-Path $CliSysoFile) {
    Remove-Item -Path $CliSysoFile -Force
}
Pop-Location

# 复制前端资源
$FrontendUiDir = Join-Path $ProjectRoot "tlcpchan-ui\ui"
if (Test-Path $FrontendUiDir) {
    Write-Host "[INFO] Copying frontend resources..." -ForegroundColor Green
    Copy-Item -Path $FrontendUiDir -Destination (Join-Path $OutputDir "ui") -Recurse -Force
}

# 复制信任证书
$TrustedCertsDir = Join-Path $ProjectRoot "trustedcerts"
if (Test-Path $TrustedCertsDir) {
    Write-Host "[INFO] Copying trusted certificates..." -ForegroundColor Green
    Copy-Item -Path $TrustedCertsDir -Destination (Join-Path $OutputDir "rootcerts") -Recurse -Force
}

# 创建 zip 包
Write-Host "[INFO] Creating zip package..." -ForegroundColor Green
New-Item -ItemType Directory -Path $DistDir -Force | Out-Null
$ZipPath = Join-Path $DistDir "tlcpchan_${Version}_windows_amd64.zip"
Compress-Archive -Path (Join-Path $BuildDir "windows-amd64") -DestinationPath $ZipPath -Force
if ($LASTEXITCODE -ne 0) {
    Write-Host "[WARN] Failed to create zip package, skipping" -ForegroundColor Yellow
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build completed" -ForegroundColor Green
Write-Host "  Output directory: $OutputDir " -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# TLCP Channel Windows MSI 打包脚本
# 使用 WiX Toolset 自动下载和构建 MSI 安装包

param(
    [switch]$SkipDownload
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TLCP Channel MSI 打包脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 设置目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ReleaseDir = Join-Path $ScriptDir ".."
$ProjectRoot = Join-Path $ReleaseDir ".."
$WixDir = Join-Path $ReleaseDir "wix"
$BuildDir = Join-Path $ProjectRoot "build"
$DistDir = Join-Path $ProjectRoot "dist"
$OutputDir = Join-Path $BuildDir "windows-amd64"

# 从 tlcpchan/main.go 中解析版本号
$MainGo = Join-Path $ProjectRoot "tlcpchan" "main.go"
if (-not (Test-Path $MainGo)) {
    Write-Host "[ERROR] main.go not found!" -ForegroundColor Red
    exit 1
}

$VersionLine = Get-Content $MainGo | Select-String -Pattern 'var\s+version\s*=' | Select-Object -First 1
if (-not $VersionLine) {
    Write-Host "[ERROR] Failed to find version in main.go!" -ForegroundColor Red
    exit 1
}

$Version = $VersionLine -replace '.*version\s*=\s*"([^"]+)".*', '$1'

Write-Host "[INFO] 版本: $Version" -ForegroundColor Green
Write-Host "[INFO] 项目根目录: $ProjectRoot" -ForegroundColor Green

# 检查是否已经有编译产物
if (-not (Test-Path $OutputDir)) {
    Write-Host "[ERROR] 未找到编译产物，请先运行 build.bat" -ForegroundColor Red
    exit 1
}

# WiX 配置
$WixVersion = "3.14.1"
$WixZipUrl = "https://github.com/wixtoolset/wix3/releases/download/wix3141rtm/wix314-binaries.zip"
$WixZipPath = Join-Path $WixDir "wix314-binaries.zip"
$WixToolPath = Join-Path $WixDir "wix-bin"

# 下载 WiX
if (-not $SkipDownload) {
    if (-not (Test-Path $WixToolPath)) {
        Write-Host "[INFO] 正在下载 WiX Toolset v$WixVersion..." -ForegroundColor Yellow
        
        if (-not (Test-Path $WixDir)) {
            New-Item -ItemType Directory -Path $WixDir | Out-Null
        }
        
        if (-not (Test-Path $WixZipPath)) {
            try {
                Invoke-WebRequest -Uri $WixZipUrl -OutFile $WixZipPath -UseBasicParsing
            } catch {
                Write-Host "[ERROR] WiX 下载失败: $_" -ForegroundColor Red
                exit 1
            }
        }
        
        Write-Host "[INFO] 解压 WiX..." -ForegroundColor Yellow
        Expand-Archive -Path $WixZipPath -DestinationPath $WixToolPath -Force
    }
}

# 检查 WiX
$CandlePath = Join-Path $WixToolPath "candle.exe"
$LightPath = Join-Path $WixToolPath "light.exe"

if (-not (Test-Path $CandlePath) -or -not (Test-Path $LightPath)) {
    Write-Host "[ERROR] 未找到 WiX 工具，请检查: $WixToolPath" -ForegroundColor Red
    exit 1
}

Write-Host "[INFO] 使用 WiX: $WixToolPath" -ForegroundColor Green

# 创建临时的 WiX 源文件目录
$WixSourceDir = Join-Path $BuildDir "wix-source"
if (Test-Path $WixSourceDir) {
    Remove-Item -Path $WixSourceDir -Recurse -Force
}
New-Item -ItemType Directory -Path $WixSourceDir | Out-Null

# 复制编译产物到 WiX 源目录
Write-Host "[INFO] 准备 WiX 源文件..." -ForegroundColor Yellow
Copy-Item -Path (Join-Path $OutputDir "*") -Destination $WixSourceDir -Recurse

# 创建 WiX .wxs 文件
$WxsPath = Join-Path $WixSourceDir "tlcpchan.wxs"

$WxsContent = @"
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" Name="TLCP Channel" Language="1033" Version="$Version" Manufacturer="TLCP Channel Team" UpgradeCode="12345678-1234-1234-1234-123456789012">
    <Package InstallerVersion="200" Compressed="yes" InstallScope="perMachine" />
    
    <MajorUpgrade DowngradeErrorMessage="A newer version of [ProductName] is already installed." />
    <MediaTemplate EmbedCab="yes" />
    
    <Feature Id="ProductFeature" Title="TLCP Channel" Level="1">
      <ComponentGroupRef Id="ProductComponents" />
    </Feature>
    
    <Property Id="WIXUI_INSTALLDIR" Value="INSTALLFOLDER" />
    <UIRef Id="WixUI_Minimal" />
    
    <Environment Id="PathEnv" Name="PATH" Value="[INSTALLFOLDER]" Permanent="no" Part="last" Action="set" System="yes" />
  </Product>
  
  <Fragment>
    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="TLCP Channel" />
      </Directory>
      <Directory Id="CommonAppDataFolder">
        <Directory Id="APPDATAFOLDER" Name="TLCP Channel">
          <Directory Id="KEYSTORESFOLDER" Name="keystores" />
          <Directory Id="LOGSFOLDER" Name="logs" />
          <Directory Id="ROOTCERTSFOLDER" Name="rootcerts" />
        </Directory>
      </Directory>
    </Directory>
  </Fragment>
  
  <Fragment>
    <ComponentGroup Id="ProductComponents" Directory="INSTALLFOLDER">
      <Component Id="tlcpchan_exe" Guid="{$(New-Guid)}">
        <File Id="tlcpchan_exe" Name="tlcpchan.exe" Source="tlcpchan.exe" KeyPath="yes" />
      </Component>
      <Component Id="tlcpchan_cli_exe" Guid="{$(New-Guid)}">
        <File Id="tlcpchan_cli_exe" Name="tlcpchan-cli.exe" Source="tlcpchan-cli.exe" />
      </Component>
      <Component Id="tlcpchan_ui_exe" Guid="{$(New-Guid)}">
        <File Id="tlcpchan_ui_exe" Name="tlcpchan-ui.exe" Source="tlcpchan-ui.exe" />
      </Component>
      <Component Id="install_bat" Guid="{$(New-Guid)}">
        <File Id="install_bat" Name="install.bat" Source="install.bat" />
      </Component>
      <Component Id="uninstall_bat" Guid="{$(New-Guid)}">
        <File Id="uninstall_bat" Name="uninstall.bat" Source="uninstall.bat" />
      </Component>
    </ComponentGroup>
  </Fragment>
</Wix>
"@

$WxsContent | Out-File -FilePath $WxsPath -Encoding utf8

# 编译 WiX
Write-Host "[INFO] 编译 WiX..." -ForegroundColor Yellow

Push-Location $WixSourceDir

try {
    & $CandlePath -nologo tlcpchan.wxs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] candle.exe 执行失败" -ForegroundColor Red
        exit 1
    }
    
    & $LightPath -nologo -ext WixUIExtension -out "tlcpchan_$Version.msi" tlcpchan.wixobj
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] light.exe 执行失败" -ForegroundColor Red
        exit 1
    }
    
    # 复制 MSI 到 dist 目录
    if (-not (Test-Path $DistDir)) {
        New-Item -ItemType Directory -Path $DistDir | Out-Null
    }
    
    $MsiOutputPath = Join-Path $DistDir "tlcpchan_${Version}_windows_amd64.msi"
    Copy-Item -Path "tlcpchan_$Version.msi" -Destination $MsiOutputPath -Force
    
    Write-Host "[INFO] MSI 包已创建: $MsiOutputPath" -ForegroundColor Green
    
} finally {
    Pop-Location
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MSI 打包完成！" -ForegroundColor Green
Write-Host "  输出目录: $DistDir" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

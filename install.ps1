# ============================================
# TLCP Channel One-Click Installer (Windows)
# GitHub: https://github.com/Trisia/tlcpchan
# ============================================

# Set output encoding to UTF-8 to avoid character encoding issues
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# Error handling
$ErrorActionPreference = "Stop"

# Project configuration
$RepoOwner = "Trisia"
$RepoName = "tlcpchan"
$GitHubApi = "https://api.github.com/repos/${RepoOwner}/${RepoName}"
$InstallDir = "C:\Program Files\TLCP Channel"

# Write colored error message
function Write-ErrorColored {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

# Write colored success message
function Write-SuccessColored {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

# Write colored info message
function Write-InfoColored {
    param([string]$Message)
    Write-Host "  $Message" -ForegroundColor Yellow
}

# Write plain message
function Write-Plain {
    param([string]$Message)
    Write-Host $Message
}

# Detect system architecture
function Detect-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($arch -ne "AMD64") {
        Write-ErrorColored "TLCP Channel only supports x64 architecture"
        Write-Plain ""
        Write-InfoColored "Current architecture: $arch"
        exit 1
    }
    return "amd64"
}

# Check if already installed
function Check-Installed {
    if (Test-Path $InstallDir) {
        Write-ErrorColored "TLCP Channel is already installed at $InstallDir"
        Write-Plain ""
        Write-InfoColored "To reinstall, please uninstall the old version first"
        Write-Plain ""
        Write-InfoColored "Uninstall steps:"
        Write-InfoColored "  1. Stop the application if running"
        Write-InfoColored "  2. Uninstall via Control Panel or run: msiexec /x {ProductCode}"
        Write-InfoColored "  3. Manually delete directory if needed: $InstallDir"
        exit 1
    }
}

# Check administrator permissions
function Check-Permissions {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    if (-not $principal.IsInRole([Security.Principal.WindowsBuiltinRole]::Administrator)) {
        Write-ErrorColored "Administrator privileges are required"
        Write-Plain ""
        Write-InfoColored "Please run PowerShell as Administrator and try again"
        Write-InfoColored "Right-click PowerShell -> 'Run as Administrator'"
        exit 1
    }
}

# Get latest version from GitHub
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "$GitHubApi/releases/latest" -Headers @{"Accept"="application/vnd.github.v3+json"}
        $version = $response.tag_name -replace '^v', ''
        if ([string]::IsNullOrEmpty($version)) {
            Write-ErrorColored "Failed to get latest version information"
            exit 1
        }
        return $version
    } catch {
        Write-ErrorColored "Failed to get latest version information: $_"
        exit 1
    }
}

# Build download URL for MSI package
function Build-DownloadUrl {
    param([string]$Version)

    # MSI version format: add .0 suffix (e.g., 1.0.1 -> 1.0.1.0)
    $msiVersion = "${Version}.0"
    $filename = "tlcpchan_${msiVersion}_windows_amd64.msi"
    $url = "https://github.com/${RepoOwner}/${RepoName}/releases/download/v${Version}/${filename}"
    return $url
}

# Download MSI package
function Download-Package {
    param([string]$Url, [string]$OutputDir)

    $filename = [System.IO.Path]::GetFileName($Url)
    $outputPath = Join-Path $OutputDir $filename

    Write-Plain "📥 Downloading installation package..."

    try {
        Invoke-WebRequest -Uri $Url -OutFile $outputPath -UseBasicParsing
    } catch {
        Write-ErrorColored "Failed to download installation package: $_"
        return $null
    }

    # Check file size
    if (-not (Test-Path $outputPath) -or (Get-Item $outputPath).Length -eq 0) {
        Write-ErrorColored "Downloaded file is empty"
        Remove-Item $outputPath -ErrorAction SilentlyContinue
        return $null
    }

    Write-Plain ""
    return $outputPath
}

# Install MSI package silently
function Install-MSI {
    param([string]$MsiPath)

    Write-Plain "📦 Installing TLCP Channel..."
    Write-Plain ""

    try {
        $process = Start-Process -FilePath "msiexec.exe" -ArgumentList "/i", "`"$MsiPath`"", "/qn", "/norestart" -Wait -PassThru -NoNewWindow

        if ($process.ExitCode -ne 0) {
            Write-ErrorColored "MSI installation failed with exit code: $($process.ExitCode)"
            Write-Plain ""
            Write-InfoColored "Please check Windows Event Viewer for detailed error logs"
            return $false
        }

        return $true
    } catch {
        Write-ErrorColored "Failed to install MSI package: $_"
        return $false
    }
}

# Verify installation
function Verify-Install {
    param([string]$InstallDir)

    Write-Plain "🔍 Verifying installation..."

    $tlcpchanPath = Join-Path $InstallDir "tlcpchan.exe"

    if (-not (Test-Path $tlcpchanPath)) {
        Write-ErrorColored "Installation verification failed: tlcpchan.exe not found"
        return $false
    }

    try {
        $output = & $tlcpchanPath -version 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-ErrorColored "Failed to execute tlcpchan.exe"
            return $false
        }

        # Extract version number from output
        if ($output -match '(\d+\.\d+\.\d+)') {
            $installedVersion = $matches[1]
            Write-Plain ""
            Write-Plain "✓ Version: $installedVersion"
            Write-Plain ""
            return $true
        } else {
            Write-ErrorColored "Unable to parse version information"
            return $false
        }
    } catch {
        Write-ErrorColored "Installation verification failed: $_"
        return $false
    }
}

# Show success information
function Show-SuccessInfo {
    param([string]$Version)

    Write-SuccessColored "TLCP Channel Installation successful!"
    Write-Plain ""
    Write-Plain "📦 Version Information"
    Write-InfoColored "Version: $Version"
    Write-Plain ""
    Write-Plain "📂 Installation Location"
    Write-InfoColored "Install Directory: $InstallDir"
    Write-InfoColored "Executable: $InstallDir\tlcpchan.exe"
    Write-InfoColored "CLI Tool: $InstallDir\tlcpchan-cli.exe"
    Write-Plain ""
    Write-Plain "🚀 Start Service"
    Write-Plain "  Run: cd `"$InstallDir`"; .\tlcpchan.exe"
    Write-Plain "  Background: Start-Process -FilePath `"$InstallDir\tlcpchan.exe`" -WindowStyle Hidden"
    Write-Plain ""
    Write-Plain "🌐 Web UI"
    Write-InfoColored "http://localhost:20080"
    Write-Plain ""
    Write-Plain "📖 More Information"
    Write-InfoColored "Documentation: https://github.com/Trisia/tlcpchan"
}

# Main function
function Main {
    # Detect architecture
    $arch = Detect-Arch
    Write-Plain "🔍 Detecting system environment..."
    Write-Plain ""
    Write-Plain "  OS: Windows"
    Write-Plain "  Architecture: $arch"
    Write-Plain ""

    # Check if already installed
    Check-Installed

    # Check permissions
    Check-Permissions

    # Get latest version
    Write-Plain "🔍 Getting latest version information..."
    $version = Get-LatestVersion
    Write-InfoColored "Latest version: v$version"
    Write-Plain ""

    # Build download URL
    $downloadUrl = Build-DownloadUrl -Version $version

    # Create temporary directory
    $tmpDir = Join-Path $env:TEMP "tlcpMSI_install_$([Guid]::NewGuid())"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download MSI package
        $msiPath = Download-Package -Url $downloadUrl -OutputDir $tmpDir
        if ($null -eq $msiPath) {
            exit 1
        }

        # Install MSI package
        if (-not (Install-MSI -MsiPath $msiPath)) {
            Write-Plain ""
            Write-ErrorColored "Installation failed, please check error messages"
            exit 1
        }

        # Verify installation
        if (-not (Verify-Install -InstallDir $InstallDir)) {
            Write-Plain ""
            Write-ErrorColored "Installation verification failed"
            Write-Plain ""
            Write-InfoColored "Please check if TLCP Channel was installed correctly in:"
            Write-InfoColored "  $InstallDir"
            exit 1
        }

        # Show success information
        Show-SuccessInfo -Version $version

    } finally {
        # Cleanup temporary directory
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Execute main function
Main

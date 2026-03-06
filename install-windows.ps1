# Simple installation script for expe3000-go on Windows
# This script must be run as Administrator.

$InstallDir = "C:\Program Files\expe3000-go"
$Executables = @("expe3000.exe", "expe3000-gui.exe")

# 1. Create the installation directory if it doesn't exist
if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating installation directory: $InstallDir" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# 2. Copy the executables
foreach ($exe in $Executables) {
    if (Test-Path $exe) {
        Write-Host "Installing $exe..." -ForegroundColor Green
        Copy-Item -Path $exe -Destination $InstallDir -Force
    } else {
        Write-Warning "Could not find $exe in the current folder. Skipping."
    }
}

# 3. Add to System PATH
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to the system PATH..." -ForegroundColor Cyan
    $NewPath = "$CurrentPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
    Write-Host "Success! You may need to restart your terminal for changes to take effect." -ForegroundColor Green
} else {
    Write-Host "$InstallDir is already in the PATH." -ForegroundColor Yellow
}

Write-Host "`nInstallation complete. You can now run 'expe3000' or 'expe3000-gui' from any terminal." -ForegroundColor White
pause

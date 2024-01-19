$ErrorActionPreference = "Stop"

$PSScriptRoot = Split-Path -Path $PSCommandPath -Parent

Import-Module -WarningAction Ignore -Name "$PSScriptRoot\common.psm1" -Force

$buildData = Get-BuildData
$os = $buildData.OS
$arch = $buildData.ARCH

$ccgPluginInstaller = "./bin/ccg-plugin-installer-$os-$arch.exe"

# Ensure that before the DLL is installed, all expected files are not present, the COM class is not registered, and the DLL's CLSID is not registered as expected.

Write-Host "Ensuring that current machine does not have DLL installed..." -ForegroundColor Yellow

Check-DLLUninstalled

# Ensure that once the DLL is installed, all expected files are present, the COM class is registered, and the DLL's CLSID is registered as expected.

Write-Host "Installing DLL..." -ForegroundColor Yellow

& "$ccgPluginInstaller" install --debug

Write-Host "Checking if DLL is installed..." -ForegroundColor Yellow

Check-DLLInstalled

# Ensure that once the DLL can be re-installed

Write-Host "Re-installing DLL..." -ForegroundColor Yellow

& "$ccgPluginInstaller" install --debug

Write-Host "Checking if DLL is installed..." -ForegroundColor Yellow

Check-DLLInstalled

# Ensure that after the DLL is uninstalled, all expected files are not present, the COM class is not registered, and the DLL's CLSID is not registered as expected.

Write-Host "Uninstalling DLL..." -ForegroundColor Yellow

& "$ccgPluginInstaller" uninstall --debug

Write-Host "Checking if DLL is uninstalled..." -ForegroundColor Yellow

Check-DLLUninstalled

Write-Host "SUCCESS: Verified that DLL can be installed and uninstalled." -ForegroundColor Green

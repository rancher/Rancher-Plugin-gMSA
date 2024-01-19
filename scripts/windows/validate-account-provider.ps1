$ErrorActionPreference = "Stop"

$PSScriptRoot = Split-Path -Path $PSCommandPath -Parent

Import-Module -WarningAction Ignore -Name "$PSScriptRoot\common.psm1" -Force

$buildData = Get-BuildData
$os = $buildData.OS
$arch = $buildData.ARCH

$gmsaAccountProvider = "./bin/gmsa-account-provider-$os-$arch.exe"

# Ensure that a cleanup operation successfully runs on a machine that has never had the Account Provider installed.

Write-Host "Ensuring cleanup can be run on clean host..." -ForegroundColor Yellow

& "$gmsaAccountProvider" cleanup

# TODO: Ensure that a run operation successfully works on a machine with or without certs.
# Requires a Kubernetes cluster to be running.
# & C:\Users\adminuser\gmsa-account-provider-$suffix.exe run
# 
Write-Host "SUCCESS: Verified that Account Provider can be cleaned up." -ForegroundColor Green

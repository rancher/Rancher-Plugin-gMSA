$ErrorActionPreference = "Stop"

$PSScriptRoot = Split-Path -Path $PSCommandPath -Parent

Import-Module -WarningAction Ignore -Name "$PSScriptRoot\common.psm1" -Force

Clone-Directory -From "bin" -To "dist"

Docker-Build -Apps (Get-ChildItem -Path ./package -Directory)

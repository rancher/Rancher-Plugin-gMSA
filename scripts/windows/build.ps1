$ErrorActionPreference = "Stop"

$PSScriptRoot = Split-Path -Path $PSCommandPath -Parent

Import-Module -WarningAction Ignore -Name "$PSScriptRoot\common.psm1" -Force

Go-Build -Apps (Get-ChildItem -Path ./cmd -Directory)

$ErrorActionPreference = "Stop"

$PSScriptRoot = Split-Path -Path $PSCommandPath -Parent

Import-Module -WarningAction Ignore -Name "$PSScriptRoot\common.psm1" -Force

Write-Host "Running tests" -ForegroundColor yellow

go test -cover -tags=test ./...

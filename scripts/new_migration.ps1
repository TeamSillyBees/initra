[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern("^[a-zA-Z0-9_]+$")]
    [string]$Name
)

$ErrorActionPreference = "Stop"

# 迁移文件仍然由开发者手写；该脚本只负责生成带时间戳的 Atlas versioned migration 模板。
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$MigrationDir = Join-Path $RepoRoot "db\migrations"
$Timestamp = Get-Date -Format "yyyyMMddHHmmss"
$File = Join-Path $MigrationDir "$Timestamp`_$Name.sql"

if (-not (Test-Path $MigrationDir)) {
    New-Item -ItemType Directory -Path $MigrationDir | Out-Null
}

@"
-- 请在这里编写 Atlas versioned migration SQL。
"@ | Set-Content -Path $File -Encoding UTF8

Write-Host "created migration: $File"

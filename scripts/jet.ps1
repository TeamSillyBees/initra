[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

# Jet generator 会清空目标目录后重新生成，请不要把手写代码放进 internal/gen/jet。
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$JetBin = if ([string]::IsNullOrWhiteSpace($env:JET_BIN)) { "jet" } else { $env:JET_BIN }
$JetDsn = if ([string]::IsNullOrWhiteSpace($env:JET_DSN)) { "postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable" } else { $env:JET_DSN }
$JetSchema = if ([string]::IsNullOrWhiteSpace($env:JET_SCHEMA)) { "public" } else { $env:JET_SCHEMA }
$JetPath = if ([string]::IsNullOrWhiteSpace($env:JET_PATH)) { "./internal/gen/jet" } else { $env:JET_PATH }

Push-Location $RepoRoot
try {
    & $JetBin "-dsn=$JetDsn" "-schema=$JetSchema" "-path=$JetPath"
    if ($LASTEXITCODE -ne 0) {
        throw "jet failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

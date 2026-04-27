[CmdletBinding()]
param(
    [string]$Env = $(if ([string]::IsNullOrWhiteSpace($env:ENV)) { "local" } else { $env:ENV })
)

$ErrorActionPreference = "Stop"

# Windows/PowerShell 一键开发入口：整理依赖、启动中间件、执行迁移、生成 Jet 代码并运行服务。
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")

Push-Location $RepoRoot
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        throw "go mod tidy failed with exit code $LASTEXITCODE"
    }

    docker compose up -d postgres redis
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose up failed with exit code $LASTEXITCODE"
    }

    & (Join-Path $PSScriptRoot "atlas.ps1") migrate apply --env $Env
    & (Join-Path $PSScriptRoot "jet.ps1")

    $env:APP_ENV = $Env
    go run ./cmd/server
    if ($LASTEXITCODE -ne 0) {
        throw "go run failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

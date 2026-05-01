param(
    [string]$Env = $env:APP_ENV,
    [string]$ConfigDir = ".\configs",
    [string]$Schema = "public",
    [string]$Dest = ".\internal\gen\jet"
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($Env)) {
    $Env = "dev"
}

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    go run ./tools/jetgen -env $Env -config-dir $ConfigDir -schema $Schema -dest $Dest
}
finally {
    Pop-Location
}

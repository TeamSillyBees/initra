param(
    [string]$Out = (Join-Path $env:TEMP "initra-basic-server.exe")
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    go build -o $Out ./cmd/server
}
finally {
    Pop-Location
}

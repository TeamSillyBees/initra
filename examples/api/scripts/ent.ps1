param()

$ErrorActionPreference = "Stop"

Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    go generate ./internal/ent
} finally {
    Pop-Location
}


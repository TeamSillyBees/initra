param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$AtlasArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    atlas -c file://db/atlas.hcl @AtlasArgs
}
finally {
    Pop-Location
}

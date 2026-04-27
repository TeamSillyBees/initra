[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$AtlasArgs
)

$ErrorActionPreference = "Stop"

# 固定通过 db/atlas.hcl 读取 Atlas 环境配置，避免不同工作目录下命令行为不一致。
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$AtlasBin = if ([string]::IsNullOrWhiteSpace($env:ATLAS_BIN)) { "atlas" } else { $env:ATLAS_BIN }
$ConfigPath = if ([string]::IsNullOrWhiteSpace($env:CONFIG_PATH)) { "file://db/atlas.hcl" } else { $env:CONFIG_PATH }

Push-Location $RepoRoot
try {
    & $AtlasBin -c $ConfigPath @AtlasArgs
    if ($LASTEXITCODE -ne 0) {
        throw "atlas failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

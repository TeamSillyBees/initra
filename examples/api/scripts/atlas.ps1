param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$AtlasArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    if ($AtlasArgs.Count -ge 2 -and $AtlasArgs[0] -eq "migrate" -and $AtlasArgs[1] -eq "diff") {
        $name = $null
        $envName = $null
        $configDir = "configs"
        $devURL = $null
        for ($i = 2; $i -lt $AtlasArgs.Count; $i++) {
            $arg = $AtlasArgs[$i]
            if ($arg -eq "--dev-url" -or $arg -eq "-dev-url") {
                $i++
                if ($i -ge $AtlasArgs.Count) {
                    throw "$arg requires a value"
                }
                $devURL = $AtlasArgs[$i]
            }
            elseif ($arg.StartsWith("--dev-url=")) {
                $devURL = $arg.Substring("--dev-url=".Length)
            }
            elseif ($arg.StartsWith("-dev-url=")) {
                $devURL = $arg.Substring("-dev-url=".Length)
            }
            elseif ($arg -eq "--env" -or $arg -eq "-env" -or $arg -eq "-e") {
                $i++
                if ($i -ge $AtlasArgs.Count) {
                    throw "$arg requires a value"
                }
                $envName = $AtlasArgs[$i]
            }
            elseif ($arg.StartsWith("--env=")) {
                $envName = $arg.Substring("--env=".Length)
            }
            elseif ($arg.StartsWith("-env=")) {
                $envName = $arg.Substring("-env=".Length)
            }
            elseif ($arg -eq "--config-dir" -or $arg -eq "-config-dir") {
                $i++
                if ($i -ge $AtlasArgs.Count) {
                    throw "$arg requires a value"
                }
                $configDir = $AtlasArgs[$i]
            }
            elseif ($arg.StartsWith("--config-dir=")) {
                $configDir = $arg.Substring("--config-dir=".Length)
            }
            elseif ($arg.StartsWith("-config-dir=")) {
                $configDir = $arg.Substring("-config-dir=".Length)
            }
            elseif ($arg.StartsWith("-")) {
                throw "unsupported migrate diff option $arg; use --env, --config-dir, --dev-url or run atlas directly"
            }
            elseif ($null -eq $name) {
                $name = $arg
            }
            else {
                throw "migrate diff accepts one migration name"
            }
        }
        if ([string]::IsNullOrWhiteSpace($name)) {
            throw "migrate diff requires a migration name"
        }
        $optionalArgs = @()
        if (![string]::IsNullOrWhiteSpace($envName)) {
            $optionalArgs += @("-env", $envName)
        }
        if (![string]::IsNullOrWhiteSpace($devURL)) {
            $optionalArgs += @("-dev-url", $devURL)
        }
        go run ./internal/ent/migratediff/main.go $name -config-dir $configDir @optionalArgs
        return
    }
    atlas -c file://db/atlas.hcl @AtlasArgs
}
finally {
    Pop-Location
}

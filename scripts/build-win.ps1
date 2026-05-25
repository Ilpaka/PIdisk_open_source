param()

$ErrorActionPreference = "Stop"

Set-Location (Join-Path $PSScriptRoot "..")

$version = (& git describe --tags --always --dirty 2>$null)
if (-not $version) { $version = "dev" }
$commit = (& git rev-parse --short HEAD 2>$null)
if (-not $commit) { $commit = "unknown" }
$buildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$ldFlags = "-X github.com/pidisk/pidisk/internal/version.Version=$version " +
           "-X github.com/pidisk/pidisk/internal/version.Commit=$commit " +
           "-X github.com/pidisk/pidisk/internal/version.BuildTime=$buildTime"

wails build -clean -platform "windows/amd64" -ldflags $ldFlags

[CmdletBinding()]
param(
    [ValidatePattern('^\d+(\.\d+){0,2}([\-+][0-9A-Za-z.-]+)?$')]
    [string]$AppVersion
)

$ErrorActionPreference = 'Stop'

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$go = Get-Command 'go' -ErrorAction SilentlyContinue
if ($null -eq $go) {
    $defaultGo = 'C:\Program Files\Go\bin\go.exe'
    if (Test-Path $defaultGo) {
        $go = Get-Item $defaultGo
    }
    else {
        throw "Required command 'go' was not found. See docs/build-windows.md."
    }
}

$icon = Join-Path $repositoryRoot 'app.png'
if (-not (Test-Path $icon)) {
    throw "Packaging icon was not found: $icon"
}

$fyneModule = 'fyne.io/tools/cmd/fyne@v1.7.2'
$packageArguments = @(
    'run', $fyneModule, 'package',
    '--target', 'windows',
    '--src', $repositoryRoot,
    '--name', 'CG',
    '--icon', $icon,
    '--app-id', 'com.github.g70245.cg',
    '--release'
)
if ($AppVersion) {
    $packageArguments += @('--app-version', $AppVersion)
}

Push-Location $repositoryRoot
try {
    Write-Host "Go: $(& $go.Source version)"
    Write-Host 'Packaging with fyne.io/tools v1.7.2'

    & $go.Source @packageArguments
    if ($LASTEXITCODE -ne 0) {
        throw "Fyne packaging failed with exit code $LASTEXITCODE."
    }

    $packageOutput = Join-Path $repositoryRoot 'CG.exe'
    if (-not (Test-Path $packageOutput)) {
        throw "Fyne completed without producing the expected package: $packageOutput"
    }

    $outputDirectory = Join-Path $repositoryRoot 'dist'
    $output = Join-Path $outputDirectory 'CG.exe'
    New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
    Move-Item -Force -Path $packageOutput -Destination $output

    Write-Host "Package succeeded: $output"
}
finally {
    Pop-Location
}

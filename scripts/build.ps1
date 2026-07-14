[CmdletBinding()]
param(
    [switch]$SkipDependencyDownload
)

$ErrorActionPreference = 'Stop'

function Get-RequiredCommand {
    param(
        [Parameter(Mandatory)]
        [string]$Name,
        [string[]]$FallbackPaths = @()
    )

    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -ne $command) {
        return $command.Source
    }

    foreach ($path in $FallbackPaths) {
        if (Test-Path $path) {
            return $path
        }
    }

    throw "Required command '$Name' was not found. See docs/build-windows.md."
}

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$go = Get-RequiredCommand -Name 'go' -FallbackPaths @('C:\Program Files\Go\bin\go.exe')
$gcc = Get-RequiredCommand -Name 'gcc'

Push-Location $repositoryRoot
try {
    Write-Host "Go:  $(& $go version)"
    Write-Host "GCC: $((& $gcc --version)[0])"

    if (-not (Test-Path 'go.sum')) {
        throw 'go.sum is missing. Restore it from version control before building.'
    }

    if (-not $SkipDependencyDownload) {
        & $go mod download
    }
    & $go mod verify

    $outputDirectory = Join-Path $repositoryRoot 'dist'
    $output = Join-Path $outputDirectory 'cg.exe'
    New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

    & $go build -trimpath -o $output .
    if ($LASTEXITCODE -ne 0) {
        throw "Go build failed with exit code $LASTEXITCODE."
    }

    Write-Host "Build succeeded: $output"
}
finally {
    Pop-Location
}

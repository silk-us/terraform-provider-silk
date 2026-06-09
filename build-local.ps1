<#
    .SYNOPSIS
    Build terraform-provider-silk against a LOCAL silk-sdp-go-sdk checkout
    instead of the version pinned in go.mod. Restores go.mod / go.sum on exit
    so nothing accidentally gets committed.

    .DESCRIPTION
    PowerShell port of build-local.sh for Windows hosts. Reads VERSION and
    BINARY from the Makefile, injects a `replace` directive pointing at the
    local SDK, runs `go mod tidy`, builds, and rolls go.mod / go.sum back.

    .PARAMETER All
    Build every release target (darwin/linux/freebsd/openbsd/solaris/windows).

    .PARAMETER Install
    Build for the host and copy into %APPDATA%\terraform.d\plugins\... so
    Terraform can pick the provider up locally.

    .PARAMETER Sdk
    Path to the local SDK checkout. Default: ..\silk-sdp-go-sdk relative to
    this script.

    .EXAMPLE
    .\build-local.ps1
    # Build for current host (windows_amd64 etc.)

    .EXAMPLE
    .\build-local.ps1 -All

    .EXAMPLE
    .\build-local.ps1 -Install

    .EXAMPLE
    .\build-local.ps1 -Sdk ..\some\other\silk-sdp-go-sdk
#>

[CmdletBinding(DefaultParameterSetName = 'Host')]
param(
    [Parameter(ParameterSetName = 'All')]
    [switch] $All,

    [Parameter(ParameterSetName = 'Install')]
    [switch] $Install,

    [Parameter()]
    [string] $Sdk
)

$ErrorActionPreference = 'Stop'

# -------------------------------------------------------------------------
# Paths
# -------------------------------------------------------------------------

$providerDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $Sdk) { $Sdk = Join-Path $providerDir '..\silk-sdp-go-sdk' }

if (-not (Test-Path (Join-Path $Sdk 'silksdp') -PathType Container)) {
    Write-Error "SDK not found at $Sdk (expected silksdp\ inside)"
}
$sdkDir = (Resolve-Path $Sdk).Path

# Mode
$mode = 'host'
if ($All)     { $mode = 'all' }
if ($Install) { $mode = 'install' }

# -------------------------------------------------------------------------
# Read VERSION / BINARY from Makefile so this stays in sync
# -------------------------------------------------------------------------

function Get-MakefileVar {
    param(
        [Parameter(Mandatory)][string] $Path,
        [Parameter(Mandatory)][string] $Name
    )
    $line = Select-String -Path $Path -Pattern "^$Name\s*=" | Select-Object -First 1
    if (-not $line) { Write-Error "Could not find $Name in $Path" }
    ($line.Line -split '=', 2)[1].Trim()
}

$makefile = Join-Path $providerDir 'Makefile'
$version  = Get-MakefileVar -Path $makefile -Name 'VERSION'
$name     = Get-MakefileVar -Path $makefile -Name 'NAME'
$binary   = (Get-MakefileVar -Path $makefile -Name 'BINARY') -replace '\$\{NAME\}', $name

# -------------------------------------------------------------------------
# Snapshot go.mod / go.sum and arrange for restore
# -------------------------------------------------------------------------

Set-Location $providerDir
$null = New-Item -ItemType Directory -Path (Join-Path $providerDir 'bin') -Force

$goMod    = Join-Path $providerDir 'go.mod'
$goSum    = Join-Path $providerDir 'go.sum'
$goModBak = "$goMod.bak"
$goSumBak = "$goSum.bak"

Copy-Item $goMod $goModBak -Force
Copy-Item $goSum $goSumBak -Force

function Restore-GoMod {
    if (Test-Path $script:goModBak) { Move-Item $script:goModBak $script:goMod -Force }
    if (Test-Path $script:goSumBak) { Move-Item $script:goSumBak $script:goSum -Force }
    Write-Host "Restored go.mod / go.sum"
}

function Invoke-Go {
    param([Parameter(Mandatory)][string[]] $GoArgs)
    & go @GoArgs
    if ($LASTEXITCODE -ne 0) {
        throw "go $($GoArgs -join ' ') failed with exit $LASTEXITCODE"
    }
}

try {
    # ---------------------------------------------------------------------
    # Inject replace directive and tidy
    # ---------------------------------------------------------------------

    Invoke-Go -GoArgs @('mod', 'edit', "-replace=github.com/silk-us/silk-sdp-go-sdk=$sdkDir")
    Invoke-Go -GoArgs @('mod', 'tidy')

    Write-Host "Building against local SDK at: $sdkDir"
    Write-Host "Provider version: $version"

    switch ($mode) {
        'host' {
            $hostOs   = (& go env GOOS).Trim()
            $hostArch = (& go env GOARCH).Trim()
            $ext = if ($hostOs -eq 'windows') { '.exe' } else { '' }
            $out = ".\bin\${binary}_${version}_${hostOs}_${hostArch}${ext}"
            Invoke-Go -GoArgs @('build', '-o', $out)
            Write-Host "Built: $out"
        }

        'install' {
            $hostOs   = (& go env GOOS).Trim()
            $hostArch = (& go env GOARCH).Trim()
            $ext = if ($hostOs -eq 'windows') { '.exe' } else { '' }

            # Terraform on Windows looks under %APPDATA%\terraform.d\plugins
            $pluginRoot = if ($env:APPDATA) {
                Join-Path $env:APPDATA 'terraform.d\plugins'
            } else {
                Join-Path $env:USERPROFILE '.terraform.d\plugins'
            }
            $installDir = Join-Path $pluginRoot "localdomain\provider\silk\$version\${hostOs}_${hostArch}"
            $null = New-Item -ItemType Directory -Path $installDir -Force

            $installedBinary = Join-Path $installDir "${binary}${ext}"
            Invoke-Go -GoArgs @('build', '-o', $installedBinary)

            $out = ".\bin\${binary}_${version}_${hostOs}_${hostArch}${ext}"
            Copy-Item $installedBinary $out -Force
            Write-Host "Built: $out"
            Write-Host "Installed: $installedBinary"
        }

        'all' {
            $targets = @(
                @{ os = 'darwin';  arch = 'amd64' }
                @{ os = 'darwin';  arch = 'arm64' }
                @{ os = 'freebsd'; arch = '386'   }
                @{ os = 'freebsd'; arch = 'amd64' }
                @{ os = 'freebsd'; arch = 'arm'   }
                @{ os = 'linux';   arch = '386'   }
                @{ os = 'linux';   arch = 'amd64' }
                @{ os = 'linux';   arch = 'arm'   }
                @{ os = 'openbsd'; arch = '386'   }
                @{ os = 'openbsd'; arch = 'amd64' }
                @{ os = 'solaris'; arch = 'amd64' }
                @{ os = 'windows'; arch = '386'   }
                @{ os = 'windows'; arch = 'amd64' }
            )
            $savedGoos   = $env:GOOS
            $savedGoarch = $env:GOARCH
            try {
                foreach ($t in $targets) {
                    $env:GOOS   = $t.os
                    $env:GOARCH = $t.arch
                    $ext = if ($t.os -eq 'windows') { '.exe' } else { '' }
                    $out = ".\bin\${binary}_${version}_$($t.os)_$($t.arch)${ext}"
                    Invoke-Go -GoArgs @('build', '-o', $out)
                    Write-Host "Built: $out"
                }
            } finally {
                $env:GOOS   = $savedGoos
                $env:GOARCH = $savedGoarch
            }
        }
    }

    Write-Host "Done."
}
finally {
    Restore-GoMod
}

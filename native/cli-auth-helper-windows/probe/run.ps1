#Requires -Version 7.0
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

if (-not $IsWindows) { throw 'This probe requires Windows.' }
if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) {
    throw 'Run from an x64 Visual Studio Native Tools environment.'
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$outputDirectory = Join-Path $repoRoot 'out'
$evidencePath = Join-Path $outputDirectory 'windows-cli-protection-probe.json'
if (Test-Path -LiteralPath $evidencePath) { Remove-Item -LiteralPath $evidencePath -Force }
$temporaryDirectory = Join-Path ([IO.Path]::GetTempPath()) ('riido-cng-probe-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temporaryDirectory | Out-Null
$process = $null
try {
    $binary = Join-Path $temporaryDirectory 'riido-cng-probe.exe'
    $object = Join-Path $temporaryDirectory 'riido-cng-probe.obj'
    $source = Join-Path $PSScriptRoot 'main.cpp'
    & cl.exe /nologo /W4 /WX /std:c++20 /EHsc /MT "/Fe:$binary" "/Fo:$object" $source /link ncrypt.lib bcrypt.lib
    if ($LASTEXITCODE -ne 0) { throw 'Native probe compilation failed.' }

    $stdout = Join-Path $temporaryDirectory 'stdout.json'
    $stderr = Join-Path $temporaryDirectory 'stderr.txt'
    $process = Start-Process -FilePath $binary -ArgumentList '--ci-silent' -PassThru -NoNewWindow `
        -RedirectStandardOutput $stdout -RedirectStandardError $stderr
    if (-not $process.WaitForExit(30000)) {
        $process.Kill($true)
        $process.WaitForExit()
        throw 'Native probe timed out; test-key cleanup is unverified.'
    }
    if ((Get-Item $stdout).Length -gt 16384) { throw 'Probe receipt exceeded its bound.' }
    $receipt = Get-Content -Raw $stdout | ConvertFrom-Json
    $evidence = [ordered]@{
        source_revision = $env:GITHUB_SHA
        runner_image = $env:ImageOS
        runner_image_version = $env:ImageVersion
        architecture = $env:PROCESSOR_ARCHITECTURE
        probe_sha256 = (Get-FileHash $binary -Algorithm SHA256).Hash.ToLowerInvariant()
        process_exit = $process.ExitCode
        measurement = $receipt
    }
    New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
    $evidence | ConvertTo-Json -Depth 8 | Set-Content -Encoding utf8 $evidencePath

    if ($process.ExitCode -ne 0 -or $receipt.version -ne 1 -or
        $receipt.production_supported -cne $false -or $receipt.measurement_complete -cne $true -or
        $receipt.cleanup_complete -cne $true -or $receipt.packaging -cne 'unpackaged') {
        throw 'Native measurement or owned-key cleanup did not complete.'
    }
    $baseline = $receipt.baseline
    if ($baseline.finalize_status -ne 0 -or $baseline.properties_match -cne $true -or
        $baseline.public_shape_valid -cne $true -or $baseline.private_export_status -ne 2148073488 -or
        $baseline.original_sign_status -ne 0 -or $baseline.reopened_sign_status -ne 0 -or
        $baseline.child_sign_status -ne 0) {
        throw 'The nonexportable baseline was not established.'
    }
    $states = @('interactive_setup_required', 'inconclusive_provider_error', 'policy_readback_failed',
                'private_export_allowed', 'silent_use_allowed', 'silent_use_denied_only')
    if ($receipt.candidate_state -cnotin $states) { throw 'Unknown candidate measurement state.' }
    Write-Output "Measurement completed: candidate=$($receipt.candidate_state); production_supported=false."
} finally {
    if ($null -ne $process) {
        if (-not $process.HasExited) {
            $process.Kill($true)
            $process.WaitForExit()
        }
        $process.Dispose()
    }
    Remove-Item -LiteralPath $temporaryDirectory -Recurse -Force
}
